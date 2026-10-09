import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import CoinMarketValueSection from '../CoinMarketValueSection.vue'
import { setLocale } from '@/i18n'
import type { CoinMarketValue } from '@/api/client'

const mocks = vi.hoisted(() => ({
  getCoinMarketValue: vi.fn(),
}))

vi.mock('@/api/client', () => mocks)

function marketValue(overrides: Partial<CoinMarketValue> = {}): CoinMarketValue {
  return {
    status: 'ok',
    numistaTypeId: 2933,
    numistaUrl: 'https://en.numista.com/catalogue/pieces2933.html',
    issues: [
      { id: 101, year: 1967, mintage: 1000000 },
      { id: 102, year: 1968, mintLetter: 'L', mintage: 5000000 },
    ],
    selectedIssueId: 102,
    currency: 'USD',
    prices: [
      { grade: 'f', price: 0.5 },
      { grade: 'vf', price: 1 },
      { grade: 'unc', price: 2.75 },
    ],
    fetchedAt: '2026-10-09T10:00:00Z',
    stale: false,
    catalogPhotos: [],
    ...overrides,
  }
}

function mountSection(props: { numistaTypeId: number | null; coinGrade?: string } = { numistaTypeId: 2933, coinGrade: 'VF' }) {
  return mount(CoinMarketValueSection, { props: { coinId: 7, ...props } })
}

describe('CoinMarketValueSection', () => {
  beforeEach(() => {
    mocks.getCoinMarketValue.mockReset()
    localStorage.clear()
    setLocale('en')
  })

  afterEach(() => setLocale('en'))

  it('shows the estimated price per grade and highlights the coin grade', async () => {
    mocks.getCoinMarketValue.mockResolvedValue({ data: marketValue() })

    const wrapper = mountSection()
    await flushPromises()

    expect(mocks.getCoinMarketValue).toHaveBeenCalledWith(7, { currency: 'USD', issueId: undefined, lang: 'en' })
    const items = wrapper.findAll('li')
    expect(items.map(item => item.find('span').text())).toEqual(['F', 'VF', 'UNC'])
    expect(items[1]!.text()).toContain('Very Fine')
    expect(items[1]!.text()).toContain('$1.00')
    expect(items[1]!.text()).toContain('Your grade')
    expect(items[0]!.text()).not.toContain('Your grade')
    expect(wrapper.text()).toContain('Community estimates from Numista')
    expect(wrapper.get('a').attributes('href')).toBe('https://en.numista.com/catalogue/pieces2933.html')
  })

  it('reloads for another issue or currency and remembers the choice', async () => {
    mocks.getCoinMarketValue.mockResolvedValue({ data: marketValue() })
    const wrapper = mountSection()
    await flushPromises()

    const [currencySelect, issueSelect] = wrapper.findAll('select')
    expect(issueSelect!.findAll('option').map(option => option.text())).toEqual([
      '1967 · 1,000,000 minted',
      '1968 · L · 5,000,000 minted',
    ])

    await issueSelect!.setValue('101')
    await flushPromises()
    expect(mocks.getCoinMarketValue).toHaveBeenLastCalledWith(7, { currency: 'USD', issueId: 101, lang: 'en' })
    expect(localStorage.getItem('numista:marketValueIssue:7')).toBe('101')

    await currencySelect!.setValue('EUR')
    await flushPromises()
    expect(mocks.getCoinMarketValue).toHaveBeenLastCalledWith(7, { currency: 'EUR', issueId: 101, lang: 'en' })
    expect(localStorage.getItem('numista:marketValueCurrency')).toBe('EUR')
  })

  it('defaults to euros in Portuguese and explains provider problems', async () => {
    setLocale('pt-PT')
    mocks.getCoinMarketValue.mockResolvedValue({ data: marketValue({ status: 'quota_limited', prices: [], fetchedAt: undefined }) })

    const wrapper = mountSection()
    await flushPromises()

    expect(mocks.getCoinMarketValue).toHaveBeenCalledWith(7, { currency: 'EUR', issueId: undefined, lang: 'pt' })
    expect(wrapper.text()).toContain('Foi atingido o limite mensal de pedidos ao Numista')
    expect(wrapper.findAll('li')).toHaveLength(0)
  })

  it('marks stored values that could not be refreshed', async () => {
    mocks.getCoinMarketValue.mockResolvedValue({ data: marketValue({ stale: true }) })

    const wrapper = mountSection()
    await flushPromises()

    expect(wrapper.text()).toContain('Could not refresh; showing the last stored values.')
  })

  it('asks to link a Numista type instead of calling the API when the coin is not linked', async () => {
    const wrapper = mountSection({ numistaTypeId: null })
    await flushPromises()

    expect(mocks.getCoinMarketValue).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Link this coin to a Numista type')
    expect(wrapper.find('select').exists()).toBe(false)
  })

  it('shows a message when the request fails', async () => {
    mocks.getCoinMarketValue.mockRejectedValue(new Error('network'))

    const wrapper = mountSection()
    await flushPromises()

    expect(wrapper.text()).toContain('Market value could not be loaded.')
  })
})
