import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CollectionWorldMap from '../CollectionWorldMap.vue'
import type { WorldCountryCollection } from '@/utils/worldCountries'

interface FakeLayer {
  tooltip: string
  handlers: Record<string, () => void>
  bindTooltip: (text: string) => void
  on: (event: string, handler: () => void) => void
}

const mocks = vi.hoisted(() => ({
  geoJSONOptions: [] as Array<{ style: (feature: unknown) => Record<string, unknown>; onEachFeature: (feature: unknown, layer: unknown) => void }>,
  fitBounds: vi.fn(),
  remove: vi.fn(),
  loadWorldCountries: vi.fn(),
}))

vi.mock('leaflet', () => ({
  map: vi.fn(() => ({ fitBounds: mocks.fitBounds, remove: mocks.remove })),
  geoJSON: vi.fn((_data: unknown, options: (typeof mocks.geoJSONOptions)[number]) => {
    mocks.geoJSONOptions.push(options)
    return { addTo: vi.fn(function (this: unknown) { return this }), remove: vi.fn() }
  }),
}))
vi.mock('leaflet/dist/leaflet.css', () => ({}))
vi.mock('@/utils/worldCountries', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/utils/worldCountries')>()),
  loadWorldCountries: mocks.loadWorldCountries,
}))

const square = { type: 'Polygon' as const, coordinates: [[[0, 0], [1, 0], [1, 1], [0, 0]]] }
const COUNTRIES: WorldCountryCollection = {
  type: 'FeatureCollection',
  features: [
    { type: 'Feature', properties: { iso: 'PT', name: 'Portugal', namePt: 'Portugal', aliases: ['Portugal', 'Portuguese Republic'] }, geometry: square },
    { type: 'Feature', properties: { iso: 'AO', name: 'Angola', namePt: 'Angola', aliases: ['Angola'] }, geometry: square },
    { type: 'Feature', properties: { iso: 'ES', name: 'Spain', namePt: 'Espanha', aliases: ['Spain', 'Espanha'] }, geometry: square },
  ],
}

function fakeLayer(): FakeLayer {
  const layer: FakeLayer = {
    tooltip: '',
    handlers: {},
    bindTooltip: (text) => { layer.tooltip = text },
    on: (event, handler) => { layer.handlers[event] = handler },
  }
  return layer
}

function mountMap(selectedCountry = '') {
  return mount(CollectionWorldMap, {
    props: { countries: [{ country: 'Portugal', count: 3 }, { country: 'Angola', count: 1 }, { country: 'Corinth', count: 2 }], selectedCountry },
    attachTo: document.body,
  })
}

describe('CollectionWorldMap', () => {
  beforeEach(() => {
    mocks.geoJSONOptions.length = 0
    mocks.fitBounds.mockReset()
    mocks.loadWorldCountries.mockReset()
    mocks.loadWorldCountries.mockResolvedValue(COUNTRIES)
  })

  it('shades countries by coin count and filters the gallery when one is clicked', async () => {
    const wrapper = mountMap()
    await flushPromises()

    expect(mocks.fitBounds).toHaveBeenCalled()
    const { style, onEachFeature } = mocks.geoJSONOptions.at(-1)!
    const [portugal, angola, spain] = COUNTRIES.features
    expect(style(portugal).fillOpacity).toBe(0.65)
    expect(style(angola).fillOpacity).toBe(0.5)
    expect(style(spain).fillOpacity).toBe(0.35)

    const portugalLayer = fakeLayer()
    onEachFeature(portugal, portugalLayer)
    expect(portugalLayer.tooltip).toBe('Portugal: 3 coins')
    portugalLayer.handlers.click!()
    expect(wrapper.emitted('select')).toEqual([['Portugal']])

    const spainLayer = fakeLayer()
    onEachFeature(spain, spainLayer)
    expect(spainLayer.tooltip).toBe('Spain: 0 coins')
    expect(spainLayer.handlers.click).toBeUndefined()
    wrapper.unmount()
  })

  it('lists every country by coin count, marks those off the map and toggles the filter', async () => {
    const wrapper = mountMap('Portugal')
    await flushPromises()

    const chips = wrapper.findAll('li button')
    expect(chips.map((chip) => chip.text())).toEqual(['Portugal · 3', 'Corinth · 2 *', 'Angola · 1'])
    expect(chips[0]!.attributes('aria-pressed')).toBe('true')
    expect(wrapper.text()).toContain('Showing coins from Portugal.')
    expect(mocks.geoJSONOptions.at(-1)!.style(COUNTRIES.features[0]).weight).toBe(2.5)

    await chips[0]!.trigger('click')
    await chips[1]!.trigger('click')
    await wrapper.findAll('button').find((button) => button.text() === 'Show all countries')!.trigger('click')
    expect(wrapper.emitted('select')).toEqual([[''], ['Corinth'], ['']])
    wrapper.unmount()
  })

  it('stays hidden without countries and explains a map that cannot load', async () => {
    expect(mount(CollectionWorldMap, { props: { countries: [], selectedCountry: '' } }).find('section').exists()).toBe(false)

    mocks.loadWorldCountries.mockRejectedValue(new Error('offline'))
    const wrapper = mountMap()
    await flushPromises()
    expect(wrapper.text()).toContain('The map could not be loaded.')
    expect(wrapper.findAll('li button')).toHaveLength(3)
    wrapper.unmount()
  })
})
