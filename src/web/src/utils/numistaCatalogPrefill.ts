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

/** Maps a Numista composition ("Silver (.900)", "Cuproníquel", "Ouro") to an app material. */
export function materialFromComposition(composition: string | undefined, materials: string[]): string {
  const text = fold(composition ?? '')
  let material = 'Other'
  if (/nordic gold|ouro nordico/.test(text)) material = 'Other'
  else if (/\bgold\b|\bouro\b/.test(text)) material = 'Gold'
  else if (/silver|prata/.test(text)) material = 'Silver'
  else if (/electrum|eletro|electro/.test(text)) material = 'Electrum'
  else if (/bronze/.test(text)) material = 'Bronze'
  else if (/^(copper|cobre)\b/.test(text) && !/nickel|niquel|zinc|brass|latao|alumin/.test(text)) material = 'Copper'
  return pickConfigured(materials, material, 'Other')
}

/** Suggests one of the configured categories from the issuer, title and dates. */
export function categoryForType(type: NumistaCatalogType, categories: string[]): string {
  const text = fold(`${type.issuer.name} ${type.issuer.code} ${type.title}`)
  const year = type.maxYear ?? type.minYear
  let category = 'Other'
  if (/byzant|bizant/.test(text)) category = 'Byzantine'
  else if (/roman|romano|_rom_|\brome\b|\broma\b/.test(text) && (year === undefined || year < 500)) category = 'Roman'
  else if (/greek|greece|grecia|grega|greco|hellen|helen/.test(text) && (year === undefined || year < 100)) category = 'Greek'
  else if (year !== undefined && year >= 1500) category = 'Modern'
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
