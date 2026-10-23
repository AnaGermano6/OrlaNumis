// admin endpoints. Split out of the former monolithic client.ts;
// re-exported from '@/api/client' so existing imports keep working.
import { api } from '@/api/http'
import type {
  AdminHealthSummaryResponse,
  AdminWishlistSearchAlertRunListResponse,
  AppSettings,
  AvailabilityCycleDetail,
  AvailabilityCycleListResponse,
  AvailabilityCycleTriggerResponse,
  AvailabilityRun,
  CatalogRegistry,
  DeepIdentificationObservabilitySummary,
  LogEntry,
  MintLocation,
  NumistaHealthSummary,
  NumistaValueStatus,
  SecurityEventFilters,
  SecurityEventsResponse,
  SecurityExposureCheck,
  SecuritySummary,
  UserInfo,
  ValuationRun,
} from '@/types'
import type { MintLocationInput } from '@/api/endpoints/collection'

type ConnectivityResult = { available: boolean; message: string }

export const adminCreateMintLocation = (data: MintLocationInput) =>
  api.post<MintLocation>('/admin/mint-locations', data)

export const adminUpdateMintLocation = (id: number, data: MintLocationInput) =>
  api.put<MintLocation>(`/admin/mint-locations/${id}`, data)

export const adminDeleteMintLocation = (id: number) => api.delete(`/admin/mint-locations/${id}`)

export const adminCreateCatalog = (payload: { catalog: string; displayName: string; era: string; volumeRequired: boolean }) =>
  api.post<CatalogRegistry>('/admin/catalogs', payload)

export const adminUpdateCatalog = (id: number, payload: { catalog: string; displayName: string; era: string; volumeRequired: boolean }) =>
  api.put<CatalogRegistry>(`/admin/catalogs/${id}`, payload)

export const adminDeleteCatalog = (id: number) => api.delete(`/admin/catalogs/${id}`)

// Admin
export const getUsers = () => api.get<UserInfo[]>('/admin/users')

export const deleteUser = (id: number) => api.delete(`/admin/users/${id}`)

export const resetUserPassword = (id: number, newPassword: string) =>
  api.post(`/admin/users/${id}/reset-password`, { newPassword })

export const updateUserRole = (id: number, role: UserInfo['role']) =>
  api.put(`/admin/users/${id}/role`, { role })

export const unlockUser = (id: number) => api.post(`/admin/users/${id}/unlock`)

export const getAppSettings = () => api.get<AppSettings>('/admin/settings')

export const getAppSettingDefaults = () => api.get<AppSettings>('/admin/settings/defaults')

export const updateAppSettings = (settings: { key: string; value: string }[]) =>
  api.put('/admin/settings', settings)

/** Which version runs, the latest released one (with its Markdown notes) and whether Watchtower can update the container. */
export interface UpdateStatus {
  current: string
  latest?: string
  updateAvailable: boolean
  notes?: string
  releaseUrl?: string
  publishedAt?: string
  checkedAt?: string
  status: 'ok' | 'disabled' | 'unavailable'
  watchtower: boolean
  image: string
}

export const getUpdateStatus = () => api.get<UpdateStatus>('/admin/updates')

/** Asks GitHub now, instead of using the last check. */
export const checkForUpdates = () => api.post<UpdateStatus>('/admin/updates/check')

/** Backs up, then asks Watchtower to update the container. */
export const applyUpdate = () => api.post<{ message: string }>('/admin/updates/apply')

/** A full backup: the whole database and every upload, in one zip. */
export interface BackupInfo {
  name: string
  /** Made by the schedule, by hand, of the data a restore replaced, or imported. */
  kind: 'auto' | 'manual' | 'pre-restore' | 'pre-update' | 'uploaded'
  /** When its data was copied (an imported backup's own date). */
  createdAt: string
  size: number
  appVersion?: string
  users: number
  /** Coins and banknotes. */
  coins: number
  /** Photos and other uploads. */
  files: number
}

export interface BackupsResponse {
  backups: BackupInfo[]
  running: boolean
  enabled: boolean
  startTime: string
  keep: number
  /** Seconds to the next automatic backup. */
  nextRunIn: number
  folder: string
  lastRestore?: { name: string; at: string; ok: boolean; error?: string }
}

export const getBackups = () => api.get<BackupsResponse>('/admin/backups')

/** Starts a backup, made in the background (`running` until it is listed). */
export const createBackup = () => api.post('/admin/backups')

export const downloadBackup = (name: string, onProgress?: (fraction: number) => void) =>
  api.get<Blob>(`/admin/backups/${encodeURIComponent(name)}`, {
    responseType: 'blob',
    onDownloadProgress: (event) => { if (event.total) onProgress?.(event.loaded / event.total) },
  })

