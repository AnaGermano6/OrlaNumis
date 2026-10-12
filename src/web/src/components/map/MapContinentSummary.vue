<template>
  <section class="rounded-md border border-border-subtle bg-card p-4" aria-labelledby="map-continents-title">
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
      <h2 id="map-continents-title" class="m-0 font-display text-lg font-medium text-heading">{{ t('map.continents.title') }}</h2>
      <button v-if="selected" type="button" class="btn btn-secondary btn-sm" @click="emit('select', '')">{{ t('map.continents.all') }}</button>
    </div>
    <ul class="m-0 grid list-none gap-2 p-0 sm:grid-cols-2 xl:grid-cols-4" data-testid="map-continents">
      <li v-for="row in totals" :key="row.continent">
        <button
          type="button"
          class="flex w-full flex-col gap-0.5 rounded-sm border px-3 py-2 text-left transition-colors hover:border-gold"
          :class="row.continent === selected ? 'border-gold bg-gold-glow' : 'border-border-subtle'"
          :aria-pressed="row.continent === selected"
          @click="emit('select', row.continent === selected ? '' : row.continent)"
        >
          <span class="font-medium text-text-primary">{{ t(CONTINENT_KEYS[row.continent]) }}</span>
          <span class="text-chip text-text-secondary">
            {{ t('map.country.countries', { count: row.places }) }} · {{ t('map.country.coins', { count: row.coins }) }}<template v-if="row.value > 0"> · {{ formatCurrency(row.value) }}</template>
          </span>
        </button>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
// Places, coins and value by continent; choosing a continent narrows the
// list of countries to it and zooms the map there.
import { t, type MessageKey } from '@/i18n'
import type { ContinentTotals } from '@/utils/collectionMap'
import type { Continent } from '@/utils/mapGeography'
import { formatCurrency } from '@/utils/format'

defineProps<{
  totals: ContinentTotals[]
  /** The chosen continent, or ''. */
  selected: Continent | 'other' | ''
}>()

const emit = defineEmits<{ select: [continent: Continent | 'other' | ''] }>()

const CONTINENT_KEYS: Record<Continent | 'other', MessageKey> = {
  europe: 'map.continents.europe',
  asia: 'map.continents.asia',
  africa: 'map.continents.africa',
  northAmerica: 'map.continents.northAmerica',
  southAmerica: 'map.continents.southAmerica',
  oceania: 'map.continents.oceania',
  antarctica: 'map.continents.antarctica',
  other: 'map.continents.other',
}
</script>
