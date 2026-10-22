import { ref, onBeforeUnmount, onMounted, watch } from 'vue'
import { EXTINCT_COUNTRIES } from '@/utils/countryFilter'
import { useCoinsStore } from '@/stores/coins'
import { fillCountriesFromNumista, getCoinCountries, getCoinDenominations, getSets, type CoinCountryCount, type CoinDenominationCount } from '@/api/client'
import { COIN_GROUPINGS, type CoinGroupBy } from '@/utils/coinAttributes'
import { numistaLang } from '@/utils/numistaCatalog'
import { loadCoinVocabulary } from '@/i18n/coinVocabulary'
import { usePwa } from '@/composables/usePwa'
import { normalizeCoinSetType } from '@/types'
import type { CollectionSetOption } from '@/types'

const RANDOM_SEED_KEY = 'coins:randomSeed'
const GROUP_BY_KEY = 'collection:groupBy'
const COUNTRY_FILL_KEY = 'coins:countryFillDone'

function readGroupBy(): CoinGroupBy {
  try {
    const saved = localStorage.getItem(GROUP_BY_KEY) as CoinGroupBy | null
    if (saved && COIN_GROUPINGS.includes(saved)) return saved
  } catch {
    // Storage unavailable: no grouping.
  }
  return 'none'
}
const PWA_RESUME_THRESHOLD_MS = 30_000

function generateRandomSeed(): number {
  return Math.floor(Math.random() * 1_000_000) + 1
}

// Module-level flag: true once any consumer of this composable has mounted in
// the current app lifetime. Resets when JS modules reload (i.e., PWA relaunch
// or full page reload).
let appLifecycleStarted = false

