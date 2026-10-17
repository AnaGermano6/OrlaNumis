import { flushPromises, shallowMount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import SetDetailPage from '@/pages/SetDetailPage.vue'
import { buildRomanDenariusCore } from '@/test/fixtures/coins'

const mockAddCoinToSet = vi.fn()
const mockDeleteSet = vi.fn()
const mockGetCoins = vi.fn()
const mockGetCoinsInSet = vi.fn()
const mockGetSet = vi.fn()
const mockGetSetCompletion = vi.fn()
const mockGetSetChecklist = vi.fn()
const mockReorderSetCoins = vi.fn()
const mockRemoveCoinFromSet = vi.fn()
const mockUpdateSet = vi.fn()
const mockPush = vi.fn()

vi.mock('@/api/client', () => ({
  addCoinToSet: (...args: unknown[]) => mockAddCoinToSet(...args),
  deleteSet: (...args: unknown[]) => mockDeleteSet(...args),
  getCoins: (...args: unknown[]) => mockGetCoins(...args),
  getCoinsInSet: (...args: unknown[]) => mockGetCoinsInSet(...args),
  getSet: (...args: unknown[]) => mockGetSet(...args),
  getSetCompletion: (...args: unknown[]) => mockGetSetCompletion(...args),
  getSetChecklist: (...args: unknown[]) => mockGetSetChecklist(...args),
  reorderSetCoins: (...args: unknown[]) => mockReorderSetCoins(...args),
  removeCoinFromSet: (...args: unknown[]) => mockRemoveCoinFromSet(...args),
  updateSet: (...args: unknown[]) => mockUpdateSet(...args),
}))

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { id: '7' } }),
  useRouter: () => ({ push: mockPush }),
}))

const mockIsPwa = vi.hoisted(() => ({ value: false }))

vi.mock('@/composables/usePwa', () => ({
  usePwa: () => ({ isPwa: mockIsPwa.value }),
}))

const defaultSet = {
  id: 7,
  name: 'Twelve Caesars',
  color: '#c9a84c',
  setType: 'goal',
  coinCount: 13,
  totalValue: 1300,
  totalInvested: 900,
}

const defaultStubs = {
  SetCompletionChecklist: true,
  DesktopPageContext: {
    props: ['title'],
    template: '<div><slot /><slot name="actions" /><slot name="desktop" /></div>',
  },
  AppIconButton: {
    props: ['active', 'disabled', 'title'],
    emits: ['click'],
    template: '<button :disabled="disabled" :title="title" @click="$emit(\'click\')"><slot /></button>',
  },
}

function mockSetDetailLoad(coins = [
  buildRomanDenariusCore({ id: 1, name: 'Augustus Denarius', diameterMm: 19 }),
  buildRomanDenariusCore({ id: 2, name: 'Tiberius Denarius', diameterMm: null }),
], setOverrides: Record<string, unknown> = {}) {
  mockGetSet.mockResolvedValue({ data: { ...defaultSet, ...setOverrides } })
  mockGetCoinsInSet.mockResolvedValue({ data: { coins } })
  mockGetCoins.mockResolvedValue({ data: { coins: [], total: 0 } })
  mockGetSetCompletion.mockResolvedValue({
    data: { totalTargets: 12, completedTargets: 2, completionPercentage: 16.7, missingTargets: [] },
  })
}

