import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import DesktopCollectionHeader from '../DesktopCollectionHeader.vue'
import { setLocale } from '@/i18n'

function mountHeader(props: Record<string, unknown> = {}) {
  return mount(DesktopCollectionHeader, {
    props: {
      search: '', selectedDenomination: '', selectedSet: '', userSets: [{ id: 1, name: 'Euros' }], denominationOptions: [],
      sortKey: 'created_at_desc', gridSide: null, layout: 'grid', ...props,
    },
    slots: { main: '<label data-testid="main-slot">Ver por</label>', more: '<label data-testid="more-slot">Raridade</label>' },
    global: { stubs: { DenominationFilter: true, SortSelect: true } },
  })
}

describe('DesktopCollectionHeader layout', () => {
  beforeEach(() => setLocale('pt-PT'))

  it('puts the search beside grid or list and the side shown', async () => {
    const wrapper = mountHeader()

    expect(wrapper.get('input[type="text"]').attributes('placeholder')).toBe('Pesquisar moedas por nome, governante, inscrição...')
    await wrapper.get('[data-testid="gallery-layout-list"]').trigger('click')
    expect(wrapper.emitted('update:layout')).toEqual([['list']])
    // A side, then none again on a second press.
    await wrapper.get('[data-testid="gallery-side-obverse"]').trigger('click')
    expect(wrapper.emitted('update:gridSide')).toEqual([['obverse']])
    await wrapper.setProps({ gridSide: 'obverse' })
    await wrapper.get('[data-testid="gallery-side-obverse"]').trigger('click')
    expect(wrapper.emitted('update:gridSide')?.[1]).toEqual([null])
  })

  it('keeps the other filters a click away, open by themselves while one of them is on', async () => {
    const closed = mountHeader()
    expect(closed.get('[data-testid="main-slot"]').isVisible()).toBe(true)
    expect(closed.get('[data-testid="collection-more-filters"]').attributes('style')).toContain('display: none')
    const toggle = closed.get('[data-testid="collection-more-toggle"]')
    expect(toggle.text()).toBe('Mais filtros')
    await toggle.trigger('click')
    expect(closed.get('[data-testid="collection-more-filters"]').attributes('style') ?? '').not.toContain('display: none')
    expect(toggle.text()).toBe('Menos filtros')

    const open = mountHeader({ moreActive: 2 })
    expect(open.get('[data-testid="collection-more-filters"]').attributes('style') ?? '').not.toContain('display: none')
    await open.get('[data-testid="collection-more-toggle"]').trigger('click')
    expect(open.get('[data-testid="collection-more-toggle"]').text()).toBe('Mais filtros (2)')
  })
})
