import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import CollectionKpis from '../CollectionKpis.vue'
import { setLocale } from '@/i18n'

describe('CollectionKpis', () => {
  afterEach(() => setLocale('en'))

  it('shows coins, coin types and countries with the coins missing a country', () => {
    setLocale('pt-PT')
    const wrapper = mount(CollectionKpis, {
      props: { summary: { coins: 1250, types: 830, countries: 42, withoutCountry: 3, coincards: 17, totalValue: 1250, unvaluedCoins: 0, totalValueEur: 1115.47, exchangeRate: { usdPerEur: 1.1206, date: '2026-10-09' } } },
    })

    const cards = wrapper.findAll('section > *').map((card) => card.findAll('span').map((span) => span.text().replace(/\s/g, ' ')))
    expect(cards).toEqual([
      ['1250', 'Moedas'],
      ['830', 'Tipos de moeda'],
      ['42', 'Países', '3 sem país'],
      ['17', 'Em coincard'],
      ['1115,47 €', 'Valor total', 'Taxa BCE de 09/10/2026'],
    ])
    expect(wrapper.findAll('section > *')[4]!.attributes('title')).toContain('1 € = 1,1206 US$')
    expect(wrapper.findAll('section > *')[1]!.attributes('title')).toContain('mesmo tipo Numista')
  })

  it('shows placeholders until the numbers load and hides the note when every coin has a country', async () => {
    const wrapper = mount(CollectionKpis, { props: { summary: null } })
    expect(wrapper.findAll('section > *').map((card) => card.find('span').text())).toEqual(['—', '—', '—', '—', '—'])

    await wrapper.setProps({ summary: { coins: 2, types: 2, countries: 1, withoutCountry: 0, coincards: 0, totalValue: 0, unvaluedCoins: 0, totalValueEur: 0 } })
    expect(wrapper.text()).not.toContain('without a country')
    expect(wrapper.text()).toContain('Countries')
  })

  it('shows only the coincards when their KPI is chosen', async () => {
    const wrapper = mount(CollectionKpis, { props: { summary: { coins: 5, types: 5, countries: 2, withoutCountry: 3, coincards: 1, totalValue: 0, unvaluedCoins: 0, totalValueEur: 0 } } })

    const coincard = wrapper.get('button')
    expect(coincard.text()).toContain('In coincard')
    await coincard.trigger('click')
    expect(wrapper.emitted('show-coincards')).toHaveLength(1)
  })

  it('shows a euro collection\'s total as is, with no exchange rate', () => {
    setLocale('pt-PT')
    const wrapper = mount(CollectionKpis, {
      props: { summary: { coins: 2, types: 2, countries: 1, withoutCountry: 0, coincards: 0, currency: 'EUR', totalValue: 3.01, unvaluedCoins: 0, totalValueEur: 3.01, exchangeRate: { usdPerEur: 1.12, date: '2026-10-09' } } },
    })

    const value = wrapper.findAll('section > *')[4]!
    expect(value.find('span').text().replace(/\s/g, ' ')).toBe('3,01 €')
    expect(value.text()).not.toContain('BCE')
    expect(value.attributes('title')).toBe('Soma do valor atual das moedas, em euros.')
  })

  it('explains a missing rate and coins without a value', () => {
    const wrapper = mount(CollectionKpis, { props: { summary: { coins: 5, types: 5, countries: 2, withoutCountry: 0, coincards: 0, totalValue: 120, unvaluedCoins: 2, totalValueEur: null } } })

    const value = wrapper.findAll('section > *')[4]!
    expect(value.find('span').text()).toBe('—')
    expect(value.text()).toContain('Exchange rate unavailable · 2 without a value')
    expect(value.attributes('title')).toContain('could not be loaded')
  })
})
