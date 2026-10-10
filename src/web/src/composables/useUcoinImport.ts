// Imports a uCoin.net collection export: creates the coins from the file, then
// completes them from the Numista catalogue (metal, weight, mintage-based
// rarity, current value for the coin's grade). Each Numista type is looked up
// once for all its coins and every response is cached server-side, so the
// import stays inside Numista's free monthly quota; when the quota runs out
// the run stops and can be resumed later from the same page.
import { computed, reactive, ref } from 'vue'
import {
  addOfficialCoinPhotos,
  createCoin,
  getCoins,
  getCollectionSummary,
  getNumistaCatalogType,
  getNumistaTypeMarketValue,
  listCatalogs,
  searchNumistaCatalog,
  searchNumistaIssuers,
  translateText,
  updateCoin,
  type NumistaCatalogSearchItem,
  type NumistaCatalogSearchParams,
  type NumistaCatalogType,
  type NumistaDataStatus,
  type NumistaGradePrice,
} from '@/api/client'
import type { CatalogRegistry, Coin, CoinMutationPayload, CoinReferenceInput } from '@/types'
import { useCoinOptions } from '@/composables/useCoinOptions'
import { collectionCurrency, toCollectionCurrency } from '@/utils/collectionCurrency'
import { numistaLang } from '@/utils/numistaCatalog'
import { buildCatalogPrefill } from '@/utils/numistaCatalogPrefill'
import { numistaStatusMessage } from '@/utils/numistaMarketValue'
import {
  buildUcoinCoin,
  isUcoinImported,
  parseUcoinCsv,
  ucoinCoinKey,
  ucoinCoinName,
  type UcoinParseError,
  type UcoinRow,
} from '@/utils/ucoinImport'
import {
  issuerSimilarity,
  pickIssue,
  pickNumistaType,
  priceForGrade,
  ucoinLookupForCoin,
  ucoinLookupForRow,
  ucoinLookupGroupKey,
  unitQuery,
  valuationGrade,
  type UcoinLookup,
} from '@/utils/ucoinNumista'

export type UcoinImportPhase = 'idle' | 'ready' | 'creating' | 'enriching' | 'photos' | 'finished'

export interface UcoinImportCoinRef {
  id: number
  name: string
}

export interface UcoinImportOptions {
  /** Pause between API calls, to stay under the app's per-user rate limit. */
  paceMs?: number
  /** Groups looked up at the same time. */
  workers?: number
}

const NOT_FOUND_KEY = 'ucoinImport:notFound'
const PAGE_SIZE = 100
// Official photos are downloaded by the server; this keeps under its limit.
const PHOTO_PACE_MS = 250
const RATE_LIMITED_WAIT_MS = 5000
const SEARCH_PAGE_SIZE = 50
const MAX_ISSUER_PAGES = 3
// Statuses that will not get better by trying the next coin.
const STOPPING_STATUSES: NumistaDataStatus[] = ['quota_limited', 'unauthorized', 'unconfigured']

class NumistaStop extends Error {
  constructor(readonly status: NumistaDataStatus) {
    super(status)
  }
}

function fold(value: string): string {
  return value.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '').trim()
}

function readNotFound(): Set<number> {
  try {
    const saved = JSON.parse(localStorage.getItem(NOT_FOUND_KEY) ?? '[]')
    return new Set(Array.isArray(saved) ? saved.filter((id) => Number.isInteger(id)) : [])
  } catch {
    return new Set()
  }
}

function saveNotFound(ids: Set<number>) {
  try {
    localStorage.setItem(NOT_FOUND_KEY, JSON.stringify([...ids]))
  } catch {
    // Without storage they are simply looked up again next time.
  }
}

function hasNumistaReference(coin: Coin): boolean {
  return (coin.references ?? []).some((ref) => ref.catalog.toLowerCase() === 'numista')
}

function referenceKey(ref: CoinReferenceInput): string {
  return `${ref.catalog.toLowerCase()}|${(ref.volume ?? '').toLowerCase()}|${ref.number.toLowerCase()}`
}

function mergeReferences(coin: Coin, added: CoinReferenceInput[]): CoinReferenceInput[] {
  const merged = new Map<string, CoinReferenceInput>()
  for (const ref of coin.references ?? []) {
    const input: CoinReferenceInput = { catalog: ref.catalog, volume: ref.volume || undefined, number: ref.number, uri: ref.uri || undefined }
    merged.set(referenceKey(input), input)
  }
  for (const ref of added) if (!merged.has(referenceKey(ref))) merged.set(referenceKey(ref), ref)
  return [...merged.values()]
}

