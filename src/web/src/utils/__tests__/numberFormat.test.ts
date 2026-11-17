import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import { formatNumber, formatShare, numberFormat } from '@/utils/numberFormat'

describe('numberFormat', () => {
  afterEach(() => setLocale('en'))

  it('always groups the thousands, Portuguese ones too', () => {
    setLocale('pt-PT')
    // Portuguese on its own groups only from 10 000.
    expect(numberFormat({ style: 'currency', currency: 'EUR' }).format(3874.51)).toBe('3 874,51 €')
    expect(formatNumber(1600)).toBe('1 600')
    expect(formatNumber(12345.6, { maximumFractionDigits: 1 })).toBe('12 345,6')
    expect(formatNumber(950)).toBe('950')

    setLocale('en')
    expect(formatNumber(3874.51)).toBe('3,874.51')
  })

  it('gives a share a decimal only where a whole number would mislead', () => {
    setLocale('pt-PT')
    expect(formatShare(14 / 4208)).toBe('0,3%')
    expect(formatShare(0)).toBe('0%')
    expect(formatShare(0.4)).toBe('40%')
    expect(formatShare(0.996)).toBe('99,6%')
    expect(formatShare(1)).toBe('100%')
  })

  it('keeps none when asked for none', () => {
    setLocale('pt-PT')
    expect(numberFormat({ useGrouping: false }).format(1000)).toBe('1000')
  })
})
