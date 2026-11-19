import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import SetsByCountry from '../SetsByCountry.vue'
import { setLocale } from '@/i18n'
import type { CoinSetSummary } from '@/types'

vi.mock('vue-router', () => ({
  RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' },
}))

const SETS: CoinSetSummary[] = [
  { id: 10, name: 'Portugal · 2 €', color: '#c9a84c', setType: 'checklist', coinCount: 20, totalValue: 0, countries: [{ country: 'PT', owned: 20, total: 60 }], coverImageUrls: ['pt-a.jpg', 'pt-b.jpg'] },
  { id: 7, name: 'Emissão comum de 2007: Tratado de Roma', year: 2007, oneCoin: true, color: '#3a7bd5', setType: 'checklist', coinCount: 2, totalValue: 0, countries: [{ country: 'PT', owned: 1, total: 1 }, { country: 'ES', owned: 1, total: 1 }], coverImages: ['coins/roma.jpg'] },
  { id: 14, name: 'Tipos dos réis', color: '#999', setType: 'checklist', coinCount: 0, totalValue: 0, partial: true, countries: [{ country: 'PT', owned: 0, total: 1 }] },
  { id: 8, name: 'Moedas de 1991', color: '#777', setType: 'smart', coinCount: 4, totalValue: 0 },
  // Its list not read yet: no places, but not a set without any.
  { id: 15, name: 'Tipos dos réis', color: '#999', setType: 'checklist', coinCount: 0, totalValue: 0, partial: true, countries: [] },
  { id: 16, name: 'Emissão comum de 2009: UEM', year: 2009, oneCoin: true, color: '#999', setType: 'checklist', coinCount: 0, totalValue: 0, countries: [{ country: 'DE', owned: 1, total: 5 }, { country: 'AT', owned: 0, total: 4 }] },
]

const stubs = { AuthenticatedImage: { props: ['mediaPath'], template: '<img class="own" :src="mediaPath" />' } }
const text = (value: string) => value.replace(/\s+/g, ' ').trim()

describe('SetsByCountry', () => {
  beforeEach(() => setLocale('pt-PT'))
  afterEach(() => setLocale('en'))

  it("shows each country's flag, sets, coins, what it misses and how far it is, the most advanced first", () => {
    const wrapper = mount(SetsByCountry, { props: { sets: SETS }, global: { stubs } })

    const portugal = wrapper.get('[data-testid="sets-country-PT"]')
    // Its flag beside its name, as in the coin gallery.
    expect(portugal.get('h3').text()).toBe('🇵🇹Portugal')
    expect(portugal.get('[data-testid="sets-country-flag"]').attributes('aria-hidden')).toBe('true')
    expect(portugal.get('[data-testid="sets-country-sets"]').text()).toBe('3 conjuntos')
    expect(text(portugal.get('[data-testid="sets-country-coins"]').text())).toBe('21 / 62 moedas')
    expect(portugal.get('[data-testid="sets-country-missing"]').text()).toBe('faltam 41')
    expect(portugal.get('[data-testid="sets-country-percent"]').text()).toBe('34%')
    expect(portugal.text()).toContain('lista ainda a ser lida')

    const spain = wrapper.get('[data-testid="sets-country-ES"]')
    expect(spain.get('[data-testid="sets-country-sets"]').text()).toBe('1 conjunto')
    expect(spain.get('[data-testid="sets-country-missing"]').text()).toBe('Completo')
    expect(spain.get('[data-testid="sets-country-percent"]').text()).toBe('100%')

    // The most advanced first: Spain (100%), Portugal (34%), Germany (20%), Austria (0%).
    expect(wrapper.findAll('li[data-testid^="sets-country-"]').map((node) => node.attributes('data-testid'))).toEqual(['sets-country-ES', 'sets-country-PT', 'sets-country-DE', 'sets-country-AT'])
    expect(wrapper.get('[data-testid="sets-by-country-others"]').text()).toBe('1 conjunto não tem lugares por país (por filtros, livre ou de objetivo) e não conta aqui.')
  })

  it('opens a country to show its sets, each with a picture, its year or coins and whether it is complete', async () => {
    const wrapper = mount(SetsByCountry, { props: { sets: SETS }, global: { stubs } })
    const portugal = wrapper.get('[data-testid="sets-country-PT"]')
    const toggle = portugal.get('[data-testid="sets-country-toggle"]')
    expect(portugal.find('[data-testid="sets-country-list"]').exists()).toBe(false)
    expect(toggle.attributes('aria-label')).toBe('Mostrar os conjuntos de Portugal')

    await toggle.trigger('click')
    expect(toggle.attributes('aria-expanded')).toBe('true')
    const catalogue = portugal.get('[data-testid="sets-country-set-10"]')
    expect(catalogue.attributes('href')).toBe('/sets/10')
    const lines = (link: typeof catalogue) => link.findAll('.flex-1 > span').map((line) => line.text())
    expect(lines(catalogue)).toEqual(['Portugal · 2 €', '20 / 60 moedas'])
    expect(catalogue.get('img').attributes('src')).toBe('pt-a.jpg')
    expect(catalogue.find('[data-testid="sets-country-set-missing"]').exists()).toBe(true)
    const rome = portugal.get('[data-testid="sets-country-set-7"]')
    expect(lines(rome)).toEqual(['Tratado de Roma', '2007'])
    expect(rome.get('img.own').attributes('src')).toBe('coins/roma.jpg')
    expect(rome.find('[data-testid="sets-country-set-complete"]').exists()).toBe(true)

    await toggle.trigger('click')
    expect(portugal.find('[data-testid="sets-country-list"]').exists()).toBe(false)
  })

  it('finds a country by its name, without accents, and puts the countries in another order', async () => {
    const wrapper = mount(SetsByCountry, { props: { sets: SETS, search: 'austr' }, global: { stubs } })
    const shown = () => wrapper.findAll('li[data-testid^="sets-country-"]').map((node) => node.attributes('data-testid'))
    expect(shown()).toEqual(['sets-country-AT'])

    await wrapper.setProps({ search: 'ESPANHA' })
    expect(shown()).toEqual(['sets-country-ES'])
    await wrapper.setProps({ search: 'Narnia' })
    expect(wrapper.get('[data-testid="sets-by-country-no-match"]').text()).toBe('Nenhum país corresponde à pesquisa.')

    await wrapper.setProps({ search: '', sort: 'missing' })
    // Germany and Austria both miss four: by name.
    expect(shown()).toEqual(['sets-country-PT', 'sets-country-DE', 'sets-country-AT', 'sets-country-ES'])
    await wrapper.setProps({ sort: 'progressAsc' })
    expect(shown()).toEqual(['sets-country-AT', 'sets-country-DE', 'sets-country-PT', 'sets-country-ES'])
    await wrapper.setProps({ sort: 'name' })
    expect(shown()).toEqual(['sets-country-DE', 'sets-country-AT', 'sets-country-ES', 'sets-country-PT'])
  })

  it('says when no set has places by country', () => {
    const wrapper = mount(SetsByCountry, { props: { sets: [SETS[3]!] }, global: { stubs } })
    expect(wrapper.find('[data-testid="sets-by-country-empty"]').exists()).toBe(true)
    expect(wrapper.find('li[data-testid^="sets-country-"]').exists()).toBe(false)
  })
})
