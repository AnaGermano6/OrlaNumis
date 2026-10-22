// admin types. Split out of the former single-file src/types/index.ts;
// re-exported from '@/types' so existing imports keep working.
import type { NumistaSettings } from '@/types/numista'

export interface AppSettings extends Partial<NumistaSettings> {
  AIProvider: string
  OllamaURL: string
  OllamaModel: string
  ObversePrompt: string
  ReversePrompt: string
  TextExtractionPrompt: string
  OllamaTimeout: string
  SearXNGURL: string
  LogLevel: string
  PublicAppURL?: string
  DealerSearchSources?: string
  AuctionSearchSources?: string
  RegistrationMode?: string
  WishlistSearchAlertsCheckEnabled?: string
  WishlistSearchAlertsCheckStartTime?: string
  BackupEnabled?: string
  BackupStartTime?: string
  BackupKeep?: string
  CoinCategories?: string
  CoinEras?: string
  CoinCategoryLabelsPtPT?: string
  CoinEraLabelsPtPT?: string
  CoinCopilotEnabled?: string
  CoinCopilotAttributionEnabled?: string
  CoinCopilotWorkerCount?: string
  CoinCopilotMaxActivePerUser?: string
  CoinCopilotQueueDepth?: string
  CoinCopilotMaxReasoningIterations?: string
  CoinCopilotMaxToolCalls?: string
  CoinCopilotHardTimeoutSeconds?: string
  CoinCopilotMaxPersistedToolResultBytes?: string
  CoinCopilotEventRetentionHours?: string
  CoinCopilotCheckpointRetentionDays?: string
  CoinCopilotResumeWindowHours?: string
  ScheduleTimezone?: string
  NumistaValuesEnabled?: string
  NumistaValuesStartTime?: string
  NumistaValuesDailyRequests?: string
  [key: string]: string | undefined
}

export interface SecuritySummary {
  failedLogins: number
  lockedAccounts: number
  recentEvents: number
  loginFailures?: number
}

export interface SecurityEvent {
  id: number | string
  timestamp: string
  type: string
  severity: string
  username?: string | null
  ip?: string | null
  clientIp?: string | null
  outcome?: string | null
  message?: string | null
  userAgent?: string | null
  createdAt?: string
}

export interface SecurityEventFilters {
  type?: string
  severity?: string
  username?: string
  ip?: string
  clientIp?: string
  outcome?: string
  since?: string
  limit?: number
}

export interface SecurityEventsResponse {
  events: SecurityEvent[]
  total?: number
}

export interface SecurityExposureCheck {
  publicIp?: string
  proxy?: boolean
  proxyWarning?: string
  cors?: boolean
  corsWarning?: string
  publicAppUrl?: boolean
  publicAppURL?: boolean
  publicAppUrlWarning?: string
  registration?: boolean
  registrationWarning?: string
  agentToken?: boolean
  agentTokenWarning?: string
  warnings?: string[]
  checks?: Record<string, boolean | string | null | undefined>
  config?: {
    publicAppURL?: string
    trustedProxiesConfigured?: boolean
    agentInternalTokenSet?: boolean
    registrationMode?: string
    backupStatus?: string
  }
}

export type Theme = 'dark' | 'light' | 'british-museum' | 'louvre' | 'capitoline' | 'byzantine' | 'modern-greek'

export const LOG_LEVELS = ['trace', 'debug', 'info', 'warn', 'error'] as const

export interface LogEntry {
  timestamp: string
  level: string
  message: string
}
