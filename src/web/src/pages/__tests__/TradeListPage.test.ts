import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import TradeListPage from '../TradeListPage.vue'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'

const mocks = vi.hoisted(() => ({ getCoins: vi.fn(), showToast: vi.fn(), writeText: vi.fn() }))
vi.mock('@/api/client', () => ({ getCoins: mocks.getCoins, updateCoinQuantity: vi.fn(), downloadShareList: vi.fn() }))
vi.mock('@/composables/useToast', () => ({ useToast: () => ({ showToast: mocks.showToast }) }))
vi.mock('@/composables/useLayout', async () => {
  const { ref } = await import('vue')
  return { useLayout: () => ({ isMobile: ref(false) }) }
})

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots, attrs }) => () => h('a', { ...attrs, href: props.to }, slots.default?.()),
})

function mountPage() {
  return mount(TradeListPage, { global: { stubs: { RouterLink: RouterLinkStub, AuthenticatedImage: true, Teleport: true } } })
}

describe('TradeListPage', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    vi.clearAllMocks()
    mocks.getCoins.mockImplementation(async ({ kind }: { kind: string }) => ({ data: {
      coins: kind === 'coin' ? [{ id: 1, name: '1 euro 2009 – Eslováquia', quantity: 2, country: 'Eslováquia', year: 2009, images: [] } as unknown as Coin] : [],
      total: 1,
    } }))
    Object.defineProperty(navigator, 'clipboard', { value: { writeText: mocks.writeText }, configurable: true })
  })

  it('has the title at the top, its buttons beside it, and the list under them', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(wrapper.get('[data-testid="page-context"] h1').text()).toBe('Lista de Trocas')
    // The list without a title of its own.
    expect(wrapper.get('[data-testid="trade-list"]').find('h2').exists()).toBe(false)
    expect(wrapper.findAll('[data-testid="repeated-item"]')).toHaveLength(1)
  })

  it('copies the list as text, and shares it as a file of the duplicates', async () => {
    mocks.writeText.mockResolvedValue(undefined)
    const wrapper = mountPage()
    await flushPromises()

    const copy = wrapper.get('[data-testid="page-context"] [data-testid="trade-copy"]')
    expect(copy.attributes('disabled')).toBeUndefined()
    await copy.trigger('click')
    await flushPromises()
    expect(mocks.writeText).toHaveBeenCalledWith(expect.stringContaining('1 euro 2009 – Eslováquia'))

    await wrapper.get('[data-testid="page-context"] [data-testid="trade-share"]').trigger('click')
    expect(wrapper.findComponent({ name: 'ShareListDialog' }).props()).toMatchObject({ open: true, initialList: 'trades' })
  })
})
