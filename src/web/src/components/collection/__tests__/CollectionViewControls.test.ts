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
})
