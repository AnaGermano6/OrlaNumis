import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, type PropType } from 'vue'
import type { Coin } from '@/types'
import type { CoinRelease } from '@/api/client'
import AddCoinPage from '../AddCoinPage.vue'
import { releaseAddRoute } from '@/utils/releasePrefill'

const mocks = vi.hoisted(() => ({
  add: vi.fn(), push: vi.fn(), listCatalogs: vi.fn(), importCoinImage: vi.fn(), alert: vi.fn(), upload: vi.fn(), official: vi.fn(),
  releaseType: vi.fn(), catalogType: vi.fn(), matches: vi.fn(), confirm: vi.fn(), remove: vi.fn(), toast: vi.fn(),
  route: { query: {} as Record<string, unknown> },
}))
vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-router')>()
  return { ...actual, useRoute: () => mocks.route, useRouter: () => ({ push: mocks.push }) }
})
vi.mock('@/stores/coins', () => ({ useCoinsStore: () => ({ addCoin: mocks.add, removeCoin: mocks.remove }) }))
vi.mock('@/composables/useToast', () => ({ useToast: () => ({ showToast: mocks.toast }) }))
vi.mock('@/composables/usePwa', () => ({ usePwa: () => ({ isPwa: false }) }))
vi.mock('@/composables/useDialog', () => ({ useDialog: () => ({ showAlert: mocks.alert, showConfirm: mocks.confirm }) }))
vi.mock('@/api/client', () => ({
  listCatalogs: mocks.listCatalogs, importCoinImage: mocks.importCoinImage, uploadImage: mocks.upload, addOfficialCoinPhotos: mocks.official, updateCoin: vi.fn(), extractText: vi.fn(),
  getReleaseNumistaType: mocks.releaseType, getNumistaCatalogType: mocks.catalogType, getWishlistMatches: mocks.matches,
}))
vi.mock('@/composables/useCoinOptions', async () => {
  const { ref } = await import('vue')
  return { useCoinOptions: () => ({
    categoryOptions: ref(['Roman', 'Modern', 'Other']), materialOptions: ref(['Silver', 'Other']),
    eraOptions: ref(['ancient', 'modern']), loadOptions: vi.fn(),
  }) }
})

const CoinFormStub = defineComponent({
  props: { form: { type: Object as PropType<Partial<Coin>>, required: true }, submitLabel: String, loading: Boolean },
  emits: ['submit'],
  setup(_, { expose }) {
    expose({ obverseFile: null, reverseFile: null })
  },
  template: '<div class="coin-form-stub"><span class="form-name">{{ form.name }}</span><button @click="$emit(\'submit\')">{{ submitLabel }}</button></div>',
})

function release(fields: Partial<CoinRelease>): CoinRelease {
  return {
    source: 'ecb', kind: 'commemorative', country: 'PT', title: 'Sustainable development', faceValue: 2, imageUrl: 'https://ecb/pt.jpg',
    pageUrl: 'https://www.ecb.europa.eu/euro/coins/comm/html/comm_2025.en.html', from: '2025-01', to: '2025-03', upcoming: false, owned: false, ...fields,
  }
}

// Numista's type of the Portuguese 2025 €2 coin.
const NUMISTA_TYPE = {
  id: 4321, url: 'https://en.numista.com/catalogue/pieces4321.html', title: '2 Euros (Sustainable development)', issuer: { code: 'portugal', name: 'Portugal' },
  minYear: 2025, maxYear: 2025, denomination: '2 Euros', rulers: [], composition: 'Bimetallic: nickel brass clad nickel centre in copper-nickel ring',
  weightGrams: 8.5, diameterMm: 25.75, mints: ['Imprensa Nacional-Casa da Moeda'], obverse: { description: 'A tree', lettering: 'PORTUGAL\n2025' },
  reverse: { description: 'The map of Europe' }, edge: {}, references: [{ catalogue: 'KM', number: '1000' }], photos: [],
}
const CANDIDATES = [
  { id: 4321, title: '2 Euros (Sustainable development)', issuer: { code: 'portugal', name: 'Portugal' }, minYear: 2025, maxYear: 2025 },
  { id: 4322, title: '2 Euros (Rotary)', issuer: { code: 'portugal', name: 'Portugal' }, minYear: 2025, maxYear: 2025 },
]

function form(wrapper: ReturnType<typeof open>): Partial<Coin> {
  return wrapper.findComponent(CoinFormStub).props('form') as Partial<Coin>
}

function open(value: CoinRelease) {
  const location = releaseAddRoute(value) as { query: Record<string, unknown> }
  mocks.route.query = location.query
  return mount(AddCoinPage, { global: { stubs: { CoinForm: CoinFormStub, NumistaCatalogBrowser: true, RouterLink: true, DesktopPageContext: { template: '<div><slot /></div>' } } } })
}

