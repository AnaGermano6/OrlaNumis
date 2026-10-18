// Turns a new coin (from the ECB or a mint's shop) into Add Coin form values,
// named the way the collection names its coins: "2 euro 2025 – Feature".
// What belongs to the user's own coin (price, date, grade, storage) stays for
// them to fill in; the official images are added when it is saved.
import type { RouteLocationRaw } from 'vue-router'
import type { CoinRelease } from '@/api/client'
import type { CatalogPrefill } from '@/utils/numistaCatalogPrefill'
import type { Coin, CoinMutationPayload, CoinReferenceInput } from '@/types'
import { intlLocale, t } from '@/i18n'
import { rarityFromMintage } from '@/utils/coinAttributes'
import { JOINT_ISSUE, releaseCountryName } from '@/utils/coinReleases'
import { materialFromComposition } from '@/utils/numistaCatalogPrefill'

export interface ReleasePrefillOptions {
  categories: string[]
  materials: string[]
}

export interface ReleasePrefill {
  fields: Partial<Coin>
  references: CoinReferenceInput[]
}

const SOURCES = ['ecb', 'incm', 'mdp', 'moe', 'numista'] as const

const SOURCE_NAMES: Record<CoinRelease['source'], string> = {
  ecb: '',
  incm: 'INCM',
  mdp: 'Monnaie de Paris',
  moe: 'Münze Österreich',
  numista: 'Numista',
}

/** Where a release comes from: "BCE", "INCM"… */
export function releaseSourceName(source: CoinRelease['source']): string {
  return source === 'ecb' ? t('addCoin.release.ecb') : SOURCE_NAMES[source]
}

/** A face value as the collection writes it: "2 euro", "7½ euro", "200 escudos", "1 escudo". */
export function releaseDenomination(value: number, currency?: string): string {
  const whole = Math.trunc(value)
  const number = Math.abs(value - whole - 0.5) < 1e-9
    ? `${whole || ''}½`
    : new Intl.NumberFormat(intlLocale(), { maximumFractionDigits: 2, useGrouping: false }).format(value)
  if (currency === 'PTE') return `${number} ${value === 1 ? 'escudo' : 'escudos'}`
  return `${number} euro`
}

// Whether the coin circulated: a €2 commemorative is current, an escudo
// commemorative demonetized; a collector coin never circulated.
function circulation(release: CoinRelease): string {
  if (release.kind === 'collector') return 'non_circulating'
  return release.currency === 'PTE' ? 'demonetized_commemorative' : 'current_commemorative'
}

function pick(options: string[], wanted: string): string {
  return options.find((option) => option.toLowerCase() === wanted.toLowerCase())
    ?? options.find((option) => option.toLowerCase() === 'other')
    ?? wanted
}

function clip(value: string, max: number): string {
  return value.length > max ? `${value.slice(0, max - 1)}…` : value
}

/** The form values of a release, and its Numista reference when it comes from Numista. */
export function buildReleasePrefill(release: CoinRelease, options: ReleasePrefillOptions): ReleasePrefill {
  const year = Number.parseInt(release.from.slice(0, 4), 10) || null
  const denomination = releaseDenomination(release.faceValue, release.currency)
  // A joint issue's country is the one the user's coin is of: theirs to choose.
  const country = release.country && release.country !== JOINT_ISSUE ? releaseCountryName(release.country) : ''
  const notes = [
    release.variants?.length ? t('addCoin.release.variants', { value: release.variants.join(', ') }) : '',
    release.volume ? t('catalog.notes.mintage', { value: new Intl.NumberFormat(intlLocale()).format(release.volume) }) : '',
  ].filter(Boolean).join('\n')
  const numista = release.source === 'numista' ? /pieces(\d+)\.html/.exec(release.pageUrl)?.[1] : undefined
  return {
    fields: {
      name: clip([denomination, year ?? ''].filter(Boolean).join(' ') + ` – ${release.title}`, 200),
      category: pick(options.categories, 'Modern') as Coin['category'],
      // A metal only when the design comes in one ("Prata Proof").
      material: materialFromComposition(release.variants?.length === 1 ? release.variants[0] : '', options.materials) as Coin['material'],
      denomination,
      year,
      country,
      circulation: circulation(release),
      rarity: rarityFromMintage(release.volume) ?? '',
      referenceUrl: release.pageUrl,
      referenceText: releaseSourceName(release.source),
      notes,
      quantity: 1,
      // A coin still to come is wanted, not had.
      isWishlist: release.upcoming,
    },
    references: numista ? [{ catalog: 'Numista', number: numista, uri: release.pageUrl }] : [],
  }
}

// A shop's product page is the coin's own, where it is bought. The ECB's
// year page lists every coin of the year, and a wishlist coin's source page
// is one of a kind (and checked for availability): those go in the notes.
const SHOPS: Array<CoinRelease['source']> = ['incm', 'mdp', 'moe']

