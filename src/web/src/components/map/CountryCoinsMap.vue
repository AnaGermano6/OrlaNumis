<template>
  <div
    ref="mapElement"
    class="country-coins-map h-[340px] w-full overflow-hidden rounded-md border border-border-subtle bg-input md:h-[520px]"
    role="img"
    :aria-label="t('collection.worldMap.mapLabel')"
  ></div>
</template>

<script setup lang="ts">
// The collection's world map. Countries (outlines) and small territories
// (points) are shaded by the page's measure; places a target still misses
// get a dashed outline; extinct countries can be hatched over their present
// territory; mints can be shown as pins. Choosing a place selects it (again:
// clears the selection) and the map zooms to it; the page can also zoom to a
// group of places (a continent).
import 'leaflet/dist/leaflet.css'
import * as L from 'leaflet'
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { locale, t } from '@/i18n'
import { countryShadeOpacity, countryFeatureLabel, type WorldCountryCollection, type WorldCountryFeature } from '@/utils/worldCountries'
import type { MapPoint } from '@/utils/mapGeography'

export interface MapOverlay {
  key: string
  /** Present-day places it covers. */
  places: string[]
  tooltip: string
}

export interface MapPin {
  key: string
  lat: number
  lng: number
  count: number
  tooltip: string
}

const props = withDefaults(defineProps<{
  collection: WorldCountryCollection
  points: MapPoint[]
  /** Colour step (0–4) of each place, by key. */
  shades: Map<string, number>
  /** Hover text of a place, from its key and name. */
  describe: (key: string, label: string) => string
  /** Places that can be chosen (with coins, or missing from the target). */
  selectable: Set<string>
  selectedKey: string | null
  /** Places the target still misses. */
  missing?: Set<string>
  /** Extinct countries hatched over their present territory. */
  overlays?: MapOverlay[]
  /** Mints with coins. */
  pins?: MapPin[]
  selectedPin?: string | null
  /** Places to zoom to (a continent), or null. */
  focus?: string[] | null
}>(), { missing: () => new Set<string>(), overlays: () => [], pins: () => [], selectedPin: null, focus: null })

const emit = defineEmits<{
  select: [key: string | null]
  'select-pin': [key: string | null]
}>()

const WORLD: L.LatLngBoundsExpression = [[-56, -170], [78, 180]]
const HATCH_ID = 'country-map-extinct-hatch'

const mapElement = ref<HTMLElement | null>(null)
let map: L.Map | null = null
let layers: L.Layer[] = []
const outlineByKey = new Map<string, L.Layer>()
const pointByKey = new Map<string, MapPoint>()

// Leaflet writes SVG attributes, which cannot use var(); read the theme's values.
function cssColor(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim() || 'currentColor'
}

function placeStyle(key: string): L.PathOptions {
  const shade = props.shades.get(key) ?? 0
  const selected = key === props.selectedKey
  const missing = props.missing.has(key)
  return {
    color: selected ? cssColor('--accent-gold') : missing ? cssColor('--text-warning') : cssColor('--bg-card'),
    weight: selected ? 2.5 : missing ? 1.6 : 0.6,
    dashArray: missing && !selected ? '4 3' : undefined,
    fillColor: shade === 0 ? cssColor('--border-subtle') : cssColor('--accent-gold'),
    fillOpacity: countryShadeOpacity(shade),
  }
}

function bindPlace(key: string, label: string, layer: L.Layer) {
  layer.bindTooltip(props.describe(key, label), { sticky: true })
  if (props.selectable.has(key)) layer.on('click', () => emit('select', key === props.selectedKey ? null : key))
}

// The hatching is an SVG pattern in the map's own drawing.
function ensureHatch() {
  const svg = mapElement.value?.querySelector('.leaflet-overlay-pane svg')
  if (!svg || svg.querySelector(`#${HATCH_ID}`)) return
  const ns = 'http://www.w3.org/2000/svg'
  const defs = document.createElementNS(ns, 'defs')
  const pattern = document.createElementNS(ns, 'pattern')
  pattern.setAttribute('id', HATCH_ID)
  pattern.setAttribute('patternUnits', 'userSpaceOnUse')
  pattern.setAttribute('width', '6')
  pattern.setAttribute('height', '6')
  pattern.setAttribute('patternTransform', 'rotate(45)')
  const line = document.createElementNS(ns, 'line')
  line.setAttribute('x1', '0')
  line.setAttribute('y1', '0')
  line.setAttribute('x2', '0')
  line.setAttribute('y2', '6')
  line.setAttribute('stroke', cssColor('--accent-bronze'))
  line.setAttribute('stroke-width', '2')
  pattern.appendChild(line)
  defs.appendChild(pattern)
  svg.insertBefore(defs, svg.firstChild)
}

