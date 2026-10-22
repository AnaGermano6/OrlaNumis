import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import BanknotesPage from '../BanknotesPage.vue'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'

const mocks = vi.hoisted(() => ({
  getCoins: vi.fn(),
  getCoinCountries: vi.fn(),
  replace: vi.fn(),
  route: { query: {} as Record<string, string> },
  bulkAction: vi.fn(),
  getSets: vi.fn(),
  getCollectionSummary: vi.fn(),
  confirm: vi.fn(),
  alert: vi.fn(),
}))

vi.mock('@/composables/useDialog', () => ({ useDialog: () => ({ showConfirm: mocks.confirm, showAlert: mocks.alert }) }))

vi.mock('vue-router', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-router')>(),
  useRoute: () => mocks.route,
  useRouter: () => ({ replace: mocks.replace }),
}))

vi.mock('@/api/client', () => ({
  getCoins: mocks.getCoins,
  getCoinCountries: mocks.getCoinCountries,
  bulkAction: mocks.bulkAction,
  getSets: mocks.getSets,
  getCollectionSummary: mocks.getCollectionSummary,
}))
vi.mock('@/components/CoinCard.vue', () => ({
  default: defineComponent({
    props: { coin: { type: Object, required: true }, selectable: Boolean, selected: Boolean },
    emits: ['toggle-select'],
    setup: (props, { emit }) => () => h('article', {
      'data-testid': 'card', 'data-selectable': String(props.selectable), 'data-selected': String(props.selected),
      onClick: () => emit('toggle-select', (props.coin as Coin).id),
    }, (props.coin as Coin).name),
  }),
}))
vi.mock('@/components/collection/CoinListTable.vue', () => ({
  default: defineComponent({
    props: { coins: { type: Array, required: true }, showPresentation: { type: Boolean, default: true } },
    setup: (props) => () => h('table', { 'data-testid': 'list', 'data-presentation': String(props.showPresentation) }, `${props.coins.length}`),
  }),
}))

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots }) => () => h('a', { href: props.to }, slots.default?.()),
})

function banknote(id: number, name: string, country = 'Portugal', extra: Partial<Coin> = {}): Coin {
  return { id, name, kind: 'banknote', country, images: [], ...extra } as unknown as Coin
}

function mountPage() {
  return mount(BanknotesPage, { global: { stubs: { RouterLink: RouterLinkStub, 'router-link': RouterLinkStub } } })
}

