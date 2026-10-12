// Mints of modern coins, for the collection map: coins name their mint in
// free text ("Imprensa Nacional - Casa da Moeda (INCM)", "Royal Mint of Madrid
// (M)") or by a mint mark that only means something within its country ("D"
// is Munich on a German coin, Denver on an American one).
import { normalizeMintName } from '@/utils/mintMap'

export interface MintPlace {
  /** "builtin:<id>" for these, "location:<id>" for the app's mint locations. */
  key: string
  name: string
  lat: number
  lng: number
  aliases: string[]
}

function mint(id: string, name: string, lat: number, lng: number, ...aliases: string[]): MintPlace {
  return { key: `builtin:${id}`, name, lat, lng, aliases }
}

export const MODERN_MINTS: MintPlace[] = [
  mint('incm', 'Imprensa Nacional - Casa da Moeda (Lisboa)', 38.716, -9.163, 'Imprensa Nacional - Casa da Moeda', 'INCM', 'Casa da Moeda de Lisboa', 'Lisboa', 'Lisbon'),
  mint('porto', 'Casa da Moeda do Porto', 41.15, -8.61, 'Porto'),
  mint('royal-mint', 'Royal Mint (Llantrisant)', 51.54, -3.38, 'Royal Mint', 'Llantrisant', 'The Royal Mint'),
  mint('tower-hill', 'Royal Mint (Tower Hill, Londres)', 51.51, -0.074, 'Royal Mint (Tower Hill)', 'Tower Hill', 'London', 'Londres'),
  mint('paris', 'Monnaie de Paris', 48.857, 2.339, 'Casa da Moeda de Paris', 'Paris', 'Pessac', 'Paris Mint'),
  mint('utrecht', 'Koninklijke Nederlandse Munt (Utrecht)', 52.09, 5.12, 'Royal Dutch Mint', 'Koninklijke Nederlandse Munt', 'KNM', 'Utrecht', 'Houten'),
  mint('madrid', 'Real Casa de la Moneda (Madrid)', 40.43, -3.67, 'Royal Mint of Madrid', 'Real Casa de la Moneda', 'FNMT', 'Fábrica Nacional de Moneda y Timbre', 'Madrid'),
  mint('brussels', 'Monnaie Royale de Belgique (Bruxelas)', 50.85, 4.35, 'Casa da Moeda Real da Bélgica', 'Royal Belgian Mint', 'Koninklijke Munt van België', 'Brussels', 'Bruxelas'),
  mint('rome', 'Zecca di Roma (IPZS)', 41.9, 12.5, 'Roma', 'Rome', 'IPZS', 'Istituto Poligrafico e Zecca dello Stato', 'Zecca di Roma'),
  mint('vienna', 'Münze Österreich (Viena)', 48.2, 16.38, 'Austrian Mint', 'Münze Österreich', 'Vienna', 'Viena', 'Wien'),
  mint('kremnica', 'Mincovňa Kremnica', 48.7, 18.92, 'Kremnica', 'Kremnica Mint'),
  mint('croatia', 'Hrvatska kovnica novca (Sveta Nedelja)', 45.8, 15.78, 'Croatian Mint', 'Hrvatska kovnica novca'),
  mint('vilnius', 'Lietuvos monetų kalykla (Vilnius)', 54.69, 25.28, 'Lithuanian Mint', 'LMK', 'Vilnius'),
  mint('warsaw', 'Mennica Polska (Varsóvia)', 52.23, 21.0, 'Casa da Moeda da Polônia', 'Casa da Moeda da Polónia', 'Mennica Polska', 'Polish Mint', 'Warsaw', 'Varsóvia', 'MW'),
  mint('finland', 'Suomen Rahapaja (Vantaa)', 60.29, 25.04, 'Casa da Moeda da Finlândia', 'Mint of Finland', 'Suomen Rahapaja', 'Vantaa'),
  mint('dublin', 'Irish Mint (Dublin)', 53.27, -6.22, 'Irish Mint', 'Central Bank of Ireland', 'Dublin'),
  mint('berlin', 'Staatliche Münze Berlin', 52.52, 13.4, 'Berlin', 'Berlim'),
  mint('munich', 'Bayerisches Hauptmünzamt (Munique)', 48.14, 11.58, 'Munich', 'Munique', 'München'),
  mint('stuttgart', 'Staatliche Münzen (Estugarda)', 48.78, 9.18, 'Stuttgart', 'Estugarda'),
  mint('karlsruhe', 'Staatliche Münzen (Karlsruhe)', 49.01, 8.4, 'Karlsruhe'),
  mint('hamburg', 'Hamburgische Münze (Hamburgo)', 53.55, 10.0, 'Hamburg', 'Hamburgo'),
  mint('bern', 'Swissmint (Berna)', 46.95, 7.45, 'Swissmint', 'Bern', 'Berna'),
  mint('philadelphia', 'United States Mint (Filadélfia)', 39.95, -75.15, 'Philadelphia', 'Filadélfia'),
  mint('denver', 'United States Mint (Denver)', 39.74, -104.99, 'Denver'),
  mint('san-francisco', 'United States Mint (São Francisco)', 37.77, -122.42, 'San Francisco', 'São Francisco'),
  mint('west-point', 'United States Mint (West Point)', 41.39, -73.96, 'West Point'),
  mint('winnipeg', 'Royal Canadian Mint (Winnipeg)', 49.84, -97.06, 'Royal Canadian Mint of Winnipeg', 'Winnipeg'),
  mint('ottawa', 'Royal Canadian Mint (Ottawa)', 45.43, -75.7, 'Royal Canadian Mint', 'Ottawa'),
  mint('mexico', 'Casa de Moneda de México', 22.15, -100.98, 'Mexican Mint', 'Casa de Moneda de México', 'Mo'),
  mint('brazil', 'Casa da Moeda do Brasil (Rio de Janeiro)', -22.9, -43.2, 'Casa da Moeda do Brasil'),
  mint('lima', 'Casa Nacional de Moneda (Lima)', -12.05, -77.03, 'Lima', 'Casa Nacional de Moneda'),
  mint('ibague', 'Fábrica de Moneda de Ibagué', 4.44, -75.23, 'Fábrica de Moneda de Ibagué', 'Ibagué'),
  mint('cairo', 'Egyptian Mint Authority (Cairo)', 30.04, 31.24, 'Egyptian Mint Authority', 'Egyptian Mint', 'Cairo'),
  mint('pretoria', 'South African Mint (Pretória)', -25.86, 28.19, 'South African Mint', 'Pretória', 'Pretoria'),
  mint('canberra', 'Royal Australian Mint (Camberra)', -35.31, 149.1, 'Royal Australian Mint', 'Canberra', 'Camberra'),
  mint('osaka', 'Japan Mint (Osaca)', 34.7, 135.52, 'Japan Mint', 'Osaka', 'Osaca'),
  mint('peruri', 'Perum Peruri (Indonésia)', -6.3, 107.3, 'Perum Peruri', 'Peruri'),
  mint('shah-alam', 'Royal Mint of Malaysia (Shah Alam)', 3.07, 101.52, 'Shah Alam', 'Royal Mint of Malaysia'),
  mint('lahore', 'Pakistan Mint (Lahore)', 31.55, 74.34, 'Lahore', 'Pakistan Mint'),
]

