// A coin just added to the collection may be one the wishlist still lists
// (saved earlier from the releases, a dealer page…): the user is offered to
// take it off the wishlist.
import { getWishlistMatches } from '@/api/client'
import { t } from '@/i18n'
import { useCoinsStore } from '@/stores/coins'
import { useDialog } from '@/composables/useDialog'
import { useToast } from '@/composables/useToast'

export function useWishlistLeave() {
  const store = useCoinsStore()
  const { showConfirm } = useDialog()
  const { showToast } = useToast()

  /** Asks to take the wishlist coins the collection coin is off the wishlist; never fails. */
  async function offerToLeaveWishlist(coinId: number): Promise<void> {
    let matches: Array<{ id: number; name: string }>
    try {
      matches = (await getWishlistMatches(coinId)).data.coins ?? []
    } catch {
      return
    }
    if (!matches.length) return
    const count = matches.length
    const leave = await showConfirm(t('wishlist.leave.message', { count, names: matches.map((match) => `«${match.name}»`).join(', ') }), {
      title: t('wishlist.leave.title'),
      confirmLabel: t('wishlist.leave.confirm', { count }),
      cancelLabel: t('wishlist.leave.keep'),
    })
    if (!leave) return
    try {
      for (const match of matches) await store.removeCoin(match.id)
      showToast(t('wishlist.leave.done', { count }), 'success')
    } catch {
      showToast(t('wishlist.leave.failed'), 'error')
    }
  }

  return { offerToLeaveWishlist }
}
