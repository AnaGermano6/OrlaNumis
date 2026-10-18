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

  it('hides the filter when Numista says nothing of the coins', () => {
    const wrapper = mount(CollectionViewControls, { props: baseProps })
    expect(wrapper.find('[data-testid="classification-filter"]').exists()).toBe(false)
  })

  it('shows the flag of each country, and keeps the plain name as the value', async () => {
    const wrapper = mount(CollectionViewControls, {
      props: { ...baseProps, countryOptions: [{ country: 'Croácia', count: 4 }, { country: 'Jugoslávia', count: 1 }] },
    })

    const select = wrapper.get('[data-testid="country-filter"]')
    expect(select.findAll('option').map((option) => option.text()).slice(1)).toEqual(['🇭🇷 Croácia (4)', 'Jugoslávia (1)'])
    await select.setValue('Croácia')
    expect(wrapper.emitted('update:country')).toEqual([['Croácia']])
  })
})
