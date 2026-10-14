// Saves a new coin (from the releases or the catalogue) to the wishlist at
// once, with its official images.
import { addOfficialCoinPhotos, type CoinRelease } from '@/api/client'
import { useCoinsStore } from '@/stores/coins'
import { useCoinOptions } from '@/composables/useCoinOptions'
import { buildReleaseWishlistCoin } from '@/utils/releasePrefill'

/** Saved, or already on the wishlist (the server refuses a second coin of the same shop page). */
export type WishlistSaveResult = 'saved' | 'already'

export function useReleaseWishlist() {
  const store = useCoinsStore()
  const { categoryOptions, materialOptions, loadOptions } = useCoinOptions()
  let optionsLoaded: Promise<void> | null = null

  async function saveToWishlist(release: CoinRelease): Promise<WishlistSaveResult> {
    optionsLoaded ??= loadOptions()
    await optionsLoaded
    let coinId: number
    try {
      coinId = (await store.addCoin(buildReleaseWishlistCoin(release, { categories: categoryOptions.value, materials: materialOptions.value }))).id
    } catch (error) {
      const code = (error as { response?: { data?: { code?: string } } })?.response?.data?.code
      if (code === 'wishlist_duplicate') return 'already'
      throw error
    }
    try {
      await addOfficialCoinPhotos(coinId, ['obverse', 'reverse'])
    } catch {
      // The official images are a convenience; the coin is saved.
    }
    return 'saved'
  }

  return { saveToWishlist }
}
