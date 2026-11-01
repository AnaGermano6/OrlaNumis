<template>
  <div class="container-bleed sticky top-[60px] z-(--z-sticky) mb-3 flex items-center gap-2 bg-surface py-2">
    <SearchBar class="!max-w-none flex-1" :model-value="search" @update:model-value="$emit('update:search', $event)" />
    <button
      type="button"
      class="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-sm border border-border-subtle bg-card text-text-secondary transition-all hover:border-gold hover:bg-[var(--accent-gold-dim)] hover:text-gold"
      :class="menuOpen || activeFilters ? 'border-gold text-gold' : ''"
      :aria-label="activeFilters ? `${t('collection.menu.openFilters')} · ${t('collection.menu.activeFilters', { count: activeFilters })}` : t('collection.menu.openFilters')"
      :aria-expanded="menuOpen"
      data-testid="gallery-filters-button"
      @click="$emit('update:menuOpen', true)"
    >
      <SlidersHorizontal :size="22" />
      <span
        v-if="activeFilters"
        class="absolute -top-1.5 -right-1.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-gold px-1 text-sm font-semibold text-surface"
        aria-hidden="true"
        data-testid="gallery-filters-count"
      >{{ activeFilters }}</span>
    </button>
  </div>

  <BaseModal :open="menuOpen" :aria-labelledby="titleId" :padded="false" testid="gallery-filters" @close="$emit('update:menuOpen', false)">
    <div class="flex items-center justify-between gap-3 border-b border-border-subtle px-5 py-3">
      <h2 :id="titleId" class="m-0 text-md text-heading">{{ t('collection.menu.filters') }}</h2>
      <button
        type="button"
        class="inline-flex h-9 w-9 items-center justify-center rounded-sm text-text-muted transition-colors hover:bg-white/5 hover:text-text-primary"
        :aria-label="t('common.close')"
        @click="$emit('update:menuOpen', false)"
      >
        <X :size="20" />
      </button>
    </div>

    <div class="flex flex-col gap-4 px-5 py-4">
      <div class="flex flex-col gap-[0.4rem]">
        <span class="section-label mb-0">{{ t('collection.menu.selection') }}</span>
        <button class="h-10 w-full rounded-sm border border-border-subtle bg-card px-[0.7rem] text-left text-body font-medium text-text-secondary transition-all hover:border-gold" :class="selectMode ? 'border-gold bg-[var(--accent-gold-dim)] text-gold' : ''" @click="$emit('toggle-select-mode')">
          {{ selectMode ? t('collection.menu.exitSelection') : t('collection.menu.enableSelection') }}
        </button>
      </div>
      <div class="flex flex-col gap-[0.4rem]">
        <span class="section-label mb-0">{{ t('collection.menu.denomination') }}</span>
        <DenominationFilter :model-value="selectedDenomination" :denominations="denominationOptions" @update:model-value="$emit('update:selectedDenomination', $event)" />
      </div>
      <div v-if="userSets.length" class="flex flex-col gap-[0.4rem]">
        <span class="section-label mb-0">{{ t('collection.menu.set') }}</span>
        <select :value="selectedSet" @change="$emit('update:selectedSet', ($event.target as HTMLSelectElement).value)" class="form-select h-10 w-full bg-card py-0 text-body">
          <option value="">{{ t('collection.allSets') }}</option>
          <option v-for="set in userSets" :key="set.id" :value="String(set.id)">{{ set.name }}</option>
        </select>
      </div>
      <div class="flex flex-col gap-[0.4rem]">
        <span class="section-label mb-0">{{ t('collection.menu.sort') }}</span>
        <SortSelect :model-value="sortKey" @update:model-value="$emit('update:sortKey', $event)" />
      </div>
      <slot name="controls" />
      <!-- One choice of view: a grid, a list, or one coin at a time to swipe. -->
      <div class="flex flex-col gap-[0.4rem]">
        <span class="section-label mb-0">{{ t('collection.menu.view') }}</span>
        <div class="flex h-10 w-full rounded-sm border border-border-subtle bg-input p-[2px]" role="group" :aria-label="t('collection.menu.view')">
          <button
            v-for="option in VIEWS"
            :key="option.id"
            type="button"
            class="inline-flex min-w-0 flex-1 items-center justify-center gap-1 rounded-[6px] px-2 text-chip transition-all"
            :class="currentView === option.id ? 'bg-gold font-semibold text-surface' : 'font-medium text-text-secondary hover:bg-card-hover hover:text-text-primary'"
            :aria-pressed="currentView === option.id"
            :data-testid="`gallery-view-${option.id}`"
            @click="chooseView(option.id)"
          >
            <component :is="option.icon" :size="16" aria-hidden="true" />
            {{ t(option.label) }}
          </button>
        </div>
      </div>
      <div v-if="currentView === 'grid'" class="flex flex-col gap-[0.4rem]">
        <span class="section-label mb-0">{{ t('collection.menu.face') }}</span>
        <div class="flex h-10 w-full rounded-sm border border-border-subtle bg-input p-[2px]" role="group" :aria-label="t('collection.menu.face')">
          <button
            v-for="side in (['obverse', 'reverse'] as const)"
            :key="side"
            type="button"
            class="min-w-0 flex-1 rounded-[6px] px-2 text-chip transition-all"
            :class="gridSide === side ? 'bg-gold font-semibold text-surface' : 'font-medium text-text-secondary hover:bg-card-hover hover:text-text-primary'"
            :aria-pressed="gridSide === side"
            @click="$emit('update:gridSide', gridSide === side ? null : side)"
          >
            {{ t(`common.${side}`) }}
          </button>
        </div>
      </div>
    </div>

    <!-- Back to the coins, saying how many the filters keep. -->
    <div class="sticky bottom-0 mt-auto border-t border-border-subtle bg-card px-5 py-3">
      <button type="button" class="btn btn-primary w-full justify-center" data-testid="gallery-filters-done" @click="$emit('update:menuOpen', false)">
        {{ t('collection.menu.showResults', { count: total }) }}
      </button>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
