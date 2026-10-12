import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import type { CatalogRegistry, Coin } from '@/types'
import { collectionCurrency } from '@/utils/collectionCurrency'
import {
  buildUcoinCoin,
  isUcoinImported,
  isXlsxFile,
  matchUcoinRows,
  parseCsv,
  parseUcoinCsv,
  parseUcoinDate,
  parseUcoinXlsx,
  ucoinCoinKey,
  ucoinCoinName,
  ucoinMintMark,
  ucoinPurchaseUpdate,
  type UcoinRow,
} from '@/utils/ucoinImport'
import { buildXlsx } from './xlsxFixture'

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

// The columns of uCoin's Excel export (Portuguese), with the purchase details the CSV lacks.
const XLSX_HEADER = ['País', 'Periodo', 'Moeda', 'Denominação', 'Ano', 'Var.', 'Assunto', 'Diametro, mm', 'Conservação', 'Preço, EUR [uCoin]', 'Número', 'Cor', 'Quantidade', 'Necessita reposição', 'Publicado', 'Trocas', 'Compra', 'Preço, EUR [My]', 'Grading company', 'Grading number', 'Grade', 'Etiqueta', 'Comentário']

function excelRow(fields: Partial<Record<(typeof XLSX_HEADER)[number], string | number>>): (string | number)[] {
  return XLSX_HEADER.map((name) => fields[name] ?? '')
}

describe('uCoin Excel import', () => {
  afterEach(() => setLocale('en'))

  it('reads the purchase date, the price paid, uCoin\'s price, swaps, quantities and grading', async () => {
    const { rows, error } = await parseUcoinXlsx(await buildXlsx([
      XLSX_HEADER,
      excelRow({ País: 'Alemanha', Denominação: '2 euro', Ano: 2026, 'Var.': 'D', Assunto: 'Bremen, Casa do Clima', 'Preço, EUR [uCoin]': 2.86, Número: 'UC# 113', Quantidade: 1, Compra: 46174, 'Preço, EUR [My]': 2 }),
      excelRow({ País: 'Espanha', Denominação: '1 peseta', Ano: 1975, 'Var.': '76', 'Preço, EUR [uCoin]': 0.3, Quantidade: 2, Compra: 0, Trocas: 'Atlantius.ue /uid275798' }),
      excelRow({ País: 'Portugal', Denominação: '400 reis', Ano: 1643, Assunto: 'Cruz com letra ""P"" em cada quadrante', 'Grading company': 'PCGS', 'Grading number': '12345678', Grade: 'AU55', Quantidade: 1 }),
    ]))

    expect(error).toBeUndefined()
    expect(rows).toHaveLength(4)
    expect(rows[0]).toMatchObject({ yearText: '2026 D', variety: 'D', price: 2, priceCurrency: 'EUR', value: 2.86, valueCurrency: 'EUR', purchaseDate: '2026-06-01', swapWith: '' })
    // Two of the same coin; the variety joins the year as in the CSV; no purchase date.
    expect(rows[1]).toBe(rows[2])
    expect(rows[1]).toMatchObject({ yearText: '1975 76', purchaseDate: undefined, price: undefined, swapWith: 'Atlantius.ue' })
    expect(rows[3]).toMatchObject({ subject: 'Cruz com letra "P" em cada quadrante', slab: { company: 'PCGS', grade: 'AU55', number: '12345678' } })
  })

  it('builds coins with the purchase details and a graded coin\'s grade', async () => {
    setLocale('pt-PT')
    const { rows } = await parseUcoinXlsx(await buildXlsx([
      XLSX_HEADER,
      excelRow({ País: 'Espanha', Denominação: '1 peseta', Ano: 1975, 'Var.': '76', 'Preço, EUR [uCoin]': 0.3, Compra: 45931, Trocas: 'Atlantius.ue /uid275798' }),
      excelRow({ País: 'Portugal', Denominação: '400 reis', Ano: 1643, Conservação: 'VF', 'Grading company': 'PCGS', 'Grading number': '12345678', Grade: 'AU55' }),
    ]))
    const euros = { ...OPTIONS, usdPerEur: undefined }
    collectionCurrency.value = 'EUR'
    try {
      expect(buildUcoinCoin(rows[0]!, euros)).toMatchObject({ purchaseDate: '2025-10-01', purchaseLocation: 'Troca no uCoin com Atlantius.ue', currentValue: 0.3, purchasePrice: undefined })
      const graded = buildUcoinCoin(rows[1]!, euros)
      expect(graded).toMatchObject({ grade: 'PCGS AU55', presentation: 'slab' })
      expect(graded.notes).toContain('Certificado de grading: 12345678')
    } finally {
      collectionCurrency.value = 'USD'
    }
  })

  it('tells an unreadable file from one without uCoin\'s columns', async () => {
    expect((await parseUcoinXlsx(new TextEncoder().encode('not a workbook').buffer)).error).toBe('unreadable')
    expect((await parseUcoinXlsx(await buildXlsx([['Nome', 'Valor'], ['x', 1]]))).error).toBe('columns')
    expect(isXlsxFile({ name: 'Coleção.XLSX', type: '' })).toBe(true)
    expect(isXlsxFile({ name: 'coleção.csv', type: 'text/csv' })).toBe(false)
  })

  it('reads purchase dates written as text too', () => {
    expect(parseUcoinDate('01.06.2026')).toBe('2026-06-01')
    expect(parseUcoinDate('2026-06-01')).toBe('2026-06-01')
    expect(parseUcoinDate('31/02/2026')).toBeUndefined()
    expect(parseUcoinDate('0')).toBeUndefined()
  })
})

