// The collection's purchase-price ranges. The server names them in dollars
// ("Under $50"), but their bounds are amounts in the currency every page
// uses, so they are shown in it: "Menos de 50 €", "50 € – 200 €".
import { intlLocale, t } from '@/i18n'
import { collectionCurrency } from '@/utils/collectionCurrency'

const PRICE_RANGE_BOUNDS: Record<string, [number, number | null]> = {
  'Under $50': [0, 50],
  '$50 - $200': [50, 200],
  '$200 - $500': [200, 500],
  '$500 - $1K': [500, 1000],
  '$1K+': [1000, null],
}

export function priceRangeLabel(range: string): string {
  const bounds = PRICE_RANGE_BOUNDS[range]
  if (!bounds) return range
  const money = (value: number) => new Intl.NumberFormat(intlLocale(), { style: 'currency', currency: collectionCurrency.value, maximumFractionDigits: 0 }).format(value)
  const [low, high] = bounds
  if (high === null) return t('stats.priceRanges.over', { amount: money(low) })
  if (low === 0) return t('stats.priceRanges.under', { amount: money(high) })
  return `${money(low)} – ${money(high)}`
}
