import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import TimelinePage from '../TimelinePage.vue'
import type { Coin } from '@/types'

const mocks = vi.hoisted(() => ({ getCoins: vi.fn() }))
vi.mock('@/api/client', () => mocks)

function item(fields: Partial<Coin>): Coin {
  return { id: 1, name: 'Denarius', category: 'Roman', purchaseDate: '2026-08-10', purchasePrice: 10, isWishlist: false, images: [], ...fields } as Coin
}

const ITEMS = [
  item({ id: 1, name: '2 euro 2025 – Erasmus', purchasePrice: 3 }),
  item({ id: 2, name: '500 escudos – Chapa 13', kind: 'banknote', purchaseDate: '2026-07-02', purchasePrice: 25 }),
  item({ id: 3, name: '5000 escudos – Antero', kind: 'banknote', purchaseDate: '2026-07-20', purchasePrice: 40 }),
  item({ id: 4, name: 'Wanted coin', isWishlist: true }),
]

function mountPage() {
  return mount(TimelinePage, {
    global: {
      stubs: {
        PullToRefresh: { template: '<div><slot /></div>' },
        DesktopPageContext: { template: '<div><slot /></div>' },
        AuthenticatedImage: true,
        RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' },
      },
    },
  })
}

describe('TimelinePage', () => {
  beforeEach(() => {
    mocks.getCoins.mockReset()
    mocks.getCoins.mockResolvedValue({ data: { coins: ITEMS } })
  })

  it('shows all the coins, or all the banknotes, never both', async () => {
    const wrapper = mountPage()
    await flushPromises()

    const select = wrapper.get<HTMLSelectElement>('[data-testid="timeline-kind"]')
    expect(select.findAll('option').map((option) => option.text())).toEqual(['All Coins', 'All Banknotes'])
    expect(wrapper.text()).toContain('2 euro 2025 – Erasmus')
    expect(wrapper.text()).not.toContain('Chapa 13')
    expect(wrapper.text()).not.toContain('Wanted coin')
    expect(wrapper.text()).toContain('Coins')
    expect(wrapper.text()).toContain('1 coin')

    await select.setValue('banknote')
    const text = wrapper.text()
    expect(text).not.toContain('Erasmus')
    // Newest first, in one month.
    expect(text.indexOf('Antero')).toBeLessThan(text.indexOf('Chapa 13'))
    expect(text).toContain('Banknotes')
    expect(text).toContain('2 banknotes')
    expect(text).toMatch(/65[.,]00/)
  })

  it('says so when no banknote has a purchase date', async () => {
    mocks.getCoins.mockResolvedValue({ data: { coins: [ITEMS[0]] } })
    const wrapper = mountPage()
    await flushPromises()

    await wrapper.get('[data-testid="timeline-kind"]').setValue('banknote')
    expect(wrapper.text()).toContain('Add purchase dates to your banknotes to see them on the timeline.')
  })
})
