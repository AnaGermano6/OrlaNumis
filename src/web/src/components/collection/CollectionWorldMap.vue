<template>
  <section v-if="countries.length" class="mb-4 rounded-md border border-border-subtle bg-card p-4" aria-labelledby="world-map-title">
    <div class="mb-1 flex flex-wrap items-center justify-between gap-2">
      <h2 id="world-map-title" class="m-0 font-display text-lg font-medium text-heading">{{ t('collection.worldMap.title') }}</h2>
      <button v-if="selectedCountry" type="button" class="btn btn-secondary btn-sm" @click="$emit('select', '')">{{ t('collection.worldMap.showAll') }}</button>
    </div>
    <p class="mb-3 mt-0 text-chip text-text-muted" role="status">
      {{ selectedCountry ? t('collection.worldMap.showing', { country: selectedCountry }) : t('collection.worldMap.hint') }}
    </p>

    <div
      ref="mapElement"
      class="world-map h-[240px] w-full overflow-hidden rounded-sm border border-border-subtle bg-input md:h-[340px]"
      role="img"
      :aria-label="t('collection.worldMap.mapLabel')"
    ></div>
    <p v-if="loadFailed" class="mb-0 mt-2 text-body text-text-secondary" role="status">{{ t('collection.worldMap.loadFailed') }}</p>

    <div class="mt-2 flex flex-wrap items-center gap-3 text-chip text-text-muted" :aria-label="t('collection.worldMap.legend')">
      <span>{{ t('collection.worldMap.legend') }}</span>
      <span v-for="step in LEGEND" :key="step.shade" class="inline-flex items-center gap-1">
        <span class="inline-block h-3 w-4 rounded-[2px] border border-border-subtle" :style="{ backgroundColor: shadeColor(step.shade), opacity: shadeOpacity(step.shade) }"></span>
        {{ step.label }}
      </span>
    </div>

    <ul class="m-0 mt-3 flex list-none flex-wrap gap-2 p-0" :aria-label="t('collection.worldMap.countries')">
      <li v-for="entry in rankedCountries" :key="entry.country">
        <button
          type="button"
          class="chip border border-border-subtle"
          :class="{ 'border-gold': entry.country === selectedCountry }"
          :aria-pressed="entry.country === selectedCountry"
          :title="entry.onMap ? undefined : t('collection.worldMap.notOnMap')"
          @click="$emit('select', entry.country === selectedCountry ? '' : entry.country)"
        >
          {{ entry.country }} · {{ entry.count }}<span v-if="!entry.onMap" aria-hidden="true"> *</span>
        </button>
      </li>
    </ul>
    <p v-if="distribution.unmatched.length" class="mb-0 mt-2 text-2xs text-text-muted">* {{ t('collection.worldMap.notOnMap') }}</p>
    <p class="mb-0 mt-1 text-2xs text-text-muted">{{ t('collection.worldMap.attribution') }}</p>
  </section>
</template>

<script setup lang="ts">
// World map of the collection: each country is shaded by its number of coins;
// choosing a country (on the map or in the list) filters the gallery to it.
import 'leaflet/dist/leaflet.css'
import * as L from 'leaflet'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import type { CoinCountryCount } from '@/api/client'
import { intlLocale, locale, t } from '@/i18n'
import {
  buildCountryIndex,
  countryShade,
  distributeCountries,
  loadWorldCountries,
  type CountryDistribution,
  type WorldCountryCollection,
  type WorldCountryFeature,
} from '@/utils/worldCountries'

const props = defineProps<{
  countries: CoinCountryCount[]
  /** The gallery's country filter (a stored country name), '' for all. */
  selectedCountry: string
}>()

const emit = defineEmits<{
  /** A country name to filter by, or '' to show every country. */
  select: [country: string]
}>()

const LEGEND = [
  { shade: 1, label: '1' },
  { shade: 2, label: '2–4' },
  { shade: 3, label: '5–9' },
  { shade: 4, label: '10+' },
] as const