/** The wishlist coin of a release, as saved at once: its source page in the notes unless it is a shop's. */
export function buildReleaseWishlistCoin(release: CoinRelease, options: ReleasePrefillOptions): CoinMutationPayload {
  const { fields, references } = buildReleasePrefill(release, options)
  const shop = SHOPS.includes(release.source)
  const source = !shop && release.pageUrl ? t('addCoin.release.source', { source: releaseSourceName(release.source), url: release.pageUrl }) : ''
  return {
    name: fields.name,
    category: fields.category,
    material: fields.material,
    denomination: fields.denomination,
    year: fields.year,
    country: fields.country || undefined,
    circulation: fields.circulation,
    rarity: fields.rarity || undefined,
    referenceUrl: shop ? release.pageUrl : undefined,
    referenceText: shop ? fields.referenceText : undefined,
    notes: [fields.notes, source].filter(Boolean).join('\n') || undefined,
    quantity: 1,
    isWishlist: true,
    references: references.length ? references : undefined,
  }
}

// A catalogue's references, once each.
function mergeReferences(...lists: CoinReferenceInput[][]): CoinReferenceInput[] {
  const seen = new Set<string>()
  return lists.flat().filter((reference) => {
    const key = `${reference.catalog.toLowerCase()}|${reference.number}`
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}

/**
 * A release's form values completed with its Numista type's (see
 * buildCatalogPrefill): what the coin is made of, weighs and measures, its
 * mint, ruler, sides, catalogue references and Numista's notes. The release
 * keeps what the collection writes its own way (its name, face value,
 * country, year, circulation and source page), and the metal and rarity it
 * says itself.
 */
export function completeReleasePrefill(release: CoinRelease, numista: CatalogPrefill, options: ReleasePrefillOptions): ReleasePrefill {
  const own = buildReleasePrefill(release, options)
  const found = numista.fields
  const variantMetal = release.variants?.length === 1 ? materialFromComposition(release.variants[0], options.materials) : ''
  const metal = variantMetal && variantMetal !== materialFromComposition('', options.materials) ? variantMetal : ''
  // The release's mintage, when it says one, rather than Numista's.
  const mintage = t('catalog.notes.mintage', { value: '' }).trim()
  const numistaNotes = (found.notes ?? '').split('\n').filter((line) => line && !(release.volume && line.startsWith(mintage)))
  const notes = [...new Set([...(own.fields.notes ?? '').split('\n'), ...numistaNotes].filter(Boolean))].join('\n')
  return {
    fields: {
      ...own.fields,
      era: found.era || own.fields.era,
      ruler: found.ruler || own.fields.ruler,
      mint: found.mint || own.fields.mint,
      material: (metal || found.material || own.fields.material) as Coin['material'],
      weightGrams: found.weightGrams ?? own.fields.weightGrams,
      diameterMm: found.diameterMm ?? own.fields.diameterMm,
      presentation: found.presentation || own.fields.presentation,
      rarity: own.fields.rarity || found.rarity,
      obverseInscription: found.obverseInscription || own.fields.obverseInscription,
      reverseInscription: found.reverseInscription || own.fields.reverseInscription,
      obverseDescription: found.obverseDescription || own.fields.obverseDescription,
      reverseDescription: found.reverseDescription || own.fields.reverseDescription,
      notes: clip(notes, 5000),
    },
    references: mergeReferences(numista.references, own.references),
  }
}

/** The Add Coin page, filled in with a release. */
export function releaseAddRoute(release: CoinRelease): RouteLocationRaw {
  const { source, kind, country, title, variants, faceValue, currency, imageUrl, pageUrl, from, volume, upcoming } = release
  return { name: 'add-coin', query: { release: JSON.stringify({ source, kind, country, title, variants, faceValue, currency, imageUrl, pageUrl, from, volume, upcoming }) } }
}

/** The release the Add Coin page was opened with; null when there is none or it does not read. */
export function releaseFromQuery(value: unknown): CoinRelease | null {
  if (typeof value !== 'string' || !value) return null
  let raw: Record<string, unknown>
  try {
    raw = JSON.parse(value) as Record<string, unknown>
  } catch {
    return null
  }
  if (!raw || typeof raw !== 'object') return null
  const { source, kind, country, title, variants, faceValue, currency, imageUrl, pageUrl, from, volume, upcoming } = raw
  if (!SOURCES.includes(source as CoinRelease['source']) || (kind !== 'commemorative' && kind !== 'collector')) return null
  if (typeof title !== 'string' || !title.trim() || typeof faceValue !== 'number' || !(faceValue > 0) || typeof from !== 'string') return null
  return {
    source: source as CoinRelease['source'],
    kind,
    country: typeof country === 'string' ? country : '',
    title: title.trim(),
    variants: Array.isArray(variants) ? variants.filter((variant): variant is string => typeof variant === 'string') : undefined,
    faceValue,
    currency: currency === 'PTE' ? 'PTE' : undefined,
    imageUrl: typeof imageUrl === 'string' && /^https:\/\//.test(imageUrl) ? imageUrl : '',
    pageUrl: typeof pageUrl === 'string' && /^https:\/\//.test(pageUrl) ? pageUrl : '',
    from,
    to: from,
    volume: typeof volume === 'number' && volume > 0 ? volume : undefined,
    upcoming: upcoming === true,
    owned: false,
  }
}
