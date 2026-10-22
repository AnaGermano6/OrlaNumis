import { intlLocale, locale, t } from '@/i18n'
import type { NumistaCatalogPhoto, NumistaIssueSummary } from '@/api/client'
import { noteVariantLabel } from '@/utils/banknoteAttributes'

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

// The day an issue is dated, when Numista has it (banknotes); years before 100
// would be read as 19xx.
function issueDay(issue: NumistaIssueSummary): number | undefined {
  return issue.year !== undefined && issue.year >= 100 && issue.month ? Date.UTC(issue.year, issue.month - 1, issue.day || 1) : undefined
}

/** "13/09/1978" when Numista has the day an issue is dated, else its year or years. */
export function numistaIssueDate(issue: NumistaIssueSummary): string {
  const day = issueDay(issue)
  if (day !== undefined) {
    const parts: Intl.DateTimeFormatOptions = issue.day ? { day: '2-digit', month: '2-digit', year: 'numeric' } : { month: '2-digit', year: 'numeric' }
    return new Intl.DateTimeFormat(intlLocale(), { ...parts, timeZone: 'UTC' }).format(day)
  }
  if (issue.year !== undefined) return formatNumistaYear(issue.year)
  if (issue.minYear !== undefined || issue.maxYear !== undefined) return numistaYearRange(issue.minYear, issue.maxYear)
  return t('coinDetail.marketValue.undated')
}

/** The date printed on a banknote of the issue ("13 de setembro de 1978"), else its year; '' when undated. */
export function numistaIssuePrintedDate(issue: NumistaIssueSummary): string {
  const day = issueDay(issue)
  if (day !== undefined) {
    const parts: Intl.DateTimeFormatOptions = issue.day ? { day: 'numeric', month: 'long', year: 'numeric' } : { month: 'long', year: 'numeric' }
    return new Intl.DateTimeFormat(intlLocale(), { ...parts, timeZone: 'UTC' }).format(day)
  }
  return issue.year !== undefined ? formatNumistaYear(issue.year) : ''
}

const NAME_SUFFIXES = new Set(['neto', 'filho', 'sobrinho', 'júnior', 'junior', 'jr', 'jr.'])

/** How a signatory is usually called: "Lopes", "Campos Neto". */
export function signerSurname(name: string): string {
  const words = name.trim().split(/\s+/)
  const last = words.at(-1) ?? ''
  return words.length > 2 && NAME_SUFFIXES.has(last.toLowerCase()) ? words.slice(-2).join(' ') : last
}

/**
 * "1968 · L · 5 000 000 minted" for an issue picker; a banknote's issue reads
 * "13/09/1978 · P176a · Vilar / Reis · 29 645 000 minted".
 */
export function numistaIssueLabel(issue: NumistaIssueSummary): string {
  const parts: string[] = [numistaIssueDate(issue)]
  if (issue.mintLetter) parts.push(issue.mintLetter)
  if (issue.pick) parts.push(`P${issue.pick}`)
  if (issue.signatures?.length) parts.push(issue.signatures.map((signature) => signerSurname(signature.name)).join(' / '))
  if (issue.variant) parts.push(noteVariantLabel(issue.variant))
  if (issue.mintage) parts.push(t('coinDetail.marketValue.mintage', { count: new Intl.NumberFormat(intlLocale()).format(issue.mintage) }))
  // A banknote issue's comment repeats its Pick number and signatories.
  if (issue.comment && !issue.pick && !issue.signatures?.length) parts.push(issue.comment.length > 40 ? `${issue.comment.slice(0, 39)}…` : issue.comment)
  return parts.join(' · ')
}

/** "© Author · CC BY-NC" credit for a catalogue photo. */
export function catalogPhotoCredit(photo: NumistaCatalogPhoto): string {
  return [photo.copyright ? `© ${photo.copyright}` : '', photo.license ?? ''].filter(Boolean).join(' · ')
}
