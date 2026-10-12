// Reads a collection exported from uCoin.net (CSV or Excel) and turns each row
// into a coin for this app. Only what the file says is used here; catalogue
// details (metal, weight, rarity, value) are added afterwards from Numista.
import { t } from '@/i18n'
import type { CatalogRegistry, Category, Coin, CoinMutationPayload, CoinReferenceInput } from '@/types'
import { toCollectionCurrency } from '@/utils/collectionCurrency'
import { eraForYear } from '@/utils/numistaCatalogPrefill'
import { excelSerialDate, readXlsxRows } from '@/utils/xlsxReader'

/** Reference catalogues Numista can search by number; uCoin's own "UC#" is not one. */
export const UCOIN_SEARCHABLE_CATALOGUES = ['KM', 'Y'] as const
export type UcoinCatalogueCode = (typeof UCOIN_SEARCHABLE_CATALOGUES)[number]

export interface UcoinRow {
  /** Line in the file (the header is line 1), for messages. */
  line: number
  country: string
  period: string
  denomination: string
  /** The year as written, with any mint mark: "2026 D". */
  yearText: string
  year?: number
  /** uCoin's "Var." column: a mint mark (D, BI) or a variety (76, ☷). */
  variety: string
  subject: string
  diameterMm?: number
  grade: string
  /** What was paid ("Preço, EUR" in a CSV, "Preço, EUR [My]" in Excel). */
  price?: number
  /** Currency of the price column ("Preço, EUR" → EUR). */
  priceCurrency: string
  /** uCoin's own price for the coin ("Preço, EUR [uCoin]"). */
  value?: number
  valueCurrency: string
  /** Purchase date, YYYY-MM-DD ("Compra"). */
  purchaseDate?: string
  /** uCoin member the coin came from in a swap ("Trocas"). */
  swapWith: string
  /** Graded and sealed: "PCGS", "MS65", certificate number. */
  slab?: { company: string; grade: string; number: string }
  /** Catalogue number as written: "KM# 84", "UC# 113". */
  reference: string
  catalogue?: { code: string; number: string }
  comment: string
}

export type UcoinParseError = 'empty' | 'columns' | 'unreadable'

export interface UcoinParseResult {
  rows: UcoinRow[]
  error?: UcoinParseError
}

type Column = 'country' | 'period' | 'denomination' | 'year' | 'variety' | 'subject' | 'diameter' | 'grade' | 'number' | 'comment'
  | 'purchaseDate' | 'swap' | 'quantity' | 'gradingCompany' | 'gradingNumber'

// Header names in the languages uCoin exports in, accents and case folded.
const HEADER_NAMES: Record<Column, string[]> = {
  country: ['pais', 'country'],
  period: ['periodo', 'period'],
  denomination: ['denominacao', 'denomination'],
  year: ['ano', 'year'],
  variety: ['var.', 'var', 'variety', 'mint mark'],
  subject: ['assunto', 'subject', 'theme'],
  diameter: ['diametro, mm', 'diameter, mm'],
  grade: ['conservacao', 'condition', 'grade'],
  number: ['numero', 'number', 'catalog number'],
  comment: ['comentario', 'comment'],
  purchaseDate: ['compra', 'data de compra', 'purchase', 'purchase date', 'bought'],
  swap: ['trocas', 'troca', 'swap', 'swaps', 'exchange'],
  quantity: ['quantidade', 'quantity', 'qty'],
  gradingCompany: ['grading company', 'empresa de certificacao'],
  gradingNumber: ['grading number', 'numero de certificacao'],
}

// Price columns: "Preço, EUR" (CSV: what was paid), "Preço, EUR [My]" (paid)
// and "Preço, EUR [uCoin]" (uCoin's price).
const PRICE_HEADER = /^(preco|price)(,|$)/
const MY_PRICE = /\[(my|minha|meu|mine)\]/
const UCOIN_PRICE = /\[ucoin\]/


function fold(value: string): string {
  return value.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '').trim()
}

function clip(value: string, max: number): string {
  return value.length > max ? `${value.slice(0, max - 1)}…` : value
}

