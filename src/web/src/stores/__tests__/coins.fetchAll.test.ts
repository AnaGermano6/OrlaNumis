import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const mocks = vi.hoisted(() => ({ getCoins: vi.fn() }))
vi.mock('@/api/client', () => mocks)

import { useCoinsStore } from '../coins'

function page(ids: number[], total: number) {
  return { data: { coins: ids.map((id) => ({ id, name: `Coin ${id}` })), total } }
}

describe('coins store fetchAllCoins', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mocks.getCoins.mockReset()
  })

  it('loads every page with the filters and reports the full total', async () => {
    mocks.getCoins
      .mockResolvedValueOnce(page(Array.from({ length: 100 }, (_, i) => i + 1), 150))
      .mockResolvedValueOnce(page(Array.from({ length: 50 }, (_, i) => i + 101), 150))
    const store = useCoinsStore()

    await store.fetchAllCoins({ country: 'Portugal', wishlist: 'false', sold: 'false' })

    expect(mocks.getCoins).toHaveBeenNthCalledWith(1, { country: 'Portugal', wishlist: 'false', sold: 'false', page: 1, limit: 100 })
    expect(mocks.getCoins).toHaveBeenNthCalledWith(2, expect.objectContaining({ page: 2, limit: 100 }))
    expect(store.coins).toHaveLength(150)
    expect(store.total).toBe(150)
    expect(store.truncated).toBe(false)
    expect(store.loading).toBe(false)
  })

  it('ignores a grouped load that a newer paged load replaced', async () => {
    let finish!: (value: unknown) => void
    mocks.getCoins.mockImplementationOnce(() => new Promise((resolve) => { finish = resolve }))
    mocks.getCoins.mockResolvedValueOnce(page([7], 1))
    const store = useCoinsStore()

    const grouped = store.fetchAllCoins({})
    await store.fetchCoins({ page: 1 })
    finish(page([1, 2, 3], 3))
    await grouped

    expect(store.coins.map((coin) => coin.id)).toEqual([7])
  })
})
