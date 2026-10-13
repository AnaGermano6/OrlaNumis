import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import { setCollectionCurrency } from '@/utils/collectionCurrency'
import { priceRangeLabel } from '@/utils/priceRanges'

describe('priceRangeLabel', () => {
  afterEach(() => {
    setLocale('en')
    setCollectionCurrency('USD')
  })

  it('names the ranges in the currency every page uses', () => {
    setLocale('pt-PT')
    setCollectionCurrency('EUR')
    const labels = ['Under $50', '$50 - $200', '$1K+', 'Unknown'].map((range) => priceRangeLabel(range).replace(/\s/g, ' '))
    expect(labels).toEqual(['Menos de 50 €', '50 € – 200 €', '1000 € ou mais', 'Unknown'])

    setLocale('en')
    setCollectionCurrency('USD')
    expect(priceRangeLabel('$200 - $500')).toBe('$200 – $500')
  })
})
