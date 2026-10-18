// Flags of countries that no longer exist, which have no emoji (Unicode has
// today's countries only): small drawings of each, shown as images where a
// page can show one (group titles, the list's country). A menu's options are
// text only, and show none. The names are those the server counts as extinct
// (services/extinct_countries.go), in Portuguese and English. A country whose
// flag is a living country's (Czechoslovakia's, the Ottoman Empire's, a
// colony's) gets that emoji instead: see countryFlags.ts.
import { foldCountryName, lookupCountry } from '@/utils/worldCountries'

const RED = '#d52b1e'
const WHITE = '#ffffff'
const BLUE = '#003893'
const BLACK = '#000000'
const GOLD = '#ffce00'
const GREEN = '#007a3d'

// A five-pointed star, its top point up (or turned by rotation degrees).
function star(cx: number, cy: number, outer: number, fill: string, extra = '', rotation = -90): string {
  const inner = outer * 0.382
  const points = Array.from({ length: 10 }, (_, index) => {
    const radius = index % 2 ? inner : outer
    const angle = ((rotation + index * 36) * Math.PI) / 180
    return `${(cx + radius * Math.cos(angle)).toFixed(2)},${(cy + radius * Math.sin(angle)).toFixed(2)}`
  })
  return `<polygon points="${points.join(' ')}" fill="${fill}"${extra}/>`
}

// Equal bands, top to bottom (or hoist to fly), on a 30 × 20 flag.
function bands(colors: string[], vertical = false): string {
  const size = (vertical ? 30 : 20) / colors.length
  return colors.map((color, index) => vertical
    ? `<rect x="${(index * size).toFixed(2)}" width="${size.toFixed(2)}" height="20" fill="${color}"/>`
    : `<rect y="${(index * size).toFixed(2)}" width="30" height="${size.toFixed(2)}" fill="${color}"/>`).join('')
}

// The Union Jack in a box of the flag (Hawaii's canton).
function unionJack(x: number, y: number, width: number, height: number): string {
  return `<svg x="${x}" y="${y}" width="${width}" height="${height}" viewBox="0 0 60 30" preserveAspectRatio="none">`
    + '<rect width="60" height="30" fill="#012169"/>'
    + '<path d="M0,0 60,30 M60,0 0,30" stroke="#fff" stroke-width="6"/>'
    + '<path d="M0,0 60,30 M60,0 0,30" stroke="#c8102e" stroke-width="2"/>'
    + '<path d="M30,0 V30 M0,15 H60" stroke="#fff" stroke-width="10"/>'
    + '<path d="M30,0 V30 M0,15 H60" stroke="#c8102e" stroke-width="6"/></svg>'
}

interface HistoricFlag {
  names: string[]
  /** The drawing, on a 30 × 20 flag. */
  draw: () => string
}

