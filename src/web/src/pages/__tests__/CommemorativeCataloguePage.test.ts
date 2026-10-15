import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CommemorativeCataloguePage from '../CommemorativeCataloguePage.vue'
import { setLocale } from '@/i18n'
import type { CatalogueCountry, CoinRelease } from '@/api/client'

const mocks = vi.hoisted(() => ({
  getCoinCatalogueCountries: vi.fn(),
  getCoinCatalogue: vi.fn(),
  createSet: vi.fn(),
  replace: vi.fn(),
  push: vi.fn(),
  saveToWishlist: vi.fn(),
  hideCatalogueCoin: vi.fn(),
  joinCatalogueCoin: vi.fn(),
  showCatalogueCoin: vi.fn(),
  route: { query: {} as Record<string, string> },
}))
vi.mock('@/api/client', () => ({
  getCoinCatalogueCountries: mocks.getCoinCatalogueCountries,
  getCoinCatalogue: mocks.getCoinCatalogue,
  createSet: mocks.createSet,
  hideCatalogueCoin: mocks.hideCatalogueCoin,
  joinCatalogueCoin: mocks.joinCatalogueCoin,
  showCatalogueCoin: mocks.showCatalogueCoin,
}))
vi.mock('vue-router', () => ({ useRoute: () => mocks.route, useRouter: () => ({ replace: mocks.replace, push: mocks.push }) }))
vi.mock('@/composables/useReleaseWishlist', () => ({ useReleaseWishlist: () => ({ saveToWishlist: mocks.saveToWishlist }) }))
vi.mock('@/composables/useToast', () => ({ useToast: () => ({ showToast: vi.fn() }) }))

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
  coin({ title: 'Sustainable development', from: '2025-01', to: '2025-03', wishlist: true }),
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

  it("makes a set of the country's coins, of the face value chosen, and opens it", async () => {
    mocks.createSet.mockResolvedValue({ data: { id: 31 } })
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    await wrapper.get('[data-testid="catalogue-create-set"]').trigger('click')
    await flushPromises()
    expect(mocks.createSet).toHaveBeenLastCalledWith(expect.objectContaining({
      name: 'Portugal', setType: 'checklist', checklist: { source: 'catalogue', country: 'PT' },
    }))
    expect(mocks.push).toHaveBeenLastCalledWith({ name: 'set-detail', params: { id: 31 } })

    await wrapper.get('[data-testid="catalogue-value-10"]').trigger('click')
    await wrapper.get('[data-testid="catalogue-create-set"]').trigger('click')
    await flushPromises()
    expect(mocks.createSet).toHaveBeenLastCalledWith(expect.objectContaining({
      // Intl puts a no-break space before the euro sign.
      name: expect.stringMatching(/^Portugal · 10\s€$/), checklist: { source: 'catalogue', country: 'PT', faceValue: 10 },
    }))

    // The escudos: their currency goes with them.
    await wrapper.get('[data-testid="catalogue-country-PT-PTE"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="catalogue-create-set"]').trigger('click')
    await flushPromises()
    expect(mocks.createSet).toHaveBeenLastCalledWith(expect.objectContaining({
      name: 'Portugal · Escudos', checklist: { source: 'catalogue', country: 'PT', currency: 'PTE' },
    }))

    mocks.createSet.mockRejectedValue(new Error('offline'))
    await wrapper.get('[data-testid="catalogue-create-set"]').trigger('click')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Não foi possível criar o conjunto. Tente novamente.')
  })

  it('saves a coin the collection does not have to the wishlist', async () => {
    mocks.saveToWishlist.mockResolvedValue('saved')
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    const card = (title: string) => wrapper.findAll('[data-testid="release"]').find((node) => node.find('.line-clamp-3').text() === title)!
    // Had: nothing to save. On the wishlist already: marked.
    expect(card('Rotary').find('[data-testid="release-wishlist"]').exists()).toBe(false)
    expect(card('Sustainable development').get('[data-testid="release-on-wishlist"]').text()).toBe('Na lista de desejos')
    expect(card('Sustainable development').get('[data-testid="release-wishlist"]').attributes('disabled')).toBeDefined()
    // No other button here: the wishlist's has its words.
    const faróis = card('Faróis').get('[data-testid="release-wishlist"]')
    expect(faróis.text()).toBe('Lista de desejos')
    expect(faróis.attributes('aria-label')).toBe('Guardar na lista de desejos')
    expect(card('Faróis').find('[data-testid="release-add"]').exists()).toBe(false)

    await faróis.trigger('click')
    await flushPromises()
    expect(mocks.saveToWishlist).toHaveBeenCalledWith(expect.objectContaining({ title: 'Faróis' }))
    expect(card('Faróis').get('[data-testid="release-wishlist"]').text()).toBe('Na lista de desejos')
  })

  it('hides a coin, or joins it to another, and shows it again', async () => {
    // The server: Portugal's coins with ids, less those hidden or joined.
    const withIds = PORTUGAL.map((release) => ({ ...release, id: release.title }))
    let hidden: CoinRelease[] = []
    const find = (id: string) => withIds.find((release) => release.id === id)!
    mocks.getCoinCatalogue.mockImplementation(async () => ({ data: {
      country: 'PT', coins: withIds.filter((release) => !hidden.some((other) => other.id === release.id)), hidden, complete: true,
    } }))
    mocks.hideCatalogueCoin.mockImplementation(async (id: string) => { hidden = [...hidden, find(id)] })
    mocks.joinCatalogueCoin.mockImplementation(async (id: string, sameAs: string) => { hidden = [...hidden, { ...find(id), sameAs }] })
    mocks.showCatalogueCoin.mockImplementation(async (id: string) => { hidden = hidden.filter((release) => release.id !== id) })
    const wrapper = mount(CommemorativeCataloguePage, { global: { stubs: { teleport: true } } })
    await flushPromises()
    const card = (title: string) => wrapper.findAll('[data-testid="release"]').find((node) => node.find('.line-clamp-3').text() === title)!
    expect(wrapper.find('[data-testid="catalogue-hidden"]').exists()).toBe(false)

    // Faróis, of 2025: the other coins, those of its year first.
    await card('Faróis').get('[data-testid="release-manage"]').trigger('click')
    const options = () => wrapper.get('[data-testid="catalogue-options"]')
    expect(options().get('h3').text()).toBe('Faróis')
    const choices = () => options().findAll('[data-testid="catalogue-same-option"]').map((option) => option.find('.truncate').text())
    expect(choices()).toEqual(['World scouting', 'Sustainable development', 'Rotary', 'Mosteiro dos Jerónimos'])
    await options().get('[data-testid="catalogue-same-search"]').setValue('jeronimos')
    expect(choices()).toEqual(['Mosteiro dos Jerónimos'])

    // Not saved: said, and the options stay open.
    mocks.hideCatalogueCoin.mockRejectedValueOnce(new Error('offline'))
    await options().get('[data-testid="catalogue-hide"]').trigger('click')
    await flushPromises()
    expect(options().get('[role="alert"]').text()).toBe('Não foi possível guardar a alteração. Tente novamente.')

    await options().get('[data-testid="catalogue-hide"]').trigger('click')
    await flushPromises()
    expect(mocks.hideCatalogueCoin).toHaveBeenLastCalledWith('Faróis')
    expect(wrapper.find('[data-testid="catalogue-options"]').exists()).toBe(false)
    expect(cards(wrapper)).not.toContain('Faróis')
    expect(wrapper.get('[data-testid="catalogue-hidden"] summary').text()).toBe('1 moeda escondida ou juntada')
    // The counts of every country again.
    expect(mocks.getCoinCatalogueCountries).toHaveBeenCalledTimes(2)

    await card('Rotary').get('[data-testid="release-manage"]').trigger('click')
    await options().findAll('[data-testid="catalogue-same-option"]').find((option) => option.text().includes('World scouting'))!.trigger('click')
    await flushPromises()
    expect(mocks.joinCatalogueCoin).toHaveBeenCalledWith('Rotary', 'World scouting')
    const rows = () => wrapper.findAll('[data-testid="catalogue-hidden-coin"]')
    expect(rows().map((row) => text(row))).toEqual([
      expect.stringMatching(/^Faróis.* 2025 Mostrar$/),
      expect.stringMatching(/^Rotary.* 2026 · a mesma que «World scouting» Mostrar$/),
    ])

    await rows()[0].get('[data-testid="catalogue-show"]').trigger('click')
    await flushPromises()
    expect(mocks.showCatalogueCoin).toHaveBeenCalledWith('Faróis')
    expect(cards(wrapper)).toContain('Faróis')
    expect(rows()).toHaveLength(1)
  })

  it('says so when the catalogue cannot be loaded', async () => {
    mocks.getCoinCatalogueCountries.mockRejectedValue(new Error('offline'))
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Não foi possível carregar o catálogo.')
  })
})
