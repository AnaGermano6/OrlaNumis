<template>
  <div class="container-bleed sticky top-[60px] z-(--z-sticky) bg-surface pb-2">
    <div class="mb-4 flex flex-col gap-4 rounded-md border border-border-subtle bg-card p-4 shadow-card">
      <!-- The search, then how the coins show: in a grid or a list, and which side. -->
      <div class="flex flex-wrap items-center gap-3">
        <SearchBar
          class="!max-w-none min-w-64 flex-1"
          box
          :model-value="search"
          @update:model-value="$emit('update:search', $event)"
        />
        <ViewToggle
          :model-value="layout"
          :options="layouts"
          :label="t('collection.view.layout')"
          test-id-prefix="gallery-layout"
          @update:model-value="$emit('update:layout', $event ?? 'grid')"
        />
        <ViewToggle
          :model-value="gridSide"
          :options="sides"
          :label="t('collection.menu.face')"
          clearable
          test-id-prefix="gallery-side"
          @update:model-value="$emit('update:gridSide', $event)"
        />
      </div>

      <!-- The usual filters, each the same size; the others a click away. -->
      <div class="grid grid-cols-2 items-end gap-3 lg:grid-cols-[repeat(4,minmax(0,1fr))_auto]" data-testid="collection-filters">
        <label class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('collection.menu.denomination') }}</span>
          <DenominationFilter :model-value="selectedDenomination" :denominations="denominationOptions" @update:model-value="$emit('update:selectedDenomination', $event)" />
        </label>
        <label class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('collection.menu.sort') }}</span>
          <SortSelect :model-value="sortKey" @update:model-value="$emit('update:sortKey', $event)" />
        </label>

        <slot name="main" />

        <button
          type="button"
          class="col-span-2 h-10 cursor-pointer justify-self-start whitespace-nowrap rounded-sm px-1 text-body font-medium text-gold underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2 lg:col-span-1"
          :aria-expanded="moreOpen"
          aria-controls="collection-more-filters"
          data-testid="collection-more-toggle"
          @click="moreOpen = !moreOpen"
        >
          {{ moreOpen ? t('collection.view.fewerFilters') : moreActive ? t('collection.view.moreFiltersOn', { count: moreActive }) : t('collection.view.moreFilters') }}
        </button>
      </div>

      <div v-show="moreOpen" id="collection-more-filters" class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5" data-testid="collection-more-filters">
        <label v-if="userSets.length" class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('collection.menu.set') }}</span>
          <select :value="selectedSet" class="form-select h-10 w-full bg-card py-0 text-body transition-colors hover:border-border-accent" @change="$emit('update:selectedSet', ($event.target as HTMLSelectElement).value)">
            <option value="">{{ t('collection.allSets') }}</option>
            <option v-for="set in userSets" :key="set.id" :value="String(set.id)">{{ set.name }}</option>
          </select>
        </label>
        <slot name="more" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// The gallery's search and filters on a wider screen: the search with how
// the coins show (grid or list, which side), then the usual filters (face
// value, order, grouping, country) and, a click away, the others (set,
// period, kind of coin, presentation, rarity), open by themselves while one
// of them is on.
import { computed, ref } from 'vue'
import { LayoutGrid, List, Square } from 'lucide-vue-next'
import type { CollectionSetOption, ImageType } from '@/types'
import type { CoinDenominationCount } from '@/api/client'
import DenominationFilter from '@/components/collection/DenominationFilter.vue'
import ViewToggle from '@/components/collection/ViewToggle.vue'
import SearchBar from '@/components/SearchBar.vue'
import SortSelect from '@/components/SortSelect.vue'
import { t } from '@/i18n'

const props = withDefaults(defineProps<{
  search: string
  selectedDenomination: string
  selectedSet: string
  userSets: CollectionSetOption[]
  denominationOptions: CoinDenominationCount[]
  sortKey: string
  gridSide: ImageType | null
  layout: 'grid' | 'list'
  /** How many of the other filters are on. */
  moreActive?: number
}>(), { moreActive: 0 })

defineEmits<{
  'update:search': [value: string]
  'update:selectedDenomination': [value: string]
  'update:selectedSet': [value: string]
  'update:sortKey': [value: string]
  'update:gridSide': [value: ImageType | null]
  'update:layout': [value: 'grid' | 'list']
}>()

const moreOpen = ref(props.moreActive > 0)

const layouts = computed(() => [
  { value: 'grid' as const, label: t('collection.view.grid'), icon: LayoutGrid },
  { value: 'list' as const, label: t('collection.view.list'), icon: List },
])
const sides = computed(() => [
  { value: 'obverse' as ImageType, label: t('common.obverse'), icon: Square },
  { value: 'reverse' as ImageType, label: t('common.reverse'), icon: Square },
])
</script>
