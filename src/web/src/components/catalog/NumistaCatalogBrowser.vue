<template>
  <div class="grid min-w-0 gap-4">
    <NumistaCatalogTypePreview
      v-if="selectedTypeId"
      :type-id="selectedTypeId"
      :year-hint="searchedYear"
      @back="selectedTypeId = null"
      @use="$emit('use', $event)"
    />

    <template v-else>
      <form class="card grid min-w-0 gap-3 p-4" role="search" @submit.prevent="runSearch(1)">
        <p class="m-0 text-body text-text-secondary">{{ t('catalog.intro') }}</p>
        <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)_7rem]">
          <div class="relative flex min-w-0 flex-col gap-1">
            <label class="form-label" for="catalog-country">{{ t('catalog.country') }}</label>
            <div v-if="country" class="flex min-h-[2.5rem] items-center gap-2">
              <span class="chip min-w-0 truncate border border-gold">{{ country.name }}</span>
              <button type="button" class="btn btn-ghost btn-sm" :aria-label="t('catalog.clearCountry')" @click="clearCountry">✕</button>
            </div>
            <template v-else>
              <input
                id="catalog-country"
                v-model="countryQuery"
                class="form-input"
                type="text"
                autocomplete="off"
                role="combobox"
                aria-autocomplete="list"
                aria-controls="catalog-country-options"
                :aria-expanded="suggestionsOpen"
                :aria-activedescendant="suggestionsOpen && activeSuggestion >= 0 ? `catalog-country-${activeSuggestion}` : undefined"
                :placeholder="t('catalog.countryPlaceholder')"
                @input="onCountryInput"
                @keydown.down.prevent="moveSuggestion(1)"
                @keydown.up.prevent="moveSuggestion(-1)"
                @keydown.enter="onCountryEnter"
                @keydown.escape="suggestionsOpen = false"
                @blur="suggestionsOpen = false"
              />
              <ul
                v-if="suggestionsOpen"
                id="catalog-country-options"
                role="listbox"
                class="absolute left-0 right-0 top-full z-20 m-0 mt-1 max-h-64 list-none overflow-y-auto rounded-sm border border-border-subtle bg-card p-1 shadow-lg"
              >
                <li
                  v-for="(issuer, index) in suggestions"
                  :id="`catalog-country-${index}`"
                  :key="issuer.code"
                  role="option"
                  :aria-selected="index === activeSuggestion"
                  class="cursor-pointer rounded-sm px-2 py-1.5 text-body"
                  :class="index === activeSuggestion ? 'bg-gold-glow text-text-primary' : 'text-text-primary hover:bg-card-hover'"
                  @mousedown.prevent="chooseCountry(issuer)"
                >
                  {{ issuer.name }}<span v-if="issuer.parent" class="text-text-muted"> · {{ issuer.parent }}</span>
                </li>
              </ul>
              <p v-if="countryMessage" class="m-0 text-chip text-text-muted">{{ countryMessage }}</p>
            </template>
          </div>
          <div class="flex min-w-0 flex-col gap-1">
            <label class="form-label" for="catalog-query">{{ t('catalog.query') }}</label>
            <input id="catalog-query" v-model="query" class="form-input" type="search" maxlength="200" :placeholder="t('catalog.queryPlaceholder')" />
          </div>
          <div class="flex min-w-0 flex-col gap-1">
            <label class="form-label" for="catalog-year">{{ t('catalog.year') }}</label>
            <input id="catalog-year" v-model="year" class="form-input" type="number" inputmode="numeric" min="-5000" max="3000" step="1" />
          </div>
        </div>
        <div class="flex flex-wrap items-center justify-between gap-2">
          <p class="m-0 text-chip text-text-muted">{{ t('catalog.yearHint') }}</p>
          <button type="submit" class="btn btn-primary" :disabled="!canSearch || searching">
            {{ searching && page === 1 ? t('common.searching') : t('common.search') }}
          </button>
        </div>
      </form>

      <section v-if="searched" class="grid min-w-0 gap-3" aria-live="polite">
        <p v-if="searchFailed" class="m-0 text-body text-text-secondary" role="status">{{ t('catalog.searchFailed') }}</p>
        <p v-else-if="status !== 'ok'" class="m-0 text-body text-text-secondary" role="status">{{ numistaStatusMessage(status) }}</p>
        <p v-else-if="!results.length && !searching" class="m-0 text-body text-text-secondary" role="status">{{ t('catalog.noResults') }}</p>
        <p v-else-if="results.length" class="m-0 text-chip text-text-muted">{{ t('catalog.resultCount', { count: total }) }}</p>

        <ul v-if="results.length" class="m-0 grid list-none gap-2 p-0 md:grid-cols-2">
          <li v-for="item in results" :key="item.id" class="min-w-0">
            <button
              type="button"
              class="flex w-full min-w-0 items-center gap-3 rounded-sm border border-border-subtle bg-card p-2 text-left transition-colors hover:border-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
              @click="selectedTypeId = item.id"
            >
              <CatalogPhoto :url="item.obverseThumbnail" :alt="t('catalog.obversePhoto')" size="3.5rem" />
              <CatalogPhoto :url="item.reverseThumbnail" :alt="t('catalog.reversePhoto')" size="3.5rem" />
              <span class="min-w-0">
                <strong class="block font-medium text-text-primary [overflow-wrap:anywhere]">{{ item.title }}</strong>
                <span class="block text-chip text-text-secondary">{{ [item.issuer.name, numistaYearRange(item.minYear, item.maxYear)].filter(Boolean).join(' · ') }}</span>
              </span>
            </button>
          </li>
        </ul>

        <button v-if="results.length < total && status === 'ok'" type="button" class="btn btn-secondary w-fit" :disabled="searching" @click="runSearch(page + 1)">
          {{ searching ? t('common.loading') : t('catalog.loadMore') }}
        </button>
        <p class="m-0 text-chip text-text-muted">{{ t('catalog.attribution') }}</p>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
