// Finds the Numista type of a coin imported from uCoin. A "KM# 84" style
// number is searched directly; otherwise the issuer's types of that year are
// compared with the coin's denomination and subject (uCoin writes subjects in
// the user's language, Numista titles are mostly English, so words are matched
// loosely: "Turíngia" ≈ "Thuringia", "Berlim" ≈ "Berlin").
import type { NumistaCatalogSearchItem } from '@/api/endpoints/numistaCatalog'
import type { NumistaGradePrice, NumistaIssueSummary } from '@/api/endpoints/coins'
import type { Coin } from '@/types'
import { NUMISTA_GRADES, numistaGradeForCoinGrade, type NumistaGrade } from '@/utils/numistaGrades'
import { UCOIN_SEARCHABLE_CATALOGUES, ucoinMintMark, type UcoinCatalogueCode, type UcoinRow } from '@/utils/ucoinImport'

export interface UcoinLookup {
  country: string
  denomination: string
  year?: number
  mintMark: string
  /** A variety written after the year that is not a mint mark ("76" in a star). */
  variety: string
  subject: string
  catalogue?: { code: UcoinCatalogueCode; number: string }
}

function fold(value: string): string {
  return value.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '').trim()
}

/**
 * The Numista type in a link or number as people paste it: "380096",
 * "https://pt.numista.com/380096", ".../catalogue/pieces380096.html".
 */
