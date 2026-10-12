// Geography for the collection map that the country outlines do not carry:
// small states and islands too small for the 1:110m outlines (drawn as
// points), the continent of each place, the euro's issuers, and the present
// territory of countries that no longer exist.
import { foldCountryName } from '@/utils/worldCountries'

export type Continent = 'europe' | 'asia' | 'africa' | 'northAmerica' | 'southAmerica' | 'oceania' | 'antarctica'

export const CONTINENTS: Continent[] = ['europe', 'asia', 'africa', 'northAmerica', 'southAmerica', 'oceania', 'antarctica']

/** A place drawn as a point: ISO code (or a region code such as PT-20), names and position. */
export interface MapPoint {
  iso: string
  name: string
  namePt: string
  aliases: string[]
  lat: number
  lng: number
}

function point(iso: string, name: string, namePt: string, lat: number, lng: number, ...aliases: string[]): MapPoint {
  return { iso, name, namePt, lat, lng, aliases: [name, namePt, ...aliases] }
}

export const SMALL_TERRITORIES: MapPoint[] = [
  point('AD', 'Andorra', 'Andorra', 42.55, 1.6),
  point('MC', 'Monaco', 'Mónaco', 43.74, 7.42, 'Mônaco'),
  point('SM', 'San Marino', 'São Marinho', 43.94, 12.46),
  point('VA', 'Vatican City', 'Vaticano', 41.9, 12.45, 'Cidade do Vaticano', 'Estado da Cidade do Vaticano', 'Santa Sé', 'Holy See', 'Vatican'),
  point('MT', 'Malta', 'Malta', 35.9, 14.45),
  point('LI', 'Liechtenstein', 'Listenstaine', 47.16, 9.55, 'Liechtenstein'),
  point('IM', 'Isle of Man', 'Ilha de Man', 54.24, -4.55),
  point('JE', 'Jersey', 'Jersey', 49.21, -2.13),
  point('GG', 'Guernsey', 'Guernsey', 49.45, -2.58, 'Guernesey'),
  point('GI', 'Gibraltar', 'Gibraltar', 36.14, -5.35),
  point('FO', 'Faroe Islands', 'Ilhas Faroé', 62.0, -6.9, 'Ilhas Feroé'),
  point('PT-20', 'Azores', 'Açores', 37.74, -25.67, 'Região Autónoma dos Açores', 'Arquipélago dos Açores'),
  point('PT-30', 'Madeira', 'Madeira', 32.65, -16.91, 'Região Autónoma da Madeira', 'Madeira, Arquipélago da'),
  point('SG', 'Singapore', 'Singapura', 1.35, 103.82),
  point('BH', 'Bahrain', 'Barém', 26.07, 50.55, 'Bahrein', 'Barein'),
  point('HK', 'Hong Kong', 'Hong Kong', 22.32, 114.17),
  point('MO', 'Macau', 'Macau', 22.2, 113.55, 'Macao'),
  point('MV', 'Maldives', 'Maldivas', 3.2, 73.22),
  point('CV', 'Cape Verde', 'Cabo Verde', 15.1, -23.6, 'Cabo Verde'),
  point('SC', 'Seychelles', 'Seicheles', -4.68, 55.49, 'Seychelles'),
  point('MU', 'Mauritius', 'Maurícias', -20.3, 57.58, 'Maurícia', 'Ilhas Maurício'),
  point('KM', 'Comoros', 'Comores', -11.88, 43.87),
  point('ST', 'São Tomé and Príncipe', 'São Tomé e Príncipe', 0.19, 6.61),
  point('SH', 'Saint Helena', 'Santa Helena', -15.97, -5.71),
  point('BB', 'Barbados', 'Barbados', 13.19, -59.54),
  point('AG', 'Antigua and Barbuda', 'Antígua e Barbuda', 17.06, -61.8),
  point('DM', 'Dominica', 'Dominica', 15.41, -61.37),
  point('GD', 'Grenada', 'Granada', 12.12, -61.68),
  point('KN', 'Saint Kitts and Nevis', 'São Cristóvão e Neves', 17.36, -62.78),
  point('LC', 'Saint Lucia', 'Santa Lúcia', 13.91, -60.98),
  point('VC', 'Saint Vincent and the Grenadines', 'São Vicente e Granadinas', 13.25, -61.2),
  point('AW', 'Aruba', 'Aruba', 12.52, -69.97),
  point('CW', 'Curaçao', 'Curaçau', 12.17, -68.98, 'Curacao'),
  point('SX', 'Sint Maarten', 'São Martinho', 18.04, -63.05),
  point('BM', 'Bermuda', 'Bermudas', 32.32, -64.76),
  point('KY', 'Cayman Islands', 'Ilhas Caimão', 19.31, -81.25, 'Ilhas Cayman'),
  point('TO', 'Tonga', 'Tonga', -21.18, -175.2),
  point('CK', 'Cook Islands', 'Ilhas Cook', -21.24, -159.78),
  point('WS', 'Samoa', 'Samoa', -13.76, -172.1),
  point('KI', 'Kiribati', 'Quiribáti', 1.45, 173.0, 'Kiribati'),
  point('TV', 'Tuvalu', 'Tuvalu', -8.52, 179.2),
  point('NR', 'Nauru', 'Nauru', -0.52, 166.93),
  point('PW', 'Palau', 'Palau', 7.51, 134.58),
  point('FM', 'Micronesia', 'Micronésia', 6.92, 158.16),
  point('MH', 'Marshall Islands', 'Ilhas Marshall', 7.13, 171.18),
]

