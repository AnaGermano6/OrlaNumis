import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import CollectionRepeated from '../CollectionRepeated.vue'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'

const mocks = vi.hoisted(() => ({ getCoins: vi.fn(), updateCoinQuantity: vi.fn(), showToast: vi.fn(), writeText: vi.fn() }))
vi.mock('@/api/client', () => ({ getCoins: mocks.getCoins, updateCoinQuantity: mocks.updateCoinQuantity }))
vi.mock('@/composables/useToast', () => ({ useToast: () => ({ showToast: mocks.showToast }) }))

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots, attrs }) => () => h('a', { ...attrs, href: props.to }, slots.default?.()),
})

function coin(id: number, name: string, quantity: number, country = 'Portugal', year = 2022): Coin {
  return { id, name, quantity, country, year, images: [] } as unknown as Coin
}

const COINS = [coin(1, '1 euro 2009 – Eslováquia', 2, 'Eslováquia', 2009), coin(2, '2 euro 2022 – Erasmus', 3)]
const BANKNOTES = [coin(9, '1000 escudos', 2, 'Portugal', 1996)]

function render() {
  return mount(CollectionRepeated, { global: { stubs: { RouterLink: RouterLinkStub, AuthenticatedImage: true } } })
}

const titles = (wrapper: ReturnType<typeof render>) => wrapper.findAll('[data-testid="repeated-item"] a.line-clamp-2').map((link) => link.text())

describe('CollectionRepeated (the trade list)', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    vi.clearAllMocks()
    mocks.getCoins.mockImplementation(async ({ kind }: { kind: string }) => ({ data: { coins: (kind === 'coin' ? COINS : BANKNOTES).map((item) => ({ ...item })), total: 2 } }))
    Object.defineProperty(navigator, 'clipboard', { value: { writeText: mocks.writeText }, configurable: true })
  })
  afterEach(() => setLocale('en'))

  it('lists the coins held more than once, or the banknotes, each with how many and how many to trade', async () => {
    const wrapper = render()
    await flushPromises()

    expect(mocks.getCoins).toHaveBeenCalledWith(expect.objectContaining({ kind: 'coin', repeated: 'true', wishlist: 'false' }))
    expect(mocks.getCoins).toHaveBeenCalledWith(expect.objectContaining({ kind: 'banknote', repeated: 'true', wishlist: 'false' }))
    expect(wrapper.get('h2').text()).toBe('Lista de Trocas')
    expect(wrapper.findAll('[data-testid^="trade-kind-"]').map((tab) => tab.text())).toEqual(['Moedas (2)', 'Notas (1)'])
    expect(wrapper.get('[data-testid="trade-kind-coin"]').attributes('aria-pressed')).toBe('true')
    expect(titles(wrapper)).toEqual(['1 euro 2009 – Eslováquia', '2 euro 2022 – Erasmus'])
    const first = wrapper.get('[data-testid="repeated-item"]')
    expect(first.get('[data-testid="repeated-origin"]').text().replace(/\s+/g, ' ')).toContain('Eslováquia · 2009')
    expect(first.get('[data-testid="repeated-total"]').text()).toBe('Total: 2')
    expect(first.get('[data-testid="repeated-spare"]').text()).toBe('1 para troca')

    await wrapper.get('[data-testid="trade-kind-banknote"]').trigger('click')
    expect(titles(wrapper)).toEqual(['1000 escudos'])
  })

  it('finds a coin by its name, or the coins of a country', async () => {
    const wrapper = render()
    await flushPromises()

    const country = wrapper.get('[data-testid="trade-country"]')
    expect(country.findAll('option').map((option) => option.text())).toEqual(['Todos os países', '🇸🇰 Eslováquia', '🇵🇹 Portugal'])
    await country.setValue('Portugal')
    expect(titles(wrapper)).toEqual(['2 euro 2022 – Erasmus'])
    await country.setValue('')
    await wrapper.get('[data-testid="trade-search"] input').setValue('eslovaquia')
    expect(titles(wrapper)).toEqual(['1 euro 2009 – Eslováquia'])
    await wrapper.get('[data-testid="trade-search"] input').setValue('narnia')
    expect(wrapper.get('[data-testid="trade-no-match"]').text()).toBe('Nenhuma repetida corresponde à pesquisa.')
  })

  it('changes how many to trade by its quantity alone, down to none (it stays until opened again)', async () => {
    mocks.updateCoinQuantity.mockResolvedValue({ data: {} })
    const wrapper = render()
    await flushPromises()
    const first = () => wrapper.get('[data-testid="repeated-item"]')

    await first().get('[data-testid="repeated-more"]').trigger('click')
    await flushPromises()
    expect(mocks.updateCoinQuantity).toHaveBeenLastCalledWith(1, 3)
    expect(first().get('[data-testid="repeated-spare"]').text()).toBe('2 para troca')
    await first().get('[data-testid="repeated-fewer"]').trigger('click')
    await first().get('[data-testid="repeated-fewer"]').trigger('click')
    await flushPromises()
    expect(mocks.updateCoinQuantity).toHaveBeenLastCalledWith(1, 1)
    expect(first().get('[data-testid="repeated-spare"]').text()).toBe('0 para troca')
    expect(first().get('[data-testid="repeated-total"]').text()).toBe('Total: 1')
    expect(first().get('[data-testid="repeated-fewer"]').attributes('disabled')).toBeDefined()

    // Not saved: said, and the count stays.
    mocks.updateCoinQuantity.mockRejectedValueOnce(new Error('offline'))
    await first().get('[data-testid="repeated-more"]').trigger('click')
    await flushPromises()
    expect(mocks.showToast).toHaveBeenLastCalledWith('Não foi possível guardar a quantidade. Tente novamente.', 'error')
    expect(first().get('[data-testid="repeated-total"]').text()).toBe('Total: 1')
  })

  it('copies the list shown as text, and asks for the file to share', async () => {
    mocks.writeText.mockResolvedValue(undefined)
    const wrapper = render()
    await flushPromises()

    await wrapper.get('[data-testid="trade-copy"]').trigger('click')
    await flushPromises()
    expect(mocks.writeText).toHaveBeenCalledWith('Moedas repetidas para troca:\n• 1 euro 2009 – Eslováquia (Eslováquia, 2009): 1 para troca\n• 2 euro 2022 – Erasmus (Portugal, 2022): 2 para troca')
    expect(mocks.showToast).toHaveBeenLastCalledWith('Lista copiada.', 'success')

    await wrapper.get('[data-testid="trade-share"]').trigger('click')
    expect(wrapper.emitted('share')).toHaveLength(1)
  })

  it('says when there is none, and opens on the banknotes when only they are held twice', async () => {
    mocks.getCoins.mockImplementation(async ({ kind }: { kind: string }) => ({ data: { coins: kind === 'coin' ? [] : BANKNOTES, total: 0 } }))
    const wrapper = render()
    await flushPromises()
    expect(wrapper.get('[data-testid="trade-kind-banknote"]').attributes('aria-pressed')).toBe('true')
    await wrapper.get('[data-testid="trade-kind-coin"]').trigger('click')
    expect(wrapper.get('[data-testid="repeated-empty"]').text()).toContain('Ainda não tem moedas repetidas')
  })
})
