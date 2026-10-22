// Turns a coin type picked in the Numista catalogue into Add Coin form values.
// Only catalogue facts are filled in; what belongs to the user's own coin
// (price, date, place, storage) stays for them to complete.
import { intlLocale, t } from '@/i18n'
import type { NumistaCatalogType, NumistaIssueSummary } from '@/api/client'
import type { CatalogRegistry, Category, Coin, CoinReferenceInput, Material } from '@/types'
import { formatNumistaYear, numistaYearRange } from '@/utils/numistaCatalog'
import { numistaGradeShort, type NumistaGrade } from '@/utils/numistaGrades'
import type { RealPhotoChoice } from '@/utils/coinPhotos'
import { rarityFromMintage } from '@/utils/coinAttributes'

export interface CatalogSelection {
  type: NumistaCatalogType
  issue?: NumistaIssueSummary
  grade?: NumistaGrade
  /** Numista's price for the chosen grade, used as the coin's current value. */
  currentValue?: number
  wishlist: boolean
  /** Real photos to import into the coin once it is saved. */
  photos?: RealPhotoChoice
}

export interface CatalogPrefillOptions {
  categories: string[]
  eras: string[]
  materials: string[]
  /** Catalogues configured in Admin; other references go into the notes. */
  registry: CatalogRegistry[]
}

export interface CatalogPrefill {
  fields: Partial<Coin>
  references: CoinReferenceInput[]
}

function fold(value: string): string {
  return value.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '')
}

function clip(value: string, max: number): string {
  return value.length > max ? `${value.slice(0, max - 1)}…` : value
}

function pickConfigured(options: string[], wanted: string, fallback: string): string {
  const match = options.find((option) => option.toLowerCase() === wanted.toLowerCase())
  if (match) return match
  return options.find((option) => option.toLowerCase() === fallback.toLowerCase()) ?? options[0] ?? fallback
}

// Alloys come before the metals they are named after ("ouro nórdico" is not
// gold, "latão de níquel" is not nickel). The same rules as the API's
// MaterialFromComposition, which fills existing coins.
const COMPOSITION_MATERIALS: [Material, RegExp][] = [
  ['Nordic gold', /ouro nordico|nordic gold/],
  ['Aluminium bronze', /alumini?(?:o|um)[- ](?:(?:niquel|nickel)[- ])?bronze|bronze[- ]alumini?(?:o|um)|aluminum[- ]bronze|bronzital/],
  ['Stainless steel', /inoxidavel|stainless|acmonital/],
  ['Nickel brass', /latao de niquel|nickel[- ]brass|nickel silver|german silver|alpaca|maillechort|argentan/],
  ['Copper-nickel', /cupro[- ]?(?:niquel|nickel)|(?:copper|cobre)[- ](?:nickel|niquel)/],
  ['Billon', /\bbilhao\b|\bbillon\b/],
  ['Gold', /\bouro\b|\bgold\b/],
  ['Silver', /\bprata\b|\bsilver\b/],
  ['Electrum', /electrum|\beletro\b|\belectro\b/],
  ['Bronze', /bronze/],
  ['Brass', /\blatao\b|\bbrass\b|\btombac/],
  ['Nickel', /\bniquel\b|\bnickel\b/],
  ['Aluminium', /\balumini?(?:o|um)\b|\baluminum\b/],
  ['Zinc', /\bzinco?\b/],
  ['Copper', /\bcobre\b|\bcopper\b/],
]

function decodeEntities(value: string): string {
  return value.replace(/&amp;/g, '&').replace(/&nbsp;/g, ' ')
}

/**
 * Maps a Numista composition ("Silver (.900)", "Cuproníquel", "Aço banhado
 * em cobre", "Bimetálica: …") to an app material. A plated or clad coin
 * counts by its core, a plated steel one as plated steel.
 */
export function materialFromComposition(composition: string | undefined, materials: string[]): string {
  return pickConfigured(materials, compositionMaterial(composition ?? ''), 'Other')
}

function compositionMaterial(composition: string): Material {
  let text = fold(decodeEntities(composition)).split(/\s+/).filter(Boolean).join(' ')
  // The alloy in brackets names other metals ("Cuproníquel (75% copper").
  const bracket = text.indexOf('(')
  if (bracket > 0 && text.slice(0, bracket).trim()) text = text.slice(0, bracket).trim()
  if (/\b(?:bi|tri)-?metal/.test(text)) return 'Bimetallic'
  const core = /^(.+?) (?:banhad[oa]s?|revestid[oa]s?|chapead[oa]s?)\b/.exec(text)?.[1]
    ?? /^.+?[- ](?:plated|clad|coated) (.+)$/.exec(text)?.[1]
  if (core && /\baco\b|\bsteel\b|\bferro\b|\biron\b/.test(core) && !/inox|stainless/.test(core)) return 'Plated steel'
  const subject = core ?? text
  return COMPOSITION_MATERIALS.find(([, pattern]) => pattern.test(subject))?.[0] ?? 'Other'
}