const CONTINENT_CODES: Record<Continent, string> = {
  europe: 'AD AL AT BA BE BG BY CH CY CYN CZ DE DK EE ES FI FO FR GB GG GI GR HR HU IE IM IS IT JE LI LT LU LV MC MD ME MK MT NL NO PL PT PT-20 PT-30 RO RS RU SE SI SK SM UA VA XK',
  asia: 'AE AF AM AZ BD BH BN BT CN GE HK ID IL IN IQ IR JO JP KG KH KP KR KW KZ LA LB LK MM MN MO MV MY NP OM PH PK PS QA SA SG SY TH TJ TL TM TR TW UZ VN YE',
  africa: 'AO BF BI BJ BW CD CF CG CI CM CV DJ DZ EG EH ER ET GA GH GM GN GQ GW KE KM LR LS LY MA MG ML MR MU MW MZ NA NE NG RW SC SD SH SL SN SO SOL SS ST SZ TD TG TN TZ UG ZA ZM ZW',
  northAmerica: 'AG AW BB BM BS BZ CA CR CU CW DM DO GD GL GT HN HT JM KN KY LC MX NI PA PR SV SX TT US VC',
  southAmerica: 'AR BO BR CL CO EC FK GY PE PY SR UY VE',
  oceania: 'AU CK FJ FM KI MH NC NR NZ PG PW SB TO TV VU WS',
  antarctica: 'AQ TF',
}

const continentByIso = new Map<string, Continent>(
  CONTINENTS.flatMap((continent) => CONTINENT_CODES[continent].split(' ').map((iso) => [iso, continent] as [string, Continent])),
)

/** The continent of a map place, by ISO (or region) code. */
export function continentOf(iso: string): Continent | undefined {
  return continentByIso.get(iso)
}

/** Every place on a continent, by code. */
export function continentPlaces(continent: Continent): string[] {
  return CONTINENT_CODES[continent].split(' ')
}

/** The countries that issue euro coins: the euro area (21 in 2026) and the four microstates with an agreement. */
export const EURO_ISSUERS = 'AT BE BG HR CY EE FI FR DE GR IE IT LV LT LU MT NL PT SK SI ES AD MC SM VA'.split(' ')

