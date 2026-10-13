// coins endpoints. Split out of the former monolithic client.ts;
// re-exported from '@/api/client' so existing imports keep working.
import { api } from '@/api/http'
import type {
  AIJob,
  AIJobStartResponse,
  CatalogRegistry,
  Coin,
  CoinHealthItem,
  CoinHealthListResponse,
  CoinImage,
  CoinJournal,
  CoinKind,
  CoinListResponse,
  CoinLookupImageRole,
  CoinLookupResponse,
  CoinMutationPayload,
  CoinRecommendation,
  CoinReference,
  CoinReferenceInput,
  CoinValueHistory,
  IntakeCommitRequest,
  IntakeCommitResponse,
  IntakeDraft,
  LegacyMigrationResult,
  NumistaEnrichmentRequest,
  NumistaLookupOutcome,
  NumistaLookupRequest,
  NumistaQueryProposal,
  NumistaQueryProposalRequest,
  NumistaSearchResponse,
  QuickCaptureDraft,
  QuickCaptureDraftInput,
  QuickCaptureDraftListResponse,
  QuickCaptureDraftStatus,
  QuickCaptureDraftUpdateInput,
  QuickCapturePromoteRequest,
  QuickCapturePromotionResponse,
  ShipmentEnvelopeResponse,
  ShipmentStatus,
  ShipmentUpsertInput,
  ValueSnapshot,
} from '@/types'
import { appendOptionalFormValue } from '@/api/endpoints/_shared'

// Coins
export const getCoins = (params?: {
  /** Only coins or only banknotes; every item when omitted. */
  kind?: CoinKind
  category?: string
  era?: string
  country?: string
  /** "true": only coins of countries that no longer exist. */
  extinct?: string
  /** Only coins still legal tender, or no longer. */
  circulation?: string
  /** "true": only commemorative coins; "false": only the others. */
  commemorative?: string
  /** Only coins of one Numista kind of coin (a CollectionClassification key). */
  classification?: string
  presentation?: string
  rarity?: string
  search?: string
  wishlist?: string
  sold?: string
  tag?: string
  set?: string
  page?: number
  limit?: number
  sort?: string
  order?: string
  seed?: number
}) => api.get<CoinListResponse>('/coins', { params })

const NULLABLE_FIELDS: (keyof Coin)[] = ['weightGrams', 'diameterMm', 'widthMm', 'heightMm', 'purchasePrice', 'currentValue', 'purchaseDate', 'storageLocationId', 'storageSlot', 'romanImperialFigureId', 'mintLocationId', 'year']

function sanitizeCoin(coin: CoinMutationPayload): CoinMutationPayload {
  const clean: Record<string, unknown> = { ...coin }
  for (const field of NULLABLE_FIELDS) {
    if (clean[field] === '' || clean[field] === undefined) {
      clean[field] = null
    }
  }
  delete clean.storageLocation
  delete clean.mintLocation
  // Default currentValue to purchasePrice if not set (preserve 0 as valid)
  if (clean.currentValue == null && clean.purchasePrice != null) {
    clean.currentValue = clean.purchasePrice
  }
  // Convert date-only strings (YYYY-MM-DD) to RFC3339 for Go
  if (typeof clean.purchaseDate === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(clean.purchaseDate)) {
    clean.purchaseDate = clean.purchaseDate + 'T00:00:00Z'
  }
  return clean as CoinMutationPayload
}

export const getCoin = (id: number) => api.get<Coin>(`/coins/${id}`)

/** Whether a coin is still legal tender and whether it is commemorative; null when nobody knows. */
export interface CoinStatus {
  /** Still legal tender (a collector coin can be, while not "current" in the KPI). */
  current: boolean | null
  /** Issued for collectors rather than for circulation (Numista's "non-circulating coins"). */
  nonCirculating?: boolean | null
  /** When it stopped being legal tender, if Numista says (usually YYYY-MM-DD). */
  demonetizationDate?: string
  commemorative: boolean | null
  commemoratedTopic?: string
}

