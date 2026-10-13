<template>
  <section class="rounded-md border border-border-subtle bg-card p-4" :aria-label="t('map.controls.title')">
    <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <label class="flex min-w-0 flex-col gap-1">
        <span class="section-label mb-0">{{ t('map.controls.colorBy') }}</span>
        <select v-model="metric" class="form-select h-[38px] bg-card text-body" data-testid="map-metric">
          <option value="coins">{{ t('map.controls.metricCoins') }}</option>
          <option value="value">{{ t('map.controls.metricValue') }}</option>
          <option value="types">{{ t('map.controls.metricTypes') }}</option>
        </select>
      </label>
      <label class="flex min-w-0 flex-col gap-1">
        <span class="section-label mb-0">{{ t('map.controls.target') }}</span>
        <select v-model="target" class="form-select h-[38px] bg-card text-body" data-testid="map-target">
          <option value="">{{ t('map.controls.targetNone') }}</option>
          <option value="euro">{{ t('map.controls.targetEuro') }}</option>
          <option value="europe">{{ t('map.controls.targetEurope') }}</option>
          <option value="world">{{ t('map.controls.targetWorld') }}</option>
        </select>
      </label>
      <label class="flex min-w-0 flex-col gap-1">
        <span class="section-label mb-0">{{ t('map.controls.circulation') }}</span>
        <select :value="filters.circulation" class="form-select h-[38px] bg-card text-body" data-testid="map-circulation" @change="setFilter('circulation', ($event.target as HTMLSelectElement).value)">
          <option value="">{{ t('map.controls.allCoins') }}</option>
          <option value="current">{{ t('map.controls.current') }}</option>
          <option value="demonetized">{{ t('map.controls.demonetized') }}</option>
          <option value="nonCirculating">{{ t('map.controls.nonCirculating') }}</option>
        </select>
      </label>
      <label class="flex min-w-0 flex-col gap-1">
        <span class="section-label mb-0">{{ t('map.controls.commemorative') }}</span>
        <select :value="filters.commemorative" class="form-select h-[38px] bg-card text-body" data-testid="map-commemorative" @change="setFilter('commemorative', ($event.target as HTMLSelectElement).value)">
          <option value="">{{ t('map.controls.allCoins') }}</option>
          <option value="true">{{ t('map.controls.commemorativeOnly') }}</option>
          <option value="false">{{ t('map.controls.regularOnly') }}</option>
        </select>
      </label>
      <label v-if="eras.length > 1 || filters.era" class="flex min-w-0 flex-col gap-1">
        <span class="section-label mb-0">{{ t('map.controls.era') }}</span>
        <select :value="filters.era" class="form-select h-[38px] bg-card text-body" data-testid="map-era" @change="setFilter('era', ($event.target as HTMLSelectElement).value)">
          <option value="">{{ t('map.controls.allEras') }}</option>
          <option v-for="era in eras" :key="era" :value="era">{{ eraLabel(era) }}</option>
        </select>
      </label>
      <label class="flex min-w-0 flex-col gap-1">
        <span class="section-label mb-0">{{ t('map.controls.yearFrom') }}</span>
        <input :value="filters.yearFrom ?? ''" type="number" inputmode="numeric" class="form-input h-[38px]" data-testid="map-year-from" @change="setYear('yearFrom', ($event.target as HTMLInputElement).value)" />
      </label>
      <label class="flex min-w-0 flex-col gap-1">
        <span class="section-label mb-0">{{ t('map.controls.yearTo') }}</span>
        <input :value="filters.yearTo ?? ''" type="number" inputmode="numeric" class="form-input h-[38px]" data-testid="map-year-to" @change="setYear('yearTo', ($event.target as HTMLInputElement).value)" />
      </label>
      <fieldset class="m-0 flex min-w-0 flex-col gap-1 border-0 p-0">
        <legend class="section-label mb-1">{{ t('map.controls.layers') }}</legend>
        <label class="flex items-center gap-2 text-body text-text-primary">
          <input v-model="showMints" type="checkbox" class="h-4 w-4 accent-[var(--accent-gold)]" data-testid="map-layer-mints" />
          {{ t('map.controls.mints') }}
        </label>
        <label class="flex items-center gap-2 text-body text-text-primary">
          <input v-model="showExtinct" type="checkbox" class="h-4 w-4 accent-[var(--accent-gold)]" data-testid="map-layer-extinct" />
          {{ t('map.controls.extinct') }}
        </label>
      </fieldset>
    </div>
    <div v-if="hasMapFilters(filters)" class="mt-3 flex flex-wrap items-center gap-3 text-chip text-text-secondary" role="status">
      <span>{{ t('map.controls.filtered', { shown: shownCount, total: totalCount }) }}</span>
      <button type="button" class="btn btn-secondary btn-sm" @click="filters = { ...EMPTY_MAP_FILTERS }">{{ t('map.controls.clearFilters') }}</button>
    </div>
  </section>
</template>

<script setup lang="ts">
// The map's options: what the colours measure, which countries a target
// still misses, which coins count (legal tender, commemorative, period,
// years) and the extra layers (mints, extinct countries).
import { t } from '@/i18n'
import { eraLabel } from '@/i18n/coinVocabulary'
import { EMPTY_MAP_FILTERS, hasMapFilters, type MapFilters, type MapMetric, type MapTarget } from '@/utils/collectionMap'

defineProps<{
  /** The periods of the collection's coins. */
  eras: string[]
  shownCount: number
  totalCount: number
}>()

const metric = defineModel<MapMetric>('metric', { required: true })
const target = defineModel<MapTarget>('target', { required: true })
const filters = defineModel<MapFilters>('filters', { required: true })
const showMints = defineModel<boolean>('showMints', { required: true })
const showExtinct = defineModel<boolean>('showExtinct', { required: true })

function setFilter<K extends 'era' | 'circulation' | 'commemorative'>(key: K, value: string) {
  filters.value = { ...filters.value, [key]: value }
}

function setYear(key: 'yearFrom' | 'yearTo', value: string) {
  const year = value.trim() === '' ? null : Number.parseInt(value, 10)
  filters.value = { ...filters.value, [key]: year === null || Number.isNaN(year) ? null : year }
}
</script>