/** Splits CSV text into rows of fields: quoted fields, "" escapes, CRLF, BOM. */
export function parseCsv(text: string): string[][] {
  const rows: string[][] = []
  let row: string[] = []
  let field = ''
  let quoted = false
  let fieldStarted = false
  const source = text.replace(/^\uFEFF/, '')
  for (let i = 0; i < source.length; i += 1) {
    const char = source[i]
    if (quoted) {
      if (char === '"' && source[i + 1] === '"') {
        field += '"'
        i += 1
      } else if (char === '"') {
        quoted = false
      } else {
        field += char
      }
      continue
    }
    if (char === '"' && !fieldStarted) {
      quoted = true
      fieldStarted = true
    } else if (char === ',') {
      row.push(field)
      field = ''
      fieldStarted = false
    } else if (char === '\n' || char === '\r') {
      if (char === '\r' && source[i + 1] === '\n') i += 1
      row.push(field)
      if (row.some((value) => value.trim())) rows.push(row)
      row = []
      field = ''
      fieldStarted = false
    } else if (char === ' ' && !fieldStarted) {
      // Spaces after a comma ("Ano", "Var.") are not part of the field.
    } else {
      field += char
      fieldStarted = true
    }
  }
  row.push(field)
  if (row.some((value) => value.trim())) rows.push(row)
  return rows
}

/** A purchase date as written: an Excel serial number, DD.MM.YYYY, DD/MM/YYYY or YYYY-MM-DD. */
export function parseUcoinDate(value: string): string | undefined {
  const text = value.trim()
  if (!text || text === '0') return undefined
  if (/^\d+(\.\d+)?$/.test(text)) return excelSerialDate(text)
  const iso = /^(\d{4})-(\d{1,2})-(\d{1,2})/.exec(text)
  const european = /^(\d{1,2})[./](\d{1,2})[./](\d{4})$/.exec(text)
  const [year, month, day] = iso ? [iso[1], iso[2], iso[3]] : european ? [european[3], european[2], european[1]] : []
  if (!year || !month || !day) return undefined
  const date = new Date(Date.UTC(Number(year), Number(month) - 1, Number(day)))
  return Number.isNaN(date.getTime()) || date.getUTCDate() !== Number(day) ? undefined : date.toISOString().slice(0, 10)
}

function parseNumber(value: string): number | undefined {
  const normalized = value.trim().replace(/\s/g, '').replace(',', '.')
  if (!normalized) return undefined
  const number = Number(normalized)
  return Number.isFinite(number) && number > 0 ? number : undefined
}

function parseCatalogue(reference: string): UcoinRow['catalogue'] {
  const match = /^\s*([A-Za-z]{1,10})\s*#\s*(\S.*?)\s*$/.exec(reference)
  return match?.[1] && match[2] ? { code: match[1].toUpperCase(), number: match[2] } : undefined
}

function priceCurrency(header: string | undefined): string {
  return (header?.split(',')[1] ?? '').replace(/\[[^\]]*\]/g, '').trim().toUpperCase()
}

/** At most this many coins from one line ("Quantidade"). */
const MAX_QUANTITY = 50

