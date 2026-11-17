import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { beforeEach, describe, expect, it } from 'vitest'
import StatsCountryLeaders from '../StatsCountryLeaders.vue'
import StatsDecades from '../StatsDecades.vue'
import StatsPreciousMetals from '../StatsPreciousMetals.vue'
import { setLocale } from '@/i18n'

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots, attrs }) => () => h('a', { ...attrs, href: props.to }, slots.default?.()),
})
const stubs = { RouterLink: RouterLinkStub, 'router-link': RouterLinkStub }
const text = (value: string) => value.replace(/\s+/g, ' ').trim()

beforeEach(() => setLocale('pt-PT'))

describe('StatsCountryLeaders', () => {
  it("lists the countries the collection has most of, as a share of each, and opens each one's catalogue", () => {
    const wrapper = mount(StatsCountryLeaders, { props: { countries: [
      { country: 'FR', commemoratives: 60, collector: 140, ownedCommemoratives: 20, ownedCollector: 0 },
      { country: 'PT', commemoratives: 40, collector: 10, ownedCommemoratives: 30, ownedCollector: 5 },
      { country: 'PT', currency: 'PTE', commemoratives: 100, collector: 100, ownedCommemoratives: 46, ownedCollector: 4 },
      { country: 'MT', commemoratives: 20, collector: 0, ownedCommemoratives: 0, ownedCollector: 0 },
    ] }, global: { stubs } })

    const rows = wrapper.findAll('li')
    expect(rows.map((row) => row.findAll('a > span:first-child > span').map((part) => text(part.text())))).toEqual([
      ['1.🇵🇹 Portugal', '70% (35/50)'], ['2.🇵🇹 Portugal (escudos)', '25% (50/200)'], ['3.🇫🇷 França', '10% (20/200)'],
    ])
    expect(rows.map((row) => row.get('a').attributes('href'))).toEqual([
      '/sets/commemoratives?country=PT', '/sets/commemoratives?country=PT&currency=PTE', '/sets/commemoratives?country=FR',
    ])
  })
})

describe('StatsPreciousMetals', () => {
  it('shows the fine gold and silver, their whole weight, and the pieces of unknown fineness', () => {
    const wrapper = mount(StatsPreciousMetals, { props: { metals: {
      gold: { fineGrams: 0, grossGrams: 0, pieces: 0, unknownFineness: 0 },
      silver: { fineGrams: 1450.24, grossGrams: 1700, pieces: 12, unknownFineness: 2 },
    } } })

    const silver = wrapper.get('[data-testid="metal-silver"]')
    expect(silver.get('[data-testid="metal-fine"]').text()).toBe('1\u00a0450,2 g')
    expect(text(silver.text())).toContain('Peso total 1 700 g · 12 peças')
    expect(silver.get('[data-testid="metal-unknown"]').text()).toBe('2 sem toque conhecido (só no peso total)')
    const gold = wrapper.get('[data-testid="metal-gold"]')
    expect(gold.get('[data-testid="metal-fine"]').text()).toBe('0 g')
    expect(gold.text()).toContain('Nenhuma moeda de ouro')
  })

  it("lists the coins' other materials under them, most first, without gold, silver or the banknotes' paper", () => {
    const wrapper = mount(StatsPreciousMetals, { props: {
      metals: { gold: { fineGrams: 0, grossGrams: 0, pieces: 0, unknownFineness: 0 }, silver: { fineGrams: 0, grossGrams: 0, pieces: 0, unknownFineness: 0 } },
      byMaterial: [
        { material: 'Copper-nickel', count: 118 }, { material: 'Gold', count: 1 }, { material: 'Bimetallic', count: 122 },
        { material: 'Paper', count: 17 }, { material: 'Silver', count: 4 }, { material: 'Billon', count: 6 },
      ],
    } })

    const others = wrapper.get('[data-testid="metal-others"]')
    expect(others.get('h3').text()).toBe('Outros materiais')
    expect(others.findAll('li').map((item) => item.attributes('data-testid'))).toEqual(['material-Bimetallic', 'material-Copper-nickel', 'material-Billon'])
    expect(others.get('[data-testid="material-Bimetallic"]').attributes('title')).toBe('122 moedas diferentes')
  })

  it('stays out of the way when the server does not say', () => {
    expect(mount(StatsPreciousMetals, { props: { metals: undefined } }).find('[data-testid="stats-metals"]').exists()).toBe(false)
  })
})

describe('StatsDecades', () => {
  it('counts the coins by decade, those before 1900 together', () => {
    const wrapper = mount(StatsDecades, { props: { byDecade: [
      { decade: -40, count: 1 }, { decade: 1640, count: 1 }, { decade: 1880, count: 3 }, { decade: 1900, count: 2 }, { decade: 1990, count: 75 },
    ] } })

    const bars = wrapper.findAll('li')
    expect(bars.map((bar) => bar.attributes('title'))).toEqual(['Antes de 1900: 5', '1900–1909: 2', '1990–1999: 75'])
    expect(bars.map((bar) => bar.get('span').text())).toEqual(['5', '2', '75'])
    // In a strip that scrolls, each decade written in full.
    expect(wrapper.get('[data-testid="decades-scroller"]').classes()).toContain('overflow-x-auto')
    expect(wrapper.text()).toContain('<1900')
  })
})
