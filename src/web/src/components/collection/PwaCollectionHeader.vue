<template>
  <div class="sticky top-[60px] z-[150] mb-3 flex items-center gap-2 bg-surface py-2">
    <SearchBar class="!max-w-none flex-1" :model-value="search" @update:model-value="$emit('update:search', $event)" />
    <div class="relative shrink-0">
      <button class="flex h-10 w-10 items-center justify-center rounded-sm border border-border-subtle bg-card text-text-secondary transition-all hover:border-gold hover:bg-[var(--accent-gold-dim)] hover:text-gold" @click="$emit('update:menuOpen', !menuOpen)" :class="menuOpen ? 'border-gold bg-[var(--accent-gold-dim)] text-gold' : ''">
        <SlidersHorizontal :size="22" />
      </button>
      <Transition
        enter-active-class="transition-all duration-200 ease-in-out"
        enter-from-class="-translate-y-2 opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-active-class="transition-all duration-200 ease-in-out"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="-translate-y-2 opacity-0"
      >
        <div v-if="menuOpen" class="absolute right-0 top-[calc(100%+0.5rem)] z-[100] flex min-w-[260px] flex-col gap-3 rounded-md border border-border-subtle bg-card p-4 shadow-[0_8px_30px_var(--overlay-40)]">
          <div class="flex flex-col gap-[0.4rem]">
            <span class="section-label mb-0">{{ t('collection.menu.selection') }}</span>
            <button class="w-full rounded-sm border border-border-subtle bg-card px-[0.7rem] py-2 text-left text-body font-medium text-text-secondary transition-all hover:border-gold" :class="selectMode ? 'border-gold bg-[var(--accent-gold-dim)] text-gold' : ''" @click="$emit('toggle-select-mode')">
              {{ selectMode ? t('collection.menu.exitSelection') : t('collection.menu.enableSelection') }}
            </button>
          </div>
          <div class="flex flex-col gap-[0.4rem]">
            <span class="section-label mb-0">{{ t('collection.menu.denomination') }}</span>
            <DenominationFilter :model-value="selectedDenomination" :denominations="denominationOptions" @update:model-value="$emit('update:selectedDenomination', $event)" />
          </div>
          <div v-if="userTags.length" class="flex flex-col gap-[0.4rem]">
            <span class="section-label mb-0">{{ t('collection.menu.set') }}</span>
            <select :value="selectedTag" @change="$emit('update:selectedTag', ($event.target as HTMLSelectElement).value)" class="form-select w-full bg-card text-body">
              <option value="">{{ t('collection.allSets') }}</option>
              <option v-for="tag in userTags" :key="tag.filterValue" :value="tag.filterValue">{{ tag.name }}</option>
            </select>
          </div>
          <div class="flex flex-col gap-[0.4rem]">
            <span class="section-label mb-0">{{ t('collection.menu.sort') }}</span>
            <SortSelect :model-value="sortKey" @update:model-value="$emit('update:sortKey', $event)" />
          </div>
          <slot name="controls" />
          <div class="flex flex-col gap-[0.4rem]">
            <span class="section-label mb-0">{{ t('collection.menu.view') }}</span>
            <div class="flex flex-wrap items-center gap-2">
              <div class="flex overflow-hidden rounded-sm border border-border-subtle">
                <button class="bg-card px-[0.6rem] py-1.5 text-text-secondary transition-all hover:bg-card-hover" :class="viewMode === 'swipe' ? 'bg-[var(--accent-gold-dim)] text-gold' : ''" @click="$emit('update:viewMode', 'swipe')" :title="t('collection.menu.swipeView')">
                  <Layers :size="18" />
                </button>
                <button class="bg-card px-[0.6rem] py-1.5 text-text-secondary transition-all hover:bg-card-hover" :class="viewMode === 'grid' ? 'bg-[var(--accent-gold-dim)] text-gold' : ''" @click="$emit('update:viewMode', 'grid')" :title="t('collection.menu.gridView')">
                  <LayoutGrid :size="18" />
                </button>
              </div>
            </div>
          </div>
          <div v-if="viewMode === 'grid'" class="flex flex-col gap-[0.4rem]">
            <span class="section-label mb-0">{{ t('collection.menu.face') }}</span>
            <div class="flex flex-wrap gap-[0.4rem]">
              <button class="chip" :class="{ active: gridSide === 'obverse' }" @click="$emit('update:gridSide', gridSide === 'obverse' ? null : 'obverse')">{{ t('common.obverse') }}</button>
              <button class="chip" :class="{ active: gridSide === 'reverse' }" @click="$emit('update:gridSide', gridSide === 'reverse' ? null : 'reverse')">{{ t('common.reverse') }}</button>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </div>
  <div v-if="menuOpen" class="fixed inset-0 z-[90]" @click="$emit('update:menuOpen', false)"></div>
</template>

<script setup lang="ts">
import type { CollectionSetOption, ImageType } from '@/types'
import type { CoinDenominationCount } from '@/api/client'
import DenominationFilter from '@/components/collection/DenominationFilter.vue'
import SearchBar from '@/components/SearchBar.vue'
import SortSelect from '@/components/SortSelect.vue'
import { Layers, LayoutGrid, SlidersHorizontal } from 'lucide-vue-next'
import { t } from '@/i18n'

defineProps<{
  search: string
  selectMode: boolean
  menuOpen: boolean
  selectedDenomination: string
  selectedTag: string
  userTags: CollectionSetOption[]
  denominationOptions: CoinDenominationCount[]
  sortKey: string
  viewMode: 'grid' | 'swipe'
  gridSide: ImageType | null
}>()

defineEmits<{
  'update:search': [value: string]
  'update:menuOpen': [value: boolean]
  'update:selectedDenomination': [value: string]
  'update:selectedTag': [value: string]
  'update:sortKey': [value: string]
  'update:viewMode': [value: 'grid' | 'swipe']
  'update:gridSide': [value: ImageType | null]
  'toggle-select-mode': []
}>()
</script>
