import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CollectionOverviewTabs from '../CollectionOverviewTabs.vue'
import { setLocale } from '@/i18n'
import { useCoinsStore } from '@/stores/coins'

const mocks = vi.hoisted(() => ({ getCollectionSummary: vi.fn(), getCoins: vi.fn(), push: vi.fn() }))

vi.mock('@/api/client', () => ({ getCollectionSummary: mocks.getCollectionSummary, getCoins: mocks.getCoins }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: mocks.push }), RouterLink: { template: '<a><slot /></a>' } }))

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots }) => () => h('a', { href: props.to }, slots.default?.()),
})

function summary(coins: number, countries: number, value: number, extra: Record<string, unknown> = {}) {
  return { data: { coins, countries, totalValue: value, currency: 'EUR', currentCoins: 0, commemorativeCoins: 0, coincards: 0, extinctCountries: 0, classifications: [], highlights: { rarest: null, oldest: null, mostValuable: null }, ...extra } }
}

// Whether a tab's panel shows (v-show hides the others).
function shown(wrapper: ReturnType<typeof mountPage>, selector: string): boolean {
  return !(wrapper.get(selector).attributes('style') ?? '').includes('display: none')
}

function mountPage() {
  return mount(CollectionOverviewTabs, { global: { stubs: {
    RouterLink: RouterLinkStub, 'router-link': RouterLinkStub, AuthenticatedImage: true,
  } } })
}

