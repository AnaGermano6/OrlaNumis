<template>
  <div ref="containerRef" class="container">
    <div v-if="store.loading && !coin" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="coin">
      <div class="md:sticky md:top-[61px] md:z-10 md:border-b md:border-border-subtle md:bg-surface md:px-0 md:py-3 md:shadow-[0_4px_12px_var(--overlay-30)]">
        <CoinDetailHeaderActions
          :is-wishlist="coin.isWishlist"
          :is-banknote="isBanknote"
          :coin-id="coin.id"
          :sharing="sharing"
          :duplicating="duplicating"
          @share="handleShare"
          @duplicate="handleDuplicate"
          @edit="handleEdit"
          @delete="handleDelete"
        />
      </div>

      <div class="mx-auto grid max-w-[1400px] grid-cols-1 items-start gap-8 md:grid-cols-[400px_minmax(0,1fr)]">
        <!-- T009-T011: Dual-side hero media -->
        <div class="order-1 self-start md:sticky md:top-[125px] md:h-fit">
          <!-- A banknote is wide: its sides are stacked and not zoomed (zooming crops them). -->
          <div class="grid gap-4" :class="isBanknote ? 'grid-cols-1' : 'grid-cols-2'" :data-kind="isBanknote ? 'banknote' : 'coin'">
            <div class="relative w-full overflow-hidden rounded-md border border-border-subtle bg-card" :class="isBanknote ? 'aspect-[2/1]' : 'aspect-square'">
              <AuthenticatedImage
                v-if="obverseImage"
                :media-path="obverseImage.filePath"
                :alt="sideLabel('obverse')"
                class="h-full w-full cursor-pointer object-contain transition-[opacity,transform] duration-200 hover:opacity-[0.85]"
                :class="isBanknote ? '' : 'scale-[1.28] hover:scale-[1.32]'"
                @click="openLightbox(obverseImage)"
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
                class="h-full w-full cursor-pointer object-contain transition-[opacity,transform] duration-200 hover:opacity-[0.85]"
                :class="isBanknote ? '' : 'scale-[1.28] hover:scale-[1.32]'"
                @click="openLightbox(reverseImage)"
              />
              <div v-else class="flex h-full w-full flex-col items-center justify-center gap-2">
                <span class="text-body font-medium text-text-secondary">{{ sideLabel('reverse') }}</span>
                <span class="text-sm italic text-text-muted">{{ t('common.noImage') }}</span>
              </div>
            </div>
          </div>
          <button
            v-if="!obverseImage || !reverseImage"
            type="button"
            class="btn btn-secondary btn-sm mt-3 w-full justify-center"
            @click="showRealPhotos = true"
          >
            {{ t('photos.findReal') }}
          </button>
          <!-- Official photos sometimes name the sides the other way round. -->
          <button
            v-else
            type="button"
            class="btn btn-ghost btn-sm mt-3 w-full justify-center"
            :disabled="swappingSides"
            @click="swapSides"
          >
            {{ t('photos.swapSides') }}
          </button>
          <div v-if="coin.isWishlist" class="mt-3">
            <button class="btn btn-primary w-full justify-center" @click="showPurchaseModal = true">
              {{ t('coinDetail.markPurchased') }}
            </button>
          </div>
        </div>

        <!-- Info -->
        <div class="order-2">
          <!-- T012: Title hierarchy -->
          <div class="mb-6">
            <h1 class="mt-2 font-display text-2xl font-semibold text-heading">{{ coin.name }}</h1>
            <p v-if="coin.ruler" class="mt-1 text-md text-text-secondary">{{ coin.ruler }}</p>
            <div v-if="coin.category || coinStatus?.commemorative || isBanknote" class="mt-3 flex flex-wrap gap-2">
              <span v-if="isBanknote" class="chip-sm" data-testid="banknote-badge">{{ t('banknotes.badge') }}</span>
              <span
                v-if="coin.category"
                class="badge"
                :style="{
                  backgroundColor: colorForLabelBackground(coin.category),
                  color: colorForLabel(coin.category),
                  border: `1px solid ${colorForLabelBackground(coin.category, 30)}`,
                }"
              >{{ categoryLabel(coin.category) }}</span>
              <!-- Commemorative coins, as Numista says of their type or as set by hand. -->
              <span
                v-if="coinStatus?.commemorative"
                class="badge"
                data-testid="commemorative-badge"
                :style="COMMEMORATIVE_BADGE_STYLE"
              >{{ t('coinDetail.badges.commemorative') }}</span>
              <span v-if="coin.isWishlist" class="chip-sm">{{ t('nav.wishlist') }}</span>
            </div>
          </div>

          <div v-if="coin.obverseInscription || coin.reverseInscription || coin.obverseDescription || coin.reverseDescription" class="mb-6">
            <h3 class="mb-3 font-display text-base font-medium text-heading">{{ t('coinFields.inscription') }}</h3>
            <div class="rounded-sm border border-border-subtle bg-card p-4">
              <div class="grid gap-6 md:grid-cols-2">
                <div v-if="coin.obverseInscription || coin.obverseDescription" class="flex flex-col gap-3">
                  <h4 class="m-0 text-base font-medium text-heading">{{ sideLabel('obverse') }}</h4>
                  <div v-if="coin.obverseInscription" class="flex flex-col gap-1">
                    <span class="section-label">{{ t('coinFields.inscriptionLabel') }}</span>
                    <span class="text-base italic text-text-secondary">{{ coin.obverseInscription }}</span>
                  </div>
                  <p v-if="coin.obverseDescription" class="m-0 text-base leading-6 text-text-secondary">{{ coin.obverseDescription }}</p>
                </div>
                <div v-if="coin.reverseInscription || coin.reverseDescription" class="flex flex-col gap-3">
                  <h4 class="m-0 text-base font-medium text-heading">{{ sideLabel('reverse') }}</h4>
                  <div v-if="coin.reverseInscription" class="flex flex-col gap-1">
                    <span class="section-label">{{ t('coinFields.inscriptionLabel') }}</span>
                    <span class="text-base italic text-text-secondary">{{ coin.reverseInscription }}</span>
                  </div>
                  <p v-if="coin.reverseDescription" class="m-0 text-base leading-6 text-text-secondary">{{ coin.reverseDescription }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- T014-T016: Metadata table -->
          <div v-if="metadataRows.length" class="mb-6">
            <h3 class="mb-3 font-display text-base font-medium text-heading">{{ t('common.details') }}</h3>
            <CoinDetailMetadataTable :rows="metadataRows" />
          </div>

          <CoinReferencesSection
            :coin-id="coin.id"
            :references="coin.references ?? []"
            :coin-name="coin.name"
            :coin-ruler="coin.ruler ?? ''"
            :coin-denomination="coin.denomination ?? ''"
            :coin-mint="coin.mint ?? ''"
            :coin-date-range="coin.dateRange ?? coin.era ?? ''"
            :coin-material="coin.material"
            :coin-obverse-inscription="coin.obverseInscription ?? ''"
            :coin-reverse-inscription="coin.reverseInscription ?? ''"
            @changed="refreshCoin"
          />

          <CoinMarketValueSection
            :coin-id="coin.id"
            :numista-type-id="numistaTypeId"
            :coin-grade="coin.grade"
          />

          <CoinSetsSection
            :sets="coin.sets ?? []"
            :coin-id="coin.id"
            @changed="refreshCoin"
          />

          <CoinListingStatus
            :coin-id="coin.id"
            :listing-status="coin.listingStatus"
            :listing-check-reason="coin.listingCheckReason"
            :listing-checked-at="coin.listingCheckedAt"
            @dismissed="refreshCoin"
          />
        </div>
      </div>
    </div>

    <RealPhotosModal
      v-if="showRealPhotos && coin"
      :coin-id="coin.id"
      :coin-name="coin.name"
      :numista-type-id="numistaTypeId"
      :details="coinPhotoDetails(coin)"
      :has-primary-image="Boolean(coin.images?.some((image) => image.isPrimary))"
      :kind="isBanknote ? 'banknote' : 'coin'"
      @close="showRealPhotos = false"
      @saved="refreshCoin"
    />
    <PurchaseModal v-if="showPurchaseModal && coin" :coin="coin" @close="showPurchaseModal = false" @confirm="confirmPurchase" />
    <ImageLightbox
      v-if="lightboxImage && coin"
      :coin-id="coin.id"
      :image-id="lightboxImage.id"
      :image-path="lightboxImage.filePath"
      :image-type="lightboxImage.imageType"
      @close="lightboxImage = null"
      @saved="handleImageSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCoinsStore } from '@/stores/coins'
import PurchaseModal from '@/components/PurchaseModal.vue'
import ImageLightbox from '@/components/ImageLightbox.vue'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import CoinDetailHeaderActions from '@/components/coin/CoinDetailHeaderActions.vue'
import CoinSetsSection from '@/components/coin/CoinSetsSection.vue'
import CoinDetailMetadataTable from '@/components/coin/CoinDetailMetadataTable.vue'
import CoinListingStatus from '@/components/coin/CoinListingStatus.vue'
import CoinReferencesSection from '@/components/coin/CoinReferencesSection.vue'
import CoinMarketValueSection from '@/components/coin/CoinMarketValueSection.vue'
import RealPhotosModal from '@/components/photos/RealPhotosModal.vue'
import { coinPhotoDetails } from '@/utils/coinPhotos'
import { deleteCoin, duplicateCoin, getCoinStatus, purchaseCoin, swapCoinImageSides, type CoinStatus } from '@/api/client'
import { colorForLabel, colorForLabelBackground } from '@/utils/categoryColor'
import { useDialog } from '@/composables/useDialog'
import { useCoinDetailMetadataRows } from '@/composables/useCoinDetailMetadataRows'
import { useCoinShareCard } from '@/composables/useCoinShareCard'
import { useCoinDetailSwipeNav } from '@/composables/useCoinDetailSwipeNav'
import type { CoinImage } from '@/types'
import { t } from '@/i18n'
import { categoryLabel } from '@/i18n/coinVocabulary'

const { showConfirm, showAlert } = useDialog()
const route = useRoute()
const router = useRouter()
const store = useCoinsStore()

const showPurchaseModal = ref(false)
const lightboxImage = ref<CoinImage | null>(null)
const { sharing, shareCoinCard } = useCoinShareCard()
const duplicating = ref(false)

const containerRef = ref<HTMLElement | null>(null)

// Swipe navigation is suppressed while any non-teleported modal is open.
const swipeEnabled = computed(() => !showPurchaseModal.value)
useCoinDetailSwipeNav(containerRef, { enabled: swipeEnabled })

const coin = computed(() => store.currentCoin)
const isBanknote = computed(() => coin.value?.kind === 'banknote')
// A banknote has a front and a back, a coin an obverse and a reverse.
function sideLabel(side: 'obverse' | 'reverse'): string {
  if (isBanknote.value) return side === 'obverse' ? t('banknoteForm.front') : t('banknoteForm.back')
  return side === 'obverse' ? t('common.obverse') : t('common.reverse')
}
// Numista type linked through a "Numista" catalogue reference, if any.
const numistaTypeId = computed(() => {
  const reference = coin.value?.references?.find((item) => item.catalog.trim().toLowerCase() === 'numista')
  const id = Number(reference?.number)
  return Number.isInteger(id) && id > 0 ? id : null
})

// T010: Deterministic media slot logic
const obverseImage = computed(() => coin.value?.images?.find(i => i.imageType === 'obverse') ?? null)
const reverseImage = computed(() => coin.value?.images?.find(i => i.imageType === 'reverse') ?? null)

// Whether the coin is still legal tender and whether it is commemorative.
const coinStatus = ref<CoinStatus | null>(null)
const COMMEMORATIVE_BADGE_STYLE = {
  backgroundColor: 'color-mix(in srgb, var(--accent-gold) 20%, transparent)',
  color: 'var(--accent-gold)',
  border: '1px solid color-mix(in srgb, var(--accent-gold) 30%, transparent)',
}
watch(() => coin.value?.id, async (id) => {
  coinStatus.value = null
  if (!id) return
  try {
    const { data } = await getCoinStatus(id)
    if (coin.value?.id === id) coinStatus.value = data
  } catch {
    // The details show without it.
  }
}, { immediate: true })

// T015: Metadata rows
const metadataRows = computed(() => {
  if (!coin.value) return []
  return useCoinDetailMetadataRows(coin.value, coinStatus.value).rows.value
})
onMounted(async () => {
  await store.fetchCoin(Number(route.params.id))
})

function refreshCoin() {
  if (coin.value) {
    store.fetchCoin(coin.value.id)
  }
}

const showRealPhotos = ref(false)
const swappingSides = ref(false)

async function swapSides() {
  if (!coin.value || swappingSides.value) return
  swappingSides.value = true
  try {
    await swapCoinImageSides(coin.value.id)
    await store.fetchCoin(coin.value.id)
  } catch {
    await showAlert(t('photos.swapSidesFailed'), { title: t('common.error') })
  } finally {
    swappingSides.value = false
  }
}

function openLightbox(image: CoinImage) {
  lightboxImage.value = image
}

function handleImageSaved() {
  refreshCoin()
}

async function handleShare() {
  if (!coin.value) return
  await shareCoinCard(coin.value)
}

async function confirmPurchase(data: { purchasePrice?: number; purchaseDate?: string; purchaseLocation?: string }) {
  if (!coin.value) return
  try {
    await purchaseCoin(coin.value.id, data)
    showPurchaseModal.value = false
    await store.fetchCoin(coin.value.id)
  } catch {
    showPurchaseModal.value = false
  }
}

async function handleDelete() {
  if (!coin.value || !await showConfirm(t('coinDetail.confirmDelete'), { title: t('coinDetail.deleteTitle'), variant: 'danger' })) return
  const coinId = coin.value.id
  await deleteCoin(coinId)
  // Back to where it was listed: the banknotes, or the coin gallery.
  router.push(isBanknote.value ? '/banknotes' : '/')
}

async function handleDuplicate() {
  if (!coin.value || duplicating.value) return
  duplicating.value = true
  try {
    const res = await duplicateCoin(coin.value.id)
    router.push(`/coin/${res.data.id}`)
  } catch {
    await showAlert(t('coinDetail.duplicateFailed'), { title: t('common.error') })
  } finally {
    duplicating.value = false
  }
}

function handleEdit() {
  if (!coin.value) return
  router.push(`/edit/${coin.value.id}`)
}
</script>