export const deleteBackup = (name: string) => api.delete(`/admin/backups/${encodeURIComponent(name)}`)

/** Prepares the backup to replace all the data; the app then restarts to apply it. */
export const restoreBackup = (name: string) => api.post<{ restarting: boolean }>(`/admin/backups/${encodeURIComponent(name)}/restore`)

/** Keeps a downloaded backup in the backup folder, after checking it; restoring it is a second step. */
export const uploadBackup = (file: File, onProgress?: (fraction: number) => void) =>
  api.put<BackupInfo>('/admin/backups/upload', file, {
    headers: { 'Content-Type': 'application/zip' },
    onUploadProgress: (event) => { if (event.total) onProgress?.(event.loaded / event.total) },
  })

export const getAdminNumistaHealth = () =>
  api.get<NumistaHealthSummary>('/admin/numista/health')

export const getAdminDeepIdentificationObservability = () =>
  api.get<DeepIdentificationObservabilitySummary>('/admin/deep-identification/observability')

export const getSecuritySummary = () =>
  api.get<SecuritySummary | { summary?: Partial<SecuritySummary>; backupStatus?: string }>('/admin/security/summary')

export const getSecurityEvents = (filters?: SecurityEventFilters) => {
  const params = filters ? { ...filters } : undefined
  if (params?.ip && !params.clientIp) {
    params.clientIp = params.ip
    delete params.ip
  }
  return api.get<SecurityEventsResponse>('/admin/security/events', { params })
}

export const getSecurityExposureCheck = () =>
  api.get<SecurityExposureCheck>('/admin/security/exposure-check')

export const getAdminLogs = (limit = 500, level?: string) => {
  const params: Record<string, string> = { limit: String(limit) }
  if (level) params.level = level
  return api.get<{ logs: LogEntry[]; count: number; logLevel: string }>('/admin/logs', { params })
}

export const testAnthropicConnection = () =>
  api.get<ConnectivityResult>('/admin/test-anthropic')

export const testSearXNGConnection = () =>
  api.get<ConnectivityResult>('/admin/test-searxng')

export const getAdminHealthSummary = () =>
  api.get<AdminHealthSummaryResponse>('/admin/health/summary')

// Legacy admin availability runs (pre-Feature-353 UserID=0 aggregate rows).
// Retained read-only for the "Legacy" section of the admin schedule history.
export const getAvailabilityRuns = (page = 1, limit = 20) =>
  api.get<{ runs: AvailabilityRun[]; total: number }>('/admin/availability-runs', { params: { page, limit } })

export const getAvailabilityRunDetail = (runId: number) =>
  api.get<AvailabilityRun>(`/admin/availability-runs/${runId}`)

export const triggerAvailabilityCheck = () =>
  api.post<AvailabilityCycleTriggerResponse>('/admin/availability/run')

// Admin availability cycles (Feature 353) — parent roll-up rows with expandable per-user children.
export const getAvailabilityCycles = (page = 1, limit = 5) =>
  api.get<AvailabilityCycleListResponse>('/admin/availability-cycles', { params: { page, limit } })

export const getAvailabilityCycleDetail = (cycleId: number) =>
  api.get<AvailabilityCycleDetail>(`/admin/availability-cycles/${cycleId}`)

export const getAdminWishlistSearchAlertRuns = (page = 1, limit = 5) =>
  api.get<AdminWishlistSearchAlertRunListResponse>('/admin/wishlist-search-alert-runs', { params: { page, limit } })

// Valuation Runs
export const getValuationRuns = (page = 1, limit = 20) =>
  api.get<{ runs: ValuationRun[]; total: number }>('/admin/valuation-runs', { params: { page, limit } })

export const getValuationRunDetail = (runId: number) =>
  api.get<ValuationRun>(`/admin/valuation-runs/${runId}`)

export const triggerValuation = () =>
  api.post<{ message: string; users: number }>('/admin/valuation-runs/trigger')

export const cancelValuationRun = (runId: number) =>
  api.post<{ message: string }>(`/admin/valuation-runs/${runId}/cancel`)

// The daily batches of coin values from Numista's prices.
export const getNumistaValuesStatus = () => api.get<NumistaValueStatus>('/admin/numista-values')

export const runNumistaValuesNow = () => api.post<{ message: string }>('/admin/numista-values/run')

/** Result of switching the collection currency. */
export interface CollectionCurrencyChange {
  currency: string
  previous: string
  /** What every amount was multiplied by, at the ECB rates of rateDate; absent when nothing was converted. */
  factor?: number
  rateDate?: string
  coinsConverted: number
}

/** Switches the collection currency, converting every stored price and value. */
export const changeCollectionCurrency = (currency: string) =>
  api.post<CollectionCurrencyChange>('/admin/collection-currency', { currency })
