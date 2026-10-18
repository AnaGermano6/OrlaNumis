import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CoinReleases from '../CoinReleases.vue'
import { setLocale } from '@/i18n'
import type { CoinRelease } from '@/api/client'
import { releaseAddRoute } from '@/utils/releasePrefill'

const mocks = vi.hoisted(() => ({ getCoinReleases: vi.fn(), saveToWishlist: vi.fn(), removeFromWishlist: vi.fn(), showToast: vi.fn() }))
vi.mock('@/api/client', () => ({ getCoinReleases: mocks.getCoinReleases }))
vi.mock('@/composables/useReleaseWishlist', () => ({ useReleaseWishlist: () => ({ saveToWishlist: mocks.saveToWishlist, removeFromWishlist: mocks.removeFromWishlist }) }))
vi.mock('@/composables/useToast', () => ({ useToast: () => ({ showToast: mocks.showToast }) }))

// Text with the non-breaking spaces of formatted numbers as plain spaces.
function text(element: { text: () => string }): string {
  return element.text().replace(/\s/g, ' ')
}

function release(fields: Partial<CoinRelease>): CoinRelease {
  return { source: 'ecb', kind: 'commemorative', country: 'PT', title: 'Feature', faceValue: 2, imageUrl: 'https://www.ecb.europa.eu/x.jpg', pageUrl: 'https://www.ecb.europa.eu/p', from: '2025-06', to: '2025-06', upcoming: false, owned: false, ...fields }
}

const GROUPS = [
  { source: 'ecb', year: 2025, releases: [
    release({ title: 'Lithuanian Minor', country: 'LT', from: '2025-10', to: '2025-12', upcoming: true, imageUrl: 'https://www.ecb.europa.eu/lt.jpg', volume: 500000 }),
    release({ title: 'World scouting', country: 'PT', owned: true, imageUrl: 'https://www.ecb.europa.eu/pt.jpg' }),
    release({ title: '35 years of the Erasmus programme', country: '', from: '2025-01', to: '2025-12', imageUrl: 'https://www.ecb.europa.eu/eu.jpg' }),
  ] },
  { source: 'mints', year: 2026, releases: [
    release({ source: 'mdp', country: 'FR', title: 'Year of the goat', faceValue: 10, faceValues: [10, 200], from: '2027-01', to: '2027-12', upcoming: true, imageUrl: 'https://mdp/goat.jpg' }),
    release({ source: 'incm', title: 'Rotary', faceValue: 2, variants: ['BNC', 'Proof'], from: '2026-07', to: '2026-07', imageUrl: 'https://cdn.shopify.com/r.jpg', pageUrl: 'https://loja.incm.pt/products/rotary' }),
    release({ source: 'incm', title: 'FIFA World Cup', faceValue: 7.5, from: '2026-05', to: '2026-05', imageUrl: 'https://cdn.shopify.com/f.jpg' }),
  ] },
]