export function parseNumistaTypeId(text: string): number | undefined {
  const value = text.trim()
  const match = /^(\d{1,7})$/.exec(value)
    ?? /numista\.com\/(?:catalogue\/pieces)?(\d{1,7})(?:\.html)?(?:[/?#]|$)/i.exec(value)
    ?? /pieces(\d{1,7})\.html/i.exec(value)
  const id = match ? Number(match[1]) : NaN
  return Number.isInteger(id) && id > 0 ? id : undefined
}

/** The Numista type a coin is linked to, if any. */
export function numistaTypeOf(coin: Pick<Coin, 'references'>): number | undefined {
  const reference = coin.references?.find((ref) => ref.catalog.trim().toLowerCase() === 'numista')
  return reference ? parseNumistaTypeId(reference.number) : undefined
}

/**
 * Whether a coin linked to Numista has its details: completing a coin writes
 * the type's page into its notes ("…/catalogue/pieces380096.html"); a link
 * chosen by hand while Numista could not answer has not got them yet.
 */
export function hasNumistaDetails(coin: Pick<Coin, 'notes'>, typeId: number): boolean {
  return new RegExp(`pieces${typeId}\\.html|numista\\.com/${typeId}(?!\\d)`).test(coin.notes ?? '')
}

/** What to look up for an imported coin, read back from the coin itself. */
export function ucoinLookupForCoin(coin: Pick<Coin, 'name' | 'country' | 'denomination' | 'year' | 'mint' | 'references'>): UcoinLookup {
  const country = coin.country?.trim() ?? ''
  const separator = coin.name.indexOf(' – ')
  const detail = separator >= 0 ? coin.name.slice(separator + 3).trim() : ''
  // The name starts with "<denomination> <year> <mint mark or variety>".
  const denomination = coin.denomination?.trim() ?? ''
  const head = (separator >= 0 ? coin.name.slice(0, separator) : coin.name).trim()
  const yearText = denomination && head.startsWith(denomination) ? head.slice(denomination.length).trim() : ''
  const mintMark = ucoinMintMark(coin.mint ?? '')
  const variety = yearText.split(/\s+/).slice(1).filter((token) => token !== mintMark).join(' ')
  const reference = (coin.references ?? []).find((ref) =>
    (UCOIN_SEARCHABLE_CATALOGUES as readonly string[]).includes(ref.catalog.toUpperCase()) && ref.number.trim())
  return {
    country,
    denomination,
    year: coin.year ?? undefined,
    mintMark,
    variety,
    // The name ends with the subject, or with the country when there is none.
    subject: fold(detail) === fold(country) ? '' : detail,
    catalogue: reference ? { code: reference.catalog.toUpperCase() as UcoinCatalogueCode, number: reference.number.trim() } : undefined,
  }
}

/** What to look up for a row of the file, before the coin exists. */
export function ucoinLookupForRow(row: UcoinRow): UcoinLookup {
  const catalogue = row.catalogue && (UCOIN_SEARCHABLE_CATALOGUES as readonly string[]).includes(row.catalogue.code)
    ? { code: row.catalogue.code as UcoinCatalogueCode, number: row.catalogue.number }
    : undefined
  return {
    country: row.country.trim(),
    denomination: row.denomination.trim(),
    year: row.year,
    mintMark: ucoinMintMark(row.variety),
    variety: ucoinMintMark(row.variety) ? '' : row.variety.trim(),
    subject: row.subject.trim(),
    catalogue,
  }
}

/** Coins with the same key are the same Numista type and are looked up once. */
export function ucoinLookupGroupKey(lookup: UcoinLookup): string {
  if (lookup.catalogue) return `${fold(lookup.country)}|${lookup.catalogue.code} ${fold(lookup.catalogue.number)}`
  return `${fold(lookup.country)}|${fold(lookup.denomination)}|${fold(lookup.subject)}|${lookup.year ?? ''}`
}

// ---- Denominations -------------------------------------------------------

const VULGAR_FRACTIONS: Record<string, number> = { '½': 0.5, '¼': 0.25, '¾': 0.75, '⅓': 1 / 3, '⅔': 2 / 3, '⅛': 0.125 }

// Units worth a fraction of the main one, so "2 Euro Cents" never passes for "2 euro".
const SUBUNIT = /\b(cent|cents|centimo|centimos|centime|centimes|centavo|centavos|centesimo|centesimos|centesimi|penny|pence|pfennig|reichspfennig|groschen|grosz|groszy|kopek|kopeks|kopiyok|kopiiok|kopiika|kopiikas|kopiyka|kopiyky|kopecks|halala|halalas|halerov|haleru|heller|hellers|rappen|ore|stotinki|para|fils|agorot|baisa|millimes|qapik|lipa|jiao|avos|bututs|piastres|sen|senti|seniti|santims|sentimo|dime|nickel|mils?)\b/

export interface ParsedDenomination {
  value?: number
  subunit: boolean
}

/** "2½ euro" → 2.5 (main unit); "10 cêntimos de euro" → 10 (subunit). */
export function parseDenomination(text: string): ParsedDenomination {
  const folded = fold(text).replace(/⁄/g, '/')
  let value: number | undefined
  const fraction = /^(\d+)\s*\/\s*(\d+)/.exec(folded)
  const number = /^(\d+(?:[.,]\d+)?)?\s*([½¼¾⅓⅔⅛])?/.exec(folded)
  if (fraction && Number(fraction[2]) > 0) value = Number(fraction[1]) / Number(fraction[2])
  else if (number && (number[1] || number[2])) value = Number((number[1] ?? '0').replace(',', '.')) + (number[2] ? (VULGAR_FRACTIONS[number[2]] ?? 0) : 0)
  return { value, subunit: SUBUNIT.test(folded) }
}

/** The denomination at the start of a Numista title, before "(" or " - ". */
function titleDenomination(title: string): string {
  return title.split(/\s\(|\s-\s/)[0] ?? title
}


function sameDenomination(a: ParsedDenomination, b: ParsedDenomination): boolean {
  return a.value !== undefined && b.value !== undefined && Math.abs(a.value - b.value) < 1e-6 && a.subunit === b.subunit
}

// ---- Loose word matching -------------------------------------------------

const STOP_WORDS = new Set([
  'a', 'o', 'os', 'as', 'de', 'da', 'do', 'das', 'dos', 'e', 'em', 'no', 'na', 'nos', 'nas', 'um', 'uma', 'com', 'por', 'para', 'sem',
  'the', 'of', 'and', 'in', 'on', 'for', 'to', 'with', 'th', 'st', 'nd', 'rd',
  'aniversario', 'anniversary', 'anos', 'ano', 'years', 'year', 'marca', 'euro', 'euros', 'cent', 'cents',
])

// Portuguese words for the metals and finishes Numista titles name in English.
const SYNONYMS: Record<string, string> = { prata: 'silver', ouro: 'gold', prova: 'proof', cuproniquel: 'cupronickel', cobre: 'copper' }

function words(text: string): string[] {
  return fold(text)
    // "75th" and "85.º" are the number 75 and 85.
    .replace(/(\d)([a-z])/g, '$1 $2')
    .split(/[^a-z0-9]+/)
    .filter((word) => (/^\d+$/.test(word) ? word.length >= 2 : word.length >= 3) && !STOP_WORDS.has(word))
    .map((word) => SYNONYMS[word] ?? word)
}

function editDistance(a: string, b: string): number {
  // One row of the classic dynamic-programming table, reused.
  const row = Array.from({ length: b.length + 1 }, (_, i) => i)
  for (let i = 1; i <= a.length; i += 1) {
    let diagonal = i - 1
    row[0] = i
    for (let j = 1; j <= b.length; j += 1) {
      const above = row[j] ?? 0
      row[j] = Math.min(above + 1, (row[j - 1] ?? 0) + 1, diagonal + (a[i - 1] === b[j - 1] ? 0 : 1))
      diagonal = above
    }
  }
  return row[b.length] ?? 0
}

// Endings that differ between languages for the same name: Turíngia /
// Thüringen, Saxónia / Saxony, Baviera / Bavaria.
const ENDINGS = ['ia', 'en', 'um', 'us', 'os', 'as', 'es', 'a', 'o', 'e', 'y', 's']

function stem(word: string): string {
  const ending = ENDINGS.find((suffix) => word.length - suffix.length >= 4 && word.endsWith(suffix))
  return ending ? word.slice(0, -ending.length) : word
}

function closeEnough(a: string, b: string): boolean {
  const shorter = Math.min(a.length, b.length)
  const allowed = shorter >= 8 ? 2 : shorter >= 5 ? 1 : 0
  return allowed > 0 && editDistance(a, b) <= allowed
}

function similarWords(a: string, b: string): boolean {
  if (a === b) return true
  if (/^\d+$/.test(a) || /^\d+$/.test(b)) return false
  // A long word inside another, not at its start: "Dirndl" in "Aufdirndln"
  // (a shared start is weighed below: "Willem" is not "Willemstad").
  const [shorter, longer] = a.length <= b.length ? [a, b] : [b, a]
  if (shorter.length >= 6 && longer.indexOf(shorter) > 0) return true
  let prefix = 0
  while (prefix < a.length && prefix < b.length && a[prefix] === b[prefix]) prefix += 1
  // A shared start counts only between words of similar length: "Lisboa" is
  // "Lisbon", but "Brandt" is not "Brandenburg".
  if (prefix >= 5 && Math.abs(a.length - b.length) <= 3) return true
  return closeEnough(a, b) || closeEnough(stem(a), stem(b))
}

/**
 * How many of the subject's words appear, loosely, in the title. Words of the
 * coin's country are left out: "França 2016" says nothing that tells "EURO
 * 2016" from "François Mitterrand".
 */
export function subjectScore(subject: string, title: string, country = ''): number {
  const titleWords = words(title)
  const countryWords = words(country)
  return words(subject)
    .filter((word) => !countryWords.some((own) => similarWords(word, own)))
    .filter((word) => titleWords.some((candidate) => similarWords(word, candidate))).length
}

/** How well a Numista issuer name fits the country written in uCoin (0 = not at all). */
export function issuerSimilarity(country: string, issuerName: string): number {
  const wanted = fold(country)
  const name = fold(issuerName)
  if (!wanted || !name) return 0
  if (wanted === name) return 3
  const head = (name.split(/[,(]/)[0] ?? '').trim()
  if (head === wanted) return 2
  if (head && (wanted.startsWith(head) || head.startsWith(wanted))) return 1.5
  const nameWords = words(name)
  return words(wanted).some((word) => nameWords.some((candidate) => similarWords(word, candidate))) ? 1 : 0
}

// ---- Choosing the type, issue and price ---------------------------------

// Collector versions of a coin (silver, gold, proof, pattern...). A coin from
// uCoin is taken to be the ordinary one unless its subject says otherwise.
const SPECIAL_VERSION = /\b(silver|gold|platinum|palladium|proof|piedfort|pattern|essai|trial|prova|prata|ouro|colou?red)\b/

function specialVersionPenalty(item: NumistaCatalogSearchItem, lookup: UcoinLookup): number {
  return SPECIAL_VERSION.test(fold(item.title)) && !SPECIAL_VERSION.test(fold(lookup.subject)) ? 0.5 : 0
}

// English words Numista titles use for units written in Portuguese, to narrow
// an issuer's list of types; checked in order ("cêntimos de euro" is euro).
const UNIT_QUERIES: [RegExp, string][] = [
  [/\beuro/, 'euro'],
  [/\blibra/, 'pound'],
  [/\bdolar/, 'dollar'],
  [/\bxelim|\bshilling/, 'shilling'],
  [/\bfranco/, 'franc'],
  [/\bmarco/, 'mark'],
  [/\bpeseta/, 'peseta'],
  [/\bescudo/, 'escudo'],
  [/\blira/, 'lira'],
  [/\biene|\byen\b/, 'yen'],
  [/\bpeso/, 'peso'],
  [/\brand\b/, 'rand'],
]

/** A word to search an issuer's types with, for the coin's unit, if known. */
export function unitQuery(denomination: string): string | undefined {
  const folded = fold(denomination)
  return UNIT_QUERIES.find(([pattern]) => pattern.test(folded))?.[1]
}

function yearFits(item: NumistaCatalogSearchItem, year: number | undefined): boolean {
  if (year === undefined) return true
  if (item.minYear !== undefined && year < item.minYear) return false
  if (item.maxYear !== undefined && year > item.maxYear) return false
  return true
}

/** True when the subject has words to compare ("sem marca" has none). */
function hasSubject(lookup: UcoinLookup): boolean {
  return words(lookup.subject).length > 0
}

/**
 * How many of the subject's words, numbers left out, the title has: "100°
 * aniversário" shares "100" with every "100 Escudos" title, which names nothing.
 */
function namedSubjectScore(lookup: UcoinLookup, title: string): number {
  return subjectScore(lookup.subject.replace(/\d+/g, ' '), title, lookup.country)
}

export interface PickNumistaTypeOptions {
  /**
   * Only a type the subject names (by a word, not a number) will do: for
   * lists that hold more than the coin's own country, such as Portugal's
   * types searched for a coin of the Azores.
   */
  requireSubject?: boolean
}

function span(item: NumistaCatalogSearchItem): number {
  return item.minYear !== undefined && item.maxYear !== undefined ? item.maxYear - item.minYear : 0
}

/** Best of several equally scored candidates, or none when it is a tie. */
function unique<T>(scored: { item: T; score: number }[]): T | undefined {
  const best = Math.max(...scored.map((entry) => entry.score))
  const top = scored.filter((entry) => entry.score === best)
  return top.length === 1 ? top[0]?.item : undefined
}

// Words naming a version of a design rather than the design itself: "5 Euros
// (Jaap Eden; silver plated copper)" and "(Jaap Eden; silver)" are one design.
const VERSION_WORDS = new Set([
  'silver', 'gold', 'plated', 'copper', 'bronze', 'steel', 'nickel', 'cupronickel', 'brass', 'magnetic', 'non',
  'coloured', 'colored', 'colourised', 'colorized', 'proof', 'piedfort', 'brilliant', 'uncirculated', 'normal', 'obverse',
])

function designKey(title: string): string {
  return words(title).filter((word) => !VERSION_WORDS.has(word) && !/^\d+$/.test(word)).join(' ')
}

/**
 * Like unique, but a tie between versions of one design (metals, finishes,
 * colours) goes to its ordinary version, the lowest-numbered when several are.
 */
function uniqueDesign(scored: { item: NumistaCatalogSearchItem; score: number }[], lookup: UcoinLookup): NumistaCatalogSearchItem | undefined {
  const best = Math.max(...scored.map((entry) => entry.score))
  const top = scored.filter((entry) => entry.score === best).map((entry) => entry.item)
  return ordinaryVersion(top, lookup)
}

/** The ordinary version when the items are versions of one design, else undefined (unless there is one). */
function ordinaryVersion(items: NumistaCatalogSearchItem[], lookup: UcoinLookup): NumistaCatalogSearchItem | undefined {
  if (items.length <= 1) return items[0]
  const key = designKey(items[0]?.title ?? '')
  if (!key || items.some((item) => designKey(item.title) !== key)) return undefined
  return [...items].sort((a, b) => specialVersionPenalty(a, lookup) - specialVersionPenalty(b, lookup) || a.id - b.id)[0]
}

/**
 * Picks the coin's type among search results, or undefined when no result is
 * a clear match. Results of a catalogue-number search are already precise;
 * results of an issuer search are narrowed by denomination and subject.
 */
export function pickNumistaType(
  candidates: NumistaCatalogSearchItem[],
  lookup: UcoinLookup,
  source: 'catalogue' | 'issuer',
  options: PickNumistaTypeOptions = {},
): NumistaCatalogSearchItem | undefined {
  const denomination = parseDenomination(lookup.denomination)

  if (source === 'catalogue') {
    // Catalogue numbers repeat across countries (KM# 412 is a Norwegian and an
    // Egyptian coin), so the issuer must fit the country. The year only breaks
    // ties: uCoin and Numista sometimes date the same type differently (a
    // Spanish 1975 peseta struck in 1976, a commemorative dated by its event).
    const fitting = candidates.filter((item) => issuerSimilarity(lookup.country, item.issuer.name) > 0)
    if (fitting.length <= 1) return fitting[0]
    const scored = fitting.map((item) => ({
      item,
      score: issuerSimilarity(lookup.country, item.issuer.name) * 10
        + (sameDenomination(denomination, parseDenomination(titleDenomination(item.title))) ? 5 : 0)
        + (yearFits(item, lookup.year) ? 2 : 0)
        + subjectScore(lookup.subject, item.title, lookup.country)
        - specialVersionPenalty(item, lookup),
    }))
    return uniqueDesign(scored, lookup)
  }

  const pool = candidates.filter((item) => yearFits(item, lookup.year))
  const sameValue = pool.filter((item) => sameDenomination(denomination, parseDenomination(titleDenomination(item.title))))
  if (!sameValue.length) return undefined
  if (options.requireSubject && !hasSubject(lookup)) return undefined
  if (hasSubject(lookup)) {
    const matches = sameValue.map((item) => ({ item, matched: subjectScore(lookup.subject, item.title, lookup.country) }))
    const match = uniqueDesign(matches.map(({ item, matched }) => ({ item, score: matched - specialVersionPenalty(item, lookup) })), lookup)
    if (match && matches.some((entry) => entry.item === match && entry.matched > 0)) {
      return !options.requireSubject || namedSubjectScore(lookup, match.title) > 0 ? match : undefined
    }
    return !options.requireSubject && sameValue.length === 1 ? sameValue[0] : undefined
  }
  // No subject: the regular circulation type, struck over several years,
  // rather than a one-year commemorative or a collector version.
  if (sameValue.length === 1) return sameValue[0]
  return unique(sameValue.map((item) => ({ item, score: span(item) - specialVersionPenalty(item, lookup) * 1000 })))
}

/**
 * The types a person can choose from for a coin the import did not find:
 * those of its value that fit its year, or every type of that year when none
 * has its value; each once.
 */
export function numistaCandidates(candidates: NumistaCatalogSearchItem[], lookup: UcoinLookup): NumistaCatalogSearchItem[] {
  const denomination = parseDenomination(lookup.denomination)
  const pool = candidates.filter((item) => yearFits(item, lookup.year))
  const sameValue = pool.filter((item) => sameDenomination(denomination, parseDenomination(titleDenomination(item.title))))
  return [...new Map((sameValue.length ? sameValue : pool).map((item) => [item.id, item])).values()]
}

/** The issue of the coin's year (and mint mark or variety), when Numista lists it. */
export function pickIssue(issues: NumistaIssueSummary[], lookup: Pick<UcoinLookup, 'year' | 'mintMark' | 'variety'>): NumistaIssueSummary | undefined {
  if (lookup.year === undefined) return issues.length === 1 ? issues[0] : undefined
  const year = lookup.year
  const sameYear = issues.filter((issue) => issue.year === year
    || (issue.year === undefined && issue.minYear !== undefined && issue.maxYear !== undefined && issue.minYear <= year && year <= issue.maxYear))
  if (!sameYear.length) return undefined
  if (lookup.mintMark) {
    const mark = fold(lookup.mintMark)
    const withMark = sameYear.find((issue) => fold(issue.mintLetter ?? '') === mark)
    if (withMark) return withMark
  }
  if (lookup.variety) {
    const variety = fold(lookup.variety)
    const withVariety = sameYear.find((issue) => fold(`${issue.comment ?? ''} ${issue.mintLetter ?? ''}`).includes(variety))
    if (withVariety) return withVariety
  }
  return sameYear.find((issue) => !issue.mintLetter && !issue.comment) ?? sameYear.find((issue) => !issue.mintLetter) ?? sameYear[0]
}

/** Numista grade used to value a coin: its own grade, proof as UNC, else VF. */
export function valuationGrade(grade: string | null | undefined): NumistaGrade {
  const mapped = numistaGradeForCoinGrade(grade)
  if (mapped) return mapped
  return /\b(PRF|PROOF|PP|BE)\b/i.test(grade ?? '') ? 'unc' : 'vf'
}

/** Price for the grade, else the nearest grade Numista has a price for (lower first). */
export function priceForGrade(prices: NumistaGradePrice[], grade: NumistaGrade): number | undefined {
  const byGrade = new Map(prices.filter((price) => price.price > 0).map((price) => [price.grade, price.price]))
  const index = NUMISTA_GRADES.indexOf(grade)
  for (let distance = 0; distance < NUMISTA_GRADES.length; distance += 1) {
    for (const candidate of [NUMISTA_GRADES[index - distance], NUMISTA_GRADES[index + distance]]) {
      if (candidate && byGrade.has(candidate)) return byGrade.get(candidate)
    }
  }
  return undefined
}
