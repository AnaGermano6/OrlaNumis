<template>
  <div class="container">
    <!-- On a phone, the title at the top; on a wider screen, in the list of countries. -->
    <div class="md:hidden">
      <header class="page-header">
        <h1>{{ t('sets.catalogue.title') }}</h1>
      </header>
    </div>

    <p v-if="failed" class="text-body text-text-secondary" role="alert">{{ t('sets.catalogue.failed') }}</p>
    <p v-else-if="!countries.length" class="text-body text-text-muted" data-testid="catalogue-loading">{{ t('sets.catalogue.loading') }}</p>
    <div v-else class="grid items-start gap-6 md:grid-cols-[16rem_minmax(0,1fr)] lg:grid-cols-[18rem_minmax(0,1fr)]">
      <!-- On a phone, a list to choose from. -->
      <select
        class="form-select w-full md:hidden"
        :value="selected"
        :aria-label="t('sets.catalogue.countries')"
        data-testid="catalogue-country-select"
        @change="choose(($event.target as HTMLSelectElement).value)"
      >
        <optgroup v-for="group in groups" :key="group.id" :label="group.title">
          <option v-for="entry in group.entries" :key="entryKey(entry)" :value="entryKey(entry)">
            {{ withFlagOfCode(entry.country, entryName(entry)) }} — {{ t('sets.catalogue.progress', { owned: owned(entry), total: total(entry) }) }}
          </option>
        </optgroup>
      </select>

      <!-- On a wider screen, the countries beside: searched for, each with how far it is; the euro's or the other currencies'. -->
      <aside class="sticky top-20 hidden max-h-[calc(100vh-6rem)] flex-col gap-4 rounded-lg border border-border-subtle bg-card p-4 md:flex" :aria-label="t('sets.catalogue.countries')" data-testid="catalogue-countries">
        <h1 class="m-0 text-center font-display text-xl leading-tight text-heading">{{ t('sets.catalogue.title') }}</h1>
        <SearchBar v-model="countrySearch" box class="!max-w-none" :placeholder="t('sets.byCountry.search')" data-testid="catalogue-search" />
        <ul ref="countryList" class="-mx-1 m-0 flex min-h-0 flex-1 list-none flex-col gap-2 overflow-y-auto px-1 py-0.5">
          <li v-for="entry in listed" :key="entryKey(entry)">
            <button
              type="button"
              class="flex w-full items-center gap-3 rounded-md border bg-input px-3 py-2 text-left transition-colors hover:border-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
              :class="entryKey(entry) === selected ? 'border-gold' : 'border-transparent'"
              :aria-pressed="entryKey(entry) === selected"
              :data-testid="`catalogue-country-${entryKey(entry)}`"
              @click="choose(entryKey(entry))"
            >
              <span class="shrink-0 font-sans" aria-hidden="true" data-testid="catalogue-flag">{{ flagOfCode(entry.country) }}</span>
              <span class="min-w-0 flex-1 truncate text-body font-medium text-text-primary">{{ releaseCountryName(entry.country) }}</span>
              <span class="shrink-0 text-chip text-text-secondary">{{ formatShare(share(entry)) }}</span>
              <span v-if="entryKey(entry) === selected" class="shrink-0 text-chip text-text-muted">{{ owned(entry) }}/{{ total(entry) }}</span>
              <span v-else class="block h-1.5 w-12 shrink-0 overflow-hidden rounded-full bg-card" aria-hidden="true">
                <span class="block h-full rounded-full bg-gold" :style="{ width: `${percent(entry)}%` }"></span>
              </span>
            </button>
          </li>
          <li v-if="!listed.length" class="px-1 text-body text-text-muted" data-testid="catalogue-search-none">{{ t('sets.byCountry.noMatch') }}</li>
        </ul>
        <!-- The euro coins, or the other currencies'. -->
        <div v-if="groups.length > 1" class="flex rounded-full border border-border-subtle bg-input p-1" role="group" :aria-label="t('sets.catalogue.countries')">
          <button
            v-for="group in groups"
            :key="group.id"
            type="button"
            class="min-w-0 flex-1 truncate rounded-full px-2 py-1.5 text-chip transition-colors"
            :class="group.id === listGroup ? 'bg-text-primary font-semibold text-surface' : 'text-text-secondary hover:text-text-primary'"
            :aria-pressed="group.id === listGroup"
            :data-testid="`catalogue-group-${group.id}`"
            @click="listGroup = group.id"
          >
            {{ group.title }}
          </button>
        </div>
      </aside>

      <section v-if="selected" class="min-w-0" aria-labelledby="catalogue-country-title" data-testid="catalogue-country">
        <!-- The country: its flag, name and how far the collection is; a set of its coins. -->
        <div class="flex flex-wrap items-center justify-between gap-4 border-b border-border-subtle pb-4">
          <div v-if="selectedSummary" class="min-w-0">
            <div class="min-w-0">
              <h2 id="catalogue-country-title" class="m-0 truncate font-display text-2xl text-heading">
                <span v-if="flagOfCode(selectedSummary.country)" class="mr-2 font-sans" aria-hidden="true" data-testid="catalogue-title-flag">{{ flagOfCode(selectedSummary.country) }}</span>{{ entryName(selectedSummary) }}
              </h2>
              <div class="mt-1 flex flex-wrap items-center gap-x-4 gap-y-1">
                <span class="text-body text-text-secondary" data-testid="catalogue-owned">
                  {{ t('sets.catalogue.ownedOf', { owned: owned(selectedSummary), total: total(selectedSummary), percent: formatShare(share(selectedSummary)) }) }}
                </span>
                <span class="block h-2 w-40 overflow-hidden rounded-full bg-input sm:w-64" aria-hidden="true">
                  <span class="block h-full rounded-full bg-gold" :style="{ width: `${percent(selectedSummary)}%` }"></span>
                </span>
              </div>
            </div>
          </div>
          <!-- A set of these coins (of the face value chosen), kept up to date by itself. -->
          <button
            v-if="selectedSummary && coins.length"
            type="button"
            class="btn btn-secondary"
            :disabled="creatingSet"
            :title="t('sets.catalogue.createSetHint')"
            data-testid="catalogue-create-set"
            @click="createChecklistSet"
          >
            <Plus :size="16" /> {{ t('sets.catalogue.createSet') }}
          </button>
        </div>
        <p v-if="createSetFailed" class="mb-0 mt-3 text-body text-error-bg" role="alert">{{ t('sets.catalogue.createSetFailed') }}</p>

        <!-- Its coins of a year, of a face value, the collection's or the missing; and without the special editions. -->
        <div class="mb-5 flex flex-wrap items-center gap-3 border-b border-border-subtle py-4">
          <select v-if="years.length > 1" v-model="year" class="form-select !w-auto !rounded-full !py-1.5 text-body" :aria-label="t('sets.catalogue.filters.year')" data-testid="catalogue-year-filter">
            <option value="">{{ t('sets.catalogue.filters.yearAll') }}</option>
            <option v-for="option in years" :key="option" :value="option">{{ t('sets.catalogue.filters.yearOf', { year: option }) }}</option>
          </select>
          <select v-if="faceValues.length > 1" v-model="faceValueChoice" class="form-select !w-auto !rounded-full !py-1.5 text-body" :aria-label="t('sets.catalogue.filters.value')" data-testid="catalogue-value-filter">
            <option value="">{{ t('sets.catalogue.filters.valueAll') }}</option>
            <option v-for="value in faceValues" :key="value" :value="String(value)">{{ t('sets.catalogue.filters.valueOf', { value: formatFaceValue(value, selectedSummary?.currency) }) }}</option>
          </select>
          <div class="flex items-center rounded-full border border-border-subtle bg-input p-1" role="group" :aria-label="t('sets.catalogue.filters.state')">
            <span class="px-2 text-body text-text-secondary">{{ t('sets.catalogue.filters.state') }}:</span>
            <button
              v-for="option in OWNED"
              :key="option"
              type="button"
              class="rounded-full px-3 py-1 text-body transition-colors"
              :class="option === ownedFilter ? 'bg-gold font-semibold text-surface' : 'text-text-secondary hover:text-text-primary'"
              :aria-pressed="option === ownedFilter"
              :data-testid="`catalogue-owned-${option}`"
              @click="ownedFilter = option; limit = PAGE"
            >
              {{ t(`sets.catalogue.ownedFilter.${option}`) }}
            </button>
          </div>
          <!-- As in the new releases: without the gold coins and those in proof or FDC. -->
          <button
            v-if="hasSpecialEditions"
            type="button"
            class="chip border"
            :class="noSpecialEditions ? '!border-gold !text-gold' : ''"
            :aria-pressed="noSpecialEditions"
            :title="t('collectionOverview.releases.noSpecialEditionsHint')"
            data-testid="catalogue-no-special"
            @click="noSpecialEditions = !noSpecialEditions; limit = PAGE"
          >
            {{ t('collectionOverview.releases.noSpecialEditions') }}
          </button>
        </div>

        <p v-if="!loadingCountry && !complete" class="mb-4 mt-0 rounded-md border border-border-subtle bg-card p-3 text-body text-text-secondary" role="status" data-testid="catalogue-partial">
          {{ t('sets.catalogue.partialNumista') }}
        </p>
        <p v-if="loadingCountry" class="text-body text-text-muted">{{ t('common.loading') }}</p>
        <p v-else-if="countryFailed" class="text-body text-text-secondary" role="alert">{{ t('sets.catalogue.failed') }}</p>
        <p v-else-if="!filtered.length" class="text-body text-text-secondary" data-testid="catalogue-empty">{{ t('sets.catalogue.empty') }}</p>
        <template v-else>
          <!-- The coins, newest first, two in a row on a wide screen. -->
          <ul class="m-0 grid list-none gap-4 p-0 xl:grid-cols-2" data-testid="catalogue-coins">
            <CoinReleaseCard
              v-for="coin in filtered.slice(0, limit)"
              :key="coin.source + coin.imageUrl + coin.title"
              :release="coin"
              :show-country="false"
              layout="row"
              addable
              wishlistable
              :manageable="Boolean(coin.id)"
              @manage="openOptions(coin)"
            />
          </ul>
          <button
            v-if="filtered.length > limit"
            type="button"
            class="mt-4 rounded-sm text-body text-gold underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
            data-testid="catalogue-more"
            @click="limit += PAGE"
          >
            {{ t('sets.catalogue.more', { count: filtered.length - limit }) }}
          </button>
        </template>

        <!-- The coins hidden, or joined to another: each can be listed again. -->
        <details v-if="!loadingCountry && hidden.length" class="mt-6 rounded-md border border-border-subtle bg-card px-4 py-3" data-testid="catalogue-hidden">
          <summary class="cursor-pointer text-body font-medium text-text-secondary">{{ t('sets.catalogue.hidden.title', { count: hidden.length }) }}</summary>
          <p class="mb-1 mt-2 text-chip text-text-muted">{{ t('sets.catalogue.hidden.body') }}</p>
          <ul class="m-0 list-none p-0">
            <li v-for="coin in hidden" :key="coin.id" class="flex items-center gap-3 border-t border-border-subtle py-2 first:border-t-0" data-testid="catalogue-hidden-coin">
              <span class="min-w-0 flex-1">
                <span class="block truncate text-body text-text-primary">{{ coin.title }}</span>
                <span class="block text-chip text-text-muted">
                  {{ releaseFaceValues(coin) }} · {{ coin.from.slice(0, 4) }}<template v-if="coin.sameAs"> · {{ t('sets.catalogue.hidden.sameAs', { title: titleOf(coin.sameAs) }) }}</template>
                </span>
              </span>
              <button type="button" class="btn btn-secondary btn-sm shrink-0" :disabled="saving" data-testid="catalogue-show" @click="showAgain(coin)">
                <Eye :size="14" /> {{ t('sets.catalogue.hidden.show') }}
              </button>
            </li>
          </ul>
          <p v-if="showFailed" class="mb-0 mt-2 text-body text-loss" role="alert">{{ t('sets.catalogue.manage.failed') }}</p>
        </details>

        <p class="mb-0 mt-6 text-chip text-text-muted">{{ t('sets.catalogue.sources') }}</p>
      </section>
    </div>

    <CatalogueCoinOptions :coin="options" :coins="coins" :saving="saving" :failed="optionsFailed" @hide="hide" @join="join" @close="options = null" />
  </div>
