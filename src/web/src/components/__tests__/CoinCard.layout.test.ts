import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import CoinCard from '@/components/CoinCard.vue'
import type { Coin } from '@/types'
import { setLocale } from '@/i18n'

vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))

function card(extra: Partial<Coin>, props: Record<string, unknown> = {}) {
  const coin = {
    id: 1, name: '1 Escudo', category: 'Modern', material: 'Bronze', grade: '', quantity: 1,
    currentValue: null, purchasePrice: null, images: [], ...extra,
  } as unknown as Coin
  return mount(CoinCard, { props: { coin, ...props }, global: { stubs: { AuthenticatedImage: true, SafeExternalLink: true } } })
}

describe('CoinCard layout', () => {
  beforeEach(() => setLocale('pt-PT'))
  afterEach(() => setLocale('en'))

  it('shows its year, material and kind under its name', () => {
    const details = card({ year: 1924, monetarySystem: 'Escudo (1911-2001)' }).get('[data-testid="coin-card-details"]')
    expect(details.text()).toBe('1924 · Bronze · Escudo')
  })

  it('shows its dates when it has no year, and leaves out an unknown material', () => {
    const details = card({ year: null, dateRange: '1910-1920', material: 'Other' } as Partial<Coin>).get('[data-testid="coin-card-details"]')
    expect(details.text()).toBe('1910-1920 · Moderna')
  })

  it('shows its grade in short in the corner, with its name on hover', () => {
    const grade = card({ grade: 'vf' }).get('[data-testid="coin-card-grade"]')
    expect(grade.text()).toBe('VF')
    expect(card({ grade: 'UNC' }).get('[data-testid="coin-card-grade"]').attributes('title')).toContain('UNC — ')
    expect(card({ grade: '' }).find('[data-testid="coin-card-grade"]').exists()).toBe(false)
  })

  it('shows how many and what it is worth at the bottom', () => {
    const wrapper = card({ quantity: 3, currentValue: 12.5 })
    expect(wrapper.get('[data-testid="coin-card-quantity"]').text()).toBe('3x')
    expect(wrapper.get('[data-testid="coin-card-value"]').text()).toContain('12,50')
    expect(card({}).find('[data-testid="coin-card-value"]').exists()).toBe(false)
  })

  it('shows the price asked on the wishlist, without a grade or a count', () => {
    const wrapper = card({ purchasePrice: 8, grade: 'VF', quantity: 2 }, { wishlist: true })
    expect(wrapper.get('[data-testid="coin-card-value"]').text()).toContain('8,00')
    expect(wrapper.find('[data-testid="coin-card-grade"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="coin-card-quantity"]').text()).toBe('')
  })
})
