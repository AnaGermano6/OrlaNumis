import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import SetsByCountry from '../SetsByCountry.vue'
import { setLocale } from '@/i18n'
import type { CoinSetSummary } from '@/types'

vi.mock('vue-router', () => ({
  RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' },
}))

const SETS: CoinSetSummary[] = [
  { id: 10, name: 'Portugal · 2 €', color: '#c9a84c', setType: 'checklist', coinCount: 20, totalValue: 0, countries: [{ country: 'PT', owned: 20, total: 60 }] },
  { id: 7, name: 'Tratado de Roma', color: '#3a7bd5', setType: 'checklist', coinCount: 2, totalValue: 0, countries: [{ country: 'PT', owned: 1, total: 1 }, { country: 'ES', owned: 1, total: 1 }] },
  { id: 14, name: 'Tipos dos réis', color: '#999', setType: 'checklist', coinCount: 0, totalValue: 0, partial: true, countries: [{ country: 'PT', owned: 0, total: 1 }] },
  { id: 8, name: 'Moedas de 1991', color: '#777', setType: 'smart', coinCount: 4, totalValue: 0 },
  // Its list not read yet: no places, but not a set without any.
  { id: 15, name: 'Tipos dos réis', color: '#999', setType: 'checklist', coinCount: 0, totalValue: 0, partial: true, countries: [] },
]

describe('SetsByCountry', () => {
  beforeEach(() => setLocale('pt-PT'))
  afterEach(() => setLocale('en'))

  it('shows what is missing in each country, and in each of its sets', () => {
    const wrapper = mount(SetsByCountry, { props: { sets: SETS } })

    const portugal = wrapper.get('[data-testid="sets-country-PT"]')
    expect(portugal.text()).toContain('Portugal')
    expect(portugal.get('[data-testid="sets-country-missing"]').text()).toBe('faltam 41')
    expect(portugal.text()).toContain('21 de 62 · em 3 conjuntos · lista ainda a ser lida')
    expect(portugal.get('[data-testid="sets-country-set-10"]').text()).toContain('faltam 40 de 60')
    expect(portugal.get('[data-testid="sets-country-set-7"]').text()).toContain('Completo')
    expect(portugal.get('[data-testid="sets-country-set-14"]').text()).toContain('falta 1 de 1')
    expect(portugal.get('[data-testid="sets-country-set-10"]').attributes('href')).toBe('/sets/10')

    const spain = wrapper.get('[data-testid="sets-country-ES"]')
    expect(spain.get('[data-testid="sets-country-missing"]').text()).toBe('Completo')
    expect(spain.text()).toContain('em 1 conjunto')
    // Portugal has the most places, so it comes first.
    expect(wrapper.findAll('details').map((node) => node.attributes('data-testid'))).toEqual(['sets-country-PT', 'sets-country-ES'])
    expect(wrapper.get('[data-testid="sets-by-country-others"]').text()).toBe('1 conjunto não tem lugares por país (por filtros, livre ou de objetivo) e não conta aqui.')
  })

  it('says when no set has places by country', () => {
    const wrapper = mount(SetsByCountry, { props: { sets: [SETS[3]!] } })
    expect(wrapper.find('[data-testid="sets-by-country-empty"]').exists()).toBe(true)
    expect(wrapper.find('details').exists()).toBe(false)
  })
})