</template>

<script setup lang="ts">
// Euro coins, country by country: every €2 commemorative the ECB has
// published, the €5, €10 and €20 coins and the other coins the mints of
// Portugal, France and Austria sell. Other coins apart: Portugal's
// commemorative and collector coins in escudos. The collection's own are
// marked "In my collection"; filters narrow the list to a year, a face
// value, the collection's or the missing coins, and leave out the special
// editions (gold, proof, FDC). The countries are listed beside, to search,
// the euro's or the other currencies'.
// A coin can be hidden, or said to be the same as another (one coin listed
// twice); those are listed apart, to show again. The country (and currency)
// chosen is kept in the address.
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Eye, Plus } from 'lucide-vue-next'
import {
  createSet, getCoinCatalogue, getCoinCatalogueCountries, hideCatalogueCoin, joinCatalogueCoin, showCatalogueCoin,
  type CatalogueCountry, type CoinRelease,
} from '@/api/client'
import CatalogueCoinOptions from '@/components/collection/CatalogueCoinOptions.vue'
import CoinReleaseCard from '@/components/collection/CoinReleaseCard.vue'
import SearchBar from '@/components/SearchBar.vue'
import { intlLocale, t } from '@/i18n'
import { formatFaceValue, isSpecialEdition, releaseCountryName, releaseFaceValues } from '@/utils/coinReleases'
import { flagOfCode, withFlagOfCode } from '@/utils/countryFlags'
import { formatShare } from '@/utils/numberFormat'
import { randomSetColor } from '@/utils/setColors'

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
// The value filter's choice, as its menu holds it ('' for every one).
const faceValueChoice = computed({
  get: () => (faceValue.value === null ? '' : String(faceValue.value)),
  set: (value: string) => {
    faceValue.value = value === '' ? null : Number(value)
    limit.value = PAGE
  },
})
/** The year shown ("2026"); '' for every one. */
const year = ref('')
/** The countries listed beside: the euro's, or the other currencies'. */
const listGroup = ref<'euro' | 'other'>('euro')
/** Part of a country's name, to find it in the list. */
const countrySearch = ref('')
const countryList = ref<HTMLElement | null>(null)

