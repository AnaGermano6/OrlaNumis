// Reads a collection exported from uCoin.net (CSV) and turns each row into a
// coin for this app. Only what the file says is used here; catalogue details
// (metal, weight, rarity, value) are added afterwards from Numista.
import { t } from '@/i18n'
import type { CatalogRegistry, Category, Coin, CoinMutationPayload, CoinReferenceInput } from '@/types'
import { toCollectionCurrency } from '@/utils/collectionCurrency'
import { eraForYear } from '@/utils/numistaCatalogPrefill'

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
  price?: number
  /** Currency of the price column ("Preço, EUR" → EUR). */
  priceCurrency: string
  /** Catalogue number as written: "KM# 84", "UC# 113". */
  reference: string
  catalogue?: { code: string; number: string }
  comment: string
}

export type UcoinParseError = 'empty' | 'columns'

export interface UcoinParseResult {
  rows: UcoinRow[]
  error?: UcoinParseError
}

type Column = 'country' | 'period' | 'denomination' | 'year' | 'variety' | 'subject' | 'diameter' | 'grade' | 'price' | 'number' | 'comment'

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
  price: ['preco', 'price'],
  number: ['numero', 'number', 'catalog number'],
  comment: ['comentario', 'comment'],
}

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

export function parseUcoinCsv(text: string): UcoinParseResult {
  const [headerCells, ...dataRows] = parseCsv(text)
  if (!headerCells || !dataRows.length) return { rows: [], error: 'empty' }
  const header = headerCells.map(fold)
  const columns = new Map<Column, number>()
  let priceCurrency = ''
  for (const [column, names] of Object.entries(HEADER_NAMES) as [Column, string[]][]) {
    const index = header.findIndex((name) => names.includes(name) || (column === 'price' && names.some((n) => name.startsWith(`${n},`))))
    if (index >= 0) columns.set(column, index)
  }
  const priceIndex = columns.get('price')
  if (priceIndex !== undefined) priceCurrency = (headerCells[priceIndex]?.split(',')[1] ?? '').trim().toUpperCase()
  if (!columns.has('country') || !columns.has('denomination') || !columns.has('year')) return { rows: [], error: 'columns' }

  const cell = (fields: string[], column: Column) => {
    const index = columns.get(column)
    return index === undefined ? '' : (fields[index] ?? '').trim()
  }
  const rows: UcoinRow[] = dataRows.map((fields, index) => {
    const yearText = cell(fields, 'year')
    const yearMatch = /^-?\d{1,4}/.exec(yearText)
    const reference = cell(fields, 'number')
    return {
      line: index + 2,
      country: cell(fields, 'country'),
      period: cell(fields, 'period'),
      denomination: cell(fields, 'denomination'),
      yearText,
      year: yearMatch ? Number(yearMatch[0]) : undefined,
      variety: cell(fields, 'variety'),
      subject: cell(fields, 'subject'),
      diameterMm: parseNumber(cell(fields, 'diameter')),
      grade: cell(fields, 'grade'),
      price: parseNumber(cell(fields, 'price')),
      priceCurrency,
      reference,
      catalogue: parseCatalogue(reference),
      comment: cell(fields, 'comment'),
    }
  })
  return { rows: rows.filter((row) => row.country || row.denomination) }
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

/** The row's price in the collection currency, or undefined when it cannot be converted. */
function priceInCollectionCurrency(row: UcoinRow, usdPerEur: number | undefined): number | undefined {
  if (row.price === undefined) return undefined
  const price = toCollectionCurrency(row.price, row.priceCurrency, usdPerEur)
  return price === undefined ? undefined : roundCents(price)
}

export function buildUcoinCoin(row: UcoinRow, options: UcoinCoinOptions): CoinMutationPayload {
  const mintMark = ucoinMintMark(row.variety)
  const price = priceInCollectionCurrency(row, options.usdPerEur)
  const registered = row.catalogue
    ? options.registry.find((entry) => entry.catalog.toLowerCase() === row.catalogue!.code.toLowerCase() && !entry.volumeRequired)
    : undefined
  const references: CoinReferenceInput[] = registered && row.catalogue ? [{ catalog: registered.catalog, number: row.catalogue.number }] : []
  const notes = [
    t('ucoinImport.notes.source', { reference: row.reference || '—' }),
    row.period ? t('ucoinImport.notes.period', { value: row.period }) : '',
    row.variety && !mintMark ? t('ucoinImport.notes.variety', { value: row.variety }) : '',
    row.price !== undefined && price === undefined ? t('ucoinImport.notes.price', { value: `${row.price} ${row.priceCurrency}`.trim() }) : '',
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
    grade: clip(row.grade, 100) || undefined,
    purchasePrice: price,
    notes: clip(notes, 5000),
    isWishlist: false,
    references,
  }
}
