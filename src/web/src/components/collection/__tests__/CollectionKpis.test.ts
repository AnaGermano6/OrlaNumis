import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import CollectionKpis from '../CollectionKpis.vue'
import { setLocale } from '@/i18n'

describe('CollectionKpis', () => {
  afterEach(() => setLocale('en'))

  it('shows coins, current and commemorative coins and countries with the coins missing a country', () => {
    setLocale('pt-PT')
    const wrapper = mount(CollectionKpis, {
      props: { summary: { coins: 1250, currentCoins: 600, circulationUnknown: 40, commemorativeCoins: 210, commemorativeUnknown: 40, types: 830, countries: 42, withoutCountry: 3, coincards: 17, totalValue: 1250, unvaluedCoins: 0, totalValueEur: 1115.47, exchangeRate: { usdPerEur: 1.1206, date: '2026-10-09' } } },
    })

    const cards = wrapper.findAll('section > *').map((card) => card.findAll('span').map((span) => span.text().replace(/\s/g, ' ')))
    expect(cards).toEqual([
      ['1250', 'Moedas'],
      ['600', 'Moedas correntes'],
      ['210', 'Moedas comemorativas'],
      ['42', 'Países', '3 sem país'],
      ['17', 'Em coincard'],
      ['1115,47 €', 'Valor total', 'Taxa BCE de 09/10/2026'],
    ])
    expect(wrapper.findAll('section > *')[5]!.attributes('title')).toContain('1 € = 1,1206 US$')
    expect(wrapper.findAll('section > *')[1]!.attributes('title')).toContain('curso legal')
    expect(wrapper.findAll('section > *')[2]!.attributes('title')).toContain('catálogo Numista')
  })

  it('shows placeholders until the numbers load and hides the note when every coin has a country', async () => {
    const wrapper = mount(CollectionKpis, { props: { summary: null } })
    expect(wrapper.findAll('section > *').map((card) => card.find('span').text())).toEqual(['—', '—', '—', '—', '—', '—'])

    await wrapper.setProps({ summary: { coins: 2, types: 2, countries: 1, withoutCountry: 0, coincards: 0, totalValue: 0, unvaluedCoins: 0, totalValueEur: 0 } })
    expect(wrapper.text()).not.toContain('without a country')
    expect(wrapper.text()).toContain('Countries')
  })

  it('counts the extinct countries among the countries and shows their coins when chosen', async () => {
    setLocale('pt-PT')
    const wrapper = mount(CollectionKpis, { props: { summary: { coins: 5, types: 5, countries: 9, extinctCountries: 3, withoutCountry: 0, coincards: 0, totalValue: 0, unvaluedCoins: 0, totalValueEur: 0 } } })
    const line = wrapper.get('[data-testid="kpi-extinct-countries"]')
    expect(line.text()).toBe('3 extintos')
    expect(line.attributes('title')).toContain('já não existem')
    await line.trigger('click')
    expect(wrapper.emitted('show-extinct')).toHaveLength(1)

    await wrapper.setProps({ summary: { coins: 5, types: 5, countries: 9, extinctCountries: 0, withoutCountry: 0, coincards: 0, totalValue: 0, unvaluedCoins: 0, totalValueEur: 0 } })
    expect(wrapper.find('[data-testid="kpi-extinct-countries"]').exists()).toBe(false)
  })

  it('filters the gallery to the current coins when their KPI is chosen, and back', async () => {
    const summary = { coins: 5, currentCoins: 3, circulationUnknown: 1, types: 5, countries: 2, withoutCountry: 0, coincards: 1, totalValue: 0, unvaluedCoins: 0, totalValueEur: 0 }
    const wrapper = mount(CollectionKpis, { props: { summary } })

    // Only the number: no count of coins without data, no link.
    const card = wrapper.get('[data-testid="kpi-current"]')
    expect(card.element.tagName).toBe('BUTTON')
    expect(card.findAll('span').map((span) => span.text())).toEqual(['3', 'Current coins'])
    expect(card.findAll('button')).toHaveLength(0)
    expect(card.attributes('aria-pressed')).toBe('false')
    await card.trigger('click')
    expect(wrapper.emitted('toggle-current')).toHaveLength(1)

    // While the filter is on, the card stands out; choosing it again shows every coin.
    await wrapper.setProps({ currentOnly: true })
    expect(wrapper.get('[data-testid="kpi-current"]').attributes('aria-pressed')).toBe('true')
    expect(wrapper.get('[data-testid="kpi-current"]').classes()).toContain('border-gold')
    await wrapper.get('[data-testid="kpi-current"]').trigger('click')
    expect(wrapper.emitted('toggle-current')).toHaveLength(2)
  })

  it('filters the gallery to the commemorative coins when their KPI is chosen, and back', async () => {
    const summary = { coins: 5, currentCoins: 3, commemorativeCoins: 2, commemorativeUnknown: 1, types: 5, countries: 2, withoutCountry: 0, coincards: 1, totalValue: 0, unvaluedCoins: 0, totalValueEur: 0 }
    const wrapper = mount(CollectionKpis, { props: { summary } })

    const card = wrapper.get('[data-testid="kpi-commemorative"]')
    expect(card.findAll('span').map((span) => span.text())).toEqual(['2', 'Commemorative coins'])
    expect(card.findAll('button')).toHaveLength(0)
    await card.trigger('click')
    expect(wrapper.emitted('toggle-commemorative')).toHaveLength(1)
    expect(wrapper.emitted('toggle-current')).toBeUndefined()

    await wrapper.setProps({ commemorativeOnly: true })
    expect(wrapper.get('[data-testid="kpi-commemorative"]').attributes('aria-pressed')).toBe('true')
    expect(wrapper.get('[data-testid="kpi-commemorative"]').classes()).toContain('border-gold')
    expect(wrapper.get('[data-testid="kpi-current"]').classes()).not.toContain('border-gold')
  })

  it('shows only the coincards when their KPI is chosen', async () => {
    const wrapper = mount(CollectionKpis, { props: { summary: { coins: 5, types: 5, countries: 2, withoutCountry: 3, coincards: 1, totalValue: 0, unvaluedCoins: 0, totalValueEur: 0 } } })

    const coincard = wrapper.get('[data-testid="kpi-coincards"]')
    expect(coincard.text()).toContain('In coincard')
    await coincard.trigger('click')
    expect(wrapper.emitted('show-coincards')).toHaveLength(1)
  })

  it('shows a euro collection\'s total as is, with no exchange rate', () => {
    setLocale('pt-PT')
    const wrapper = mount(CollectionKpis, {
      props: { summary: { coins: 2, types: 2, countries: 1, withoutCountry: 0, coincards: 0, currency: 'EUR', totalValue: 3.01, unvaluedCoins: 0, totalValueEur: 3.01, exchangeRate: { usdPerEur: 1.12, date: '2026-10-09' } } },
    })

    const value = wrapper.findAll('section > *')[5]!
    expect(value.find('span').text().replace(/\s/g, ' ')).toBe('3,01 €')
    expect(value.text()).not.toContain('BCE')
    expect(value.attributes('title')).toBe('Soma do valor atual das moedas, em euros.')
  })

  it('explains a missing rate and coins without a value', () => {
    const wrapper = mount(CollectionKpis, { props: { summary: { coins: 5, types: 5, countries: 2, withoutCountry: 0, coincards: 0, totalValue: 120, unvaluedCoins: 2, totalValueEur: null } } })

    const value = wrapper.findAll('section > *')[5]!
    expect(value.find('span').text()).toBe('—')
    expect(value.text()).toContain('Exchange rate unavailable · 2 without a value')
    expect(value.attributes('title')).toContain('could not be loaded')
  })
})
