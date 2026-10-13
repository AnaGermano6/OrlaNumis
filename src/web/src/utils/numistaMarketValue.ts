import type { NumistaDataStatus } from '@/api/client'
import { t, type MessageKey } from '@/i18n'
import { collectionCurrency } from '@/utils/collectionCurrency'

/** Currencies offered for Numista price estimates, besides the collection's. */
export const NUMISTA_CURRENCIES = ['EUR', 'USD', 'GBP'] as const

/** The currencies to view Numista's prices in: the collection's, then the usual ones. */
export function numistaCurrencies(): string[] {
  return [...new Set<string>([collectionCurrency.value, ...NUMISTA_CURRENCIES])]
}

const CURRENCY_KEY = 'numista:marketValueCurrency'

/** The viewer's last chosen currency, else the one every page uses. */
export function readPreferredCurrency(): string {
  try {
    const saved = localStorage.getItem(CURRENCY_KEY)
    if (saved && numistaCurrencies().includes(saved)) return saved
  } catch {
    // Storage can be unavailable; fall through to the locale default.
  }
  return collectionCurrency.value
}

export function savePreferredCurrency(code: string) {
  try {
    localStorage.setItem(CURRENCY_KEY, code)
  } catch {
    // The choice still applies for this visit.
  }
}

const STATUS_MESSAGES: Record<Exclude<NumistaDataStatus, 'ok'>, MessageKey> = {
  not_linked: 'coinDetail.marketValue.notLinked',
  no_issues: 'coinDetail.marketValue.noIssues',
  no_prices: 'coinDetail.marketValue.noPrices',
  unconfigured: 'coinDetail.marketValue.unconfigured',
  unauthorized: 'coinDetail.marketValue.unauthorized',
  quota_limited: 'coinDetail.marketValue.quotaLimited',
  unavailable: 'coinDetail.marketValue.unavailable',
}

/** Explains a non-ok Numista status to the user. */
export function numistaStatusMessage(status: NumistaDataStatus): string {
  return status === 'ok' ? '' : t(STATUS_MESSAGES[status])
}