export const getCoinStatus = (id: number) => api.get<CoinStatus>(`/coins/${id}/status`)

/** Collection coin IDs by legal tender and commemorative status; a coin nobody knows about is in neither list of a pair. */
export interface CoinStatusGroups {
  /** Circulation coins still legal tender. */
  current: number[]
  demonetized: number[]
  /** Collector (non-circulating) coins, not demonetized. */
  nonCirculating?: number[]
  commemorative: number[]
  regular: number[]
}

export const getCoinStatusGroups = () => api.get<CoinStatusGroups>('/coins/statuses')

export const createCoin = (coin: CoinMutationPayload) => api.post<Coin>('/coins', sanitizeCoin(coin))

export interface MatchCategoryEraResponse {
  match: string
  matched: boolean
}

export const matchCategoryEra = (type: 'category' | 'era', value: string) =>
  api.post<MatchCategoryEraResponse>('/coins/match-category-era', { type, value })

export interface CoinOptionLabels {
  categories: Record<string, string>
  eras: Record<string, string>
}

export interface CoinOptionsResponse {
  categories: string[]
  eras: string[]
  /** Admin-defined display names keyed by locale (e.g. `pt-PT`), then by stored value. */
  labels: Record<string, CoinOptionLabels>
  /** Collection currency coin prices and values are kept in. */
  currency?: string
}

/** Configured category and era choices; readable by every signed-in user. */
export const getCoinOptions = () => api.get<CoinOptionsResponse>('/coin-options')

export type NumistaDataStatus =
  | 'ok'
  | 'not_linked'
  | 'no_issues'
  | 'no_prices'
  | 'unconfigured'
  | 'unauthorized'
  | 'quota_limited'
  | 'unavailable'

/** One issue (year / mint mark variant) of a Numista type. Year is negative for BC. */
export interface NumistaIssueSummary {
  id: number
  year?: number
  minYear?: number
  maxYear?: number
  mintLetter?: string
  mintage?: number
  comment?: string
}

/** Numista's estimated price for one grade: g, vg, f, vf, xf, au or unc. */
export interface NumistaGradePrice {
  grade: string
  price: number
}

/** A reference photo from the Numista catalogue; always show its credit. */
export interface NumistaCatalogPhoto {
  side: 'obverse' | 'reverse'
  thumbnail: string
  picture?: string
  copyright?: string
  license?: string
}

export interface CoinMarketValue {
  status: NumistaDataStatus
  numistaTypeId?: number
  numistaUrl?: string
  issues: NumistaIssueSummary[]
  selectedIssueId?: number
  currency: string
  prices: NumistaGradePrice[]
  /** When Numista last returned these prices. */
  fetchedAt?: string
  /** True when a refresh failed and stored values are shown. */
  stale: boolean
  catalogPhotos: NumistaCatalogPhoto[]
}

/** Numista's estimated price per grade for the coin's linked Numista type (cached server-side for days). */
export const getCoinMarketValue = (coinId: number, params: { currency: string; issueId?: number; lang?: string }) =>
  api.get<CoinMarketValue>(`/coins/${coinId}/market-value`, { params })

export async function createIntakeDraft(images: File[], coinCardImage?: File) {
  const formData = new FormData()
  for (const image of images) {
    formData.append('images', image)
  }
  if (coinCardImage) {
    formData.append('coinCardImage', coinCardImage)
  }
  return api.post<IntakeDraft>('/coins/intake/draft', formData)
}


