import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import { healthItemLabel } from '@/utils/healthItems'

describe('healthItemLabel', () => {
  afterEach(() => setLocale('en'))

  it('names what an item asks for in the language of the app', () => {
    setLocale('pt-PT')
    expect(healthItemLabel('metadata.monetarySystem')).toBe('Sistema monetário')
    expect(healthItemLabel('valuation.value', 'Current valuation')).toBe('Valor atual')
    setLocale('en')
    expect(healthItemLabel('images.coverage')).toBe('Photos of both sides')
  })

  it("keeps the server's label, or the key, for an item it does not know", () => {
    expect(healthItemLabel('metadata.weight', 'Weight')).toBe('Weight')
    expect(healthItemLabel('metadata.weight')).toBe('metadata.weight')
  })
})
