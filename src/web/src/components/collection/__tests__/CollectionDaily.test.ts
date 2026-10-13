import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import CollectionDaily from '../CollectionDaily.vue'
import { setLocale } from '@/i18n'
import { setCollectionCurrency } from '@/utils/collectionCurrency'
import type { Coin } from '@/types'

const mocks = vi.hoisted(() => ({ getDailyCoin: vi.fn() }))
vi.mock('@/api/client', () => ({ getDailyCoin: mocks.getDailyCoin }))

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots }) => () => h('a', { href: props.to }, slots.default?.()),
})

function render(kind: 'coin' | 'banknote') {
  return mount(CollectionDaily, { props: { kind }, global: { stubs: { RouterLink: RouterLinkStub, AuthenticatedImage: true } } })
}

describe('CollectionDaily', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    setCollectionCurrency('EUR')
    mocks.getDailyCoin.mockReset()
    vi.useFakeTimers({ toFake: ['Date'] })
    vi.setSystemTime(new Date(2026, 9, 13, 23, 30))
  })
  afterEach(() => {
    vi.useRealTimers()
    setCollectionCurrency('USD')
  })

  it("shows the viewer's day's coin with its photo, details and a link", async () => {
    mocks.getDailyCoin.mockResolvedValue({ data: { date: '2026-10-13', coin: {
      id: 7, name: '2 euro 2026 – Rotary', country: 'Portugal', year: 2026, grade: 'UNC', currentValue: 3.5,
      obverseDescription: 'A roda do Rotary.', images: [{ filePath: 'uploads/r.jpg', imageType: 'obverse', isPrimary: true }],
    } as unknown as Coin } })
    const wrapper = render('coin')
    await flushPromises()

    // The local date, even late at night.
    expect(mocks.getDailyCoin).toHaveBeenCalledWith('coin', '2026-10-13')
    expect(wrapper.text()).toContain('Moeda do dia')
    expect(wrapper.text().replace(/\s/g, ' ')).toContain('Portugal · 2026 · UNC · 3,50 €')
    expect(wrapper.text()).toContain('A roda do Rotary.')
    expect(wrapper.findAll('a').every((link) => link.attributes('href') === '/coin/7')).toBe(true)
  })

  it('shows nothing when the collection has no banknote', async () => {
    mocks.getDailyCoin.mockResolvedValue({ data: { date: '2026-10-13', coin: null } })
    const wrapper = render('banknote')
    await flushPromises()

    expect(mocks.getDailyCoin).toHaveBeenCalledWith('banknote', '2026-10-13')
    expect(wrapper.find('article').exists()).toBe(false)
  })
})
