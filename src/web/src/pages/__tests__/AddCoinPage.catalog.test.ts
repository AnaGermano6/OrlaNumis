import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, type PropType } from 'vue'
import type { Coin } from '@/types'
import type { NumistaCatalogType } from '@/api/client'
import type { CatalogSelection } from '@/utils/numistaCatalogPrefill'
import AddCoinPage from '../AddCoinPage.vue'

const mocks = vi.hoisted(() => ({
  add: vi.fn(), push: vi.fn(), listCatalogs: vi.fn(), importCoinImage: vi.fn(), alert: vi.fn(), upload: vi.fn(), official: vi.fn(),
}))
// The photo the user took of the reverse in the form, if any.
let ownReverse: File | null = null
vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-router')>()
  return { ...actual, useRoute: () => ({ query: {} }), useRouter: () => ({ push: mocks.push }) }
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
    expose({ obverseFile: null, get reverseFile() { return ownReverse } })
  },
  template: '<div class="coin-form-stub"><span class="form-name">{{ form.name }}</span><button @click="$emit(\'submit\')">{{ submitLabel }}</button></div>',
})
const BrowserStub = defineComponent({ name: 'NumistaCatalogBrowser', emits: ['use'], template: '<div class="catalog-stub"></div>' })

const TYPE: NumistaCatalogType = {
  id: 5,
  url: 'https://en.numista.com/catalogue/pieces5.html',
  title: '50 Escudos',
  issuer: { code: 'portugal', name: 'Portugal' },
  minYear: 1968,
  maxYear: 1972,
  denomination: '50 Escudos',
  rulers: [],
  composition: 'Silver (.650)',
  weightGrams: 18,
  diameterMm: 34,
  mints: ['Lisbon'],
  obverse: {},
  reverse: {},
  edge: {},
  references: [{ catalogue: 'KM', number: '599' }],
  photos: [],
}

function render() {
  return mount(AddCoinPage, { global: { stubs: { CoinForm: CoinFormStub, NumistaCatalogBrowser: BrowserStub, RouterLink: true, DesktopPageContext: { template: '<div><slot /></div>' } } } })
}

async function chooseFromCatalog(wrapper: ReturnType<typeof render>, selection: CatalogSelection) {
  const catalogButton = wrapper.findAll('button').find((button) => button.text() === 'Catalog')!
  await catalogButton.trigger('click')
  wrapper.findComponent(BrowserStub).vm.$emit('use', selection)
  await flushPromises()
}