describe('BanknotesPage', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    localStorage.clear()
    mocks.route.query = {}
    mocks.getCoins.mockResolvedValue({ data: { coins: [banknote(1, '1000 escudos'), banknote(2, '500 escudos')], total: 2 } })
    mocks.getCoinCountries.mockResolvedValue({ data: [{ country: 'Portugal', count: 2 }] })
    mocks.getCollectionSummary.mockResolvedValue({ data: { classifications: [] } })
  })

  afterEach(() => {
    vi.clearAllMocks()
    vi.useRealTimers()
  })

  it('lists only banknotes of the collection, without the numbers of the Collection page', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(mocks.getCoins).toHaveBeenCalledWith(expect.objectContaining({ kind: 'banknote', wishlist: 'false', page: 1, limit: 100, sort: 'updated_at', order: 'desc' }))
    expect(mocks.getCoinCountries).toHaveBeenCalledWith('banknote')
    expect(wrapper.findAll('[data-testid="card"]').map((card) => card.text())).toEqual(['1000 escudos', '500 escudos'])
    expect(wrapper.find('[data-testid="banknotes-kpi-banknotes"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="banknotes-add"]').attributes('href')).toBe('/banknotes/add')
    expect(wrapper.get('h1').text()).toBe('Galeria de notas')
  })

  it('filters by country and search, and remembers the list layout', async () => {
    vi.useFakeTimers()
    const wrapper = mountPage()
    await flushPromises()

    await wrapper.get('[data-testid="banknotes-country"]').trigger('click')
    const options = [...document.querySelectorAll<HTMLElement>('[data-testid="country-select-option"]')]
    expect(options.map((option) => option.textContent?.trim())).toEqual(['Todos os países', '🇵🇹Portugal (2)'])
    options[1]!.click()
    await flushPromises()
    expect(mocks.getCoins).toHaveBeenLastCalledWith(expect.objectContaining({ kind: 'banknote', country: 'Portugal' }))

    await wrapper.get('input[type="text"]').setValue('escudos')
    vi.advanceTimersByTime(300)
    await flushPromises()
    expect(mocks.getCoins).toHaveBeenLastCalledWith(expect.objectContaining({ search: 'escudos', country: 'Portugal' }))

    const listButton = wrapper.findAll('button').find((button) => button.text() === 'Lista')!
    await listButton.trigger('click')
    expect(wrapper.find('[data-testid="list"]').exists()).toBe(true)
    expect(localStorage.getItem('banknotes:layout')).toBe('list')
  })

  it('invites to add the first banknote when there is none', async () => {
    mocks.getCoins.mockResolvedValue({ data: { coins: [], total: 0 } })
    const wrapper = mountPage()
    await flushPromises()

    const empty = wrapper.get('[data-testid="banknotes-empty"]')
    expect(empty.text()).toContain('Ainda não tem notas de banco')
    expect(empty.get('a').attributes('href')).toBe('/banknotes/add')
  })

  it("shows only the duplicates to trade when opened from the Collection page's, until the filter is removed", async () => {
    mocks.route.query = { repeated: 'true' }
    const wrapper = mountPage()
    await flushPromises()

    expect(mocks.getCoins).toHaveBeenLastCalledWith(expect.objectContaining({ kind: 'banknote', repeated: 'true' }))
    const chip = wrapper.get('[data-testid="banknotes-filter-repeated"]')
    expect(chip.text()).toBe('Só repetidas para troca')
    await chip.trigger('click')
    await flushPromises()
    expect(mocks.replace).toHaveBeenCalledWith({ query: {} })
    expect(mocks.getCoins).toHaveBeenLastCalledWith(expect.objectContaining({ repeated: undefined }))
    expect(wrapper.find('[data-testid="banknotes-filter-repeated"]').exists()).toBe(false)
  })

  it('shows the banknotes country by country, every page of them, and the list without presentation', async () => {
    const first = Array.from({ length: 100 }, (_, index) => banknote(index + 1, `Nota ${index + 1}`, index % 2 ? 'Angola' : 'Portugal'))
    mocks.getCoins
      .mockResolvedValueOnce({ data: { coins: first, total: 102 } })
      .mockResolvedValueOnce({ data: { coins: [banknote(101, 'Sem país', ''), banknote(102, '50 Dólares', 'Estados Unidos')], total: 102 } })
    const wrapper = mountPage()
    await flushPromises()

    expect(mocks.getCoins).toHaveBeenNthCalledWith(2, expect.objectContaining({ page: 2, limit: 100 }))
    const groups = wrapper.findAll('[data-testid="banknotes-group"]')
    expect(groups.map((group) => group.get('h2').text().replace(/\s+/g, ' '))).toEqual([
      '🇦🇴 Angola 50 notas', '🇺🇸 Estados Unidos 1 nota', '🇵🇹 Portugal 50 notas', 'Sem país 1 nota',
    ])
    expect(groups[1]!.findAll('[data-testid="card"]').map((card) => card.text())).toEqual(['50 Dólares'])

    await wrapper.findAll('button').find((button) => button.text() === 'Lista')!.trigger('click')
    expect(wrapper.findAll('[data-testid="list"]').map((list) => list.attributes('data-presentation'))).toEqual(['false', 'false', 'false', 'false'])
  })

  it('selects banknotes, every one at once, to add to a set or delete, as in the coin gallery', async () => {
    mocks.getSets.mockResolvedValue({ data: { sets: [{ id: 3, name: 'Notas de Portugal', color: '#c9a84c', setType: 'standard' }, { id: 4, name: 'Catálogo', color: '#fff', setType: 'checklist' }] } })
    mocks.bulkAction.mockResolvedValue({ data: {} })
    const wrapper = mountPage()
    await flushPromises()
    expect(wrapper.find('[data-testid="banknotes-select-bar"]').exists()).toBe(false)

    await wrapper.get('[data-testid="banknotes-select"]').trigger('click')
    await flushPromises()
    expect(wrapper.get('[data-testid="banknotes-select"]').attributes('aria-pressed')).toBe('true')
    expect(wrapper.findAll('[data-testid="card"]').map((card) => card.attributes('data-selectable'))).toEqual(['true', 'true'])
    await wrapper.findAll('[data-testid="card"]')[1]!.trigger('click')
    expect(wrapper.get('[data-testid="banknotes-selected-count"]').text()).toBe('1 selecionada')
    await wrapper.findAll('[data-testid="banknotes-select-bar"] button').find((button) => button.text() === 'Selecionar tudo')!.trigger('click')
    expect(wrapper.text()).toContain('2 notas selecionadas')

    // Only the standard sets, as in the coin gallery.
    const picker = wrapper.findComponent({ name: 'BulkSetPickerModal' })
    expect((picker.props('sets') as Array<{ id: number }>).map((set) => set.id)).toEqual([3])
    picker.vm.$emit('select', 3)
    await flushPromises()
    expect(mocks.bulkAction).toHaveBeenCalledWith([1, 2], 'set', { setId: 3 })
    expect(wrapper.find('[data-testid="banknotes-select-bar"]').exists()).toBe(false)

    // Deleted after confirming; not when the user says no.
    await wrapper.get('[data-testid="banknotes-select"]').trigger('click')
    await wrapper.findAll('[data-testid="banknotes-select-bar"] button').find((button) => button.text() === 'Selecionar tudo')!.trigger('click')
    mocks.confirm.mockResolvedValueOnce(false).mockResolvedValueOnce(true)
    const remove = () => wrapper.findAll('button').find((button) => button.text() === 'Eliminar')!
    await remove().trigger('click')
    await flushPromises()
    expect(mocks.bulkAction).toHaveBeenCalledTimes(1)
    await remove().trigger('click')
    await flushPromises()
    expect(mocks.confirm).toHaveBeenLastCalledWith('Eliminar 2 notas? Esta ação não pode ser anulada.', expect.objectContaining({ variant: 'danger' }))
    expect(mocks.bulkAction).toHaveBeenLastCalledWith([1, 2], 'delete')
    expect(wrapper.find('[data-testid="banknotes-select-bar"]').exists()).toBe(false)
  })

  it('views all the banknotes, or by country, monetary system, year or rarity, and remembers the choice', async () => {
    mocks.getCoins.mockResolvedValue({ data: { coins: [
      banknote(1, '20 Escudos 1971', 'Portugal', { year: 1971, rarity: 'common', monetarySystem: 'Escudo (1911-2001)' }),
      banknote(2, '100 Escudos 1972', 'Angola', { year: 1972, rarity: 'scarce', monetarySystem: 'Escudo (1958-1977)' }),
      banknote(3, '20 Escudos 1978', 'Portugal', { year: 1978, rarity: 'common', monetarySystem: 'Escudo (1911-2001)' }),
    ], total: 3 } })
    const wrapper = mountPage()
    await flushPromises()
    const headings = () => wrapper.findAll('[data-testid="banknotes-group"] h2').map((heading) => heading.text().replace(/\s+/g, ' '))

    const groupBy = wrapper.get('[data-testid="banknotes-group-by"]')
    expect(groupBy.findAll('option').map((option) => option.text())).toEqual(['Todas as notas', 'País', 'Sistema monetário', 'Ano', 'Raridade'])
    expect(headings()).toEqual(['🇦🇴 Angola 1 nota', '🇵🇹 Portugal 2 notas'])

    // All of them in one list, in the order asked for, without groups.
    await groupBy.setValue('none')
    expect(headings()).toEqual([])
    const sections = wrapper.findAll('[data-testid="banknotes-group"]')
    expect(sections).toHaveLength(1)
    expect(sections[0]!.attributes('aria-label')).toBe('Todas as notas')
    expect(sections[0]!.text()).toContain('20 Escudos 1971')
    expect(sections[0]!.text()).toContain('100 Escudos 1972')
    expect(localStorage.getItem('banknotes:groupBy')).toBe('none')

    await groupBy.setValue('monetarySystem')
    expect(headings()).toEqual(['Escudo (1911-2001) 2 notas', 'Escudo (1958-1977) 1 nota'])

    await groupBy.setValue('year')
    expect(headings()).toEqual(['1971 1 nota', '1972 1 nota', '1978 1 nota'])
    expect(wrapper.find('[data-testid="group-flag"]').exists()).toBe(false)
    expect(localStorage.getItem('banknotes:groupBy')).toBe('year')

    await groupBy.setValue('rarity')
    expect(headings()).toEqual(['Comum 2 notas', 'Escassa 1 nota'])
    // The banknotes are asked for once: grouping is done here.
    expect(mocks.getCoins).toHaveBeenCalledTimes(1)
  })

  it("filters by Numista's kind of banknote, offered only when Numista says one", async () => {
    const plain = mountPage()
    await flushPromises()
    expect(mocks.getCollectionSummary).toHaveBeenCalledWith('banknote')
    expect(plain.find('[data-testid="banknotes-classification"]').exists()).toBe(false)

    mocks.getCollectionSummary.mockResolvedValue({ data: { classifications: [
      { key: 'standard', label: 'Cédulas circulantes comuns', count: 12 },
      { key: 'Cédulas de emergência', label: 'Cédulas de emergência', count: 1 },
    ] } })
    const wrapper = mountPage()
    await flushPromises()
    const select = wrapper.get('[data-testid="banknotes-classification"]')
    expect(select.findAll('option').map((option) => option.text())).toEqual(['Todos os tipos', 'Circulantes comuns (12)', 'Cédulas de emergência (1)'])
    await select.setValue('standard')
    await flushPromises()
    expect(mocks.getCoins).toHaveBeenLastCalledWith(expect.objectContaining({ kind: 'banknote', classification: 'standard' }))
  })
})