// The country chosen in sight in its list (the list scrolls, not the page).
async function showChosen() {
  await nextTick()
  const list = countryList.value
  const chosen = list?.querySelector<HTMLElement>('[aria-pressed="true"]')
  if (!list || !chosen) return
  const top = chosen.offsetTop - list.offsetTop
  if (top < list.scrollTop || top + chosen.offsetHeight > list.scrollTop + list.clientHeight) {
    list.scrollTop = Math.max(0, top - list.clientHeight / 2 + chosen.offsetHeight / 2)
  }
}
const ownedFilter = ref<(typeof OWNED)[number]>('all')
/** Leaves out the special editions, on every country. */
const noSpecialEditions = ref(false)
const limit = ref(PAGE)
const creatingSet = ref(false)
const createSetFailed = ref(false)
/** The coins hidden, or joined to another (sameAs). */
const hidden = ref<CoinRelease[]>([])
/** The coin whose options are open. */
const options = ref<CoinRelease | null>(null)
const saving = ref(false)
const optionsFailed = ref(false)
const showFailed = ref(false)

function total(country: CatalogueCountry): number {
  return country.commemoratives + country.collector
}

function owned(country: CatalogueCountry): number {
  return country.ownedCommemoratives + country.ownedCollector
}

function share(country: CatalogueCountry): number {
  return total(country) ? owned(country) / total(country) : 0
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

// The other currencies' coins, named by their currency when they have one ("Escudos").
function otherTitle(entries: CatalogueCountry[]): string {
  const currencies = [...new Set(entries.map((entry) => entry.currency ?? ''))]
  return currencies.length === 1 && currencies[0] ? currencyName(currencies[0]) : t('sets.catalogue.otherCoins')
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
    { id: 'euro' as const, title: t('sets.catalogue.euroCoins'), entries: euro },
    { id: 'other' as const, title: otherTitle(other), entries: other },
  ].filter((group) => group.entries.length)
})
const selectedSummary = computed(() => countries.value.find((entry) => entryKey(entry) === selected.value))

