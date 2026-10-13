<template>
  <div class="container">
    <DesktopPageContext :title="t('sets.catalogue.title')">
      <header class="page-header">
        <h1>{{ t('sets.catalogue.title') }}</h1>
      </header>
    </DesktopPageContext>
    <p class="mb-1 mt-0 text-body text-text-secondary">{{ t('sets.catalogue.intro') }}</p>
    <p class="mb-4 mt-0 text-chip text-text-muted">{{ t('sets.catalogue.sources') }}</p>

    <p v-if="failed" class="text-body text-text-secondary" role="alert">{{ t('sets.catalogue.failed') }}</p>
    <p v-else-if="!countries.length" class="text-body text-text-muted" data-testid="catalogue-loading">{{ t('sets.catalogue.loading') }}</p>
    <template v-else>
      <h2 class="sr-only">{{ t('sets.catalogue.countries') }}</h2>
      <!-- On a phone, a list to choose from; the grid of countries above that. -->
      <select
        class="form-select w-full sm:hidden"
        :value="selected"
        :aria-label="t('sets.catalogue.countries')"
        data-testid="catalogue-country-select"
        @change="choose(($event.target as HTMLSelectElement).value)"
      >
        <option v-for="country in sortedCountries" :key="country.country" :value="country.country">
          {{ releaseCountryName(country.country) }} — {{ t('sets.catalogue.progress', { owned: owned(country), total: total(country) }) }}
        </option>
      </select>
      <ul class="m-0 hidden list-none grid-cols-2 gap-2 p-0 sm:grid sm:grid-cols-3 lg:grid-cols-5">
        <li v-for="country in sortedCountries" :key="country.country">
          <button
            type="button"
            class="w-full rounded-md border bg-card p-3 text-left transition-colors hover:border-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
            :class="country.country === selected ? 'border-gold' : 'border-border-subtle'"
            :aria-pressed="country.country === selected"
            :data-testid="`catalogue-country-${country.country}`"
            @click="choose(country.country)"
          >
            <span class="block truncate font-medium text-text-primary">{{ releaseCountryName(country.country) }}</span>
            <span class="block text-chip text-text-muted">{{ t('sets.catalogue.progress', { owned: owned(country), total: total(country) }) }}</span>
            <span class="mt-2 block h-1.5 overflow-hidden rounded-full bg-input" aria-hidden="true">
              <span class="block h-full rounded-full bg-gold" :style="{ width: `${percent(country)}%` }"></span>
            </span>
          </button>
        </li>
      </ul>

      <section v-if="selected" class="mt-6" aria-labelledby="catalogue-country-title" data-testid="catalogue-country">
        <div class="mb-3 flex flex-wrap items-baseline justify-between gap-2">
          <h2 id="catalogue-country-title" class="m-0 font-display text-lg font-medium text-heading">{{ releaseCountryName(selected) }}</h2>
          <span v-if="selectedSummary" class="text-body text-text-secondary" data-testid="catalogue-owned">
            {{ t('sets.catalogue.owned', { owned: owned(selectedSummary), total: total(selectedSummary) }) }}
          </span>
        </div>

        <div class="mb-4 flex flex-wrap items-center gap-2">
          <template v-if="hasBothKinds">
            <button
              v-for="option in KINDS"
              :key="option"
              type="button"
              class="chip border"
              :class="option === kind ? 'border-gold text-gold' : 'border-border-subtle'"
              :aria-pressed="option === kind"
              :data-testid="`catalogue-kind-${option}`"
              @click="kind = option; limit = PAGE"
            >
              {{ t(`sets.catalogue.kinds.${option}`) }}
            </button>
            <span class="mx-1 h-5 w-px bg-border-subtle" aria-hidden="true"></span>
          </template>
          <button
            v-for="option in OWNED"
            :key="option"
            type="button"
            class="chip border"
            :class="option === ownedFilter ? 'border-gold text-gold' : 'border-border-subtle'"
            :aria-pressed="option === ownedFilter"
            :data-testid="`catalogue-owned-${option}`"
            @click="ownedFilter = option; limit = PAGE"
          >
            {{ t(`sets.catalogue.ownedFilter.${option}`) }}
          </button>
        </div>

        <p v-if="!loadingCountry && !complete" class="mb-4 mt-0 rounded-md border border-border-subtle bg-card p-3 text-body text-text-secondary" role="status" data-testid="catalogue-partial">
          {{ t('sets.catalogue.partialNumista') }}
        </p>
        <p v-if="loadingCountry" class="text-body text-text-muted">{{ t('common.loading') }}</p>
        <p v-else-if="countryFailed" class="text-body text-text-secondary" role="alert">{{ t('sets.catalogue.failed') }}</p>
        <p v-else-if="!filtered.length" class="text-body text-text-secondary" data-testid="catalogue-empty">{{ t('sets.catalogue.empty') }}</p>
        <template v-else>
          <section v-for="group in byYear" :key="group.year" class="mb-5" :aria-labelledby="`catalogue-year-${group.year}`">
            <h3 :id="`catalogue-year-${group.year}`" class="mb-2 mt-0 text-base font-medium text-text-secondary" data-testid="catalogue-year">
              {{ group.year }}
              <span class="text-chip text-text-muted">· {{ t('sets.catalogue.progress', { owned: group.owned, total: group.coins.length }) }}</span>
            </h3>
            <ul class="m-0 grid list-none grid-cols-2 gap-3 p-0 md:grid-cols-4 xl:grid-cols-6">
              <CoinReleaseCard v-for="coin in group.coins" :key="coin.source + coin.imageUrl + coin.title" :release="coin" :show-country="false" />
            </ul>
          </section>
          <button
            v-if="filtered.length > limit"
            type="button"
            class="rounded-sm text-body text-gold underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
            data-testid="catalogue-more"
            @click="limit += PAGE"
          >
            {{ t('sets.catalogue.more', { count: filtered.length - limit }) }}
          </button>
        </template>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
