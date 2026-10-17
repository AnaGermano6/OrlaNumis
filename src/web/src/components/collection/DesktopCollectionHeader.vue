<template>
  <div class="sticky top-[60px] z-50 mx-[-2rem] bg-surface px-8 pb-2">
    <div class="mb-4 flex flex-col gap-3 rounded-md border border-border-subtle bg-card p-4 shadow-card">
      <SearchBar
        class="!max-w-none"
        :model-value="search"
        @update:model-value="$emit('update:search', $event)"
      />

      <!-- Every filter the same size: a grid of equal columns, a field per cell. -->
      <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5" data-testid="collection-filters">
        <label class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('collection.menu.denomination') }}</span>
          <DenominationFilter :model-value="selectedDenomination" :denominations="denominationOptions" @update:model-value="$emit('update:selectedDenomination', $event)" />
        </label>
        <label v-if="userSets.length" class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('collection.menu.set') }}</span>
          <select :value="selectedSet" @change="$emit('update:selectedSet', ($event.target as HTMLSelectElement).value)" class="form-select h-10 w-full bg-card py-0 text-body transition-colors hover:border-border-accent">
            <option value="">{{ t('collection.allSets') }}</option>
            <option v-for="set in userSets" :key="set.id" :value="String(set.id)">{{ set.name }}</option>
          </select>
        </label>
        <label class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('collection.menu.sort') }}</span>
          <SortSelect :model-value="sortKey" @update:model-value="$emit('update:sortKey', $event)" />
        </label>

        <slot name="controls" />

        <div class="flex min-w-0 flex-col gap-1">
          <span class="section-label mb-0">{{ t('collection.menu.face') }}</span>
          <div class="flex h-10 w-full rounded-sm border border-border-subtle bg-input p-[2px]" role="group" :aria-label="t('collection.menu.face')">
            <button
              type="button"
              class="min-w-0 flex-1 rounded-[6px] bg-transparent px-2 text-chip font-medium text-text-secondary transition-all hover:bg-card-hover hover:text-text-primary"
              :class="gridSide === 'obverse' ? 'bg-gold text-surface hover:bg-gold hover:text-surface font-semibold' : ''"
              :aria-pressed="gridSide === 'obverse'"
              @click="$emit('update:gridSide', gridSide === 'obverse' ? null : 'obverse')"
            >
              {{ t('common.obverse') }}
            </button>
            <button
              type="button"
              class="min-w-0 flex-1 rounded-[6px] bg-transparent px-2 text-chip font-medium text-text-secondary transition-all hover:bg-card-hover hover:text-text-primary"
              :class="gridSide === 'reverse' ? 'bg-gold text-surface hover:bg-gold hover:text-surface font-semibold' : ''"
              :aria-pressed="gridSide === 'reverse'"
              @click="$emit('update:gridSide', gridSide === 'reverse' ? null : 'reverse')"
            >
              {{ t('common.reverse') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { CollectionSetOption, ImageType } from '@/types'
import type { CoinDenominationCount } from '@/api/client'
import DenominationFilter from '@/components/collection/DenominationFilter.vue'
import SearchBar from '@/components/SearchBar.vue'
import SortSelect from '@/components/SortSelect.vue'
import { t } from '@/i18n'

defineProps<{
  search: string
  selectedDenomination: string
  selectedSet: string
  userSets: CollectionSetOption[]
  denominationOptions: CoinDenominationCount[]
  sortKey: string
  gridSide: ImageType | null
}>()

defineEmits<{
  'update:search': [value: string]
  'update:selectedDenomination': [value: string]
  'update:selectedSet': [value: string]
  'update:sortKey': [value: string]
  'update:gridSide': [value: ImageType | null]
}>()
</script>
