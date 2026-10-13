import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h, nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { useCoinsStore } from '@/stores/coins'
import { useCollectionFilters } from '../useCollectionFilters'

const mocks = vi.hoisted(() => ({
  getCoins: vi.fn(),
  getSets: vi.fn(),
  getTags: vi.fn(),
  getCoinOptions: vi.fn(),
  getCoinCountries: vi.fn(),
  fillCountriesFromNumista: vi.fn(),
}))
vi.mock('@/api/client', () => mocks)

function mountFilters() {
  let filters!: ReturnType<typeof useCollectionFilters>
  const wrapper = mount(defineComponent({
    setup() {
      filters = useCollectionFilters()
      return () => h('div')
    },
  }))
  return { filters, wrapper }
}

describe('useCollectionFilters grouping', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    sessionStorage.clear()
    Object.values(mocks).forEach((mock) => mock.mockReset())
    mocks.getCoins.mockResolvedValue({ data: { coins: [], total: 0 } })
    mocks.getSets.mockResolvedValue({ data: { sets: [] } })
    mocks.getTags.mockResolvedValue({ data: { tags: [] } })
    mocks.getCoinOptions.mockResolvedValue({ data: { categories: [], eras: [], labels: {} } })
    mocks.getCoinCountries.mockResolvedValue({ data: [{ country: 'Portugal', count: 3 }] })
  })

  it('sends the new filters and loads every coin when grouping', async () => {
    const store = useCoinsStore()
    const fetchAll = vi.spyOn(store, 'fetchAllCoins')
    const fetchPage = vi.spyOn(store, 'fetchCoins')
    const { filters, wrapper } = mountFilters()
    await filters.fetchUserTags()
    expect(filters.countryOptions.value).toEqual([{ country: 'Portugal', count: 3 }])
    // The gallery is the coins'; banknotes have their own.
    expect(mocks.getCoinCountries).toHaveBeenCalledWith('coin')

    filters.selectedPresentation.value = 'coincard'
    await nextTick()
    expect(fetchPage).toHaveBeenLastCalledWith(expect.objectContaining({ kind: 'coin', presentation: 'coincard', page: 1 }))

    mocks.fillCountriesFromNumista.mockResolvedValue({ data: { status: 'ok', updated: 0, remaining: 0, failed: 0 } })
    filters.groupBy.value = 'year'
    await nextTick()
    expect(fetchAll).toHaveBeenLastCalledWith(expect.objectContaining({ presentation: 'coincard', wishlist: 'false', sold: 'false' }))
    expect(fetchAll.mock.lastCall![0]).not.toHaveProperty('page')
    expect(localStorage.getItem('collection:groupBy')).toBe('year')
    expect(mocks.fillCountriesFromNumista).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('fills countries from Numista once per session when grouping by country', async () => {
    mocks.fillCountriesFromNumista
      .mockResolvedValueOnce({ data: { status: 'ok', updated: 25, remaining: 3, failed: 0 } })
      .mockResolvedValueOnce({ data: { status: 'ok', updated: 3, remaining: 0, failed: 0 } })
    const { filters, wrapper } = mountFilters()

    filters.groupBy.value = 'country'
    await nextTick()
    expect(filters.fillingCountries.value).toBe(true)
    await flushPromises()

    expect(mocks.fillCountriesFromNumista).toHaveBeenCalledTimes(2)
    expect(mocks.fillCountriesFromNumista).toHaveBeenCalledWith('en')
    expect(filters.fillingCountries.value).toBe(false)
    expect(mocks.getCoinCountries).toHaveBeenCalled()

    filters.groupBy.value = 'none'
    await nextTick()
    filters.groupBy.value = 'country'
    await flushPromises()
    expect(mocks.fillCountriesFromNumista).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })
})
