// The world map's country outlines, and how the collection's free-text
// countries (Numista issuer names, in English or Portuguese, or typed by
// hand) are matched to them (see utils/collectionMap). Issuers that are not
// modern countries ("Corinth", "Roman Empire") stay unmatched.
import type { Feature, FeatureCollection, MultiPolygon, Polygon } from 'geojson'

export interface WorldCountryProperties {
  /** ISO 3166-1 alpha-2 code (alpha-3 for the few without one). */
  iso: string
  name: string
  namePt: string
  aliases: string[]
}

export type WorldCountryFeature = Feature<Polygon | MultiPolygon, WorldCountryProperties>
export type WorldCountryCollection = FeatureCollection<Polygon | MultiPolygon, WorldCountryProperties>

let pending: Promise<WorldCountryCollection> | null = null

/** Country outlines (Natural Earth, public domain), loaded on first use as a separate chunk. */
export function loadWorldCountries(): Promise<WorldCountryCollection> {
  pending ??= import('@/assets/geo/world-countries.json')
    .then((module) => module.default as unknown as WorldCountryCollection)
    .catch((error: unknown) => {
      pending = null
      throw error
    })
  return pending
}

/** "  República   Checa " → "republica checa": accents, case, punctuation and spacing ignored. */
export function foldCountryName(value: string): string {
  return value
    .toLowerCase()
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .replace(/[^a-z0-9]+/g, ' ')
    .trim()
}

/** Names collections use that the map's own names and the browser's leave out. */
export const EXTRA_COUNTRY_ALIASES: Record<string, string[]> = {
  US: ['Estados Unidos da América', 'EUA', 'E.U.A.', 'USA'],
  GB: ['Grã-Bretanha', 'Great Britain'],
  NL: ['Holanda', 'Holland'],
  CZ: ['República Tcheca', 'Tchéquia'],
  RU: ['Federação Russa'],
  IR: ['Irã'],
  VN: ['Vietnã'],
}

/**
 * The ISO code of a stored country name: the whole name, else without what is
 * in brackets, else the part before a comma ("Alemanha, República Federal da").
 */
export function lookupCountry(index: Map<string, string>, name: string): string | undefined {
  const candidates = [name, name.replace(/\([^)]*\)/g, ' '), name.split(',')[0] ?? '']
  for (const candidate of candidates) {
    const iso = index.get(foldCountryName(candidate))
    if (iso) return iso
  }
  return undefined
}

/** Fill opacity of a colour step: countries without coins stay faint; a single coin already stands out. */
export function countryShadeOpacity(shade: number): number {
  return [0.35, 0.5, 0.65, 0.8, 0.95][shade] ?? 0.35
}

/** A map country's name in the app's language. */
export function countryFeatureLabel(feature: WorldCountryFeature, language: string): string {
  return language === 'pt-PT' ? feature.properties.namePt : feature.properties.name
}
