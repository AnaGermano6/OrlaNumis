import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import CoinCard from '@/components/CoinCard.vue'
import type { Coin } from '@/types'
import { setLocale } from '@/i18n'

vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))

function card(extra: Partial<Coin>) {
  const coin = { id: 1, name: '1 Escudo 1986', category: 'Modern', material: 'Nickel brass', images: [], ...extra } as unknown as Coin
  return mount(CoinCard, { props: { coin }, global: { stubs: { AuthenticatedImage: true, SafeExternalLink: true } } })
}

describe('CoinCard badge', () => {
  beforeEach(() => setLocale('pt-PT'))
  afterEach(() => setLocale('en'))

  it("shows the coin's monetary system without its dates, in place of the category", () => {
    const badge = card({ monetarySystem: 'Escudo (1911-2001)' }).get('[data-testid="coin-card-badge"]')
    expect(badge.text()).toBe('Escudo')
    expect(badge.attributes('title')).toBe('Escudo (1911-2001)')
  })

  it('shows the category while the coin has no monetary system', () => {
    expect(card({}).get('[data-testid="coin-card-badge"]').text()).toBe('Moderna')
  })
})
