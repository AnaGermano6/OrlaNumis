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
        <div v-if="isPwa" class="pwa-actions">
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

    <!-- Every commemorative and collector coin of each country, with the collection's own -->
    <RouterLink
      to="/sets/commemoratives"
      class="mb-4 flex items-center gap-3 rounded-md border border-border-subtle bg-card p-4 no-underline transition-colors hover:border-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
      data-testid="sets-catalogue-link"
    >
      <BookOpenCheck :size="24" class="shrink-0 text-gold" />
      <span class="min-w-0">
        <span class="block font-medium text-text-primary">{{ t('sets.catalogue.title') }}</span>
        <span class="block text-body text-text-secondary">{{ t('sets.catalogue.link') }}</span>
      </span>
      <ChevronRight :size="20" class="ml-auto shrink-0 text-text-muted" />
    </RouterLink>

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
      <!-- The sets one by one, or their places country by country. -->
      <div v-if="hasCountries" class="mb-4 flex h-10 max-w-[320px] rounded-sm border border-border-subtle bg-input p-[2px]" role="group" :aria-label="t('sets.byCountry.viewLabel')" data-testid="sets-view">
        <button
          v-for="option in VIEWS"
          :key="option"
          type="button"
          class="min-w-0 flex-1 rounded-[6px] px-2 text-chip transition-all"
          :class="view === option ? 'bg-gold text-surface font-semibold' : 'bg-transparent font-medium text-text-secondary hover:bg-card-hover hover:text-text-primary'"
          :aria-pressed="view === option"
          :data-testid="`sets-view-${option}`"
          @click="view = option"
        >
          {{ t(`sets.byCountry.views.${option}`) }}
        </button>
      </div>
      <SetsByCountry v-if="hasCountries && view === 'countries'" :sets="sets" />
      <div v-else class="flex flex-col gap-4">
        <SetDashboardCard
          v-for="set in sets"
          :key="set.id"
          :set="set"
          @click="goToSet(set.id)"
        />
      </div>
    </template>

    <Teleport to="body">
      <div v-if="showCreateModal" class="fixed inset-0 z-[1000] flex items-center justify-center bg-black/60 px-4 py-8" @click.self="showCreateModal = false">
        <div class="card max-h-[85vh] w-full max-w-[520px] overflow-y-auto">
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
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { BookOpenCheck, ChevronRight, CirclePlus, Layers3, Plus, X } from 'lucide-vue-next'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { getApiErrorMessage, getSets, createSet as createSetApi, createSetFromCsv } from '@/api/client'
import type { CoinSetSummary, CreateCoinSetRequest } from '@/types'
import SetDashboardCard from '@/components/sets/SetDashboardCard.vue'
import SetsByCountry from '@/components/sets/SetsByCountry.vue'
import SetCreationWizard from '@/components/sets/SetCreationWizard.vue'
import { usePwa } from '@/composables/usePwa'
import { randomSetColor } from '@/utils/setColors'
import { useDialog } from '@/composables/useDialog'
import { t } from '@/i18n'

const router = useRouter()
const { isPwa } = usePwa()
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
