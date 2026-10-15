import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import { getCoinDenominations, getCoinOptions, getCoins, getSets, getTags } from '@/api/client'
import { EXTINCT_COUNTRIES } from '@/utils/countryFilter'
import { resetCoinVocabulary } from '@/i18n/coinVocabulary'
import { useCollectionFilters } from '../useCollectionFilters'

vi.mock('@/api/client', () => ({
  getCoins: vi.fn(),
  getCoinDenominations: vi.fn(),
  getCoinOptions: vi.fn(),
  getSets: vi.fn(),
  getTags: vi.fn(),
}))

describe('useCollectionFilters', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    sessionStorage.clear()
    vi.clearAllMocks()
    resetCoinVocabulary()
  })

  it('keeps tag and set filters when coin options are unavailable', async () => {
    vi.mocked(getTags).mockResolvedValue({
      data: {
        tags: [
          { id: 1, name: 'Favorites', color: '#c9a84c' },
        ],
      },
    } as Awaited<ReturnType<typeof getTags>>)
    vi.mocked(getSets).mockResolvedValue({
      data: {
        sets: [
          { id: 2, name: 'Twelve Caesars', color: '#b08d57', setType: 'standard' },
        ],
      },
    } as Awaited<ReturnType<typeof getSets>>)
    vi.mocked(getCoinOptions).mockRejectedValue({ response: { status: 500 } })

    let filters!: ReturnType<typeof useCollectionFilters>
    const wrapper = mount(defineComponent({
      setup() {
        filters = useCollectionFilters()
        return () => h('div')
      },
    }))
    await filters.fetchUserTags()

    expect(filters.userTags.value).toEqual([
      {
        id: 1,
        name: 'Favorites',
        color: '#c9a84c',
        filterValue: 'tag:1',
        source: 'tag',
      },
      {
        id: 2,
        name: 'Twelve Caesars',
        color: '#b08d57',
        filterValue: 'set:2',
        source: 'set',
      },
    ])
    wrapper.unmount()
  })

  it('asks for the coins of every extinct country rather than one country', async () => {
    vi.mocked(getCoins).mockResolvedValue({ data: { coins: [], total: 0 } } as unknown as Awaited<ReturnType<typeof getCoins>>)
    let filters!: ReturnType<typeof useCollectionFilters>
    const wrapper = mount(defineComponent({
      setup() {
        filters = useCollectionFilters()
        return () => h('div')
      },
    }))

    filters.selectedCountry.value = EXTINCT_COUNTRIES
    await vi.waitFor(() => expect(getCoins).toHaveBeenCalled())
    expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ extinct: 'true', country: undefined })

    filters.selectedCountry.value = 'Portugal'
    await vi.waitFor(() => expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ country: 'Portugal', extinct: undefined }))

    // The legal tender filter goes to the server as is.
    filters.selectedCirculation.value = 'current'
    await vi.waitFor(() => expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ circulation: 'current', country: 'Portugal' }))

    // So does the commemorative one, together with the others.
    filters.selectedCommemorative.value = 'true'
    await vi.waitFor(() => expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ commemorative: 'true', circulation: 'current' }))

    // And the Numista kind of coin.
    filters.selectedClassification.value = 'non-circulating'
    await vi.waitFor(() => expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ classification: 'non-circulating', commemorative: 'true' }))
    wrapper.unmount()
  })

  it("offers the collection's face values and filters by one", async () => {
    vi.mocked(getTags).mockResolvedValue({ data: { tags: [] } } as unknown as Awaited<ReturnType<typeof getTags>>)
    vi.mocked(getSets).mockResolvedValue({ data: { sets: [] } } as unknown as Awaited<ReturnType<typeof getSets>>)
    vi.mocked(getCoins).mockResolvedValue({ data: { coins: [], total: 0 } } as unknown as Awaited<ReturnType<typeof getCoins>>)
    vi.mocked(getCoinDenominations).mockResolvedValue({
      data: [{ denomination: '2 euro', count: 92 }, { denomination: '5 euro', count: 36 }],
    } as unknown as Awaited<ReturnType<typeof getCoinDenominations>>)

    let filters!: ReturnType<typeof useCollectionFilters>
    const wrapper = mount(defineComponent({
      setup() {
        filters = useCollectionFilters()
        return () => h('div')
      },
    }))
    await filters.fetchUserTags()

    expect(getCoinDenominations).toHaveBeenCalledWith('coin')
    expect(filters.denominationOptions.value.map((option) => option.denomination)).toEqual(['2 euro', '5 euro'])
    filters.selectedDenomination.value = '2 euro'
    await vi.waitFor(() => expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ denomination: '2 euro', kind: 'coin' }))
    wrapper.unmount()
  })
})
