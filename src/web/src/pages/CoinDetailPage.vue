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

      <div class="mx-auto flex max-w-[1000px] flex-col pt-6">
        <!-- The coin as a museum label: both sides with their names, its
             name, its key facts in a row and what is said of it. -->
        <article class="card mb-6" data-testid="coin-summary">
          <!-- A banknote is wide: its sides are side by side on a wider
               screen, stacked on a phone, and not zoomed (zooming crops them). -->
          <div class="grid gap-6" :class="isBanknote ? 'grid-cols-1 md:grid-cols-2' : 'grid-cols-2'" :data-kind="isBanknote ? 'banknote' : 'coin'">
            <figure v-for="side in SIDES" :key="side" class="m-0 flex min-w-0 flex-col items-center gap-3" :data-testid="`coin-side-${side}`">
              <div class="relative w-full overflow-hidden rounded-md" :class="isBanknote ? 'aspect-[2/1]' : 'aspect-square max-w-[320px]'">
                <AuthenticatedImage
                  v-if="sideImage(side)"
                  :media-path="sideImage(side)!.filePath"
                  :alt="sideLabel(side)"
                  class="h-full w-full cursor-pointer object-contain transition-[opacity,transform] duration-200 hover:opacity-[0.85]"
                  :class="isBanknote ? '' : 'scale-[1.28] hover:scale-[1.32]'"
                  @click="openLightbox(sideImage(side)!)"
                />
                <div v-else class="flex h-full w-full items-center justify-center rounded-md border border-dashed border-border-subtle">
                  <span class="text-sm italic text-text-muted">{{ t('common.noImage') }}</span>
                </div>
              </div>
              <figcaption class="text-body text-text-secondary">{{ sideLabel(side) }}</figcaption>
            </figure>
          </div>
          <div class="mt-4 flex flex-wrap justify-center gap-2">
            <button
              v-if="!obverseImage || !reverseImage"
              type="button"
              class="btn btn-secondary btn-sm"
              @click="showRealPhotos = true"
            >
              {{ t('photos.findReal') }}
            </button>
            <!-- Official photos sometimes name the sides the other way round. -->
            <button
              v-else
              type="button"
              class="btn btn-ghost btn-sm"
              :disabled="swappingSides"
              @click="swapSides"
            >
              {{ t('photos.swapSides') }}
            </button>
            <button v-if="coin.isWishlist" class="btn btn-primary btn-sm" @click="showPurchaseModal = true">
              {{ t('coinDetail.markPurchased') }}
            </button>
          </div>

          <!-- T012: Title hierarchy -->
          <h1 class="mb-0 mt-6 font-display text-2xl font-semibold text-heading">{{ coin.name }}</h1>
          <p v-if="coin.ruler" class="mb-0 mt-1 text-md text-text-secondary">{{ coin.ruler }}</p>
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

          <!-- Its details, inscriptions, history, catalogue references (with its
               market value) and sets, one at a time, so the page keeps one length. The rule under the tabs
               is a shadow: a border would make the bar scroll by a pixel. -->
          <div class="mt-6 flex overflow-x-auto shadow-[inset_0_-1px_0_var(--border-subtle)]" role="tablist" :aria-label="t('coinDetail.tabs.label')" @keydown="onTabKeydown">
            <button
              v-for="item in tabs"
              :id="`coin-tab-${item}`"
              :key="item"
              type="button"
              role="tab"
              class="shrink-0 whitespace-nowrap border-b-2 px-4 py-3 text-base transition-colors md:flex-1"
              :class="item === activeTab ? 'border-gold font-medium text-text-primary' : 'border-transparent text-text-secondary hover:text-text-primary'"
              :aria-selected="item === activeTab"
              :aria-controls="`coin-panel-${item}`"
              :tabindex="item === activeTab ? 0 : -1"
              :data-testid="`coin-tab-${item}`"
              @click="tab = item"
            >
              {{ tabLabel(item) }}
            </button>
          </div>
          <div class="min-h-[14rem] pt-5">
            <div v-if="hasInscription" v-show="activeTab === 'inscription'" id="coin-panel-inscription" role="tabpanel" aria-labelledby="coin-tab-inscription">
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

            <div v-show="activeTab === 'details'" id="coin-panel-details" role="tabpanel" aria-labelledby="coin-tab-details">
              <!-- Its metal, weight, diameter and grade (a banknote's material,
                   size and grade) in a row, ruled above and below; then the others. -->
              <dl v-if="keyFacts.length" class="m-0 grid grid-cols-2 border-y border-border-subtle" :class="KEY_FACT_COLUMNS[keyFacts.length]" data-testid="coin-key-facts">
                <div v-for="(fact, index) in keyFacts" :key="fact.key" class="flex min-w-0 flex-col gap-1 border-border-subtle py-3" :class="keyFactRules(index)">
                  <dt class="text-body text-text-secondary">{{ fact.label }}</dt>
                  <dd class="m-0 text-md text-text-primary" :class="fact.valueClass === 'gold' ? 'text-gold' : ''">{{ fact.value }}</dd>
                </div>
              </dl>
              <!-- T014-T016: Metadata table, without the key facts. -->
              <CoinDetailMetadataTable v-if="otherRows.length" class="mt-4" :rows="otherRows" />
            </div>

            <!-- What the collector wrote of it. -->
            <div v-if="coin.notes" v-show="activeTab === 'history'" id="coin-panel-history" role="tabpanel" aria-labelledby="coin-tab-history">
              <p class="m-0 whitespace-pre-wrap text-base leading-7 text-text-secondary" data-testid="coin-notes">{{ coin.notes }}</p>
            </div>

            <div v-show="activeTab === 'references'" id="coin-panel-references" role="tabpanel" aria-labelledby="coin-tab-references">
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
            </div>

            <div v-show="activeTab === 'sets'" id="coin-panel-sets" role="tabpanel" aria-labelledby="coin-tab-sets">
              <CoinSetsSection
                :sets="coin.sets ?? []"
                :coin-id="coin.id"
                :coin="coin"
                :status="coinStatus"
                @changed="refreshCoin"
              />
            </div>
          </div>
        </article>

        <!-- Whether the listing it was bought from is still for sale. -->
        <div>
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
const SIDES = ['obverse', 'reverse'] as const
function sideImage(side: (typeof SIDES)[number]): CoinImage | null {
  return side === 'obverse' ? obverseImage.value : reverseImage.value
}

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

