<template>
  <div class="container flex flex-col gap-4 pb-10">
    <DesktopPageContext :title="t('map.title')">
      <template #actions>
        <AppIconButton to="/stats" :title="t('stats.backToStats')"><ArrowLeft :size="20" /></AppIconButton>
      </template>
      <header class="page-header flex-row items-center">
        <div>
          <p class="section-label">{{ t('stats.insights') }}</p>
          <h1>{{ t('map.title') }}</h1>
        </div>
        <router-link class="btn btn-secondary btn-sm" to="/stats" :aria-label="t('stats.backToStats')"><ArrowLeft :size="20" /></router-link>
      </header>
    </DesktopPageContext>

    <div v-if="loading" class="card flex items-center gap-3 p-6 text-text-secondary" role="status">
      <div class="spinner"></div>
      <span>{{ t('map.country.loading') }}</span>
    </div>

    <div v-else-if="failed" class="card flex flex-col items-start gap-3 p-6" role="alert">
      <p class="m-0 text-text-secondary">{{ t('map.country.failed') }}</p>
      <button class="btn btn-primary" type="button" @click="load">{{ t('common.tryAgain') }}</button>
    </div>

    <div v-else-if="!coins.some((coin) => coin.country?.trim())" class="card flex flex-col items-start gap-3 p-6">
      <p class="m-0 text-text-secondary">{{ t('map.country.empty') }}</p>
      <router-link class="btn btn-primary" to="/add">{{ t('nav.addCoin') }}</router-link>
    </div>

    <template v-else>
      <section class="flex flex-wrap gap-2" :aria-label="t('map.country.summary')" data-testid="map-summary">
        <span class="chip border border-border-subtle">{{ t('map.country.countries', { count: grouped.byKey.size }) }}</span>
        <span class="chip border border-border-subtle">{{ t('map.country.coins', { count: filteredCoins.length }) }}</span>
        <span v-if="offMapCount" class="chip border border-border-subtle" :title="t('collection.worldMap.notOnMap')">{{ t('map.country.offMap', { count: offMapCount }) }}</span>
        <span v-if="grouped.withoutCountry" class="chip border border-border-subtle">{{ t('map.country.withoutCountry', { count: grouped.withoutCountry }) }}</span>
      </section>

      <MapControlsPanel
        v-model:metric="metric"
        v-model:target="target"
        v-model:filters="filters"
        v-model:show-mints="showMints"
        v-model:show-extinct="showExtinct"
        :eras="eras"
        :shown-count="filteredCoins.length"
        :total-count="coins.length"
      />

      <section class="rounded-md border border-border-subtle bg-card p-4">
        <CountryCoinsMap
          v-if="world"
          ref="mapRef"
          :collection="world"
          :points="SMALL_TERRITORIES"
          :shades="shades"
          :describe="describe"
          :selectable="selectable"
          :selected-key="selectedKey"
          :missing="missing"
          :overlays="overlays"
          :pins="pins"
          :selected-pin="selectedPin"
          :focus="focus"
          @select="selectPlace"
          @select-pin="selectPin"
        />
        <p v-else class="m-0 text-body text-text-secondary" role="status">{{ t('collection.worldMap.loadFailed') }}</p>
        <div class="mt-3 flex flex-wrap items-center justify-between gap-3 text-chip text-text-muted">
          <div class="flex flex-wrap items-center gap-3" data-testid="map-legend">
            <span>{{ legendTitle }}</span>
            <span v-for="(label, index) in legendLabels" :key="label" class="inline-flex items-center gap-1">
              <span class="inline-block h-3 w-4 rounded-[2px] border border-border-subtle bg-gold" :style="{ opacity: countryShadeOpacity(index + 1) }"></span>
              {{ label }}
            </span>
            <span v-if="target" class="inline-flex items-center gap-1">
              <span class="inline-block h-3 w-4 rounded-[2px] border border-dashed border-[var(--text-warning)]"></span>
              {{ t('map.legend.missing') }}
            </span>
            <span v-if="showExtinct" class="inline-flex items-center gap-1">
              <span class="extinct-swatch inline-block h-3 w-4 rounded-[2px] border border-dashed border-bronze"></span>
              {{ t('map.legend.extinct') }}
            </span>
            <span v-if="showMints" class="inline-flex items-center gap-1">
              <span class="inline-block h-3 w-3 rounded-full border-2 border-gold bg-card"></span>
              {{ t('map.legend.mint') }}
            </span>
          </div>
          <button v-if="world" type="button" class="btn btn-secondary btn-sm" @click="resetView">{{ t('map.country.resetView') }}</button>
        </div>
        <p v-if="showMints && mintGroups.unplaced" class="mb-0 mt-1 text-2xs text-text-muted">{{ t('map.mints.unplaced', { count: mintGroups.unplaced }) }}</p>
        <p class="mb-0 mt-1 text-2xs text-text-muted">{{ t('collection.worldMap.attribution') }}</p>
      </section>

      <MapTargetProgress
        v-if="target"
        :title="targetTitle"
        :progress="progress"
        :label-for="labelFor"
        :selected-key="selectedKey"
        @select="selectPlace"
      />

      <MapContinentSummary :totals="continents" :selected="continent" @select="selectContinent" />

      <div class="grid gap-4 md:grid-cols-[minmax(240px,320px)_1fr]">
        <MapPlaceList :places="listedPlaces" :metric="metric" :selected-key="selectedKey" @select="selectPlace" />
        <MapSelectionPanel
          :title="panel.title"
          :kicker="panel.kicker"
          :details="panel.details"
          :note="panel.note"
          :coins="panel.coins"
          :empty-text="panel.emptyText"
          :gallery-country="panel.galleryCountry"
          @clear="clearSelection"
          @show-in-gallery="showInGallery"
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
// The collection on a world map. Countries are coloured by number of coins,
// value or types; filters narrow the coins (legal tender, commemorative,
// period, years); a target (the euro countries, Europe, the world) shows the
// countries still missing; extinct countries can be hatched over today's
// map and mints shown as pins; totals by continent zoom the map. Choosing a
// country, a continent's country or a mint lists its coins.
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft } from 'lucide-vue-next'
import { getCoinStatusGroups, getCoins, getMintLocations } from '@/api/client'
import type { Coin } from '@/types'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import CountryCoinsMap, { type MapOverlay, type MapPin } from '@/components/map/CountryCoinsMap.vue'
import MapContinentSummary from '@/components/map/MapContinentSummary.vue'
import MapControlsPanel from '@/components/map/MapControlsPanel.vue'
import MapPlaceList from '@/components/map/MapPlaceList.vue'
import MapSelectionPanel from '@/components/map/MapSelectionPanel.vue'
import MapTargetProgress from '@/components/map/MapTargetProgress.vue'
import { intlLocale, locale, t } from '@/i18n'
import { useCoinsStore } from '@/stores/coins'
import {
  EMPTY_MAP_FILTERS,
  buildMapPlaces,
  continentTotals,
  filterMapCoins,
  groupCoinsByMintPlace,
  groupCoinsByPlace,
  metricShade,
  metricSteps,
  metricValue,
  placeKeyOf,
  placeLabel,
  targetProgress,
  type CoinStatusSets,
  type MapFilters,
  type MapMetric,
  type MapTarget,
  type PlaceStats,
} from '@/utils/collectionMap'
import { SMALL_TERRITORIES, continentPlaces, type Continent } from '@/utils/mapGeography'
import { MODERN_MINTS, type MintPlace } from '@/utils/modernMints'
import { formatCurrency } from '@/utils/format'
import { formatNumistaYear } from '@/utils/numistaCatalog'
import { countryShadeOpacity, loadWorldCountries, type WorldCountryCollection } from '@/utils/worldCountries'