describe('CollectionOverviewTabs', () => {
  beforeEach(() => {
    localStorage.clear()
    setActivePinia(createPinia())
    mocks.getCoins.mockResolvedValue({ data: { coins: [], total: 3 } })
    setLocale('pt-PT')
    vi.clearAllMocks()
    mocks.getCollectionSummary.mockImplementation((kind?: string) => Promise.resolve(
      kind === 'coin' ? summary(452, 84, 3838.11, {
        currentCoins: 192, commemorativeCoins: 200, uniqueCoins: 440, extinctCountries: 5,
        classifications: [
          { key: 'standard', label: 'Standard circulation coin', count: 200 },
          { key: 'circulating-commemorative', label: 'Circulating commemorative coin', count: 190 },
          { key: 'non-circulating', label: 'Non-circulating coin', count: 30 },
          { key: 'Padrão', label: 'Padrão', count: 12 },
        ],
      })
        : kind === 'banknote' ? summary(6, 4, 25, { uniqueCoins: 5, invested: 0, investedPieces: 0, circulation: { current: 2, demonetized: 2, nonCirculating: 0, unknown: 1 }, faceValue: 14.96, withoutFaceValue: 0, valueOverFace: 67.1, highlights: { rarest: null, oldest: { id: 1820, name: '2000 Escudos 1991', country: 'Portugal', year: 1991, dateRange: '1991', rarity: '', grade: '', currentValue: 25 }, mostValuable: null } })
          : summary(455, 84, 3863.11, { uniqueCoins: 441, faceValue: 912.4, withoutFaceValue: 31, valueOverFace: -2.5, extinctCountries: 5, invested: 515.98, investedPieces: 98 }),
    ))
  })

  it('sums up the whole collection, its coins and its banknotes, each in a tab', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(mocks.getCollectionSummary).toHaveBeenCalledWith()
    expect(mocks.getCollectionSummary).toHaveBeenCalledWith('coin')
    expect(mocks.getCollectionSummary).toHaveBeenCalledWith('banknote')
    expect(wrapper.findAll('[role="tab"]').map((tab) => tab.text())).toEqual(['Toda a coleção', 'Moedas', 'Notas'])
    expect(shown(wrapper, '#overview-panel-all')).toBe(true)
    expect(shown(wrapper, '#overview-panel-coins')).toBe(false)
    // Every item held, counting quantities, then each different one once.
    expect(wrapper.get('[data-testid="overview-all-items"]').text()).toBe('455Peças totais441 únicas')
    // What they are worth, against their face value.
    expect(wrapper.get('[data-testid="overview-all-value"]').text()).toContain('3\u00a0863,11')
    expect(wrapper.get('[data-testid="overview-all-face"]').text().replace(/\s/g, ' ')).toMatch(/^Facial: 912,40 \S+ \(-2,5%\)$/)
    expect(wrapper.get('[data-testid="overview-all-face"]').attributes('title')).toContain('31 sem valor facial')
    expect(wrapper.get('[data-testid="overview-all-countries"]').text()).toBe('84Países5 extintos')
    // What the pieces with a purchase price cost.
    const invested = wrapper.get('[data-testid="overview-all-invested"]')
    expect(invested.text().replace(/\s/g, ' ')).toMatch(/^515,98 \S+Valor investidoEm 98 peças com preço de compra$/)
    expect(invested.attributes('title')).toContain('As peças sem preço de compra não contam.')
    // The whole collection has only its numbers: the distributions are in the coins' and banknotes' tabs.
    expect(wrapper.find('#overview-panel-all [data-testid="coin-distribution"]').exists()).toBe(false)
    expect(wrapper.find('#overview-panel-all [data-testid="banknote-distribution"]').exists()).toBe(false)

    // The coins: their numbers, highlights and those held more than once.
    await wrapper.get('[data-testid="overview-tab-coins"]').trigger('click')
    expect(shown(wrapper, '#overview-panel-coins')).toBe(true)
    expect(shown(wrapper, '#overview-panel-all')).toBe(false)
    expect(wrapper.get('[data-testid="overview-coin-items"]').text()).toBe('452Moedas440 únicas')
    // The coins on the wishlist, which open it; only the coins' tab has them.
    expect(mocks.getCoins).toHaveBeenCalledWith(expect.objectContaining({ kind: 'coin', wishlist: 'true' }))
    const wishlist = wrapper.get('[data-testid="overview-coin-wishlist"]')
    expect(wishlist.text()).toBe('3Na lista de desejos')
    expect(wishlist.attributes('href')).toBe('/wishlist')
    expect(wrapper.find('[data-testid="overview-all-wishlist"]').exists()).toBe(false)
    // The coins by kind of coin, the rest together with those Numista says nothing of.
    const distribution = wrapper.get('#overview-panel-coins [data-testid="coin-distribution"]')
    expect(distribution.text()).toContain('Distribuição das moedas (440)')
    expect(distribution.get('[data-testid="distribution-standard"]').text()).toBe('200 Circulantes comuns (45%)')
    expect(distribution.get('[data-testid="distribution-other"]').text()).toBe('20 Outras (5%)')
    expect(distribution.get('[data-testid="distribution-other"] span').attributes('title')).toBe('Padrão: 12\nSem dados do Numista: 8')
    expect(distribution.get('[data-testid="distribution-current"]').text()).toBe('192 com curso legal')
    expect(distribution.find('[data-testid="distribution-coincards"]').exists()).toBe(false)

    // A coin summary without a face value (an older server) shows none.
    expect(wrapper.find('[data-testid="overview-coin-face"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Destaques das moedas')

    // The banknotes', in their own tab.
    await wrapper.get('[data-testid="overview-tab-banknotes"]').trigger('click')
    expect(wrapper.get('[data-testid="overview-banknote-items"]').text()).toBe('6Notas5 únicas')
    // No banknote has a purchase price: nothing to add up.
    expect(wrapper.get('[data-testid="overview-banknote-invested"]').text().replace(/\s+/g, ' ')).toBe('—Valor investidoNenhuma com preço de compra')
    expect(wrapper.get('[data-testid="overview-banknote-face"]').text().replace(/\s/g, ' ')).toMatch(/^Facial: 14,96 \S+ \(\+67,1%\)$/)
    expect(wrapper.text()).toContain('Destaques das notas')
    expect(wrapper.text()).toContain('2000 Escudos 1991')
    expect(wrapper.text()).toContain('Ainda nenhuma nota tem raridade.')
    // The banknotes still legal tender and no longer; each opens them.
    const byCirculation = wrapper.get('#overview-panel-banknotes [data-testid="banknote-distribution"]')
    expect(byCirculation.text()).toContain('Distribuição das notas (5)')
    expect(byCirculation.findAll('li').map((item) => item.text())).toEqual(['2 Com curso legal (40%)', '2 Desmonetizadas (40%)', '1 Outras (20%)'])
    expect(byCirculation.get('[data-testid="distribution-other"] span').attributes('title')).toBe('Sem dados do Numista: 1')
    await byCirculation.get('[data-testid="distribution-demonetized"] button').trigger('click')
    expect(mocks.push).toHaveBeenLastCalledWith({ name: 'banknotes', query: { circulation: 'demonetized' } })
    // Those held more than once are on the wishlist's page, not here.
    expect(wrapper.findComponent({ name: 'CollectionRepeated' }).exists()).toBe(false)
  })

  it('loads its numbers again when asked (the statistics pull to refresh)', async () => {
    const wrapper = mountPage()
    await flushPromises()
    mocks.getCollectionSummary.mockClear()

    await (wrapper.vm as unknown as { load: () => Promise<void> }).load()
    expect(mocks.getCollectionSummary).toHaveBeenCalledTimes(3)
    expect(mocks.getCoins).toHaveBeenCalledTimes(2)
  })

  it('keeps the tab chosen for the next visit', async () => {
    const first = mountPage()
    await flushPromises()
    await first.get('[data-testid="overview-tab-banknotes"]').trigger('click')
    first.unmount()

    const next = mountPage()
    await flushPromises()
    expect(next.get('[data-testid="overview-tab-banknotes"]').attributes('aria-selected')).toBe('true')
    expect(shown(next, '#overview-panel-banknotes')).toBe(true)
  })

  it('opens the gallery with only the chosen filter on', async () => {
    const store = useCoinsStore()
    store.searchQuery = 'escudo'
    store.selectedCountry = 'Portugal'
    const wrapper = mountPage()
    await flushPromises()

    await wrapper.get('#overview-panel-coins [data-testid="distribution-current"]').trigger('click')
    expect(store.selectedCirculation).toBe('current')
    expect(store.searchQuery).toBe('')
    expect(store.selectedCountry).toBe('')
    expect(mocks.push).toHaveBeenCalledWith('/')

    // A kind of coin, by its key.
    await wrapper.get('#overview-panel-coins [data-testid="distribution-circulating-commemorative"] button').trigger('click')
    expect(store.selectedClassification).toBe('circulating-commemorative')
    expect(store.selectedCirculation).toBe('')

    // The coins of countries that no longer exist.
    await wrapper.get('[data-testid="overview-coin-extinct"]').trigger('click')
    expect(store.selectedCountry).not.toBe('')
    expect(store.selectedClassification).toBe('')
  })

  it('turns off the duplicates filter when it opens the gallery with another', async () => {
    const store = useCoinsStore()
    store.selectedRepeated = true
    const wrapper = mountPage()
    await flushPromises()

    await wrapper.get('#overview-panel-coins [data-testid="distribution-current"]').trigger('click')
    expect(store.selectedRepeated).toBe(false)
  })

  it('invites to add a banknote when there is none', async () => {
    mocks.getCollectionSummary.mockImplementation((kind?: string) => Promise.resolve(kind === 'banknote' ? summary(0, 0, 0) : summary(452, 84, 3838.11)))
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.get('[data-testid="overview-no-banknotes"]').get('a').attributes('href')).toBe('/banknotes/add')
  })
})
