import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UcoinNotFoundCoin from '../UcoinNotFoundCoin.vue'
import type { Coin } from '@/types'

const coin = { id: 70, name: '5 euro 2023 – COC', country: 'Países Baixos' } as unknown as Coin
const netherlands = { code: 'pays-bas', name: 'Países Baixos' }

function mountCoin(chooseType = vi.fn().mockResolvedValue('pending'), candidatesFor = vi.fn().mockResolvedValue([
  { id: 380095, title: '5 Euros - Willem-Alexander (COC)', issuer: netherlands, minYear: 2023, maxYear: 2023 },
  { id: 352548, title: '5 Euros - Willem-Alexander (Willemstad)', issuer: netherlands, minYear: 2023, maxYear: 2023 },
])) {
  const wrapper = mount(UcoinNotFoundCoin, {
    props: { coin, candidatesFor, chooseType },
    global: { stubs: { RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' } } },
  })
  return { wrapper, chooseType, candidatesFor }
}

describe('UcoinNotFoundCoin', () => {
  it('lists the candidate types with a link to Numista and links the one chosen', async () => {
    const { wrapper, chooseType, candidatesFor } = mountCoin()
    expect(candidatesFor).not.toHaveBeenCalled()
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(candidatesFor).toHaveBeenCalledWith(coin)
    const rows = wrapper.findAll('li li')
    expect(rows.map((row) => row.text())).toEqual([
      expect.stringContaining('5 Euros - Willem-Alexander (COC)'),
      expect.stringContaining('5 Euros - Willem-Alexander (Willemstad)'),
    ])
    expect(rows[0]!.get('a').attributes('href')).toBe('https://pt.numista.com/380095')

    await rows[0]!.get('button').trigger('click')
    await flushPromises()
    expect(chooseType).toHaveBeenCalledWith(coin, 380095)
    expect(wrapper.text()).toContain('Linked to Numista. Its details')
    // Nothing left to choose.
    expect(wrapper.find('form').exists()).toBe(false)
  })

  it('takes a pasted Numista link, and says when it is not one', async () => {
    const { wrapper, chooseType } = mountCoin(vi.fn().mockResolvedValue('completed'))
    await wrapper.get('button').trigger('click')
    await flushPromises()

    await wrapper.get('input').setValue('https://pt.ucoin.net/coin/netherlands-5-euro-2023/')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.text()).toContain('That is not a Numista link or number.')
    expect(chooseType).not.toHaveBeenCalled()

    await wrapper.get('input').setValue('https://pt.numista.com/380096')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(chooseType).toHaveBeenCalledWith(coin, 380096)
    expect(wrapper.text()).toContain('Linked to Numista and completed.')
  })
})
