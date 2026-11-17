import { beforeEach, describe, expect, it } from 'vitest'
import { coinSetProposals } from '@/utils/coinSetProposals'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'

function coin(fields: Partial<Coin>): Coin {
  return { id: 1, name: 'Coin', kind: 'coin', country: 'Portugal', isWishlist: false, images: [], ...fields } as unknown as Coin
}

// The proposals as their names, with plain spaces.
function names(proposals: ReturnType<typeof coinSetProposals>): string[] {
  return proposals.map((proposal) => proposal.name.replace(/\s/g, ' '))
}

describe('coinSetProposals', () => {
  beforeEach(() => setLocale('pt-PT'))

  it("proposes a €2 commemorative's album, its country's commemoratives and its year", () => {
    const proposals = coinSetProposals(coin({ monetarySystem: 'Euro', denomination: '2 euro', year: 2026 }), { current: true, commemorative: true })
    expect(names(proposals)).toEqual(['Portugal · 2 €', 'Comemorativas de Portugal', 'Moedas de 2026'])
    expect(proposals[0]!.request).toMatchObject({ name: 'Portugal · 2 €'.replace(' €', ' €'), setType: 'checklist', checklist: { source: 'catalogue', country: 'PT', faceValue: 2 } })
    // The coins only: Portugal's banknotes are not commemoratives of Portugal.
    expect(proposals[1]!.request).toMatchObject({ setType: 'smart', smartCriteria: { operator: 'and', rules: [
      { field: 'country', op: 'eq', value: 'Portugal' },
      { field: 'classification', op: 'eq', value: 'commemorative' },
      { field: 'kind', op: 'eq', value: 'coin' },
      { field: 'isWishlist', op: 'eq', value: false },
    ] } })
    expect(proposals.map((proposal) => proposal.kind)).toEqual(['album', 'smart', 'smart'])
  })

  it("proposes a circulation euro coin's annual series, its year and its country", () => {
    const proposals = coinSetProposals(coin({ monetarySystem: 'Euro', denomination: '50 cêntimos', year: 2002 }), { current: true, commemorative: false })
    expect(names(proposals)).toEqual(['Portugal 2002: série anual', 'Moedas de 2002', 'Moedas de Portugal'])
    expect(proposals[0]!.request.checklist).toEqual({ source: 'challenge', challenge: 'annual-series', country: 'PT', year: 2002 })
  })

  it("proposes a collector coin's album of its face value", () => {
    const proposals = coinSetProposals(coin({ monetarySystem: 'Euro', denomination: '7,5 euros', year: 2026 }), { current: true, nonCirculating: true, commemorative: false })
    expect(proposals[0]!.request.checklist).toEqual({ source: 'catalogue', country: 'PT', faceValue: 7.5 })
  })

  it("proposes Portugal's escudos and réis albums", () => {
    expect(coinSetProposals(coin({ monetarySystem: 'Escudo (1911-2001)', denomination: '100 escudos', year: 1989 }), { current: false, commemorative: false })[0]!.request.checklist)
      .toEqual({ source: 'challenge', challenge: 'escudo-types' })
    const commemorative = coinSetProposals(coin({ monetarySystem: 'Escudo (1911-2001)', denomination: '1000 escudos', year: 1998 }), { current: false, commemorative: true })
    expect(names(commemorative)[0]).toBe('Portugal · Escudos')
    expect(commemorative[0]!.request.checklist).toEqual({ source: 'catalogue', country: 'PT', currency: 'PTE' })
    expect(coinSetProposals(coin({ monetarySystem: 'Real (decimalizado, 1835-1910)', denomination: '1000 réis', year: 1899 }), null)[0]!.request.checklist)
      .toEqual({ source: 'challenge', challenge: 'reis-types' })
  })

  it('proposes a banknote the sets of the collection\'s banknotes: of its year, of its country', () => {
    const proposals = coinSetProposals(coin({ kind: 'banknote', monetarySystem: 'Escudo (1911-2001)', denomination: '1000 escudos', year: 1996 }), { current: false, commemorative: false })
    expect(names(proposals)).toEqual(['Notas de 1996', 'Notas de Portugal'])
    expect(proposals[1]!.request.smartCriteria).toEqual({ operator: 'and', rules: [
      { field: 'country', op: 'eq', value: 'Portugal' },
      { field: 'kind', op: 'eq', value: 'banknote' },
      { field: 'isWishlist', op: 'eq', value: false },
    ] })
  })

  it('has no album for a coin outside the catalogue, and nothing for a wished coin', () => {
    expect(names(coinSetProposals(coin({ country: 'Uganda', monetarySystem: 'Shilling (1966-date)', denomination: '5 shillings', year: 1987 }), null)))
      .toEqual(['Moedas de 1987', 'Moedas de Uganda'])
    expect(coinSetProposals(coin({ isWishlist: true, monetarySystem: 'Euro', denomination: '2 euro', year: 2026 }), null)).toEqual([])
  })
})
