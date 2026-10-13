<template>
  <div class="container">
    <DesktopPageContext :title="t('sets.catalogue.title')">
      <header class="page-header">
        <h1>{{ t('sets.catalogue.title') }}</h1>
      </header>
    </DesktopPageContext>
    <p class="mb-4 mt-0 text-chip text-text-muted">{{ t('sets.catalogue.sources') }}</p>

    <p v-if="failed" class="text-body text-text-secondary" role="alert">{{ t('sets.catalogue.failed') }}</p>
    <p v-else-if="!countries.length" class="text-body text-text-muted" data-testid="catalogue-loading">{{ t('sets.catalogue.loading') }}</p>
    <template v-else>
      <!-- On a phone, a list to choose from; the grids of countries above that. -->
      <select
        class="form-select w-full sm:hidden"
        :value="selected"
        :aria-label="t('sets.catalogue.countries')"
        data-testid="catalogue-country-select"
        @change="choose(($event.target as HTMLSelectElement).value)"
      >
        <optgroup v-for="group in groups" :key="group.id" :label="group.title">
          <option v-for="entry in group.entries" :key="entryKey(entry)" :value="entryKey(entry)">
            {{ entryName(entry) }} — {{ t('sets.catalogue.progress', { owned: owned(entry), total: total(entry) }) }}
          </option>
        </optgroup>
      </select>
      <section
        v-for="group in groups"
        :key="group.id"
        class="mb-4 hidden sm:block"
        :aria-labelledby="`catalogue-group-${group.id}`"
        :data-testid="`catalogue-group-${group.id}`"
      >
        <h2 :id="`catalogue-group-${group.id}`" class="mb-2 mt-0 text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ group.title }}</h2>
        <ul class="m-0 grid list-none grid-cols-2 gap-2 p-0 sm:grid-cols-3 lg:grid-cols-5">
          <li v-for="entry in group.entries" :key="entryKey(entry)">
            <button
              type="button"
              class="w-full rounded-md border bg-card p-3 text-left transition-colors hover:border-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
              :class="entryKey(entry) === selected ? 'border-gold' : 'border-border-subtle'"
              :aria-pressed="entryKey(entry) === selected"
              :data-testid="`catalogue-country-${entryKey(entry)}`"
              @click="choose(entryKey(entry))"
            >
              <span class="block truncate font-medium text-text-primary">{{ releaseCountryName(entry.country) }}</span>
              <span class="block text-chip text-text-muted">
                <template v-if="entry.currency">{{ currencyName(entry.currency) }} · </template>{{ t('sets.catalogue.progress', { owned: owned(entry), total: total(entry) }) }}
              </span>
              <span class="mt-2 block h-1.5 overflow-hidden rounded-full bg-input" aria-hidden="true">
                <span class="block h-full rounded-full bg-gold" :style="{ width: `${percent(entry)}%` }"></span>
              </span>
            </button>
          </li>
        </ul>
      </section>

      <section v-if="selected" class="mt-6" aria-labelledby="catalogue-country-title" data-testid="catalogue-country">
        <div class="mb-3 flex flex-wrap items-baseline justify-between gap-2">
          <h2 id="catalogue-country-title" class="m-0 font-display text-lg font-medium text-heading">{{ selectedSummary ? entryName(selectedSummary) : '' }}</h2>
          <span v-if="selectedSummary" class="text-body text-text-secondary" data-testid="catalogue-owned">
            {{ t('sets.catalogue.owned', { owned: owned(selectedSummary), total: total(selectedSummary) }) }}
          </span>
        </div>

        <div class="mb-4 flex flex-wrap items-center gap-2">
          <!-- A filter for each face value the country's coins have. -->
          <template v-if="faceValues.length > 1">
            <button
              type="button"
              class="chip border"
              :class="faceValue === null ? 'border-gold text-gold' : 'border-border-subtle'"
              :aria-pressed="faceValue === null"
              data-testid="catalogue-value-all"
              @click="faceValue = null; limit = PAGE"
            >
              {{ t('sets.catalogue.allValues') }}
            </button>
            <button
              v-for="value in faceValues"
              :key="value"
              type="button"
              class="chip border"
              :class="value === faceValue ? 'border-gold text-gold' : 'border-border-subtle'"
              :aria-pressed="value === faceValue"
              :data-testid="`catalogue-value-${value}`"
              @click="faceValue = value; limit = PAGE"
            >
              {{ formatFaceValue(value, selectedSummary?.currency) }}
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
// Euro coins, country by country: every €2 commemorative the ECB has
// published, the €5, €10 and €20 coins and the other coins the mints of
// Portugal, France and Austria sell. Other coins apart: Portugal's
// commemorative and collector coins in escudos. The collection's own are
// marked "In my collection"; a filter for each face value narrows the list.
// The country (and currency) chosen is kept in the address.
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getCoinCatalogue, getCoinCatalogueCountries, type CatalogueCountry, type CoinRelease } from '@/api/client'
import CoinReleaseCard from '@/components/collection/CoinReleaseCard.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import { intlLocale, t } from '@/i18n'
import { formatFaceValue, releaseCountryName } from '@/utils/coinReleases'

