import type { Coin, CoinImage, CoinReference, CoinSet, MintLocation } from '@/types'

export const GOLDEN_COIN_FIXTURE_NAMES = [
  'roman-denarius-core',
  'greek-tetradrachm-valued',
  'byzantine-solidus-set-member',
  'wishlist-aureus-target',
  'legacy-era-provincial-bronze',
  'diocletian-follis',
  'image-heavy-drachm',
  'reference-rich-denarius',
] as const

export type GoldenCoinFixtureName = typeof GOLDEN_COIN_FIXTURE_NAMES[number]

export type GoldenCoinTrait =
  | 'roman'
  | 'greek'
  | 'byzantine'
  | 'wishlist'
  | 'set-member'
  | 'image-heavy'
  | 'legacy-custom-era'
  | 'valued'
  | 'reference-rich'
  | 'mint-alias'
  | 'mint-unmatched'
  | 'mint-unknown'

export interface GoldenCoinFixtureInfo {
  name: GoldenCoinFixtureName
  traits: readonly GoldenCoinTrait[]
}

export type CoinFixtureOverrides = Partial<Omit<Coin, 'images' | 'references' | 'sets'>> & {
  images?: readonly CoinImage[]
  references?: readonly CoinReference[]
  sets?: readonly CoinSet[]
}

interface CoinFixtureDefinition {
  coin: Coin
  traits: readonly GoldenCoinTrait[]
}

const fixtureUserId = 101
const createdAt = '2026-01-15T12:00:00Z'
const updatedAt = '2026-01-15T12:00:00Z'

const sets = {
  twelveCaesars: {
    id: 401,
    userId: fixtureUserId,
    name: 'Twelve Caesars',
    description: 'Representative imperial portrait set',
    color: '#c9a84c',
    icon: 'Crown',
    setType: 'goal',
    parentSetId: null,
    targetCompletionDate: '2026-12-31',
    createdAt,
    updatedAt,
  },
  byzantineGold: {
    id: 402,
    userId: fixtureUserId,
    name: 'Byzantine Gold',
    description: 'Gold issues for set-membership workflow tests',
    color: '#b08d57',
    icon: 'CircleDot',
    setType: 'standard',
    parentSetId: null,
    targetCompletionDate: null,
    createdAt,
    updatedAt,
  },
} satisfies Record<string, CoinSet>

const mintLocations: MintLocation[] = [
  {
    id: 1,
    displayName: 'Rome',
    lat: 41.9028,
    lng: 12.4964,
    region: 'Italy',
    aliases: ['Roma', 'Rome mint'],
    createdAt,
    updatedAt,
  },
  {
    id: 2,
    displayName: 'Constantinople',
    lat: 41.0082,
    lng: 28.9784,
    region: 'Thrace',
    aliases: ['Byzantium', 'Istanbul', 'Constantinopolis', 'Constantinople mint'],
    createdAt,
    updatedAt,
  },
  {
    id: 3,
    displayName: 'Athens',
    lat: 37.9838,
    lng: 23.7275,
    region: 'Greece',
    aliases: ['Athenai', 'Athenae', 'Athens mint'],
    createdAt,
    updatedAt,
  },
  {
    id: 4,
    displayName: 'Syracuse',
    lat: 37.0755,
    lng: 15.2866,
    region: 'Sicily',
    aliases: ['Syracusa', 'Syracuse Sicily'],
    createdAt,
    updatedAt,
  },
]

function buildImage(coinId: number, id: number, imageType: CoinImage['imageType'], isPrimary = false): CoinImage {
  return {
    id,
    coinId,
    filePath: `/uploads/test-fixtures/${coinId}-${imageType}-${id}.webp`,
    imageType,
    isPrimary,
    createdAt,
  }
}

function buildReference(coinId: number, id: number, catalog: string, number: string, volume = ''): CoinReference {
  return {
    id,
    coinId,
    catalog,
    volume,
    number,
    uri: `https://example.test/catalog/${catalog.toLowerCase()}/${number}`,
    createdAt,
    updatedAt,
  }
}

