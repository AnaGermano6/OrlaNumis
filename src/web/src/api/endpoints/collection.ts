// collection endpoints. Split out of the former monolithic client.ts;
// re-exported from '@/api/client' so existing imports keep working.
import { api } from '@/api/http'
import type {
  GeocodeCandidate,
  MintLocation,
} from '@/types'

export type MintLocationInput = {
  displayName: string
  lat: number
  lng: number
  region?: string
  aliases: string[]
}

export type MintLocationsResponse = MintLocation[] | { mintLocations?: MintLocation[] }

// Mint Locations
export const getMintLocations = () => api.get<MintLocationsResponse>('/mint-locations')

export const createMintLocation = (data: MintLocationInput) =>
  api.post<MintLocation>('/mint-locations', data)

export const updateMintLocation = (id: number, data: MintLocationInput) =>
  api.put<MintLocation>(`/mint-locations/${id}`, data)

export const deleteMintLocation = (id: number) => api.delete(`/mint-locations/${id}`)

export const geocodeMintName = (query: string) =>
  api.get<{ candidates: GeocodeCandidate[] }>('/mint-locations/geocode', { params: { query } })
