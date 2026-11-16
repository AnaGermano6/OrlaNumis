import type { CollectionSummary } from '@/api/client'
import { intlLocale, t } from '@/i18n'
import { collectionCurrency, currencyName } from '@/utils/collectionCurrency'
import { formatCurrency } from '@/utils/format'

/** A card of the face value, or of how much more than it the items are worth. */
export interface FaceValueKpi {
  key: 'faceValue' | 'overFace'
  display: string
  label: string
  hint: string
  note: string
}

// The face value of a summary's items, in the currency every page uses, and
// how much more than it they are worth ("+85,3 %"), for the overview's cards.
export function faceValueKpis(summary: CollectionSummary | null): FaceValueKpi[] {
  const face = summary?.faceValue
  const over = summary?.valueOverFace
  return [
    {
      key: 'faceValue',
      display: face == null ? '—' : formatCurrency(face),
      label: t('collection.kpis.faceValue'),
      hint: t('collection.kpis.faceValueHint', { currency: currencyName(collectionCurrency.value) }),
      note: summary?.withoutFaceValue ? t('collection.kpis.withoutFaceValue', { count: summary.withoutFaceValue }) : '',
    },
    {
      key: 'overFace',
      display: over == null ? '—' : new Intl.NumberFormat(intlLocale(), { style: 'percent', signDisplay: 'exceptZero', maximumFractionDigits: 1 }).format(over / 100),
      label: t('collection.kpis.overFace'),
      hint: t('collection.kpis.overFaceHint'),
      note: '',
    },
  ]
}