const PAGE_SIZE = 100

const route = useRoute()
const router = useRouter()
const store = useCoinsStore()

const loading = ref(true)
const failed = ref(false)
const coins = ref<Coin[]>([])
const statuses = ref<CoinStatusSets>({ current: new Set(), demonetized: new Set(), nonCirculating: new Set(), commemorative: new Set(), regular: new Set() })
const world = ref<WorldCountryCollection | null>(null)
const mintLocations = ref<MintPlace[]>([])
const mapRef = ref<InstanceType<typeof CountryCoinsMap> | null>(null)

const metric = ref<MapMetric>('coins')
const target = ref<MapTarget>('')
const filters = ref<MapFilters>({ ...EMPTY_MAP_FILTERS })
const showMints = ref(false)
const showExtinct = ref(false)
const continent = ref<Continent | 'other' | ''>('')
const selectedKey = ref<string | null>(null)
const selectedPin = ref<string | null>(null)

const places = computed(() => buildMapPlaces(world.value))
const filteredCoins = computed(() => filterMapCoins(coins.value, filters.value, statuses.value))
const grouped = computed(() => groupCoinsByPlace(filteredCoins.value, places.value, locale.value))
const eras = computed(() => [...new Set(coins.value.map((coin) => coin.era).filter((era): era is string => Boolean(era)))].sort())
const offMapCount = computed(() => [...grouped.value.byKey.values()].filter((place) => place.kind === 'offMap').length)