// Every €2 commemorative coin the ECB has published, every euro country's
// €5, €10 and €20 collector coins (with the other coins the mints of
// Portugal, France and Austria sell), and Portugal's commemorative and
// collector coins in escudos, country by country, the collection's own
// marked "In my collection". The country chosen is kept in the address.
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getCoinCatalogue, getCoinCatalogueCountries, type CatalogueCountry, type CoinRelease } from '@/api/client'
import CoinReleaseCard from '@/components/collection/CoinReleaseCard.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import { intlLocale, t } from '@/i18n'
import { releaseCountryName } from '@/utils/coinReleases'

const KINDS = ['all', 'commemoratives', 'collector'] as const
const OWNED = ['all', 'owned', 'missing'] as const
/** Coins shown at first, and added by "Show more". */
const PAGE = 48

const route = useRoute()
const router = useRouter()

const countries = ref<CatalogueCountry[]>([])
const failed = ref(false)
const selected = ref('')
const coins = ref<CoinRelease[]>([])
/** False while some coins from Numista are still to be listed. */
const complete = ref(true)
const loadingCountry = ref(false)
const countryFailed = ref(false)
const kind = ref<(typeof KINDS)[number]>('all')
const ownedFilter = ref<(typeof OWNED)[number]>('all')
const limit = ref(PAGE)

function total(country: CatalogueCountry): number {
  return country.commemoratives + country.collector
}

function owned(country: CatalogueCountry): number {
  return country.ownedCommemoratives + country.ownedCollector
}

function percent(country: CatalogueCountry): number {
  return total(country) ? Math.round((owned(country) / total(country)) * 100) : 0
}

const sortedCountries = computed(() => [...countries.value].sort((a, b) =>
  releaseCountryName(a.country).localeCompare(releaseCountryName(b.country), intlLocale())))
const selectedSummary = computed(() => countries.value.find((country) => country.country === selected.value))
const hasBothKinds = computed(() => !!selectedSummary.value?.commemoratives && !!selectedSummary.value?.collector)

const filtered = computed(() => coins.value.filter((coin) =>
  (kind.value === 'all' || (kind.value === 'commemoratives') === (coin.kind === 'commemorative'))
  && (ownedFilter.value === 'all' || (ownedFilter.value === 'owned') === coin.owned)))

// The coins shown, by year, newest first, with how many of each year the collection has.
const byYear = computed(() => {
  const years: { year: string; coins: CoinRelease[]; owned: number }[] = []
  for (const coin of filtered.value.slice(0, limit.value)) {
    const year = coin.from.slice(0, 4)
    let group = years[years.length - 1]
    if (group?.year !== year) {
      group = { year, coins: [], owned: 0 }
      years.push(group)
    }
    group.coins.push(coin)
    if (coin.owned) group.owned++
  }
  return years
})

async function choose(country: string) {
  if (country === selected.value && coins.value.length) return
  selected.value = country
  kind.value = 'all'
  ownedFilter.value = 'all'
  limit.value = PAGE
  void router.replace({ query: { ...route.query, country } })
  loadingCountry.value = true
  countryFailed.value = false
  try {
    const { data } = await getCoinCatalogue(country)
    if (selected.value === country) {
      coins.value = data.coins
      complete.value = data.complete !== false
    }
  } catch {
    countryFailed.value = true
    coins.value = []
  } finally {
    loadingCountry.value = false
  }
}

onMounted(async () => {
  try {
    countries.value = (await getCoinCatalogueCountries()).data.countries
  } catch {
    failed.value = true
    return
  }
  // The country in the address, else the one the collection has most coins of.
  const wanted = String(route.query.country ?? '').toUpperCase()
  const start = countries.value.find((country) => country.country === wanted)
    ?? [...countries.value].sort((a, b) => owned(b) - owned(a))[0]
  if (start) await choose(start.country)
})
</script>
