<template>
  <section class="flex flex-col" :aria-label="t('collectionOverview.tabs')" data-testid="collection-overview-tabs">
    <p v-if="failed" class="text-body text-text-secondary" role="alert">{{ t('collectionOverview.failed') }}</p>

    <!-- The whole collection, its coins or its banknotes, one at a time. -->
    <AppTabs v-model="tab" class="mb-5" :tabs="tabItems" id-prefix="overview" :label="t('collectionOverview.tabs')" />

    <div v-show="tab === 'all'" id="overview-panel-all" class="flex flex-col gap-4" role="tabpanel" aria-labelledby="overview-tab-all">
      <OverviewKpis :summary="all" kind="all" />
    </div>

    <div v-show="tab === 'coins'" id="overview-panel-coins" class="flex flex-col gap-4" role="tabpanel" aria-labelledby="overview-tab-coins">
      <OverviewKpis :summary="coins" kind="coin" extinct-link :wishlist="coinWishlist" @show-extinct="openGallery({ country: EXTINCT_COUNTRIES })" />
      <CollectionDistribution :summary="coins" @open="openGallery" />
      <CollectionHighlights :summary="coins" :title="t('collectionOverview.coins.highlights')" />
    </div>

    <div v-show="tab === 'banknotes'" id="overview-panel-banknotes" class="flex flex-col gap-4" role="tabpanel" aria-labelledby="overview-tab-banknotes">
      <OverviewKpis :summary="banknotes" kind="banknote" />
      <template v-if="banknotes && banknotes.coins > 0">
        <CollectionDistribution kind="banknote" :summary="banknotes" @open="openBanknotes" />
        <CollectionHighlights :summary="banknotes" kind="banknote" :title="t('collectionOverview.banknotes.highlights')" />
      </template>
      <p v-else-if="banknotes" class="m-0 text-body text-text-secondary" data-testid="overview-no-banknotes">
        {{ t('banknotes.empty.title') }}.
        <router-link to="/banknotes/add" class="text-gold underline">{{ t('banknotes.add') }}</router-link>
      </p>
    </div>
  </section>
</template>

<script setup lang="ts">
// The collection at a glance, in three tabs: the whole collection, its coins
// and its banknotes. Each has three headline numbers (how many pieces and how
// many different ones, what they are worth against their face value, how
// many countries); then the coins by kind of coin and the banknotes by
// circulation, whose parts open the coin gallery or the banknotes filtered,
// and their highlights (those held more than once are on the wishlist's
// page). The tab chosen is kept for the next visit. At the top of the
// statistics.
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { getCoins, getCollectionSummary, type CollectionSummary } from '@/api/client'
import AppTabs from '@/components/AppTabs.vue'
import CollectionDistribution, { type DistributionFilter } from '@/components/collection/CollectionDistribution.vue'
import CollectionHighlights from '@/components/collection/CollectionHighlights.vue'
import OverviewKpis from '@/components/collection/OverviewKpis.vue'
import { useOpenGallery } from '@/composables/useOpenGallery'
import { t } from '@/i18n'
import { EXTINCT_COUNTRIES } from '@/utils/countryFilter'

const router = useRouter()

const all = ref<CollectionSummary | null>(null)
const coins = ref<CollectionSummary | null>(null)
const banknotes = ref<CollectionSummary | null>(null)
/** The coins on the wishlist (the banknotes' are not counted). */
const coinWishlist = ref<number | null>(null)
const failed = ref(false)

type OverviewTab = 'all' | 'coins' | 'banknotes'
const TAB_KEY = 'overview:tab'

// The tab chosen last time, when the browser kept it.
function savedTab(): OverviewTab {
  try {
    const saved = localStorage.getItem(TAB_KEY)
    return saved === 'coins' || saved === 'banknotes' ? saved : 'all'
  } catch {
    return 'all'
  }
}

const tab = ref<OverviewTab>(savedTab())
watch(tab, (value) => {
  try {
    localStorage.setItem(TAB_KEY, value)
  } catch {
    // A private window, or storage turned off: the tab is not kept.
  }
})

const tabItems = computed(() => [
  { id: 'all' as const, label: t('collectionOverview.all.title') },
  { id: 'coins' as const, label: t('collectionOverview.coins.title') },
  { id: 'banknotes' as const, label: t('collectionOverview.banknotes.title') },
])

// A part of a coin card opens the gallery with only its filter on.
const openGallery = useOpenGallery()
// A part of the banknotes' card opens those still legal tender, or no longer.
function openBanknotes(filter: DistributionFilter) {
  void router.push({ name: 'banknotes', query: filter.circulation ? { circulation: filter.circulation } : {} })
}

// Loads the three summaries again (the statistics' pull to refresh).
async function load() {
  const [allResult, coinResult, banknoteResult, wishlistResult] = await Promise.allSettled([
    getCollectionSummary(),
    getCollectionSummary('coin'),
    getCollectionSummary('banknote'),
    getCoins({ kind: 'coin', wishlist: 'true', page: 1, limit: 1 }),
  ])
  if (allResult.status === 'fulfilled') all.value = allResult.value.data
  if (coinResult.status === 'fulfilled') coins.value = coinResult.value.data
  if (banknoteResult.status === 'fulfilled') banknotes.value = banknoteResult.value.data
  if (wishlistResult.status === 'fulfilled') coinWishlist.value = wishlistResult.value.data.total
  failed.value = [allResult, coinResult, banknoteResult].some((result) => result.status === 'rejected')
}

onMounted(load)
defineExpose({ load })
</script>
