<template>
  <!-- Every filter the same size, as in the coin gallery; one under another in a phone's sheet. -->
  <div
    :class="stacked ? 'flex flex-col gap-4' : ['grid grid-cols-1 gap-3 sm:grid-cols-2', showClassification ? 'lg:grid-cols-5' : 'lg:grid-cols-4']"
    data-testid="banknotes-filters"
  >
    <label class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('collection.menu.sort') }}</span>
      <select :value="sortKey" class="form-select h-10 w-full bg-card py-0 text-body" data-testid="banknotes-sort" @change="$emit('update:sortKey', ($event.target as HTMLSelectElement).value)">
        <option value="updated_at_desc">{{ t('collection.sort.lastUpdated') }}</option>
        <option value="created_at_desc">{{ t('collection.sort.newest') }}</option>
        <option value="created_at_asc">{{ t('collection.sort.oldest') }}</option>
        <option value="current_value_desc">{{ t('collection.sort.priceHighLow') }}</option>
        <option value="current_value_asc">{{ t('collection.sort.priceLowHigh') }}</option>
      </select>
    </label>
    <label class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('collection.view.groupBy') }}</span>
      <select :value="groupBy" class="form-select h-10 w-full bg-card py-0 text-body" data-testid="banknotes-group-by" @change="$emit('update:groupBy', ($event.target as HTMLSelectElement).value as BanknoteGrouping)">
        <option v-for="option in BANKNOTE_GROUPINGS" :key="option" :value="option">{{ t(GROUPING_LABELS[option]) }}</option>
      </select>
    </label>
    <div class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('coinFields.country') }}</span>
      <CountrySelect :model-value="country" :options="countryChoices" :label="t('coinFields.country')" data-testid="banknotes-country" @update:model-value="$emit('update:country', $event)" />
    </div>
    <!-- Numista's kinds of banknote in the collection, when it says any. -->
    <label v-if="showClassification" class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('banknotes.classification') }}</span>
      <select :value="classification" class="form-select h-10 w-full bg-card py-0 text-body" data-testid="banknotes-classification" @change="$emit('update:classification', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ t('collection.view.allClassifications') }}</option>
        <option v-for="option in classifications" :key="option.key" :value="option.key">{{ collectionClassificationLabel(option) }} ({{ option.count }})</option>
      </select>
    </label>
    <div class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('collection.view.layout') }}</span>
      <div class="flex h-10 w-full rounded-sm border border-border-subtle bg-input p-[2px]" role="group" :aria-label="t('collection.view.layout')">
        <button
          v-for="option in LAYOUTS"
          :key="option.id"
          type="button"
          class="inline-flex min-w-0 flex-1 items-center justify-center gap-1 rounded-[6px] px-2 text-chip transition-all"
          :class="layout === option.id ? 'bg-gold text-surface font-semibold' : 'bg-transparent font-medium text-text-secondary hover:bg-card-hover hover:text-text-primary'"
          :aria-pressed="layout === option.id"
          @click="$emit('update:layout', option.id)"
        >
          <component :is="option.icon" v-if="stacked" :size="16" aria-hidden="true" />
          {{ t(option.label) }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// The banknote gallery's order, grouping, filters and layout: in a row above
// the banknotes on a wider screen, in the filters sheet on a phone.
import { computed } from 'vue'
import { LayoutGrid, List } from 'lucide-vue-next'
import type { CollectionClassification } from '@/api/client'
import CountrySelect, { type CountrySelectOption } from '@/components/CountrySelect.vue'
import { t, type MessageKey } from '@/i18n'
import { BANKNOTE_GROUPINGS, collectionClassificationLabel, type BanknoteGrouping } from '@/utils/coinAttributes'

export type BanknoteLayout = 'grid' | 'list'

const props = defineProps<{
  sortKey: string
  groupBy: BanknoteGrouping
  country: string
  classification: string
  layout: BanknoteLayout
  countryChoices: CountrySelectOption[]
  classifications: CollectionClassification[]
  /** One under another, as in a phone's sheet. */
  stacked?: boolean
}>()

defineEmits<{
  'update:sortKey': [value: string]
  'update:groupBy': [value: BanknoteGrouping]
  'update:country': [value: string]
  'update:classification': [value: string]
  'update:layout': [value: BanknoteLayout]
}>()

const GROUPING_LABELS: Record<BanknoteGrouping, MessageKey> = {
  none: 'banknotes.groupNone',
  country: 'coinFields.country',
  monetarySystem: 'coinFields.monetarySystem',
  year: 'collection.view.groupYear',
  rarity: 'coinFields.rarity',
}
const LAYOUTS: { id: BanknoteLayout, label: MessageKey, icon: typeof LayoutGrid }[] = [
  { id: 'grid', label: 'collection.view.grid', icon: LayoutGrid },
  { id: 'list', label: 'collection.view.list', icon: List },
]

const showClassification = computed(() => props.classifications.length > 0 || props.classification !== '')
</script>