export function useCollectionFilters() {
  const store = useCoinsStore()
  const { isPwa } = usePwa()

  // On the first composable mount of a PWA's lifetime, treat it as a fresh
  // launch and clear any cached random seed so the next random fetch reshuffles.
  if (isPwa && !appLifecycleStarted) {
    sessionStorage.removeItem(RANDOM_SEED_KEY)
    store.galleryIndex = 0
  }
  appLifecycleStarted = true

  const selectedDenomination = ref(store.selectedDenomination)
  const search = ref(store.searchQuery)
  const page = ref(1)
  const sortKey = ref(store.activeSortKey || localStorage.getItem('defaultSort') || 'updated_at_desc')
  const selectedSet = ref('')
  const userSets = ref<CollectionSetOption[]>([])
  const denominationOptions = ref<CoinDenominationCount[]>([])
  const selectedCountry = ref(store.selectedCountry)
  const selectedPresentation = ref(store.selectedPresentation)
  const selectedRarity = ref(store.selectedRarity)
  const selectedCirculation = ref(store.selectedCirculation)
  const selectedCommemorative = ref(store.selectedCommemorative)
  const selectedClassification = ref(store.selectedClassification)
  const selectedRepeated = ref(store.selectedRepeated)
  const groupBy = ref<CoinGroupBy>(readGroupBy())
  const countryOptions = ref<CoinCountryCount[]>([])
  const fillingCountries = ref(false)

  let debounceTimer: ReturnType<typeof setTimeout>
  let hiddenAt = 0

  async function fetchUserSets() {
    try {
      const [setRes] = await Promise.all([getSets(), loadCoinVocabulary(), loadCountryOptions(), loadDenominationOptions()])
      userSets.value = (setRes.data?.sets ?? [])
        .filter((set) => normalizeCoinSetType(set.setType) === 'standard')
        .map((set) => ({ id: set.id, name: set.name, color: set.color }))
    } catch { /* ignore */ }
  }

  async function loadDenominationOptions() {
    try {
      const res = await getCoinDenominations('coin')
      denominationOptions.value = res.data ?? []
    } catch {
      denominationOptions.value = []
    }
  }

  async function loadCountryOptions() {
    try {
      const res = await getCoinCountries('coin')
      countryOptions.value = res.data ?? []
    } catch {
      countryOptions.value = []
    }
  }

  // Coins linked to Numista but saved without a country get it from their
  // Numista type the first time the collection is grouped by country in a
  // browser session (a few Numista requests, cached server-side for 30 days).
  async function fillCountriesOnce() {
    try {
      if (sessionStorage.getItem(COUNTRY_FILL_KEY)) return
      sessionStorage.setItem(COUNTRY_FILL_KEY, '1')
    } catch {
      // Without session storage the fill may run again on a later visit; it is cheap.
    }
    fillingCountries.value = true
    let updated = 0
    try {
      for (let round = 0; round < 8; round += 1) {
        const res = await fillCountriesFromNumista(numistaLang())
        updated += res.data.updated
        if (res.data.status !== 'ok' || res.data.remaining === 0 || res.data.updated === 0) break
      }
    } catch {
      // Countries stay as they are; the user can still set them by hand.
    } finally {
      fillingCountries.value = false
    }
    if (updated > 0) {
      await loadCountryOptions()
      loadCoins()
    }
  }

  // The list's filters and order, as the server takes them (without the page).
  function listParams() {
    const [sort, order] = sortKey.value.split('_').length === 3
      ? [sortKey.value.split('_').slice(0, 2).join('_'), sortKey.value.split('_')[2]]
      : [sortKey.value.split('_')[0], sortKey.value.split('_')[1]]
    // For random sort, generate a per-session seed (stable across pagination within a session)
    let seed: number | undefined
    if (sort === 'random') {
      const cached = sessionStorage.getItem(RANDOM_SEED_KEY)
      if (cached) {
        seed = parseInt(cached, 10)
      } else {
        seed = generateRandomSeed()
        sessionStorage.setItem(RANDOM_SEED_KEY, String(seed))
      }
    }

    const params = {
      // Banknotes have their own gallery.
      kind: 'coin' as const,
      denomination: selectedDenomination.value || undefined,
      country: selectedCountry.value && selectedCountry.value !== EXTINCT_COUNTRIES ? selectedCountry.value : undefined,
      extinct: selectedCountry.value === EXTINCT_COUNTRIES ? 'true' : undefined,
      presentation: selectedPresentation.value || undefined,
      rarity: selectedRarity.value || undefined,
      circulation: selectedCirculation.value || undefined,
      commemorative: selectedCommemorative.value || undefined,
      classification: selectedClassification.value || undefined,
      repeated: selectedRepeated.value ? 'true' : undefined,
      search: search.value || undefined,
      set: selectedSet.value || undefined,
      wishlist: 'false',
      sort,
      order,
      seed,
    }
    return params
  }

  function loadCoins() {
    store.selectedDenomination = selectedDenomination.value
    store.selectedCountry = selectedCountry.value
    store.selectedPresentation = selectedPresentation.value
    store.selectedRarity = selectedRarity.value
    store.selectedCirculation = selectedCirculation.value
    store.selectedCommemorative = selectedCommemorative.value
    store.selectedClassification = selectedClassification.value
    store.selectedRepeated = selectedRepeated.value
    store.searchQuery = search.value
    const params = listParams()
    // Grouped views show every matching coin in sections, so they load all pages.
    if (groupBy.value !== 'none') void store.fetchAllCoins(params)
    else void store.fetchCoins({ ...params, page: page.value })
  }

  watch([selectedCountry, selectedPresentation, selectedRarity, selectedCirculation, selectedCommemorative, selectedClassification, selectedRepeated], () => {
    page.value = 1
    loadCoins()
  })

  watch(groupBy, (value) => {
    try {
      localStorage.setItem(GROUP_BY_KEY, value)
    } catch {
      // The choice still applies for this visit.
    }
    page.value = 1
    loadCoins()
    if (value === 'country') void fillCountriesOnce()
  })

  watch(selectedSet, () => {
    page.value = 1
    loadCoins()
  })

  watch(selectedDenomination, () => {
    page.value = 1
    loadCoins()
  })

  watch(search, () => {
    clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => {
      page.value = 1
      loadCoins()
    }, 300)
  })

  watch(page, loadCoins)
  watch(sortKey, () => {
    store.activeSortKey = sortKey.value
    page.value = 1
    // Reset the random seed when the user re-selects Random so the order shuffles.
    const sort = sortKey.value.split('_').slice(0, -1).join('_')
    if (sort === 'random') {
      sessionStorage.setItem(RANDOM_SEED_KEY, String(generateRandomSeed()))
    }
    loadCoins()
  })

  // In PWA mode, when the app resumes after being backgrounded for a while,
  // treat it as a relaunch: reshuffle if currently sorted by Random.
  function onVisibilityChange() {
    if (document.visibilityState === 'hidden') {
      hiddenAt = Date.now()
      return
    }
    if (document.visibilityState !== 'visible' || hiddenAt === 0) return
    const hiddenFor = Date.now() - hiddenAt
    hiddenAt = 0
    if (!isPwa || hiddenFor < PWA_RESUME_THRESHOLD_MS) return
    const sort = sortKey.value.split('_').slice(0, -1).join('_')
    if (sort !== 'random') return
    sessionStorage.setItem(RANDOM_SEED_KEY, String(generateRandomSeed()))
    store.galleryIndex = 0
    page.value = 1
    loadCoins()
  }

  if (isPwa) {
    onMounted(() => {
      document.addEventListener('visibilitychange', onVisibilityChange)
    })
  }

  onBeforeUnmount(() => {
    clearTimeout(debounceTimer)
    if (isPwa) {
      document.removeEventListener('visibilitychange', onVisibilityChange)
    }
  })

  return {
    selectedDenomination,
    search,
    page,
    sortKey,
    selectedSet,
    userSets,
    denominationOptions,
    selectedCountry,
    selectedPresentation,
    selectedCirculation,
    selectedCommemorative,
    selectedClassification,
    selectedRepeated,
    selectedRarity,
    groupBy,
    countryOptions,
    fillingCountries,
    fetchUserSets,
    loadCoins,
    loadCountryOptions,
    listParams,
    fillCountriesOnce,
  }
}
