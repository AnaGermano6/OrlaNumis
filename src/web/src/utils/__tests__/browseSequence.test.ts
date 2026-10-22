import { beforeEach, describe, expect, it } from 'vitest'
import { browseNeighbours, rememberBrowseSequence } from '@/utils/browseSequence'

describe('browse sequence', () => {
  beforeEach(() => sessionStorage.clear())

  it("gives a coin's place in the remembered list and the coins around it", () => {
    rememberBrowseSequence('gallery', [7, 3, 9, 4])
    expect(browseNeighbours(7)).toEqual({ origin: 'gallery', position: 1, total: 4, previous: null, next: 3 })
    expect(browseNeighbours(9)).toEqual({ origin: 'gallery', position: 3, total: 4, previous: 3, next: 4 })
    expect(browseNeighbours(4)).toEqual({ origin: 'gallery', position: 4, total: 4, previous: 9, next: null })
  })

  it('keeps only the latest list', () => {
    rememberBrowseSequence('gallery', [1, 2, 3])
    rememberBrowseSequence('banknotes', [20, 21])
    expect(browseNeighbours(2)).toBeNull()
    expect(browseNeighbours(21)).toEqual({ origin: 'banknotes', position: 2, total: 2, previous: 20, next: null })
  })

  it('has nothing for a coin out of the list, a list of one, or what it cannot read', () => {
    rememberBrowseSequence('wishlist', [5])
    expect(browseNeighbours(5)).toBeNull()
    expect(browseNeighbours(6)).toBeNull()
    sessionStorage.setItem('coinBrowse:sequence', '{not json')
    expect(browseNeighbours(5)).toBeNull()
    sessionStorage.setItem('coinBrowse:sequence', JSON.stringify({ origin: 'elsewhere', ids: [1, 2] }))
    expect(browseNeighbours(1)).toBeNull()
  })
})
