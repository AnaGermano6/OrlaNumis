import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import { collectionCurrencySymbol, currencyName, setCollectionCurrency, toCollectionCurrency } from '@/utils/collectionCurrency'
import { formatCurrency } from '@/utils/format'

describe('collection currency', () => {
  afterEach(() => {
    setCollectionCurrency('USD')
    setLocale('en')
  })

  it('formats amounts in the collection currency unless told otherwise', () => {
    setLocale('pt-PT')
    expect(formatCurrency(3.37).replace(/\s/g, ' ')).toBe('3,37 US$')
    setCollectionCurrency('eur')
    expect(formatCurrency(3.01).replace(/\s/g, ' ')).toBe('3,01 €')
    expect(formatCurrency(3.37, 'USD').replace(/\s/g, ' ')).toBe('3,37 US$')
    expect(collectionCurrencySymbol()).toBe('€')
    setCollectionCurrency('GBP')
    expect(collectionCurrencySymbol()).toBe('£')
    expect(formatCurrency(3).replace(/\s/g, ' ')).toBe('3,00 £')
    // Unknown codes leave the currency as it was.
    setCollectionCurrency('XYZ')
    expect(collectionCurrencySymbol()).toBe('£')
  })

  it('names currencies in the app language', () => {
    setLocale('pt-PT')
    expect(currencyName('EUR')).toBe('Euro (€)')
    expect(currencyName('GBP')).toBe('Libra esterlina britânica (£)')
    setLocale('en')
    expect(currencyName('CHF')).toBe('Swiss Franc (CHF)')
  })

  it('converts between any two currencies of the ECB rates, through the euro', () => {
    const perEur = { USD: 1.25, GBP: 0.8 }
    expect(toCollectionCurrency(10, 'USD', undefined)).toBe(10)
    expect(toCollectionCurrency(10, 'EUR', perEur)).toBe(12.5)
    expect(toCollectionCurrency(10, 'EUR', undefined)).toBeUndefined()
    setCollectionCurrency('EUR')
    expect(toCollectionCurrency(12.5, 'USD', perEur)).toBe(10)
    expect(toCollectionCurrency(7.5, 'EUR', undefined)).toBe(7.5)
    setCollectionCurrency('GBP')
    expect(toCollectionCurrency(12.5, 'USD', perEur)).toBe(8)
    expect(toCollectionCurrency(10, 'CHF', perEur)).toBeUndefined()
  })
})
