import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import CollectionMapPage from '../CollectionMapPage.vue'
import { setLocale } from '@/i18n'
import { useCoinsStore } from '@/stores/coins'
import type { Coin } from '@/types'
import type { WorldCountryCollection } from '@/utils/worldCountries'

const mocks = vi.hoisted(() => ({
  getCoins: vi.fn(),
  getCoinStatusGroups: vi.fn(),
  getMintLocations: vi.fn(),
  loadWorldCountries: vi.fn(),
  replace: vi.fn(),
  push: vi.fn(),
  query: {} as Record<string, string>,
  mapProps: [] as Record<string, unknown>[],
}))

vi.mock('@/api/client', () => ({
  getCoins: mocks.getCoins,
  getCoinStatusGroups: mocks.getCoinStatusGroups,
  getMintLocations: mocks.getMintLocations,
}))
vi.mock('vue-router', () => ({
  useRoute: () => ({ query: mocks.query }),
  useRouter: () => ({ replace: mocks.replace, push: mocks.push }),
  RouterLink: { template: '<a><slot /></a>' },
}))
vi.mock('@/utils/worldCountries', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/utils/worldCountries')>()),
  loadWorldCountries: mocks.loadWorldCountries,
}))
vi.mock('@/components/map/CountryCoinsMap.vue', () => ({
  default: defineComponent({
    props: {
      collection: { type: Object, required: true },
      points: { type: Array, required: true },
      shades: { type: Map, required: true },
      describe: { type: Function, required: true },
      selectable: { type: Set, required: true },
      selectedKey: { type: String, default: null },
      missing: { type: Set, default: () => new Set() },
      overlays: { type: Array, default: () => [] },
      pins: { type: Array, default: () => [] },
      selectedPin: { type: String, default: null },
      focus: { type: Array, default: null },
    },
    emits: ['select', 'select-pin'],
    setup(props) {
      return () => {
        mocks.mapProps.push({ ...props })
        return h('div', { 'data-testid': 'map' })
      }
    },
  }),
}))

const square = { type: 'Polygon' as const, coordinates: [[[0, 0], [1, 0], [1, 1], [0, 0]]] }
const WORLD: WorldCountryCollection = {
  type: 'FeatureCollection',
  features: [
    { type: 'Feature', properties: { iso: 'PT', name: 'Portugal', namePt: 'Portugal', aliases: ['Portugal'] }, geometry: square },
    { type: 'Feature', properties: { iso: 'FR', name: 'France', namePt: 'França', aliases: ['França'] }, geometry: square },
    { type: 'Feature', properties: { iso: 'RS', name: 'Serbia', namePt: 'Sérvia', aliases: ['Serbia'] }, geometry: square },
    { type: 'Feature', properties: { iso: 'ES', name: 'Spain', namePt: 'Espanha', aliases: ['Espanha'] }, geometry: square },
  ],
}

function coin(id: number, fields: Partial<Coin>): Coin {
  return { id, name: `Coin ${id}`, era: 'modern', images: [], references: [], currentValue: null, ...fields } as Coin
}

const COINS = [
  coin(1, { name: '2 euro 2007', country: 'Portugal', year: 2007, currentValue: 3, mint: 'Imprensa Nacional - Casa da Moeda' }),
  coin(2, { name: '1 escudo 1980', country: 'Portugal', year: 1980, currentValue: 0.5 }),
  coin(3, { name: '2 euro 2020', country: 'França', year: 2020, currentValue: 2 }),
  coin(4, { name: '1 euro 2010', country: 'Mónaco', year: 2010, currentValue: 10 }),
  coin(5, { name: '10 dinar 1987', country: 'Jugoslávia', year: 1987 }),
]

function lastMap() {
  return mocks.mapProps.at(-1)!
}

async function mountPage() {
  const wrapper = mount(CollectionMapPage, {
    global: { stubs: { DesktopPageContext: { template: '<div><slot /></div>' }, AppIconButton: true, RouterLink: { template: '<a><slot /></a>' }, CoinListTable: { props: ['coins'], template: '<ol><li v-for="c in coins" :key="c.id">{{ c.name }}</li></ol>' } } },
  })
  await flushPromises()
  return wrapper
}

