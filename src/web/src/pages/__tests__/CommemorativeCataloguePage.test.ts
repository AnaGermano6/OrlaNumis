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

function country(code: string, commemoratives: number, ownedCommemoratives: number, collector = 0, ownedCollector = 0, currency?: string): CatalogueCountry {
  return { country: code, currency, commemoratives, ownedCommemoratives, collector, ownedCollector }
}

function coin(fields: Partial<CoinRelease>): CoinRelease {
  const kind = fields.source && fields.source !== 'ecb' ? 'collector' : 'commemorative'
  return { source: 'ecb', kind, country: 'PT', title: 'Coin', faceValue: 2, imageUrl: `https://img/${fields.title ?? 'coin'}.jpg`, pageUrl: 'https://page', from: '2025-06', to: '2025-06', upcoming: false, owned: false, ...fields }
}

const PORTUGAL = [
  coin({ source: 'incm', title: 'Rotary', from: '2026-07', to: '2026-07', owned: true }),
  coin({ title: 'World scouting', owned: true }),
  coin({ title: 'Sustainable development', from: '2025-01', to: '2025-03' }),
  coin({ source: 'incm', title: 'Faróis', faceValue: 5, faceValues: [5, 10], from: '2025-11', to: '2025-11' }),
  coin({ source: 'numista', title: 'Mosteiro dos Jerónimos', faceValue: 10, imageUrl: '', from: '2004-01', to: '2004-12' }),
]

const ESCUDOS = [
  coin({ source: 'numista', title: 'Colombo e Portugal', faceValue: 200, currency: 'PTE', imageUrl: '', from: '1991-01', to: '1991-12', owned: true }),
  coin({ source: 'numista', title: 'Discovery of Azores', faceValue: 100, currency: 'PTE', imageUrl: '', from: '1989-01', to: '1989-12' }),
]

function cards(wrapper: ReturnType<typeof mount>): string[] {
  return wrapper.findAll('[data-testid="release"]').map((card) => card.find('.line-clamp-3').text())
}

function text(element: { text: () => string }): string {
  return element.text().replace(/\s+/g, ' ')
}

