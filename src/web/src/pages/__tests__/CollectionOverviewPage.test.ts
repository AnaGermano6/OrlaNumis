import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CollectionOverviewPage from '../CollectionOverviewPage.vue'
import { setLocale } from '@/i18n'
import { useCoinsStore } from '@/stores/coins'

const mocks = vi.hoisted(() => ({ getCollectionSummary: vi.fn(), push: vi.fn() }))

vi.mock('@/api/client', () => ({ getCollectionSummary: mocks.getCollectionSummary }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: mocks.push }), RouterLink: { template: '<a><slot /></a>' } }))

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots }) => () => h('a', { href: props.to }, slots.default?.()),
})

function summary(coins: number, countries: number, value: number, extra: Record<string, unknown> = {}) {
  return { data: { coins, countries, totalValue: value, currency: 'EUR', currentCoins: 0, commemorativeCoins: 0, coincards: 0, extinctCountries: 0, classifications: [], highlights: { rarest: null, oldest: null, mostValuable: null }, ...extra } }
}

function mountPage() {
  return mount(CollectionOverviewPage, { global: { stubs: {
    RouterLink: RouterLinkStub, 'router-link': RouterLinkStub, AuthenticatedImage: true,
    CollectionRepeated: true, CoinReleases: true,
  } } })
}

describe('CollectionOverviewPage', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    setLocale('pt-PT')
    vi.clearAllMocks()
    mocks.getCollectionSummary.mockImplementation((kind?: string) => Promise.resolve(
      kind === 'coin' ? summary(452, 84, 3838.11, { currentCoins: 192, commemorativeCoins: 200, uniqueCoins: 440 })
        : kind === 'banknote' ? summary(3, 1, 25, { uniqueCoins: 1, highlights: { rarest: null, oldest: { id: 1820, name: '2000 Escudos 1991', country: 'Portugal', year: 1991, dateRange: '1991', rarity: '', grade: '', currentValue: 25 }, mostValuable: null } })
          : summary(455, 84, 3863.11, { uniqueCoins: 441 }),
    ))
  })

  it('sums up the whole collection, then its coins and its banknotes', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(mocks.getCollectionSummary).toHaveBeenCalledWith()
    expect(mocks.getCollectionSummary).toHaveBeenCalledWith('coin')
    expect(mocks.getCollectionSummary).toHaveBeenCalledWith('banknote')
    // Every item held, counting quantities, then each different one once.
    expect(wrapper.get('[data-testid="overview-kpi-items"]').text()).toBe('455Moedas e notas')
    expect(wrapper.get('[data-testid="overview-kpi-unique"]').text()).toBe('441Moedas e notas únicas')
    expect(wrapper.get('[data-testid="overview-kpi-value"]').text()).toContain('3863,11')
    // The gallery's numbers and highlights for the coins…
    expect(wrapper.get('[data-testid="kpi-coins"]').text()).toContain('452')
    expect(wrapper.get('[data-testid="kpi-unique"]').text()).toContain('440')
    expect(wrapper.get('[data-testid="kpi-current"]').text()).toContain('192')
    expect(wrapper.text()).toContain('Destaques das moedas')
    // …and the banknotes', under their own title.
    expect(wrapper.get('#overview-banknotes-title').text()).toBe('Notas')
    expect(wrapper.get('[data-testid="banknotes-kpi-banknotes"]').text()).toBe('3Notas')
    expect(wrapper.get('[data-testid="banknotes-kpi-unique"]').text()).toBe('1Notas únicas')
    expect(wrapper.text()).toContain('Destaques das notas')
    expect(wrapper.text()).toContain('2000 Escudos 1991')
    expect(wrapper.text()).toContain('Ainda nenhuma nota tem raridade.')
    expect(wrapper.get('[data-testid="overview-open-banknotes"]').attributes('href')).toBe('/banknotes')
    // The repeated coins and banknotes and the new releases; no coin or banknote of the day.
    expect(wrapper.text()).not.toContain('Moeda do dia')
    expect(wrapper.text()).not.toContain('Nota do dia')
    expect(wrapper.findAll('collection-repeated-stub').map((stub) => stub.attributes('kind'))).toEqual(['coin', 'banknote'])
    expect(wrapper.find('coin-releases-stub').exists()).toBe(true)
  })

  it('opens the gallery with only the chosen card filter on', async () => {
    const store = useCoinsStore()
    store.searchQuery = 'escudo'
    store.selectedCountry = 'Portugal'
    const wrapper = mountPage()
    await flushPromises()

    await wrapper.get('[data-testid="kpi-current"]').trigger('click')
    expect(store.selectedCirculation).toBe('current')
    expect(store.searchQuery).toBe('')
    expect(store.selectedCountry).toBe('')
    expect(mocks.push).toHaveBeenCalledWith('/')
  })

  it('opens the coin gallery, or the banknotes, with only the duplicates to trade', async () => {
    const store = useCoinsStore()
    store.selectedCountry = 'Portugal'
    const wrapper = mountPage()
    await flushPromises()

    const [coins, banknotes] = wrapper.findAllComponents({ name: 'CollectionRepeated' })
    coins!.vm.$emit('open')
    expect(store.selectedRepeated).toBe(true)
    expect(store.selectedCountry).toBe('')
    expect(mocks.push).toHaveBeenLastCalledWith('/')
    banknotes!.vm.$emit('open')
    expect(mocks.push).toHaveBeenLastCalledWith({ name: 'banknotes', query: { repeated: 'true' } })

    // Another card's filter turns it off.
    await wrapper.get('[data-testid="kpi-current"]').trigger('click')
    expect(store.selectedRepeated).toBe(false)
  })

  it('invites to add a banknote when there is none', async () => {
    mocks.getCollectionSummary.mockImplementation((kind?: string) => Promise.resolve(kind === 'banknote' ? summary(0, 0, 0) : summary(452, 84, 3838.11)))
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.get('[data-testid="overview-no-banknotes"]').get('a').attributes('href')).toBe('/banknotes/add')
  })
})
