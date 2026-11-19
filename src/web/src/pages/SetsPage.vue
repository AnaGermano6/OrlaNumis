<template>
  <div class="container">
    <DesktopPageContext :title="t('nav.sets')">
      <template #actions>
        <AppIconButton :title="t('sets.list.create')" @click="openCreateModal">
          <CirclePlus :size="20" />
        </AppIconButton>
      </template>
      <div class="page-header">
        <h1>{{ t('nav.sets') }}</h1>
        <div v-if="isMobile" class="pwa-actions">
          <button class="pwa-icon-btn" :title="t('sets.list.create')" @click="openCreateModal">
            <CirclePlus :size="22" />
          </button>
        </div>
        <div v-else class="header-actions">
          <button class="btn btn-primary" @click="openCreateModal">
            <Plus :size="16" /> {{ t('sets.list.create') }}
          </button>
        </div>
      </div>
    </DesktopPageContext>

    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div>
      <p>{{ t('sets.list.loading') }}</p>
    </div>

    <div v-else-if="sets.length === 0" class="empty-state">
      <Layers3 :size="48" />
      <h3>{{ t('sets.list.emptyTitle') }}</h3>
      <p>{{ t('sets.list.emptyBody') }}</p>
      <button class="btn btn-primary mt-4" @click="openCreateModal">
        <Plus :size="16" /> {{ t('sets.list.createFirst') }}
      </button>
    </div>

    <template v-else>
      <!-- The sets one by one, or their places country by country; at the right, all of them, those to complete or the complete ones. -->
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div v-if="hasCountries" class="flex h-11 w-full max-w-[320px] rounded-sm border border-border-subtle bg-input p-[3px]" role="group" :aria-label="t('sets.byCountry.viewLabel')" data-testid="sets-view">
          <button
            v-for="option in VIEWS"
            :key="option"
            type="button"
            class="min-w-0 flex-1 rounded-[6px] px-2 text-body transition-all"
            :class="view === option ? 'bg-gold text-surface font-semibold' : 'bg-transparent font-medium text-text-secondary hover:bg-card-hover hover:text-text-primary'"
            :aria-pressed="view === option"
            :data-testid="`sets-view-${option}`"
            @click="view = option"
          >
            {{ t(`sets.byCountry.views.${option}`) }}
          </button>
        </div>
        <!-- By country: a country searched for, and their order. -->
        <div v-if="byCountry" class="ml-auto flex w-full flex-wrap items-center justify-end gap-3 sm:w-auto">
          <SearchBar v-model="countrySearch" box class="!max-w-none w-full sm:w-64" :placeholder="t('sets.byCountry.search')" data-testid="sets-country-search" />
          <select v-model="countrySort" class="form-select !w-auto min-w-44" :aria-label="t('sets.byCountry.sort.label')" data-testid="sets-country-sort">
            <option v-for="option in COUNTRY_SORTS" :key="option" :value="option">{{ t(`sets.byCountry.sort.${option}`) }}</option>
          </select>
        </div>
        <select v-else v-model="status" class="form-select ml-auto !w-auto min-w-44" :aria-label="t('sets.list.status.label')" data-testid="sets-status">
          <option v-for="option in STATUSES" :key="option" :value="option">{{ t(`sets.list.status.${option}`) }}</option>
        </select>
      </div>
      <SetsByCountry v-if="byCountry" :sets="sets" :search="countrySearch" :sort="countrySort" />
      <p v-else-if="!shownSets.length" class="empty-state" data-testid="sets-status-none">{{ t('sets.list.status.none') }}</p>
      <div v-else class="grid gap-4 lg:grid-cols-2" data-testid="sets-cards">
        <SetDashboardCard
          v-for="set in shownSets"
          :key="set.id"
          :set="set"
          @click="goToSet(set.id)"
        />
      </div>
    </template>

    <BaseModal :open="showCreateModal" :aria-label="t('sets.list.createNew')" max-width="md:max-w-[520px]" @close="showCreateModal = false">
      <div class="mb-6 flex items-center justify-between gap-4">
        <h2 class="m-0 text-lg text-text-primary">{{ t('sets.list.createNew') }}</h2>
        <button class="inline-flex rounded-sm p-1 text-text-secondary transition-colors hover:text-text-primary" @click="showCreateModal = false">
          <X :size="20" />
        </button>
      </div>
      <SetCreationWizard
        :initial-value="newSet"
        :submit-label="t('common.create')"
        @submit="createSet"
        @cancel="showCreateModal = false"
      />
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { CirclePlus, Layers3, Plus, X } from 'lucide-vue-next'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { getApiErrorMessage, getSets, createSet as createSetApi, createSetFromCsv } from '@/api/client'
import type { CoinSetSummary, CreateCoinSetRequest } from '@/types'
import SetDashboardCard from '@/components/sets/SetDashboardCard.vue'
import SetsByCountry, { type CountrySort } from '@/components/sets/SetsByCountry.vue'
import SearchBar from '@/components/SearchBar.vue'
import SetCreationWizard from '@/components/sets/SetCreationWizard.vue'
import { useLayout } from '@/composables/useLayout'
import { randomSetColor } from '@/utils/setColors'
import { useDialog } from '@/composables/useDialog'
import { t } from '@/i18n'
import BaseModal from '@/components/ui/BaseModal.vue'

