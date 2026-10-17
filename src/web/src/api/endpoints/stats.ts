// stats endpoints. Split out of the former monolithic client.ts;
// re-exported from '@/api/client' so existing imports keep working.
import { api } from '@/api/http'
import type {
  TimeMachineBounds,
  TimeMachineSnapshot,
  CollectionHealthSummary,
  StatsResponse,
} from '@/types'

// Stats
export const getStats = () => api.get<StatsResponse>('/stats')

export const getDistribution = () => api.get<{ cells: { era: string; category: string; count: number }[] }>('/stats/distribution')

export const getCollectionHealthSummary = () => api.get<CollectionHealthSummary>('/stats/health')

// Coin of the Day

// Collection Time Machine
export const getTimeMachineSnapshot = (date: string) =>
  api.get<TimeMachineSnapshot>('/stats/time-machine', { params: { date } })

export const getTimeMachineBounds = () =>
  api.get<TimeMachineBounds>('/stats/time-machine/bounds')
