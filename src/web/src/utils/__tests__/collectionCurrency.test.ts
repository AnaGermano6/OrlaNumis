import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import { collectionCurrencySymbol, setCollectionCurrency, toCollectionCurrency } from '@/utils/collectionCurrency'
import { formatCurrency } from '@/utils/format'

describe('collection currency', () => {
  afterEach(() => {
    setCollectionCurrency('USD')
    setLocale('en')
  })

  it('formats amounts in the collection currency unless told otherwise', () => {
    setLocale('pt-PT')
    expect(formatCurrency(3.37).replace(/\s/g, ' ')).toBe('3,37 US$')
    setCollectionCurrency('EUR')
    expect(formatCurrency(3.01).replace(/\s/g, ' ')).toBe('3,01 €')
    expect(formatCurrency(3.37, 'USD').replace(/\s/g, ' ')).toBe('3,37 US$')
    expect(collectionCurrencySymbol()).toBe('€')
    // Unknown codes leave the currency as it was.
    setCollectionCurrency('GBP')
    expect(collectionCurrencySymbol()).toBe('€')
  })

  it('converts between dollars and euros with the ECB rate', () => {
    expect(toCollectionCurrency(10, 'USD', undefined)).toBe(10)
    expect(toCollectionCurrency(10, 'EUR', 1.25)).toBe(12.5)
    expect(toCollectionCurrency(10, 'EUR', undefined)).toBeUndefined()
    setCollectionCurrency('EUR')
    expect(toCollectionCurrency(12.5, 'USD', 1.25)).toBe(10)
    expect(toCollectionCurrency(7.5, 'EUR', undefined)).toBe(7.5)
    expect(toCollectionCurrency(7.5, 'GBP', 1.25)).toBeUndefined()
  })
})