function fold(value: string): string {
  return value.toLocaleLowerCase(intlLocale()).normalize('NFD').replace(/[̀-ͯ]/g, '')
}

// The countries of the group chosen, those whose name has what is searched for.
const listed = computed(() => {
  const query = fold(countrySearch.value.trim())
  return (groups.value.find((group) => group.id === listGroup.value) ?? groups.value[0])?.entries
    .filter((entry) => !query || fold(releaseCountryName(entry.country)).includes(query)) ?? []
})

// Another year shows its coins from the first.
watch(year, () => {
  limit.value = PAGE
})

// The years of the country's coins, newest first.
const years = computed(() => [...new Set(coins.value.map((coin) => coin.from.slice(0, 4)))].sort((a, b) => b.localeCompare(a)))
// The face values the country's coins have, from the lowest.
const faceValues = computed(() => [...new Set(coins.value.flatMap(coinValues))].sort((a, b) => a - b))

const hasSpecialEditions = computed(() => coins.value.some(isSpecialEdition))

const filtered = computed(() => coins.value.filter((coin) =>
  (!year.value || coin.from.startsWith(year.value))
  && (faceValue.value === null || coinValues(coin).includes(faceValue.value))
  && (ownedFilter.value === 'all' || (ownedFilter.value === 'owned') === coin.owned)
  && (!noSpecialEditions.value || !isSpecialEdition(coin))))

