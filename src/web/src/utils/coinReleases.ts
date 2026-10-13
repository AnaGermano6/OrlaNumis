// Labels of the coins the ECB lists and the mints sell (new releases and the
// catalogue of commemorative and collector coins).
import type { CoinRelease } from '@/api/client'
import { intlLocale, t } from '@/i18n'

/** The country key of the euro area's joint issues. */
export const JOINT_ISSUE = 'EU'

/** A country's name from its ISO code; the euro area for a joint issue. */
export function releaseCountryName(code: string): string {
  if (!code || code === JOINT_ISSUE) return t('collectionOverview.releases.jointIssue')
  try {
    return new Intl.DisplayNames([intlLocale()], { type: 'region' }).of(code) ?? code
  } catch {
    return code
  }
}

/** A face value in euros (whatever the collection currency), or in escudos ("PTE"). */
export function formatFaceValue(value: number, currency = 'EUR'): string {
  if (currency === 'PTE') return t('sets.catalogue.escudos', { value: new Intl.NumberFormat(intlLocale()).format(value) })
  return new Intl.NumberFormat(intlLocale(), { style: 'currency', currency, minimumFractionDigits: Number.isInteger(value) ? 0 : 2 }).format(value)
}

/** "2 €", "200 escudos", or every face value of a design: "10 € · 200 €". */
export function releaseFaceValues(release: CoinRelease): string {
  return (release.faceValues ?? [release.faceValue]).map((value) => formatFaceValue(value, release.currency)).join(' · ')
}

function month(value: string): Date {
  const [year, monthNumber] = value.split('-').map(Number)
  return new Date(year ?? 1970, (monthNumber ?? 1) - 1, 1)
}

/** "novembro de 2025", "outubro – dezembro de 2025", or the year alone. */
export function releasePeriod(release: CoinRelease): string {
  if (!release.from) return release.issueText ?? ''
  const year = release.from.slice(0, 4)
  if (release.from === `${year}-01` && release.to === `${year}-12`) return year
  const format = new Intl.DateTimeFormat(intlLocale(), { month: 'long', year: 'numeric' })
  return release.from === release.to ? format.format(month(release.from)) : format.formatRange(month(release.from), month(release.to))
}

export function formatCount(value: number): string {
  return new Intl.NumberFormat(intlLocale()).format(value)
}
