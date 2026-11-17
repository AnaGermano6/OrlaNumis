import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CoinSetsSection from '../CoinSetsSection.vue'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'

const mocks = vi.hoisted(() => ({ getSets: vi.fn(), createSet: vi.fn(), getCoinRecommendations: vi.fn() }))
vi.mock('@/api/client', () => ({
  getSets: mocks.getSets,
  createSet: mocks.createSet,
  getCoinRecommendations: mocks.getCoinRecommendations,
  addCoinToSet: vi.fn(),
  removeCoinFromSet: vi.fn(),
  acceptCoinRecommendation: vi.fn(),
  rejectCoinRecommendation: vi.fn(),
}))

const gulbenkian = { id: 1830, name: '2 euro 2026', kind: 'coin', country: 'Portugal', monetarySystem: 'Euro', denomination: '2 euro', year: 2026, isWishlist: false, images: [] } as unknown as Coin

describe('CoinSetsSection', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    vi.clearAllMocks()
    mocks.getCoinRecommendations.mockResolvedValue({ data: { recommendations: [] } })
    mocks.createSet.mockResolvedValue({ data: { id: 9 } })
  })

  it('proposes the sets the coin could start: makes one, opens one made before', async () => {
    mocks.getSets.mockResolvedValueOnce({ data: { sets: [{ id: 4, name: 'Moedas de 2026', color: '#fff', setType: 'smart' }] } })
      .mockResolvedValue({ data: { sets: [{ id: 4, name: 'Moedas de 2026', color: '#fff', setType: 'smart' }, { id: 9, name: 'Portugal · 2 €', color: '#fff', setType: 'checklist' }] } })
    const wrapper = mount(CoinSetsSection, {
      props: { sets: [], coinId: 1830, coin: gulbenkian, status: { current: true, commemorative: true } },
      global: { stubs: { RouterLink: RouterLinkStub } },
    })
    await flushPromises()

    const proposals = wrapper.get('[data-testid="set-proposals"]')
    expect(proposals.get('h4').text()).toBe('Criar um conjunto para esta moeda')
    // Made before: opened, not made again.
    const year = wrapper.get('[data-testid="set-proposal-year"]')
    expect(year.findComponent(RouterLinkStub).props('to')).toEqual({ name: 'set-detail', params: { id: 4 } })

    await wrapper.get('[data-testid="set-proposal-album"] [data-testid="set-proposal-create"]').trigger('click')
    await flushPromises()
    expect(mocks.createSet).toHaveBeenCalledWith(expect.objectContaining({
      name: 'Portugal · 2 €', setType: 'checklist', checklist: { source: 'catalogue', country: 'PT', faceValue: 2 }, color: expect.stringMatching(/^#/),
    }))
    // Made now: it opens.
    expect(wrapper.get('[data-testid="set-proposal-album"]').findComponent(RouterLinkStub).props('to')).toEqual({ name: 'set-detail', params: { id: 9 } })
  })

  it('suggests the sets of the collection the coin would fit only when there are some', async () => {
    mocks.getSets.mockResolvedValue({ data: { sets: [] } })
    const wrapper = mount(CoinSetsSection, {
      props: { sets: [], coinId: 1830, coin: gulbenkian, status: { current: true, commemorative: true } },
      global: { stubs: { RouterLink: RouterLinkStub } },
    })
    await flushPromises()
    expect(wrapper.find('[data-testid="set-suggestions"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('Conjuntos sugeridos')

    mocks.getCoinRecommendations.mockResolvedValue({ data: { recommendations: [{ id: 3, targetType: 'set', targetId: 4, targetName: 'Notas de Portugal', score: 80, confidence: 'high', reasons: ['Mesmo país'], status: 'pending' }] } })
    const suggested = mount(CoinSetsSection, {
      props: { sets: [], coinId: 1830, coin: gulbenkian, status: { current: true, commemorative: true } },
      global: { stubs: { RouterLink: RouterLinkStub } },
    })
    await flushPromises()
    expect(suggested.get('[data-testid="set-suggestions"]').text()).toContain('Notas de Portugal')
  })

  it('says so when a set cannot be made', async () => {
    mocks.getSets.mockResolvedValue({ data: { sets: [] } })
    mocks.createSet.mockRejectedValue(new Error('offline'))
    const wrapper = mount(CoinSetsSection, {
      props: { sets: [], coinId: 1830, coin: gulbenkian, status: { current: true, commemorative: true } },
      global: { stubs: { RouterLink: RouterLinkStub } },
    })
    await flushPromises()
    // Three at most: the album, the country's commemoratives and the year.
    expect(wrapper.findAll('[data-testid^="set-proposal-"][data-testid$="album"], [data-testid="set-proposal-commemorative"], [data-testid="set-proposal-year"]')).toHaveLength(3)
    expect(wrapper.find('[data-testid="set-proposal-country"]').exists()).toBe(false)
    await wrapper.get('[data-testid="set-proposal-commemorative"] [data-testid="set-proposal-create"]').trigger('click')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Não foi possível criar o conjunto. Tente novamente.')
  })
})
