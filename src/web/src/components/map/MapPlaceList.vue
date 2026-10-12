<template>
  <section class="rounded-md border border-border-subtle bg-card p-4" aria-labelledby="map-places-title">
    <h2 id="map-places-title" class="mb-3 mt-0 font-display text-lg font-medium text-heading">{{ t('map.country.listTitle') }}</h2>
    <input v-model="search" type="search" class="form-input mb-3 w-full" :placeholder="t('map.country.search')" :aria-label="t('map.country.search')" />
    <ul class="m-0 flex max-h-[520px] list-none flex-col gap-1 overflow-y-auto p-0" data-testid="map-country-list">
      <li v-for="place in shown" :key="place.key">
        <button
          type="button"
          class="flex w-full flex-col gap-1 rounded-sm border px-3 py-2 text-left transition-colors hover:border-gold"
          :class="place.key === selectedKey ? 'border-gold bg-gold-glow' : 'border-transparent'"
          :aria-pressed="place.key === selectedKey"
          @click="emit('select', place.key === selectedKey ? null : place.key)"
        >
          <span class="flex items-center justify-between gap-2 text-body">
            <span class="truncate text-text-primary">{{ place.label }}<span v-if="place.kind === 'offMap'" class="text-text-muted" :title="t('map.country.offMapMark')"> *</span></span>
            <span class="whitespace-nowrap text-text-secondary">{{ measure(place) }}</span>
          </span>
          <span class="block h-1 rounded-full bg-gold" :style="{ width: `${Math.max(4, (metricValue(place, metric) / max) * 100)}%`, opacity: 0.7 }" aria-hidden="true"></span>
        </button>
      </li>
    </ul>
    <p v-if="!shown.length" class="mb-0 mt-2 text-body text-text-muted">{{ places.length ? t('map.country.noMatch') : t('map.places.empty') }}</p>
    <p v-if="places.some((place) => place.kind === 'offMap')" class="mb-0 mt-2 text-2xs text-text-muted">* {{ t('collection.worldMap.notOnMap') }}</p>
  </section>
</template>

<script setup lang="ts">
// The places of the collection, by the chosen measure, with a search; a
// place is chosen from here as from the map.
import { computed, ref } from 'vue'
import { t } from '@/i18n'
import { metricValue, type MapMetric, type PlaceStats } from '@/utils/collectionMap'
import { formatCurrency } from '@/utils/format'
import { foldCountryName } from '@/utils/worldCountries'

const props = defineProps<{
  /** In order. */
  places: PlaceStats[]
  metric: MapMetric
  selectedKey: string | null
}>()

const emit = defineEmits<{ select: [key: string | null] }>()

const search = ref('')

const shown = computed(() => {
  const wanted = foldCountryName(search.value)
  if (!wanted) return props.places
  return props.places.filter((place) => [place.label, ...place.names].some((name) => foldCountryName(name).includes(wanted)))
})

const max = computed(() => Math.max(1, ...props.places.map((place) => metricValue(place, props.metric))))

function measure(place: PlaceStats): string {
  if (props.metric === 'value') return formatCurrency(place.value)
  return String(props.metric === 'types' ? place.types : place.count)
}
</script>
