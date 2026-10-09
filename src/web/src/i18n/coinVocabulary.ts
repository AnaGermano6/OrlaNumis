// Display names for coin categories and eras.
//
// Categories and eras are admin-configured free text that is stored on each
// coin and validated by the API, so the stored value never changes with the
// UI language. Translation is display-only: option `value`s, filters, API
// payloads, and stats keys keep the stored text, and only what the user reads
// goes through categoryLabel()/eraLabel().
//
// A Portuguese name comes from, in order:
//   1. the admin's own name for that value (Admin → Coin Properties),
//   2. the built-in dictionary below (defaults and common numismatic names),
//   3. the stored value unchanged.
import { reactive } from 'vue'
import { getCoinOptions, type CoinOptionLabels } from '@/api/client'
import { CATEGORIES, COIN_ERAS } from '@/types'
import { parseOptionList } from '@/utils/options'
import { locale, type Locale } from '@/i18n'

export type CoinVocabularyKind = 'category' | 'era'

// Keys are lower-case stored values. Labels are feminine adjectives so they
// read naturally for both a coin ("moeda romana") and an era ("época antiga");
// the first letter follows the stored value's case.
const BUILT_IN_PT: Record<string, string> = {
  // Built-in categories and eras.
  roman: 'romana',
  greek: 'grega',
  byzantine: 'bizantina',
  modern: 'moderna',
  other: 'outra',
  ancient: 'antiga',
  medieval: 'medieval',
  // Common additions.
  'roman provincial': 'romana provincial',
  'roman republican': 'romana republicana',
  'roman imperial': 'romana imperial',
  'greek imperial': 'grega imperial',
  provincial: 'provincial',
  republican: 'republicana',
  imperial: 'imperial',
  archaic: 'arcaica',
  classical: 'clássica',
  hellenistic: 'helenística',
  'late antiquity': 'antiguidade tardia',
  'early modern': 'moderna inicial',
  contemporary: 'contemporânea',
  celtic: 'celta',
  etruscan: 'etrusca',
  thracian: 'trácia',
  macedonian: 'macedónia',
  lydian: 'lídia',
  persian: 'persa',
  achaemenid: 'aqueménida',
  parthian: 'parta',
  sasanian: 'sassânida',
  sassanian: 'sassânida',
  seleucid: 'selêucida',
  ptolemaic: 'ptolemaica',
  egyptian: 'egípcia',
  phoenician: 'fenícia',
  carthaginian: 'cartaginesa',
  nabataean: 'nabateia',
  judaean: 'judaica',
  judean: 'judaica',
  jewish: 'judaica',
  bactrian: 'bactriana',
  'indo-greek': 'indo-grega',
  kushan: 'cuchana',
  indian: 'indiana',
  chinese: 'chinesa',
  japanese: 'japonesa',
  islamic: 'islâmica',
  ottoman: 'otomana',
  mughal: 'mogol',
  crusader: 'cruzada',
  visigothic: 'visigótica',
  ostrogothic: 'ostrogótica',
  vandal: 'vândala',
  merovingian: 'merovíngia',
  carolingian: 'carolíngia',
  'anglo-saxon': 'anglo-saxónica',
  colonial: 'colonial',
  world: 'mundial',
  portuguese: 'portuguesa',
  spanish: 'espanhola',
  british: 'britânica',
  english: 'inglesa',
  french: 'francesa',
  german: 'alemã',
  italian: 'italiana',
  american: 'americana',
}

interface CoinVocabularyState {
  categories: string[]
  eras: string[]
  /** Admin-defined names keyed by locale, then by stored value. */
  labels: Record<string, CoinOptionLabels>
  loaded: boolean
}

/** Configured choices and admin names. Reactive, so labels re-render once loaded. */
export const coinVocabulary = reactive<CoinVocabularyState>({
  categories: [...CATEGORIES],
  eras: [...COIN_ERAS],
  labels: {},
  loaded: false,
})

let pending: Promise<void> | null = null

/**
 * Fetch the configured choices and admin names. Concurrent callers share one
 * request; call again after the admin saves Coin Properties to pick up changes.
 * Needs a signed-in user; on failure the current (or default) vocabulary stays.
 */
export function loadCoinVocabulary(): Promise<void> {
  pending ??= (async () => {
    try {
      const { data } = await getCoinOptions()
      coinVocabulary.categories = parseOptionList(data.categories?.join('\n'), CATEGORIES)
      coinVocabulary.eras = parseOptionList(data.eras?.join('\n'), COIN_ERAS)
      coinVocabulary.labels = data.labels ?? {}
      coinVocabulary.loaded = true
    } catch {
      // Keep what we have; labels fall back to the built-in dictionary.
    } finally {
      pending = null
    }
  })()
  return pending
}

/** Back to the built-in defaults (used on sign-out and in tests). */
export function resetCoinVocabulary() {
  coinVocabulary.categories = [...CATEGORIES]
  coinVocabulary.eras = [...COIN_ERAS]
  coinVocabulary.labels = {}
  coinVocabulary.loaded = false
}

function matchFirstLetterCase(source: string, label: string): string {
  const first = source.charAt(0)
  return first !== first.toLowerCase() ? label.charAt(0).toUpperCase() + label.slice(1) : label
}

function adminLabel(kind: CoinVocabularyKind, value: string, target: Locale): string | undefined {
  const names = coinVocabulary.labels[target]?.[kind === 'category' ? 'categories' : 'eras']
  if (!names) return undefined
  if (names[value]) return names[value]
  const lower = value.toLowerCase()
  const key = Object.keys(names).find(candidate => candidate.toLowerCase() === lower)
  return key ? names[key] : undefined
}

/** The built-in Portuguese name for a stored value, or undefined when there is none. */
export function builtInCoinOptionLabel(value: string): string | undefined {
  const trimmed = value.trim()
  const label = BUILT_IN_PT[trimmed.toLowerCase()]
  return label ? matchFirstLetterCase(trimmed, label) : undefined
}

/** Display name for a stored category or era in the current locale. */
export function coinOptionLabel(kind: CoinVocabularyKind, value: string | null | undefined): string {
  const raw = value ?? ''
  const trimmed = raw.trim()
  if (!trimmed || locale.value === 'en') return raw
  return adminLabel(kind, trimmed, locale.value) ?? builtInCoinOptionLabel(trimmed) ?? raw
}

export function categoryLabel(value: string | null | undefined): string {
  return coinOptionLabel('category', value)
}

export function eraLabel(value: string | null | undefined): string {
  return coinOptionLabel('era', value)
}