describe('CoinReleases', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    mocks.getCoinReleases.mockReset()
    mocks.getCoinReleases.mockResolvedValue({ data: { groups: GROUPS } })
  })

  it("shows the ECB's latest list, what is to come and what the collection has", async () => {
    const wrapper = mount(CoinReleases, { global: { stubs: { RouterLink: RouterLinkStub } } })
    await flushPromises()

    expect(wrapper.get('[data-testid="releases-tab-ecb"]').text()).toBe('2 € comemorativas 2025')
    const cards = wrapper.findAll('[data-testid="release"]')
    expect(cards).toHaveLength(3)
    expect(text(cards[0]!)).toContain('Lituânia · 2 €')
    expect(text(cards[0]!)).toContain('outubro')
    expect(text(cards[0]!)).toContain('dezembro de 2025')
    expect(text(cards[0]!)).toContain('Tiragem: 500 000')
    expect(cards[0]!.find('[data-testid="release-upcoming"]').exists()).toBe(true)
    expect(text(cards[1]!)).toContain('junho de 2025')
    expect(cards[1]!.find('[data-testid="release-owned"]').text()).toBe('Na coleção')
    // A joint issue of the whole year.
    expect(text(cards[2]!)).toContain('Área do euro')
    expect(text(cards[2]!)).toContain('2025')
    // The list is last year's while the ECB has not published this year's.
    expect(wrapper.get('[data-testid="releases-note"]').text()).toContain('2025')
    expect(cards[0]!.get('a').attributes()).toMatchObject({ href: 'https://www.ecb.europa.eu/p', target: '_blank', rel: 'noopener noreferrer' })
  })

  it('offers to add each coin to the collection, and to save it to the wishlist at once', async () => {
    const wrapper = mount(CoinReleases, { global: { stubs: { RouterLink: RouterLinkStub } } })
    await flushPromises()

    const cards = wrapper.findAll('[data-testid="release"]')
    // Still to come: the wishlist only, with its words.
    expect(cards[0]!.find('[data-testid="release-add"]').exists()).toBe(false)
    expect(cards[0]!.get('[data-testid="release-wishlist"]').text()).toBe('Lista de desejos')
    expect(cards[0]!.get('[data-testid="release-wishlist"]').attributes('aria-label')).toBe('Guardar na lista de desejos')
    // Had: marked, with neither button.
    expect(cards[1]!.get('[data-testid="release-owned"]').text()).toBe('Na coleção')
    expect(cards[1]!.find('[data-testid="release-add"]').exists()).toBe(false)
    expect(cards[1]!.find('[data-testid="release-wishlist"]').exists()).toBe(false)
    // Neither: both, the wishlist as a heart.
    expect(cards[2]!.get('[data-testid="release-add"]').attributes('aria-label')).toBe('Adicionar à coleção')
    expect(cards[2]!.getComponent(RouterLinkStub).props('to')).toEqual(releaseAddRoute(GROUPS[0]!.releases[2]! as CoinRelease))
    const heart = cards[2]!.get('[data-testid="release-wishlist"]')
    expect(heart.text()).toBe('')
    expect(heart.attributes('aria-label')).toBe('Guardar na lista de desejos')
  })

  it('saves a coin to the wishlist and marks it, or says it could not', async () => {
    mocks.saveToWishlist.mockResolvedValueOnce({ result: 'saved', coinId: 31 }).mockRejectedValueOnce(new Error('offline'))
    const wrapper = mount(CoinReleases, { global: { stubs: { RouterLink: RouterLinkStub } } })
    await flushPromises()

    const card = () => wrapper.findAll('[data-testid="release"]')[2]!
    await card().get('[data-testid="release-wishlist"]').trigger('click')
    await flushPromises()
    expect(mocks.saveToWishlist).toHaveBeenCalledWith(GROUPS[0]!.releases[2])
    expect(mocks.showToast).toHaveBeenCalledWith('Guardada na lista de desejos: 35 years of the Erasmus programme', 'success')
    expect(card().get('[data-testid="release-on-wishlist"]').text()).toBe('Na lista de desejos')
    // A second press takes it off again.
    expect(card().get('[data-testid="release-wishlist"]').attributes()).toMatchObject({ 'aria-pressed': 'true', 'aria-label': 'Retirar da lista de desejos' })
    expect(card().get('[data-testid="release-wishlist"]').attributes('disabled')).toBeUndefined()

    await wrapper.findAll('[data-testid="release"]')[0]!.get('[data-testid="release-wishlist"]').trigger('click')
    await flushPromises()
    expect(mocks.showToast).toHaveBeenLastCalledWith('Não foi possível guardar na lista de desejos. Tente novamente.', 'error')
    expect(wrapper.findAll('[data-testid="release"]')[0]!.find('[data-testid="release-on-wishlist"]').exists()).toBe(false)
  })

  it('takes a coin off the wishlist with a second press, or says it could not', async () => {
    const onWishlist = { ...GROUPS[0]!.releases[2]!, wishlist: true, wishlistCoinId: 44 }
    mocks.getCoinReleases.mockResolvedValue({ data: { groups: [{ ...GROUPS[0]!, releases: [onWishlist] }] } })
    mocks.saveToWishlist.mockClear()
    mocks.removeFromWishlist.mockRejectedValueOnce(new Error('offline')).mockResolvedValueOnce(undefined)
    const wrapper = mount(CoinReleases, { global: { stubs: { RouterLink: RouterLinkStub } } })
    await flushPromises()

    const heart = () => wrapper.get('[data-testid="release-wishlist"]')
    expect(wrapper.get('[data-testid="release-on-wishlist"]').text()).toBe('Na lista de desejos')
    expect(heart().attributes('aria-label')).toBe('Retirar da lista de desejos')

    await heart().trigger('click')
    await flushPromises()
    expect(mocks.removeFromWishlist).toHaveBeenCalledWith(44)
    expect(mocks.showToast).toHaveBeenLastCalledWith('Não foi possível retirar da lista de desejos. Tente novamente.', 'error')
    expect(wrapper.find('[data-testid="release-on-wishlist"]').exists()).toBe(true)

    await heart().trigger('click')
    await flushPromises()
    expect(mocks.showToast).toHaveBeenLastCalledWith('Retirada da lista de desejos: 35 years of the Erasmus programme', 'success')
    expect(wrapper.find('[data-testid="release-on-wishlist"]').exists()).toBe(false)
    expect(heart().attributes()).toMatchObject({ 'aria-pressed': 'false', 'aria-label': 'Guardar na lista de desejos' })
    expect(mocks.saveToWishlist).not.toHaveBeenCalled()
  })

  it('still takes a coin the collection has off the wishlist', async () => {
    const both = { ...GROUPS[0]!.releases[2]!, owned: true, wishlist: true, wishlistCoinId: 45 }
    mocks.getCoinReleases.mockResolvedValue({ data: { groups: [{ ...GROUPS[0]!, releases: [both] }] } })
    mocks.removeFromWishlist.mockResolvedValueOnce(undefined)
    const wrapper = mount(CoinReleases, { global: { stubs: { RouterLink: RouterLinkStub } } })
    await flushPromises()

    expect(wrapper.find('[data-testid="release-add"]').exists()).toBe(false)
    const heart = wrapper.get('[data-testid="release-wishlist"]')
    expect(heart.attributes('aria-label')).toBe('Retirar da lista de desejos')
    await heart.trigger('click')
    await flushPromises()
    expect(mocks.removeFromWishlist).toHaveBeenCalledWith(45)
    // Off the wishlist, a coin the collection has offers nothing more.
    expect(wrapper.find('[data-testid="release-wishlist"]').exists()).toBe(false)
  })

  it("filters what is to come and what is missing, and shows the mints' list", async () => {
    const wrapper = mount(CoinReleases, { global: { stubs: { RouterLink: RouterLinkStub } } })
    await flushPromises()

    await wrapper.get('[data-testid="releases-filter-upcoming"]').trigger('click')
    expect(wrapper.findAll('[data-testid="release"]').map((card) => card.find('.line-clamp-3').text())).toEqual(['Lithuanian Minor'])
    await wrapper.get('[data-testid="releases-filter-missing"]').trigger('click')
    expect(wrapper.findAll('[data-testid="release"]')).toHaveLength(2)

    await wrapper.get('[data-testid="releases-filter-all"]').trigger('click')
    await wrapper.get('[data-testid="releases-tab-mints"]').trigger('click')
    expect(wrapper.get('[data-testid="releases-tab-mints"]').text()).toBe('Moedas de coleção')
    const cards = wrapper.findAll('[data-testid="release"]')
    expect(cards).toHaveLength(3)
    // A coin dated by its year alone.
    // One design in several face values.
    expect(text(cards[0]!)).toContain('França · 10 € · 200 €')
    expect(text(cards[0]!)).toContain('2027')
    expect(text(cards[1]!)).toContain('Rotary')
    expect(text(cards[1]!)).toContain('BNC · Proof')
    expect(text(cards[1]!)).toContain('julho de 2026')
    expect(text(cards[2]!)).toContain('7,50 €')
    expect(wrapper.get('[data-testid="releases-note"]').text()).toContain('só com o ano')
  })

  it('filters each list by country, and starts again with every country on another list', async () => {
    const wrapper = mount(CoinReleases, { global: { stubs: { RouterLink: RouterLinkStub } } })
    await flushPromises()

    const select = wrapper.get('[data-testid="releases-country"]')
    // The ECB list's countries by name, the joint issue as the euro area.
    expect(select.findAll('option').map((option) => text(option))).toEqual(['Todos os países (3)', 'Área do euro (1)', 'Lituânia (1)', 'Portugal (1)'])
    await select.setValue('PT')
    expect(wrapper.findAll('[data-testid="release"]').map((card) => card.find('.line-clamp-3').text())).toEqual(['World scouting'])
    await select.setValue('EU')
    expect(wrapper.findAll('[data-testid="release"]').map((card) => card.find('.line-clamp-3').text())).toEqual(['35 years of the Erasmus programme'])

    await wrapper.get('[data-testid="releases-tab-mints"]').trigger('click')
    const mints = wrapper.get('[data-testid="releases-country"]')
    expect((mints.element as HTMLSelectElement).value).toBe('')
    expect(mints.findAll('option').map((option) => text(option))).toEqual(['Todos os países (3)', 'França (1)', 'Portugal (2)'])
    await mints.setValue('FR')
    expect(wrapper.findAll('[data-testid="release"]').map((card) => card.find('.line-clamp-3').text())).toEqual(['Year of the goat'])
    await mints.setValue('PT')
    expect(wrapper.findAll('[data-testid="release"]')).toHaveLength(2)
  })

  it('says so when the lists cannot be loaded', async () => {
    mocks.getCoinReleases.mockRejectedValue(new Error('offline'))
    const wrapper = mount(CoinReleases, { global: { stubs: { RouterLink: RouterLinkStub } } })
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Não foi possível carregar os lançamentos.')
  })
})
