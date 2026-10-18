import { describe, expect, it } from 'vitest'
import { countryFlag, countryFlagCode, flagOfCode, withCountryFlag, withFlagOfCode } from '@/utils/countryFlags'

describe('country flags', () => {
  it('finds the flag of a country written in Portuguese or English, with or without accents', () => {
    expect(countryFlag('Croácia')).toBe('🇭🇷')
    expect(countryFlag('Croatia')).toBe('🇭🇷')
    expect(countryFlag('  africa do sul ')).toBe('🇿🇦')
    expect(countryFlagCode('Estados Unidos da América')).toBe('US')
    expect(countryFlagCode('Hong Kong')).toBe('HK')
  })

  it('knows the names collections use that the browser does not', () => {
    expect(countryFlagCode('Região Autónoma dos Açores')).toBe('PT')
    expect(countryFlagCode('Vaticano')).toBe('VA')
    expect(countryFlagCode('Macedónia')).toBe('MK')
    expect(countryFlagCode('Kosovo')).toBe('XK')
  })

  it('gives no flag to a country that no longer exists or a name it does not know', () => {
    for (const name of ['União Soviética', 'Jugoslávia', 'Checoslováquia', 'Império Otomano', 'Atlântida', '', null, undefined]) {
      expect(countryFlag(name)).toBe('')
    }
  })

  it('puts the flag before the name, or leaves the name alone', () => {
    expect(withCountryFlag('Portugal')).toBe('🇵🇹 Portugal')
    expect(withCountryFlag('Jugoslávia')).toBe('Jugoslávia')
  })
})

describe('flags of ISO codes', () => {
  it('turns a code into its flag, and gives none to a former or private code', () => {
    expect(flagOfCode('PT')).toBe('🇵🇹')
    expect(flagOfCode('fr')).toBe('🇫🇷')
    expect(flagOfCode('EU')).toBe('🇪🇺')
    expect(flagOfCode('XK')).toBe('🇽🇰')
    for (const code of ['SU', 'YU', 'XA', 'PTE', '', null, undefined]) expect(flagOfCode(code)).toBe('')
  })

  it('puts the flag of a code before a name', () => {
    expect(withFlagOfCode('PT', 'Portugal · Escudos')).toBe('🇵🇹 Portugal · Escudos')
    expect(withFlagOfCode('SU', 'União Soviética')).toBe('União Soviética')
  })
})