function baseCoin(id: number, name: string): Coin {
  return {
    id,
    name,
    category: 'Modern',
    denomination: 'Denarius',
    ruler: 'Trajan',
    era: 'ancient',
    mint: 'Rome',
    material: 'Silver',
    weightGrams: 3.35,
    diameterMm: 18,
    grade: 'VF',
    quantity: 1,
    obverseInscription: 'IMP TRAIANO AVG GER DAC P M TR P',
    reverseInscription: 'SPQR OPTIMO PRINCIPI',
    obverseDescription: 'Laureate bust right',
    reverseDescription: 'Victory standing left',
    rarityRating: 'Common',
    purchasePrice: 180,
    currentValue: 225,
    purchaseDate: '2024-03-15T00:00:00Z',
    purchaseLocation: 'Fixture Dealer',
    vendorSku: '',
    vendorInvoice: '',
    notes: 'Deterministic frontend fixture coin.',
    aiAnalysis: '',
    obverseAnalysis: '',
    reverseAnalysis: '',
    referenceUrl: '',
    referenceText: '',
    isWishlist: false,
    listingStatus: 'unlisted',
    listingCheckedAt: null,
    listingCheckReason: '',
    userId: fixtureUserId,
    images: [buildImage(id, id * 10 + 1, 'obverse', true), buildImage(id, id * 10 + 2, 'reverse')],
    references: [],
    sets: [],
    createdAt,
    updatedAt,
  }
}

const fixtureDefinitions = {
  'roman-denarius-core': {
    coin: {
      ...baseCoin(1001, 'Trajan Denarius Core'),
      referenceUrl: 'https://example.test/coins/roman-denarius-core',
      referenceText: 'RIC II Trajan 147',
    },
    traits: ['roman'],
  },
  'greek-tetradrachm-valued': {
    coin: {
      ...baseCoin(1002, 'Athens Owl Tetradrachm Valued'),
      category: 'Other',
      denomination: 'Tetradrachm',
      ruler: 'Athens',
      mint: 'Athens',
      material: 'Silver',
      weightGrams: 17.18,
      diameterMm: 24,
      purchasePrice: 950,
      currentValue: 1250,
      purchaseDate: '2023-09-10T00:00:00Z',
      obverseInscription: '',
      reverseInscription: 'ΑΘΕ',
      obverseDescription: 'Helmeted head of Athena right',
      reverseDescription: 'Owl standing right, olive sprig and crescent',
    },
    traits: ['greek', 'valued'],
  },
  'byzantine-solidus-set-member': {
    coin: {
      ...baseCoin(1003, 'Justinian I Solidus Set Member'),
      category: 'Other',
      denomination: 'Solidus',
      ruler: 'Justinian I',
      mint: 'Constantinople',
      material: 'Gold',
      weightGrams: 4.48,
      diameterMm: 21,
      purchasePrice: 700,
      currentValue: 875,
      sets: [{ ...sets.byzantineGold }],
    },
    traits: ['byzantine', 'set-member'],
  },
  'wishlist-aureus-target': {
    coin: {
      ...baseCoin(1004, 'Augustus Aureus Wishlist Target'),
      denomination: 'Aureus',
      ruler: 'Augustus',
      material: 'Gold',
      weightGrams: 7.9,
      diameterMm: 20,
      purchasePrice: null,
      currentValue: 8500,
      purchaseDate: null,
      purchaseLocation: '',
      isWishlist: true,
      notes: 'Wishlist target used for purchase workflow tests.',
    },
    traits: ['roman', 'wishlist', 'valued'],
  },
  'legacy-era-provincial-bronze': {
    coin: {
      ...baseCoin(1006, 'Alexandria Provincial Bronze'),
      denomination: 'Drachm',
      ruler: 'Antoninus Pius',
      era: 'Roman Provincial Year 12',
      mint: 'Alexandria',
      material: 'Bronze',
      referenceText: 'Legacy handwritten tray note',
      notes: 'Coin with a custom legacy era value.',
    },
    traits: ['roman', 'legacy-custom-era'],
  },
  'diocletian-follis': {
    coin: {
      ...baseCoin(1007, 'Diocletian Follis'),
      denomination: 'Follis',
      ruler: 'Diocletian',
      material: 'Bronze',
    },
    traits: ['roman'],
  },
  'image-heavy-drachm': {
    coin: {
      ...baseCoin(1008, 'Syracuse Drachm Image Heavy'),
      category: 'Other',
      denomination: 'Drachm',
      ruler: 'Syracuse',
      mint: 'Syracuse',
      images: [
        buildImage(1008, 10081, 'obverse', true),
        buildImage(1008, 10082, 'reverse'),
        buildImage(1008, 10083, 'detail'),
        buildImage(1008, 10084, 'other'),
      ],
    },
    traits: ['greek', 'image-heavy'],
  },
  'reference-rich-denarius': {
    coin: {
      ...baseCoin(1009, 'Vespasian Denarius Reference Rich'),
      ruler: 'Vespasian',
      referenceUrl: 'https://example.test/coins/reference-rich-denarius',
      referenceText: 'RIC II.1 Vespasian 772; BMCRE 161; RSC 554',
      references: [
        buildReference(1009, 1, 'RIC', '772', 'II.1'),
        buildReference(1009, 2, 'BMCRE', '161'),
        buildReference(1009, 3, 'RSC', '554'),
      ],
      sets: [{ ...sets.twelveCaesars }],
    },
    traits: ['roman', 'reference-rich', 'set-member'],
  },
} satisfies Record<GoldenCoinFixtureName, CoinFixtureDefinition>

