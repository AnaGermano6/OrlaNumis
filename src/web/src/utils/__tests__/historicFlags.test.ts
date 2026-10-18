import { describe, expect, it } from 'vitest'
import { historicFlagImage, historicFlagNames } from '@/utils/historicFlags'

function svgOf(image: string): Document {
  expect(image.startsWith('data:image/svg+xml,')).toBe(true)
  return new DOMParser().parseFromString(decodeURIComponent(image.slice('data:image/svg+xml,'.length)), 'image/svg+xml')
}

describe('flags of countries that no longer exist', () => {
  it('draws one for every name, as a well-formed picture', () => {
    expect(historicFlagNames().length).toBeGreaterThan(60)
    for (const name of historicFlagNames()) {
      const doc = svgOf(historicFlagImage(name))
      expect(doc.getElementsByTagName('parsererror')).toHaveLength(0)
      expect(doc.documentElement.getAttribute('viewBox')).toBe('0 0 30 20')
    }
  })

  it("reads Numista's forms, and the same country in either language", () => {
    const yugoslavia = historicFlagImage('Jugoslávia')
    expect(historicFlagImage('Jugoslávia, República Federal Socialista da')).toBe(yugoslavia)
    expect(historicFlagImage('Yugoslavia')).toBe(yugoslavia)
    expect(historicFlagImage('União Soviética (URSS)')).toBe(historicFlagImage('USSR'))
    // The kingdom flew the tricolour without the star.
    expect(historicFlagImage('Reino da Jugoslávia')).not.toBe(yugoslavia)
    expect(svgOf(yugoslavia).getElementsByTagName('polygon')).toHaveLength(1)
  })

  it('draws none for a living country or a name it does not know', () => {
    for (const name of ['Portugal', 'Checoslováquia', 'Atlântida', '', null, undefined]) expect(historicFlagImage(name)).toBe('')
  })
})
