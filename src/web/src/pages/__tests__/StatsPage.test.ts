import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import StatsPage from '@/pages/StatsPage.vue'
import type { StatsResponse } from '@/types'

const stats: StatsResponse = {
  totalCoins: 42,
  totalWishlist: 7,
  byCategory: [{ category: 'Roman', count: 20 }, { category: 'Greek', count: 15 }],
  byMaterial: [],
  byGrade: [],
  byEra: [{ era: 'Republic', count: 10 }, { era: 'Empire', count: 32 }],
  byMonetarySystem: [{ monetarySystem: 'Escudo (1911-2001)', count: 30 }, { monetarySystem: 'Euro', count: 12 }],
  byRuler: [{ ruler: 'Augustus', count: 5 }],
  byPriceRange: [],
  values: {
    totalPurchasePrice: 1200,
    totalCurrentValue: 1500,
    avgPurchasePrice: 120,
    avgCurrentValue: 150,
  },
}

const store = vi.hoisted(() => ({
  stats: null as StatsResponse | null,
  fetchStats: vi.fn(),
}))

vi.mock('@/stores/coins', () => ({
  useCoinsStore: () => store,
}))

describe('StatsPage', () => {
  beforeEach(() => {
    document.body.innerHTML = `
      <div id="desktop-page-title"></div>
      <div id="desktop-page-actions"></div>
    `
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('renders summary metrics and distribution charts; excludes value detail block', () => {
    store.stats = stats

    const wrapper = mount(StatsPage, {
      global: {
        stubs: {
          PullToRefresh: { template: '<div><slot /></div>' },
          StatsHeatMap: { template: '<div class="stub-heatmap" />', methods: { fetchDistribution: () => {} } },
          StatsCoinFlowChart: { template: '<div class="stub-coin-flow" />' },
        },
      },
    })

    expect(wrapper.text()).toContain('Coins Owned')
    expect(wrapper.text()).toContain('On Wishlist')
    expect(wrapper.text()).toContain('42')
    expect(wrapper.text()).toContain('7')

    // The Stats page has no value detail block
    expect(wrapper.text()).not.toContain('Value Summary')
    expect(wrapper.text()).not.toContain('Total Invested')

    // Distribution charts are present
    expect(wrapper.text()).toContain('By Monetary System')
    expect(wrapper.text()).toContain('Escudo (1911-2001)')
    expect(wrapper.text()).toContain('By Coin Type')
    expect(wrapper.text()).toContain('By Era')
    expect(wrapper.text()).toContain('Top Rulers')
    expect(wrapper.find('.stub-heatmap').exists()).toBe(true)

    expect(wrapper.text()).not.toContain('Open Mint Map')
    expect(wrapper.findAll('a[href^="/stats/"]')).toHaveLength(0)
  })

  it('leaves out categories and eras with one value, and adds up the less common systems', () => {
    store.stats = {
      ...stats,
      byCategory: [{ category: 'Modern', count: 42 }],
      byEra: [{ era: 'modern', count: 42 }],
      byMonetarySystem: Array.from({ length: 14 }, (_, i) => ({ monetarySystem: `System ${i + 1}`, count: 20 - i })),
    }

    const wrapper = mount(StatsPage, {
      global: {
        stubs: {
          PullToRefresh: { template: '<div><slot /></div>' },
          StatsHeatMap: { template: '<div class="stub-heatmap" />', methods: { fetchDistribution: () => {} } },
          StatsCoinFlowChart: { template: '<div class="stub-coin-flow" />' },
        },
      },
    })

    expect(wrapper.text()).not.toContain('By Coin Type')
    expect(wrapper.text()).not.toContain('By Era')
    expect(wrapper.find('.stub-heatmap').exists()).toBe(false)
    const systems = wrapper.get('[data-testid="stats-monetary-systems"]').text()
    expect(systems).toContain('System 12')
    expect(systems).not.toContain('System 13')
    expect(systems).toContain('2 other systems')
    expect(systems).toContain('13')
  })

  it('renders stats from the normal stats store without querying quick-capture drafts', () => {
    store.stats = stats

    const wrapper = mount(StatsPage, {
      global: {
        stubs: {
          PullToRefresh: { template: '<div><slot /></div>' },
          StatsHeatMap: { template: '<div />', methods: { fetchDistribution: () => {} } },
          StatsCoinFlowChart: { template: '<div />' },
        },
      },
    })

    expect(store.fetchStats).toHaveBeenCalled()
    expect(wrapper.text()).not.toContain('Quick Capture Draft')
  })
})
