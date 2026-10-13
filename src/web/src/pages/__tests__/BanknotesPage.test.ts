import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import BanknotesPage from '../BanknotesPage.vue'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'

const mocks = vi.hoisted(() => ({
  getCoins: vi.fn(),
  getCoinCountries: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  getCoins: mocks.getCoins,
  getCoinCountries: mocks.getCoinCountries,
}))
vi.mock('@/components/CoinCard.vue', () => ({
  default: defineComponent({
    props: { coin: { type: Object, required: true } },
    setup: (props) => () => h('article', { 'data-testid': 'card' }, (props.coin as Coin).name),
  }),
}))
vi.mock('@/components/collection/CoinListTable.vue', () => ({
  default: defineComponent({
    props: { coins: { type: Array, required: true } },
    setup: (props) => () => h('table', { 'data-testid': 'list' }, `${props.coins.length}`),
  }),
}))

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots }) => () => h('a', { href: props.to }, slots.default?.()),
})

function banknote(id: number, name: string): Coin {
  return { id, name, kind: 'banknote', country: 'Portugal', images: [] } as unknown as Coin
}

function mountPage() {
  return mount(BanknotesPage, { global: { stubs: { RouterLink: RouterLinkStub, 'router-link': RouterLinkStub } } })
}

describe('BanknotesPage', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    localStorage.clear()
    mocks.getCoins.mockResolvedValue({ data: { coins: [banknote(1, '1000 escudos'), banknote(2, '500 escudos')], total: 2 } })
    mocks.getCoinCountries.mockResolvedValue({ data: [{ country: 'Portugal', count: 2 }] })
  })

  afterEach(() => {
    vi.clearAllMocks()
    vi.useRealTimers()
  })

  it('lists only banknotes of the collection, without the numbers of the Collection page', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(mocks.getCoins).toHaveBeenCalledWith(expect.objectContaining({ kind: 'banknote', wishlist: 'false', sold: 'false', page: 1, sort: 'updated_at', order: 'desc' }))
    expect(mocks.getCoinCountries).toHaveBeenCalledWith('banknote')
    expect(wrapper.findAll('[data-testid="card"]').map((card) => card.text())).toEqual(['1000 escudos', '500 escudos'])
    expect(wrapper.find('[data-testid="banknotes-kpi-banknotes"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="banknotes-add"]').attributes('href')).toBe('/banknotes/add')
  })

  it('filters by country and search, and remembers the list layout', async () => {
    vi.useFakeTimers()
    const wrapper = mountPage()
    await flushPromises()

    await wrapper.get('[data-testid="banknotes-country"]').setValue('Portugal')
    await flushPromises()
    expect(mocks.getCoins).toHaveBeenLastCalledWith(expect.objectContaining({ kind: 'banknote', country: 'Portugal' }))

    await wrapper.get('input[type="text"]').setValue('escudos')
    vi.advanceTimersByTime(300)
    await flushPromises()
    expect(mocks.getCoins).toHaveBeenLastCalledWith(expect.objectContaining({ search: 'escudos', country: 'Portugal' }))

    const listButton = wrapper.findAll('button').find((button) => button.text() === 'Lista')!
    await listButton.trigger('click')
    expect(wrapper.find('[data-testid="list"]').exists()).toBe(true)
    expect(localStorage.getItem('banknotes:layout')).toBe('list')
  })

  it('invites to add the first banknote when there is none', async () => {
    mocks.getCoins.mockResolvedValue({ data: { coins: [], total: 0 } })
    const wrapper = mountPage()
    await flushPromises()

    const empty = wrapper.get('[data-testid="banknotes-empty"]')
    expect(empty.text()).toContain('Ainda não tem notas de banco')
    expect(empty.get('a').attributes('href')).toBe('/banknotes/add')
  })
})
