<template>
  <!-- Its fields join the header's grid of equal columns; stacked, one per line. -->
  <div :class="stacked ? 'flex flex-col gap-3' : 'contents'">
    <label class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('collection.view.groupBy') }}</span>
      <select class="form-select h-10 w-full bg-card py-0 text-body" :value="groupBy" @change="$emit('update:groupBy', ($event.target as HTMLSelectElement).value as CoinGroupBy)">
        <option v-for="grouping in COIN_GROUPINGS" :key="grouping" :value="grouping">{{ groupingLabel(grouping) }}</option>
      </select>
    </label>

    <div class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('coinFields.country') }}</span>
      <!-- Its own menu, so every country shows its flag (drawn ones too). -->
      <CountrySelect :model-value="country" :options="countryChoices" :label="t('coinFields.country')" data-testid="country-filter" @update:model-value="$emit('update:country', $event)" />
    </div>

    <!-- The chosen country's periods, the most recent first. -->
    <label v-if="periodOptions.length || period" class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('coinFields.period') }}</span>
      <select class="form-select h-10 w-full bg-card py-0 text-body" :value="period" data-testid="period-filter" @change="$emit('update:period', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ t('collection.view.allPeriods') }}</option>
        <option v-for="option in periodOptions" :key="option.period" :value="option.period">{{ periodLabel(option) }} ({{ option.count }})</option>
      </select>
    </label>

    <label v-if="classifications.length || classification" class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('collection.view.classification') }}</span>
      <select class="form-select h-10 w-full bg-card py-0 text-body" :value="classification" data-testid="classification-filter" @change="$emit('update:classification', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ t('collection.view.allClassifications') }}</option>
        <option v-for="option in classifications" :key="option.key" :value="option.key">{{ classificationLabel(option) }} ({{ option.count }})</option>
      </select>
    </label>

    <label class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('coinFields.presentation') }}</span>
      <select class="form-select h-10 w-full bg-card py-0 text-body" :value="presentation" @change="$emit('update:presentation', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ t('collection.view.allPresentations') }}</option>
        <option v-for="option in COIN_PRESENTATIONS" :key="option" :value="option">{{ presentationLabel(option) }}</option>
      </select>
    </label>

    <label class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('coinFields.rarity') }}</span>
      <select class="form-select h-10 w-full bg-card py-0 text-body" :value="rarity" @change="$emit('update:rarity', ($event.target as HTMLSelectElement).value)">
        <option value="">{{ t('collection.view.allRarities') }}</option>
        <option v-for="option in COIN_RARITIES" :key="option" :value="option">{{ rarityLabel(option) }}</option>
      </select>
    </label>

    <div v-if="showLayout" class="flex min-w-0 flex-col gap-1">
      <span class="section-label mb-0">{{ t('collection.view.layout') }}</span>
      <div class="flex h-10 w-full rounded-sm border border-border-subtle bg-input p-[2px]" role="group" :aria-label="t('collection.view.layout')">
        <button
          v-for="option in LAYOUTS"
          :key="option.value"
          type="button"
          class="inline-flex min-w-0 flex-1 items-center justify-center gap-1 rounded-[6px] px-2 text-chip transition-all"
          :class="layout === option.value ? 'bg-gold font-semibold text-surface' : 'font-medium text-text-secondary hover:bg-card-hover hover:text-text-primary'"
          :aria-pressed="layout === option.value"
          @click="$emit('update:layout', option.value)"
        >
          <component :is="option.icon" :size="16" />
          {{ t(option.label) }}
        </button>
      </div>
    </div>

    <p v-if="fillingCountries" class="col-span-full m-0 text-chip text-text-muted" role="status">{{ t('collection.view.fillingCountries') }}</p>
  </div>
</template>

<script setup lang="ts">
// "View by" grouping, grid/list layout and the country, presentation and
// rarity filters of the collection, shared by the desktop and PWA headers.
import { LayoutGrid, List } from 'lucide-vue-next'
import { computed } from 'vue'
import { EXTINCT_COUNTRIES } from '@/utils/countryFilter'
import CountrySelect, { type CountrySelectOption } from '@/components/CountrySelect.vue'
import type { CoinCountryCount, CoinPeriodCount, CollectionClassification } from '@/api/client'
import { periodYears } from '@/utils/coinPeriods'
import { formatNumistaYear } from '@/utils/numistaCatalog'
import { t, type MessageKey } from '@/i18n'
import {
  COIN_GROUPINGS,
  COIN_PRESENTATIONS,
  COIN_RARITIES,
  collectionClassificationLabel,
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
  period: 'coinFields.period',
  monetarySystem: 'coinFields.monetarySystem',
  year: 'collection.view.groupYear',
  rarity: 'coinFields.rarity',
  presentation: 'coinFields.presentation',
}

const props = withDefaults(defineProps<{
  groupBy: CoinGroupBy
  layout: CollectionLayout
  country: string
  /** '' or one of the chosen country's periods. */
  period?: string
  presentation: string
  rarity: string
  countryOptions: CoinCountryCount[]
  /** The chosen country's periods; none until a country is chosen. */
  periodOptions?: CoinPeriodCount[]
  /** Numista's kinds of coin in the collection, offered as a filter. */
  classifications?: CollectionClassification[]
  /** '' or the chosen kind's key. */
  classification?: string
  showLayout?: boolean
  /** One control per line, for the PWA menu. */
  stacked?: boolean
  fillingCountries?: boolean
  /** How many of the countries no longer exist; offered as one choice. */
  extinctCountries?: number
}>(), { showLayout: true, stacked: false, fillingCountries: false, extinctCountries: 0, classifications: () => [], classification: '', period: '', periodOptions: () => [] })

defineEmits<{
  'update:groupBy': [value: CoinGroupBy]
  'update:layout': [value: CollectionLayout]
  'update:country': [value: string]
  'update:period': [value: string]
  'update:classification': [value: string]
  'update:presentation': [value: string]
  'update:rarity': [value: string]
}>()

// A period whose name gives no years shows those of its coins.
function periodLabel(option: CoinPeriodCount): string {
  if (periodYears(option.period) || option.fromYear === undefined) return option.period
  const to = option.toYear ?? option.fromYear
  const years = option.fromYear === to ? formatNumistaYear(option.fromYear) : `${formatNumistaYear(option.fromYear)}–${formatNumistaYear(to)}`
  return `${option.period} ${years}`
}

// "All countries", "Extinct countries" when there are any, then each country with its count.
const countryChoices = computed((): CountrySelectOption[] => [
  { value: '', label: t('collection.view.allCountries') },
  ...(props.extinctCountries ? [{ value: EXTINCT_COUNTRIES, label: t('collection.view.extinctCountries', { count: props.extinctCountries }) }] : []),
  ...props.countryOptions.map((option) => ({ value: option.country, label: `${option.country} (${option.count})`, country: option.country })),
])

function groupingLabel(grouping: CoinGroupBy): string {
  return t(GROUPING_KEYS[grouping])
}

function classificationLabel(option: CollectionClassification): string {
  return collectionClassificationLabel(option)
}
</script>