export async function createQuickCaptureDraft(input: QuickCaptureDraftInput) {
  const formData = new FormData()
  appendOptionalFormValue(formData, 'workingTitle', input.workingTitle)
  appendOptionalFormValue(formData, 'dateRange', input.dateRange)
  appendOptionalFormValue(formData, 'era', input.era)
  appendOptionalFormValue(formData, 'acquisitionSource', input.acquisitionSource)
  appendOptionalFormValue(formData, 'notes', input.notes)
  appendOptionalFormValue(formData, 'source', input.source)
  appendOptionalFormValue(formData, 'ngcCertNumber', input.ngcCertNumber)
  appendOptionalFormValue(formData, 'ngcLookupUrl', input.ngcLookupUrl)
  appendOptionalFormValue(formData, 'ngcGrade', input.ngcGrade)
  appendOptionalFormValue(formData, 'labelText', input.labelText)
  appendOptionalFormValue(formData, 'aiConfidence', input.aiConfidence)
  appendOptionalFormValue(formData, 'selectedNumistaId', input.selectedNumistaId)
  appendOptionalFormValue(formData, 'selectedNumistaUrl', input.selectedNumistaUrl)
  if (input.purchasePrice !== undefined && input.purchasePrice !== null) {
    formData.append('purchasePrice', String(input.purchasePrice))
  }
  if (input.obverseImage) formData.append('obverseImage', input.obverseImage)
  if (input.reverseImage) formData.append('reverseImage', input.reverseImage)
  for (const image of input.detailImages ?? []) {
    formData.append('detailImages', image)
  }
  return api.post<QuickCaptureDraft>('/quick-capture/drafts', formData)
}

export const listQuickCaptureDrafts = (params?: { status?: QuickCaptureDraftStatus; page?: number; limit?: number }) =>
  api.get<QuickCaptureDraftListResponse>('/quick-capture/drafts', { params })

export const getQuickCaptureDraft = (id: number) => api.get<QuickCaptureDraft>(`/quick-capture/drafts/${id}`)

export async function updateQuickCaptureDraft(id: number, input: QuickCaptureDraftUpdateInput) {
  const formData = new FormData()
  formData.append('workingTitle', input.workingTitle)
  formData.append('dateRange', input.dateRange)
  formData.append('era', input.era)
  formData.append('acquisitionSource', input.acquisitionSource)
  formData.append('notes', input.notes)
  appendOptionalFormValue(formData, 'source', input.source)
  appendOptionalFormValue(formData, 'ngcCertNumber', input.ngcCertNumber)
  appendOptionalFormValue(formData, 'ngcLookupUrl', input.ngcLookupUrl)
  appendOptionalFormValue(formData, 'ngcGrade', input.ngcGrade)
  appendOptionalFormValue(formData, 'labelText', input.labelText)
  appendOptionalFormValue(formData, 'aiConfidence', input.aiConfidence)
  appendOptionalFormValue(formData, 'selectedNumistaId', input.selectedNumistaId)
  appendOptionalFormValue(formData, 'selectedNumistaUrl', input.selectedNumistaUrl)
  if (input.clearSelectedNumista) formData.append('clearSelectedNumista', 'true')
  if (input.purchasePrice !== null && input.purchasePrice !== undefined) {
    formData.append('purchasePrice', String(input.purchasePrice))
  }
  if (input.removeImageIds) {
    formData.append('removeImageIds', input.removeImageIds)
  }
  if (input.replaceObverse) formData.append('replaceObverse', 'true')
  if (input.replaceReverse) formData.append('replaceReverse', 'true')
  if (input.obverseImage) formData.append('obverseImage', input.obverseImage)
  if (input.reverseImage) formData.append('reverseImage', input.reverseImage)
  for (const file of input.detailImages ?? []) {
    formData.append('detailImages', file)
  }
  return api.put<QuickCaptureDraft>(`/quick-capture/drafts/${id}`, formData)
}

export const discardQuickCaptureDraft = (id: number) =>
  api.post<QuickCaptureDraft>(`/quick-capture/drafts/${id}/discard`)

export const promoteQuickCaptureDraft = (id: number, request: QuickCapturePromoteRequest) =>
  api.post<QuickCapturePromotionResponse>(`/quick-capture/drafts/${id}/promote`, request)

export const commitIntakeDraft = (request: IntakeCommitRequest) =>
  api.post<IntakeCommitResponse>('/coins/intake/commit', {
    ...request,
    overrides: request.overrides ? sanitizeCoin(request.overrides) : undefined,
  })

