// The currency coin prices and values are kept in and every page shows: US
// dollars unless an admin switched the collection to another (Settings ›
// Appearance). Loaded with the coin options at sign-in.
import { ref } from 'vue'
import { intlLocale } from '@/i18n'

/** The euro and the currencies the ECB publishes a daily euro rate for. */
export const COLLECTION_CURRENCIES = [
  'EUR', 'USD', 'GBP', 'CHF', 'JPY', 'CAD', 'AUD', 'NZD', 'BRL', 'CNY', 'HKD', 'SGD', 'KRW', 'INR', 'IDR',
  'MYR', 'PHP', 'THB', 'MXN', 'ZAR', 'ILS', 'TRY', 'SEK', 'NOK', 'DKK', 'ISK', 'PLN', 'CZK', 'HUF', 'RON',
] as const
export type CollectionCurrency = (typeof COLLECTION_CURRENCIES)[number]

export const collectionCurrency = ref<CollectionCurrency>('USD')

function isCollectionCurrency(code: string): code is CollectionCurrency {
  return (COLLECTION_CURRENCIES as readonly string[]).includes(code)
}

export function setCollectionCurrency(code: string | null | undefined) {
  const upper = (code ?? '').trim().toUpperCase()
  if (isCollectionCurrency(upper)) collectionCurrency.value = upper
}

/** A currency's symbol in the app's language: "€", "$", "£", "CHF". */
export function currencySymbol(code: string): string {
  try {
    return new Intl.NumberFormat(intlLocale(), { style: 'currency', currency: code, currencyDisplay: 'narrowSymbol' })
      .formatToParts(0).find((part) => part.type === 'currency')?.value ?? code
  } catch {
    return code
  }
}

/** The collection currency's symbol, for field labels. */
export function collectionCurrencySymbol(): string {
  return currencySymbol(collectionCurrency.value)
}

/** "Euro (€)", "Libra esterlina (£)": a currency's name in the app's language. */
export function currencyName(code: string): string {
  let name = code
  try {
    name = new Intl.DisplayNames([intlLocale()], { type: 'currency' }).of(code) ?? code
  } catch {
    // An older browser without currency names shows the code.
  }
  const symbol = currencySymbol(code)
  const label = name.charAt(0).toLocaleUpperCase(intlLocale()) + name.slice(1)
  return symbol && symbol !== code ? `${label} (${symbol})` : `${label} (${code})`
}

/**
 * An amount in another currency in the collection currency, with the ECB
 * rates (how much of each currency one euro buys). Undefined when it cannot
 * be converted.
 */
export function toCollectionCurrency(amount: number, from: string, perEur: Record<string, number> | undefined): number | undefined {
  const target = collectionCurrency.value
  if (from === target) return amount
  const rate = (code: string) => (code === 'EUR' ? 1 : perEur?.[code])
  const fromRate = rate(from)
  const toRate = rate(target)
  if (!fromRate || !toRate || fromRate <= 0 || toRate <= 0) return undefined
  return (amount / fromRate) * toRate
}
