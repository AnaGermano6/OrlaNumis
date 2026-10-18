// The flag of a stored country name, as an emoji made from its ISO code
// ("Croácia" → 🇭🇷). Country names come in Portuguese or English, as Numista
// or a person writes them. A country that no longer exists (USSR,
// Yugoslavia, Ottoman Empire…) or a name that is not recognised has no flag.
import { EXTRA_COUNTRY_ALIASES, foldCountryName, lookupCountry } from '@/utils/worldCountries'

// Codes the browser names that have no flag: former countries, and private
// or reserved codes.
const NO_FLAG = new Set(['AN', 'BU', 'CS', 'DD', 'DY', 'FX', 'HV', 'NH', 'NT', 'QO', 'QU', 'RH', 'SU', 'TP', 'VD', 'YD', 'YU', 'ZR', 'ZZ'])

// Names collections use that the browser does not, with the flag they fly.
const FLAG_ALIASES: Record<string, string[]> = {
  ...EXTRA_COUNTRY_ALIASES,
  PT: ['Região Autónoma dos Açores', 'Açores', 'Azores', 'Região Autónoma da Madeira', 'Madeira'],
  MK: ['Macedónia', 'Macedônia', 'Macedonia'],
  VA: ['Vaticano', 'Vatican', 'Santa Sé', 'Holy See'],
  MU: ['Maurícias', 'Mauritius'],
  EU: ['Área do euro', 'Euro area', 'Eurozone', 'Zona euro'],
}

let index: Map<string, string> | null = null

function countryIndex(): Map<string, string> {
  if (index) return index
  index = new Map()
  const add = (name: string, code: string) => {
    const folded = foldCountryName(name)
    if (folded && !index!.has(folded)) index!.set(folded, code)
  }
  for (const [code, aliases] of Object.entries(FLAG_ALIASES)) {
    for (const alias of aliases) add(alias, code)
  }
  const displays = ['pt-PT', 'pt', 'en'].flatMap((language) => {
    try {
      return [new Intl.DisplayNames([language], { type: 'region' })]
    } catch {
      return []
    }
  })
  for (let first = 65; first <= 90; first += 1) {
    for (let second = 65; second <= 90; second += 1) {
      const code = String.fromCharCode(first, second)
      if (NO_FLAG.has(code) || code.startsWith('X') && code !== 'XK') continue
      for (const display of displays) {
        let name: string | undefined
        try {
          name = display.of(code)
        } catch {
          name = undefined
        }
        // "Hong Kong, RAE da China" is also just "Hong Kong".
        if (name && name !== code) for (const variant of [name, name.split(',')[0] ?? '']) add(variant, code)
      }
    }
  }
  return index
}

/** The ISO code a stored country name has a flag for, or undefined. */
export function countryFlagCode(country: string | null | undefined): string | undefined {
  const name = country?.trim()
  if (!name) return undefined
  const code = lookupCountry(countryIndex(), name)
  return code && /^[A-Z]{2}$/.test(code) && !NO_FLAG.has(code) ? code : undefined
}

/** The flag emoji of an ISO country code ("HR" → 🇭🇷), or '' when it has none. */
export function flagOfCode(code: string | null | undefined): string {
  const upper = code?.trim().toUpperCase() ?? ''
  if (!/^[A-Z]{2}$/.test(upper) || NO_FLAG.has(upper) || upper.startsWith('X') && upper !== 'XK') return ''
  return String.fromCodePoint(...[...upper].map((letter) => 0x1f1e6 + letter.charCodeAt(0) - 65))
}

/** The flag emoji of a stored country name, or '' when it has none. */
export function countryFlag(country: string | null | undefined): string {
  return flagOfCode(countryFlagCode(country))
}

/** "🇭🇷 Croácia", or the name alone when it has no flag. */
export function withCountryFlag(country: string): string {
  const flag = countryFlag(country)
  return flag ? `${flag} ${country}` : country
}

/** A name with the flag of an ISO code before it, for text alone (a list's options). */
export function withFlagOfCode(code: string | null | undefined, name: string): string {
  const flag = flagOfCode(code)
  return flag ? `${flag} ${name}` : name
}