function clip(value: string, max: number): string {
  return value.length > max ? `${value.slice(0, max - 1)}…` : value
}

/** Fields Numista adds to an imported coin; what came from the file is kept. */
export function ucoinEnrichment(
  coin: Coin,
  type: NumistaCatalogType,
  prefill: ReturnType<typeof buildCatalogPrefill>,
): CoinMutationPayload {
  const fields = prefill.fields
  return {
    // updateCoin sends every nullable field it is not given as null, which
    // would clear it: keep the coin's own values for those.
    purchasePrice: coin.purchasePrice,
    purchaseDate: coin.purchaseDate,
    storageLocationId: coin.storageLocationId,
    storageSlot: coin.storageSlot ?? null,
    romanImperialFigureId: coin.romanImperialFigureId,
    mintLocationId: coin.mintLocationId,
    year: coin.year ?? null,
    category: fields.category ?? coin.category,
    era: fields.era || coin.era,
    ruler: coin.ruler || fields.ruler,
    // A mint mark from the file is kept; Numista's mint only when there is just one.
    mint: coin.mint || (type.mints.length === 1 ? fields.mint : '') || '',
    material: fields.material && fields.material !== 'Other' ? fields.material : coin.material,
    weightGrams: coin.weightGrams ?? fields.weightGrams ?? null,
    diameterMm: coin.diameterMm ?? fields.diameterMm ?? null,
    rarity: coin.rarity || fields.rarity || '',
    obverseInscription: coin.obverseInscription || fields.obverseInscription,
    reverseInscription: coin.reverseInscription || fields.reverseInscription,
    obverseDescription: coin.obverseDescription || fields.obverseDescription,
    reverseDescription: coin.reverseDescription || fields.reverseDescription,
    currentValue: fields.currentValue ?? coin.currentValue,
    notes: clip([coin.notes, fields.notes].filter(Boolean).join('\n'), 5000),
    references: mergeReferences(coin, prefill.references),
  }
}

