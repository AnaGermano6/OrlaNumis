import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Coin, CoinMutationPayload, CoinReferenceInput } from '@/types'
import { useUcoinImport } from '../useUcoinImport'
import { setCollectionCurrency } from '@/utils/collectionCurrency'
import { locale, setLocale } from '@/i18n'

const mocks = vi.hoisted(() => ({
  addOfficialCoinPhotos: vi.fn(),
  createCoin: vi.fn(),
  getCoins: vi.fn(),
  getCoinOptions: vi.fn(),
  getCollectionSummary: vi.fn(),
  getNumistaCatalogType: vi.fn(),
  getNumistaTypeMarketValue: vi.fn(),
  listCatalogs: vi.fn(),
  searchNumistaCatalog: vi.fn(),
  searchNumistaIssuers: vi.fn(),
  translateText: vi.fn(),
  updateCoin: vi.fn(),
}))

vi.mock('@/api/client', () => mocks)

const CSV = [
  '"País","Periodo","Denominação","Ano", "Var.", "Moeda", "Assunto", "Diametro, mm","Conservação","Preço, EUR","Número","Etiqueta","Comentário"',
  '"África do Sul","República","5 cêntimos","1984","","Rand","","17.35","","","KM# 84","",""',
  '"África do Sul","República","5 cêntimos","1985","","Rand","","17.35","","","KM# 84","",""',
  '"Alemanha","União Europeia (Euro)","2 euro","2026 D","D","Euro","Bremen, Casa do Clima","25.75","UNC","2.00","UC# 113","",""',
  '"Atlântida","Reino","1 concha","1900","","Concha","","","","","KM# 999","",""',
  '"Japão","Akihito (Heisei)","50 ienes","1991","","Yen japonês","","21","VG","1.00","Y# 101","",""',
].join('\n')

const SOUTH_AFRICA = { code: 'afrique_du_sud', name: 'África do Sul' }
const GERMANY = { code: 'allemagne', name: 'Alemanha, República Federal da' }
const JAPAN = { code: 'japon', name: 'Japão' }

function numistaType(id: number, issuer: { code: string; name: string }, extra: Record<string, unknown>) {
  return {
    id, url: `https://en.numista.com/${id}`, title: 'Type', issuer, rulers: [], mints: [], obverse: {}, reverse: {}, edge: {}, references: [], photos: [],
    ...extra,
  }
}

const TYPES: Record<number, unknown> = {
  1791: {
    status: 'ok',
    type: numistaType(1791, SOUTH_AFRICA, { title: '5 Cents', minYear: 1970, maxYear: 1989, composition: 'Nickel', weightGrams: 2.5, diameterMm: 17, mints: ['Pretoria'], references: [{ catalogue: 'KM', number: '84' }] }),
    issues: [{ id: 11, year: 1984, mintage: 20_000_000 }, { id: 12, year: 1985, mintage: 500_000 }],
  },
  474550: {
    status: 'ok',
    type: numistaType(474550, GERMANY, { title: '2 Euros (Bundesländer II - "Bremen")', minYear: 2026, maxYear: 2026, composition: 'Bimetallic: nickel brass clad nickel centre in copper-nickel ring', weightGrams: 8.5, mints: ['Berlin', 'Munich'] }),
    issues: [{ id: 21, year: 2026, mintLetter: 'A', mintage: 6_000_000 }, { id: 22, year: 2026, mintLetter: 'D', mintage: 6_300_000 }],
  },
  2120: {
    status: 'ok',
    type: numistaType(2120, JAPAN, { title: '50 Yen - Heisei', minYear: 1989, maxYear: 2019, composition: 'Copper-nickel', weightGrams: 4 }),
    issues: [{ id: 31, year: 1991, mintage: 2_000_000 }],
  },
}

const PRICES: Record<number, { grade: string; price: number }[]> = {
  11: [{ grade: 'vf', price: 0.25 }],
  12: [{ grade: 'xf', price: 2 }],
  22: [{ grade: 'xf', price: 2.6 }, { grade: 'unc', price: 3.1 }],
  31: [{ grade: 'vg', price: 0.4 }],
}

// Fields the API client sends as null when an update leaves them out.
const NULLABLE_FIELDS = ['weightGrams', 'diameterMm', 'purchasePrice', 'currentValue', 'purchaseDate', 'storageLocationId', 'storageSlot', 'romanImperialFigureId', 'mintLocationId', 'year']

let collection: Coin[]
let nextId: number

