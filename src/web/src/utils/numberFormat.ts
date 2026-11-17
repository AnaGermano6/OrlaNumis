import { intlLocale } from '@/i18n'

/**
 * Intl.NumberFormat in the app's language that always groups thousands:
 * "3 874,51 €" (Portuguese, by default, groups only from 10 000). A format
 * that asks for no grouping (useGrouping: false) keeps none.
 */
export function numberFormat(options: Intl.NumberFormatOptions = {}): Intl.NumberFormat {
  // "always" is newer than the TypeScript library the app builds with.
  const useGrouping = (options.useGrouping === false ? false : 'always') as Intl.NumberFormatOptions['useGrouping']
  return new Intl.NumberFormat(intlLocale(), { ...options, useGrouping })
}

/**
 * A share of a whole as a percentage, whole but for what would round to a
 * misleading 0% or 100%: "0,3%" for 14 of 4208, "99,6%" for one short.
 */
export function formatShare(value: number): string {
  const precise = (value > 0 && value < 0.01) || (value > 0.99 && value < 1)
  return numberFormat({ style: 'percent', maximumFractionDigits: precise ? 1 : 0 }).format(value)
}

/** A number in the app's language, its thousands always grouped. */
export function formatNumber(value: number, options?: Intl.NumberFormatOptions): string {
  return numberFormat(options).format(value)
}
