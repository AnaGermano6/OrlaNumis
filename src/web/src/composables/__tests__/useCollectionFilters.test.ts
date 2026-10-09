import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import { COIN_ERAS } from '@/types'
import { getCoinOptions, getSets, getTags } from '@/api/client'
import { resetCoinVocabulary } from '@/i18n/coinVocabulary'
import { useCollectionFilters } from '../useCollectionFilters'

vi.mock('@/api/client', () => ({
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
    expect(filters.eraOptions.value).toEqual(COIN_ERAS)
    wrapper.unmount()
  })

  it('uses the configured era choices', async () => {
    vi.mocked(getTags).mockResolvedValue({ data: { tags: [] } } as unknown as Awaited<ReturnType<typeof getTags>>)
    vi.mocked(getSets).mockResolvedValue({ data: { sets: [] } } as unknown as Awaited<ReturnType<typeof getSets>>)
    vi.mocked(getCoinOptions).mockResolvedValue({
      data: { categories: ['Roman'], eras: ['ancient', 'Republican'], labels: {} },
    } as unknown as Awaited<ReturnType<typeof getCoinOptions>>)

    let filters!: ReturnType<typeof useCollectionFilters>
    const wrapper = mount(defineComponent({
      setup() {
        filters = useCollectionFilters()
        return () => h('div')
      },
    }))
    await filters.fetchUserTags()

    expect(filters.eraOptions.value).toEqual(['ancient', 'Republican'])
    wrapper.unmount()
  })
})
