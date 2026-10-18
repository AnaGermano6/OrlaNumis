// Saves a new coin (from the releases or the catalogue) to the wishlist at
// once, with its official images; and takes it off again.
import { addOfficialCoinPhotos, type CoinRelease } from '@/api/client'
import { useCoinsStore } from '@/stores/coins'
import { useCoinOptions } from '@/composables/useCoinOptions'
import { buildReleaseWishlistCoin } from '@/utils/releasePrefill'

/** Saved, or already on the wishlist (the server refuses a second coin of the same shop page). */
export type WishlistSaveResult = 'saved' | 'already'

export interface WishlistSave {
  result: WishlistSaveResult
  /** The wishlist coin: the one saved, or the one already there. */
  coinId?: number
}

export function useReleaseWishlist() {
  const store = useCoinsStore()
  const { categoryOptions, materialOptions, loadOptions } = useCoinOptions()
  let optionsLoaded: Promise<void> | null = null

  async function saveToWishlist(release: CoinRelease): Promise<WishlistSave> {
    optionsLoaded ??= loadOptions()
    await optionsLoaded
    let coinId: number
    try {
      coinId = (await store.addCoin(buildReleaseWishlistCoin(release, { categories: categoryOptions.value, materials: materialOptions.value }))).id
    } catch (error) {
      const data = (error as { response?: { data?: { code?: string; coinId?: number } } })?.response?.data
      if (data?.code === 'wishlist_duplicate') return { result: 'already', coinId: data.coinId }
      throw error
    }
    try {
      await addOfficialCoinPhotos(coinId, ['obverse', 'reverse'])
    } catch {
      // The official images are a convenience; the coin is saved.
    }
    return { result: 'saved', coinId }
  }

  /** Takes a coin off the wishlist: a wishlist coin is only the wish, so it is deleted. */
  async function removeFromWishlist(coinId: number): Promise<void> {
    await store.removeCoin(coinId)
  }

  return { saveToWishlist, removeFromWishlist }
}
