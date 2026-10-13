// Numista catalogue browsing: find a coin type and add it with its details
// filled in. Responses are cached server-side to stay inside Numista's free plan.
import { api } from '@/api/http'
import type { CoinMarketValue, NumistaCatalogPhoto, NumistaDataStatus, NumistaIssueSummary } from '@/api/endpoints/coins'

export interface NumistaCatalogIssuer {
  code: string
  name: string
  /** Wider region or country the issuer belongs to, when different. */
  parent?: string
}

export interface NumistaCatalogSearchItem {
  id: number
  title: string
  issuer: NumistaCatalogIssuer
  minYear?: number
  maxYear?: number
  obverseThumbnail?: string
  reverseThumbnail?: string
}

export interface NumistaCatalogSearchResult {
  status: NumistaDataStatus
  count: number
  page: number
  types: NumistaCatalogSearchItem[]
}

export interface NumistaCatalogSide {
  description?: string
  lettering?: string
}

export interface NumistaCatalogReference {
  catalogue: string
  number: string
}

export interface NumistaCatalogType {
  id: number
  /** Canonical Numista page, as required for a Numista catalogue reference. */
  url: string
  title: string
  issuer: NumistaCatalogIssuer
  minYear?: number
  maxYear?: number
  denomination?: string
  currency?: string
  rulers: string[]
  composition?: string
  technique?: string
  weightGrams?: number
  /** A coin's diameter, or a banknote's width. */
  diameterMm?: number
  /** A banknote's height. */
  heightMm?: number
  thicknessMm?: number
  shape?: string
  /** Banknotes: who printed them, their watermark and the issuing bank. */
  printers?: string[]
  watermark?: string
  issuingEntity?: string
  /** "coin", "banknote" or "exonumia". */
  category?: string
  series?: string
  /** Numista's comments say the type's coins are issued in a coin card. */
  issuedInCoincard?: boolean
  mints: string[]
  obverse: NumistaCatalogSide
  reverse: NumistaCatalogSide
  edge: NumistaCatalogSide
  references: NumistaCatalogReference[]
  photos: NumistaCatalogPhoto[]
}

export interface NumistaCatalogTypeResult {
  status: NumistaDataStatus
  type?: NumistaCatalogType
  issues: NumistaIssueSummary[]
}

export const searchNumistaIssuers = (q: string, lang: string) =>
  api.get<{ status: NumistaDataStatus; issuers: NumistaCatalogIssuer[] }>('/numista/catalog/issuers', { params: { q, lang } })

export interface NumistaCatalogSearchParams {
  q?: string
  issuer?: string
  /** Reference catalogue code ("KM" or "Y"), searched together with number. */
  catalogue?: string
  number?: string
  year?: number
  page?: number
  /** Types per page, 1–50 (default 20). */
  count?: number
  lang: string
  /** Banknotes instead of coins. */
  category?: 'coin' | 'banknote'
}

export const searchNumistaCatalog = (params: NumistaCatalogSearchParams) =>
  api.get<NumistaCatalogSearchResult>('/numista/catalog/search', { params })

export const getNumistaCatalogType = (typeId: number, lang: string) =>
  api.get<NumistaCatalogTypeResult>(`/numista/catalog/types/${typeId}`, { params: { lang } })

export const getNumistaTypeMarketValue = (typeId: number, params: { currency: string; issueId?: number; year?: number }) =>
  api.get<CoinMarketValue>(`/numista/catalog/types/${typeId}/market-value`, { params })
