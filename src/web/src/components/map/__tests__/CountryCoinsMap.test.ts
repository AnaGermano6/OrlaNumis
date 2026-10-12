import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CountryCoinsMap from '../CountryCoinsMap.vue'
import type { WorldCountryCollection } from '@/utils/worldCountries'
import type { MapPoint } from '@/utils/mapGeography'

interface FakeLayer {
  kind: string
  latLng?: [number, number]
  options: Record<string, unknown>
  tooltip: string
  handlers: Record<string, () => void>
  bindTooltip: (text: string) => FakeLayer
  on: (event: string, handler: () => void) => FakeLayer
  addTo: () => FakeLayer
  remove: () => void
}

const mocks = vi.hoisted(() => ({
  geoJSON: [] as Array<{ data: { features: unknown[] }; options: { style?: (feature: unknown) => Record<string, unknown>; onEachFeature?: (feature: unknown, layer: unknown) => void }; layer: unknown }>,
  markers: [] as unknown[],
  fitBounds: vi.fn(),
}))

function fakeLayer(kind: string, options: Record<string, unknown> = {}, latLng?: [number, number]): FakeLayer {
  const layer: FakeLayer = {
    kind,
    latLng,
    options,
    tooltip: '',
    handlers: {},
    bindTooltip: (text) => { layer.tooltip = text; return layer },
    on: (event, handler) => { layer.handlers[event] = handler; return layer },
    addTo: () => layer,
    remove: () => {},
  }
  return layer
}

vi.mock('leaflet', () => ({
  map: vi.fn(() => ({ fitBounds: mocks.fitBounds, remove: vi.fn() })),
  geoJSON: vi.fn((data: { features: unknown[] }, options: (typeof mocks.geoJSON)[number]['options']) => {
    const layer = fakeLayer('geojson', options as Record<string, unknown>)
    mocks.geoJSON.push({ data, options, layer })
    return layer
  }),
  circleMarker: vi.fn((latLng: [number, number], options: Record<string, unknown>) => {
    const layer = fakeLayer('marker', options, latLng)
    mocks.markers.push(layer)
    return layer
  }),
  latLngBounds: vi.fn(),
}))
vi.mock('leaflet/dist/leaflet.css', () => ({}))

const square = { type: 'Polygon' as const, coordinates: [[[0, 0], [1, 0], [1, 1], [0, 0]]] }
const WORLD: WorldCountryCollection = {
  type: 'FeatureCollection',
  features: [
    { type: 'Feature', properties: { iso: 'PT', name: 'Portugal', namePt: 'Portugal', aliases: [] }, geometry: square },
    { type: 'Feature', properties: { iso: 'ES', name: 'Spain', namePt: 'Espanha', aliases: [] }, geometry: square },
    { type: 'Feature', properties: { iso: 'RS', name: 'Serbia', namePt: 'Sérvia', aliases: [] }, geometry: square },
  ],
}
const POINTS: MapPoint[] = [{ iso: 'MC', name: 'Monaco', namePt: 'Mónaco', aliases: [], lat: 43.7, lng: 7.4 }]

function mountMap(props: Record<string, unknown> = {}) {
  return mount(CountryCoinsMap, {
    props: {
      collection: WORLD,
      points: POINTS,
      shades: new Map([['PT', 4], ['MC', 1]]),
      describe: (key: string, label: string) => `${label} (${key})`,
      selectable: new Set(['PT', 'MC', 'ES']),
      selectedKey: null,
      ...props,
    },
    attachTo: document.body,
  })
}

function outlineLayer(iso: string) {
  const countries = mocks.geoJSON.find((entry) => entry.options.onEachFeature)!
  const feature = WORLD.features.find((entry) => entry.properties.iso === iso)!
  const layer = fakeLayer('outline')
  countries.options.onEachFeature!(feature, layer)
  return { style: countries.options.style!(feature), layer }
}

describe('CountryCoinsMap', () => {
  beforeEach(() => {
    mocks.geoJSON.length = 0
    mocks.markers.length = 0
    mocks.fitBounds.mockReset()
  })

  it('shades countries and small territories, describes them and selects them', async () => {
    const wrapper = mountMap()
    await flushPromises()

    const portugal = outlineLayer('PT')
    expect(portugal.style.fillOpacity).toBe(0.95)
    expect(portugal.layer.tooltip).toBe('Portugal (PT)')
    portugal.layer.handlers.click!()
    expect(wrapper.emitted('select')).toEqual([['PT']])

    // Serbia has no coins and is not missing: it cannot be chosen.
    expect(outlineLayer('RS').layer.handlers.click).toBeUndefined()

    // Monaco is too small for the outlines: a point.
    const monaco = mocks.markers[0] as FakeLayer
    expect(monaco.latLng).toEqual([43.7, 7.4])
    expect(monaco.options.fillOpacity).toBe(0.5)
    expect(monaco.tooltip).toBe('Monaco (MC)')
    monaco.handlers.click!()
    expect(wrapper.emitted('select')).toEqual([['PT'], ['MC']])
    wrapper.unmount()
  })

  it('outlines missing countries, hatches extinct ones and pins mints', async () => {
    const wrapper = mountMap({
      missing: new Set(['ES']),
      overlays: [{ key: 'name:Jugoslávia', places: ['RS'], tooltip: 'Jugoslávia (já não existe): 1 moeda' }],
      pins: [{ key: 'builtin:incm', lat: 38.7, lng: -9.2, count: 4, tooltip: 'INCM: 4 moedas' }],
    })
    await flushPromises()

    expect(outlineLayer('ES').style.dashArray).toBe('4 3')
    expect(outlineLayer('PT').style.dashArray).toBeUndefined()

    const extinct = mocks.geoJSON.find((entry) => !entry.options.onEachFeature)!
    expect(extinct.data.features).toHaveLength(1)
    const hatchLayer = extinct.layer as FakeLayer
    expect(hatchLayer.tooltip).toBe('Jugoslávia (já não existe): 1 moeda')
    hatchLayer.handlers.click!()
    expect(wrapper.emitted('select')).toEqual([['name:Jugoslávia']])

    const pin = (mocks.markers as FakeLayer[]).find((marker) => marker.tooltip === 'INCM: 4 moedas')!
    expect(pin.latLng).toEqual([38.7, -9.2])
    pin.handlers.click!()
    expect(wrapper.emitted('select-pin')).toEqual([['builtin:incm']])
    wrapper.unmount()
  })

  it('marks the chosen country and clears it when chosen again', async () => {
    const wrapper = mountMap({ selectedKey: 'PT' })
    await flushPromises()
    const portugal = outlineLayer('PT')
    expect(portugal.style.weight).toBe(2.5)
    portugal.layer.handlers.click!()
    expect(wrapper.emitted('select')).toEqual([[null]])
    wrapper.unmount()
  })
})
