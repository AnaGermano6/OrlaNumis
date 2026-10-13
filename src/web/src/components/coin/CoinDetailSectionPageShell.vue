<template>
  <div ref="containerRef" class="container">
    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="coin" class="mx-auto max-w-[900px]">
      <div class="mb-4 grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3">
        <button class="btn btn-ghost btn-sm inline-flex items-center gap-1" @click="navigateToOverview">
          <ChevronLeft :size="16" />
          {{ t('coinDetail.backToOverview') }}
        </button>
        <CoinDetailOverflowMenu
          :coin-id="coin.id"
          :is-banknote="isBanknote"
          :duplicating="duplicating"
          @duplicate="handleDuplicate"
        />
      </div>

      <!-- A banknote is wide: its front and back are stacked. -->
      <div class="mb-4 grid gap-4" :class="isBanknote ? 'grid-cols-1' : 'grid-cols-2'">
        <div class="relative w-full overflow-hidden rounded-md border border-border-subtle bg-card" :class="isBanknote ? 'aspect-[2/1]' : 'aspect-square'">
          <AuthenticatedImage
            v-if="obverseImage"
            :media-path="obverseImage.filePath"
            :alt="sideLabel('obverse')"
            class="h-full w-full object-contain"
          />
          <div v-else class="flex h-full w-full flex-col items-center justify-center gap-2">
            <span class="text-body font-medium text-text-secondary">{{ sideLabel('obverse') }}</span>
            <span class="text-sm italic text-text-muted">{{ t('common.noImage') }}</span>
          </div>
        </div>
        <div class="relative w-full overflow-hidden rounded-md border border-border-subtle bg-card" :class="isBanknote ? 'aspect-[2/1]' : 'aspect-square'">
          <AuthenticatedImage
            v-if="reverseImage"
            :media-path="reverseImage.filePath"
            :alt="sideLabel('reverse')"
            class="h-full w-full object-contain"
          />
          <div v-else class="flex h-full w-full flex-col items-center justify-center gap-2">
            <span class="text-body font-medium text-text-secondary">{{ sideLabel('reverse') }}</span>
            <span class="text-sm italic text-text-muted">{{ t('common.noImage') }}</span>
          </div>
        </div>
      </div>

      <!-- Coin identity banner -->
      <div class="mb-6 border-b border-border-subtle pb-4">
        <h1 class="mb-1 text-2xl font-medium text-heading">{{ coin.name }}</h1>
        <p v-if="coin.ruler" class="text-base text-text-secondary">{{ coin.ruler }}</p>
      </div>

      <!-- Section title -->
      <h2 class="mb-6 text-lg font-medium text-heading">{{ sectionTitle }}</h2>

      <!-- Content slot -->
      <div>
        <slot :coin="coin" :refresh="refreshCoin"></slot>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useCoinDetailSwipeNav } from '@/composables/useCoinDetailSwipeNav'
import { useRouter } from 'vue-router'
import { ChevronLeft } from 'lucide-vue-next'
import { useCoinDetailContext } from '@/composables/useCoinDetailContext'
import { useDialog } from '@/composables/useDialog'
import { duplicateCoin } from '@/api/client'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import CoinDetailOverflowMenu from '@/components/coin/CoinDetailOverflowMenu.vue'
import { t } from '@/i18n'

defineProps<{
  sectionTitle: string
}>()

const { coin, loading, refreshCoin, navigateToOverview } = useCoinDetailContext()
const obverseImage = computed(() => coin.value?.images?.find((image) => image.imageType === 'obverse') ?? null)
const reverseImage = computed(() => coin.value?.images?.find((image) => image.imageType === 'reverse') ?? null)
// A banknote has a front and a back, a coin an obverse and a reverse.
const isBanknote = computed(() => coin.value?.kind === 'banknote')
function sideLabel(side: 'obverse' | 'reverse'): string {
  if (isBanknote.value) return side === 'obverse' ? t('banknoteForm.front') : t('banknoteForm.back')
  return side === 'obverse' ? t('common.obverse') : t('common.reverse')
}
const duplicating = ref(false)

const containerRef = ref<HTMLElement | null>(null)

useCoinDetailSwipeNav(containerRef)
const router = useRouter()
const { showAlert } = useDialog()

async function handleDuplicate() {
  if (!coin.value || duplicating.value) return
  duplicating.value = true
  try {
    const res = await duplicateCoin(coin.value.id)
    await router.push(`/coin/${res.data.id}`)
  } catch {
    await showAlert(t('coinDetail.duplicateFailed'), { title: t('common.error') })
  } finally {
    duplicating.value = false
  }
}
</script>