describe('CommemorativeCataloguePage', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    vi.clearAllMocks()
    mocks.route.query = {}
    mocks.getCoinCatalogueCountries.mockResolvedValue({ data: { countries: [
      country('FR', 20, 1, 50, 0), country('PT', 3, 2, 2, 1), country('DE', 30, 0), country('PT', 40, 1, 6, 0, 'PTE'),
    ] } })
    mocks.getCoinCatalogue.mockImplementation(async (code: string, currency?: string) => ({ data: {
      country: code,
      coins: code === 'PT' ? (currency === 'PTE' ? ESCUDOS : PORTUGAL) : [coin({ country: code, title: 'Notre-Dame' })],
      complete: code !== 'PT' || currency === 'PTE',
    } }))
  })

  it('puts the euro coins and the other coins apart, and opens the country the collection has most of', async () => {
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    // Only the sources above: no introduction.
    expect(wrapper.text()).toContain('Fontes:')
    expect(wrapper.text()).not.toContain('Todas as moedas de 2 € comemorativas publicadas')
    const euro = wrapper.get('[data-testid="catalogue-group-euro"]')
    expect(euro.get('h2').text()).toBe('Moedas do euro')
    expect(euro.findAll('button').map((button) => button.find('span').text())).toEqual(['Alemanha', 'França', 'Portugal'])
    const other = wrapper.get('[data-testid="catalogue-group-other"]')
    expect(other.get('h2').text()).toBe('Outras moedas')
    expect(text(other.get('[data-testid="catalogue-country-PT-PTE"]'))).toContain('Escudos · 1 de 46')

    // Portugal in euros has the most coins of the collection: its euro coins only.
    expect(mocks.getCoinCatalogue).toHaveBeenCalledWith('PT', undefined)
    expect(mocks.replace).toHaveBeenCalledWith({ query: { country: 'PT' } })
    expect(wrapper.get('#catalogue-country-title').text()).toBe('Portugal')
    expect(wrapper.get('[data-testid="catalogue-owned"]').text()).toBe('Tem 3 de 5')
    // In the server's order: the newest first.
    expect(cards(wrapper)).toEqual(['Rotary', 'World scouting', 'Sustainable development', 'Faróis', 'Mosteiro dos Jerónimos'])
    expect(wrapper.findAll('[data-testid="catalogue-year"]').map((heading) => text(heading))).toEqual(['2026 · 1 de 1', '2025 · 1 de 3', '2004 · 0 de 1'])
    expect(wrapper.get('[data-testid="catalogue-partial"]').text()).toContain('quota mensal')
  })

  it('filters by each face value the country has, and by what the collection has', async () => {
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    expect(wrapper.findAll('[data-testid^="catalogue-value-"]').map((chip) => text(chip))).toEqual(['Todos os valores', '2 €', '5 €', '10 €'])
    await wrapper.get('[data-testid="catalogue-value-10"]').trigger('click')
    // A design in several face values counts for each.
    expect(cards(wrapper)).toEqual(['Faróis', 'Mosteiro dos Jerónimos'])
    await wrapper.get('[data-testid="catalogue-value-2"]').trigger('click')
    expect(cards(wrapper)).toEqual(['Rotary', 'World scouting', 'Sustainable development'])
    await wrapper.get('[data-testid="catalogue-owned-missing"]').trigger('click')
    expect(cards(wrapper)).toEqual(['Sustainable development'])
    await wrapper.get('[data-testid="catalogue-value-all"]').trigger('click')
    expect(cards(wrapper)).toEqual(['Sustainable development', 'Faróis', 'Mosteiro dos Jerónimos'])
  })

  it("shows Portugal's escudos on their own, with their face values in escudos", async () => {
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    await wrapper.get('[data-testid="catalogue-country-PT-PTE"]').trigger('click')
    await flushPromises()
    expect(mocks.getCoinCatalogue).toHaveBeenLastCalledWith('PT', 'PTE')
    expect(mocks.replace).toHaveBeenLastCalledWith({ query: { country: 'PT', currency: 'PTE' } })
    expect(wrapper.get('#catalogue-country-title').text()).toBe('Portugal · Escudos')
    expect(cards(wrapper)).toEqual(['Colombo e Portugal', 'Discovery of Azores'])
    expect(wrapper.findAll('[data-testid^="catalogue-value-"]').map((chip) => text(chip))).toEqual(['Todos os valores', '100 escudos', '200 escudos'])
    await wrapper.get('[data-testid="catalogue-value-100"]').trigger('click')
    expect(cards(wrapper)).toEqual(['Discovery of Azores'])
    expect(wrapper.find('[data-testid="release-no-image"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="catalogue-partial"]').exists()).toBe(false)
  })

  it('opens the country and currency in the address, and the phone list chooses the same way', async () => {
    mocks.route.query = { country: 'pt', currency: 'pte' }
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()
    expect(mocks.getCoinCatalogue).toHaveBeenCalledWith('PT', 'PTE')

    const select = wrapper.get('[data-testid="catalogue-country-select"]')
    expect(select.findAll('optgroup').map((group) => group.attributes('label'))).toEqual(['Moedas do euro', 'Outras moedas'])
    expect(select.findAll('option').map((option) => option.text())).toEqual(['Alemanha — 0 de 30', 'França — 1 de 70', 'Portugal — 3 de 5', 'Portugal · Escudos — 1 de 46'])
    await select.setValue('FR')
    await flushPromises()
    expect(mocks.getCoinCatalogue).toHaveBeenLastCalledWith('FR', undefined)
    expect(cards(wrapper)).toEqual(['Notre-Dame'])
    // One face value: no value filter.
    expect(wrapper.find('[data-testid="catalogue-value-all"]').exists()).toBe(false)
  })

  it('says so when the catalogue cannot be loaded', async () => {
    mocks.getCoinCatalogueCountries.mockRejectedValue(new Error('offline'))
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Não foi possível carregar o catálogo.')
  })
})
