import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, type PropType } from 'vue'
import type { Coin } from '@/types'
import type { CoinRelease } from '@/api/client'
import AddCoinPage from '../AddCoinPage.vue'
import { releaseAddRoute } from '@/utils/releasePrefill'

const mocks = vi.hoisted(() => ({
  add: vi.fn(), push: vi.fn(), listCatalogs: vi.fn(), importCoinImage: vi.fn(), alert: vi.fn(), upload: vi.fn(), official: vi.fn(),
  route: { query: {} as Record<string, unknown> },
}))
vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-router')>()
  return { ...actual, useRoute: () => mocks.route, useRouter: () => ({ push: mocks.push }) }
})
vi.mock('@/stores/coins', () => ({ useCoinsStore: () => ({ addCoin: mocks.add }) }))
vi.mock('@/composables/usePwa', () => ({ usePwa: () => ({ isPwa: false }) }))
vi.mock('@/composables/useDialog', () => ({ useDialog: () => ({ showAlert: mocks.alert, showConfirm: vi.fn() }) }))
vi.mock('@/api/client', () => ({
  listCatalogs: mocks.listCatalogs, importCoinImage: mocks.importCoinImage, uploadImage: mocks.upload, addOfficialCoinPhotos: mocks.official, updateCoin: vi.fn(), extractText: vi.fn(),
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
})
