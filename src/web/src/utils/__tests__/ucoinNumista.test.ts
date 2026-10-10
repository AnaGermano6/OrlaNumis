import { describe, expect, it } from 'vitest'
import type { NumistaCatalogSearchItem } from '@/api/client'
import type { Coin } from '@/types'
import {
  issuerSimilarity,
  parseDenomination,
  pickIssue,
  pickNumistaType,
  priceForGrade,
  subjectScore,
  ucoinLookupForCoin,
  unitQuery,
  ucoinLookupForRow,
  ucoinLookupGroupKey,
  valuationGrade,
  type UcoinLookup,
} from '@/utils/ucoinNumista'
import { parseUcoinCsv } from '@/utils/ucoinImport'

const GERMANY = { code: 'allemagne', name: 'Alemanha, República Federal da' }

// Numista's answer for issuer=allemagne, year=2026, q=euro (pt), trimmed.
const GERMANY_2026: NumistaCatalogSearchItem[] = [
  { id: 105, title: '1 Euro Cent', issuer: GERMANY, minYear: 2002, maxYear: 2026 },
  { id: 106, title: '2 Euro Cents', issuer: GERMANY, minYear: 2002, maxYear: 2026 },
  { id: 6323, title: '2 Euros (2nd map)', issuer: GERMANY, minYear: 2008, maxYear: 2026 },
  { id: 544530, title: '2 Euros (Konrad Adenauer)', issuer: GERMANY, minYear: 2026, maxYear: 2026 },
  { id: 474550, title: '2 Euros (Bundesländer II - "Bremen")', issuer: GERMANY, minYear: 2026, maxYear: 2026 },
  { id: 570281, title: '50 Euros (Biblical Magi)', issuer: GERMANY, minYear: 2026, maxYear: 2026 },
]

function lookup(fields: Partial<UcoinLookup>): UcoinLookup {
  return { country: 'Alemanha', denomination: '2 euro', year: 2026, mintMark: '', variety: '', subject: '', ...fields }
}

