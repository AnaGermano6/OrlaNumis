import type { CoinSetType } from '@/types/sets'

export type QuickAccessTargetType = 'coin' | 'coin_set'

export interface QuickAccessCoinDTO {
  name: string
  classification: 'owned' | 'wishlist'
  primaryImageUrl: string
}

export interface QuickAccessCoinSetDTO {
  name: string
  setType: CoinSetType
  color: string
  icon: string
}

interface QuickAccessItemBase {
  id: number
  pinnedAt: string
}

export type QuickAccessCoinItem = QuickAccessItemBase & {
  type: 'coin'
  coin: QuickAccessCoinDTO
  coinSet?: never
}

export type QuickAccessCoinSetItem = QuickAccessItemBase & {
  type: 'coin_set'
  coin?: never
  coinSet: QuickAccessCoinSetDTO
}

export type QuickAccessItem =
  | QuickAccessCoinItem
  | QuickAccessCoinSetItem

export interface QuickAccessListDTO {
  items: QuickAccessItem[]
}
