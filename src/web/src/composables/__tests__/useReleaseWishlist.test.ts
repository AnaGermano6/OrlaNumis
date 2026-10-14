import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { CoinRelease } from '@/api/client'
import { setLocale } from '@/i18n'
import { useReleaseWishlist } from '@/composables/useReleaseWishlist'

const mocks = vi.hoisted(() => ({ addCoin: vi.fn(), official: vi.fn() }))
vi.mock('@/api/client', () => ({ addOfficialCoinPhotos: mocks.official }))
vi.mock('@/stores/coins', () => ({ useCoinsStore: () => ({ addCoin: mocks.addCoin }) }))
vi.mock('@/composables/useCoinOptions', async () => {
  const { ref } = await import('vue')
  return { useCoinOptions: () => ({ categoryOptions: ref(['Modern', 'Other']), materialOptions: ref(['Silver', 'Other']), loadOptions: vi.fn().mockResolvedValue(undefined) }) }
})

const ROTARY: CoinRelease = {
  source: 'incm', kind: 'commemorative', country: 'PT', title: 'Rotary', faceValue: 2, imageUrl: 'https://cdn/rotary.jpg',
  pageUrl: 'https://loja.incm.pt/products/rotary', from: '2026-07', to: '2026-07', upcoming: false, owned: false,
}

describe('useReleaseWishlist', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    vi.clearAllMocks()
    mocks.addCoin.mockResolvedValue({ id: 12 })
    mocks.official.mockResolvedValue({ data: { images: [] } })
  })

  it('saves the coin to the wishlist with its official images', async () => {
    const { saveToWishlist } = useReleaseWishlist()

    expect(await saveToWishlist(ROTARY)).toBe('saved')
    expect(mocks.addCoin).toHaveBeenCalledWith(expect.objectContaining({ name: '2 euro 2026 – Rotary', isWishlist: true, referenceUrl: 'https://loja.incm.pt/products/rotary', category: 'Modern' }))
    expect(mocks.official).toHaveBeenCalledWith(12, ['obverse', 'reverse'])

    // The official images are a convenience: the coin is saved without them.
    mocks.official.mockRejectedValue(new Error('offline'))
    expect(await saveToWishlist(ROTARY)).toBe('saved')
  })

  it('says when the wishlist has the coin already, and fails otherwise', async () => {
    const { saveToWishlist } = useReleaseWishlist()

    mocks.addCoin.mockRejectedValue({ response: { status: 409, data: { code: 'wishlist_duplicate' } } })
    expect(await saveToWishlist(ROTARY)).toBe('already')
    expect(mocks.official).not.toHaveBeenCalled()

    mocks.addCoin.mockRejectedValue(new Error('offline'))
    await expect(saveToWishlist(ROTARY)).rejects.toThrow('offline')
  })
})