// The gallery's top on a phone: the search, which stays in sight, and a
// button that opens every filter and view option in a sheet, showing how many
// filters are on.
import { computed, useId } from 'vue'
import type { CollectionSetOption, ImageType } from '@/types'
import type { CoinDenominationCount } from '@/api/client'
import type { CollectionLayout } from '@/components/collection/CollectionViewControls.vue'
import type { MessageKey } from '@/i18n'
import DenominationFilter from '@/components/collection/DenominationFilter.vue'
import SearchBar from '@/components/SearchBar.vue'
import SortSelect from '@/components/SortSelect.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import { Layers, LayoutGrid, List, SlidersHorizontal, X } from 'lucide-vue-next'
import { t } from '@/i18n'

const props = withDefaults(defineProps<{
  search: string
  selectMode: boolean
  menuOpen: boolean
  selectedDenomination: string
  selectedSet: string
  userSets: CollectionSetOption[]
  denominationOptions: CoinDenominationCount[]
  sortKey: string
  viewMode: 'grid' | 'swipe'
  layout: CollectionLayout
  gridSide: ImageType | null
  /** How many filters are on, shown on the button. */
  activeFilters?: number
  /** How many coins the filters keep. */
  total?: number
}>(), { activeFilters: 0, total: 0 })

const emit = defineEmits<{
  'update:search': [value: string]
  'update:menuOpen': [value: boolean]
  'update:selectedDenomination': [value: string]
  'update:selectedSet': [value: string]
  'update:sortKey': [value: string]
  'update:viewMode': [value: 'grid' | 'swipe']
  'update:layout': [value: CollectionLayout]
  'update:gridSide': [value: ImageType | null]
  'toggle-select-mode': []
}>()

const titleId = useId()

type View = 'grid' | 'list' | 'swipe'
const VIEWS: { id: View, label: MessageKey, icon: typeof LayoutGrid }[] = [
  { id: 'grid', label: 'collection.view.grid', icon: LayoutGrid },
  { id: 'list', label: 'collection.view.list', icon: List },
  { id: 'swipe', label: 'collection.menu.swipe', icon: Layers },
]

const currentView = computed<View>(() => (props.viewMode === 'swipe' ? 'swipe' : props.layout))

function chooseView(view: View) {
  if (view === 'swipe') {
    emit('update:viewMode', 'swipe')
    return
  }
  emit('update:viewMode', 'grid')
  emit('update:layout', view)
}
</script>
