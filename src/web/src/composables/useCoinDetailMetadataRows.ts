// T015: Coin detail metadata rows composable for #219
// This composable generates consistent metadata row schema from coin data

import { computed } from 'vue'
import type { Coin, CoinDetailMetadataRow } from '@/types'
import type { CoinStatus } from '@/api/client'
import { formatCurrency, materialLabel } from '@/utils/format'
import { sanitizeExternalUrl } from '@/composables/useSafeExternalLink'
import { intlLocale, t } from '@/i18n'
import { eraLabel } from '@/i18n/coinVocabulary'
import { presentationLabel, rarityLabel } from '@/utils/coinAttributes'
import { banknoteGradeLabel, banknoteMaterialLabel, noteVariantLabel, pickNumber } from '@/utils/banknoteAttributes'

// "Corrente · Comemorativa", "Não circulante · Comemorativa", "Desmonetizada
// em 28/02/2002"…; '' when nothing is known.
function statusValue(status: CoinStatus | null | undefined): string {
  if (!status) return ''
  const parts: string[] = []
  // A collector coin is "non-circulating" even while it is legal tender.
  if (status.nonCirculating) parts.push(t('coinDetail.nonCirculatingCoin'))
  else if (status.current === true) parts.push(t('coinDetail.currentCoin'))
  if (status.current === false) {
    const date = formatDemonetizationDate(status.demonetizationDate)
    parts.push(date ? t('coinDetail.demonetizedOn', { date }) : t('coinDetail.demonetizedCoin'))
  }
  if (status.commemorative) parts.push(t('coinDetail.commemorativeCoin'))
  return parts.join(' · ')
}

function formatDemonetizationDate(value: string | undefined): string {
  if (!value) return ''
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return value
  const date = new Date(`${value}T00:00:00Z`)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString(intlLocale(), { timeZone: 'UTC' })
}

export function useCoinDetailMetadataRows(coin: Coin | null, status?: CoinStatus | null) {
  const rows = computed<CoinDetailMetadataRow[]>(() => {
    if (!coin) return []

    const result: CoinDetailMetadataRow[] = []
    const banknote = coin.kind === 'banknote'

    // Purchase & Value
    if (coin.purchasePrice != null) {
      result.push({
        key: 'purchasePrice',
        label: t('coinFields.purchasePrice'),
        value: formatCurrency(coin.purchasePrice),
      })
    }

    if (coin.currentValue != null) {
      result.push({
        key: 'currentValue',
        label: t('coinFields.currentValue'),
        value: formatCurrency(coin.currentValue),
        valueClass: 'gold',
      })
    }

    if (coin.purchaseDate) {
      result.push({
        key: 'purchaseDate',
        label: t('coinFields.purchaseDate'),
        value: new Date(coin.purchaseDate).toLocaleDateString(intlLocale()),
      })
    }

    if (coin.vendorSku) {
      result.push({
        key: 'vendorSku',
        label: t('coinFields.vendorSku'),
        value: coin.vendorSku,
      })
    }

    if (coin.vendorInvoice) {
      result.push({
        key: 'vendorInvoice',
        label: t('coinFields.vendorInvoice'),
        value: coin.vendorInvoice,
      })
    }

    const circulation = statusValue(status)
    if (circulation) {
      result.push({ key: 'circulation', label: t('coinFields.circulation'), value: circulation })
    }

    if (coin.country) {
      result.push({ key: 'country', label: t('coinFields.country'), value: coin.country })
    }

    if (coin.monetarySystem) {
      result.push({ key: 'monetarySystem', label: t('coinFields.monetarySystem'), value: coin.monetarySystem })
    }
    if (coin.period) {
      result.push({ key: 'period', label: t('coinFields.period'), value: coin.period })
    }

    // How many are held; one for coins saved before the field existed.
    result.push({ key: 'quantity', label: t('coinFields.quantity'), value: String(coin.quantity && coin.quantity > 0 ? coin.quantity : 1) })

    if (coin.presentation) {
      result.push({ key: 'presentation', label: t('coinFields.presentation'), value: presentationLabel(coin.presentation) })
    }

    if (coin.rarity) {
      result.push({ key: 'rarity', label: t('coinFields.rarity'), value: rarityLabel(coin.rarity) })
    }

    // Physical attributes
    if (coin.denomination) {
      result.push({
        key: 'denomination',
        label: t('coinFields.denomination'),
        value: coin.denomination,
      })
    }

    // What sets a banknote apart: its issue, catalogue number and this note's own marks.
    if (banknote) {
      const pick = pickNumber(coin.references)
      const dimensions = coin.widthMm != null && coin.heightMm != null ? `${coin.widthMm} × ${coin.heightMm} mm` : ''
      for (const [key, label, value] of [
        ['series', t('banknoteForm.series'), coin.series],
        ['pick', t('banknoteForm.pick'), pick ? `P${pick}` : ''],
        ['serialNumber', t('banknoteForm.serialNumber'), coin.serialNumber],
        ['signatures', t('banknoteForm.signatures'), coin.signatures],
        ['noteVariant', t('banknoteForm.variant'), coin.noteVariant ? noteVariantLabel(coin.noteVariant) : ''],
        ['printer', t('banknoteForm.printer'), coin.printer],
        ['dimensions', t('banknoteForm.dimensions'), dimensions],
        ['watermark', t('banknoteForm.watermark'), coin.watermark],
        ['certifiedGrade', t('banknoteForm.certifiedGrade'), coin.certifiedGrade],
        ['certificationNumber', t('banknoteForm.certificationNumber'), coin.certificationNumber],
      ] as const) {
        if (value) result.push({ key, label, value, valueClass: key === 'certifiedGrade' ? 'gold' : undefined })
      }
    }

    if (coin.era) {
      result.push({
        key: 'era',
        label: t('coinFields.era'),
        value: eraLabel(coin.era),
      })
    }

    if (coin.mint && !banknote) {
      result.push({
        key: 'mint',
        label: t('coinFields.mint'),
        value: coin.mint,
      })
    }

    if (coin.material) {
      result.push({
        key: 'material',
        label: t('coinFields.material'),
        value: banknote ? banknoteMaterialLabel(coin.material) : materialLabel(coin.material),
      })
    }

    if (coin.weightGrams != null && !banknote) {
      result.push({
        key: 'weight',
        label: t('coinFields.weight'),
        value: `${coin.weightGrams}g`,
      })
    }

    if (coin.diameterMm != null && !banknote) {
      result.push({
        key: 'diameter',
        label: t('coinFields.diameter'),
        value: `${coin.diameterMm}mm`,
      })
    }

    if (coin.grade) {
      result.push({
        key: 'grade',
        label: banknote ? t('banknoteForm.grade') : t('coinFields.grade'),
        value: banknote ? banknoteGradeLabel(coin.grade) : coin.grade,
        valueClass: 'gold',
      })
    }

    // Purchase location (full-width last row, optional link)
    if (coin.purchaseLocation) {
      const safeUrl = sanitizeExternalUrl(coin.referenceUrl)
      result.push({
        key: 'purchaseLocation',
        label: '',
        value: coin.purchaseLocation,
        fullWidth: true,
        url: safeUrl,
      })
    }

    return result
  })

  return {
    rows,
  }
}
