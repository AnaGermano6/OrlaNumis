import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CommemorativeCataloguePage from '../CommemorativeCataloguePage.vue'
import { setLocale } from '@/i18n'
import type { CatalogueCountry, CoinRelease } from '@/api/client'

const mocks = vi.hoisted(() => ({
  getCoinCatalogueCountries: vi.fn(),
  getCoinCatalogue: vi.fn(),
  replace: vi.fn(),
  route: { query: {} as Record<string, string> },
}))
vi.mock('@/api/client', () => ({ getCoinCatalogueCountries: mocks.getCoinCatalogueCountries, getCoinCatalogue: mocks.getCoinCatalogue }))
vi.mock('vue-router', () => ({ useRoute: () => mocks.route, useRouter: () => ({ replace: mocks.replace }) }))

function country(code: string, commemoratives: number, ownedCommemoratives: number, collector = 0, ownedCollector = 0): CatalogueCountry {
  return { country: code, commemoratives, ownedCommemoratives, collector, ownedCollector }
}

function coin(fields: Partial<CoinRelease>): CoinRelease {
  const kind = fields.source && fields.source !== 'ecb' ? 'collector' : 'commemorative'
  return { source: 'ecb', kind, country: 'PT', title: 'Coin', faceValue: 2, imageUrl: `https://img/${fields.title ?? 'coin'}.jpg`, pageUrl: 'https://page', from: '2025-06', to: '2025-06', upcoming: false, owned: false, ...fields }
}

const PORTUGAL = [
  coin({ source: 'incm', title: 'Rotary', from: '2026-07', to: '2026-07', owned: true }),
  coin({ title: 'World scouting', owned: true }),
  coin({ title: 'Sustainable development', from: '2025-01', to: '2025-03' }),
  coin({ source: 'incm', title: 'Faróis', faceValue: 5, from: '2025-11', to: '2025-11' }),
  coin({ title: '35 years of the Erasmus programme', from: '2022-07', to: '2022-07', owned: true }),
  // A commemorative in escudos, from Numista: no picture.
  coin({ source: 'numista', kind: 'commemorative', title: 'Colombo e Portugal', faceValue: 200, currency: 'PTE', imageUrl: '', from: '1991-01', to: '1991-12', owned: true }),
]

function cards(wrapper: ReturnType<typeof mount>): string[] {
  return wrapper.findAll('[data-testid="release"]').map((card) => card.find('.line-clamp-3').text())
}

describe('CommemorativeCataloguePage', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    vi.clearAllMocks()
    mocks.route.query = {}
    mocks.getCoinCatalogueCountries.mockResolvedValue({ data: { countries: [country('FR', 20, 1, 50, 0), country('PT', 3, 2, 2, 1), country('DE', 30, 0)] } })
    mocks.getCoinCatalogue.mockImplementation(async (code: string) => ({ data: { country: code, coins: code === 'PT' ? PORTUGAL : [coin({ country: code, title: 'Notre-Dame' })], complete: code !== 'PT' } }))
  })

  it("opens the country the collection has most coins of, by year, marking the user's own", async () => {
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    // Countries by name, with how many of their coins the collection has.
    const countries = wrapper.findAll('button[data-testid^="catalogue-country-"]')
    expect(countries.map((button) => button.find('span').text())).toEqual(['Alemanha', 'França', 'Portugal'])
    expect(wrapper.get('[data-testid="catalogue-country-PT"]').text()).toContain('3 de 5')
    expect(wrapper.get('[data-testid="catalogue-country-PT"]').attributes('aria-pressed')).toBe('true')
    expect(mocks.getCoinCatalogue).toHaveBeenCalledWith('PT')
    expect(mocks.replace).toHaveBeenCalledWith({ query: { country: 'PT' } })

    expect(wrapper.get('[data-testid="catalogue-owned"]').text()).toBe('Tem 3 de 5')
    expect(wrapper.findAll('[data-testid="catalogue-year"]').map((heading) => heading.text().replace(/\s+/g, ' '))).toEqual(['2026 · 1 de 1', '2025 · 1 de 3', '2022 · 1 de 1', '1991 · 1 de 1'])
    expect(wrapper.findAll('[data-testid="release-owned"]').map((label) => label.text())).toEqual(['Na coleção', 'Na coleção', 'Na coleção', 'Na coleção'])
    // The escudo coin: its value in escudos, without a picture; and the list of escudos is not complete yet.
    const escudo = wrapper.findAll('[data-testid="release"]').at(-1)!
    expect(escudo.text()).toContain('200 escudos')
    expect(escudo.find('[data-testid="release-no-image"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="catalogue-partial"]').text()).toContain('quota mensal')
    // The country is not repeated on each coin.
    expect(wrapper.get('[data-testid="release"]').text()).not.toContain('Portugal')
  })

  it('filters by kind of coin and by what the collection has', async () => {
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    await wrapper.get('[data-testid="catalogue-owned-missing"]').trigger('click')
    expect(cards(wrapper)).toEqual(['Sustainable development', 'Faróis'])
    await wrapper.get('[data-testid="catalogue-kind-collector"]').trigger('click')
    expect(cards(wrapper)).toEqual(['Faróis'])
    await wrapper.get('[data-testid="catalogue-owned-owned"]').trigger('click')
    expect(cards(wrapper)).toEqual(['Rotary'])
    // Commemoratives in euros and in escudos.
    await wrapper.get('[data-testid="catalogue-kind-commemoratives"]').trigger('click')
    expect(cards(wrapper)).toEqual(['World scouting', '35 years of the Erasmus programme', 'Colombo e Portugal'])
  })

  it('opens the country in the address and switches country on request', async () => {
    mocks.route.query = { country: 'de' }
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()
    expect(mocks.getCoinCatalogue).toHaveBeenCalledWith('DE')
    // A country without collector coins has no kind filter.
    expect(wrapper.find('[data-testid="catalogue-kind-all"]').exists()).toBe(false)

    await wrapper.get('[data-testid="catalogue-country-FR"]').trigger('click')
    await flushPromises()
    expect(mocks.getCoinCatalogue).toHaveBeenLastCalledWith('FR')
    expect(cards(wrapper)).toEqual(['Notre-Dame'])
    expect(wrapper.find('[data-testid="catalogue-partial"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="catalogue-kind-all"]').exists()).toBe(true)

    // The phone's list chooses the same way.
    const select = wrapper.get('[data-testid="catalogue-country-select"]')
    expect(select.findAll('option').map((option) => option.text())).toEqual(['Alemanha — 0 de 30', 'França — 1 de 70', 'Portugal — 3 de 5'])
    await select.setValue('PT')
    await flushPromises()
    expect(mocks.getCoinCatalogue).toHaveBeenLastCalledWith('PT')
  })

  it('says so when the catalogue cannot be loaded', async () => {
    mocks.getCoinCatalogueCountries.mockRejectedValue(new Error('offline'))
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Não foi possível carregar o catálogo.')
  })
})