describe('SetDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockIsPwa.value = false
    vi.stubGlobal('confirm', vi.fn(() => true))
    mockSetDetailLoad()
  })

  it('shows the set as an album of its coins, without the tray', async () => {
    const wrapper = shallowMount(SetDetailPage, {
      global: { stubs: defaultStubs },
    })
    await flushPromises()

    const album = wrapper.findComponent({ name: 'SetAlbum' })
    expect(album.props('slots')).toEqual([
      expect.objectContaining({ key: 'coin:1', title: 'Augustus Denarius', owned: true, coinId: 1 }),
      expect.objectContaining({ key: 'coin:2', title: 'Tiberius Denarius', owned: true, coinId: 2 }),
    ])
    expect(album.props('defaultGroup')).toBe('none')
    expect(album.props('selectable')).toBe(false)
    expect(wrapper.find('#set-tray-size-slider').exists()).toBe(false)
    expect(wrapper.findComponent({ name: 'MuseumTray' }).exists()).toBe(false)
  })

  it("leaves a goal set's wishlist coins as places to fill", async () => {
    mockSetDetailLoad([
      buildRomanDenariusCore({ id: 1, name: 'Owned Denarius', isWishlist: false }),
      buildRomanDenariusCore({ id: 2, name: 'Wishlist Denarius', isWishlist: true }),
    ])

    const wrapper = shallowMount(SetDetailPage, {
      global: { stubs: defaultStubs },
    })
    await flushPromises()

    expect(wrapper.findComponent({ name: 'SetAlbum' }).props('slots')).toEqual([
      expect.objectContaining({ coinId: 1, owned: true, wishlist: false }),
      expect.objectContaining({ coinId: 2, owned: false, wishlist: true }),
    ])
  })

  it("shows a checklist set's slots by year, what it lists, and that it keeps up by itself", async () => {
    mockSetDetailLoad([], { setType: 'checklist', checklist: { source: 'catalogue', country: 'PT', faceValue: 2 } })
    mockGetSetChecklist.mockResolvedValue({ data: { complete: false, slots: [
      { key: 'ecb:rotary', title: 'Rotary', country: 'PT', year: 2026, faceValue: 2, owned: false, imageUrl: 'https://ecb/rotary.jpg', pageUrl: 'https://ecb/page' },
      { key: 'ecb:gago', title: 'Gago Coutinho', country: 'PT', year: 2022, faceValue: 2, owned: true, coinId: 5, coinName: 'Mine', coinImage: 'coins/gago.jpg' },
    ] } })

    const wrapper = shallowMount(SetDetailPage, {
      global: { stubs: defaultStubs },
    })
    await flushPromises()

    expect(mockGetSetChecklist).toHaveBeenCalledWith(7)
    expect(mockGetSetCompletion).not.toHaveBeenCalled()
    const album = wrapper.findComponent({ name: 'SetAlbum' })
    expect(album.props('defaultGroup')).toBe('year')
    expect(album.props('slots')).toEqual([
      expect.objectContaining({ key: 'ecb:rotary', title: 'Rotary', detail: '€2', year: 2026, owned: false, imageUrl: 'https://ecb/rotary.jpg', pageUrl: 'https://ecb/page' }),
      expect.objectContaining({ key: 'ecb:gago', title: 'Gago Coutinho', owned: true, coinId: 5, coinImage: 'coins/gago.jpg' }),
    ])
    const info = wrapper.get('[data-testid="set-checklist-info"]').text()
    expect(info).toContain('Catalogue · Portugal · €2')
    expect(info).toContain('Kept up to date by itself')
    expect(wrapper.find('[data-testid="set-checklist-partial"]').exists()).toBe(true)
    // Its coins are the catalogue's: none are added by hand.
    await wrapper.findAll('button').find((node) => node.text().includes('Actions'))!.trigger('click')
    expect(wrapper.findAll('button').some((node) => node.text().includes('Add Coin'))).toBe(false)
  })

  it("edits a smart set's filters", async () => {
    mockSetDetailLoad([], { setType: 'smart', smartCriteria: { operator: 'and', rules: [
      { field: 'country', op: 'eq', value: 'Portugal' },
      { field: 'isWishlist', op: 'eq', value: false },
    ] } })
    mockUpdateSet.mockResolvedValue({ data: {} })

    const wrapper = shallowMount(SetDetailPage, {
      global: { stubs: defaultStubs },
    })
    await flushPromises()
    expect(wrapper.get('[data-testid="set-smart-info"]').text()).toContain('Portugal')
    expect(wrapper.text()).toContain('No coin of your collection matches these filters yet.')

    await wrapper.findAll('button').find((node) => node.text().includes('Actions'))!.trigger('click')
    await wrapper.findAll('button').find((node) => node.text().includes('Edit Set'))!.trigger('click')
    const filters = wrapper.findComponent({ name: 'SetSimpleFilters' })
    expect(filters.props('modelValue')).toEqual(expect.objectContaining({ country: 'Portugal', yearFrom: null }))
    filters.vm.$emit('update:modelValue', { ...filters.props('modelValue'), yearFrom: 2002, yearTo: 2010 })
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(mockUpdateSet).toHaveBeenCalledWith(7, expect.objectContaining({ smartCriteria: { operator: 'and', rules: [
      { field: 'country', op: 'eq', value: 'Portugal' },
      { field: 'year', op: 'between', value: [2002, 2010] },
      { field: 'isWishlist', op: 'eq', value: false },
    ] } }))
  })

  it('says when a smart set has rules the filters cannot show', async () => {
    mockSetDetailLoad([], { setType: 'smart', smartCriteria: { operator: 'or', rules: [{ field: 'country', op: 'eq', value: 'Portugal' }] } })
    mockUpdateSet.mockResolvedValue({ data: {} })
    const wrapper = shallowMount(SetDetailPage, {
      global: { stubs: defaultStubs },
    })
    await flushPromises()

    await wrapper.findAll('button').find((node) => node.text().includes('Actions'))!.trigger('click')
    await wrapper.findAll('button').find((node) => node.text().includes('Edit Set'))!.trigger('click')
    expect(wrapper.findComponent({ name: 'SetSimpleFilters' }).exists()).toBe(false)
    expect(wrapper.find('[data-testid="edit-advanced-rules"]').exists()).toBe(true)
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mockUpdateSet.mock.calls[0]?.[1]).not.toHaveProperty('smartCriteria')
  })

  it('opens analytics and trends page from actions menu', async () => {
    const wrapper = shallowMount(SetDetailPage, {
      global: { stubs: defaultStubs },
    })
    await flushPromises()

    const actionsButton = wrapper.findAll('button').find((node) => node.text().includes('Actions'))
    expect(actionsButton?.exists()).toBe(true)
    await actionsButton!.trigger('click')

    const infoAction = wrapper.findAll('button').find((node) => node.text().includes('Analytics & Value Trend'))
    expect(infoAction?.exists()).toBe(true)
    await infoAction!.trigger('click')

    expect(mockPush).toHaveBeenCalledWith({ name: 'set-insights', params: { id: 7 } })
  })

})