// Present-day territory of countries that no longer exist (approximate: the
// countries that came out of them, or that hold their land today).
const EXTINCT_TERRITORY_LIST: [string[], string][] = [
  [['União Soviética', 'União das Repúblicas Socialistas Soviéticas', 'URSS', 'Soviet Union', 'USSR', 'Union of Soviet Socialist Republics'], 'RU UA BY MD EE LV LT GE AM AZ KZ UZ TM KG TJ'],
  [['Império Russo', 'Russian Empire'], 'RU UA BY MD EE LV LT FI GE AM AZ KZ UZ TM KG TJ'],
  [['Jugoslávia', 'Iugoslávia', 'Reino da Jugoslávia', 'Yugoslavia', 'Kingdom of Yugoslavia'], 'RS HR SI BA MK ME XK'],
  [['Sérvia e Montenegro', 'Serbia and Montenegro'], 'RS ME XK'],
  [['Checoslováquia', 'Tchecoslováquia', 'Czechoslovakia'], 'CZ SK'],
  [['Alemanha Oriental', 'República Democrática Alemã', 'RDA', 'East Germany', 'German Democratic Republic', 'GDR', 'DDR'], 'DE'],
  [['Antilhas Holandesas', 'Antilhas Neerlandesas', 'Netherlands Antilles'], 'CW SX'],
  [['República Árabe Unida', 'United Arab Republic'], 'EG SY'],
  [['Iémen do Sul', 'Iêmen do Sul', 'South Yemen', 'Iémen do Norte', 'Iêmen do Norte', 'North Yemen'], 'YE'],
  [['Vietname do Sul', 'Vietnã do Sul', 'South Vietnam'], 'VN'],
  [['Zaire', 'Katanga'], 'CD'],
  [['Rodésia', 'Rhodesia', 'Rodésia do Sul', 'Southern Rhodesia'], 'ZW'],
  [['Rodésia e Niassalândia', 'Federação da Rodésia e Niassalândia', 'Rhodesia and Nyasaland'], 'ZW ZM MW'],
  [['Biafra'], 'NG'],
  [['Manchukuo'], 'CN'],
  [['Tannu Tuva'], 'RU'],
  [['Império Otomano', 'Ottoman Empire'], 'TR SY IQ LB JO IL PS'],
  [['Áustria-Hungria', 'Império Austro-Húngaro', 'Austria-Hungary', 'Austro-Hungarian Empire'], 'AT HU CZ SK SI HR BA'],
  [['Império Austríaco', 'Austrian Empire'], 'AT HU CZ SK SI HR'],
  [['Império Alemão', 'German Empire', 'Prússia', 'Prussia', 'Baviera', 'Bavaria', 'Saxónia', 'Saxônia', 'Saxony', 'Württemberg', 'Hanôver', 'Hanover', 'Sarre', 'Saar', 'Saarland', 'Protetorado do Sarre'], 'DE'],
  [['Estados Papais', 'Estados Pontifícios', 'Papal States', 'Reino das Duas Sicílias', 'Duas Sicílias', 'Two Sicilies', 'Kingdom of the Two Sicilies', 'Reino da Sardenha', 'Kingdom of Sardinia'], 'IT'],
  [['Reino do Havai', 'Reino do Havaí', 'Kingdom of Hawaii', 'Havai', 'Havaí', 'Hawaii'], 'US'],
  [['Terra Nova', 'Newfoundland'], 'CA'],
  [['Transvaal', 'República Sul-Africana (ZAR)', 'Estado Livre de Orange', 'Orange Free State'], 'ZA'],
  [['Cidade Livre de Danzig', 'Danzig', 'Free City of Danzig'], 'PL'],
  [['Indochina Francesa', 'French Indochina'], 'VN LA KH'],
  [['África Ocidental Francesa', 'French West Africa'], 'SN ML MR GN CI BF NE BJ'],
  [['África Equatorial Francesa', 'French Equatorial Africa'], 'GA CG CF TD'],
  [['África Oriental Britânica', 'British East Africa'], 'KE UG TZ'],
  [['África Ocidental Britânica', 'British West Africa'], 'NG GH SL GM'],
  [['Índias Orientais Holandesas', 'Índias Orientais Neerlandesas', 'Netherlands East Indies', 'Dutch East Indies'], 'ID'],
  [['Índia Britânica', 'British India'], 'IN PK BD MM'],
  [['Índia Portuguesa', 'Portuguese India'], 'IN'],
  [['Estabelecimentos do Estreito', 'Straits Settlements'], 'MY SG'],
]

const extinctTerritories = new Map<string, string[]>(
  EXTINCT_TERRITORY_LIST.flatMap(([names, codes]) => names.map((name) => [foldCountryName(name), codes.split(' ')] as [string, string[]])),
)

/**
 * Present-day places of a country that no longer exists, or [] for any other
 * country. Numista's forms count too: "Jugoslávia, República Federal
 * Socialista da" by the name before the comma, "União Soviética (URSS)" with
 * or without the brackets.
 */
export function extinctTerritory(country: string): string[] {
  const trimmed = country.trim()
  const candidates = [trimmed, trimmed.replace(/\([^)]*\)/g, ' '), trimmed.split(',')[0] ?? '', ...[...trimmed.matchAll(/\(([^)]*)\)/g)].map((match) => match[1] ?? '')]
  for (const candidate of candidates) {
    const territory = extinctTerritories.get(foldCountryName(candidate))
    if (territory) return territory
  }
  return []
}
