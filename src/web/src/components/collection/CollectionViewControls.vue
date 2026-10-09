<template>
  <div class="flex min-w-0 flex-wrap items-end gap-3" :class="{ 'flex-col items-stretch': stacked }">
    <label class="flex min-w-[150px] flex-col gap-1" :class="{ 'min-w-0': stacked }">
      <span class="section-label mb-0">{{ t('collection.view.groupBy') }}</span>
      <select class="form-select h-[38px] bg-card text-body" :value="groupBy" @change="$emit('update:groupBy', ($event.target as HTMLSelectElement).value as CoinGroupBy)">
        <option v-for="grouping in COIN_GROUPINGS" :key="grouping" :value="grouping">{{ groupingLabel(grouping) }}</option>
      </select>
    </label>

    <label class="flex min-w-[150px] flex-col gap-1" :class="{ 'min-w-0': stacked }">
      <span class="section-label mb-0">{{ t('coinFields.country') }}</span>
      <select class="form-select h-[38px] bg-card text-body" :value="country" @change="$emit('update:country', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ t('collection.view.allCountries') }}</option>
        <option v-for="option in countryOptions" :key="option.country" :value="option.country">{{ option.country }} ({{ option.count }})</option>
      </select>
    </label>

    <label class="flex min-w-[150px] flex-col gap-1" :class="{ 'min-w-0': stacked }">
      <span class="section-label mb-0">{{ t('coinFields.presentation') }}</span>
      <select class="form-select h-[38px] bg-card text-body" :value="presentation" @change="$emit('update:presentation', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ t('collection.view.allPresentations') }}</option>
        <option v-for="option in COIN_PRESENTATIONS" :key="option" :value="option">{{ presentationLabel(option) }}</option>
      </select>
    </label>

    <label class="flex min-w-[150px] flex-col gap-1" :class="{ 'min-w-0': stacked }">
      <span class="section-label mb-0">{{ t('coinFields.rarity') }}</span>
      <select class="form-select h-[38px] bg-card text-body" :value="rarity" @change="$emit('update:rarity', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ t('collection.view.allRarities') }}</option>
        <option v-for="option in COIN_RARITIES" :key="option" :value="option">{{ rarityLabel(option) }}</option>
      </select>
    </label>

    <div v-if="showLayout" class="flex flex-col gap-1">
      <span class="section-label mb-0">{{ t('collection.view.layout') }}</span>
      <div class="inline-flex h-[38px] w-fit rounded-sm border border-border-subtle bg-input p-[2px]" role="group" :aria-label="t('collection.view.layout')">
        <button
          v-for="option in LAYOUTS"
          :key="option.value"
          type="button"
          class="inline-flex items-center gap-1 rounded-[6px] px-3 text-chip transition-all"
          :class="layout === option.value ? 'bg-gold font-semibold text-surface' : 'font-medium text-text-secondary hover:bg-card-hover hover:text-text-primary'"
          :aria-pressed="layout === option.value"
          @click="$emit('update:layout', option.value)"
        >
          <component :is="option.icon" :size="16" />
          {{ t(option.label) }}
        </button>
      </div>
    </div>

    <p v-if="fillingCountries" class="m-0 basis-full text-chip text-text-muted" role="status">{{ t('collection.view.fillingCountries') }}</p>
  </div>
</template>

<script setup lang="ts">
// "View by" grouping, grid/list layout and the country, presentation and
// rarity filters of the collection, shared by the desktop and PWA headers.
import { LayoutGrid, List } from 'lucide-vue-next'
import type { CoinCountryCount } from '@/api/client'
import { t, type MessageKey } from '@/i18n'
import {
  COIN_GROUPINGS,
  COIN_PRESENTATIONS,
  COIN_RARITIES,
  presentationLabel,
  rarityLabel,
  type CoinGroupBy,
} from '@/utils/coinAttributes'

export type CollectionLayout = 'grid' | 'list'

const LAYOUTS: { value: CollectionLayout; label: MessageKey; icon: typeof LayoutGrid }[] = [
  { value: 'grid', label: 'collection.view.grid', icon: LayoutGrid },
  { value: 'list', label: 'collection.view.list', icon: List },
]

const GROUPING_KEYS: Record<CoinGroupBy, MessageKey> = {
  none: 'collection.view.groupNone',
  country: 'coinFields.country',
  year: 'collection.view.groupYear',
  rarity: 'coinFields.rarity',
  presentation: 'coinFields.presentation',
}

withDefaults(defineProps<{
  groupBy: CoinGroupBy
  layout: CollectionLayout
  country: string
  presentation: string
  rarity: string
  countryOptions: CoinCountryCount[]
  showLayout?: boolean
  /** One control per line, for the PWA menu. */
  stacked?: boolean
  fillingCountries?: boolean
}>(), { showLayout: true, stacked: false, fillingCountries: false })

defineEmits<{
  'update:groupBy': [value: CoinGroupBy]
  'update:layout': [value: CollectionLayout]
  'update:country': [value: string]
  'update:presentation': [value: string]
  'update:rarity': [value: string]
}>()

function groupingLabel(grouping: CoinGroupBy): string {
  return t(GROUPING_KEYS[grouping])
}
</script>
