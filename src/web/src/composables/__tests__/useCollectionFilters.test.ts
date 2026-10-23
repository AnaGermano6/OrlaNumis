import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import { getCoinDenominations, getCoinOptions, getCoinPeriods, getCoins, getSets } from '@/api/client'
import { EXTINCT_COUNTRIES } from '@/utils/countryFilter'
import { resetCoinVocabulary } from '@/i18n/coinVocabulary'
import { useCollectionFilters } from '../useCollectionFilters'

vi.mock('@/api/client', () => ({
  getCoins: vi.fn(),
  getCoinDenominations: vi.fn(),
  getCoinOptions: vi.fn(),
  getCoinPeriods: vi.fn(),
  getSets: vi.fn(),
}))

describe('useCollectionFilters', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    sessionStorage.clear()
    vi.clearAllMocks()
    resetCoinVocabulary()
  })

  it('offers the standard sets as filters, even when coin options are unavailable, and filters by one', async () => {
    vi.mocked(getSets).mockResolvedValue({
      data: {
        sets: [
          { id: 2, name: 'Euro 2007', color: '#b08d57', setType: 'standard' },
          { id: 3, name: 'Smart', color: '#6b7280', setType: 'smart' },
        ],
      },
    } as Awaited<ReturnType<typeof getSets>>)
    vi.mocked(getCoins).mockResolvedValue({ data: { coins: [], total: 0 } } as unknown as Awaited<ReturnType<typeof getCoins>>)
    vi.mocked(getCoinOptions).mockRejectedValue({ response: { status: 500 } })

    let filters!: ReturnType<typeof useCollectionFilters>
    const wrapper = mount(defineComponent({
      setup() {
        filters = useCollectionFilters()
        return () => h('div')
      },
    }))
    await filters.fetchUserSets()

    expect(filters.userSets.value).toEqual([{ id: 2, name: 'Euro 2007', color: '#b08d57' }])
    filters.selectedSet.value = '2'
    await vi.waitFor(() => expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ set: '2' }))
    wrapper.unmount()
  })

  it("offers the chosen country's periods, filters by one, and starts again with another country", async () => {
    vi.mocked(getCoins).mockResolvedValue({ data: { coins: [], total: 0 } } as unknown as Awaited<ReturnType<typeof getCoins>>)
    vi.mocked(getCoinPeriods).mockResolvedValue({ data: [{ period: 'República Portuguesa 1986–2001', count: 48 }] } as unknown as Awaited<ReturnType<typeof getCoinPeriods>>)
    let filters!: ReturnType<typeof useCollectionFilters>
    const wrapper = mount(defineComponent({
      setup() {
        filters = useCollectionFilters()
        return () => h('div')
      },
    }))

    filters.selectedCountry.value = 'Portugal'
    await vi.waitFor(() => expect(filters.periodOptions.value).toHaveLength(1))
    expect(getCoinPeriods).toHaveBeenCalledWith('Portugal')
    filters.selectedPeriod.value = 'República Portuguesa 1986–2001'
    await vi.waitFor(() => expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ country: 'Portugal', period: 'República Portuguesa 1986–2001' }))

    vi.mocked(getCoinPeriods).mockResolvedValue({ data: [] } as unknown as Awaited<ReturnType<typeof getCoinPeriods>>)
    filters.selectedCountry.value = 'Espanha'
    await vi.waitFor(() => expect(filters.selectedPeriod.value).toBe(''))
    await vi.waitFor(() => expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ country: 'Espanha', period: undefined }))
    filters.selectedCountry.value = ''
    await vi.waitFor(() => expect(filters.periodOptions.value).toEqual([]))
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

    // And only the duplicates to trade, off when not asked for.
    expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ repeated: undefined })
    filters.selectedRepeated.value = true
    await vi.waitFor(() => expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ repeated: 'true', classification: 'non-circulating' }))
    wrapper.unmount()
  })

  it("offers the collection's face values and filters by one", async () => {
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
    await filters.fetchUserSets()

    expect(getCoinDenominations).toHaveBeenCalledWith('coin')
    expect(filters.denominationOptions.value.map((option) => option.denomination)).toEqual(['2 euro', '5 euro'])
    filters.selectedDenomination.value = '2 euro'
    await vi.waitFor(() => expect(vi.mocked(getCoins).mock.lastCall?.[0]).toMatchObject({ denomination: '2 euro', kind: 'coin' }))
    wrapper.unmount()
  })
})