describe('matching the file with coins imported earlier', () => {
  function row(fields: Partial<UcoinRow>): UcoinRow {
    return { line: 2, country: 'Espanha', period: '', denomination: '1 peseta', yearText: '1975 76', year: 1975, variety: '76', subject: '', grade: '', priceCurrency: 'EUR', valueCurrency: 'EUR', swapWith: '', reference: '', comment: '', ...fields }
  }
  function coin(id: number, name: string, country = 'Espanha', fields: Partial<Coin> = {}): Coin {
    return { id, name, country, notes: 'Importada do uCoin', purchaseDate: null, purchasePrice: null, purchaseLocation: '', currentValue: null, ...fields } as unknown as Coin
  }

  it('pairs rows by name, then by denomination, year and variety when Excel left the subject out', () => {
    const { matched, pending } = matchUcoinRows(
      [
        row({ subject: 'Bremen', denomination: '2 euro', yearText: '2026 D', year: 2026, country: 'Alemanha' }),
        row({}),
        row({}),
        row({ denomination: '40 para', yearText: '1909 0', year: 1909, variety: '0', country: 'Império Otomano' }),
        row({ denomination: '5 pesetas', yearText: '1957 58', year: 1957, variety: '58', subject: 'Nova' }),
      ],
      [
        coin(3, '1 peseta 1975 76 – 76 dentro da estrela'),
        coin(1, '2 euro 2026 D – Bremen', 'Alemanha'),
        coin(4, '40 para 1909 ٥ – "٥" Abaixo da Tughra', 'Império Otomano'),
      ],
    )
    expect(matched.map(({ coin }) => coin.id)).toEqual([1, 3, 4])
    // A second 1 peseta 1975 and a coin with a subject of its own are new.
    expect(pending.map((entry) => entry.denomination)).toEqual(['1 peseta', '5 pesetas'])
  })

  it('adds only what a coin does not have yet, keeping its other values', () => {
    const purchase = { purchaseDate: '2026-06-01', purchasePrice: 2, purchaseLocation: 'Troca no uCoin com Atlantius.ue', currentValue: 2.86 }
    const update = ucoinPurchaseUpdate(coin(1, 'x', 'y', { purchasePrice: 1.5, currentValue: 4, weightGrams: 8.5 } as Partial<Coin>), purchase)
    expect(update).toMatchObject({ purchaseDate: '2026-06-01', purchaseLocation: 'Troca no uCoin com Atlantius.ue', purchasePrice: 1.5, currentValue: 4, weightGrams: 8.5 })
    expect(ucoinPurchaseUpdate(coin(2, 'x', 'y', { purchaseDate: '2025-01-01T00:00:00Z', purchasePrice: 2, purchaseLocation: 'Feira', currentValue: 3 } as Partial<Coin>), purchase)).toBeNull()
  })
})