export const updateCoin = (id: number, coin: CoinMutationPayload, params?: Record<string, string>) =>
  api.put<Coin>(`/coins/${id}`, sanitizeCoin(coin), { params })

export const purchaseCoin = (id: number, data?: { purchasePrice?: number; purchaseDate?: string; purchaseLocation?: string; shipment?: ShipmentUpsertInput }) =>
  api.post<Coin>(`/coins/${id}/purchase`, data || {})

export const sellCoin = (id: number, soldPrice: number | null, soldTo: string) =>
  api.post<Coin>(`/coins/${id}/sell`, { soldPrice, soldTo })

export const duplicateCoin = (id: number) => api.post<Coin>(`/coins/${id}/duplicate`)

export const deleteCoin = (id: number) => api.delete(`/coins/${id}`)

export const getCoinReferences = (coinId: number) => api.get<CoinReference[]>(`/coins/${coinId}/references`)

export const createCoinReference = (coinId: number, reference: CoinReferenceInput) =>
  api.post<CoinReference>(`/coins/${coinId}/references`, reference)

export const updateCoinReference = (coinId: number, referenceId: number, reference: CoinReferenceInput) =>
  api.put<CoinReference>(`/coins/${coinId}/references/${referenceId}`, reference)

export const deleteCoinReference = (coinId: number, referenceId: number) =>
  api.delete(`/coins/${coinId}/references/${referenceId}`)

export const migrateLegacyReferences = () =>
  api.post<LegacyMigrationResult>('/references/migrate-legacy')

export const addTagToCoin = (coinId: number, tagId: number) => api.post(`/coins/${coinId}/tags`, { tagId })

export const removeTagFromCoin = (coinId: number, tagId: number) => api.delete(`/coins/${coinId}/tags/${tagId}`)

export const getCoinRecommendations = (coinId: number) =>
  api.get<{ recommendations: CoinRecommendation[] }>(`/coins/${coinId}/recommendations`)

export const acceptCoinRecommendation = (coinId: number, recommendationId: number) =>
  api.post(`/coins/${coinId}/recommendations/${recommendationId}/accept`)

export const rejectCoinRecommendation = (coinId: number, recommendationId: number) =>
  api.post(`/coins/${coinId}/recommendations/${recommendationId}/reject`)

// Catalog Registry
export const listCatalogs = async () => {
  const res = await api.get<{ catalogs: CatalogRegistry[] }>('/catalogs')
  return res.data.catalogs ?? []
}

// Bulk Operations
export const bulkAction = (
  coinIds: number[],
  action: string,
  opts?: { tagId?: number; setId?: number; storageLocationId?: number | null }
) => {
  const payload: { coinIds: number[]; action: string; tagId?: number; setId?: number; storageLocationId?: number | null } = {
    coinIds,
    action,
  }
  if (opts?.tagId !== undefined) payload.tagId = opts.tagId
  if (opts?.setId !== undefined) payload.setId = opts.setId
  if (opts?.storageLocationId !== undefined) payload.storageLocationId = opts.storageLocationId
  return api.post<{ message: string; affected: number; coins?: Coin[] }>('/coins/bulk', payload)
}

// Journal
export const getJournalEntries = (coinId: number) => api.get<CoinJournal[]>(`/coins/${coinId}/journal`)

export const addJournalEntry = (coinId: number, entry: string) =>
  api.post<CoinJournal>(`/coins/${coinId}/journal`, { entry })

export const deleteJournalEntry = (coinId: number, entryId: number) =>
  api.delete(`/coins/${coinId}/journal/${entryId}`)

// Numista
export const searchNumista = (q: string) => api.get<NumistaSearchResponse>('/numista/search', { params: { q } })

export const proposeNumistaQuery = (request: NumistaQueryProposalRequest) =>
  api.post<NumistaQueryProposal>('/numista/query-proposal', request)

export const lookupNumista = (request: NumistaLookupRequest) =>
  api.post<NumistaLookupOutcome>('/numista/lookup', request)

export const enrichNumista = (request: NumistaEnrichmentRequest, signal?: AbortSignal) =>
  api.post<NumistaLookupOutcome>('/numista/enrich', request, { signal })

