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

// "Corrente · Comemorativa", "Desmonetizada em 28/02/2002"…; '' when nothing is known.
function statusValue(status: CoinStatus | null | undefined): string {
  if (!status) return ''
  const parts: string[] = []
  if (status.current === true) parts.push(t('coinDetail.currentCoin'))
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

    if (coin.era) {
      result.push({
        key: 'era',
        label: t('coinFields.era'),
        value: eraLabel(coin.era),
      })
    }

    if (coin.mint) {
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
        value: materialLabel(coin.material),
      })
    }

    if (coin.weightGrams != null) {
      result.push({
        key: 'weight',
        label: t('coinFields.weight'),
        value: `${coin.weightGrams}g`,
      })
    }

    if (coin.diameterMm != null) {
      result.push({
        key: 'diameter',
        label: t('coinFields.diameter'),
        value: `${coin.diameterMm}mm`,
      })
    }

    if (coin.grade) {
      result.push({
        key: 'grade',
        label: t('coinFields.grade'),
        value: coin.grade,
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
