import { ref } from 'vue'
import en from './locales/en'
import ptPT from './locales/pt-PT'
import type { MessageKey, MessageParams, Messages, PluralMessage } from './types'

export type { MessageKey, MessageParams } from './types'

export type Locale = 'en' | 'pt-PT'

export interface LocaleOption {
  code: Locale
  /** Name of the language in that language, so it is recognisable whatever the current locale. */
  label: string
  short: string
}

export const SUPPORTED_LOCALES: readonly LocaleOption[] = [
  { code: 'en', label: 'English', short: 'EN' },
  { code: 'pt-PT', label: 'Português', short: 'PT' },
]

const STORAGE_KEY = 'locale'

const catalogs: Record<Locale, Messages> = { en, 'pt-PT': ptPT }

// BCP 47 tags handed to Intl formatters. English stays en-US so existing
// currency and date output is unchanged.
const INTL_TAGS: Record<Locale, string> = { en: 'en-US', 'pt-PT': 'pt-PT' }

function isLocale(value: unknown): value is Locale {
  return SUPPORTED_LOCALES.some((option) => option.code === value)
}

function detectInitialLocale(): Locale {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (isLocale(saved)) return saved
  } catch {
    // Storage can be unavailable (private mode, blocked site data).
  }
  const preferred = typeof navigator !== 'undefined' ? navigator.language : ''
  return preferred?.toLowerCase().startsWith('pt') ? 'pt-PT' : 'en'
}

/** Current UI locale. Reading it inside a render or computed makes that output follow locale changes. */
export const locale = ref<Locale>(detectInitialLocale())

function applyDocumentLang() {
  if (typeof document !== 'undefined') {
    document.documentElement.lang = locale.value
  }
}

applyDocumentLang()

export function setLocale(next: Locale) {
  if (!isLocale(next)) return
  locale.value = next
  try {
    localStorage.setItem(STORAGE_KEY, next)
  } catch {
    // Non-persistent choice still applies for this session.
  }
  applyDocumentLang()
}

/** Locale tag for Intl / toLocale*String calls (e.g. `en-US`, `pt-PT`). */
export function intlLocale(): string {
  return INTL_TAGS[locale.value]
}

function lookup(messages: Messages, key: string): unknown {
  let node: unknown = messages
  for (const part of key.split('.')) {
    if (!node || typeof node !== 'object') return undefined
    node = (node as Record<string, unknown>)[part]
  }
  return node
}

function isPluralMessage(value: unknown): value is PluralMessage {
  if (!value || typeof value !== 'object') return false
  const { one, other } = value as Partial<PluralMessage>
  return typeof one === 'string' && typeof other === 'string' && Object.keys(value).length === 2
}

const pluralRulesCache = new Map<string, Intl.PluralRules>()

function selectPlural(message: PluralMessage, count: number): string {
  const tag = intlLocale()
  let rules = pluralRulesCache.get(tag)
  if (!rules) {
    rules = new Intl.PluralRules(tag)
    pluralRulesCache.set(tag, rules)
  }
  return rules.select(count) === 'one' ? message.one : message.other
}

function interpolate(template: string, params?: MessageParams): string {
  if (!params) return template
  return template.replace(/\{(\w+)\}/g, (match, name: string) =>
    Object.prototype.hasOwnProperty.call(params, name) ? String(params[name]) : match,
  )
}

/**
 * Translate `key` into the current locale, filling `{name}` placeholders from
 * `params`. Plural messages pick their form from `params.count`. Falls back to
 * English, then to the key itself, so a gap never renders blank.
 */
export function t(key: MessageKey, params?: MessageParams): string {
  let message = lookup(catalogs[locale.value], key)
  if (message === undefined) message = lookup(catalogs.en, key)

  if (typeof message === 'string') return interpolate(message, params)
  if (isPluralMessage(message)) {
    const count = Number(params?.count ?? 0)
    return interpolate(selectPlural(message, count), params)
  }

  if (import.meta.env.DEV) console.warn(`[i18n] Missing message: ${key}`)
  return key
}
