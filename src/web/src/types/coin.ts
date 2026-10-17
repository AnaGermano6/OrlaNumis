// coin types. Split out of the former single-file src/types/index.ts;
// re-exported from '@/types' so existing imports keep working.
import type { MintLocation } from '@/types/collection'
import type { NumistaEvidence, NumistaLookupOutcome, SelectedNumistaReference } from '@/types/numista'
import type { CoinSet } from '@/types/sets'

/** What a collection item is; the galleries show coins and banknotes apart. */
export type CoinKind = 'coin' | 'banknote'

export interface Coin {
  id: number
  name: string
  /** "coin" unless the item is a banknote. */
  kind?: CoinKind
  category: Category
  denomination: string
  ruler: string
  era: string
  dateRange?: string | null
  mint: string
  mintLocationId: number | null
  mintLocation: Pick<MintLocation, 'id' | 'displayName' | 'lat' | 'lng'> | null
  material: Material
  weightGrams: number | null
  diameterMm: number | null
  grade: string
  /** Banknote details: series or issue ("Chapa 13"), the note's serial number, signatories, printer. */
  series?: string
  serialNumber?: string
  signatures?: string
  printer?: string
  widthMm?: number | null
  heightMm?: number | null
  watermark?: string
  /** A replacement (star) note or a specimen; '' for an ordinary note. */
  noteVariant?: '' | 'replacement' | 'specimen'
  /** A grading company's grade ("PMG 66 EPQ") and its certificate number. */
  certifiedGrade?: string
  certificationNumber?: string
  obverseInscription: string
  reverseInscription: string
  obverseDescription: string
  reverseDescription: string
  rarityRating: string
  /** First year of dateRange (negative for BC), derived by the server. */
  year?: number | null
  /** Issuing country or authority. */
  country?: string
  /** How many of this coin or note the collector has (1 or more). */
  quantity: number
  /** How the coin is kept or was issued, e.g. "coincard" (see COIN_PRESENTATIONS). */
  presentation?: string
  /** Fixed rarity scale (see COIN_RARITIES); rarityRating is a legacy free-text field. */
  rarity?: string
  /** Circulation set by hand (see COIN_CIRCULATIONS); empty means "as Numista says". */
  circulation?: string
  purchasePrice: number | null
  currentValue: number | null
  purchaseDate: string | null
  purchaseLocation: string
  vendorSku: string
  vendorInvoice: string
  notes: string
  aiAnalysis: string
  obverseAnalysis: string
  reverseAnalysis: string
  referenceUrl: string
  referenceText: string
  isWishlist: boolean
  listingStatus: string
  listingCheckedAt: string | null
  listingCheckReason: string
  sourceAlertCandidateId?: number | null
  userId: number
  images: CoinImage[]
  references?: CoinReference[]
  sets?: CoinSet[]
  createdAt: string
  updatedAt: string
}

export interface CoinReference {
  id: number
  coinId: number
  catalog: string
  volume: string
  number: string
  uri: string
  createdAt: string
  updatedAt: string
}

export interface CoinReferenceInput {
  catalog: string
  volume?: string
  number: string
  uri?: string
}

export interface CatalogRegistry {
  id: number
  catalog: string
  displayName: string
  era: 'ancient' | 'medieval' | 'modern'
  volumeRequired: boolean
  createdAt?: string
  updatedAt?: string
}

export interface LegacyMigrationResult {
  succeeded: number
  skipped: number
  failed: number
  message?: string
}

export type CoinMutationPayload = Partial<Omit<Coin, 'references' | 'mintLocation'>> & {
  references?: CoinReferenceInput[]
}

export type IntakeConfidenceLevel = 'low' | 'medium' | 'high'

export interface IntakeConfidenceSummary {
  overall: IntakeConfidenceLevel
  uncertainFields: string[]
}

export interface IntakeEvidenceItem {
  type: string
  source: string
  field: string
  value: string
  confidence: IntakeConfidenceLevel
  notes?: string
}

export interface IntakeDraft {
  draftId: number
  status: 'drafted' | 'confirmed' | 'discarded' | 'expired'
  coin: CoinMutationPayload
  confidenceSummary: IntakeConfidenceSummary
  evidence: IntakeEvidenceItem[]
  unresolvedFields: string[]
  expiresAt: string
}

export interface CoinLookupNGCData {
  certNumber: string
  normalizedCert: string
  lookupURL: string
  grade?: string
  description?: string
}

export interface CoinLookupExtractedData {
  ngc?: CoinLookupNGCData
  labelText?: string
  coinFields?: Record<string, unknown>
  confidence: IntakeConfidenceLevel
  rawAnalysis: string
}

export interface CoinLookupNumistaCandidate {
  id: string
  title: string
  issuer: string
  year: string
  thumbnail?: string
  url: string
}

export interface CoinLookupResponse {
  extractedData: CoinLookupExtractedData
  numistaCandidates: CoinLookupNumistaCandidate[]
  proposedNumistaQuery?: string
  numistaEvidence?: NumistaEvidence
  numistaLookup?: NumistaLookupOutcome | null
  prefilledDraft?: CoinMutationPayload
  candidateReferences?: CoinReferenceInput[]
  /** Present only when the collector opted in; model guess, not market comps. */
  priceEstimate?: CoinLookupPriceEstimate
  /** Source-backed range from current dealer listings; preferred over priceEstimate. */
  priceComparables?: CoinLookupPriceComparables
}

