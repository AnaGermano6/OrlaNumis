<template>
  <div class="container">
    <DesktopPageContext :title="t('collectionOverview.title')">
      <header class="page-header">
        <h1>{{ t('collectionOverview.title') }}</h1>
      </header>
    </DesktopPageContext>

    <p v-if="failed" class="text-body text-text-secondary" role="alert">{{ t('collectionOverview.failed') }}</p>

    <!-- The whole collection: coins and banknotes together -->
    <section class="mb-8" aria-labelledby="overview-all-title">
      <h2 id="overview-all-title" class="mb-3 font-display text-lg font-medium text-heading">{{ t('collectionOverview.all.title') }}</h2>
      <div class="grid grid-cols-2 gap-3 md:grid-cols-4">
        <div
          v-for="kpi in totals"
          :key="kpi.key"
          class="flex min-w-0 flex-col items-center gap-1 rounded-md border border-border-subtle bg-card px-2 py-4 text-center md:px-4"
          :data-testid="`overview-kpi-${kpi.key}`"
        >
          <!-- A long amount gets a smaller size on phones, to fit half the row. -->
          <span class="whitespace-nowrap font-display font-semibold text-text-primary md:text-2xl" :class="kpi.display.length > 10 ? 'text-base' : 'text-xl'">{{ kpi.display }}</span>
          <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ kpi.label }}</span>
        </div>
      </div>
    </section>

    <!-- Coins: the gallery's numbers and highlights -->
    <section class="mb-8" aria-labelledby="overview-coins-title">
      <div class="mb-3 flex flex-wrap items-baseline justify-between gap-2">
        <h2 id="overview-coins-title" class="m-0 font-display text-lg font-medium text-heading">{{ t('collectionOverview.coins.title') }}</h2>
        <router-link to="/" class="text-body text-gold underline-offset-2 hover:underline" data-testid="overview-open-gallery">{{ t('collectionOverview.coins.open') }}</router-link>
      </div>
      <CollectionKpis
        :summary="coins"
        @toggle-current="openGallery({ circulation: 'current' })"
        @toggle-commemorative="openGallery({ commemorative: 'true' })"
        @show-coincards="openGallery({ presentation: 'coincard' })"
        @show-extinct="openGallery({ country: EXTINCT_COUNTRIES })"
      />
      <CollectionHighlights :summary="coins" :title="t('collectionOverview.coins.highlights')" />
      <CollectionRepeated kind="coin" :summary="coins" @open="openGallery({ repeated: true })" />
      <CollectionMissingPhotos />
    </section>

    <!-- Banknotes: their numbers and highlights -->
    <section class="mb-8" aria-labelledby="overview-banknotes-title">
      <div class="mb-3 flex flex-wrap items-baseline justify-between gap-2">
        <h2 id="overview-banknotes-title" class="m-0 font-display text-lg font-medium text-heading">{{ t('collectionOverview.banknotes.title') }}</h2>
        <router-link to="/banknotes" class="text-body text-gold underline-offset-2 hover:underline" data-testid="overview-open-banknotes">{{ t('collectionOverview.banknotes.open') }}</router-link>
      </div>
      <BanknoteKpis :summary="banknotes" />
      <CollectionHighlights :summary="banknotes" kind="banknote" :title="t('collectionOverview.banknotes.highlights')" />
      <CollectionRepeated v-if="banknotes && banknotes.coins > 0" kind="banknote" :summary="banknotes" @open="router.push({ name: 'banknotes', query: { repeated: 'true' } })" />
      <p v-if="banknotes && banknotes.coins === 0" class="m-0 text-body text-text-secondary" data-testid="overview-no-banknotes">
        {{ t('banknotes.empty.title') }}.
        <router-link to="/banknotes/add" class="text-gold underline">{{ t('banknotes.add') }}</router-link>
      </p>
    </section>

    <CoinReleases />
  </div>
</template>

<script setup lang="ts">
// The collection at a glance: how many items (counting quantities, and each
// different one once), countries and how much they are worth altogether, with
// a coin and a banknote of the day; then the coins (the gallery's numbers and
// highlights, whose cards open the gallery filtered) and the banknotes, each
// with the ones held more than once; last, new and upcoming coins.
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getCollectionSummary, type CollectionSummary } from '@/api/client'
import BanknoteKpis from '@/components/collection/BanknoteKpis.vue'
import CollectionHighlights from '@/components/collection/CollectionHighlights.vue'
import CollectionKpis from '@/components/collection/CollectionKpis.vue'
import CollectionMissingPhotos from '@/components/collection/CollectionMissingPhotos.vue'
import CollectionRepeated from '@/components/collection/CollectionRepeated.vue'
import CoinReleases from '@/components/collection/CoinReleases.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import { useCoinsStore } from '@/stores/coins'
import { intlLocale, t } from '@/i18n'
import { EXTINCT_COUNTRIES } from '@/utils/countryFilter'
import { formatCurrency } from '@/utils/format'

const router = useRouter()
const store = useCoinsStore()

const all = ref<CollectionSummary | null>(null)
const coins = ref<CollectionSummary | null>(null)
const banknotes = ref<CollectionSummary | null>(null)
const failed = ref(false)

function count(value: number | undefined): string {
  return value === undefined ? '—' : new Intl.NumberFormat(intlLocale()).format(value)
}

const totals = computed(() => [
  { key: 'items', display: count(all.value?.coins), label: t('collectionOverview.all.items') },
  { key: 'unique', display: count(all.value ? all.value.uniqueCoins ?? all.value.coins : undefined), label: t('collectionOverview.all.unique') },
  { key: 'countries', display: count(all.value?.countries), label: t('collectionOverview.all.countries') },
  {
    key: 'value',
    display: all.value ? formatCurrency(all.value.totalValue) : '—',
    label: t('collectionOverview.all.value'),
  },
])

// A coin card opens the gallery with only that card's filter on.
function openGallery(filter: { circulation?: string; commemorative?: string; presentation?: string; country?: string; repeated?: boolean }) {
  store.searchQuery = ''
  store.selectedDenomination = ''
  store.selectedRarity = ''
  store.selectedClassification = ''
  store.selectedCirculation = filter.circulation ?? ''
  store.selectedCommemorative = filter.commemorative ?? ''
  store.selectedPresentation = filter.presentation ?? ''
  store.selectedCountry = filter.country ?? ''
  store.selectedRepeated = filter.repeated ?? false
  void router.push('/')
}

onMounted(async () => {
  const [allResult, coinResult, banknoteResult] = await Promise.allSettled([
    getCollectionSummary(),
    getCollectionSummary('coin'),
    getCollectionSummary('banknote'),
  ])
  if (allResult.status === 'fulfilled') all.value = allResult.value.data
  if (coinResult.status === 'fulfilled') coins.value = coinResult.value.data
  if (banknoteResult.status === 'fulfilled') banknotes.value = banknoteResult.value.data
  failed.value = [allResult, coinResult, banknoteResult].some((result) => result.status === 'rejected')
})
</script>
