import { readonly, ref } from 'vue'
import type { Coin } from '@/types'
import { getShareCardFilename, getShareImageUrls, renderCoinShareCard } from '@/utils/coinShareCard'
import type { CoinShareCardContext } from '@/utils/coinShareCard'
import { useDialog } from '@/composables/useDialog'
import { t } from '@/i18n'

export type CoinShareResultMode = 'shared' | 'downloaded'

export interface CoinShareResult {
  mode: CoinShareResultMode
}

export interface CoinShareCardOptions {
  context?: CoinShareCardContext
}

interface ShareNavigator {
  canShare?: (data?: ShareData) => boolean
  share?: (data: ShareData) => Promise<void>
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  anchor.rel = 'noopener'
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  URL.revokeObjectURL(url)
}

export function useCoinShareCard() {
  const sharing = ref(false)
  const { showAlert } = useDialog()

  async function shareCoinCard(coin: Coin, options: CoinShareCardOptions = {}): Promise<CoinShareResult> {
    sharing.value = true
    try {
      const imageUrls = getShareImageUrls(coin)
      const blob = await renderCoinShareCard({
        coin,
        imageUrl: imageUrls[0] ?? null,
        imageUrls,
        appName: t('app.brandFull'),
        context: options.context,
      })
      const filename = getShareCardFilename(coin)
      const file = new File([blob], filename, { type: 'image/png' })
      const contextSummary = options.context?.summary?.trim()
      const navigatorWithShare = navigator as unknown as ShareNavigator
      const shareData: ShareData = {
        files: [file],
        title: coin.name || t('shareCard.fallbackTitle'),
        text: contextSummary
          ? `${contextSummary}\n\n${t('shareCard.sharedFrom', { app: t('app.brandFull') })}`
          : t('shareCard.sharedFrom', { app: t('app.brandFull') }),
      }

      if (navigatorWithShare.share && navigatorWithShare.canShare?.(shareData)) {
        await navigatorWithShare.share(shareData)
        return { mode: 'shared' }
      }

      downloadBlob(blob, filename)
      return { mode: 'downloaded' }
    } catch (error) {
      const message = error instanceof Error ? error.message : t('shareCard.generateFailed')
      await showAlert(message, { title: t('shareCard.shareFailedTitle') })
      throw error
    } finally {
      sharing.value = false
    }
  }

  return {
    sharing: readonly(sharing),
    shareCoinCard,
  }
}
