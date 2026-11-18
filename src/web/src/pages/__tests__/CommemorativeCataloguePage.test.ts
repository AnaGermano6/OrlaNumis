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
  removeFromWishlist: vi.fn(),
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
vi.mock('vue-router', () => ({
  useRoute: () => mocks.route,
  useRouter: () => ({ replace: mocks.replace, push: mocks.push }),
  RouterLink: { props: ['to'], template: '<a href="#"><slot /></a>' },
}))
vi.mock('@/composables/useReleaseWishlist', () => ({ useReleaseWishlist: () => ({ saveToWishlist: mocks.saveToWishlist, removeFromWishlist: mocks.removeFromWishlist }) }))
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

  it('lists the euro countries beside, the escudos apart, and opens the country the collection has most of', async () => {
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    // The sources, under the coins.
    expect(wrapper.text()).toContain('Fontes:')
    const list = wrapper.get('[data-testid="catalogue-countries"]')
    expect(list.get('h1').text()).toBe('Catálogo de comemorativas')
    const names = () => list.findAll('[data-testid^="catalogue-country-"]').map((button) => button.find('.truncate').text())
    expect(names()).toEqual(['Alemanha', 'França', 'Portugal'])
    expect(list.findAll('[data-testid="catalogue-flag"]').every((flag) => flag.attributes('aria-hidden') === 'true')).toBe(true)
    // The country chosen says how many of its coins the collection has; the others, a bar.
    const parts = (code: string) => list.get(`[data-testid="catalogue-country-${code}"]`).findAll(':scope > span').map((part) => part.text())
    expect(parts('PT')).toEqual(['🇵🇹', 'Portugal', '60%', '3/5'])
    expect(parts('FR')).toEqual(['🇫🇷', 'França', '1%', ''])
    // The euro coins, or the escudos.
    expect(list.findAll('[data-testid^="catalogue-group-"]').map((button) => button.text())).toEqual(['Moedas do euro', 'Escudos'])
    await list.get('[data-testid="catalogue-group-other"]').trigger('click')
    expect(names()).toEqual(['Portugal'])
    expect(list.find('[data-testid="catalogue-country-PT-PTE"]').exists()).toBe(true)
    await list.get('[data-testid="catalogue-group-euro"]').trigger('click')
    // A country searched for, without accents.
    await list.get('[data-testid="catalogue-search"] input').setValue('franca')
    expect(names()).toEqual(['França'])
    await list.get('[data-testid="catalogue-search"] input').setValue('narnia')
    expect(list.get('[data-testid="catalogue-search-none"]').text()).toBe('Nenhum país corresponde à pesquisa.')

    // Portugal in euros has the most coins of the collection: its euro coins only.
    expect(mocks.getCoinCatalogue).toHaveBeenCalledWith('PT', undefined)
    expect(mocks.replace).toHaveBeenCalledWith({ query: { country: 'PT' } })
    // Its flag beside its name, as in the coin gallery.
    expect(wrapper.get('#catalogue-country-title').text()).toBe('🇵🇹Portugal')
    expect(wrapper.get('[data-testid="catalogue-title-flag"]').attributes('aria-hidden')).toBe('true')
    expect(wrapper.get('[data-testid="catalogue-owned"]').text()).toBe('3 / 5 moedas (60%)')
    // In the server's order: the newest first, two in a row on a wide screen.
    expect(cards(wrapper)).toEqual(['Rotary', 'World scouting', 'Sustainable development', 'Faróis', 'Mosteiro dos Jerónimos'])
    expect(wrapper.get('[data-testid="catalogue-coins"]').classes()).toContain('xl:grid-cols-2')
    expect(wrapper.get('[data-testid="catalogue-partial"]').text()).toContain('quota mensal')
  })

  it('filters by year, by each face value the country has, and by what the collection has', async () => {
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    const year = wrapper.get('[data-testid="catalogue-year-filter"]')
    expect(year.findAll('option').map((option) => option.text())).toEqual(['Ano: Todos', 'Ano: 2026', 'Ano: 2025', 'Ano: 2004'])
    await year.setValue('2025')
    expect(cards(wrapper)).toEqual(['World scouting', 'Sustainable development', 'Faróis'])
    await year.setValue('')

    const value = wrapper.get('[data-testid="catalogue-value-filter"]')
    expect(value.findAll('option').map((option) => text(option))).toEqual(['Valor: Todos', 'Valor: 2 €', 'Valor: 5 €', 'Valor: 10 €'])
    await value.setValue('10')
    // A design in several face values counts for each.
    expect(cards(wrapper)).toEqual(['Faróis', 'Mosteiro dos Jerónimos'])
    await value.setValue('2')
    expect(cards(wrapper)).toEqual(['Rotary', 'World scouting', 'Sustainable development'])
    await wrapper.get('[data-testid="catalogue-owned-missing"]').trigger('click')
    expect(wrapper.get('[data-testid="catalogue-owned-missing"]').attributes('aria-pressed')).toBe('true')
    expect(cards(wrapper)).toEqual(['Sustainable development'])
    await value.setValue('')
    expect(cards(wrapper)).toEqual(['Sustainable development', 'Faróis', 'Mosteiro dos Jerónimos'])
  })

  it('leaves out the special editions: gold coins, and those in proof or FDC only', async () => {
    mocks.getCoinCatalogue.mockImplementation(async () => ({ data: { country: 'PT', complete: true, coins: [
      coin({ source: 'incm', title: 'Rotary', metals: ['base'] }),
      coin({ source: 'incm', title: 'FIFA World Cup', faceValue: 5, faceValues: [5, 7.5], metals: ['silver', 'base'], variants: ['Prata Proof'] }),
      coin({ source: 'incm', title: 'Amadeo de Souza-Cardoso', faceValue: 10, metals: ['silver'], specialFinish: true, variants: ['Prata Proof'] }),
      coin({ source: 'incm', title: 'Moeda Emblemática', faceValue: 8, metals: ['gold'] }),
      coin({ source: 'numista', title: 'Mosteiro dos Jerónimos', faceValue: 10, imageUrl: '' }),
    ] } }))
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    const chip = wrapper.get('[data-testid="catalogue-no-special"]')
    expect(chip.text()).toBe('Sem edições especiais')
    await chip.trigger('click')
    expect(chip.attributes('aria-pressed')).toBe('true')
    // A design also struck otherwise stays, and so does one whose metal is not known.
    expect(cards(wrapper)).toEqual(['Rotary', 'FIFA World Cup', 'Mosteiro dos Jerónimos'])
    // Its face values and precious metal in a pill.
    const fifa = wrapper.findAll('[data-testid="release"]').find((node) => node.find('.line-clamp-3').text() === 'FIFA World Cup')!
    expect(text(fifa.get('[data-testid="release-values"]'))).toBe('5 € · 7,50 € · Prata')

    // A country without special editions has nothing to leave out.
    mocks.getCoinCatalogue.mockImplementation(async (code: string) => ({ data: { country: code, complete: true, coins: [coin({ country: code, title: 'Notre-Dame', metals: ['base'] })] } }))
    await wrapper.get('[data-testid="catalogue-country-FR"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="catalogue-no-special"]').exists()).toBe(false)
    expect(cards(wrapper)).toEqual(['Notre-Dame'])
  })

  it("shows Portugal's escudos on their own, with their face values in escudos", async () => {
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    await wrapper.get('[data-testid="catalogue-group-other"]').trigger('click')
    await wrapper.get('[data-testid="catalogue-country-PT-PTE"]').trigger('click')
    await flushPromises()
    expect(mocks.getCoinCatalogue).toHaveBeenLastCalledWith('PT', 'PTE')
    expect(mocks.replace).toHaveBeenLastCalledWith({ query: { country: 'PT', currency: 'PTE' } })
    expect(wrapper.get('#catalogue-country-title').text()).toBe('🇵🇹Portugal · Escudos')
    expect(cards(wrapper)).toEqual(['Colombo e Portugal', 'Discovery of Azores'])
    expect(wrapper.get('[data-testid="catalogue-value-filter"]').findAll('option').map((option) => text(option))).toEqual(['Valor: Todos', 'Valor: 100 escudos', 'Valor: 200 escudos'])
    await wrapper.get('[data-testid="catalogue-value-filter"]').setValue('100')
    expect(cards(wrapper)).toEqual(['Discovery of Azores'])
    expect(wrapper.find('[data-testid="release-no-image"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="catalogue-partial"]').exists()).toBe(false)
  })

  it('opens the country and currency in the address, its list beside, and the phone list chooses the same way', async () => {
    mocks.route.query = { country: 'pt', currency: 'pte' }
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()
    expect(mocks.getCoinCatalogue).toHaveBeenCalledWith('PT', 'PTE')
    // The escudos' list is the one beside.
    expect(wrapper.get('[data-testid="catalogue-group-other"]').attributes('aria-pressed')).toBe('true')

    const select = wrapper.get('[data-testid="catalogue-country-select"]')
    expect(select.findAll('optgroup').map((group) => group.attributes('label'))).toEqual(['Moedas do euro', 'Escudos'])
    expect(select.findAll('option').map((option) => option.text())).toEqual(['🇩🇪 Alemanha — 0 de 30', '🇫🇷 França — 1 de 70', '🇵🇹 Portugal — 3 de 5', '🇵🇹 Portugal · Escudos — 1 de 46'])
    await select.setValue('FR')
    await flushPromises()
    expect(mocks.getCoinCatalogue).toHaveBeenLastCalledWith('FR', undefined)
    expect(cards(wrapper)).toEqual(['Notre-Dame'])
    expect(wrapper.get('[data-testid="catalogue-group-euro"]').attributes('aria-pressed')).toBe('true')
    // One face value and one year: no filters of them.
    expect(wrapper.find('[data-testid="catalogue-value-filter"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="catalogue-year-filter"]').exists()).toBe(false)
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

    expect(wrapper.get('[data-testid="catalogue-create-set"]').text()).toBe('Adicionar conjunto')
    await wrapper.get('[data-testid="catalogue-value-filter"]').setValue('10')
    await wrapper.get('[data-testid="catalogue-create-set"]').trigger('click')
    await flushPromises()
    expect(mocks.createSet).toHaveBeenLastCalledWith(expect.objectContaining({
      // Intl puts a no-break space before the euro sign.
      name: expect.stringMatching(/^Portugal · 10\s€$/), checklist: { source: 'catalogue', country: 'PT', faceValue: 10 },
    }))

    // The escudos: their currency goes with them.
    await wrapper.get('[data-testid="catalogue-group-other"]').trigger('click')
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

  it('marks the collection\'s coins, offers to add the others, and saves one to the wishlist with its heart', async () => {
    mocks.saveToWishlist.mockResolvedValue({ result: 'saved', coinId: 5 })
    const wrapper = mount(CommemorativeCataloguePage)
    await flushPromises()

    const card = (title: string) => wrapper.findAll('[data-testid="release"]').find((node) => node.find('.line-clamp-3').text() === title)!
    // Had: in the collection, nothing to add or save, and its card in gold.
    expect(card('Rotary').get('[data-testid="release-owned"]').text()).toBe('Na coleção')
    expect(card('Rotary').classes()).toContain('border-gold')
    expect(card('Rotary').find('[data-testid="release-add"]').exists()).toBe(false)
    expect(card('Rotary').find('[data-testid="release-wishlist"]').exists()).toBe(false)
    // On the wishlist already: its heart is full.
    const wished = card('Sustainable development').get('[data-testid="release-wishlist"]')
    expect(wished.attributes('aria-pressed')).toBe('true')
    expect(wished.attributes('disabled')).toBeDefined()
    // The others: a button to add it, and a heart.
    expect(card('Faróis').get('[data-testid="release-add"]').text()).toBe('Adicionar')
    const faróis = card('Faróis').get('[data-testid="release-wishlist"]')
    expect(faróis.text()).toBe('')
    expect(faróis.attributes('aria-label')).toBe('Guardar na lista de desejos')

    await faróis.trigger('click')
    await flushPromises()
    expect(mocks.saveToWishlist).toHaveBeenCalledWith(expect.objectContaining({ title: 'Faróis' }))
    expect(card('Faróis').get('[data-testid="release-wishlist"]').attributes('aria-pressed')).toBe('true')
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