export function useUcoinImport(options: UcoinImportOptions = {}) {
  const paceMs = options.paceMs ?? 120
  const workerCount = options.workers ?? 2

  const phase = ref<UcoinImportPhase>('idle')
  const fileName = ref('')
  const parseError = ref<UcoinParseError | ''>('')
  const loadFailed = ref(false)
  const rows = ref<UcoinRow[]>([])
  /** Rows of the file not yet in the collection. */
  const pendingRows = ref<UcoinRow[]>([])
  /** Imported coins still without Numista data (from this or an earlier import). */
  const awaitingNumista = ref<Coin[]>([])
  /** Imported coins Numista did not have in an earlier run, left out until asked again. */
  const earlierNotFound = ref(0)
  /** Imported coins with no image yet. */
  const withoutImages = ref<Coin[]>([])
  const progress = reactive({ toCreate: 0, created: 0, createFailed: 0, toEnrich: 0, enriched: 0, notFound: 0, failed: 0, toPhoto: 0, photoDone: 0, photographed: 0 })
  const notFoundCoins = ref<UcoinImportCoinRef[]>([])
  const failedLines = ref<number[]>([])
  /** Why Numista stopped the run (quota, key), or ''. */
  const stopMessage = ref('')
  const stopped = ref(false)
  let stopRequested = false
  let nextSlot = 0
  /** ECB rate (US dollars per euro) for amounts in the other currency. */
  let usdPerEur: number | undefined

  const countries = computed(() => new Set(rows.value.map((row) => fold(row.country)).filter(Boolean)).size)
  const alreadyImported = computed(() => rows.value.length - pendingRows.value.length)

  /** Rough number of Numista requests: a search, type and issues per type, a price per issue. */
  const estimatedRequests = computed(() => {
    const groups = new Set<string>()
    const issues = new Set<string>()
    const lookups = [...pendingRows.value.map(ucoinLookupForRow), ...awaitingNumista.value.map(ucoinLookupForCoin)]
    for (const lookup of lookups) {
      const key = ucoinLookupGroupKey(lookup)
      groups.add(key)
      issues.add(`${key}|${lookup.year ?? ''}|${lookup.mintMark}`)
    }
    return groups.size * 3 + issues.size
  })

  async function pace() {
    if (paceMs <= 0) return
    const now = Date.now()
    const wait = Math.max(0, nextSlot - now)
    nextSlot = Math.max(now, nextSlot) + paceMs
    if (wait) await new Promise((resolve) => setTimeout(resolve, wait))
  }

  async function loadCollection(): Promise<Coin[]> {
    const all: Coin[] = []
    for (let page = 1; page < 100; page += 1) {
      await pace()
      const res = await getCoins({ page, limit: PAGE_SIZE })
      const coins = res.data.coins ?? []
      all.push(...coins)
      if (!coins.length || all.length >= res.data.total) break
    }
    return all
  }

  function refreshAwaiting(collection: Coin[]) {
    const skipped = readNotFound()
    const withoutNumista = collection.filter((coin) => isUcoinImported(coin) && !hasNumistaReference(coin) && !coin.isWishlist)
    awaitingNumista.value = withoutNumista.filter((coin) => !skipped.has(coin.id))
    earlierNotFound.value = withoutNumista.length - awaitingNumista.value.length
    withoutImages.value = collection.filter((coin) => isUcoinImported(coin) && !coin.isWishlist && !coin.images?.length)
  }

  /** Loads the collection to know which coins were already imported. */
  async function init() {
    try {
      refreshAwaiting(await loadCollection())
      loadFailed.value = false
    } catch {
      loadFailed.value = true
    }
  }

  async function readFile(file: File) {
    fileName.value = file.name
    parseError.value = ''
    const parsed = parseUcoinCsv(await file.text())
    if (parsed.error || !parsed.rows.length) {
      parseError.value = parsed.error ?? 'empty'
      rows.value = []
      pendingRows.value = []
      phase.value = 'idle'
      return
    }
    let collection: Coin[]
    try {
      collection = await loadCollection()
      loadFailed.value = false
    } catch {
      loadFailed.value = true
      return
    }
    refreshAwaiting(collection)
    // The same coin can be in the file twice; skip only as many as are already in the collection.
    const existing = new Map<string, number>()
    for (const coin of collection.filter(isUcoinImported)) {
      const key = ucoinCoinKey(coin)
      existing.set(key, (existing.get(key) ?? 0) + 1)
    }
    rows.value = parsed.rows
    pendingRows.value = parsed.rows.filter((row) => {
      const key = ucoinCoinKey({ name: ucoinCoinName(row), country: row.country })
      const count = existing.get(key) ?? 0
      if (count > 0) existing.set(key, count - 1)
      return count === 0
    })
    phase.value = 'ready'
  }

  async function runPool<T>(items: T[], workers: number, task: (item: T) => Promise<void>) {
    let index = 0
    const worker = async () => {
      while (!stopRequested && index < items.length) {
        const item = items[index] as T
        index += 1
        await task(item)
      }
    }
    await Promise.all(Array.from({ length: Math.min(workers, items.length) }, worker))
  }

  // ---- Numista lookups ---------------------------------------------------

  const lang = numistaLang()
  const learnedIssuers = new Map<string, { code: string; minYear?: number; maxYear?: number }[]>()
  const issuerLookups = new Map<string, Promise<string | undefined>>()
  const typeResults = new Map<number, Promise<{ type?: NumistaCatalogType; issues: Awaited<ReturnType<typeof getNumistaCatalogType>>['data']['issues'] }>>()
  const translations = new Map<string, Promise<string | undefined>>()
  const priceResults = new Map<string, Promise<NumistaGradePrice[]>>()

  function checkStatus(status: NumistaDataStatus) {
    if (STOPPING_STATUSES.includes(status)) throw new NumistaStop(status)
  }

  async function search(params: Omit<NumistaCatalogSearchParams, 'lang'>) {
    await pace()
    const res = await searchNumistaCatalog({ ...params, lang })
    checkStatus(res.data.status)
    return res.data
  }

  function learnIssuer(country: string, item: NumistaCatalogSearchItem) {
    const key = fold(country)
    const known = learnedIssuers.get(key) ?? []
    if (!known.some((entry) => entry.code === item.issuer.code && entry.minYear === item.minYear && entry.maxYear === item.maxYear)) {
      known.push({ code: item.issuer.code, minYear: item.minYear, maxYear: item.maxYear })
    }
    learnedIssuers.set(key, known)
  }

  /** Numista issuer code for the country: one already matched near that year, else by name. */
  async function issuerCode(lookup: UcoinLookup): Promise<string | undefined> {
    const year = lookup.year ?? 0
    const learned = learnedIssuers.get(fold(lookup.country)) ?? []
    const distance = (entry: { minYear?: number; maxYear?: number }) =>
      Math.max(0, (entry.minYear ?? year) - year, year - (entry.maxYear ?? year))
    const nearest = [...learned].sort((a, b) => distance(a) - distance(b))[0]
    if (nearest && distance(nearest) <= 30) return nearest.code

    const key = fold(lookup.country)
    if (!issuerLookups.has(key)) {
      issuerLookups.set(key, (async () => {
        // "Região Autónoma dos Açores" is just "Açores" on Numista: try the last word too.
        const lastWord = lookup.country.trim().split(/\s+/).pop() ?? ''
        for (const query of [lookup.country, lastWord]) {
          if (!query || (query !== lookup.country && query.length < 4)) continue
          await pace()
          const res = await searchNumistaIssuers(query, lang)
          checkStatus(res.data.status)
          // Ties go to a country rather than a part of something larger
          // (modern Syria, not the Roman province of the same name).
          const ranked = res.data.issuers
            .map((issuer) => ({ issuer, score: issuerSimilarity(lookup.country, issuer.name) - (issuer.parent ? 0.1 : 0) }))
            .filter((entry) => entry.score > 0)
            .sort((a, b) => b.score - a.score)
          if (ranked[0]) return ranked[0].issuer.code
          if (query === lastWord) break
        }
        return undefined
      })())
    }
    return issuerLookups.get(key)
  }

  /**
   * The subject in English, for Numista's English titles: uCoin writes it in
   * the language of its site ("Jornada Mundial da Juventude" for "World Youth
   * Day"). Undefined when the app is in English or the translation fails.
   */
  function translatedSubject(subject: string): Promise<string | undefined> {
    if (lang === 'en' || !subject.trim()) return Promise.resolve(undefined)
    if (!translations.has(subject)) {
      translations.set(subject, translateText(subject.slice(0, 400), 'pt')
        .then((res) => (res.data.status === 'ok' ? res.data.text : undefined))
        .catch(() => undefined))
    }
    return translations.get(subject)!
  }

  /** Picks the type, then, when none fits, tries again with the subject in English too. */
  async function pickType(candidates: NumistaCatalogSearchItem[], lookup: UcoinLookup, source: 'catalogue' | 'issuer') {
    const match = pickNumistaType(candidates, lookup, source)
    if (match || !candidates.length || !lookup.subject.trim()) return match
    const english = await translatedSubject(lookup.subject)
    return english ? pickNumistaType(candidates, { ...lookup, subject: `${lookup.subject} ${english}` }, source) : undefined
  }

  async function resolveType(lookup: UcoinLookup): Promise<number | undefined> {
    if (lookup.catalogue) {
      const reference = { catalogue: lookup.catalogue.code, number: lookup.catalogue.number, count: SEARCH_PAGE_SIZE }
      let match = await pickType((await search({ ...reference, year: lookup.year })).types, lookup, 'catalogue')
      // Numista filters by the year written on the coin, so coins dated in
      // another calendar (1387 AH, Heisei 3) are found by issuer instead.
      if (!match) {
        const issuer = await issuerCode(lookup)
        if (issuer) match = await pickType((await search({ ...reference, issuer })).types, lookup, 'catalogue')
      }
      if (match) {
        learnIssuer(lookup.country, match)
        return match.id
      }
    }
    // Listing an issuer's types needs a year to stay short.
    if (lookup.year === undefined) return undefined
    const issuer = await issuerCode(lookup)
    if (!issuer) return undefined
    const q = unitQuery(lookup.denomination)
    const candidates: NumistaCatalogSearchItem[] = []
    for (let page = 1; page <= MAX_ISSUER_PAGES; page += 1) {
      const result = await search({ issuer, year: lookup.year, q, page, count: SEARCH_PAGE_SIZE })
      candidates.push(...result.types)
      if (result.count <= page * SEARCH_PAGE_SIZE) break
    }
    const match = await pickType(candidates, lookup, 'issuer')
    if (match) learnIssuer(lookup.country, match)
    return match?.id
  }

  function loadType(typeId: number) {
    if (!typeResults.has(typeId)) {
      typeResults.set(typeId, (async () => {
        await pace()
        const res = await getNumistaCatalogType(typeId, lang)
        checkStatus(res.data.status)
        return { type: res.data.type, issues: res.data.issues }
      })())
    }
    return typeResults.get(typeId)!
  }

  function loadPrices(typeId: number, issueId: number) {
    const key = `${typeId}|${issueId}`
    if (!priceResults.has(key)) {
      priceResults.set(key, (async () => {
        await pace()
        // Prices are asked in US dollars, the currency earlier lookups are
        // cached in, and converted to the collection currency with the ECB
        // rate; without a rate they are asked in the collection currency.
        const currency = toCollectionCurrency(1, 'USD', usdPerEur) === undefined ? collectionCurrency.value : 'USD'
        const res = await getNumistaTypeMarketValue(typeId, { currency, issueId })
        checkStatus(res.data.status)
        if (res.data.status !== 'ok') return []
        return res.data.prices.flatMap((entry) => {
          const price = toCollectionCurrency(entry.price, currency, usdPerEur)
          return price === undefined ? [] : [{ ...entry, price }]
        })
      })())
    }
    return priceResults.get(key)!
  }

  async function loadRate() {
    try {
      usdPerEur = (await getCollectionSummary()).data.exchangeRate?.usdPerEur
    } catch {
      // Prices in another currency then stay in the notes.
    }
  }

  // ---- Running the import ---------------------------------------------------

  function stopWith(error: unknown): boolean {
    if (error instanceof NumistaStop) {
      stopRequested = true
      stopMessage.value = numistaStatusMessage(error.status)
      return true
    }
    return false
  }

  async function enrichGroup(
    group: { lookup: UcoinLookup; coins: Coin[] },
    catalogOptions: { categories: string[]; eras: string[]; materials: string[]; registry: CatalogRegistry[] },
    notFound: Set<number>,
  ) {
    let typeId: number | undefined
    let type: NumistaCatalogType | undefined
    let issues: Awaited<ReturnType<typeof loadType>>['issues'] = []
    try {
      typeId = await resolveType(group.lookup)
      if (typeId) ({ type, issues } = await loadType(typeId))
    } catch (error) {
      if (!stopWith(error)) progress.failed += group.coins.length
      return
    }
    if (!typeId || !type) {
      for (const coin of group.coins) {
        notFound.add(coin.id)
        notFoundCoins.value.push({ id: coin.id, name: coin.name })
      }
      progress.notFound += group.coins.length
      saveNotFound(notFound)
      return
    }

    for (const coin of group.coins) {
      if (stopRequested) return
      try {
        const lookup = ucoinLookupForCoin(coin)
        const issue = pickIssue(issues, lookup) ?? (issues.length === 1 ? issues[0] : undefined)
        const prices = issue ? await loadPrices(typeId, issue.id) : []
        const currentValue = priceForGrade(prices, valuationGrade(coin.grade))
        const prefill = buildCatalogPrefill({ type, issue, currentValue, wishlist: false }, catalogOptions)
        await pace()
        await updateCoin(coin.id, ucoinEnrichment(coin, type, prefill))
        progress.enriched += 1
      } catch (error) {
        if (stopWith(error)) return
        progress.failed += 1
      }
    }
  }

  async function catalogOptions() {
    const { categoryOptions, eraOptions, materialOptions, loadOptions } = useCoinOptions()
    await loadOptions()
    let registry: CatalogRegistry[] = []
    try {
      registry = await listCatalogs()
    } catch {
      // Without the registry only the Numista reference is saved; others go to notes.
    }
    return { categories: categoryOptions.value, eras: eraOptions.value, materials: materialOptions.value, registry }
  }

  // ---- Official photos -------------------------------------------------------

  let nextPhotoSlot = 0
  async function pacePhoto() {
    const now = Date.now()
    const wait = Math.max(0, nextPhotoSlot - now)
    nextPhotoSlot = Math.max(now, nextPhotoSlot) + (paceMs > 0 ? PHOTO_PACE_MS : 0)
    if (wait) await new Promise((resolve) => setTimeout(resolve, wait))
  }

  /** Asks the server for the coin's official photos (ECB, Portuguese and French mints). */
  async function addPhotos(coin: Coin) {
    for (let attempt = 0; attempt < 2; attempt += 1) {
      await pacePhoto()
      try {
        const res = await addOfficialCoinPhotos(coin.id, ['obverse', 'reverse'])
        if (res.data.images.length) progress.photographed += 1
        return
      } catch (error) {
        const status = (error as { response?: { status?: number } }).response?.status
        if (status !== 429 || paceMs <= 0) return
        await new Promise((resolve) => setTimeout(resolve, RATE_LIMITED_WAIT_MS))
      }
    }
  }

  /** Gives the imported coins without an image their official photos, when there are any. */
  async function photograph(coins: Coin[]) {
    if (!coins.length || stopRequested) return
    phase.value = 'photos'
    progress.toPhoto = coins.length
    progress.photoDone = 0
    progress.photographed = 0
    await runPool(coins, workerCount, async (coin) => {
      await addPhotos(coin)
      progress.photoDone += 1
    })
  }

  /** Adds official photos to every imported coin that has none. */
  async function addMissingPhotos() {
    stopRequested = false
    stopped.value = false
    stopMessage.value = ''
    progress.toCreate = 0
    progress.toEnrich = 0
    progress.notFound = 0
    progress.failed = 0
    notFoundCoins.value = []
    try {
      refreshAwaiting(await loadCollection())
    } catch {
      loadFailed.value = true
      return
    }
    await photograph(withoutImages.value)
    stopped.value = stopRequested
    try {
      refreshAwaiting(await loadCollection())
    } catch {
      // The counts above still describe the run.
    }
    phase.value = 'finished'
  }

  /** Completes every imported coin that has no Numista data yet, then adds photos. */
  async function enrich() {
    phase.value = 'enriching'
    stopRequested = false
    stopped.value = false
    stopMessage.value = ''
    progress.enriched = 0
    progress.notFound = 0
    progress.failed = 0
    notFoundCoins.value = []
    const options = await catalogOptions()
    await loadRate()
    try {
      refreshAwaiting(await loadCollection())
    } catch {
      loadFailed.value = true
      phase.value = 'finished'
      return
    }
    const groups = new Map<string, { lookup: UcoinLookup; coins: Coin[] }>()
    for (const coin of awaitingNumista.value) {
      const lookup = ucoinLookupForCoin(coin)
      const key = ucoinLookupGroupKey(lookup)
      const group = groups.get(key) ?? { lookup, coins: [] }
      group.coins.push(coin)
      groups.set(key, group)
    }
    progress.toEnrich = awaitingNumista.value.length
    // Catalogue-number searches first: they also teach which issuer each country is.
    const ordered = [...groups.values()].sort((a, b) => Number(!a.lookup.catalogue) - Number(!b.lookup.catalogue))
    const notFound = readNotFound()
    progress.toPhoto = 0
    await runPool(ordered, workerCount, (group) => enrichGroup(group, options, notFound))
    // Photos come after the Numista data, which helps find them.
    try {
      refreshAwaiting(await loadCollection())
      await photograph(withoutImages.value)
      refreshAwaiting(await loadCollection())
    } catch {
      // The counts above still describe the run.
    }
    stopped.value = stopRequested
    phase.value = 'finished'
  }

  /** Creates the file's new coins, then (optionally) completes them from Numista. */
  async function start(withNumista: boolean) {
    if (phase.value !== 'ready') return
    phase.value = 'creating'
    stopRequested = false
    stopped.value = false
    progress.toCreate = pendingRows.value.length
    progress.created = 0
    progress.createFailed = 0
    progress.toPhoto = 0
    failedLines.value = []
    const options = await catalogOptions()
    await loadRate()
    await runPool(pendingRows.value, 3, async (row) => {
      try {
        await pace()
        await createCoin(buildUcoinCoin(row, { ...options, usdPerEur }))
        progress.created += 1
      } catch {
        progress.createFailed += 1
        failedLines.value.push(row.line)
      }
    })
    pendingRows.value = []
    if (stopRequested) {
      stopped.value = true
      phase.value = 'finished'
      return
    }
    if (withNumista) {
      await enrich()
      return
    }
    try {
      refreshAwaiting(await loadCollection())
      await photograph(withoutImages.value)
    } catch {
      // The coins are added; photos can be added later.
    }
    stopped.value = stopRequested
    phase.value = 'finished'
  }

  function stop() {
    stopRequested = true
  }

  /** Forgets which coins Numista did not have, so the next run tries them again. */
  function retryNotFound() {
    saveNotFound(new Set())
  }

  return {
    phase,
    fileName,
    parseError,
    loadFailed,
    rows,
    pendingRows,
    awaitingNumista,
    earlierNotFound,
    withoutImages,
    addMissingPhotos,
    countries,
    alreadyImported,
    estimatedRequests,
    progress,
    notFoundCoins,
    failedLines,
    stopMessage,
    stopped,
    init,
    readFile,
    start,
    enrich,
    stop,
    retryNotFound,
  }
}