function storedReferences(coinId: number, references: CoinReferenceInput[] | undefined) {
  return (references ?? []).map((ref, index) => ({
    id: coinId * 10 + index, coinId, catalog: ref.catalog, volume: ref.volume ?? '', number: ref.number, uri: ref.uri ?? '', createdAt: '', updatedAt: '',
  }))
}

function csvFile(text: string) {
  return { name: 'collection.csv', text: async () => text } as File
}

describe('useUcoinImport', () => {
  beforeEach(() => {
    Object.values(mocks).forEach((mock) => mock.mockReset())
    localStorage.clear()
    collection = []
    nextId = 0
    mocks.getCoinOptions.mockRejectedValue(new Error('offline'))
    mocks.translateText.mockResolvedValue({ data: { status: 'unavailable' } })
    mocks.addOfficialCoinPhotos.mockResolvedValue({ data: { images: [] } })
    mocks.getCollectionSummary.mockResolvedValue({ data: { exchangeRate: { usdPerEur: 1.12, date: '2026-10-09' } } })
    mocks.listCatalogs.mockResolvedValue([
      { id: 1, catalog: 'KM', displayName: 'Krause-Mishler', era: 'modern', volumeRequired: false },
      { id: 2, catalog: 'Y', displayName: 'Y Number', era: 'modern', volumeRequired: false },
      { id: 3, catalog: 'Numista', displayName: 'Numista', era: 'modern', volumeRequired: false },
    ])
    mocks.getCoins.mockImplementation(async ({ page, limit }: { page: number; limit: number }) => ({
      data: { coins: collection.slice((page - 1) * limit, page * limit), total: collection.length },
    }))
    mocks.createCoin.mockImplementation(async (payload: CoinMutationPayload) => {
      nextId += 1
      const coin = { mint: '', ruler: '', material: 'Other', rarity: '', weightGrams: null, currentValue: null, ...payload, id: nextId, references: storedReferences(nextId, payload.references) } as Coin
      collection.push(coin)
      return { data: coin }
    })
    mocks.updateCoin.mockImplementation(async (id: number, payload: CoinMutationPayload) => {
      const coin = collection.find((entry) => entry.id === id)!
      // Like the real client, a nullable field left out is sent as null.
      const sent: Record<string, unknown> = { ...payload }
      for (const field of NULLABLE_FIELDS) if (sent[field] === undefined) sent[field] = null
      Object.assign(coin, sent, { references: storedReferences(id, payload.references) })
      return { data: coin }
    })
    mocks.searchNumistaCatalog.mockImplementation(async (params: { catalogue?: string; number?: string; issuer?: string; year?: number }) => {
      // Numista's year filter misses coins dated in another calendar (Heisei 3).
      if (params.catalogue === 'Y' && params.number === '101' && params.issuer === 'japon' && params.year === undefined) {
        return { data: { status: 'ok', count: 1, page: 1, types: [{ id: 2120, title: '50 Yen - Heisei', issuer: JAPAN, minYear: 1989, maxYear: 2019 }] } }
      }
      if (params.catalogue === 'KM' && params.number === '84') {
        return { data: { status: 'ok', count: 1, page: 1, types: [{ id: 1791, title: '5 Cents (SUID-AFRIKA - SOUTH AFRICA)', issuer: SOUTH_AFRICA, minYear: 1970, maxYear: 1989 }] } }
      }
      if (params.issuer === 'allemagne' && params.year === 2026) {
        return {
          data: {
            status: 'ok', count: 3, page: 1, types: [
              { id: 6323, title: '2 Euros (2nd map)', issuer: GERMANY, minYear: 2008, maxYear: 2026 },
              { id: 544530, title: '2 Euros (Konrad Adenauer)', issuer: GERMANY, minYear: 2026, maxYear: 2026 },
              { id: 474550, title: '2 Euros (Bundesländer II - "Bremen")', issuer: GERMANY, minYear: 2026, maxYear: 2026 },
            ],
          },
        }
      }
      return { data: { status: 'ok', count: 0, page: 1, types: [] } }
    })
    mocks.searchNumistaIssuers.mockImplementation(async (q: string) => ({
      data: { status: 'ok', issuers: q === 'Alemanha' ? [GERMANY, { code: 'allemagne-pre1945', name: 'Alemanha (1871-1948)' }] : q === 'Japão' ? [JAPAN] : [] },
    }))
    mocks.getNumistaCatalogType.mockImplementation(async (typeId: number) => ({ data: TYPES[typeId] }))
    mocks.getNumistaTypeMarketValue.mockImplementation(async (_typeId: number, params: { issueId: number }) => ({
      data: { status: 'ok', currency: 'USD', prices: PRICES[params.issueId] ?? [], issues: [], stale: false, catalogPhotos: [] },
    }))
  })

  it('adds the coins of the file and completes them from Numista, one lookup per type', async () => {
    const importer = useUcoinImport({ paceMs: 0 })
    await importer.readFile(csvFile(CSV))
    expect(importer.phase.value).toBe('ready')
    expect(importer.pendingRows.value).toHaveLength(5)
    expect(importer.countries.value).toBe(4)

    await importer.start(true)

    expect(importer.phase.value).toBe('finished')
    expect(mocks.createCoin).toHaveBeenCalledTimes(5)
    expect(importer.progress).toMatchObject({ created: 5, createFailed: 0, toEnrich: 5, enriched: 4, notFound: 1, failed: 0 })
    expect(importer.notFoundCoins.value.map((coin) => coin.name)).toEqual(['1 concha 1900 – Atlântida'])

    // Both South African coins share one search and one type lookup.
    const searches = mocks.searchNumistaCatalog.mock.calls.map(([params]) => params)
    expect(searches.filter((params) => params.number === '84')).toHaveLength(1)
    expect(searches).toContainEqual(expect.objectContaining({ issuer: 'allemagne', year: 2026, q: 'euro', count: 50 }))
    expect(mocks.getNumistaCatalogType).toHaveBeenCalledTimes(3)

    const [sa1984, sa1985, bremen, atlantis, japan] = collection
    expect(sa1984).toMatchObject({ rarity: 'common', currentValue: 0.25, weightGrams: 2.5, diameterMm: 17.35, mint: 'Pretoria' })
    expect(sa1984.references?.map((ref) => `${ref.catalog} ${ref.number}`)).toEqual(['KM 84', 'Numista 1791'])
    expect(sa1985).toMatchObject({ rarity: 'scarce', currentValue: 2 })
    // The mint mark from the file picks the issue; the UNC grade picks the price.
    expect(bremen).toMatchObject({ mint: 'D', rarity: 'uncommon', currentValue: 3.1, purchasePrice: 2.24, grade: 'UNC', name: '2 euro 2026 D – Bremen, Casa do Clima' })
    expect(bremen.references?.map((ref) => `${ref.catalog} ${ref.number}`)).toEqual(['Numista 474550'])
    expect(bremen.notes).toContain('uCoin')
    expect(atlantis.references?.map((ref) => ref.catalog)).toEqual(['KM'])
    expect(japan).toMatchObject({ rarity: 'uncommon', currentValue: 0.4 })
    expect(japan.references?.map((ref) => `${ref.catalog} ${ref.number}`)).toEqual(['Y 101', 'Numista 2120'])
    expect(importer.awaitingNumista.value).toHaveLength(0)

    // Importing the same file again adds nothing.
    await importer.readFile(csvFile(CSV))
    expect(importer.pendingRows.value).toHaveLength(0)
    expect(importer.alreadyImported.value).toBe(5)
  })

  it('adds the official photos of the imported coins after the Numista details', async () => {
    mocks.addOfficialCoinPhotos.mockImplementation(async (id: number) => {
      const coin = collection.find((entry) => entry.id === id)!
      // Only the €2 coin has official images.
      const images = coin.denomination === '2 euro' ? [{ id: 1, coinId: id, filePath: 'a.jpg', imageType: 'obverse', isPrimary: true }] : []
      coin.images = images as Coin['images']
      return { data: { images } }
    })
    const importer = useUcoinImport({ paceMs: 0 })
    await importer.readFile(csvFile(CSV))
    await importer.start(true)

    expect(importer.phase.value).toBe('finished')
    expect(mocks.addOfficialCoinPhotos).toHaveBeenCalledTimes(5)
    expect(mocks.addOfficialCoinPhotos).toHaveBeenCalledWith(collection[2]!.id, ['obverse', 'reverse'])
    expect(importer.progress).toMatchObject({ toPhoto: 5, photoDone: 5, photographed: 1 })
    expect(importer.withoutImages.value).toHaveLength(4)
    // Photos are asked once the coin has its Numista reference.
    const bremenCall = mocks.addOfficialCoinPhotos.mock.invocationCallOrder[mocks.addOfficialCoinPhotos.mock.calls.findIndex(([id]) => id === collection[2]!.id)]!
    expect(bremenCall).toBeGreaterThan(Math.max(...mocks.updateCoin.mock.invocationCallOrder))

    // Coins imported earlier get theirs on request; the ones with an image are left alone.
    mocks.addOfficialCoinPhotos.mockClear()
    await importer.addMissingPhotos()
    expect(mocks.addOfficialCoinPhotos).toHaveBeenCalledTimes(4)
    expect(importer.phase.value).toBe('finished')
  })

  it('keeps importing when a photo cannot be added', async () => {
    mocks.addOfficialCoinPhotos.mockRejectedValue(new Error('offline'))
    const importer = useUcoinImport({ paceMs: 0 })
    await importer.readFile(csvFile(CSV))
    await importer.start(false)
    expect(importer.progress).toMatchObject({ created: 5, toPhoto: 5, photoDone: 5, photographed: 0 })
    expect(importer.phase.value).toBe('finished')
  })

  it('stops when the Numista quota runs out and resumes later', async () => {
    const importer = useUcoinImport({ paceMs: 0 })
    await importer.readFile(csvFile(CSV))
    const search = mocks.searchNumistaCatalog.getMockImplementation()!
    mocks.searchNumistaCatalog.mockResolvedValue({ data: { status: 'quota_limited', count: 0, page: 1, types: [] } })

    await importer.start(true)

    expect(importer.progress.created).toBe(5)
    expect(importer.stopped.value).toBe(true)
    expect(importer.stopMessage.value).not.toBe('')
    expect(mocks.updateCoin).not.toHaveBeenCalled()
    expect(importer.awaitingNumista.value).toHaveLength(5)

    mocks.searchNumistaCatalog.mockImplementation(search)
    await importer.enrich()
    expect(importer.stopped.value).toBe(false)
    expect(importer.progress.enriched).toBe(4)
    expect(mocks.createCoin).toHaveBeenCalledTimes(5)
  })

  it('remembers coins Numista does not have until asked to try again', async () => {
    const importer = useUcoinImport({ paceMs: 0 })
    await importer.readFile(csvFile(CSV))
    await importer.start(true)
    expect(importer.progress.notFound).toBe(1)

    await importer.enrich()
    expect(importer.progress.toEnrich).toBe(0)

    expect(importer.earlierNotFound.value).toBe(1)

    // A later visit offers to try them again.
    const later = useUcoinImport({ paceMs: 0 })
    await later.init()
    expect(later.earlierNotFound.value).toBe(1)
    expect(later.awaitingNumista.value).toHaveLength(0)

    importer.retryNotFound()
    await importer.enrich()
    expect(importer.progress.toEnrich).toBe(1)
    expect(importer.progress.notFound).toBe(1)
  })

  it('tries the subject in English when the Portuguese one names no Numista type', async () => {
    const previous = locale.value
    setLocale('pt-PT')
    try {
      const csv = [
        CSV.split('\n')[0],
        '"Portugal","União Europeia (Euro)","2 euro","2023","","Euro","Jornada Mundial da Juventude 2023 Lisboa","25.75","","","UC# 116","",""',
        '"Portugal","União Europeia (Euro)","2 euro","2023","","Euro","Paz Entre as Nações","25.75","","","UC# 117","",""',
      ].join('\n')
      const portugal = { code: 'portugal', name: 'Portugal' }
      mocks.searchNumistaIssuers.mockResolvedValue({ data: { status: 'ok', issuers: [portugal] } })
      mocks.searchNumistaCatalog.mockResolvedValue({
        data: {
          status: 'ok', count: 3, page: 1, types: [
            { id: 9897, title: '2 Euros (2nd map)', issuer: portugal, minYear: 2008, maxYear: 2026 },
            { id: 401, title: '2 Euros (World Youth Day)', issuer: portugal, minYear: 2023, maxYear: 2023 },
            { id: 402, title: '2 Euros (Peace Among Nations)', issuer: portugal, minYear: 2023, maxYear: 2023 },
          ],
        },
      })
      mocks.translateText.mockImplementation(async (q: string, from: string) => ({
        data: { status: 'ok', text: from === 'pt' ? ({ 'Jornada Mundial da Juventude 2023 Lisboa': 'World Youth Day Lisbon 2023', 'Paz Entre as Nações': 'Among the Nations' } as Record<string, string>)[q] : q },
      }))
      mocks.getNumistaCatalogType.mockImplementation(async (typeId: number) => ({
        data: { status: 'ok', type: numistaType(typeId, portugal, { title: 'Type' }), issues: [{ id: typeId * 10, year: 2023 }] },
      }))

      const importer = useUcoinImport({ paceMs: 0 })
      await importer.readFile(csvFile(csv))
      await importer.start(true)

      expect(importer.progress).toMatchObject({ created: 2, enriched: 2, notFound: 0 })
      expect(collection.map((coin) => coin.references?.find((ref) => ref.catalog === 'Numista')?.number)).toEqual(['401', '402'])
      expect(mocks.translateText).toHaveBeenCalledWith('Jornada Mundial da Juventude 2023 Lisboa', 'pt')
    } finally {
      setLocale(previous)
    }
  })

  it('finds issuers Numista names differently and prefers a country to a province', async () => {
    const csv = [
      CSV.split('\n')[0],
      '"Região Autónoma dos Açores","República","100 escudos","1991","","Escudo","","","","","KM# 46","",""',
      '"Síria","República Árabe","50 piastres","1974","","Libra síria","","","","","KM# 108","",""',
    ].join('\n')
    const azores = { code: 'azores', name: 'Açores' }
    const syria = { code: 'syrie', name: 'Síria' }
    mocks.searchNumistaIssuers.mockImplementation(async (q: string) => ({
      data: {
        status: 'ok',
        issuers: q === 'Açores' ? [azores] : q === 'Síria' ? [{ code: 'syria_section', name: 'Síria', parent: 'Grécia (antiga)' }, syria] : [],
      },
    }))
    mocks.searchNumistaCatalog.mockImplementation(async (params: { number?: string; issuer?: string; year?: number }) => {
      const found = params.year === undefined && ((params.issuer === 'azores' && params.number === '46') || (params.issuer === 'syrie' && params.number === '108'))
      const item = params.issuer === 'azores'
        ? { id: 900, title: '100 Escudos (Antero de Quental)', issuer: azores, minYear: 1991, maxYear: 1991 }
        : { id: 901, title: '50 Piastres', issuer: syria, minYear: 1971, maxYear: 1979 }
      return { data: { status: 'ok', count: found ? 1 : 0, page: 1, types: found ? [item] : [] } }
    })
    mocks.getNumistaCatalogType.mockImplementation(async (typeId: number) => ({
      data: { status: 'ok', type: numistaType(typeId, typeId === 900 ? azores : syria, { title: 'Type' }), issues: [{ id: typeId * 10, year: typeId === 900 ? 1991 : 1974 }] },
    }))

    const importer = useUcoinImport({ paceMs: 0 })
    await importer.readFile(csvFile(csv))
    await importer.start(true)

    expect(importer.progress).toMatchObject({ created: 2, enriched: 2, notFound: 0 })
    expect(mocks.searchNumistaIssuers.mock.calls.map(([q]) => q)).toEqual(expect.arrayContaining(['Região Autónoma dos Açores', 'Açores', 'Síria']))
    expect(collection.map((coin) => coin.references?.find((ref) => ref.catalog === 'Numista')?.number)).toEqual(['900', '901'])
  })

  describe('in a euro collection', () => {
    afterEach(() => setCollectionCurrency('USD'))

    it('keeps the file\'s euro prices and converts Numista\'s dollar prices with the ECB rate', async () => {
      setCollectionCurrency('EUR')
      const importer = useUcoinImport({ paceMs: 0 })
      await importer.readFile(csvFile(CSV.split('\n').slice(0, 4).join('\n')))
      await importer.start(true)

      const [sa1984, , bremen] = collection
      expect(bremen).toMatchObject({ purchasePrice: 2, currentValue: 2.77 })
      expect(sa1984).toMatchObject({ currentValue: 0.22 })
      // Prices are asked in dollars, as cached, and converted here.
      expect(mocks.getNumistaTypeMarketValue.mock.calls.every(([, params]) => params.currency === 'USD')).toBe(true)
    })

    it('asks Numista for euros when there is no exchange rate', async () => {
      setCollectionCurrency('EUR')
      mocks.getCollectionSummary.mockRejectedValue(new Error('offline'))
      const importer = useUcoinImport({ paceMs: 0 })
      await importer.readFile(csvFile(CSV.split('\n').slice(0, 4).join('\n')))
      await importer.start(true)

      expect(mocks.getNumistaTypeMarketValue.mock.calls.every(([, params]) => params.currency === 'EUR')).toBe(true)
      expect(collection[2]).toMatchObject({ purchasePrice: 2, currentValue: 3.1 })
    })
  })
})
