<template>
  <div class="container">
    <DesktopPageContext :title="t('banknotes.title')">
      <template #actions>
        <AppIconButton :title="selectLabel" :active="selectMode" @click="toggleSelectMode">
          <CheckSquare :size="20" />
        </AppIconButton>
        <AppIconButton to="/banknotes/add" :title="t('banknotes.add')">
          <CirclePlus :size="20" />
        </AppIconButton>
      </template>
      <header class="page-header">
        <h1>{{ t('banknotes.title') }}</h1>
        <span class="ml-auto flex flex-wrap gap-2">
          <button type="button" class="btn btn-secondary btn-sm" :aria-pressed="selectMode" data-testid="banknotes-select" @click="toggleSelectMode">{{ selectLabel }}</button>
          <router-link to="/banknotes/add" class="btn btn-primary btn-sm" data-testid="banknotes-add">{{ t('banknotes.add') }}</router-link>
        </span>
      </header>
    </DesktopPageContext>


    <div class="mb-4 flex flex-col gap-3 rounded-md border border-border-subtle bg-card p-4 shadow-card">
      <SearchBar v-model="search" class="!max-w-none" :placeholder="t('banknotes.searchPlaceholder')" />
      <!-- Every filter the same size, as in the coin gallery. -->
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2" :class="showClassification ? 'lg:grid-cols-5' : 'lg:grid-cols-4'" data-testid="banknotes-filters">
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
        <label class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('collection.view.groupBy') }}</span>
          <select v-model="groupBy" class="form-select h-10 w-full bg-card py-0 text-body" data-testid="banknotes-group-by">
            <option v-for="option in GROUPINGS" :key="option" :value="option">{{ t(GROUPING_LABELS[option]) }}</option>
          </select>
        </label>
        <div class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('coinFields.country') }}</span>
          <CountrySelect v-model="country" :options="countryChoices" :label="t('coinFields.country')" data-testid="banknotes-country" />
        </div>
        <!-- Numista's kinds of banknote in the collection, when it says any. -->
        <label v-if="showClassification" class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('banknotes.classification') }}</span>
          <select v-model="classification" class="form-select h-10 w-full bg-card py-0 text-body" data-testid="banknotes-classification">
            <option value="">{{ t('collection.view.allClassifications') }}</option>
            <option v-for="option in classifications" :key="option.key" :value="option.key">{{ collectionClassificationLabel(option) }} ({{ option.count }})</option>
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
      <!-- Opened from the Collection page's duplicates to trade. -->
      <div v-if="repeated" class="flex flex-wrap items-center gap-2">
        <button
          type="button"
          class="chip active inline-flex items-center gap-1"
          :aria-label="t('collection.activeFilters.remove', { filter: t('banknotes.repeatedOnly') })"
          data-testid="banknotes-filter-repeated"
          @click="clearRepeated"
        >
          {{ t('banknotes.repeatedOnly') }} <X :size="12" aria-hidden="true" />
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div>
      <p>{{ t('collection.loading') }}</p>
    </div>
    <p v-else-if="failed" class="text-body text-text-secondary" role="alert">{{ t('banknotes.failed') }}</p>
    <template v-else-if="banknotes.length">
      <!-- Select mode: every banknote shown, or none, at once. -->
      <div v-if="selectMode" class="mb-3 flex items-center gap-2" data-testid="banknotes-select-bar">
        <button type="button" class="btn btn-sm btn-secondary" @click="selectAll">{{ t('collection.selectAll') }}</button>
        <button type="button" class="btn btn-sm btn-secondary" @click="deselectAll">{{ t('collection.deselectAll') }}</button>
        <span class="ml-2 text-body text-text-secondary" data-testid="banknotes-selected-count">{{ t('collection.selectedCount', { count: selectedIds.size }) }}</span>
      </div>
      <!-- By country, year or rarity, as the coin gallery viewed by one of them. -->
      <section v-for="group in groups" :key="group.key" class="mb-8 last:mb-0" :aria-label="group.label" data-testid="banknotes-group">
        <h2 class="mb-3 flex items-baseline gap-2 border-b border-border-subtle pb-1 font-display text-lg font-medium text-heading">
          <span v-if="groupBy === 'country' && hasCountryFlag(group.label)" class="font-sans" aria-hidden="true" data-testid="group-flag"><CountryFlag :country="group.label" /></span>
          {{ group.label }}
          <span class="font-sans text-chip font-normal text-text-muted">{{ t('banknotes.groupCount', { count: group.coins.length }) }}</span>
        </h2>
        <CoinListTable
          v-if="layout === 'list'"
          :coins="group.coins"
          :show-presentation="false"
          :selectable="selectMode"
          :selected-ids="selectedIds"
          @toggle-select="toggleSelect"
        />
        <div v-else class="coins-grid" data-testid="banknotes-grid">
          <CoinCard
            v-for="banknote in group.coins"
            :key="banknote.id"
            :coin="banknote"
            :selectable="selectMode"
            :selected="selectedIds.has(banknote.id)"
            @toggle-select="toggleSelect"
          />
        </div>
      </section>
      <p v-if="banknotes.length < total" class="mt-4 text-body text-text-secondary" role="status">{{ t('collection.groups.truncated', { shown: banknotes.length, total }) }}</p>
    </template>
    <div v-else class="empty-state" data-testid="banknotes-empty">
      <h3>{{ filtered ? t('banknotes.empty.noMatchTitle') : t('banknotes.empty.title') }}</h3>
      <p>{{ filtered ? t('banknotes.empty.noMatchBody') : t('banknotes.empty.body') }}</p>
      <router-link v-if="!filtered" to="/banknotes/add" class="btn btn-primary mt-4">{{ t('banknotes.add') }}</router-link>
    </div>

    <div class="h-16" aria-hidden="true"></div>

    <BulkActionBar kind="banknote" :visible="selectMode && selectedIds.size > 0" :selected-count="selectedIds.size" @set="showSetPicker = true" @delete="deleteSelected" />
    <BulkSetPickerModal :open="showSetPicker" :sets="standardSets" @select="addSelectedToSet" @close="showSetPicker = false" />
  </div>
