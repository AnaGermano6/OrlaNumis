import { describe, expect, it } from 'vitest'
import { normalizeMintName } from '@/utils/mintMap'

describe('mintMap utilities', () => {
  it('normalizes case, punctuation, diacritics, and whitespace', () => {
    expect(normalizeMintName('  RÓMA--Mint  ')).toBe('roma mint')
    expect(normalizeMintName('Lugdunum / Lyon')).toBe('lugdunum lyon')
  })
})
