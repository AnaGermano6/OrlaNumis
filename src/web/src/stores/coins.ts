import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Coin, CoinMutationPayload, StatsResponse, ValueSnapshot, CollectionHealthSummary, CoinHealthItem } from '@/types'
import * as api from '@/api/client'

// Grouped views load the whole filtered collection in pages of 100, up to a limit.
const ALL_COINS_PAGE = 100
const ALL_COINS_LIMIT = 2000

export const useCoinsStore = defineStore('coins', () => {
  const coins = ref<Coin[]>([])
  const currentCoin = ref<Coin | null>(null)
  const total = ref(0)
  const loading = ref(false)
  const stats = ref<StatsResponse | null>(null)
  const valueHistory = ref<ValueSnapshot[]>([])
  const selectedEra = ref('')
  const searchQuery = ref('')
  const activeSortKey = ref('')
  const galleryIndex = ref(0)
  const selectedCountry = ref('')
  const selectedPresentation = ref('')
  const selectedRarity = ref('')
  /** '', 'current' or 'demonetized'. */
  const selectedCirculation = ref('')
  /** '', 'true' (commemorative) or 'false'. */
  const selectedCommemorative = ref('')
  /** '' or a Numista kind of coin (CollectionClassification key). */
  const selectedClassification = ref('')
  /** True when a grouped view hit ALL_COINS_LIMIT and shows only part of the collection. */
  const truncated = ref(false)
  let allCoinsGeneration = 0
  const collectionHealth = ref<CollectionHealthSummary | null>(null)
  const coinHealthList = ref<CoinHealthItem[]>([])
  const healthLoading = ref(false)

  async function fetchCoins(params?: {
    category?: string
    era?: string
    country?: string
    extinct?: string
    circulation?: string
    commemorative?: string
    classification?: string
    presentation?: string
    rarity?: string
    search?: string
    wishlist?: string
    sold?: string
    tag?: string
    set?: string
    page?: number
    sort?: string
    order?: string
    seed?: number
  }) {
    // A newer paged load supersedes any grouped load still in flight.
    allCoinsGeneration += 1
    loading.value = true
    try {
      const res = await api.getCoins(params)
      coins.value = res.data.coins || []
      total.value = res.data.total
      truncated.value = false
    } finally {
      loading.value = false
    }
  }

  /**
   * Loads every coin matching the filters (pages of 100), for grouped views
   * that need the whole collection at once. Stops at ALL_COINS_LIMIT.
   */
  async function fetchAllCoins(params: Omit<NonNullable<Parameters<typeof api.getCoins>[0]>, 'page' | 'limit'>) {
    const generation = ++allCoinsGeneration
    loading.value = true
    try {
      const all: Coin[] = []
      let reported = 0
      for (let nextPage = 1; all.length < ALL_COINS_LIMIT; nextPage += 1) {
        const res = await api.getCoins({ ...params, page: nextPage, limit: ALL_COINS_PAGE })
        if (generation !== allCoinsGeneration) return
        const pageCoins = res.data.coins || []
        reported = res.data.total
        all.push(...pageCoins)
        if (!pageCoins.length || all.length >= reported) break
      }
      coins.value = all
      total.value = reported
      truncated.value = all.length < reported
    } finally {
      if (generation === allCoinsGeneration) loading.value = false
    }
  }

  async function fetchCoin(id: number) {
    loading.value = true
    try {
      const res = await api.getCoin(id)
      currentCoin.value = res.data
    } finally {
      loading.value = false
    }
  }

  async function addCoin(coin: CoinMutationPayload) {
    const res = await api.createCoin(coin)
    return res.data
  }

  async function editCoin(id: number, coin: CoinMutationPayload) {
    const res = await api.updateCoin(id, coin)
    return res.data
  }

  async function removeCoin(id: number) {
    await api.deleteCoin(id)
    coins.value = coins.value.filter((c) => c.id !== id)
  }

  async function fetchStats() {
    const res = await api.getStats()
    stats.value = res.data
  }

  async function fetchValueHistory() {
    const res = await api.getValueHistory()
    valueHistory.value = res.data
  }

  async function fetchCollectionHealth() {
    healthLoading.value = true
    try {
      const res = await api.getCollectionHealthSummary()
      collectionHealth.value = res.data
    } catch (error) {
      collectionHealth.value = null
      throw error
    } finally {
      healthLoading.value = false
    }
  }

  async function fetchCoinHealthList(scope?: 'all' | 'needs_attention', page = 1, limit = 25) {
    healthLoading.value = true
    try {
      const res = await api.getCoinHealthList({ scope, page, limit })
      coinHealthList.value = res.data.coins
      return res.data
    } finally {
      healthLoading.value = false
    }
  }

  return {
    coins,
    currentCoin,
    total,
    loading,
    stats,
    valueHistory,
    selectedEra,
    searchQuery,
    activeSortKey,
    galleryIndex,
    selectedCountry,
    selectedPresentation,
    selectedRarity,
    selectedCirculation,
    selectedCommemorative,
    selectedClassification,
    truncated,
    fetchAllCoins,
    collectionHealth,
    coinHealthList,
    healthLoading,
    fetchCoins,
    fetchCoin,
    addCoin,
    editCoin,
    removeCoin,
    fetchStats,
    fetchValueHistory,
    fetchCollectionHealth,
    fetchCoinHealthList,
  }
})