describe('matching uCoin coins to Numista', () => {
  it.each([
    ['2 euro', 2, false],
    ['2½ euro', 2.5, false],
    ['10 cêntimos de euro', 10, true],
    ['2 Euro Cents', 2, true],
    ['50 pence', 50, true],
    ['1 libra', 1, false],
    ['1⁄14 Thaler', 1 / 14, false],
  ])('reads the denomination "%s"', (text, value, subunit) => {
    const parsed = parseDenomination(text)
    expect(parsed.value).toBeCloseTo(value)
    expect(parsed.subunit).toBe(subunit)
  })

  it('matches subjects written in Portuguese against English titles', () => {
    expect(subjectScore('Turíngia, Castelo de Wartburg em Eisenach', '2 Euros (Bundesländer II - "Thuringia")')).toBeGreaterThan(0)
    expect(subjectScore('Turíngia, Castelo de Wartburg em Eisenach', '2 Euros (Bundesländer - "Thüringen")')).toBe(1)
    expect(subjectScore('Saxônia-Anhalt, Catedral de Magdeburg', '2 Euros (Bundesländer II - "Saxony-Anhalt")')).toBe(2)
    expect(subjectScore('35º Aniversário - Programa Erasmus', '2 Euros (35 years of the Erasmus Programme)')).toBe(3)
    expect(subjectScore('30° aniversário - Queda do Muro de Berlim', '2 Euros (Fall of the Berlin Wall)')).toBeGreaterThan(0)
    expect(subjectScore('Palácio Sanssouci em Postdam, Brandeburgo', '2 Euros (Bundesländer II - "Brandenburg")')).toBeGreaterThan(0)
    expect(subjectScore('Bremen, Casa do Clima', '2 Euros (Konrad Adenauer)')).toBe(0)
    expect(subjectScore('Gesto Histórico de Willy Brandt em Varsóvia', '2 Euros (Bundesländer - "Brandenburg")')).toBe(0)
    expect(subjectScore('Palácio Sanssouci em Postdam, Brandeburgo', '2 Euros (Bundesländer - "Brandenburg")')).toBe(1)
    expect(subjectScore('Jornada Mundial da Juventude 2023 Lisboa', '2 Euros (World Youth Day Lisbon)')).toBe(1)
    expect(subjectScore('75º Aniversário das Nações Unidas', '2 Euros (75th anniversary of United Nations)')).toBe(1)
    // The country's own name is no evidence: "França" is not "François".
    expect(subjectScore('Campeonato Europeu de Futebol, França 2016', '2 Euros (François Mitterrand)', 'França')).toBe(0)
    expect(subjectScore('Faróis de Portugal - Cabo de São Vicente', '5 Euro (Portuguese tiles; 16th centuries)', 'Portugal')).toBe(0)
    expect(subjectScore('Faróis de Portugal - Cabo de São Vicente', '5 Euro (Cape St. Vincent; cupronickel)', 'Portugal')).toBe(1)
  })

  it('compares uCoin countries with Numista issuer names', () => {
    expect(issuerSimilarity('Alemanha', 'Alemanha, República Federal da')).toBe(2)
    expect(issuerSimilarity('Portugal', 'Portugal')).toBe(3)
    expect(issuerSimilarity('Estados Unidos da América', 'Estados Unidos')).toBeGreaterThan(0)
    expect(issuerSimilarity('Portugal', 'Espanha')).toBe(0)
  })

  it('takes the single result of a catalogue-number search', () => {
    const southAfrica = [{ id: 1791, title: '5 Cents (SUID-AFRIKA - SOUTH AFRICA)', issuer: { code: 'afrique_du_sud', name: 'África do Sul' }, minYear: 1970, maxYear: 1989 }]
    expect(pickNumistaType(southAfrica, lookup({ country: 'África do Sul', denomination: '5 cêntimos', year: 1984 }), 'catalogue')?.id).toBe(1791)
    // The number and issuer decide; uCoin and Numista sometimes date a type differently.
    expect(pickNumistaType(southAfrica, lookup({ country: 'África do Sul', year: 1991 }), 'catalogue')?.id).toBe(1791)
  })

  it('takes the ordinary version of a design when several versions fit', () => {
    const netherlands = { code: 'pays-bas', name: 'Países Baixos' }
    const jaapEden = [
      { id: 2002, title: '5 Euros - Willem-Alexander (Jaap Eden)', issuer: netherlands, minYear: 2019, maxYear: 2019 },
      { id: 2001, title: '5 Euros - Willem-Alexander (Jaap Eden)', issuer: netherlands, minYear: 2019, maxYear: 2019 },
      { id: 2003, title: '10 Euros - Willem-Alexander (Jaap Eden)', issuer: netherlands, minYear: 2019, maxYear: 2019 },
    ]
    expect(pickNumistaType(jaapEden, lookup({ country: 'Países Baixos', denomination: '5 euro', year: 2019, subject: 'Jaap Eden' }), 'issuer')?.id).toBe(2001)
    const assen = [
      { id: 3002, title: '5 Euro - Willem Alexander (TT Assen; silver)', issuer: netherlands, minYear: 2025, maxYear: 2025 },
      { id: 3001, title: '5 Euros - Willem Alexander (TT Assen; silver plated copper)', issuer: netherlands, minYear: 2025, maxYear: 2025 },
    ]
    expect(pickNumistaType(assen, lookup({ country: 'Países Baixos', denomination: '5 euro', year: 2025, subject: '100º Aniversário - Circuito TT Assen' }), 'issuer')?.id).toBe(3001)
    // Versions found by catalogue number too.
    const canada = { code: 'canada', name: 'Canadá' }
    const jubilee = [
      { id: 4003, title: '1 Cent - Elizabeth II (Golden Jubilee; non-magnetic)', issuer: canada, minYear: 2002, maxYear: 2002 },
      { id: 4002, title: '1 Cent - Elizabeth II (Golden Jubilee; magnetic)', issuer: canada, minYear: 2002, maxYear: 2002 },
    ]
    expect(pickNumistaType(jubilee, lookup({ country: 'Canadá', denomination: '1 cêntimo', year: 2002 }), 'catalogue')?.id).toBe(4002)
    // Two designs are still not guessed between.
    const twoDesigns = [
      { id: 5001, title: '2 Euros (World Youth Day)', issuer: { code: 'portugal', name: 'Portugal' }, minYear: 2023, maxYear: 2023 },
      { id: 5002, title: '2 Euros (Peace Among Nations)', issuer: { code: 'portugal', name: 'Portugal' }, minYear: 2023, maxYear: 2023 },
    ]
    expect(pickNumistaType(twoDesigns, lookup({ country: 'Portugal', denomination: '2 euro', year: 2023, subject: 'Paz Entre as Nações' }), 'issuer')).toBeUndefined()
    // Versions of a design the subject does not name are not taken for it.
    const austria = { code: 'autriche', name: 'Áustria' }
    const hares = [
      { id: 6002, title: "5 Euros (Dürer's Hare Trio, silver)", issuer: austria, minYear: 2026, maxYear: 2026 },
      { id: 6001, title: "5 Euros (Dürer's Hare Trio, copper)", issuer: austria, minYear: 2026, maxYear: 2026 },
    ]
    expect(pickNumistaType(hares, lookup({ country: 'Áustria', denomination: '5 euro', year: 2026, subject: 'Mergulhe na Felicidade' }), 'issuer')).toBeUndefined()
  })

  it('reads the Ukrainian kopiyok however it is spelled', () => {
    const ukraine = { code: 'ukraine', name: 'Ucrânia' }
    const types = [
      { id: 3104, title: '5 Kopiiok', issuer: ukraine, minYear: 1992, maxYear: 2018 },
      { id: 3303, title: '1 Kopiika', issuer: ukraine, minYear: 1992, maxYear: 2018 },
      { id: 9697, title: '1 Hryvnia (without mintmark)', issuer: ukraine, minYear: 1992, maxYear: 1996 },
    ]
    expect(pickNumistaType(types, lookup({ country: 'Ucrânia', denomination: '5 kopiyok', year: 1992 }), 'issuer')?.id).toBe(3104)
  })

  it('never takes a catalogue number from another country', () => {
    const norway = [{ id: 1445, title: '5 Kroner - Olav V (Type 1)', issuer: { code: 'norvege', name: 'Noruega' }, minYear: 1963, maxYear: 1973 }]
    expect(pickNumistaType(norway, lookup({ country: 'Egito', denomination: '5 piastres', year: 1967 }), 'catalogue')).toBeUndefined()
  })

  it('narrows a catalogue-number search to the coin\'s issuer', () => {
    const km84 = [
      { id: 15867, title: '2 Reichsmark (Friedrich Schiller)', issuer: { code: 'allemagne-pre1945', name: 'Alemanha (1871-1948)' }, minYear: 1934, maxYear: 1934 },
      { id: 900, title: '1 Lek', issuer: { code: 'albanie', name: 'Albânia' }, minYear: 1934, maxYear: 1935 },
    ]
    expect(pickNumistaType(km84, lookup({ denomination: '2 reichsmark', year: 1934 }), 'catalogue')?.id).toBe(15867)
  })

  it('finds commemoratives by subject and the regular coin when there is none', () => {
    expect(pickNumistaType(GERMANY_2026, lookup({ subject: 'Bremen, Casa do Clima' }), 'issuer')?.id).toBe(474550)
    expect(pickNumistaType(GERMANY_2026, lookup({ subject: 'Konrad Adenauer - 150º aniversário' }), 'issuer')?.id).toBe(544530)
    expect(pickNumistaType(GERMANY_2026, lookup({}), 'issuer')?.id).toBe(6323)
    // "2 euro" is never the 2 cent coin, and an unknown subject is not guessed.
    expect(pickNumistaType(GERMANY_2026, lookup({ denomination: '2 cêntimos de euro' }), 'issuer')?.id).toBe(106)
    expect(pickNumistaType(GERMANY_2026, lookup({ subject: 'Hamburgo, Elbphilharmonie' }), 'issuer')).toBeUndefined()
    expect(pickNumistaType(GERMANY_2026, lookup({ denomination: '20 euro' }), 'issuer')).toBeUndefined()
  })

  it('prefers the ordinary coin to its silver, gold or proof versions', () => {
    const portugal = { code: 'portugal', name: 'Portugal' }
    const tribute = [
      { id: 470475, title: '5 Euro (Camilo Castelo Branco)', issuer: portugal, minYear: 2025, maxYear: 2025 },
      { id: 470477, title: '5 Euro (Camilo Castelo Branco - silver)', issuer: portugal, minYear: 2025, maxYear: 2025 },
      { id: 470480, title: '5 Euro (Camilo Castelo Branco - gold)', issuer: portugal, minYear: 2025, maxYear: 2025 },
    ]
    const camilo = lookup({ country: 'Portugal', denomination: '5 euro', year: 2025, subject: 'Camilo Castelo Branco' })
    expect(pickNumistaType(tribute, camilo, 'issuer')?.id).toBe(470475)
    expect(pickNumistaType(tribute, { ...camilo, subject: 'Camilo Castelo Branco - Prata' }, 'issuer')?.id).toBe(470477)

    const uk = { code: 'royaume-uni', name: 'Reino Unido' }
    const pounds = [
      { id: 1, title: '1 Pound - Elizabeth II (5th portrait; 12-sided)', issuer: uk, minYear: 2016, maxYear: 2022 },
      { id: 2, title: '1 Pound - Elizabeth II (12-sided; Silver Proof)', issuer: uk, minYear: 2016, maxYear: 2022 },
    ]
    expect(pickNumistaType(pounds, lookup({ country: 'Reino Unido', denomination: '1 libra', year: 2019 }), 'issuer')?.id).toBe(1)
    // "Sem marca" (no mark) says nothing about the design: treated as no subject.
    expect(pickNumistaType(pounds, lookup({ country: 'Reino Unido', denomination: '1 libra', year: 2019, subject: 'sem marca' }), 'issuer')?.id).toBe(1)
  })

  it.each([
    ['2 euro', 'euro'],
    ['10 cêntimos de euro', 'euro'],
    ['1 libra', 'pound'],
    ['1 dólar', 'dollar'],
    ['50 ienes', 'yen'],
    ['10 hellers', undefined],
  ])('searches an issuer\'s "%s" coins with "%s"', (denomination, query) => {
    expect(unitQuery(denomination)).toBe(query)
  })

  it('picks the issue of the coin\'s year and mint mark', () => {
    const issues = [
      { id: 1, year: 2019, mintLetter: 'A', mintage: 6_000_000 },
      { id: 2, year: 2019, mintLetter: 'D', mintage: 6_300_000 },
      { id: 3, year: 2020, mintLetter: 'A', mintage: 5_000_000 },
    ]
    expect(pickIssue(issues, { year: 2019, mintMark: 'D', variety: '' })?.id).toBe(2)
    expect(pickIssue(issues, { year: 2019, mintMark: '', variety: '' })?.id).toBe(1)
    expect(pickIssue(issues, { year: 2021, mintMark: '', variety: '' })).toBeUndefined()
    expect(pickIssue([{ id: 9, minYear: 1950, maxYear: 1960 }], { year: 1955, mintMark: '', variety: '' })?.id).toBe(9)
    // Spanish pesetas carry the striking year in a star: "1975 76".
    const pesetas = [{ id: 4, year: 1975, comment: '*75' }, { id: 5, year: 1975, comment: '*76' }, { id: 6, year: 1975, comment: '*77' }]
    expect(pickIssue(pesetas, { year: 1975, mintMark: '', variety: '76' })?.id).toBe(5)
  })

  it('values a coin at its grade, else the nearest grade with a price', () => {
    const prices = [{ grade: 'f', price: 1 }, { grade: 'xf', price: 3 }, { grade: 'unc', price: 6 }]
    expect(priceForGrade(prices, 'xf')).toBe(3)
    expect(priceForGrade(prices, 'vf')).toBe(1)
    expect(priceForGrade(prices, 'g')).toBe(1)
    expect(priceForGrade([], 'vf')).toBeUndefined()
    expect(valuationGrade('UNC')).toBe('unc')
    expect(valuationGrade('PRF')).toBe('unc')
    expect(valuationGrade('')).toBe('vf')
  })

  it('reads the same lookup back from the created coin', () => {
    const csv = [
      '"País","Periodo","Denominação","Ano", "Var.", "Moeda", "Assunto", "Diametro, mm","Conservação","Preço, EUR","Número","Etiqueta","Comentário"',
      '"Alemanha","União Europeia (Euro)","2 euro","2022 D","D","Euro","35º Aniversário - Programa Erasmus","25.75","G","","UC# 106","",""',
      '"Alemanha","União Europeia (Euro)","2 euro","2018 J","J","Euro","Berlin","25.75","","","KM# 367","",""',
      '"África do Sul","República","5 cêntimos","1984","","Rand","","17.35","","","KM# 84","",""',
    ].join('\n')
    const [erasmus, berlin, southAfrica] = parseUcoinCsv(csv).rows
    const asCoin = (name: string, row: typeof erasmus, references: Coin['references'] = []) =>
      ({ name, country: row.country, denomination: row.denomination, year: row.year, mint: row.variety, references }) as Coin

    expect(ucoinLookupForCoin(asCoin('2 euro 2022 D – 35º Aniversário - Programa Erasmus', erasmus))).toEqual(ucoinLookupForRow(erasmus))
    const km = [{ id: 1, coinId: 1, catalog: 'KM', volume: '', number: '367', uri: '', createdAt: '', updatedAt: '' }]
    expect(ucoinLookupForCoin(asCoin('2 euro 2018 J – Berlin', berlin, km))).toEqual(ucoinLookupForRow(berlin))
    // Without a subject the name ends with the country, which is not a subject.
    const zaf = [{ id: 2, coinId: 2, catalog: 'KM', volume: '', number: '84', uri: '', createdAt: '', updatedAt: '' }]
    const [star] = parseUcoinCsv(`${csv.split('\n')[0]}\n"Espanha","Rei Juan Carlos I","1 peseta","1975 76","76","Peseta","76 dentro da estrela","","","","KM# 806","",""`).rows
    const km806 = [{ id: 3, coinId: 3, catalog: 'KM', volume: '', number: '806', uri: '', createdAt: '', updatedAt: '' }]
    expect(ucoinLookupForCoin({ ...asCoin('1 peseta 1975 76 – 76 dentro da estrela', star, km806), mint: '' })).toEqual(ucoinLookupForRow(star))
    expect(ucoinLookupForRow(star).variety).toBe('76')
    expect(ucoinLookupForCoin(asCoin('5 cêntimos 1984 – África do Sul', southAfrica, zaf))).toEqual(ucoinLookupForRow(southAfrica))

    // Coins of one Numista type share a key: the same KM number in other years, or the same subject and year.
    expect(ucoinLookupGroupKey(ucoinLookupForRow(berlin))).toBe(ucoinLookupGroupKey({ ...ucoinLookupForRow(berlin), year: 2019, mintMark: 'A' }))
    expect(ucoinLookupGroupKey(ucoinLookupForRow(erasmus))).toBe(ucoinLookupGroupKey({ ...ucoinLookupForRow(erasmus), mintMark: 'F' }))
  })
})