function byMetric(a: PlaceStats, b: PlaceStats): number {
  return metricValue(b, metric.value) - metricValue(a, metric.value) || a.label.localeCompare(b.label, intlLocale())
}

const listedPlaces = computed(() => [...grouped.value.byKey.values()]
  .filter((place) => !continent.value || (place.continent ?? 'other') === continent.value)
  .sort(byMetric))

const continents = computed(() => continentTotals(grouped.value.byKey))

// ---- Colours ----

const steps = computed(() => metricSteps(metric.value, [...grouped.value.byKey.values()].map((place) => metricValue(place, metric.value))))

const shades = computed(() => {
  const shadeByKey = new Map<string, number>()
  for (const place of grouped.value.byKey.values()) {
    if (place.kind !== 'offMap') shadeByKey.set(place.key, metricShade(metricValue(place, metric.value), steps.value))
  }
  return shadeByKey
})

const legendTitle = computed(() => t(metric.value === 'value' ? 'map.legend.value' : metric.value === 'types' ? 'map.legend.types' : 'map.legend.coins'))

const legendLabels = computed(() => {
  if (metric.value !== 'value') return ['1', '2–4', '5–9', '10+']
  const [low, middle, high] = steps.value.map((value) => formatCurrency(value))
  return [t('map.legend.upTo', { value: low! }), `${low}–${middle}`, `${middle}–${high}`, t('map.legend.over', { value: high! })]
})

// ---- Target, layers and focus ----

const targetTitle = computed(() => t(target.value === 'euro' ? 'map.controls.targetEuro' : target.value === 'europe' ? 'map.controls.targetEurope' : 'map.controls.targetWorld'))
const progress = computed(() => targetProgress(target.value, places.value, grouped.value.byKey))
const missing = computed(() => new Set(progress.value.missing))

const selectable = computed(() => new Set([
  ...[...grouped.value.byKey.values()].filter((place) => place.kind !== 'offMap').map((place) => place.key),
  ...progress.value.missing,
]))

const overlays = computed<MapOverlay[]>(() => (showExtinct.value
  ? [...grouped.value.byKey.values()]
      .filter((place) => place.territory.length)
      .map((place) => ({ key: place.key, places: place.territory, tooltip: t('map.extinct.tooltip', { country: place.label, count: place.count }) }))
  : []))

const mintPlaces = computed(() => [...MODERN_MINTS, ...mintLocations.value])
const mintGroups = computed(() => groupCoinsByMintPlace(filteredCoins.value, mintPlaces.value, (coin) => placeKeyOf(coin.country ?? '', places.value)))
const pins = computed<MapPin[]>(() => (showMints.value
  ? mintGroups.value.groups.map((group) => ({
      key: group.place.key,
      lat: group.place.lat,
      lng: group.place.lng,
      count: group.coins.length,
      tooltip: t('map.mints.tooltip', { mint: group.place.name, count: group.coins.length }),
    }))
  : []))

const focus = computed(() => (continent.value && continent.value !== 'other' ? continentPlaces(continent.value) : null))

function labelFor(key: string): string {
  const info = places.value.byKey.get(key)
  if (info) return placeLabel(info, locale.value)
  return grouped.value.byKey.get(key)?.label ?? key.replace(/^name:/, '')
}

function describe(key: string, label: string): string {
  const place = grouped.value.byKey.get(key)
  const text = metric.value === 'value'
    ? `${label}: ${formatCurrency(place?.value ?? 0)}`
    : metric.value === 'types'
      ? `${label}: ${t('map.panel.types', { count: place?.types ?? 0 })}`
      : t('collection.worldMap.countryCoins', { country: label, count: place?.count ?? 0 })
  return missing.value.has(key) ? `${text} · ${t('map.legend.missing')}` : text
}

// ---- Selection ----