const router = useRouter()
const { isMobile } = useLayout()
const { showAlert } = useDialog()
const loading = ref(true)
const sets = ref<CoinSetSummary[]>([])

const VIEWS = ['sets', 'countries'] as const
const VIEW_KEY = 'sets:view'
const view = ref<(typeof VIEWS)[number]>(readView())
// Only checklist sets have places by country.
const hasCountries = computed(() => sets.value.some((set) => set.countries?.length))

function readView(): (typeof VIEWS)[number] {
  try {
    return localStorage.getItem(VIEW_KEY) === 'countries' ? 'countries' : 'sets'
  } catch {
    return 'sets'
  }
}

const byCountry = computed(() => hasCountries.value && view.value === 'countries')
// By country: the country searched for, and the order of the countries.
const countrySearch = ref('')
const COUNTRY_SORTS: CountrySort[] = ['progress', 'progressAsc', 'missing', 'name']
const countrySort = ref<CountrySort>('progress')

// All the sets, those still to complete (any that cannot be completed too) or the complete ones.
const STATUSES = ['all', 'incomplete', 'complete'] as const
const status = ref<(typeof STATUSES)[number]>('all')
const shownSets = computed(() => sets.value.filter((set) => {
  if (status.value === 'all') return true
  const complete = (set.completionPercentage ?? 0) >= 100
  return status.value === 'complete' ? complete : !complete
}))

watch(view, (value) => {
  try {
    localStorage.setItem(VIEW_KEY, value)
  } catch {
    // The choice still applies for this visit.
  }
})
const showCreateModal = ref(false)
const newSet = ref({
  name: '',
  description: '',
  color: randomSetColor(),
  setType: 'standard' as const,
})

onMounted(async () => {
  await loadSets()
})

async function loadSets() {
  loading.value = true
  try {
    const res = await getSets()
    sets.value = res.data.sets
  } catch (error) {
    console.error('Failed to load sets:', error)
  } finally {
    loading.value = false
  }
}

function resetNewSet() {
  newSet.value = {
    name: '',
    description: '',
    color: randomSetColor(),
    setType: 'standard',
  }
}

function openCreateModal() {
  resetNewSet()
  showCreateModal.value = true
}

async function createSet(value: CreateCoinSetRequest, csv?: string) {
  try {
    if (csv) {
      await createSetFromCsv({ ...value, csv })
    } else {
      // The new set opens, to show its coins or to add them.
      const res = await createSetApi(value)
      showCreateModal.value = false
      resetNewSet()
      goToSet(res.data.id)
      return
    }
    showCreateModal.value = false
    resetNewSet()
    await loadSets()
  } catch (error) {
    console.error('Failed to create set:', error)
    await showAlert(getApiErrorMessage(error) || t('common.requestFailed'), { title: t('sets.list.createFailedTitle') })
  }
}

function goToSet(id: number) {
  router.push({ name: 'set-detail', params: { id } })
}
</script>
