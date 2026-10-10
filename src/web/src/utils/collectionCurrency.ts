// The currency coin prices and values are kept in: US dollars unless an admin
// switched the collection to euros. Loaded with the coin options at sign-in.
import { ref } from 'vue'

export const COLLECTION_CURRENCIES = ['USD', 'EUR'] as const
export type CollectionCurrency = (typeof COLLECTION_CURRENCIES)[number]

export const collectionCurrency = ref<CollectionCurrency>('USD')

export function setCollectionCurrency(code: string | null | undefined) {
  if (code === 'USD' || code === 'EUR') collectionCurrency.value = code
}

/** "$" or "€", for field labels. */
export function collectionCurrencySymbol(): string {
  return collectionCurrency.value === 'EUR' ? '€' : '$'
}

/**
 * An amount in another currency in the collection currency, with the ECB rate
 * (US dollars per euro). Undefined when it cannot be converted.
 */
export function toCollectionCurrency(amount: number, from: string, usdPerEur: number | undefined): number | undefined {
  const target = collectionCurrency.value
  if (from === target) return amount
  if (!usdPerEur || usdPerEur <= 0) return undefined
  if (from === 'USD' && target === 'EUR') return amount / usdPerEur
  if (from === 'EUR' && target === 'USD') return amount * usdPerEur
  return undefined
}