// Coin Lookup
export async function lookupCoin(
  images: File[],
  notes = '',
  imageRoles: CoinLookupImageRole[] = [],
  includePriceEstimate = false,
) {
  const formData = new FormData()
  for (const [index, image] of images.entries()) {
    formData.append('images', image)
    const role = imageRoles[index]
    if (role) formData.append('imageRoles', role)
  }
  if (notes.trim()) {
    formData.append('notes', notes.trim())
  }
  if (includePriceEstimate) {
    formData.append('includePriceEstimate', 'true')
  }
  return api.post<CoinLookupResponse>('/coins/lookup', formData)
}

// Value Estimation
export const estimateCoinValue = (coinId: number) =>
  api.post<AIJobStartResponse>(`/coins/${coinId}/estimate-value`)

export const getCoinValueHistory = (coinId: number) =>
  api.get<CoinValueHistory[]>(`/coins/${coinId}/value-history`)

// Images
export const uploadImage = (coinId: number, file: File, imageType: string, isPrimary: boolean, circleClip?: boolean) => {
  const formData = new FormData()
  formData.append('image', file)
  formData.append('imageType', imageType)
  formData.append('isPrimary', String(isPrimary))
  if (circleClip) {
    formData.append('circleClip', 'true')
  }
  return api.post<CoinImage>(`/coins/${coinId}/images`, formData)
}

export const deleteImage = (coinId: number, imageId: number) =>
  api.delete<{ message: string; cleanupPending?: boolean }>(`/coins/${coinId}/images/${imageId}`)

/** Swaps the coin's obverse and reverse images; returns its images. */
export const swapCoinImageSides = (coinId: number) => api.post<CoinImage[]>(`/coins/${coinId}/images/swap-sides`)

// Analysis
export const analyzeCoin = (coinId: number, side?: 'obverse' | 'reverse') => {
  const params = side ? `?side=${side}` : ''
  return api.post<AIJobStartResponse>(`/coins/${coinId}/analyze${params}`)
}

export const gradeCoin = (coinId: number) =>
  api.post<AIJobStartResponse>(`/coins/${coinId}/grade`)

export const getAIJob = (id: string | number) =>
  api.get<AIJob>(`/ai-jobs/${id}`)

export const getCoinAIJobs = (coinId: number, activeOnly = false) =>
  api.get<AIJob[] | { jobs?: AIJob[] }>(`/coins/${coinId}/ai-jobs`, { params: { activeOnly } })

export const deleteAnalysis = (coinId: number, side: 'obverse' | 'reverse') =>
  api.delete<{ coin: Coin }>(`/coins/${coinId}/analyze?side=${side}`)

export const extractText = (file: File) => {
  const formData = new FormData()
  formData.append('image', file)
  return api.post<{ text: string }>('/extract-text', formData)
}

export const getValueHistory = () => api.get<ValueSnapshot[]>('/value-history')

export const getCoinHealthList = (params?: { scope?: 'all' | 'needs_attention'; page?: number; limit?: number }) =>
  api.get<CoinHealthListResponse>('/coins/health', { params })

export const getCoinHealth = (coinId: number) =>
  api.get<CoinHealthItem>(`/coins/${coinId}/health`)

// Autocomplete suggestions
export const getSuggestions = (field: string, q: string) =>
  api.get<string[]>('/suggestions', { params: { field, q } })

export const proxyImage = (url: string) =>
  api.get('/proxy-image', { params: { url }, responseType: 'blob' })

export const scrapeImage = (url: string) =>
  api.get<{ imageUrl: string }>('/scrape-image', { params: { url } })

export const getCoinShipment = (coinId: number) => api.get<ShipmentEnvelopeResponse>(`/coins/${coinId}/shipment`)

export const upsertCoinShipment = (coinId: number, input: ShipmentUpsertInput) => api.put<ShipmentEnvelopeResponse>(`/coins/${coinId}/shipment`, input)