/** Suggests one of the configured categories from the dates: Modern from 1500. */
export function categoryForType(type: NumistaCatalogType, categories: string[]): string {
  const year = type.maxYear ?? type.minYear
  const category = year !== undefined && year >= 1500 ? 'Modern' : 'Other'
  return pickConfigured(categories, category, 'Other')
}

/** Suggests a configured era from a year, or '' (unspecified) when none fits. */
export function eraForYear(year: number | undefined, eras: string[]): string {
  if (year === undefined) return ''
  const era = year < 500 ? 'ancient' : year < 1500 ? 'medieval' : 'modern'
  return eras.find((option) => option.toLowerCase() === era) ?? ''
}

function lettering(value: string | undefined): string {
  return (value ?? '').split('\n').map((line) => line.trim()).filter(Boolean).join(' / ')
}

function decimal(value: number): string {
  return new Intl.NumberFormat(intlLocale(), { maximumFractionDigits: 2 }).format(value)
}

export function buildCatalogPrefill(selection: CatalogSelection, options: CatalogPrefillOptions): CatalogPrefill {
  const { type, issue } = selection
  const issueYear = issue?.year
  const year = issueYear ?? type.maxYear ?? type.minYear
  const dateRange = issueYear !== undefined
    ? formatNumistaYear(issueYear)
    : issue && (issue.minYear !== undefined || issue.maxYear !== undefined)
      ? numistaYearRange(issue.minYear, issue.maxYear)
      : numistaYearRange(type.minYear, type.maxYear)

  const references: CoinReferenceInput[] = [{ catalog: 'Numista', number: String(type.id), uri: type.url }]
  const otherReferences: string[] = []
  for (const reference of type.references) {
    const registered = options.registry.find((entry) => entry.catalog.toLowerCase() === reference.catalogue.toLowerCase())
    if (registered && !registered.volumeRequired && registered.catalog.toLowerCase() !== 'numista') {
      references.push({ catalog: registered.catalog, number: reference.number })
    } else {
      otherReferences.push(`${reference.catalogue} ${reference.number}`)
    }
  }

  const notes = [
    t('catalog.notes.source', { title: type.title, issuer: type.issuer.name, url: type.url }),
    type.composition ? t('catalog.notes.composition', { value: type.composition }) : '',
    type.thicknessMm ? t('catalog.notes.thickness', { value: decimal(type.thicknessMm) }) : '',
    type.shape ? t('catalog.notes.shape', { value: type.shape }) : '',
    type.edge.description ? t('catalog.notes.edge', { value: type.edge.description }) : '',
    issue?.mintage ? t('catalog.notes.mintage', { value: new Intl.NumberFormat(intlLocale()).format(issue.mintage) }) : '',
    otherReferences.length ? t('catalog.notes.otherReferences', { value: otherReferences.join(', ') }) : '',
  ].filter(Boolean).join('\n')

  const nameParts = [type.title, issueYear !== undefined ? formatNumistaYear(issueYear) : ''].filter(Boolean).join(' ')
  const mint = [type.mints.join(', '), issue?.mintLetter ? `(${issue.mintLetter})` : ''].filter(Boolean).join(' ')

  const fields: Partial<Coin> = {
    name: clip(type.issuer.name ? `${nameParts} – ${type.issuer.name}` : nameParts, 200),
    category: categoryForType(type, options.categories) as Category,
    era: eraForYear(year, options.eras),
    denomination: clip(type.denomination ?? '', 200),
    ruler: clip(type.rulers.join(', '), 200),
    dateRange: clip(dateRange, 200),
    // The issue's year, or the type's when it was struck in a single year.
    year: issueYear ?? (type.minYear !== undefined && type.minYear === type.maxYear ? type.minYear : null),
    country: clip(type.issuer.name, 120),
    monetarySystem: clip(type.monetarySystem ?? '', 120),
    // A starting point from the mintage; the user can change it.
    rarity: rarityFromMintage(issue?.mintage) ?? '',
    // Numista says the type is issued in a coin card.
    ...(type.issuedInCoincard ? { presentation: 'coincard' } : {}),
    mint: clip(mint, 200),
    material: materialFromComposition(type.composition, options.materials) as Material,
    weightGrams: type.weightGrams,
    diameterMm: type.diameterMm,
    grade: selection.grade ? numistaGradeShort(selection.grade) : '',
    obverseInscription: clip(lettering(type.obverse.lettering), 1000),
    reverseInscription: clip(lettering(type.reverse.lettering), 1000),
    obverseDescription: clip(type.obverse.description ?? '', 2000),
    reverseDescription: clip(type.reverse.description ?? '', 2000),
    // Numista estimates carry many decimals; the coin keeps cents.
    currentValue: selection.currentValue === undefined ? undefined : Math.round(selection.currentValue * 100) / 100,
    notes: clip(notes, 5000),
    isWishlist: selection.wishlist,
  }
  return { fields, references }
}