/** Reads uCoin's columns from a table (a header row, then a row per coin or per kind of coin). */
export function parseUcoinTable(table: string[][]): UcoinParseResult {
  const [headerCells, ...dataRows] = table.filter((row) => row.some((value) => value.trim()))
  if (!headerCells || !dataRows.length) return { rows: [], error: 'empty' }
  const header = headerCells.map(fold)
  const columns = new Map<Column, number>()
  for (const [column, names] of Object.entries(HEADER_NAMES) as [Column, string[]][]) {
    const index = header.findIndex((name) => names.includes(name))
    if (index >= 0) columns.set(column, index)
  }
  // A grade column besides the condition one is the grade of a graded coin.
  const slabGradeIndex = header.findIndex((name, index) => name === 'grade' && index !== columns.get('grade'))
  const priceHeaders = header.map((name, index) => ({ name, index })).filter(({ name }) => PRICE_HEADER.test(name))
  const paidIndex = (priceHeaders.find(({ name }) => MY_PRICE.test(name)) ?? priceHeaders.find(({ name }) => !name.includes('[')))?.index
  const valueIndex = priceHeaders.find(({ name }) => UCOIN_PRICE.test(name))?.index
  if (!columns.has('country') || !columns.has('denomination') || !columns.has('year')) return { rows: [], error: 'columns' }

  const cell = (fields: string[], column: Column | number | undefined) => {
    const index = typeof column === 'number' || column === undefined ? column : columns.get(column)
    return index === undefined || index < 0 ? '' : (fields[index] ?? '').trim()
  }
  const rows: UcoinRow[] = []
  dataRows.forEach((fields, index) => {
    // The CSV writes the variety after the year ("2026 D"); Excel keeps it in
    // its own column. Both name the coin the same way.
    const variety = cell(fields, 'variety')
    const writtenYear = cell(fields, 'year')
    const yearText = variety && !writtenYear.split(/\s+/).includes(variety) ? `${writtenYear} ${variety}`.trim() : writtenYear
    const yearMatch = /^-?\d{1,4}/.exec(yearText)
    const reference = cell(fields, 'number')
    const slabGrade = cell(fields, slabGradeIndex)
    const company = cell(fields, 'gradingCompany')
    const row: UcoinRow = {
      line: index + 2,
      country: cell(fields, 'country'),
      period: cell(fields, 'period'),
      denomination: cell(fields, 'denomination'),
      yearText,
      year: yearMatch ? Number(yearMatch[0]) : undefined,
      variety,
      subject: cell(fields, 'subject'),
      diameterMm: parseNumber(cell(fields, 'diameter')),
      grade: cell(fields, 'grade'),
      price: parseNumber(cell(fields, paidIndex)),
      priceCurrency: priceCurrency(paidIndex === undefined ? undefined : headerCells[paidIndex]),
      value: parseNumber(cell(fields, valueIndex)),
      valueCurrency: priceCurrency(valueIndex === undefined ? undefined : headerCells[valueIndex]),
      purchaseDate: parseUcoinDate(cell(fields, 'purchaseDate')),
      // "Atlantius.ue /uid275798": the member's name, without the account number.
      swapWith: cell(fields, 'swap').replace(/\s*\/\s*uid\d+\s*$/i, '').trim(),
      slab: company || slabGrade ? { company, grade: slabGrade, number: cell(fields, 'gradingNumber') } : undefined,
      reference,
      catalogue: parseCatalogue(reference),
      comment: cell(fields, 'comment'),
    }
    if (!row.country && !row.denomination) return
    const quantity = Math.min(MAX_QUANTITY, Math.max(1, Math.floor(parseNumber(cell(fields, 'quantity')) ?? 1)))
    for (let copy = 0; copy < quantity; copy += 1) rows.push(row)
  })
  return { rows }
}

export function parseUcoinCsv(text: string): UcoinParseResult {
  return parseUcoinTable(parseCsv(text))
}

/** Reads uCoin's Excel export (.xlsx). */
export async function parseUcoinXlsx(buffer: ArrayBuffer): Promise<UcoinParseResult> {
  let table: string[][]
  try {
    table = await readXlsxRows(buffer)
  } catch {
    return { rows: [], error: 'unreadable' }
  }
  // uCoin's Excel export keeps the CSV's doubled quotes: 'letra ""P""'.
  return parseUcoinTable(table.map((row) => row.map((value) => value.replace(/""/g, '"'))))
}

/** Whether a file is an Excel workbook rather than CSV text. */
export function isXlsxFile(file: Pick<File, 'name' | 'type'>): boolean {
  return /\.xlsx$/i.test(file.name) || file.type === 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
}

/** A mint mark is a short run of letters ("D", "BI"); anything else is a variety. */
export function ucoinMintMark(variety: string): string {
  return /^[A-Za-z]{1,3}$/.test(variety.trim()) ? variety.trim() : ''
}

/** "2 euro 2026 D – Bremen, Casa do Clima", or the country when there is no subject. */
export function ucoinCoinName(row: Pick<UcoinRow, 'denomination' | 'yearText' | 'subject' | 'country'>): string {
  const base = [row.denomination, row.yearText].filter(Boolean).join(' ')
  const detail = row.subject || row.country
  return clip(detail ? `${base} – ${detail}` : base, 200)
}

// Every imported coin's notes start with this mark ("uCoin" in any language),
// so a second import of the same file can skip coins already added.
const UCOIN_MARK = /\buCoin\b/