export const deleteCoinShipment = (coinId: number) => api.delete<{ message: string }>(`/coins/${coinId}/shipment`)

export const setCoinShipmentManualOverride = (
  coinId: number,
  input: { enabled: boolean; status: ShipmentStatus; note?: string },
) => api.put<ShipmentEnvelopeResponse>(`/coins/${coinId}/shipment/manual-override`, input)

export const syncCoinShipment = (coinId: number) => api.post<ShipmentEnvelopeResponse>(`/coins/${coinId}/shipment/sync`)

export const updateListingStatus = (coinId: number, status: string) =>
  api.put(`/coins/${coinId}/listing-status`, { status })

export interface CoinCountryCount {
  country: string
  count: number
}

/** Countries of the collection (not wishlist or sold) with coin counts. */
/** The collection's countries, of its coins or its banknotes (every item when kind is omitted). */
export const getCoinCountries = (kind?: CoinKind) => api.get<CoinCountryCount[]>('/coins/countries', kind ? { params: { kind } } : undefined)

export interface CoinCountryFillResult {
  status: NumistaDataStatus
  updated: number
  /** Linked coins still without a country; call again to continue. */
  remaining: number
  failed: number
}

/** Fills missing countries from the coins' Numista types (up to 25 per call). */
export const fillCountriesFromNumista = (lang: string, coinId?: number) =>
  api.post<CoinCountryFillResult>('/coins/country-from-numista', null, { params: { lang, coinId } })

/** One Numista kind of coin and how many collection coins are of it. */
export interface CollectionClassification {
  /** "standard", "circulating-commemorative", "non-circulating" or another kind's label. */
  key: string
  /** Numista's name for it, in the language it was looked up in. */
  label: string
  count: number
}

export interface CollectionSummary {
  /** Coins in the collection (not wishlist or sold). */
  coins: number
  /** Distinct types: same Numista type, or same name when not linked, count once. */
  types: number
  countries: number
  /** Countries of the collection that no longer exist (among countries). */
  extinctCountries?: number
  /** Coins still legal tender, as Numista says of their types. */
  currentCoins?: number
  /** Coins Numista says nothing of (not linked to it), left out of currentCoins. */
  circulationUnknown?: number
  /** Coins whose Numista type commemorates something. */
  commemorativeCoins?: number
  /** Coins Numista says nothing of, left out of commemorativeCoins. */
  commemorativeUnknown?: number
  /** Numista's kinds of coin found in the collection, the usual ones first. */
  classifications?: CollectionClassification[]
  withoutCountry: number
  /** Coins whose presentation is "coincard". */
  coincards: number
  /** Collection currency: totalValue and the highlights' values are in it. */
  currency?: string
  /** Sum of current values in the collection currency. */
  totalValue: number
  /** Coins without a current value, left out of totalValue. */
  unvaluedCoins: number
  /** totalValue in euros (converted at the ECB rate for a dollar collection); null when no rate is available. */
  totalValueEur: number | null
  exchangeRate?: { usdPerEur: number; date: string }
  highlights?: CollectionHighlights
}

/** A notable coin of the collection, shown under the gallery KPIs. */
export interface CoinHighlight {
  id: number
  name: string
  country: string
  year: number | null
  dateRange: string
  rarity: string
  grade: string
  /** In US dollars, like every stored value. */
  currentValue: number | null
  imagePath?: string
}

export interface CollectionHighlights {
  rarest: CoinHighlight | null
  oldest: CoinHighlight | null
  mostValuable: CoinHighlight | null
}

/** The ECB euro reference rates: how much of each currency one euro buys. */
export interface ExchangeRates {
  date: string
  perEur: Record<string, number>
  /** The currencies a collection can be kept in, the euro first. */
  currencies: string[]
}

export const getExchangeRates = () => api.get<ExchangeRates>('/exchange-rates')

/** The collection's headline numbers, for its coins or its banknotes (every item when kind is omitted). */
export const getCollectionSummary = (kind?: CoinKind) => api.get<CollectionSummary>('/coins/summary', kind ? { params: { kind } } : undefined)