async function save(wrapper: ReturnType<typeof open>) {
  await wrapper.findAll('button').find((button) => button.text() === 'Add to Collection' || button.text() === 'Add to Wishlist')!.trigger('click')
  await flushPromises()
  return mocks.add.mock.calls[0]![0]
}

describe('Add Coin from a new release', () => {
  beforeEach(() => {
    Object.values(mocks).forEach((mock) => typeof mock === 'function' && mock.mockReset())
    mocks.route.query = {}
    mocks.add.mockResolvedValue({ id: 42 })
    mocks.official.mockResolvedValue({ data: { images: [{ imageType: 'obverse' }, { imageType: 'reverse' }] } })
    mocks.listCatalogs.mockResolvedValue([{ catalog: 'KM', volumeRequired: false }])
    mocks.releaseType.mockResolvedValue({ data: { status: 'ok', candidates: [] } })
    mocks.matches.mockResolvedValue({ data: { coins: [] } })
    mocks.catalogType.mockResolvedValue({ data: { status: 'ok', type: NUMISTA_TYPE, issues: [{ id: 1, year: 2025, mintage: 500000 }] } })
  })

  it('fills the form with the release, says where it comes from, and adds the official images when saved', async () => {
    const wrapper = open(release({}))
    await flushPromises()

    expect(wrapper.get('.form-name').text()).toBe('2 euro 2025 – Sustainable development')
    expect(wrapper.get('[data-testid="release-source"]').text()).toContain('Filled in from the new coin: Sustainable development.')

    const payload = await save(wrapper)
    expect(payload).toMatchObject({
      name: '2 euro 2025 – Sustainable development',
      category: 'Modern',
      material: 'Other',
      denomination: '2 euro',
      year: 2025,
      country: 'Portugal',
      circulation: 'current_commemorative',
      referenceUrl: 'https://www.ecb.europa.eu/euro/coins/comm/html/comm_2025.en.html',
      referenceText: 'ECB',
      quantity: 1,
      isWishlist: false,
    })
    expect(payload.references).toBeUndefined()
    expect(mocks.official).toHaveBeenCalledWith(42, ['obverse', 'reverse'])
    expect(mocks.push).toHaveBeenCalledWith('/coin/42')
  })

  it('saves a coin from Numista with its Numista reference, and one still to come to the wishlist', async () => {
    const escudo = open(release({ source: 'numista', kind: 'collector', title: 'Expo 98', faceValue: 1000, currency: 'PTE', pageUrl: 'https://pt.numista.com/catalogue/pieces9002.html', from: '1998-01' }))
    await flushPromises()
    expect(await save(escudo)).toMatchObject({
      name: '1000 escudos 1998 – Expo 98',
      denomination: '1000 escudos',
      circulation: 'non_circulating',
      referenceText: 'Numista',
      references: [{ catalog: 'Numista', number: '9002', uri: 'https://pt.numista.com/catalogue/pieces9002.html' }],
    })

    mocks.add.mockClear()
    const upcoming = open(release({ source: 'incm', kind: 'collector', title: 'Natal', faceValue: 7.5, variants: ['Prata Proof'], from: '2026-12', upcoming: true, volume: 5000 }))
    await flushPromises()
    expect(upcoming.text()).toContain('It will be saved to your wishlist.')
    expect(await save(upcoming)).toMatchObject({
      name: '7½ euro 2026 – Natal',
      material: 'Silver',
      isWishlist: true,
      notes: 'Versions: Prata Proof\nMintage: 5,000',
    })
  })

  it('opens the empty form when the address carries no readable release', async () => {
    mocks.route.query = { release: '{not json' }
    const wrapper = mount(AddCoinPage, { global: { stubs: { CoinForm: CoinFormStub, NumistaCatalogBrowser: true, RouterLink: true, DesktopPageContext: { template: '<div><slot /></div>' } } } })
    await flushPromises()

    expect(wrapper.get('.form-name').text()).toBe('')
    expect(wrapper.find('[data-testid="release-source"]').exists()).toBe(false)
  })

  it("completes the coin with its Numista type, leaving what the user changed meanwhile", async () => {
    let answer!: (value: unknown) => void
    mocks.releaseType.mockReturnValue(new Promise((resolve) => { answer = resolve }))
    const wrapper = open(release({}))
    await flushPromises()
    expect(wrapper.get('[data-testid="release-numista"]').text()).toBe('Looking for this coin on Numista…')
    expect(mocks.releaseType).toHaveBeenCalledWith(expect.objectContaining({ source: 'ecb', country: 'PT', title: 'Sustainable development', imageUrl: 'https://ecb/pt.jpg' }))

    form(wrapper).mint = 'Lisbon'
    answer({ data: { status: 'ok', typeId: 4321, candidates: CANDIDATES } })
    await flushPromises()
    expect(mocks.catalogType).toHaveBeenCalledWith(4321, 'en')
    expect(wrapper.get('[data-testid="release-numista"]').text()).toContain('Completed with the Numista details of 2 Euros (Sustainable development)')
    expect(wrapper.get('[data-testid="release-numista"] a').attributes('href')).toBe('https://en.numista.com/catalogue/pieces4321.html')

    const payload = await save(wrapper)
    expect(payload).toMatchObject({
      name: '2 euro 2025 – Sustainable development',
      country: 'Portugal',
      mint: 'Lisbon',
      weightGrams: 8.5,
      diameterMm: 25.75,
      obverseInscription: 'PORTUGAL / 2025',
      reverseDescription: 'The map of Europe',
      references: [{ catalog: 'Numista', number: '4321', uri: 'https://en.numista.com/catalogue/pieces4321.html' }, { catalog: 'KM', number: '1000' }],
    })
    expect(payload.notes).toContain('Mintage: 500,000')
  })

  it("lets the user choose among Numista's coins when none is sure, and go back to the release alone", async () => {
    mocks.releaseType.mockResolvedValue({ data: { status: 'ok', candidates: CANDIDATES } })
    const wrapper = open(release({}))
    await flushPromises()
    expect(mocks.catalogType).not.toHaveBeenCalled()
    const candidates = wrapper.findAll('[data-testid="release-numista-candidate"]')
    expect(candidates.map((candidate) => candidate.text())).toEqual(['2 Euros (Sustainable development)2025', '2 Euros (Rotary)2025'])

    await candidates[0]!.trigger('click')
    await flushPromises()
    expect(form(wrapper).weightGrams).toBe(8.5)
    await wrapper.get('[data-testid="release-numista-other"]').trigger('click')
    expect(wrapper.get('[data-testid="release-numista-candidate"][aria-pressed="true"]').text()).toContain('Sustainable development')
    await wrapper.get('[data-testid="release-numista-only"]').trigger('click')
    expect(form(wrapper).weightGrams).toBeUndefined()
    expect(form(wrapper).name).toBe('2 euro 2025 – Sustainable development')
    expect((await save(wrapper)).references).toBeUndefined()
  })

  it("says when Numista cannot be asked, and keeps the release's details", async () => {
    mocks.releaseType.mockResolvedValue({ data: { status: 'quota_limited', candidates: [] } })
    const wrapper = open(release({}))
    await flushPromises()
    expect(wrapper.get('[data-testid="release-numista"] [role="alert"]').text()).not.toBe('')
    expect(form(wrapper).name).toBe('2 euro 2025 – Sustainable development')

    mocks.releaseType.mockResolvedValue({ data: { status: 'ok', candidates: [] } })
    const none = open(release({}))
    await flushPromises()
    expect(none.get('[data-testid="release-numista"]').text()).toContain('Numista has no coin of this country, year and face value')
  })

  it('offers to take the coin off the wishlist when it was there, before opening its page', async () => {
    mocks.matches.mockResolvedValue({ data: { coins: [{ id: 7, name: '2 euro 2025 – Sustainable development' }] } })
    mocks.confirm.mockResolvedValueOnce(true)
    const wrapper = open(release({}))
    await flushPromises()
    await save(wrapper)

    expect(mocks.matches).toHaveBeenCalledWith(42)
    expect(mocks.confirm).toHaveBeenCalledWith(
      'This coin was on your wishlist: «2 euro 2025 – Sustainable development». Take it off the wishlist now that you have it?',
      expect.objectContaining({ title: 'On your wishlist', confirmLabel: 'Take it off', cancelLabel: 'Keep on the wishlist' }),
    )
    expect(mocks.remove).toHaveBeenCalledWith(7)
    expect(mocks.toast).toHaveBeenCalledWith('Taken off the wishlist.', 'success')
    expect(mocks.push).toHaveBeenCalledWith('/coin/42')

    // Kept: nothing is deleted; and a coin on no wishlist asks nothing.
    mocks.add.mockClear()
    mocks.remove.mockClear()
    mocks.confirm.mockReset()
    mocks.confirm.mockResolvedValueOnce(false)
    const kept = open(release({}))
    await flushPromises()
    await save(kept)
    expect(mocks.confirm).toHaveBeenCalledTimes(1)
    expect(mocks.remove).not.toHaveBeenCalled()

    mocks.add.mockClear()
    mocks.confirm.mockClear()
    mocks.matches.mockResolvedValue({ data: { coins: [] } })
    const other = open(release({}))
    await flushPromises()
    await save(other)
    expect(mocks.confirm).not.toHaveBeenCalled()
  })
})
