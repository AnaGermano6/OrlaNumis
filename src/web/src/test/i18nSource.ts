import en from '@/i18n/locales/en'

function englishText(key: string): string | undefined {
  const value = key.split('.').reduce<unknown>(
    (node, part) => (node && typeof node === 'object' ? (node as Record<string, unknown>)[part] : undefined),
    en,
  )
  return typeof value === 'string' ? value : undefined
}

/**
 * Inlines the English text of parameterless t('key') calls in component source,
 * so source-scanning tests can keep asserting on user-facing copy after a
 * component moves its strings into the i18n catalog.
 */
export function withEnglishCopy(source: string): string {
  return source
    .replace(/:([\w-]+)="t\('([\w.]+)'\)"/g, (match, attr: string, key: string) => {
      const text = englishText(key)
      return text === undefined ? match : `${attr}="${text}"`
    })
    .replace(/\{\{ t\('([\w.]+)'\) \}\}/g, (match, key: string) => englishText(key) ?? match)
    .replace(/\bt\('([\w.]+)'\)/g, (match, key: string) => {
      const text = englishText(key)
      return text === undefined ? match : `'${text}'`
    })
}
