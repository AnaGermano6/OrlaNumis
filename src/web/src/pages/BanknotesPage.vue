<template>
  <div class="container">
    <DesktopPageContext :title="t('banknotes.title')">
      <template #actions>
        <AppIconButton to="/banknotes/add" :title="t('banknotes.add')">
          <CirclePlus :size="20" />
        </AppIconButton>
      </template>
      <header class="page-header">
        <h1>{{ t('banknotes.title') }}</h1>
        <router-link to="/banknotes/add" class="btn btn-primary btn-sm ml-auto" data-testid="banknotes-add">{{ t('banknotes.add') }}</router-link>
      </header>
    </DesktopPageContext>


    <div class="mb-4 flex flex-col gap-3 rounded-md border border-border-subtle bg-card p-4 shadow-card">
      <SearchBar v-model="search" class="!max-w-none" :placeholder="t('banknotes.searchPlaceholder')" />
      <!-- Every filter the same size, as in the coin gallery. -->
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-3" data-testid="banknotes-filters">
        <label class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('coinFields.country') }}</span>
          <select v-model="country" class="form-select h-10 w-full bg-card py-0 text-body" data-testid="banknotes-country">
            <option value="">{{ t('collection.view.allCountries') }}</option>
            <option v-for="option in countries" :key="option.country" :value="option.country">{{ option.country }} ({{ option.count }})</option>
          </select>
        </label>
        <label class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('collection.menu.sort') }}</span>
          <select v-model="sortKey" class="form-select h-10 w-full bg-card py-0 text-body" data-testid="banknotes-sort">
            <option value="updated_at_desc">{{ t('collection.sort.lastUpdated') }}</option>
            <option value="created_at_desc">{{ t('collection.sort.newest') }}</option>
            <option value="created_at_asc">{{ t('collection.sort.oldest') }}</option>
            <option value="current_value_desc">{{ t('collection.sort.priceHighLow') }}</option>
            <option value="current_value_asc">{{ t('collection.sort.priceLowHigh') }}</option>
          </select>
        </label>
        <div class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('collection.view.layout') }}</span>
          <div class="flex h-10 w-full rounded-sm border border-border-subtle bg-input p-[2px]" role="group" :aria-label="t('collection.view.layout')">
            <button
              v-for="option in LAYOUTS"
              :key="option"
              type="button"
              class="min-w-0 flex-1 rounded-[6px] px-2 text-chip transition-all"
              :class="layout === option ? 'bg-gold text-surface font-semibold' : 'bg-transparent font-medium text-text-secondary hover:bg-card-hover hover:text-text-primary'"
              :aria-pressed="layout === option"
              @click="layout = option"
            >
              {{ option === 'grid' ? t('collection.view.grid') : t('collection.view.list') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div>
      <p>{{ t('collection.loading') }}</p>
    </div>
    <p v-else-if="failed" class="text-body text-text-secondary" role="alert">{{ t('banknotes.failed') }}</p>
    <template v-else-if="banknotes.length">
      <CoinListTable v-if="layout === 'list'" :coins="banknotes" />
      <div v-else class="coins-grid" data-testid="banknotes-grid">
        <CoinCard v-for="banknote in banknotes" :key="banknote.id" :coin="banknote" />
      </div>
      <CollectionPagination :page="page" :total="total" :per-page="PER_PAGE" view-mode="grid" @prev="page--" @next="page++" />
    </template>
    <div v-else class="empty-state" data-testid="banknotes-empty">
      <h3>{{ filtered ? t('banknotes.empty.noMatchTitle') : t('banknotes.empty.title') }}</h3>
      <p>{{ filtered ? t('banknotes.empty.noMatchBody') : t('banknotes.empty.body') }}</p>
      <router-link v-if="!filtered" to="/banknotes/add" class="btn btn-primary mt-4">{{ t('banknotes.add') }}</router-link>
    </div>

    <div class="h-16" aria-hidden="true"></div>
  </div>
</template>

<script setup lang="ts">
// The banknotes of the collection, apart from its coins: a search, a country
// filter, an order and a grid or list (their numbers are on the Collection
// page). Their own filters, so the coin gallery's never leak in.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { getCoinCountries, getCoins, type CoinCountryCount } from '@/api/client'
import type { Coin } from '@/types'
import { CirclePlus } from 'lucide-vue-next'
import CoinCard from '@/components/CoinCard.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import CoinListTable from '@/components/collection/CoinListTable.vue'
import CollectionPagination from '@/components/CollectionPagination.vue'
import SearchBar from '@/components/SearchBar.vue'
import { t } from '@/i18n'

const PER_PAGE = 50
const LAYOUTS = ['grid', 'list'] as const
const LAYOUT_KEY = 'banknotes:layout'

const banknotes = ref<Coin[]>([])
const total = ref(0)
const loading = ref(true)
const failed = ref(false)
const countries = ref<CoinCountryCount[]>([])

const search = ref('')
const country = ref('')
const sortKey = ref('updated_at_desc')
const page = ref(1)
const layout = ref<(typeof LAYOUTS)[number]>(readLayout())

const filtered = computed(() => Boolean(search.value.trim() || country.value))

function readLayout(): (typeof LAYOUTS)[number] {
  try {
    return localStorage.getItem(LAYOUT_KEY) === 'list' ? 'list' : 'grid'
  } catch {
    return 'grid'
  }
}

watch(layout, (value) => {
  try {
    localStorage.setItem(LAYOUT_KEY, value)
  } catch {
    // The choice still applies for this visit.
  }
})

// Only the latest request's answer is shown.
let request = 0
async function load() {
  const current = ++request
  loading.value = true
  failed.value = false
  const [sort, order] = [sortKey.value.slice(0, sortKey.value.lastIndexOf('_')), sortKey.value.slice(sortKey.value.lastIndexOf('_') + 1)]
  try {
    const res = await getCoins({
      kind: 'banknote',
      wishlist: 'false',
      search: search.value.trim() || undefined,
      country: country.value || undefined,
      sort,
      order,
      page: page.value,
      limit: PER_PAGE,
    })
    if (current !== request) return
    banknotes.value = res.data.coins
    total.value = res.data.total
  } catch {
    if (current !== request) return
    banknotes.value = []
    total.value = 0
    failed.value = true
  } finally {
    if (current === request) loading.value = false
  }
}

async function loadCountries() {
  try {
    countries.value = (await getCoinCountries('banknote')).data
  } catch {
    // Without them the list is still shown, unfiltered.
  }
}

let searchTimer: ReturnType<typeof setTimeout> | undefined
watch(search, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    if (page.value !== 1) page.value = 1
    else void load()
  }, 300)
})
watch([country, sortKey], () => {
  if (page.value !== 1) page.value = 1
  else void load()
})
watch(page, () => void load())

onMounted(() => {
  void load()
  void loadCountries()
})
onBeforeUnmount(() => clearTimeout(searchTimer))
</script>