const FLAGS: HistoricFlag[] = [
  {
    // Red, with the hammer and sickle and the star in gold.
    names: ['União Soviética', 'União das Repúblicas Socialistas Soviéticas', 'URSS', 'Soviet Union', 'USSR', 'Union of Soviet Socialist Republics'],
    draw: () => `<rect width="30" height="20" fill="#cc0000"/>${star(6.3, 2.6, 1.3, 'none', ` stroke="${GOLD}" stroke-width="0.35"`)}`
      + `<path d="M4.3,5.6 A2.9,2.9 0 1 0 8.6,4.5" fill="none" stroke="${GOLD}" stroke-width="0.8"/>`
      + `<path d="M4.6,9.2 L7.8,5.4" stroke="${GOLD}" stroke-width="0.8"/><path d="M6.9,4.5 L8.7,6.3" stroke="${GOLD}" stroke-width="1"/>`,
  },
  {
    // The socialist federation's: blue, white and red, a red star edged in gold.
    names: ['Jugoslávia', 'Iugoslávia', 'Yugoslavia', 'República Federal Socialista da Jugoslávia', 'Socialist Federal Republic of Yugoslavia'],
    draw: () => bands([BLUE, WHITE, RED]) + star(15, 10.4, 5.4, RED, ` stroke="${GOLD}" stroke-width="0.5"`),
  },
  {
    // The kingdom's, and Serbia and Montenegro's: blue, white and red.
    names: ['Reino da Jugoslávia', 'Kingdom of Yugoslavia', 'República Federal da Jugoslávia', 'Federal Republic of Yugoslavia', 'Sérvia e Montenegro', 'Serbia and Montenegro'],
    draw: () => bands([BLUE, WHITE, RED]),
  },
  {
    // Black, red and gold, with the gold emblem (hammer and compass in a wreath).
    names: ['Alemanha Oriental', 'República Democrática Alemã', 'RDA', 'East Germany', 'German Democratic Republic', 'GDR', 'DDR'],
    draw: () => bands([BLACK, '#dd0000', GOLD]) + `<circle cx="15" cy="10" r="3.6" fill="none" stroke="${GOLD}" stroke-width="0.9"/>`
      + `<path d="M13.4,12 L15,7.6 L16.6,12" fill="none" stroke="${GOLD}" stroke-width="0.6"/>`,
  },
  {
    // White, a red upright band and a blue band across with five white stars.
    names: ['Antilhas Holandesas', 'Antilhas Neerlandesas', 'Netherlands Antilles'],
    draw: () => `<rect width="30" height="20" fill="${WHITE}"/><rect x="13.33" width="3.34" height="20" fill="${RED}"/>`
      + `<rect y="6.67" width="30" height="6.66" fill="#012a87"/>`
      + [[8.6, 8.9], [8.6, 11.1], [21.4, 8.9], [21.4, 11.1], [11, 10]].map(([x, y]) => star(x!, y!, 0.9, WHITE)).join(''),
  },
  {
    // Yellow, the French tricolour in the canton.
    names: ['Indochina Francesa', 'French Indochina'],
    draw: () => `<rect width="30" height="20" fill="#ffcd00"/><rect width="4" height="8" fill="#002395"/><rect x="4" width="4" height="8" fill="${WHITE}"/><rect x="8" width="4" height="8" fill="#ed2939"/>`,
  },
  {
    // Green, a yellow disc with a hand holding a torch.
    names: ['Zaire', 'República do Zaire', 'Republic of Zaire'],
    draw: () => `<rect width="30" height="20" fill="#00a651"/><circle cx="15" cy="10" r="5.6" fill="${GOLD}"/>`
      + '<rect x="14.4" y="9.6" width="1.2" height="5" fill="#6b3a1e"/><path d="M15,5.2 C13.4,7.2 13.6,9.6 15,9.8 C16.4,9.6 16.6,7.2 15,5.2 Z" fill="#e2231a"/>',
  },
  {
    // Green, white and green, the arms in the middle.
    names: ['Rodésia', 'Rhodesia'],
    draw: () => bands(['#00703c', WHITE, '#00703c'], true) + '<path d="M13.4,7 H16.6 V10.6 C16.6,12.4 15,13.2 15,13.2 C15,13.2 13.4,12.4 13.4,10.6 Z" fill="#00703c" stroke="#c4a000" stroke-width="0.4"/>',
  },
  { names: ['Império Alemão', 'German Empire'], draw: () => bands([BLACK, WHITE, '#dd0000']) },
  {
    // Yellow, three red stripes.
    names: ['Vietname do Sul', 'Vietnã do Sul', 'South Vietnam', 'República do Vietname', 'Republic of Vietnam'],
    draw: () => `<rect width="30" height="20" fill="#ffcd00"/>${[6.67, 9.33, 12].map((y) => `<rect y="${y}" width="30" height="1.33" fill="#da251d"/>`).join('')}`,
  },
  {
    names: ['República Árabe Unida', 'United Arab Republic'],
    draw: () => bands(['#ce1126', WHITE, BLACK]) + star(10, 10, 2.2, GREEN) + star(20, 10, 2.2, GREEN),
  },
  { names: ['Iémen do Norte', 'Iêmen do Norte', 'North Yemen', 'República Árabe do Iémen', 'Yemen Arab Republic'], draw: () => bands(['#ce1126', WHITE, BLACK]) + star(15, 10, 2.4, GREEN) },
  {
    names: ['Iémen do Sul', 'Iêmen do Sul', 'South Yemen', 'República Democrática Popular do Iémen', "People's Democratic Republic of Yemen"],
    draw: () => bands(['#ce1126', WHITE, BLACK]) + '<path d="M0,0 L10,10 L0,20 Z" fill="#3eaee8"/>' + star(3.6, 10, 2.1, '#ce1126'),
  },
  {
    // Red, black and green, a rising sun on the black.
    names: ['Biafra', 'República do Biafra', 'Republic of Biafra'],
    draw: () => bands(['#e2231a', BLACK, '#00843d']) + `<path d="M12.2,11 A2.8,2.8 0 0 1 17.8,11 Z" fill="${GOLD}"/>`
      + Array.from({ length: 9 }, (_, index) => {
        const angle = Math.PI + (index * Math.PI) / 8
        return `<path d="M${(15 + 3.2 * Math.cos(angle)).toFixed(2)},${(11 + 3.2 * Math.sin(angle)).toFixed(2)} L${(15 + 4.3 * Math.cos(angle)).toFixed(2)},${(11 + 4.3 * Math.sin(angle)).toFixed(2)}" stroke="${GOLD}" stroke-width="0.5"/>`
      }).join(''),
  },
  {
    // The Vierkleur: red, white and blue, a green band at the hoist.
    names: ['Transvaal', 'República Sul-Africana (ZAR)', 'South African Republic', 'Zuid-Afrikaansche Republiek'],
    draw: () => bands(['#ae1c28', WHITE, '#21468b']) + '<rect width="7.5" height="20" fill="#007749"/>',
  },
  {
    // A white cross, blue above and red below.
    names: ['Sarre', 'Saar', 'Saarland', 'Protetorado do Sarre', 'Saar Protectorate'],
    draw: () => `<rect width="30" height="10" fill="#002f6c"/><rect y="10" width="30" height="10" fill="#e30613"/><rect x="9" width="3" height="20" fill="${WHITE}"/><rect y="8.5" width="30" height="3" fill="${WHITE}"/>`,
  },
  { names: ['Império Austríaco', 'Austrian Empire'], draw: () => bands([BLACK, '#fcdd09']) },
  { names: ['Prússia', 'Prussia', 'Reino da Prússia', 'Kingdom of Prussia'], draw: () => bands([BLACK, WHITE]) },
  { names: ['Saxónia', 'Saxônia', 'Saxony', 'Reino da Saxónia', 'Kingdom of Saxony'], draw: () => bands([WHITE, '#009933']) },
  { names: ['Württemberg', 'Reino de Württemberg', 'Kingdom of Württemberg'], draw: () => bands([BLACK, '#dd0000']) },
  { names: ['Hanôver', 'Hanover', 'Reino de Hanôver', 'Kingdom of Hanover'], draw: () => bands(['#ffd700', WHITE]) },
  { names: ['Estados Papais', 'Estados Pontifícios', 'Papal States'], draw: () => bands(['#ffe000', WHITE], true) },
  {
    // Yellow, a canton of red, blue, white and black.
    names: ['Manchukuo', 'Manchúria', 'Manchukuo (Manchuria)'],
    draw: () => `<rect width="30" height="20" fill="#ffd700"/>${['#e60012', '#0033a0', WHITE, BLACK].map((color, index) => `<rect y="${index * 2.5}" width="10" height="2.5" fill="${color}"/>`).join('')}`,
  },
  {
    // Eight white, red and blue stripes, the Union Jack in the canton.
    names: ['Reino do Havai', 'Reino do Havaí', 'Kingdom of Hawaii', 'Havai', 'Havaí', 'Hawaii'],
    draw: () => [WHITE, '#c8102e', '#012169', WHITE, '#c8102e', '#012169', WHITE, '#c8102e'].map((color, index) => `<rect y="${index * 2.5}" width="30" height="2.5" fill="${color}"/>`).join('') + unionJack(0, 0, 13, 10),
  },
]

// Each folded name, and the flag's place in FLAGS (as text, as lookupCountry keeps codes).
let index: Map<string, string> | null = null

function historicIndex(): Map<string, string> {
  if (index) return index
  index = new Map()
  FLAGS.forEach((flag, position) => {
    for (const name of flag.names) index!.set(foldCountryName(name), String(position))
  })
  return index
}

const images = new Map<string, string>()

/**
 * The flag of a country that no longer exists, as an image address (an SVG
 * drawing), or '' when it is not one the app draws. Numista's forms count:
 * "Jugoslávia, República Federal Socialista da", "União Soviética (URSS)".
 */
export function historicFlagImage(country: string | null | undefined): string {
  const name = country?.trim()
  if (!name) return ''
  const position = lookupCountry(historicIndex(), name)
  if (position === undefined) return ''
  let image = images.get(position)
  if (!image) {
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 30 20">${FLAGS[Number(position)]!.draw()}</svg>`
    image = `data:image/svg+xml,${encodeURIComponent(svg)}`
    images.set(position, image)
  }
  return image
}

/** The names of every country drawn, for tests and documentation. */
export function historicFlagNames(): string[] {
  return FLAGS.flatMap((flag) => flag.names)
}