describe('Add Coin from the Numista catalog', () => {
  beforeEach(() => {
    Object.values(mocks).forEach((mock) => mock.mockReset())
    ownReverse = null
    mocks.listCatalogs.mockResolvedValue([
      { id: 1, catalog: 'KM', displayName: 'Krause-Mishler', era: 'modern', volumeRequired: false },
      { id: 2, catalog: 'Numista', displayName: 'Numista', era: 'modern', volumeRequired: false },
    ])
    mocks.add.mockResolvedValue({ id: 42 })
    mocks.official.mockResolvedValue({ data: { images: [] } })
  })

  it('fills the form from the chosen coin and saves it with its catalogue references', async () => {
    const wrapper = render()

    await chooseFromCatalog(wrapper, { type: TYPE, issue: { id: 102, year: 1968 }, grade: 'xf', currentValue: 12.5, wishlist: false })

    expect(wrapper.get('.form-name').text()).toBe('50 Escudos 1968 – Portugal')
    expect(wrapper.text()).toContain('Filled in from the Numista catalog: 50 Escudos.')

    const save = wrapper.findAll('button').find((button) => button.text() === 'Add to Collection')!
    await save.trigger('click')
    await flushPromises()

    const payload = mocks.add.mock.calls[0]![0]
    expect(payload).toMatchObject({
      name: '50 Escudos 1968 – Portugal',
      category: 'Modern',
      era: 'modern',
      dateRange: '1968',
      year: 1968,
      country: 'Portugal',
      material: 'Silver',
      weightGrams: 18,
      grade: 'XF',
      currentValue: 12.5,
      isWishlist: false,
      references: [
        { catalog: 'Numista', number: '5', uri: 'https://en.numista.com/catalogue/pieces5.html' },
        { catalog: 'KM', number: '599' },
      ],
    })
    expect(mocks.push).toHaveBeenCalledWith('/coin/42')
  })

  it('saves to the wishlist and keeps the Numista reference when the registry cannot be read', async () => {
    mocks.listCatalogs.mockRejectedValue(new Error('offline'))
    const wrapper = render()

    await chooseFromCatalog(wrapper, { type: TYPE, wishlist: true })
    expect(wrapper.text()).toContain('It will be saved to your wishlist.')

    const save = wrapper.findAll('button').find((button) => button.text() === 'Add to Wishlist')!
    await save.trigger('click')
    await flushPromises()

    const payload = mocks.add.mock.calls[0]![0]
    expect(payload.isWishlist).toBe(true)
    expect(payload.references).toEqual([{ catalog: 'Numista', number: '5', uri: 'https://en.numista.com/catalogue/pieces5.html' }])
    expect(payload.notes).toContain('Other references: KM 599')
  })

  it('imports the chosen real photos after saving and keeps going when a source refuses one', async () => {
    mocks.importCoinImage.mockResolvedValueOnce({ data: {} }).mockRejectedValueOnce(new Error('refused'))
    const wrapper = render()
    const obverse = { source: 'wikimedia' as const, thumbnail: 't', imageUrl: 'https://upload.wikimedia.org/wikipedia/commons/a/aa/O.jpg', author: 'Jane', license: 'CC0' }
    const reverse = { source: 'numista' as const, thumbnail: 't', imageUrl: 'https://en.numista.com/catalogue/photos/x/r.jpg' }

    await chooseFromCatalog(wrapper, { type: TYPE, wishlist: false, photos: { obverse, reverse } })
    expect(wrapper.text()).toContain('Real photos to add: Obverse, Reverse.')

    await wrapper.findAll('button').find((button) => button.text() === 'Add to Collection')!.trigger('click')
    await flushPromises()

    expect(mocks.importCoinImage).toHaveBeenNthCalledWith(1, 42, { url: obverse.imageUrl, imageType: 'obverse', isPrimary: true, credit: 'Obverse photo: Jane, CC0, Wikimedia Commons' })
    expect(mocks.importCoinImage).toHaveBeenNthCalledWith(2, 42, { url: reverse.imageUrl, imageType: 'reverse', isPrimary: false, credit: 'Reverse photo: Numista' })
    expect(mocks.alert).toHaveBeenCalledWith(expect.stringContaining('The source refused the photo (Reverse)'), expect.anything())
    expect(mocks.push).toHaveBeenCalledWith('/coin/42')
  })

  it('asks for the official images of the sides without a photo', async () => {
    mocks.importCoinImage.mockResolvedValue({ data: {} })
    const wrapper = render()
    const obverse = { source: 'wikimedia' as const, thumbnail: 't', imageUrl: 'https://upload.wikimedia.org/wikipedia/commons/a/aa/O.jpg', author: 'Jane', license: 'CC0' }

    await chooseFromCatalog(wrapper, { type: TYPE, wishlist: false, photos: { obverse } })
    await wrapper.findAll('button').find((button) => button.text() === 'Add to Collection')!.trigger('click')
    await flushPromises()

    expect(mocks.official).toHaveBeenCalledTimes(1)
    expect(mocks.official).toHaveBeenCalledWith(42, ['reverse'])
    expect(mocks.importCoinImage).toHaveBeenCalledWith(42, expect.objectContaining({ url: obverse.imageUrl, imageType: 'obverse', isPrimary: true }))
    expect(mocks.push).toHaveBeenCalledWith('/coin/42')
  })

  it('leaves the sides the user photographed to the user', async () => {
    mocks.upload.mockResolvedValue({ data: {} })
    mocks.official.mockResolvedValue({ data: { images: [{ id: 7, coinId: 42, filePath: 'coin-42/o.jpg', imageType: 'obverse', isPrimary: true }] } })
    ownReverse = new File(['x'], 'reverse.jpg', { type: 'image/jpeg' })
    const wrapper = render()

    await chooseFromCatalog(wrapper, { type: TYPE, wishlist: false })
    await wrapper.findAll('button').find((button) => button.text() === 'Add to Collection')!.trigger('click')
    await flushPromises()

    expect(mocks.official).toHaveBeenCalledWith(42, ['obverse'])
    // The official obverse leads, so the user's reverse is not the main image.
    expect(mocks.upload).toHaveBeenCalledWith(42, ownReverse, 'reverse', false, false)
    expect(mocks.push).toHaveBeenCalledWith('/coin/42')
  })

  it('puts the official image in place of a photo its source refused', async () => {
    mocks.importCoinImage.mockRejectedValue(new Error('refused'))
    mocks.official
      .mockResolvedValueOnce({ data: { images: [] } })
      .mockResolvedValueOnce({ data: { images: [{ id: 8, coinId: 42, filePath: 'coin-42/r.jpg', imageType: 'reverse', isPrimary: false }] } })
    const wrapper = render()
    const reverse = { source: 'numista' as const, thumbnail: 't', imageUrl: 'https://en.numista.com/catalogue/photos/x/r.jpg' }

    await chooseFromCatalog(wrapper, { type: TYPE, wishlist: false, photos: { reverse } })
    await wrapper.findAll('button').find((button) => button.text() === 'Add to Collection')!.trigger('click')
    await flushPromises()

    expect(mocks.official).toHaveBeenNthCalledWith(1, 42, ['obverse'])
    expect(mocks.official).toHaveBeenNthCalledWith(2, 42, ['reverse'])
    expect(mocks.alert).not.toHaveBeenCalled()
    expect(mocks.push).toHaveBeenCalledWith('/coin/42')
  })
})