const OWNED = ['all', 'owned', 'missing'] as const
/** Coins shown at first, and added by "Show more". */
const PAGE = 48

const route = useRoute()
const router = useRouter()

const countries = ref<CatalogueCountry[]>([])
const failed = ref(false)
/** The key of the country chosen: its ISO code, with "-PTE" for its escudos. */
const selected = ref('')
const coins = ref<CoinRelease[]>([])
/** False while some coins from Numista are still to be listed. */
const complete = ref(true)
const loadingCountry = ref(false)
const countryFailed = ref(false)
/** The face value shown; null for every one. */
const faceValue = ref<number | null>(null)
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

function entryKey(entry: CatalogueCountry): string {
  return entry.currency ? `${entry.country}-${entry.currency}` : entry.country
}

// The currency of a country's coins other than the euro: "Escudos".
function currencyName(currency: string): string {
  return currency === 'PTE' ? t('sets.catalogue.currencies.PTE') : currency
}

// "Portugal", or "Portugal · Escudos" for its coins in another currency.
function entryName(entry: CatalogueCountry): string {
  const name = releaseCountryName(entry.country)
  return entry.currency ? `${name} · ${currencyName(entry.currency)}` : name
}

function coinValues(coin: CoinRelease): number[] {
  return coin.faceValues ?? [coin.faceValue]
}

// The euro coins first, then the other currencies, each by country name.
const groups = computed(() => {
  const byName = (a: CatalogueCountry, b: CatalogueCountry) => releaseCountryName(a.country).localeCompare(releaseCountryName(b.country), intlLocale())
  const euro = countries.value.filter((entry) => !entry.currency).sort(byName)
  const other = countries.value.filter((entry) => entry.currency).sort(byName)
  return [
    { id: 'euro', title: t('sets.catalogue.euroCoins'), entries: euro },
    { id: 'other', title: t('sets.catalogue.otherCoins'), entries: other },
  ].filter((group) => group.entries.length)
})
const selectedSummary = computed(() => countries.value.find((entry) => entryKey(entry) === selected.value))
// The face values the country's coins have, from the lowest.
const faceValues = computed(() => [...new Set(coins.value.flatMap(coinValues))].sort((a, b) => a - b))

const filtered = computed(() => coins.value.filter((coin) =>
  (faceValue.value === null || coinValues(coin).includes(faceValue.value))
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

async function choose(key: string) {
  if (key === selected.value && coins.value.length) return
  const entry = countries.value.find((candidate) => entryKey(candidate) === key)
  if (!entry) return
  selected.value = key
  faceValue.value = null
  ownedFilter.value = 'all'
  limit.value = PAGE
  const query: Record<string, string> = { country: entry.country }
  if (entry.currency) query.currency = entry.currency
  void router.replace({ query })
  loadingCountry.value = true
  countryFailed.value = false
  try {
    const { data } = await getCoinCatalogue(entry.country, entry.currency)
    if (selected.value === key) {
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
  const currency = String(route.query.currency ?? '').toUpperCase()
  const start = countries.value.find((entry) => entry.country === wanted && (entry.currency ?? '') === currency)
    ?? [...countries.value].sort((a, b) => owned(b) - owned(a))[0]
  if (start) await choose(entryKey(start))
})
</script>