</template>

<script setup lang="ts">
// The banknotes of the collection, apart from its coins, country by country:
// a search, a country filter, an order and a grid or list (their numbers are
// on the Collection page). Their own filters, so the coin gallery's never
// leak in.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { bulkAction, getCoinCountries, getCoins, getCollectionSummary, getSets, type CoinCountryCount, type CollectionClassification } from '@/api/client'
import { normalizeCoinSetType, type Coin, type CollectionSetOption } from '@/types'
import { CheckSquare, CirclePlus, X } from 'lucide-vue-next'
import BulkActionBar from '@/components/BulkActionBar.vue'
import BulkSetPickerModal from '@/components/BulkSetPickerModal.vue'
import CoinCard from '@/components/CoinCard.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import CoinListTable from '@/components/collection/CoinListTable.vue'
import SearchBar from '@/components/SearchBar.vue'
import { t } from '@/i18n'
import { useBulkSelect } from '@/composables/useBulkSelect'
import { useDialog } from '@/composables/useDialog'
import { collectionClassificationLabel, groupCoins, type CoinGroupBy } from '@/utils/coinAttributes'
import type { MessageKey } from '@/i18n'
import CountryFlag from '@/components/CountryFlag.vue'
import CountrySelect, { type CountrySelectOption } from '@/components/CountrySelect.vue'
import { hasCountryFlag } from '@/utils/countryFlags'

// Every banknote is loaded, to show each country's together: pages of 100, up to 2000.
const PAGE_SIZE = 100
const MAX_BANKNOTES = 2000
const LAYOUTS = ['grid', 'list'] as const
const LAYOUT_KEY = 'banknotes:layout'
// The banknotes are always shown in groups: by country (at first), year or rarity.
const GROUPINGS = ['country', 'year', 'rarity'] as const
type BanknoteGrouping = (typeof GROUPINGS)[number] & CoinGroupBy
const GROUPING_LABELS: Record<BanknoteGrouping, MessageKey> = {
  country: 'coinFields.country',
  year: 'collection.view.groupYear',
  rarity: 'coinFields.rarity',
}
const GROUP_BY_KEY = 'banknotes:groupBy'

const banknotes = ref<Coin[]>([])
const total = ref(0)
const loading = ref(true)
const failed = ref(false)
const countries = ref<CoinCountryCount[]>([])

const route = useRoute()
const router = useRouter()

const search = ref('')
const country = ref('')
// Only the banknotes held more than once: "?repeated=true".
const repeated = ref(route.query.repeated === 'true')
const classification = ref('')
const classifications = ref<CollectionClassification[]>([])
const showClassification = computed(() => classifications.value.length > 0 || classification.value !== '')
const groupBy = ref<BanknoteGrouping>(readGroupBy())
const sortKey = ref('updated_at_desc')
const layout = ref<(typeof LAYOUTS)[number]>(readLayout())

