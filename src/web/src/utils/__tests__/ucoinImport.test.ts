import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import type { CatalogRegistry, Coin } from '@/types'
import { buildUcoinCoin, isUcoinImported, parseCsv, parseUcoinCsv, ucoinCoinKey, ucoinCoinName, ucoinMintMark } from '@/utils/ucoinImport'

// The header and rows as uCoin.net exports them in Portuguese (note the
// spaces after some commas and the "Diametro, mm" column with a comma).
const HEADER = '"País","Periodo","Denominação","Ano", "Var.", "Moeda", "Assunto", "Diametro, mm","Conservação","Preço, EUR","Número","Etiqueta","Comentário"'
const SAMPLE = [
  HEADER,
  '"África do Sul","República da África do Sul","5 cêntimos","1984","","Rand Sul Africano","","17.35","","","KM# 84","",""',
  '"Alemanha","União Europeia (Euro)","2 euro","2026 D","D","Euro","Bremen, Casa do Clima","25.75","UNC","2.00","UC# 113","",""',
  '"Polónia ","República","1 grosz","2010","","Zloty polonês","","15.5","","","Y# 276","",""',
  '"Egito","Reino","10 milésimos","1909 ٥","٥","Libra egípcia","","","VF","","KM# 307","","2"',
  '"Portugal","República Portuguesa","2½ euro","2009","","Euro","Ano Internacional da Astronomia","28","PRF","7.50","UC# 51","",""',
].join('\r\n')

const REGISTRY: CatalogRegistry[] = [
  { id: 1, catalog: 'KM', displayName: 'Krause-Mishler', era: 'modern', volumeRequired: false },
  { id: 2, catalog: 'Y', displayName: 'Y Number', era: 'modern', volumeRequired: false },
] as CatalogRegistry[]

const OPTIONS = { categories: ['Roman', 'Greek', 'Modern', 'Other'], eras: ['ancient', 'medieval', 'modern'], registry: REGISTRY, usdPerEur: 1.12 }

describe('uCoin CSV import', () => {
  afterEach(() => setLocale('en'))

  it('splits quoted CSV with escaped quotes, commas, CRLF and a BOM', () => {
    expect(parseCsv('﻿"a, b","say ""hi""", c\r\n"1",2,"3"\n')).toEqual([
      ['a, b', 'say "hi"', 'c'],
      ['1', '2', '3'],
    ])
  })

  it('reads the columns of a uCoin export', () => {
    const { rows, error } = parseUcoinCsv(SAMPLE)
    expect(error).toBeUndefined()
    expect(rows).toHaveLength(5)
    expect(rows[0]).toMatchObject({
      line: 2, country: 'África do Sul', denomination: '5 cêntimos', year: 1984, diameterMm: 17.35, grade: '',
      reference: 'KM# 84', catalogue: { code: 'KM', number: '84' }, priceCurrency: 'EUR',
    })
    expect(rows[1]).toMatchObject({ yearText: '2026 D', year: 2026, variety: 'D', subject: 'Bremen, Casa do Clima', grade: 'UNC', price: 2, catalogue: { code: 'UC', number: '113' } })
    expect(rows[2]).toMatchObject({ country: 'Polónia', catalogue: { code: 'Y', number: '276' } })
    expect(rows[3]).toMatchObject({ year: 1909, variety: '٥', comment: '2' })
  })

  it('rejects files without the uCoin columns', () => {
    expect(parseUcoinCsv('"Name","Value"\n"x","1"').error).toBe('columns')
    expect(parseUcoinCsv(HEADER).error).toBe('empty')
  })

  it('tells mint marks from varieties', () => {
    expect(ucoinMintMark('D')).toBe('D')
    expect(ucoinMintMark('BI')).toBe('BI')
    expect(ucoinMintMark('76')).toBe('')
    expect(ucoinMintMark('☷')).toBe('')
  })

  it('builds a coin from a row, converting the price in euros to dollars', () => {
    setLocale('pt-PT')
    const [southAfrica, bremen, , egypt, astronomy] = parseUcoinCsv(SAMPLE).rows

    const first = buildUcoinCoin(southAfrica, OPTIONS)
    expect(first).toMatchObject({
      name: '5 cêntimos 1984 – África do Sul', category: 'Modern', era: 'modern', material: 'Other',
      denomination: '5 cêntimos', dateRange: '1984', year: 1984, country: 'África do Sul', diameterMm: 17.35, isWishlist: false,
      references: [{ catalog: 'KM', number: '84' }],
    })
    expect(first.purchasePrice).toBeUndefined()
    expect(first.notes).toBe('Importada do uCoin (KM# 84).\nPeríodo: República da África do Sul')

    const second = buildUcoinCoin(bremen, OPTIONS)
    expect(second).toMatchObject({ name: '2 euro 2026 D – Bremen, Casa do Clima', mint: 'D', grade: 'UNC', purchasePrice: 2.24, references: [] })
    expect(second.notes).toContain('(UC# 113)')

    const third = buildUcoinCoin(egypt, OPTIONS)
    expect(third.mint).toBeUndefined()
    expect(third.notes).toContain('Variante: ٥')
    expect(third.notes).toContain('Comentário no uCoin: 2')

    // Without an exchange rate the price in euros stays in the notes.
    const fourth = buildUcoinCoin(astronomy, { ...OPTIONS, usdPerEur: undefined })
    expect(fourth.purchasePrice).toBeUndefined()
    expect(fourth.notes).toContain('Preço no uCoin: 7.5 EUR')
  })

  it('recognises imported coins and the same coin in the file and the collection', () => {
    const [row] = parseUcoinCsv(SAMPLE).rows
    const payload = buildUcoinCoin(row, OPTIONS)
    expect(isUcoinImported(payload as Coin)).toBe(true)
    expect(isUcoinImported({ notes: 'Bought in Lisbon' } as Coin)).toBe(false)
    expect(ucoinCoinKey({ name: ucoinCoinName(row), country: row.country })).toBe(ucoinCoinKey({ name: payload.name!, country: 'africa do sul ' }))
  })
})
