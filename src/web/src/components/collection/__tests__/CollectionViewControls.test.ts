import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import CollectionViewControls from '../CollectionViewControls.vue'
import { setLocale } from '@/i18n'

const baseProps = { groupBy: 'none' as const, layout: 'grid' as const, country: '', presentation: '', rarity: '', countryOptions: [] }

describe('CollectionViewControls', () => {
  afterEach(() => setLocale('en'))

  it('offers the Numista kinds of coin in the collection, with their counts', async () => {
    setLocale('pt-PT')
    const wrapper = mount(CollectionViewControls, {
      props: {
        ...baseProps,
        classifications: [
          { key: 'standard', label: 'Standard circulation coins', count: 162 },
          { key: 'circulating-commemorative', label: 'Moedas circulantes comemorativas', count: 135 },
          { key: 'non-circulating', label: 'Moedas não circulantes', count: 33 },
          { key: 'Patterns', label: 'Patterns', count: 1 },
        ],
      },
    })

    const select = wrapper.get('[data-testid="classification-filter"]')
    expect(select.findAll('option').map((option) => option.text())).toEqual([
      'Todos os tipos',
      'Circulantes comuns (162)',
      'Circulantes comemorativas (135)',
      'Não circulantes (33)',
      'Patterns (1)',
    ])
    await select.setValue('non-circulating')
    expect(wrapper.emitted('update:classification')).toEqual([['non-circulating']])
  })

  it("offers the chosen country's periods, with the years of those that do not name them", async () => {
    setLocale('pt-PT')
    const wrapper = mount(CollectionViewControls, {
      props: {
        ...baseProps,
        country: 'Portugal',
        periodOptions: [
          { period: 'União Europeia (Euro)', count: 55, fromYear: 2002, toYear: 2024 },
          { period: 'República Portuguesa 1986–2001', count: 48, fromYear: 1986, toYear: 2001 },
          { period: 'Reino de Portugal 1836–1910', count: 3, fromYear: 1880, toYear: 1900 },
        ],
      },
    })
    const select = wrapper.get('[data-testid="period-filter"]')
    expect(select.findAll('option').map((option) => option.text())).toEqual([
      'Todos os períodos',
      'União Europeia (Euro) 2002–2024 (55)',
      'República Portuguesa 1986–2001 (48)',
      'Reino de Portugal 1836–1910 (3)',
    ])
    await select.setValue('República Portuguesa 1986–2001')
    expect(wrapper.emitted('update:period')).toEqual([['República Portuguesa 1986–2001']])
  })

  it('has no period filter until a country with periods is chosen', () => {
    const wrapper = mount(CollectionViewControls, { props: baseProps })
    expect(wrapper.find('[data-testid="period-filter"]').exists()).toBe(false)
  })

  it('hides the filter when Numista says nothing of the coins', () => {
    const wrapper = mount(CollectionViewControls, { props: baseProps })
    expect(wrapper.find('[data-testid="classification-filter"]').exists()).toBe(false)
  })

  it('shows the flag of each country, drawn ones too, and keeps the plain name as the value', async () => {
    const wrapper = mount(CollectionViewControls, {
      props: { ...baseProps, extinctCountries: 1, countryOptions: [{ country: 'Croácia', count: 4 }, { country: 'Jugoslávia', count: 1 }] },
      attachTo: document.body,
    })

    const menu = wrapper.get('[data-testid="country-filter"]')
    expect(menu.attributes('role')).toBe('combobox')
    expect(menu.text()).toBe('All countries')
    await menu.trigger('click')
    const options = [...document.querySelectorAll<HTMLElement>('[data-testid="country-select-option"]')]
    expect(options.map((option) => option.textContent?.trim())).toEqual(['All countries', 'Extinct countries (1)', '🇭🇷Croácia (4)', 'Jugoslávia (1)'])
    // Yugoslavia has no emoji: its flag is drawn.
    expect(options[3]!.querySelector('img[data-testid="country-flag-image"]')).not.toBeNull()

    options[2]!.click()
    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('update:country')).toEqual([['Croácia']])
    expect(document.querySelector('[data-testid="country-select-list"]')).toBeNull()
    wrapper.unmount()
  })
})