function draw() {
  if (!map) return
  for (const layer of layers) layer.remove()
  layers = []
  outlineByKey.clear()
  pointByKey.clear()

  layers.push(L.geoJSON(props.collection, {
    style: (feature) => placeStyle((feature as WorldCountryFeature).properties.iso),
    onEachFeature: (feature, layer) => {
      const country = feature as WorldCountryFeature
      outlineByKey.set(country.properties.iso, layer)
      bindPlace(country.properties.iso, countryFeatureLabel(country, locale.value), layer)
    },
  }).addTo(map))

  for (const point of props.points) {
    pointByKey.set(point.iso, point)
    const style = placeStyle(point.iso)
    const marker = L.circleMarker([point.lat, point.lng], { ...style, radius: point.iso === props.selectedKey ? 7 : 5, weight: Math.max(style.weight ?? 1, 1) })
    bindPlace(point.iso, locale.value === 'pt-PT' ? point.namePt : point.name, marker)
    layers.push(marker.addTo(map))
  }

  if (props.overlays.length) {
    const hatch = { color: cssColor('--accent-bronze'), weight: 1.2, dashArray: '2 3', fillColor: `url(#${HATCH_ID})`, fillOpacity: 1 }
    for (const overlay of props.overlays) {
      const features = props.collection.features.filter((feature) => overlay.places.includes(feature.properties.iso))
      const onSelect = () => emit('select', overlay.key === props.selectedKey ? null : overlay.key)
      if (features.length) {
        const layer = L.geoJSON({ type: 'FeatureCollection', features } as WorldCountryCollection, { style: () => hatch })
        layer.bindTooltip(overlay.tooltip, { sticky: true })
        layer.on('click', onSelect)
        layers.push(layer.addTo(map))
      }
      for (const place of overlay.places) {
        const point = pointByKey.get(place)
        if (!point) continue
        const marker = L.circleMarker([point.lat, point.lng], { ...hatch, fillColor: cssColor('--accent-bronze'), fillOpacity: 0.5, radius: 7 })
        marker.bindTooltip(overlay.tooltip, { sticky: true })
        marker.on('click', onSelect)
        layers.push(marker.addTo(map))
      }
    }
    ensureHatch()
  }

  for (const pin of props.pins) {
    const selected = pin.key === props.selectedPin
    const marker = L.circleMarker([pin.lat, pin.lng], {
      radius: 4 + Math.min(10, Math.sqrt(pin.count) * 1.5),
      color: selected ? cssColor('--text-primary') : cssColor('--accent-gold'),
      weight: selected ? 3 : 2,
      fillColor: cssColor('--bg-card'),
      fillOpacity: 0.9,
    })
    marker.bindTooltip(pin.tooltip, { sticky: true })
    marker.on('click', () => emit('select-pin', selected ? null : pin.key))
    layers.push(marker.addTo(map))
  }
}

function boundsOf(keys: string[]): L.LatLngBounds | undefined {
  let bounds: L.LatLngBounds | undefined
  const extend = (next: L.LatLngBoundsExpression) => {
    bounds = bounds ? bounds.extend(next as L.LatLngBounds) : L.latLngBounds(next as L.LatLngBoundsLiteral)
  }
  for (const key of keys) {
    const layer = outlineByKey.get(key)
    if (layer && 'getBounds' in layer) {
      const layerBounds = (layer as L.Polygon).getBounds()
      if (layerBounds.isValid()) extend(layerBounds)
      continue
    }
    const point = pointByKey.get(key)
    if (point) extend([[point.lat - 1, point.lng - 1], [point.lat + 1, point.lng + 1]])
  }
  return bounds?.isValid() ? bounds : undefined
}

function zoomTo(keys: string[] | null) {
  if (!map) return
  const bounds = keys?.length ? boundsOf(keys) : undefined
  if (bounds) map.fitBounds(bounds, { padding: [24, 24], maxZoom: 4 })
  else map.fitBounds(WORLD)
}

function resetView() {
  map?.fitBounds(WORLD)
}

onMounted(() => {
  if (!mapElement.value) return
  map = L.map(mapElement.value, {
    attributionControl: false,
    zoomSnap: 0.25,
    minZoom: 0.75,
    maxZoom: 8,
    // The page scrolls over the map; zoom with the buttons, a double click or a pinch.
    scrollWheelZoom: false,
    worldCopyJump: false,
    maxBounds: [[-75, -200], [88, 200]],
  })
  map.fitBounds(WORLD)
  draw()
  if (props.selectedKey) zoomTo([props.selectedKey])
  else if (props.focus) zoomTo(props.focus)
})

watch(
  () => [props.shades, props.selectable, props.missing, props.overlays, props.pins, props.selectedPin, locale.value],
  draw,
)
watch(() => props.selectedKey, (key) => {
  draw()
  if (key) zoomTo([key])
})
watch(() => props.focus, (focus) => zoomTo(focus))

onBeforeUnmount(() => {
  map?.remove()
  map = null
  layers = []
  outlineByKey.clear()
  pointByKey.clear()
})

defineExpose({ resetView })
</script>

<style scoped>
/* Leaflet adds .leaflet-container (grey background) to the element itself. */
.country-coins-map.leaflet-container {
  background: var(--bg-input);
}
</style>