// Search the Numista catalogue (optionally within one country) and pick the
// coin type to add. Searches are cached server-side for a day.
import { computed, ref } from 'vue'
import {
  searchNumistaCatalog,
  searchNumistaIssuers,
  type NumistaCatalogIssuer,
  type NumistaCatalogSearchItem,
  type NumistaDataStatus,
} from '@/api/client'
import { t } from '@/i18n'
import CatalogPhoto from '@/components/catalog/CatalogPhoto.vue'
import NumistaCatalogTypePreview from '@/components/catalog/NumistaCatalogTypePreview.vue'
import { numistaLang, numistaYearRange } from '@/utils/numistaCatalog'
import type { CatalogSelection } from '@/utils/numistaCatalogPrefill'
import { numistaStatusMessage } from '@/utils/numistaMarketValue'

defineEmits<{
  use: [selection: CatalogSelection]
}>()

const country = ref<NumistaCatalogIssuer | null>(null)
const countryQuery = ref('')
const suggestions = ref<NumistaCatalogIssuer[]>([])
const suggestionsOpen = ref(false)
const activeSuggestion = ref(-1)
const countryMessage = ref('')
let countryTimer: ReturnType<typeof setTimeout> | undefined
let countrySeq = 0

const query = ref('')
const year = ref<number | string>('')
const results = ref<NumistaCatalogSearchItem[]>([])
const total = ref(0)
const page = ref(1)
const status = ref<NumistaDataStatus>('ok')
const searching = ref(false)
const searched = ref(false)
const searchFailed = ref(false)
const searchedYear = ref<number | undefined>(undefined)
const selectedTypeId = ref<number | null>(null)
let searchSeq = 0

const yearValue = computed(() => {
  const value = Number(year.value)
  return year.value !== '' && Number.isInteger(value) && value !== 0 ? value : undefined
})
const canSearch = computed(() => query.value.trim() !== '' || country.value !== null)

function onCountryInput() {
  clearTimeout(countryTimer)
  countryMessage.value = ''
  const text = countryQuery.value.trim()
  if (text.length < 2) {
    suggestions.value = []
    suggestionsOpen.value = false
    return
  }
  countryTimer = setTimeout(() => void loadCountries(text), 250)
}

async function loadCountries(text: string) {
  const seq = ++countrySeq
  try {
    const res = await searchNumistaIssuers(text, numistaLang())
    if (seq !== countrySeq) return
    if (res.data.status !== 'ok') {
      countryMessage.value = numistaStatusMessage(res.data.status)
      suggestions.value = []
    } else {
      suggestions.value = res.data.issuers
      countryMessage.value = res.data.issuers.length ? '' : t('catalog.noCountries')
    }
  } catch {
    if (seq === countrySeq) suggestions.value = []
  }
  activeSuggestion.value = suggestions.value.length ? 0 : -1
  suggestionsOpen.value = suggestions.value.length > 0
}

function moveSuggestion(step: number) {
  if (!suggestions.value.length) return
  suggestionsOpen.value = true
  activeSuggestion.value = (activeSuggestion.value + step + suggestions.value.length) % suggestions.value.length
}

function onCountryEnter(event: KeyboardEvent) {
  const issuer = suggestions.value[activeSuggestion.value]
  if (suggestionsOpen.value && issuer) {
    event.preventDefault()
    chooseCountry(issuer)
  }
}

function chooseCountry(issuer: NumistaCatalogIssuer) {
  country.value = issuer
  countryQuery.value = ''
  suggestions.value = []
  suggestionsOpen.value = false
}

function clearCountry() {
  country.value = null
}

async function runSearch(nextPage: number) {
  if (!canSearch.value) return
  const seq = ++searchSeq
  searching.value = true
  searched.value = true
  searchFailed.value = false
  if (nextPage === 1) {
    results.value = []
    total.value = 0
    searchedYear.value = yearValue.value
  }
  try {
    const res = await searchNumistaCatalog({
      q: query.value.trim() || undefined,
      issuer: country.value?.code,
      year: yearValue.value,
      page: nextPage,
      lang: numistaLang(),
    })
    if (seq !== searchSeq) return
    status.value = res.data.status
    total.value = res.data.count
    page.value = nextPage
    const known = new Set(results.value.map((item) => item.id))
    results.value = [...results.value, ...res.data.types.filter((item) => !known.has(item.id))]
  } catch {
    if (seq === searchSeq) searchFailed.value = true
  } finally {
    if (seq === searchSeq) searching.value = false
  }
}
</script>