export function isUcoinImported(coin: Pick<Coin, 'notes'>): boolean {
  return UCOIN_MARK.test(coin.notes ?? '')
}

/** Identifies the same coin in the file and in the collection. */
export function ucoinCoinKey(coin: { name: string; country?: string }): string {
  return `${fold(coin.name)}|${fold(coin.country ?? '')}`
}

/** "2 euro 2026 D – Bremen" → "2 euro 2026 d". */
function nameBase(name: string): string {
  return fold(name.split(' – ')[0] ?? name)
}

export interface UcoinMatch {
  /** Rows of coins already in the collection, each with its coin. */
  matched: { row: UcoinRow; coin: Coin }[]
  /** Rows of coins not in the collection yet. */
  pending: UcoinRow[]
}

/**
 * Pairs the file's rows with coins imported earlier, each coin once, oldest
 * first: by name and country; then, for a row with no subject (Excel leaves
 * out the variety's description the CSV had: "76 dentro da estrela"), by
 * denomination, year, variety and country; and last by denomination, year
 * and country when a single coin fits (a variety written another way).
 */
export function matchUcoinRows(rows: UcoinRow[], coins: Coin[]): UcoinMatch {
  const remaining = [...coins].sort((a, b) => a.id - b.id)
  const taken = new Set<number>()
  const matched: UcoinMatch['matched'] = []
  const free = (coin: Coin) => !taken.has(coin.id)
  const take = (row: UcoinRow, coin: Coin | undefined) => {
    if (!coin) return false
    taken.add(coin.id)
    matched.push({ row, coin })
    return true
  }
  const sameCountry = (coin: Coin, row: UcoinRow) => fold(coin.country ?? '') === fold(row.country)

  const byName = rows.filter((row) => {
    const key = ucoinCoinKey({ name: ucoinCoinName(row), country: row.country })
    return !take(row, remaining.find((coin) => free(coin) && ucoinCoinKey(coin) === key))
  })
  const byBase = byName.filter((row) => {
    if (row.subject) return true
    const base = fold(`${row.denomination} ${row.yearText}`)
    return !take(row, remaining.find((coin) => free(coin) && sameCountry(coin, row) && nameBase(coin.name) === base))
  })
  const pending = byBase.filter((row) => {
    if (row.subject || row.year === undefined) return true
    const prefix = fold(`${row.denomination} ${row.year}`)
    const fits = remaining.filter((coin) => free(coin) && sameCountry(coin, row) && (nameBase(coin.name) === prefix || nameBase(coin.name).startsWith(`${prefix} `)))
    return !(fits.length === 1 && take(row, fits[0]))
  })
  return { matched, pending }
}

export interface UcoinCoinOptions {
  categories: string[]
  eras: string[]
  /** Catalogues configured in Admin; others stay in the notes. */
  registry: CatalogRegistry[]
  /** US dollars per euro, to store a price given in the other currency. */
  usdPerEur?: number
}

function roundCents(value: number): number {
  return Math.round(value * 100) / 100
}

/** An amount from the file in the collection currency, or undefined when it cannot be converted. */
function inCollectionCurrency(amount: number | undefined, currency: string, usdPerEur: number | undefined): number | undefined {
  if (amount === undefined) return undefined
  const converted = toCollectionCurrency(amount, currency, usdPerEur)
  return converted === undefined ? undefined : roundCents(converted)
}

/** What the file says of how the coin was acquired, and uCoin's price for it. */
export interface UcoinPurchase {
  purchaseDate?: string
  purchasePrice?: number
  purchaseLocation?: string
  currentValue?: number
}

export function ucoinPurchase(row: UcoinRow, usdPerEur?: number): UcoinPurchase {
  return {
    purchaseDate: row.purchaseDate,
    purchasePrice: inCollectionCurrency(row.price, row.priceCurrency, usdPerEur),
    purchaseLocation: row.swapWith ? clip(t('ucoinImport.swapWith', { user: row.swapWith }), 500) : undefined,
    currentValue: inCollectionCurrency(row.value, row.valueCurrency, usdPerEur),
  }
}

