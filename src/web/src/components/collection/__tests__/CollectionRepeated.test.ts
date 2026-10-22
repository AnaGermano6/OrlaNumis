import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CollectionRepeated from '../CollectionRepeated.vue'
import { setLocale } from '@/i18n'
import type { CollectionSummary } from '@/api/client'
import type { Coin } from '@/types'

const mocks = vi.hoisted(() => ({ getCoins: vi.fn() }))
vi.mock('@/api/client', () => ({ getCoins: mocks.getCoins }))

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots }) => () => h('a', { href: props.to }, slots.default?.()),
})

function coin(id: number, name: string, quantity: number): Coin {
  return { id, name, quantity, country: 'Portugal', year: 1962, images: [] } as unknown as Coin
}

function render(kind: 'coin' | 'banknote', summary: Partial<CollectionSummary> | null) {
  return mount(CollectionRepeated, {
    props: { kind, summary: summary as CollectionSummary | null },
    global: { stubs: { RouterLink: RouterLinkStub, AuthenticatedImage: true } },
  })
}

describe('CollectionRepeated', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    mocks.getCoins.mockReset()
  })

  it('lists the coins held more than once, most copies first, with the copies to trade', async () => {
    mocks.getCoins.mockResolvedValue({ data: { coins: [coin(1, '1 escudo 1962', 3), coin(2, '50 centavos 1965', 2)], total: 2 } })
    const wrapper = render('coin', { repeatedCoins: 2, spareCoins: 3 })
    await flushPromises()

    expect(mocks.getCoins).toHaveBeenCalledWith(expect.objectContaining({ kind: 'coin', repeated: 'true', wishlist: 'false', sort: 'quantity', order: 'desc' }))
    expect(wrapper.get('[data-testid="repeated-counts"]').text()).toBe('2 moedas repetidas · 3 exemplares para troca')
    const items = wrapper.findAll('[data-testid="repeated-item"]')
    expect(items.map((item) => item.get('a').text())).toEqual(['1 escudo 1962', '50 centavos 1965'])
    expect(items[0]!.text()).toContain('× 3')
    expect(items[0]!.text()).toContain('2 para troca')
    expect(items[0]!.get('a').attributes('href')).toBe('/coin/1')
  })

  it('shows six at first and the rest on request', async () => {
    mocks.getCoins.mockResolvedValue({ data: { coins: Array.from({ length: 8 }, (_, i) => coin(i + 1, `Moeda ${i + 1}`, 2)), total: 8 } })
    const wrapper = render('coin', { repeatedCoins: 8, spareCoins: 8 })
    await flushPromises()

    expect(wrapper.findAll('[data-testid="repeated-item"]')).toHaveLength(6)
    await wrapper.get('[data-testid="repeated-toggle"]').trigger('click')
    expect(wrapper.findAll('[data-testid="repeated-item"]')).toHaveLength(8)
    expect(wrapper.get('[data-testid="repeated-toggle"]').text()).toBe('Mostrar menos')
  })

  it('says how a banknote gets here when none is repeated', async () => {
    mocks.getCoins.mockResolvedValue({ data: { coins: [], total: 0 } })
    const wrapper = render('banknote', { repeatedCoins: 0, spareCoins: 0 })
    await flushPromises()

    expect(mocks.getCoins).toHaveBeenCalledWith(expect.objectContaining({ kind: 'banknote' }))
    expect(wrapper.get('[data-testid="repeated-empty"]').text()).toContain('Ainda não tem notas repetidas')
    expect(wrapper.find('[data-testid="repeated-counts"]').exists()).toBe(false)
  })

  it('opens the gallery with only these, from a button shown when there are some', async () => {
    mocks.getCoins.mockResolvedValue({ data: { coins: [coin(1, '1 euro 2009', 2)], total: 1 } })
    const wrapper = render('banknote', { repeatedCoins: 1, spareCoins: 1 })
    await flushPromises()

    const open = wrapper.get('[data-testid="repeated-open"]')
    expect(open.text()).toBe('Ver na galeria')
    await open.trigger('click')
    expect(wrapper.emitted('open')).toHaveLength(1)

    mocks.getCoins.mockResolvedValue({ data: { coins: [], total: 0 } })
    const none = render('coin', { repeatedCoins: 0, spareCoins: 0 })
    await flushPromises()
    expect(none.find('[data-testid="repeated-open"]').exists()).toBe(false)
  })

  it('offers to share the list of duplicates when there are some', async () => {
    mocks.getCoins.mockResolvedValue({ data: { coins: [coin(1, '1 euro 2009', 2)], total: 1 } })
    const wrapper = render('coin', { repeatedCoins: 1, spareCoins: 1 })
    await flushPromises()

    const share = wrapper.get('[data-testid="repeated-share"]')
    expect(share.text()).toBe('Partilhar lista')
    await share.trigger('click')
    expect(wrapper.emitted('share')).toHaveLength(1)

    mocks.getCoins.mockResolvedValue({ data: { coins: [], total: 0 } })
    const none = render('coin', { repeatedCoins: 0, spareCoins: 0 })
    await flushPromises()
    expect(none.find('[data-testid="repeated-share"]').exists()).toBe(false)
  })
})