/** Mint marks that name a mint only on a given country's coins (ISO code → mark → mint id). */
const MINT_MARKS: Record<string, Record<string, string>> = {
  DE: { a: 'berlin', d: 'munich', f: 'stuttgart', g: 'karlsruhe', j: 'hamburg' },
  US: { p: 'philadelphia', d: 'denver', s: 'san-francisco', w: 'west-point' },
  CH: { b: 'bern' },
  ES: { m: 'madrid' },
  MX: { mo: 'mexico' },
  IT: { r: 'rome' },
}

/** Finds a coin's mint among the places, from its free-text mint and its country (ISO code). */
export function findMintPlace(mintText: string, countryIso: string | undefined, places: MintPlace[]): MintPlace | undefined {
  const text = mintText.trim()
  if (!text) return undefined
  const byName = new Map<string, MintPlace>()
  for (const place of places) {
    for (const name of [place.name, ...place.aliases]) {
      const normalized = normalizeMintName(name)
      if (normalized && !byName.has(normalized)) byName.set(normalized, place)
    }
  }
  const builtin = (id: string | undefined) => (id ? places.find((place) => place.key === `builtin:${id}`) : undefined)
  const marks = countryIso ? MINT_MARKS[countryIso] : undefined
  // "Royal Mint of Madrid (M)": the name, then the name without the mark, then the mark.
  const inBrackets = [...text.matchAll(/\(([^)]*)\)/g)].map((match) => match[1] ?? '')
  const candidates = [text, text.replace(/\([^)]*\)/g, ' '), ...inBrackets]
  for (const candidate of candidates) {
    const normalized = normalizeMintName(candidate)
    if (!normalized) continue
    const found = byName.get(normalized) ?? builtin(marks?.[normalized])
    if (found) return found
  }
  return undefined
}
