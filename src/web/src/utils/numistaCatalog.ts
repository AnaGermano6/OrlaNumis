import { intlLocale, locale, t } from '@/i18n'
import type { NumistaCatalogPhoto, NumistaIssueSummary } from '@/api/client'

/** Numista language for catalogue data in the current UI locale. */
export function numistaLang(): string {
  return locale.value === 'pt-PT' ? 'pt' : 'en'
}

/** A Numista year (negative for BC) as text, e.g. "1968" or "44 BC". */
export function formatNumistaYear(year: number): string {
  return year < 0 ? t('coinDetail.marketValue.yearBC', { year: -year }) : String(year)
}

/** "1953–1974", "1968" or "" from a year range. */
export function numistaYearRange(minYear?: number, maxYear?: number): string {
  if (minYear === undefined && maxYear === undefined) return ''
  if (minYear === undefined || maxYear === undefined || minYear === maxYear) return formatNumistaYear((minYear ?? maxYear)!)
  return `${formatNumistaYear(minYear)}–${formatNumistaYear(maxYear)}`
}

/** "1968 · L · 5 000 000 minted" for an issue picker. */
export function numistaIssueLabel(issue: NumistaIssueSummary): string {
  const parts: string[] = []
  if (issue.year !== undefined) parts.push(formatNumistaYear(issue.year))
  else if (issue.minYear !== undefined || issue.maxYear !== undefined) parts.push(numistaYearRange(issue.minYear, issue.maxYear))
  else parts.push(t('coinDetail.marketValue.undated'))
  if (issue.mintLetter) parts.push(issue.mintLetter)
  if (issue.mintage) parts.push(t('coinDetail.marketValue.mintage', { count: new Intl.NumberFormat(intlLocale()).format(issue.mintage) }))
  if (issue.comment) parts.push(issue.comment.length > 40 ? `${issue.comment.slice(0, 39)}…` : issue.comment)
  return parts.join(' · ')
}

/** "© Author · CC BY-NC" credit for a catalogue photo. */
export function catalogPhotoCredit(photo: NumistaCatalogPhoto): string {
  return [photo.copyright ? `© ${photo.copyright}` : '', photo.license ?? ''].filter(Boolean).join(' · ')
}