// A checklist set of the country's coins (of the face value chosen): its
// page shows them all, the collection's and the missing, as an album.
async function createChecklistSet() {
  const entry = selectedSummary.value
  if (!entry || creatingSet.value) return
  creatingSet.value = true
  createSetFailed.value = false
  const value = faceValue.value
  try {
    const { data } = await createSet({
      name: value === null ? entryName(entry) : `${entryName(entry)} · ${formatFaceValue(value, entry.currency)}`,
      color: randomSetColor(),
      setType: 'checklist',
      checklist: { source: 'catalogue', country: entry.country, ...(entry.currency ? { currency: entry.currency } : {}), ...(value === null ? {} : { faceValue: value }) },
    })
    void router.push({ name: 'set-detail', params: { id: data.id } })
  } catch {
    createSetFailed.value = true
  } finally {
    creatingSet.value = false
  }
}

async function choose(key: string) {
  if (key === selected.value && coins.value.length) return
  const entry = countries.value.find((candidate) => entryKey(candidate) === key)
  if (!entry) return
  selected.value = key
  listGroup.value = entry.currency ? 'other' : 'euro'
  createSetFailed.value = false
  showFailed.value = false
  faceValue.value = null
  year.value = ''
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
      hidden.value = data.hidden ?? []
      complete.value = data.complete !== false
    }
  } catch {
    countryFailed.value = true
    coins.value = []
    hidden.value = []
  } finally {
    loadingCountry.value = false
  }
}

// The title of a coin of the country, shown or not.
function titleOf(id: string): string {
  return [...coins.value, ...hidden.value].find((coin) => coin.id === id)?.title ?? ''
}

function openOptions(coin: CoinRelease) {
  optionsFailed.value = false
  options.value = coin
}

// After a coin is hidden, joined or shown again: the country's coins and
// every country's counts as they now are.
async function reload() {
  const entry = selectedSummary.value
  if (!entry) return
  const key = selected.value
  const [list, catalogue] = await Promise.all([getCoinCatalogueCountries(), getCoinCatalogue(entry.country, entry.currency)])
  countries.value = list.data.countries
  if (selected.value === key) {
    coins.value = catalogue.data.coins
    hidden.value = catalogue.data.hidden ?? []
    complete.value = catalogue.data.complete !== false
  }
}

async function change(request: () => Promise<unknown>): Promise<boolean> {
  if (saving.value) return false
  saving.value = true
  try {
    await request()
    await reload()
    return true
  } catch {
    return false
  } finally {
    saving.value = false
  }
}

async function hide() {
  const id = options.value?.id
  if (!id) return
  optionsFailed.value = false
  if (await change(() => hideCatalogueCoin(id))) options.value = null
  else optionsFailed.value = true
}

async function join(sameAs: string) {
  const id = options.value?.id
  if (!id) return
  optionsFailed.value = false
  if (await change(() => joinCatalogueCoin(id, sameAs))) options.value = null
  else optionsFailed.value = true
}

async function showAgain(coin: CoinRelease) {
  if (!coin.id) return
  showFailed.value = false
  if (!(await change(() => showCatalogueCoin(coin.id!)))) showFailed.value = true
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
  if (start) {
    const opening = choose(entryKey(start))
    void showChosen()
    await opening
  }
})
</script>
