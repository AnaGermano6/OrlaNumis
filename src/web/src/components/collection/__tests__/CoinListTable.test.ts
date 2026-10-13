import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { beforeEach, describe, expect, it } from 'vitest'
import CoinListTable from '../CoinListTable.vue'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots }) => () => h('a', { href: props.to }, slots.default?.()),
})

function coin(id: number, quantity: number | undefined): Coin {
  return { id, name: `Coin ${id}`, country: 'África do Sul', year: 1984, quantity, images: [] } as unknown as Coin
}

describe('CoinListTable', () => {
  beforeEach(() => setLocale('pt-PT'))

  it('shows how many of each coin are held, one for coins saved before the field', () => {
    const wrapper = mount(CoinListTable, {
      props: { coins: [coin(1, 3), coin(2, 1), coin(3, undefined)] },
      global: { stubs: { RouterLink: RouterLinkStub } },
    })

    expect(wrapper.findAll('th').map((th) => th.text())).toContain('Quantidade')
    expect(wrapper.findAll('[data-testid="list-quantity"]').map((cell) => cell.text())).toEqual(['3', '1', '1'])
    // On a phone, the line under the name says it when there are several.
    const sublines = wrapper.findAll('td span.block').map((span) => span.text())
    expect(sublines[0]).toBe('África do Sul · 1984 · × 3')
    expect(sublines[1]).toBe('África do Sul · 1984')
  })
})
