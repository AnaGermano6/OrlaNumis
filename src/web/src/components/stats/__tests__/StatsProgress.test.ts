import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { beforeEach, describe, expect, it } from 'vitest'
import StatsProgress from '../StatsProgress.vue'
import { setLocale } from '@/i18n'

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots, attrs }) => () => h('a', { ...attrs, href: props.to }, slots.default?.()),
})

const years = (from: number, to: number, owned = (index: number) => index % 10) =>
  Array.from({ length: to - from + 1 }, (_, index) => ({ year: from + index, total: 10, owned: owned(index) }))

const CATALOGUE = {
  countries: [
    { country: 'PT', commemoratives: 40, collector: 30, ownedCommemoratives: 30, ownedCollector: 5, commemorativeYears: [{ year: 2024, total: 2, owned: 2 }, { year: 2025, total: 3, owned: 1 }] },
    { country: 'FR', commemoratives: 60, collector: 300, ownedCommemoratives: 10, ownedCollector: 0 },
    { country: 'PT', currency: 'PTE', commemoratives: 100, collector: 100, ownedCommemoratives: 46, ownedCollector: 4 },
  ],
  commemorativeYears: years(2004, 2025),
  circulation: [
    { country: 'HR', years: [{ year: 2023, total: 8, owned: 1 }, { year: 2024, total: 8, owned: 0 }, { year: 2025, total: 8, owned: 3 }] },
    { country: 'PT', years: [{ year: 2024, total: 8, owned: 8 }, { year: 2025, total: 8, owned: 4 }] },
  ],
}

function mountCard(catalogue: InstanceType<typeof StatsProgress>['$props']['catalogue'] = CATALOGUE) {
  return mount(StatsProgress, { props: { catalogue }, attachTo: document.body, global: { stubs: { RouterLink: RouterLinkStub, 'router-link': RouterLinkStub } } })
}

describe('StatsProgress', () => {
  beforeEach(() => {
    setLocale('pt-PT')
  })

  it('shows how much of the €2 commemoratives the collection has, every year a scroll away', () => {
    const wrapper = mountCard()
    const chart = wrapper.get('[data-testid="progress-commemoratives"]')

    // The euro countries' €2 commemoratives (not their collector coins, nor the escudos): 40 of 100.
    expect(chart.get('[data-testid="progress-percent"]').text()).toBe('40%')
    expect(chart.get('[data-testid="progress-count"]').text()).toBe('40 de 100 moedas')
    expect(chart.get('[data-testid="progress-ring"]').attributes('href')).toBe('/sets/commemoratives')
    // Every year, each with its year in full, in a strip that scrolls.
    const bars = chart.findAll('li')
    expect(bars).toHaveLength(22)
    expect(bars[0]!.attributes('title')).toBe('2004: 0 de 10')
    expect(bars[21]!.attributes('title')).toBe('2025: 1 de 10')
    expect(chart.get('[data-testid="progress-scroller"]').classes()).toContain('overflow-x-auto')
    expect(chart.text()).toContain('2004')
    wrapper.unmount()
  })

  it('shows the circulation coins under them, the eight of each year', async () => {
    const wrapper = mountCard()
    await flushPromises()
    const chart = wrapper.get('[data-testid="progress-circulation"]')

    // Croatia 4 of 24, Portugal 12 of 16: 16 of 40.
    expect(chart.get('h3').text()).toBe('Moedas correntes')
    expect(chart.get('[data-testid="progress-percent"]').text()).toBe('40%')
    expect(chart.get('[data-testid="progress-count"]').text()).toBe('16 de 40 moedas')
    // The years of every country added up; no catalogue to open.
    expect(chart.findAll('li').map((bar) => bar.attributes('title'))).toEqual(['2023: 1 de 8', '2024: 8 de 16', '2025: 7 de 16'])
    expect(chart.get('[data-testid="progress-ring"]').element.tagName).toBe('DIV')
    wrapper.unmount()
  })

  it('has one country choice for every chart, of the countries of either', async () => {
    const wrapper = mountCard()
    const select = wrapper.get('[data-testid="progress-country"]')
    expect(wrapper.findAll('[data-testid="progress-country"]')).toHaveLength(1)
    expect(select.findAll('option').map((option) => option.text())).toEqual(['Todos os países', '🇭🇷 Croácia', '🇫🇷 França', '🇵🇹 Portugal'])

    // Portugal: its commemoratives and its circulation coins.
    await select.setValue('PT')
    const commemoratives = wrapper.get('[data-testid="progress-commemoratives"]')
    expect(commemoratives.get('[data-testid="progress-percent"]').text()).toBe('75%')
    expect(commemoratives.get('[data-testid="progress-count"]').text()).toBe('30 de 40 moedas')
    expect(commemoratives.get('[data-testid="progress-ring"]').attributes('href')).toBe('/sets/commemoratives?country=PT')
    expect(commemoratives.findAll('li').map((bar) => bar.attributes('title'))).toEqual(['2024: 2 de 2', '2025: 1 de 3'])
    const circulation = wrapper.get('[data-testid="progress-circulation"]')
    expect(circulation.get('[data-testid="progress-count"]').text()).toBe('12 de 16 moedas')
    expect(circulation.findAll('li').map((bar) => bar.attributes('title'))).toEqual(['2024: 8 de 8', '2025: 4 de 8'])

    // Croatia has no commemoratives here: the chart says so.
    await select.setValue('HR')
    expect(wrapper.get('[data-testid="progress-commemoratives"] [data-testid="progress-none"]').text()).toBe('Nenhuma moeda deste país.')
    expect(wrapper.get('[data-testid="progress-circulation"] [data-testid="progress-count"]').text()).toBe('4 de 24 moedas')

    await select.setValue('')
    expect(wrapper.get('[data-testid="progress-commemoratives"] [data-testid="progress-percent"]').text()).toBe('40%')
    wrapper.unmount()
  })

  it('shows a share that would round to 0% with a decimal', () => {
    const wrapper = mountCard({ countries: [], circulation: [{ country: 'PT', years: [{ year: 2024, total: 4208, owned: 14 }] }] })
    expect(wrapper.get('[data-testid="progress-circulation"] [data-testid="progress-percent"]').text()).toBe('0,3%')
    wrapper.unmount()
  })

  it('stays out of the way without the catalogue', () => {
    const wrapper = mountCard(null)
    expect(wrapper.find('[data-testid="stats-progress"]').exists()).toBe(false)
    wrapper.unmount()
  })
})
