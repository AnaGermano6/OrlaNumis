import { useRouter } from 'vue-router'
import { useCoinsStore } from '@/stores/coins'

/** The coin gallery's filters a card or a list can open it with. */
export interface GalleryFilter {
  classification?: string
  circulation?: string
  commemorative?: string
  presentation?: string
  country?: string
  /** Only the coins held more than once (to trade). */
  repeated?: boolean
  /** Only the coins without a grade. */
  ungraded?: boolean
  /** The gallery's search ("Terceira República"). */
  search?: string
}

/**
 * Opens the coin gallery with only the given filters on: the search and every
 * other filter are cleared first.
 */
export function useOpenGallery() {
  const router = useRouter()
  const store = useCoinsStore()
  return (filter: GalleryFilter = {}) => {
    store.searchQuery = filter.search ?? ''
    store.selectedDenomination = ''
    store.selectedRarity = ''
    store.selectedClassification = filter.classification ?? ''
    store.selectedCirculation = filter.circulation ?? ''
    store.selectedCommemorative = filter.commemorative ?? ''
    store.selectedPresentation = filter.presentation ?? ''
    store.selectedCountry = filter.country ?? ''
    store.selectedPeriod = ''
    store.selectedRepeated = filter.repeated ?? false
    store.selectedUngraded = filter.ungraded ?? false
    void router.push('/')
  }
}