// "All countries", then each country with its flag and count.
const countryChoices = computed((): CountrySelectOption[] => [
  { value: '', label: t('collection.view.allCountries') },
  ...countries.value.map((option) => ({ value: option.country, label: `${option.country} (${option.count})`, country: option.country })),
])
const filtered = computed(() => Boolean(search.value.trim() || country.value || repeated.value || classification.value))
// By country (by name), year (oldest first) or rarity, as the coin gallery; those without one last.
const groups = computed(() => groupCoins(banknotes.value, groupBy.value))

// Select mode, as in the coin gallery: banknotes to add to a set or delete
// together. The app's flag keeps the floating assistant out of the way.
const { bulkSelectActive } = useBulkSelect()
const { showAlert, showConfirm } = useDialog()
const selectMode = ref(false)
const selectedIds = ref(new Set<number>())
const showSetPicker = ref(false)
const standardSets = ref<CollectionSetOption[]>([])
const selectLabel = computed(() => (selectMode.value ? t('app.header.cancelSelection') : t('banknotes.select')))

function toggleSelectMode() {
  selectMode.value = !selectMode.value
  bulkSelectActive.value = selectMode.value
  selectedIds.value = new Set()
  if (selectMode.value && !standardSets.value.length) void loadSets()
}

function toggleSelect(id: number) {
  const next = new Set(selectedIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedIds.value = next
}

function selectAll() {
  selectedIds.value = new Set(banknotes.value.map((banknote) => banknote.id))
}

function deselectAll() {
  selectedIds.value = new Set()
}

function leaveSelectMode() {
  selectMode.value = false
  bulkSelectActive.value = false
  selectedIds.value = new Set()
}

async function loadSets() {
  try {
    standardSets.value = ((await getSets()).data?.sets ?? [])
      .filter((set) => normalizeCoinSetType(set.setType) === 'standard')
      .map((set) => ({ id: set.id, name: set.name, color: set.color }))
  } catch {
    standardSets.value = []
  }
}

async function deleteSelected() {
  const count = selectedIds.value.size
  if (!await showConfirm(t('banknotes.bulk.confirmDelete', { count }), { title: t('common.delete'), variant: 'danger' })) return
  try {
    await bulkAction([...selectedIds.value], 'delete')
    leaveSelectMode()
    void load()
    void loadCountries()
    void loadClassifications()
  } catch {
    await showAlert(t('banknotes.bulk.deleteFailed'), { title: t('common.error') })
  }
}

async function addSelectedToSet(setId: number) {
  showSetPicker.value = false
  try {
    await bulkAction([...selectedIds.value], 'set', { setId })
    leaveSelectMode()
  } catch {
    await showAlert(t('collection.bulk.applySetFailed'), { title: t('common.error') })
  }
}

function clearRepeated() {
  repeated.value = false
  const query = { ...route.query }
  delete query.repeated
  void router.replace({ query })
}

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

function readGroupBy(): BanknoteGrouping {
  try {
    const saved = localStorage.getItem(GROUP_BY_KEY)
    return GROUPINGS.find((option) => option === saved) ?? 'country'
  } catch {
    return 'country'
  }
}

watch(groupBy, (value) => {
  try {
    localStorage.setItem(GROUP_BY_KEY, value)
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
  const params = {
    kind: 'banknote' as const,
    wishlist: 'false',
    search: search.value.trim() || undefined,
    country: country.value || undefined,
    repeated: repeated.value ? 'true' : undefined,
    classification: classification.value || undefined,
    sort,
    order,
  }
  try {
    const all: Coin[] = []
    let reported = 0
    for (let page = 1; all.length < MAX_BANKNOTES; page += 1) {
      const res = await getCoins({ ...params, page, limit: PAGE_SIZE })
      if (current !== request) return
      const coins = res.data.coins ?? []
      reported = res.data.total
      all.push(...coins)
      if (!coins.length || all.length >= reported) break
    }
    banknotes.value = all
    total.value = reported
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

// Numista's kinds of banknote in the collection, from its cached types.
async function loadClassifications() {
  try {
    classifications.value = (await getCollectionSummary('banknote')).data.classifications ?? []
  } catch {
    classifications.value = []
  }
}

let searchTimer: ReturnType<typeof setTimeout> | undefined
watch(search, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => void load(), 300)
})
watch([country, sortKey, repeated, classification], () => void load())

onMounted(() => {
  void load()
  void loadCountries()
  void loadClassifications()
})
onBeforeUnmount(() => {
  clearTimeout(searchTimer)
  if (selectMode.value) bulkSelectActive.value = false
})
</script>
