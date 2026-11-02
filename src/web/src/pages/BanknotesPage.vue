<template>
  <div class="container">
    <DesktopPageContext :title="t('banknotes.title')">
      <template #actions>
        <AppIconButton :title="selectLabel" :active="selectMode" @click="toggleSelectMode">
          <CheckSquare :size="20" />
        </AppIconButton>
        <AppIconButton :title="t('collection.export.button')" data-testid="banknotes-export" @click="showExport = true">
          <FileSpreadsheet :size="20" />
        </AppIconButton>
        <AppIconButton to="/banknotes/add" :title="t('banknotes.add')">
          <CirclePlus :size="20" />
        </AppIconButton>
      </template>
      <header v-if="!isMobile" class="page-header">
        <h1>{{ t('banknotes.title') }}</h1>
        <span class="ml-auto flex flex-wrap gap-2">
          <button type="button" class="btn btn-secondary btn-sm" :aria-pressed="selectMode" data-testid="banknotes-select" @click="toggleSelectMode">{{ selectLabel }}</button>
          <button type="button" class="btn btn-secondary btn-sm" @click="showExport = true">
            <FileSpreadsheet :size="14" aria-hidden="true" /> {{ t('collection.export.button') }}
          </button>
          <router-link to="/banknotes/add" class="btn btn-primary btn-sm" data-testid="banknotes-add">{{ t('banknotes.add') }}</router-link>
        </span>
      </header>
    </DesktopPageContext>

    <!-- On a phone, as the coin gallery: the search, the filters in a sheet,
         and adding a banknote at hand. -->
    <MobileFilterBar
      v-if="isMobile"
      v-model:search="search"
      v-model:menu-open="menuOpen"
      :placeholder="t('banknotes.searchPlaceholder')"
      :active-filters="activeFilterCount"
      :done-label="t('banknotes.showResults', { count: total })"
    >
      <template #actions>
        <router-link
          to="/banknotes/add"
          class="flex h-10 w-10 shrink-0 items-center justify-center rounded-sm border border-border-subtle bg-card text-text-secondary transition-all hover:border-gold hover:bg-[var(--accent-gold-dim)] hover:text-gold"
          :aria-label="t('banknotes.add')"
          :title="t('banknotes.add')"
          data-testid="banknotes-add"
        >
          <CirclePlus :size="22" />
        </router-link>
      </template>
      <div class="flex flex-col gap-[0.4rem]">
        <span class="section-label mb-0">{{ t('collection.menu.selection') }}</span>
        <button
          type="button"
          class="h-10 w-full rounded-sm border border-border-subtle bg-card px-[0.7rem] text-left text-body font-medium text-text-secondary transition-all hover:border-gold"
          :class="selectMode ? 'border-gold bg-[var(--accent-gold-dim)] text-gold' : ''"
          :aria-pressed="selectMode"
          data-testid="banknotes-select"
          @click="toggleSelectMode"
        >
          {{ selectMode ? t('collection.menu.exitSelection') : t('collection.menu.enableSelection') }}
        </button>
      </div>
      <BanknoteFilters
        v-model:sort-key="sortKey"
        v-model:group-by="groupBy"
        v-model:country="country"
        v-model:classification="classification"
        v-model:layout="layout"
        :country-choices="countryChoices"
        :classifications="classifications"
        stacked
      />
      <button type="button" class="btn btn-secondary btn-sm w-full justify-center" @click="showExport = true">
        <FileSpreadsheet :size="14" aria-hidden="true" /> {{ t('collection.export.button') }}
      </button>
    </MobileFilterBar>

    <div v-else class="mb-4 flex flex-col gap-3 rounded-md border border-border-subtle bg-card p-4 shadow-card">
      <SearchBar v-model="search" class="!max-w-none" :placeholder="t('banknotes.searchPlaceholder')" />
      <BanknoteFilters
        v-model:sort-key="sortKey"
        v-model:group-by="groupBy"
        v-model:country="country"
        v-model:classification="classification"
        v-model:layout="layout"
        :country-choices="countryChoices"
        :classifications="classifications"
      />
    </div>

    <!-- Opened from the Collection page's duplicates to trade. -->
    <div v-if="repeated" class="mb-3 flex flex-wrap items-center gap-2">
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
      <!-- All in one list, or by country, monetary system, year or rarity, as the coin gallery. -->
      <section v-for="group in groups" :key="group.key" class="mb-8 last:mb-0" :aria-label="groupBy === 'none' ? t('banknotes.groupNone') : group.label" data-testid="banknotes-group">
        <h2 v-if="groupBy !== 'none'" class="mb-3 flex items-baseline gap-2 border-b border-border-subtle pb-1 font-display text-lg font-medium text-heading">
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
    <ExportDialog :open="showExport" kind="banknote" :params="listParams()" :count="total" :filtered="filtered" @close="showExport = false" />
  </div>