export interface CoinLookupComparableListing {
  title: string
  url: string
  sourceName?: string
  price: number
  currency: string
}

export interface CoinLookupPriceComparables {
  low: number
  high: number
  currency: string
  count: number
  query?: string
  listings: CoinLookupComparableListing[]
}

export interface CoinLookupPriceEstimate {
  low: number
  high: number
  currency: string
  basis?: string
}

export type CoinLookupImageRole = 'obverse' | 'reverse' | 'notes'

export interface IntakeCommitRequest {
  draftId: number
  confirm: boolean
  overrides?: CoinMutationPayload
}

export interface IntakeCommitResponse {
  draftId: number
  status: 'confirmed'
  coinId: number
}

export type QuickCaptureDraftStatus = 'active' | 'promoting' | 'promoted' | 'discarded'

export type QuickCaptureImageType = 'obverse' | 'reverse' | 'detail' | 'other'

export interface QuickCaptureDraftImage {
  id: number
  draftId: number
  filePath: string
  imageType: QuickCaptureImageType
  isPrimary: boolean
  displayOrder: number
  createdAt: string
}

export interface QuickCaptureDraft {
  id: number
  userId: number
  workingTitle: string
  dateRange: string
  era: string
  acquisitionSource: string
  purchasePrice: number | null
  notes: string
  source: string
  ngcCertNumber: string
  ngcLookupUrl: string
  ngcGrade: string
  labelText: string
  aiConfidence: string
  selectedNumistaReference?: SelectedNumistaReference | null
  status: QuickCaptureDraftStatus
  promotedCoinId: number | null
  promotedAt: string | null
  discardedAt: string | null
  images: QuickCaptureDraftImage[]
  createdAt: string
  updatedAt: string
}

export interface QuickCaptureDraftListResponse {
  drafts: QuickCaptureDraft[]
  total: number
  page: number
  limit: number
}

export interface QuickCaptureDraftInput {
  workingTitle?: string
  dateRange?: string
  era?: string
  acquisitionSource?: string
  purchasePrice?: number | null
  notes?: string
  source?: string
  ngcCertNumber?: string
  ngcLookupUrl?: string
  ngcGrade?: string
  labelText?: string
  aiConfidence?: string
  selectedNumistaId?: string
  selectedNumistaUrl?: string
  obverseImage?: File | null
  reverseImage?: File | null
  detailImages?: File[]
}

export interface QuickCaptureDraftUpdateInput {
  workingTitle: string
  dateRange: string
  era: string
  acquisitionSource: string
  purchasePrice: number | null
  notes: string
  source?: string
  ngcCertNumber?: string
  ngcLookupUrl?: string
  ngcGrade?: string
  labelText?: string
  aiConfidence?: string
  selectedNumistaId?: string
  selectedNumistaUrl?: string
  clearSelectedNumista?: boolean
  removeImageIds?: string // comma-separated IDs
  replaceObverse?: boolean
  replaceReverse?: boolean
  obverseImage?: File | null
  reverseImage?: File | null
  detailImages?: File[]
}

export interface QuickCapturePromoteOverrides {
  name?: string
  category?: string
  material?: string
  era?: string
  purchasePrice?: number | null
  purchaseLocation?: string
  notes?: string
}

export interface QuickCapturePromoteRequest {
  confirm: boolean
  target?: 'collection' | 'wishlist'
  overrides?: QuickCapturePromoteOverrides
}

export interface QuickCapturePromotionResponse {
  draftId: number
  status: 'promoted'
  coinId: number
  alreadyPromoted: boolean
  target: 'collection' | 'wishlist'
}

export interface CoinImage {
  id: number
  coinId: number
  filePath: string
  imageType: ImageType
  isPrimary: boolean
  createdAt: string
}

export type Category = string

// Coins are struck in metal; banknotes are printed on paper, polymer or both.
export type Material = 'Gold' | 'Silver' | 'Bronze' | 'Copper' | 'Electrum' | 'Other' | 'Paper' | 'Polymer' | 'Hybrid'

export type ImageType = 'obverse' | 'reverse' | 'detail' | 'other'

export type CoinEra = string

export const CATEGORIES: Category[] = ['Modern', 'Other']

export const MATERIALS: Material[] = ['Gold', 'Silver', 'Bronze', 'Copper', 'Electrum', 'Other']

export const COIN_ERAS: CoinEra[] = ['ancient', 'medieval', 'modern']

export const CATEGORY_COLORS: Record<string, string> = {
  Modern: '#4682b4',
  Other: '#888888',
}

export interface CoinListResponse {
  coins: Coin[]
  total: number
  page: number
  limit: number
}

export interface CoinJournal {
  id: number
  coinId: number
  userId: number
  entry: string
  createdAt: string
}

export interface ValueComparable {
  source: string
  price: string
  url: string
}

export interface ValueEstimate {
  estimatedValue: number
  confidence: 'high' | 'medium' | 'low'
  reasoning: string
  comparables: ValueComparable[]
}

export interface CoinValueHistory {
  id: number
  coinId: number
  userId: number
  value: number
  confidence: string
  recordedAt: string
  /** Additive — absent on legacy rows; infer from confidence when missing. */
  source?: 'ai_scheduled' | 'ai_estimate' | 'manual' | string
}

export interface CoinDetailMetadataRow {
  key: string
  label: string
  value: string
  valueClass?: string
  fullWidth?: boolean
  url?: string | null
}

