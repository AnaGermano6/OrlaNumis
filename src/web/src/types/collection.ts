// collection types. Split out of the former single-file src/types/index.ts;
// re-exported from '@/types' so existing imports keep working.

export interface MintLocation {
  id: number
  userId?: number | null
  displayName: string
  lat: number
  lng: number
  region: string
  aliases: string[]
  createdAt: string
  updatedAt: string
}

export interface GeocodeCandidate {
  displayName: string
  lat: number
  lng: number
}