describe('CollectionMapPage', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    setLocale('pt-PT')
    Object.values(mocks).forEach((mock) => typeof mock === 'function' && 'mockReset' in mock && mock.mockReset())
    mocks.mapProps.length = 0
    for (const key of Object.keys(mocks.query)) delete mocks.query[key]
    // Two pages of coins.
    mocks.getCoins.mockImplementation(async ({ page }: { page: number }) => ({ data: { coins: page === 1 ? COINS.slice(0, 3) : COINS.slice(3), total: 103 } }))
    mocks.getCoinStatusGroups.mockResolvedValue({ data: { current: [1, 3, 4], demonetized: [2, 5], commemorative: [3], regular: [1, 2, 4] } })
    mocks.getMintLocations.mockResolvedValue({ data: [] })
    mocks.loadWorldCountries.mockResolvedValue(WORLD)
  })
  afterEach(() => setLocale('en'))

  it('loads every coin and shades countries, territories and nothing off the map', async () => {
    const wrapper = await mountPage()

    expect(mocks.getCoins).toHaveBeenCalledTimes(2)
    expect(wrapper.get('[data-testid="map-summary"]').text()).toContain('4 países')
    expect(wrapper.get('[data-testid="map-summary"]').text()).toContain('5 moedas')
    const shades = lastMap().shades as Map<string, number>
    expect(Object.fromEntries(shades)).toEqual({ PT: 2, FR: 1, MC: 1 })
    expect((lastMap().describe as (key: string, label: string) => string)('PT', 'Portugal')).toBe('Portugal: 2 moedas')
    // The list ranks the places, Yugoslavia off the map.
    expect(wrapper.get('[data-testid="map-country-list"]').text().replace(/\s+/g, ' ')).toContain('Jugoslávia *')
  })

  it('lists the chosen country\'s coins, keeps the choice in the address and opens the gallery on it', async () => {
    const wrapper = await mountPage()
    await wrapper.findAll('[data-testid="map-country-list"] button').find((button) => button.text().startsWith('Portugal'))!.trigger('click')

    const panel = wrapper.get('[data-testid="map-country-panel"]')
    expect(panel.text()).toContain('Portugal')
    expect(panel.text()).toContain('Anos 1980–2007')
    expect(panel.findAll('li').map((item) => item.text())).toEqual(['1 escudo 1980', '2 euro 2007'])
    expect(mocks.replace).toHaveBeenLastCalledWith({ query: { country: 'PT' } })
    expect(lastMap().selectedKey).toBe('PT')

    await panel.findAll('button').find((button) => button.text() === 'Ver na galeria')!.trigger('click')
    expect(useCoinsStore().selectedCountry).toBe('Portugal')
    expect(mocks.push).toHaveBeenCalledWith('/')
  })

  it('colours by value and filters by legal tender, commemorative and years', async () => {
    const wrapper = await mountPage()
    await wrapper.get('[data-testid="map-metric"]').setValue('value')
    expect(wrapper.get('[data-testid="map-legend"]').text()).toContain('Valor por país')

    await wrapper.get('[data-testid="map-circulation"]').setValue('current')
    expect(wrapper.text()).toContain('3 de 5 moedas com estes filtros')
    // Values 3 €, 2 € and 10 €: steps by quarters.
    expect(Object.fromEntries(lastMap().shades as Map<string, number>)).toEqual({ PT: 2, FR: 1, MC: 4 })

    await wrapper.get('[data-testid="map-commemorative"]').setValue('true')
    expect(Object.keys(Object.fromEntries(lastMap().shades as Map<string, number>))).toEqual(['FR'])

    await wrapper.findAll('button').find((button) => button.text() === 'Limpar filtros')!.trigger('click')
    await wrapper.get('[data-testid="map-year-to"]').setValue('1999')
    await wrapper.get('[data-testid="map-year-to"]').trigger('change')
    expect(wrapper.text()).toContain('2 de 5 moedas com estes filtros')
  })

  it('shows the euro countries still missing and lets one be chosen', async () => {
    const wrapper = await mountPage()
    await wrapper.get('[data-testid="map-target"]').setValue('euro')

    const progress = wrapper.get('[data-testid="map-target-progress"]')
    expect(progress.text()).toContain('Países do euro: moedas de 3 de 25')
    expect((lastMap().missing as Set<string>).has('ES')).toBe(true)
    expect((lastMap().selectable as Set<string>).has('ES')).toBe(true)

    await progress.findAll('button').find((button) => button.text() === 'Espanha')!.trigger('click')
    expect(wrapper.get('[data-testid="map-country-panel"]').text()).toContain('Ainda não tem moedas de Espanha.')
  })

  it('hatches extinct countries, pins mints and narrows the list to a continent', async () => {
    const wrapper = await mountPage()
    await wrapper.get('[data-testid="map-layer-extinct"]').setValue(true)
    expect(lastMap().overlays).toEqual([{ key: 'name:Jugoslávia', places: ['RS', 'HR', 'SI', 'BA', 'MK', 'ME', 'XK'], tooltip: 'Jugoslávia (já não existe): 1 moeda' }])

    await wrapper.get('[data-testid="map-layer-mints"]').setValue(true)
    await flushPromises()
    expect(mocks.getMintLocations).toHaveBeenCalled()
    expect(lastMap().pins).toMatchObject([{ key: 'builtin:incm', count: 1 }])

    const continents = wrapper.get('[data-testid="map-continents"]')
    expect(continents.text()).toContain('Europa')
    await continents.findAll('button').find((button) => button.text().startsWith('Europa'))!.trigger('click')
    expect(lastMap().focus).toContain('PT')
    expect(wrapper.get('[data-testid="map-country-list"]').text()).toContain('Jugoslávia')
  })

  it('reopens on the place kept in the address', async () => {
    mocks.query.country = 'MC'
    const wrapper = await mountPage()
    expect(wrapper.get('[data-testid="map-country-panel"]').text()).toContain('Mónaco')
    expect(lastMap().selectedKey).toBe('MC')
  })
})
