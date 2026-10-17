import { afterEach, describe, expect, it } from 'vitest'
import en from '@/i18n/locales/en'
import ptPT from '@/i18n/locales/pt-PT'
import { intlLocale, locale, setLocale, SUPPORTED_LOCALES, t, type MessageKey } from '@/i18n'

type Tree = { [key: string]: unknown }

// Same rule as the runtime: a plural message is exactly { one, other }.
function isPlural(value: unknown): value is { one: string; other: string } {
  if (!value || typeof value !== 'object') return false
  const keys = Object.keys(value).sort()
  return keys.length === 2 && keys[0] === 'one' && keys[1] === 'other'
}

function leaves(tree: Tree, prefix = ''): Map<string, string[]> {
  const out = new Map<string, string[]>()
  for (const [key, value] of Object.entries(tree)) {
    const path = prefix ? `${prefix}.${key}` : key
    if (typeof value === 'string') {
      out.set(path, [value])
    } else if (isPlural(value)) {
      out.set(path, [value.one, value.other])
    } else {
      for (const [k, v] of leaves(value as Tree, path)) out.set(k, v)
    }
  }
  return out
}

const placeholders = (text: string) => [...text.matchAll(/\{(\w+)\}/g)].map((m) => m[1]).sort()

describe('i18n catalogs', () => {
  const enLeaves = leaves(en)
  const ptLeaves = leaves(ptPT)

  it('gives Portuguese exactly the same keys as English', () => {
    expect([...ptLeaves.keys()].sort()).toEqual([...enLeaves.keys()].sort())
  })

  it('has no empty messages', () => {
    for (const [key, values] of [...enLeaves, ...ptLeaves]) {
      for (const value of values) expect(value.trim(), key).not.toBe('')
    }
  })

  it('keeps the same {placeholders} in every translation', () => {
    for (const [key, enValues] of enLeaves) {
      const expected = placeholders(enValues.join(' '))
      const actual = placeholders((ptLeaves.get(key) ?? []).join(' '))
      expect([...new Set(actual)], key).toEqual([...new Set(expected)])
    }
  })
})

describe('t()', () => {
  afterEach(() => setLocale('en'))

  it('defaults to English in the test environment', () => {
    expect(locale.value).toBe('en')
    expect(t('nav.settings')).toBe('Settings')
    expect(intlLocale()).toBe('en-US')
  })

  it('switches language, persists the choice and updates <html lang>', () => {
    setLocale('pt-PT')
    expect(t('nav.settings')).toBe('Definições')
    expect(localStorage.getItem('locale')).toBe('pt-PT')
    expect(document.documentElement.lang).toBe('pt-PT')
    expect(intlLocale()).toBe('pt-PT')
  })

  it('fills placeholders', () => {
    expect(t('coinDetail.sets.remove', { name: 'Euro' })).toBe('Remove from Euro')
    setLocale('pt-PT')
    expect(t('coinDetail.sets.remove', { name: 'Euro' })).toBe('Retirar de Euro')
  })

  it('picks plural forms per locale', () => {
    expect(t('auth.login.retryIn', { count: 1 })).toMatch(/1 second\.$/)
    expect(t('auth.login.retryIn', { count: 5 })).toMatch(/5 seconds\.$/)
    setLocale('pt-PT')
    expect(t('auth.login.retryIn', { count: 1 })).toMatch(/1 segundo\.$/)
    expect(t('auth.login.retryIn', { count: 0 })).toMatch(/0 segundos\.$/)
  })

  it('returns the key for an unknown message instead of rendering blank', () => {
    expect(t('nav.doesNotExist' as MessageKey)).toBe('nav.doesNotExist')
  })

  it('ignores unsupported locales', () => {
    setLocale('fr' as never)
    expect(locale.value).toBe('en')
    expect(SUPPORTED_LOCALES.map((option) => option.code)).toEqual(['en', 'pt-PT'])
  })
})
