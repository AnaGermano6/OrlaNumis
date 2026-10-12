import { describe, expect, it } from 'vitest'
import { MODERN_MINTS, findMintPlace } from '@/utils/modernMints'

const find = (mint: string, iso?: string) => findMintPlace(mint, iso, MODERN_MINTS)?.key

describe('findMintPlace', () => {
  it('finds modern mints written as the collection writes them', () => {
    expect(find('Imprensa Nacional - Casa da Moeda')).toBe('builtin:incm')
    expect(find('Imprensa Nacional - Casa da Moeda (INCM)')).toBe('builtin:incm')
    expect(find('Imprensa Nacional - Casa da Moeda (incm)')).toBe('builtin:incm')
    expect(find('Royal Mint')).toBe('builtin:royal-mint')
    expect(find('Royal Mint (Tower Hill)')).toBe('builtin:tower-hill')
    expect(find('Casa da Moeda de Paris')).toBe('builtin:paris')
    expect(find('Royal Mint of Madrid (M)')).toBe('builtin:madrid')
    expect(find('Roma (R)')).toBe('builtin:rome')
    expect(find('Mexican Mint (Mo)')).toBe('builtin:mexico')
    expect(find('Porto (PPPP)')).toBe('builtin:porto')
  })

  it('reads a mint mark by the coin\'s country', () => {
    expect(find('D', 'DE')).toBe('builtin:munich')
    expect(find('D', 'US')).toBe('builtin:denver')
    expect(find('A', 'DE')).toBe('builtin:berlin')
    expect(find('J', 'DE')).toBe('builtin:hamburg')
    // Alone, a letter names no mint.
    expect(find('D')).toBeUndefined()
    expect(find('State Mint')).toBeUndefined()
    expect(find('')).toBeUndefined()
  })
})