/**
 * The coin's own values of the fields updateCoin sends as null when it is not
 * given them (which would clear them), to send along with a change.
 */
export function keepCoinNullables(coin: Coin): Partial<CoinMutationPayload> {
  return {
    weightGrams: coin.weightGrams,
    diameterMm: coin.diameterMm,
    purchasePrice: coin.purchasePrice,
    currentValue: coin.currentValue,
    purchaseDate: coin.purchaseDate,
    storageLocationId: coin.storageLocationId,
    storageSlot: coin.storageSlot ?? null,
    romanImperialFigureId: coin.romanImperialFigureId,
    mintLocationId: coin.mintLocationId,
    year: coin.year ?? null,
  } as Partial<CoinMutationPayload>
}

/**
 * Fills a coin imported earlier with what the file adds (purchase date, price
 * paid, where it came from, a value when it has none); what the coin already
 * has is kept. Null when there is nothing to add.
 */
export function ucoinPurchaseUpdate(coin: Coin, purchase: UcoinPurchase): CoinMutationPayload | null {
  const additions: Partial<CoinMutationPayload> = {}
  if (!coin.purchaseDate && purchase.purchaseDate) additions.purchaseDate = purchase.purchaseDate
  if (coin.purchasePrice == null && purchase.purchasePrice !== undefined) additions.purchasePrice = purchase.purchasePrice
  if (!coin.purchaseLocation?.trim() && purchase.purchaseLocation) additions.purchaseLocation = purchase.purchaseLocation
  if (coin.currentValue == null && purchase.currentValue !== undefined) additions.currentValue = purchase.currentValue
  if (!Object.keys(additions).length) return null
  return { ...keepCoinNullables(coin), ...additions } as CoinMutationPayload
}

export function buildUcoinCoin(row: UcoinRow, options: UcoinCoinOptions): CoinMutationPayload {
  const mintMark = ucoinMintMark(row.variety)
  const purchase = ucoinPurchase(row, options.usdPerEur)
  const price = purchase.purchasePrice
  const registered = row.catalogue
    ? options.registry.find((entry) => entry.catalog.toLowerCase() === row.catalogue!.code.toLowerCase() && !entry.volumeRequired)
    : undefined
  const references: CoinReferenceInput[] = registered && row.catalogue ? [{ catalog: registered.catalog, number: row.catalogue.number }] : []
  const notes = [
    t('ucoinImport.notes.source', { reference: row.reference || '—' }),
    row.period ? t('ucoinImport.notes.period', { value: row.period }) : '',
    row.variety && !mintMark ? t('ucoinImport.notes.variety', { value: row.variety }) : '',
    row.price !== undefined && price === undefined ? t('ucoinImport.notes.price', { value: `${row.price} ${row.priceCurrency}`.trim() }) : '',
    row.slab?.number ? t('ucoinImport.notes.gradingNumber', { value: row.slab.number }) : '',
    row.comment ? t('ucoinImport.notes.comment', { value: row.comment }) : '',
  ].filter(Boolean).join('\n')
  const modern = row.year === undefined || row.year >= 1500
  const category = options.categories.find((option) => option.toLowerCase() === (modern ? 'modern' : 'other'))
    ?? options.categories.find((option) => option.toLowerCase() === 'other')
    ?? 'Other'

  return {
    name: ucoinCoinName(row),
    category: category as Category,
    era: eraForYear(row.year, options.eras),
    material: 'Other',
    denomination: clip(row.denomination, 200) || undefined,
    dateRange: row.year !== undefined ? String(row.year) : clip(row.yearText, 200) || undefined,
    year: row.year ?? null,
    country: clip(row.country, 120) || undefined,
    mint: mintMark || undefined,
    diameterMm: row.diameterMm,
    // A graded coin's grade is the grading company's ("PCGS MS65").
    grade: clip(row.slab?.grade ? `${row.slab.company} ${row.slab.grade}`.trim() : row.grade, 100) || undefined,
    ...(row.slab ? { presentation: 'slab' } : {}),
    purchasePrice: price,
    purchaseDate: purchase.purchaseDate,
    purchaseLocation: purchase.purchaseLocation,
    currentValue: purchase.currentValue,
    notes: clip(notes, 5000),
    isWishlist: false,
    references,
  }
}