</template>

<script setup lang="ts">
// The banknotes of the collection, apart from its coins, country by country:
// a search, a country filter, an order and a grid or list (their numbers are
// on the Collection page). Their own filters, so the coin gallery's never
// leak in. On a phone it is laid out as the coin gallery: the search on top,
// the filters in a sheet.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { bulkAction, getCoinCountries, getCoins, getCollectionSummary, getSets, type CoinCountryCount, type CollectionClassification } from '@/api/client'
import { normalizeCoinSetType, type Coin, type CollectionSetOption } from '@/types'
import { CheckSquare, CirclePlus, FileSpreadsheet, X } from 'lucide-vue-next'
import BulkActionBar from '@/components/BulkActionBar.vue'
import BulkSetPickerModal from '@/components/BulkSetPickerModal.vue'
import CoinCard from '@/components/CoinCard.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import BanknoteFilters, { type BanknoteLayout } from '@/components/collection/BanknoteFilters.vue'
import CoinListTable from '@/components/collection/CoinListTable.vue'
import MobileFilterBar from '@/components/collection/MobileFilterBar.vue'
import SearchBar from '@/components/SearchBar.vue'
import { t } from '@/i18n'
import { useBulkSelect } from '@/composables/useBulkSelect'
import { useDialog } from '@/composables/useDialog'
import { useLayout } from '@/composables/useLayout'
import { BANKNOTE_GROUPINGS, groupCoins, type BanknoteGrouping } from '@/utils/coinAttributes'
import CountryFlag from '@/components/CountryFlag.vue'
import type { CountrySelectOption } from '@/components/CountrySelect.vue'
import ExportDialog from '@/components/collection/ExportDialog.vue'
import { rememberBrowseSequence } from '@/utils/browseSequence'
import { hasCountryFlag } from '@/utils/countryFlags'

// Every banknote is loaded, to show each country's together: pages of 100, up to 2000.
const PAGE_SIZE = 100
const MAX_BANKNOTES = 2000
const LAYOUT_KEY = 'banknotes:layout'
const GROUP_BY_KEY = 'banknotes:groupBy'

const banknotes = ref<Coin[]>([])
const total = ref(0)
const loading = ref(true)
const failed = ref(false)
const countries = ref<CoinCountryCount[]>([])

const route = useRoute()
const router = useRouter()
const { isMobile } = useLayout()
// The phone's filters sheet.
const menuOpen = ref(false)

const search = ref('')
const country = ref('')
// Only the banknotes held more than once: "?repeated=true".
const repeated = ref(route.query.repeated === 'true')
const classification = ref('')
const classifications = ref<CollectionClassification[]>([])
const groupBy = ref<BanknoteGrouping>(readGroupBy())
const sortKey = ref('updated_at_desc')
const layout = ref<BanknoteLayout>(readLayout())

// "All countries", then each country with its flag and count.
const countryChoices = computed((): CountrySelectOption[] => [
  { value: '', label: t('collection.view.allCountries') },
  ...countries.value.map((option) => ({ value: option.country, label: `${option.country} (${option.count})`, country: option.country })),
])
const filtered = computed(() => Boolean(search.value.trim() || country.value || repeated.value || classification.value))
// How many filters are on (not the search), shown on the phone's filter button.
const activeFilterCount = computed(() => [country.value, repeated.value, classification.value].filter(Boolean).length)
// The spreadsheet of the banknotes shown, or of the whole collection.
const showExport = ref(false)
// By country (by name), year (oldest first) or rarity, as the coin gallery; those without one last.
const groups = computed(() => groupCoins(banknotes.value, groupBy.value))
// A banknote's page goes to the previous and the next one in this order.
watch(groups, (value) => {
  rememberBrowseSequence('banknotes', value.flatMap((group) => group.coins.map((item) => item.id)))
}, { immediate: true })

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

function readLayout(): BanknoteLayout {
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
    return BANKNOTE_GROUPINGS.find((option) => option === saved) ?? 'country'
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

// The list's filters and order, as the server takes them (without the page).
function listParams() {
  const [sort, order] = [sortKey.value.slice(0, sortKey.value.lastIndexOf('_')), sortKey.value.slice(sortKey.value.lastIndexOf('_') + 1)]
  return {
    kind: 'banknote' as const,
    wishlist: 'false',
    search: search.value.trim() || undefined,
    country: country.value || undefined,
    repeated: repeated.value ? 'true' : undefined,
    classification: classification.value || undefined,
    sort,
    order,
  }
}

// Only the latest request's answer is shown.
let request = 0
async function load() {
  const current = ++request
  loading.value = true
  failed.value = false
  const params = listParams()
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
