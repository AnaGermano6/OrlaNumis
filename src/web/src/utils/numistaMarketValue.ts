import type { NumistaDataStatus } from '@/api/client'
import { locale, t, type MessageKey } from '@/i18n'

/** Currencies offered for Numista price estimates. */
export const NUMISTA_CURRENCIES = ['EUR', 'USD', 'GBP'] as const

const CURRENCY_KEY = 'numista:marketValueCurrency'

/** The viewer's last chosen currency, else EUR in Portuguese and USD in English. */
export function readPreferredCurrency(): string {
  try {
    const saved = localStorage.getItem(CURRENCY_KEY)
    if (saved && (NUMISTA_CURRENCIES as readonly string[]).includes(saved)) return saved
  } catch {
    // Storage can be unavailable; fall through to the locale default.
  }
  return locale.value === 'pt-PT' ? 'EUR' : 'USD'
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
