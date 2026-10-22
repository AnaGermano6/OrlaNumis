// The collection's monetary systems as bars of a chart: the most common
// ones, and the rest added up in one bar.
import { t } from '@/i18n'

/** Label of the bar that adds up the systems not shown on their own. */
export const OTHER_MONETARY_SYSTEMS = '\u0000other-monetary-systems'

export interface MonetarySystemBar {
  label: string
  count: number
}

export function monetarySystemBars(
  counts: { monetarySystem: string; count: number }[] | undefined,
  shown = 12,
): { bars: MonetarySystemBar[]; others: number } {
  const sorted = [...(counts ?? [])].sort((a, b) => b.count - a.count)
  const bars = sorted.slice(0, shown).map((item) => ({ label: item.monetarySystem, count: item.count }))
  const rest = sorted.slice(shown)
  if (rest.length) bars.push({ label: OTHER_MONETARY_SYSTEMS, count: rest.reduce((sum, item) => sum + item.count, 0) })
  return { bars, others: rest.length }
}

/** A bar's label: the system's name, or "Other systems (n)" for the bar that adds them up. */
export function monetarySystemBarLabel(label: string, others: number): string {
  return label === OTHER_MONETARY_SYSTEMS ? t('stats.charts.otherMonetarySystems', { count: others }) : label
}