const mapElement = ref<HTMLElement | null>(null)
const loadFailed = ref(false)
const collection = ref<WorldCountryCollection | null>(null)
let map: L.Map | null = null
let countryLayer: L.GeoJSON | null = null

const index = computed(() => (collection.value ? buildCountryIndex(collection.value) : new Map<string, string>()))
const distribution = computed<CountryDistribution>(() => distributeCountries(props.countries, index.value))

const rankedCountries = computed(() => {
  const unmatched = new Set(distribution.value.unmatched.map((entry) => entry.country))
  return [...props.countries]
    .sort((a, b) => b.count - a.count || a.country.localeCompare(b.country, intlLocale()))
    .map((entry) => ({ ...entry, onMap: !collection.value || !unmatched.has(entry.country) }))
})

// Leaflet writes SVG attributes, which cannot use var(); read the theme's values.
function cssColor(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || 'currentColor'
}

function shadeColor(shade: number): string {
  return shade === 0 ? cssColor('--border-subtle') : cssColor('--accent-gold')
}

function shadeOpacity(shade: number): number {
  // Countries without coins stay faint; a single coin already stands out.
  return [0.35, 0.5, 0.65, 0.8, 0.95][shade] ?? 0.35
}

function countryLabel(feature: WorldCountryFeature): string {
  return locale.value === 'pt-PT' ? feature.properties.namePt : feature.properties.name
}

function featureStyle(feature?: WorldCountryFeature): L.PathOptions {
  const country = feature ? distribution.value.byIso.get(feature.properties.iso) : undefined
  const shade = countryShade(country?.count ?? 0)
  const selected = Boolean(country && country.names.some((entry) => entry.country === props.selectedCountry))
  return {
    color: selected ? cssColor('--accent-gold') : cssColor('--bg-card'),
    weight: selected ? 2.5 : 0.6,
    fillColor: shadeColor(shade),
    fillOpacity: shadeOpacity(shade),
  }
}

function bindCountry(feature: WorldCountryFeature, layer: L.Layer) {
  const country = distribution.value.byIso.get(feature.properties.iso)
  const count = country?.count ?? 0
  layer.bindTooltip(t('collection.worldMap.countryCoins', { country: countryLabel(feature), count }), { sticky: true })
  if (country) {
    layer.on('click', () => {
      const name = country.names[0]!.country
      emit('select', name === props.selectedCountry ? '' : name)
    })
  }
}

function drawCountries() {
  if (!map || !collection.value) return
  countryLayer?.remove()
  countryLayer = L.geoJSON(collection.value, {
    style: (feature) => featureStyle(feature as WorldCountryFeature),
    onEachFeature: (feature, layer) => bindCountry(feature as WorldCountryFeature, layer),
  }).addTo(map)
}

async function initMap() {
  if (map || !mapElement.value) return
  try {
    collection.value = await loadWorldCountries()
  } catch {
    loadFailed.value = true
    return
  }
  await nextTick()
  if (!mapElement.value || map) return
  map = L.map(mapElement.value, {
    attributionControl: false,
    zoomSnap: 0.25,
    minZoom: 0.75,
    maxZoom: 6,
    scrollWheelZoom: false,
    worldCopyJump: false,
    maxBounds: [[-75, -200], [88, 200]],
  })
  map.fitBounds([[-56, -170], [78, 180]])
  drawCountries()
}

// The section only renders once the collection has countries.
watch(mapElement, (element) => {
  if (element) void initMap()
}, { immediate: true })

// Counts, the selected country or the language changed: restyle and relabel.
watch([distribution, () => props.selectedCountry, locale], drawCountries)

onBeforeUnmount(() => {
  map?.remove()
  map = null
  countryLayer = null
})
</script>

<style scoped>
/* Leaflet adds .leaflet-container (grey background) to the element itself. */
.world-map.leaflet-container {
  background: var(--bg-input);
}
</style>
