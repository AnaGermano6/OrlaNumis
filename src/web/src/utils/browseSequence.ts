// The order of the coins on the list a coin's page was opened from (the coin
// gallery, the banknotes, the wishlist), as that list shows them: its
// filters, sort and groups. The coin's page goes to the previous or the next
// one from it. Kept for the browser tab, so a reload keeps it.

export type BrowseOrigin = 'gallery' | 'banknotes' | 'wishlist'

export interface BrowseNeighbours {
  origin: BrowseOrigin
  /** The coin's place in the list, from 1. */
  position: number
  total: number
  previous: number | null
  next: number | null
}

const STORAGE_KEY = 'coinBrowse:sequence'
const MAX_IDS = 5000
const ORIGINS: BrowseOrigin[] = ['gallery', 'banknotes', 'wishlist']

/** Remembers the order a list shows its coins in. */
export function rememberBrowseSequence(origin: BrowseOrigin, ids: number[]): void {
  try {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify({ origin, ids: ids.slice(0, MAX_IDS) }))
  } catch {
    // No session storage (private window, blocked): no previous and next.
  }
}

/** The coins before and after one in the remembered list; null when it is not in it. */
export function browseNeighbours(coinId: number): BrowseNeighbours | null {
  let saved: unknown
  try {
    saved = JSON.parse(sessionStorage.getItem(STORAGE_KEY) ?? 'null')
  } catch {
    return null
  }
  if (!saved || typeof saved !== 'object') return null
  const { origin, ids } = saved as { origin?: unknown; ids?: unknown }
  if (!ORIGINS.includes(origin as BrowseOrigin) || !Array.isArray(ids)) return null
  const list = ids.filter((id): id is number => Number.isInteger(id))
  const index = list.indexOf(coinId)
  if (index < 0 || list.length < 2) return null
  return {
    origin: origin as BrowseOrigin,
    position: index + 1,
    total: list.length,
    previous: index > 0 ? list[index - 1]! : null,
    next: index < list.length - 1 ? list[index + 1]! : null,
  }
}
