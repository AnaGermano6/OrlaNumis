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

// The catalogue the progress and countries' cards read.
vi.mock('@/api/client', () => ({
  getCoinCatalogueCountries: vi.fn().mockResolvedValue({ data: { countries: [] } }),
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

  it('starts with the collection at a glance, then the distribution charts; no counts of its own', () => {
    store.stats = stats

    const wrapper = mount(StatsPage, {
      global: {
        stubs: {
          PullToRefresh: { template: '<div><slot /></div>' },
          StatsHeatMap: { template: '<div class="stub-heatmap" />', methods: { fetchDistribution: () => {} } },
          CollectionOverviewTabs: { template: '<section class="stub-overview" />', methods: { load: () => {} } },
          StatsProgress: { template: '<section class="stub-progress" />' },
        },
      },
    })

    // First the collection at a glance (its tabs), then the statistics.
    const overview = wrapper.find('.stub-overview')
    expect(overview.exists()).toBe(true)
    expect(overview.element.compareDocumentPosition(wrapper.get('[data-testid="stats-monetary-systems"]').element) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    // A line between the tabs (each ending with its pieces' condition) and the rest, the completion progress first.
    const divider = wrapper.get('[data-testid="stats-divider"]')
    expect(divider.element.tagName).toBe('HR')
    expect(overview.element.compareDocumentPosition(divider.element) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    expect(divider.element.nextElementSibling?.classList.contains('stub-progress')).toBe(true)
    // The coins owned and on the wishlist are in its tabs, not counted again here.
    expect(wrapper.text()).not.toContain('Coins Owned')
    expect(wrapper.text()).not.toContain('On Wishlist')

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
          CollectionOverviewTabs: { template: '<section class="stub-overview" />', methods: { load: () => {} } },
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
          CollectionOverviewTabs: { template: '<section />', methods: { load: () => {} } },
        },
      },
    })

    expect(store.fetchStats).toHaveBeenCalled()
    expect(wrapper.text()).not.toContain('Quick Capture Draft')
  })
})
