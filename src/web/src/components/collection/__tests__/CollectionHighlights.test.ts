import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import CollectionHighlights from '../CollectionHighlights.vue'
import { setLocale } from '@/i18n'
import { setCollectionCurrency } from '@/utils/collectionCurrency'
import type { CollectionSummary } from '@/api/client'

const BASE: CollectionSummary = {
  coins: 6, types: 6, countries: 2, withoutCountry: 4, coincards: 1, totalValue: 323.81, unvaluedCoins: 4,
  totalValueEur: 288.96, exchangeRate: { usdPerEur: 1.25, date: '2026-10-09' },
}

function mountHighlights(summary: CollectionSummary | null) {
  return mount(CollectionHighlights, {
    props: { summary },
    global: { stubs: { RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' }, AuthenticatedImage: { props: ['mediaPath'], template: '<img :src="mediaPath" />' } } },
  })
}

describe('CollectionHighlights', () => {
  afterEach(() => {
    setLocale('en')
    setCollectionCurrency('USD')
  })

  it('names the rarest, oldest and most valuable coins with what makes them stand out', () => {
    setLocale('pt-PT')
    setCollectionCurrency('EUR')
    const wrapper = mountHighlights({
      ...BASE,
      highlights: {
        rarest: { id: 5, name: '50 Escudos 1968', country: 'Portugal', year: 1968, dateRange: '1968', rarity: 'scarce', grade: 'VF', currentValue: 23.81, imagePath: 'coin-5/obverse.jpg' },
        oldest: { id: 7, name: 'Denário de César', country: '', year: -44, dateRange: 'c. 46-44 BC', rarity: '', grade: 'VF', currentValue: null },
        mostValuable: { id: 9, name: 'Aureus', country: 'Roma', year: null, dateRange: 'c. 200', rarity: '', grade: '', currentValue: 300 },
      },
    })

    const items = wrapper.findAll('li')
    expect(wrapper.find('h2').text()).toBe('Destaques da coleção')
    expect(items[0]!.text()).toContain('Mais rara')
    expect(items[0]!.get('a').attributes('href')).toBe('/coin/5')
    expect(items[0]!.text()).toContain('Escassa')
    expect(items[0]!.text()).toContain('Portugal · 1968')
    expect(items[0]!.get('img').attributes('src')).toBe('coin-5/obverse.jpg')
    expect(items[1]!.text()).toContain('44 a.C.')
    expect(items[1]!.text()).toContain('VF')
    // In the currency every page uses.
    expect(items[2]!.text().replace(/\s/g, ' ')).toContain('300,00 €')
    expect(items[2]!.text()).toContain('Roma · c. 200')
  })

  it('explains highlights that have no coin and shows values in a dollar collection in dollars', () => {
    const wrapper = mountHighlights({
      ...BASE,
      totalValueEur: null,
      exchangeRate: undefined,
      highlights: { rarest: null, oldest: null, mostValuable: { id: 9, name: 'Aureus', country: '', year: null, dateRange: '', rarity: '', grade: '', currentValue: 300 } },
    })

    const items = wrapper.findAll('li')
    expect(items[0]!.text()).toContain('No coin has a rarity yet')
    expect(items[1]!.text()).toContain('No coin has a year yet')
    expect(items[2]!.text()).toContain('$300.00')
  })

  it('stays hidden for an empty collection or before the summary loads', () => {
    expect(mountHighlights(null).find('section').exists()).toBe(false)
    expect(mountHighlights({ ...BASE, coins: 0, highlights: { rarest: null, oldest: null, mostValuable: null } }).find('section').exists()).toBe(false)
  })
})
