import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { defineComponent } from 'vue'
import CollectionContent from '../CollectionContent.vue'
import { browseNeighbours } from '@/utils/browseSequence'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'

const CoinCardStub = defineComponent({ name: 'CoinCard', props: { coin: { type: Object, required: true } }, template: '<div class="card-stub">{{ coin.name }}</div>' })

const COINS = [
  { id: 1, name: '50 Escudos', country: 'Portugal', year: 1968, grade: 'VF', rarity: 'scarce', presentation: 'coincard', currentValue: 23.81, images: [] },
  { id: 2, name: '2.50 Escudos', country: 'Angola', year: 1968, images: [] },
  { id: 3, name: '1 Escudo', country: 'Portugal', dateRange: 'c. 1930', images: [] },
] as unknown as Coin[]

function mountContent(props: Record<string, unknown>) {
  return mount(CollectionContent, {
    props: {
      loading: false, coins: COINS, selectMode: false, selectedCoinIds: new Set<number>(), selectedCount: 0,
      isPwa: false, viewMode: 'grid', gridSide: null, hasFilters: false, total: 3, page: 1, perPage: 50,
      ...props,
    },
    global: { stubs: { CoinCard: CoinCardStub, RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' }, AuthenticatedImage: true } },
  })
}

describe('CollectionContent grouping and list layout', () => {
  afterEach(() => setLocale('en'))

  it('shows one section per country with its coin count', () => {
    const wrapper = mountContent({ groupBy: 'country' })

    const headings = wrapper.findAll('h2').map((heading) => heading.text().replace(/\s+/g, ' '))
    expect(headings).toEqual(['🇦🇴 Angola 1 coin', '🇵🇹 Portugal 2 coins'])
    expect(wrapper.findAll('[data-testid="group-flag"]').map((flag) => flag.attributes('aria-hidden'))).toEqual(['true', 'true'])
    const sections = wrapper.findAll('section')
    expect(sections[1]!.findAll('.card-stub').map((card) => card.text())).toEqual(['50 Escudos', '1 Escudo'])
  })

  it("remembers the order it shows, for a coin's previous and next", () => {
    sessionStorage.clear()
    mountContent({ groupBy: 'country' })
    // Angola's coin, then Portugal's two.
    expect(browseNeighbours(2)).toEqual({ origin: 'gallery', position: 1, total: 3, previous: null, next: 1 })
    expect(browseNeighbours(3)).toEqual({ origin: 'gallery', position: 3, total: 3, previous: 1, next: null })
  })

  it('lists coins in a table with country, year, rarity, presentation and value', () => {
    setLocale('pt-PT')
    const wrapper = mountContent({ layout: 'list' })

    expect(wrapper.findAll('h2')).toHaveLength(0)
    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(3)
    const cells = rows[0]!.findAll('td').map((cell) => cell.text())
    expect(cells).toEqual(expect.arrayContaining(['🇵🇹Portugal', '1968', 'VF', 'Escassa', 'Coincard']))
    expect(rows[0]!.get('[data-testid="country-flag"]').attributes('aria-hidden')).toBe('true')
    expect(rows[0]!.get('a').attributes('href')).toBe('/coin/1')
    expect(rows[2]!.text()).toContain('c. 1930')
  })

  it('says when a grouped view shows only part of a large collection', () => {
    const wrapper = mountContent({ groupBy: 'year', truncated: true, total: 2500 })
    expect(wrapper.text()).toContain('Showing the first 3 of 2500 coins')
  })
})