const panel = computed(() => {
  if (selectedPin.value) {
    const group = mintGroups.value.groups.find((entry) => entry.place.key === selectedPin.value)
    return {
      title: group?.place.name ?? '',
      kicker: t('map.panel.mint'),
      details: group ? [t('map.country.coins', { count: group.coins.length })] : [],
      note: '',
      coins: group?.coins ?? [],
      emptyText: '',
      galleryCountry: null,
    }
  }
  if (!selectedKey.value) return { title: '', kicker: '', details: [], note: '', coins: [], emptyText: '', galleryCountry: null }
  const place = grouped.value.byKey.get(selectedKey.value)
  const title = place?.label ?? labelFor(selectedKey.value)
  const details: string[] = []
  if (place) {
    details.push(t('map.country.coins', { count: place.count }), t('map.panel.types', { count: place.types }))
    if (place.minYear !== undefined && place.maxYear !== undefined) {
      details.push(place.minYear === place.maxYear
        ? t('map.country.year', { year: formatNumistaYear(place.minYear) })
        : t('map.country.years', { from: formatNumistaYear(place.minYear), to: formatNumistaYear(place.maxYear) }))
    }
    if (place.value > 0) details.push(t('map.country.value', { value: formatCurrency(place.value) }))
  }
  return {
    title,
    kicker: '',
    details,
    note: place?.territory.length ? t('map.extinct.territory', { places: place.territory.map(labelFor).join(', ') }) : '',
    coins: place?.coins ?? [],
    emptyText: t('map.panel.noCoins', { country: title }),
    galleryCountry: place?.names[0] ?? null,
  }
})

function remember() {
  const query = { ...route.query }
  delete query.country
  delete query.mint
  if (selectedKey.value) query.country = selectedKey.value
  if (selectedPin.value) query.mint = selectedPin.value
  void router.replace({ query })
}

function selectPlace(key: string | null) {
  selectedKey.value = key
  selectedPin.value = null
  remember()
}

function selectPin(key: string | null) {
  selectedPin.value = key
  selectedKey.value = null
  remember()
}

function clearSelection() {
  selectedKey.value = null
  selectedPin.value = null
  remember()
}

function selectContinent(next: Continent | 'other' | '') {
  continent.value = next
}

function resetView() {
  continent.value = ''
  mapRef.value?.resetView()
}

function showInGallery(country: string) {
  store.selectedCountry = country
  void router.push('/')
}

// ---- Loading ----

async function loadMintLocations() {
  try {
    const res = await getMintLocations()
    const list = Array.isArray(res.data) ? res.data : res.data?.mintLocations ?? []
    mintLocations.value = list.map((location) => ({
      key: `location:${location.id}`,
      name: location.displayName,
      lat: location.lat,
      lng: location.lng,
      aliases: location.aliases ?? [],
    }))
  } catch {
    // The built-in mints are enough to draw the layer.
  }
}

watch(showMints, (show) => {
  if (show && !mintLocations.value.length) void loadMintLocations()
})

async function load() {
  loading.value = true
  failed.value = false
  try {
    const all: Coin[] = []
    for (let page = 1; ; page += 1) {
      const res = await getCoins({ wishlist: 'false', page, limit: PAGE_SIZE })
      all.push(...res.data.coins)
      if (!res.data.coins.length || page * PAGE_SIZE >= res.data.total) break
    }
    coins.value = all
  } catch {
    failed.value = true
    loading.value = false
    return
  }
  try {
    const { data } = await getCoinStatusGroups()
    statuses.value = {
      current: new Set(data.current),
      demonetized: new Set(data.demonetized),
      nonCirculating: new Set(data.nonCirculating ?? []),
      commemorative: new Set(data.commemorative),
      regular: new Set(data.regular),
    }
  } catch {
    // Without it the legal tender and commemorative filters find nothing.
  }
  try {
    world.value = await loadWorldCountries()
  } catch {
    // The lists work without the map.
    world.value = null
  }
  loading.value = false
  if (typeof route.query.country === 'string') selectedKey.value = route.query.country
  if (typeof route.query.mint === 'string') {
    showMints.value = true
    selectedPin.value = route.query.mint
  }
}

onMounted(load)
</script>

<style scoped>
.extinct-swatch {
  background: repeating-linear-gradient(45deg, var(--accent-bronze) 0 2px, transparent 2px 5px);
}
</style>