// The tabs, in this order; the inscriptions and the history only when the
// coin has some. The tab chosen stays as the previous and next coins open;
// the address's "#catalog-references" opens the references.
const TABS = ['details', 'inscription', 'history', 'references', 'sets'] as const
type CoinTab = (typeof TABS)[number]
const hasInscription = computed(() => Boolean(coin.value?.obverseInscription || coin.value?.reverseInscription || coin.value?.obverseDescription || coin.value?.reverseDescription))
const tabs = computed(() => TABS.filter((item) => (item === 'inscription' ? hasInscription.value : item === 'history' ? Boolean(coin.value?.notes) : true)))
const tab = ref<CoinTab>('details')
const activeTab = computed<CoinTab>(() => (tabs.value.includes(tab.value) ? tab.value : 'details'))
watch(() => route.hash, (hash) => {
  if (hash === '#catalog-references') tab.value = 'references'
}, { immediate: true })

function tabLabel(item: CoinTab): string {
  switch (item) {
    case 'inscription': return t('coinFields.inscription')
    case 'details': return t('common.details')
    case 'history': return t('coinDetail.tabs.history')
    case 'references': return t('coinDetail.references.title')
    default: return t('coinDetail.sets.title')
  }
}

// The arrow keys go from tab to tab.
function onTabKeydown(event: KeyboardEvent) {
  const step = event.key === 'ArrowRight' ? 1 : event.key === 'ArrowLeft' ? -1 : 0
  if (!step) return
  event.preventDefault()
  const list = tabs.value
  const next = list[(list.indexOf(activeTab.value) + step + list.length) % list.length]!
  tab.value = next
  document.getElementById(`coin-tab-${next}`)?.focus()
}

// The key facts, in a row at the top of the details; the table has the others.
const KEY_FACTS: Record<'coin' | 'banknote', string[]> = {
  coin: ['material', 'weight', 'diameter', 'grade'],
  banknote: ['material', 'dimensions', 'grade'],
}
const keyFacts = computed(() => {
  const keys = KEY_FACTS[isBanknote.value ? 'banknote' : 'coin']
  return keys.map((key) => metadataRows.value.find((row) => row.key === key)).filter((row) => row !== undefined)
})
const otherRows = computed(() => metadataRows.value.filter((row) => !keyFacts.value.includes(row)))
// As many columns as facts on a wider screen; two on a phone.
const KEY_FACT_COLUMNS: Record<number, string> = { 1: 'md:grid-cols-1', 2: 'md:grid-cols-2', 3: 'md:grid-cols-3', 4: 'md:grid-cols-4' }

// The rules between the facts: one before each fact but the first of its
// row (two a row on a phone, all in one on a wider screen), and one above
// the phone's second row.
function keyFactRules(index: number): string[] {
  return [
    index % 2 === 1 ? 'border-l pl-4' : index > 0 ? 'md:border-l md:pl-4' : '',
    index >= 2 ? 'max-md:border-t' : '',
    'pr-4',
  ]
}
onMounted(async () => {
  await store.fetchCoin(Number(route.params.id))
})

// Going to the previous or the next coin keeps this page and changes the id.
watch(() => route.params.id, (id, previous) => {
  if (!id || id === previous || route.name !== 'coin-detail') return
  lightboxImage.value = null
  showRealPhotos.value = false
  void store.fetchCoin(Number(id))
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
