import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import WishlistCard from '../WishlistCard.vue'
import type { Coin } from '@/types'
import { setLocale } from '@/i18n'
import { formatCurrency } from '@/utils/format'

const push = vi.hoisted(() => vi.fn())
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))

function card(extra: Partial<Coin> = {}) {
  const coin = {
    id: 7, name: '7½ euro 2024 – Capitais Ibero-Americanas', kind: 'coin', country: 'Portugal', denomination: '7½ euro',
    referenceText: 'INCM', referenceUrl: 'https://loja.incm.pt/7-5-euro', currentValue: null, purchasePrice: null,
    listingStatus: '', images: [], ...extra,
  } as unknown as Coin
  return mount(WishlistCard, { props: { coin }, global: { stubs: { AuthenticatedImage: true, CountryFlag: { template: '<span class="flag" />' } } } })
}

describe('WishlistCard', () => {
  beforeEach(() => setLocale('pt-PT'))
  afterEach(() => {
    setLocale('en')
    vi.clearAllMocks()
  })

  it('shows its name, its country and where it is sold, and about its face value without a price saved', () => {
    const wrapper = card()

    expect(wrapper.get('h3').text()).toBe('7½ euro 2024 – Capitais Ibero-Americanas')
    const origin = wrapper.get('[data-testid="wishlist-card-origin"]')
    expect(origin.find('.flag').exists()).toBe(true)
    expect(origin.text().replace(/\s+/g, ' ')).toBe('Portugal · INCM')
    expect(origin.get('a').attributes('href')).toBe('https://loja.incm.pt/7-5-euro')
    const price = wrapper.get('[data-testid="wishlist-card-price"]')
    expect(price.text().replace(/\s/g, ' ')).toBe('~ 7,50 €')
    expect(price.attributes('title')).toContain('7½ euro')
  })

  it('shows the price saved, its value first, as it is, in the collection currency', () => {
    expect(card({ purchasePrice: 12 }).get('[data-testid="wishlist-card-price"]').text()).toBe(formatCurrency(12))
    expect(card({ purchasePrice: 12, currentValue: 15 }).get('[data-testid="wishlist-card-price"]').text()).toBe(formatCurrency(15))
  })

  it('shows no price for one in another currency, or a banknote, without a price saved', () => {
    expect(card({ denomination: '200 escudos' }).find('[data-testid="wishlist-card-price"]').exists()).toBe(false)
    expect(card({ kind: 'banknote', denomination: '5 euro' }).find('[data-testid="wishlist-card-price"]').exists()).toBe(false)
  })

  it('names the site it is sold on when the reference has no name', () => {
    expect(card({ referenceText: '', referenceUrl: 'https://www.numista.com/catalogue/x' }).get('[data-testid="wishlist-card-origin"]').text().replace(/\s+/g, ' ')).toBe('Portugal · numista.com')
  })

  it('marks it bought or takes it off the wishlist without opening it; clicking it elsewhere opens it', async () => {
    const wrapper = card()

    await wrapper.get('[data-testid="wishlist-card-purchase"]').trigger('click')
    expect(wrapper.get('[data-testid="wishlist-card-purchase"]').text()).toBe('Marcar como comprada')
    expect(wrapper.emitted('purchase')?.[0]?.[0]).toMatchObject({ id: 7 })
    const remove = wrapper.get('[data-testid="wishlist-card-remove"]')
    expect(remove.attributes('aria-label')).toBe('Remover da lista de desejos')
    await remove.trigger('click')
    expect(wrapper.emitted('remove')?.[0]?.[0]).toMatchObject({ id: 7 })
    expect(push).not.toHaveBeenCalled()

    await wrapper.get('h3').trigger('click')
    expect(push).toHaveBeenCalledWith('/coin/7')
  })

  it('says when it can no longer be bought, until the notice is dismissed', async () => {
    const wrapper = card({ listingStatus: 'unavailable' })

    expect(wrapper.text()).toContain('Indisponível')
    await wrapper.findAll('button').find((button) => button.text() === 'Dispensar')!.trigger('click')
    expect(wrapper.emitted('dismiss-status')).toEqual([[7]])
    expect(card({ listingStatus: 'available' }).get('[data-testid="wishlist-card-status"]').attributes('title')).toBe('Disponível')
  })
})