function cloneCoin(coin: Coin): Coin {
  return {
    ...coin,
    images: coin.images.map((image) => ({ ...image })),
    references: coin.references?.map((reference) => ({ ...reference })),
    sets: coin.sets?.map((set) => ({ ...set })),
  }
}

function applyCoinOverrides(coin: Coin, overrides: CoinFixtureOverrides): Coin {
  return {
    ...coin,
    ...overrides,
    images: overrides.images === undefined ? coin.images : overrides.images.map((image) => ({ ...image })),
    references: overrides.references === undefined
      ? coin.references
      : overrides.references.map((reference) => ({ ...reference })),
    sets: overrides.sets === undefined ? coin.sets : overrides.sets.map((set) => ({ ...set })),
  }
}

export const goldenCoinFixtureCatalog: readonly GoldenCoinFixtureInfo[] = GOLDEN_COIN_FIXTURE_NAMES.map((name) => ({
  name,
  traits: fixtureDefinitions[name].traits,
}))

export function buildGoldenCoinFixture(name: GoldenCoinFixtureName, overrides: CoinFixtureOverrides = {}): Coin {
  return applyCoinOverrides(cloneCoin(fixtureDefinitions[name].coin), overrides)
}

export function buildGoldenCoinFixtures(overrides: Partial<Record<GoldenCoinFixtureName, CoinFixtureOverrides>> = {}): Coin[] {
  return GOLDEN_COIN_FIXTURE_NAMES.map((name) => buildGoldenCoinFixture(name, overrides[name]))
}

export function buildRomanDenariusCore(overrides?: CoinFixtureOverrides): Coin {
  return buildGoldenCoinFixture('roman-denarius-core', overrides)
}

export function buildGreekTetradrachmValued(overrides?: CoinFixtureOverrides): Coin {
  return buildGoldenCoinFixture('greek-tetradrachm-valued', overrides)
}

export function buildByzantineSolidusSetMember(overrides?: CoinFixtureOverrides): Coin {
  return buildGoldenCoinFixture('byzantine-solidus-set-member', overrides)
}

export function buildWishlistAureusTarget(overrides?: CoinFixtureOverrides): Coin {
  return buildGoldenCoinFixture('wishlist-aureus-target', overrides)
}

export function buildDiocletianFollis(overrides?: CoinFixtureOverrides): Coin {
  return buildGoldenCoinFixture('diocletian-follis', overrides)
}

export function buildImageHeavyDrachm(overrides?: CoinFixtureOverrides): Coin {
  return buildGoldenCoinFixture('image-heavy-drachm', overrides)
}

export function buildReferenceRichDenarius(overrides?: CoinFixtureOverrides): Coin {
  return buildGoldenCoinFixture('reference-rich-denarius', overrides)
}

export function buildTestCoinSets(): CoinSet[] {
  return Object.values(sets).map((set) => ({ ...set }))
}

export function buildTestMintLocations(): MintLocation[] {
  return mintLocations.map((location) => ({ ...location, aliases: [...location.aliases] }))
}
