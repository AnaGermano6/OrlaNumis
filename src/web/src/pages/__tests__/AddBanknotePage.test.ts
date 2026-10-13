import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AddBanknotePage from '../AddBanknotePage.vue'
import { setLocale } from '@/i18n'
import type { CatalogSelection } from '@/utils/numistaCatalogPrefill'

const mocks = vi.hoisted(() => ({
  createCoin: vi.fn(),
  uploadImage: vi.fn(),
  importCoinImage: vi.fn(),
  listCatalogs: vi.fn(),
  push: vi.fn(),
  showAlert: vi.fn(),
  front: null as File | null,
}))

vi.mock('@/api/client', () => ({
  createCoin: mocks.createCoin,
  uploadImage: mocks.uploadImage,
  importCoinImage: mocks.importCoinImage,
  listCatalogs: mocks.listCatalogs,
  getCoinOptions: vi.fn().mockResolvedValue({ data: {} }),
  getSettings: vi.fn().mockResolvedValue({ data: {} }),
}))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: mocks.push }) }))
vi.mock('@/composables/useDialog', () => ({ useDialog: () => ({ showAlert: mocks.showAlert }) }))
vi.mock('@/composables/useCoinOptions', () => ({
  useCoinOptions: () => ({
    categoryOptions: { value: ['Roman', 'Modern', 'Other'] },
    eraOptions: { value: ['ancient', 'modern'] },
    loadOptions: vi.fn().mockResolvedValue(undefined),
  }),
}))

// The form is tested on its own; here it fills the page's form and submits.
const BanknoteFormStub = defineComponent({
  props: { form: { type: Object, required: true } },
  emits: ['submit'],
  setup(props, { emit, expose }) {
    expose({ get obverseFile() { return mocks.front }, reverseFile: null, cardFile: null, removedObverseId: null, removedReverseId: null })
    return () => h('button', {
      'data-testid': 'submit',
      onClick: () => {
        Object.assign(props.form, { name: '1000 escudos 1996', serialNumber: 'CH 1234567', certifiedGrade: 'PMG 66 EPQ', references: [{ catalog: 'Pick', number: '188c' }] })
        emit('submit')
      },
    }, 'Guardar')
  },
})

const CatalogStub = defineComponent({
  props: { category: { type: String, required: true } },
  emits: ['use'],
  setup: (props) => () => h('div', { 'data-testid': 'catalog', 'data-category': props.category }),
})

function mountPage() {
  return mount(AddBanknotePage, { global: { stubs: { BanknoteForm: BanknoteFormStub, NumistaCatalogBrowser: CatalogStub } } })
}

describe('AddBanknotePage', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    setLocale('pt-PT')
    vi.clearAllMocks()
    mocks.front = null
    mocks.createCoin.mockResolvedValue({ data: { id: 81 } })
    mocks.uploadImage.mockResolvedValue({ data: {} })
    mocks.importCoinImage.mockResolvedValue({ data: {} })
    mocks.listCatalogs.mockResolvedValue([{ catalog: 'Numista', volumeRequired: false }, { catalog: 'Pick', volumeRequired: false }])
  })

  it('saves a banknote with its own details, then its photo, and opens it', async () => {
    mocks.front = new File(['front'], 'front.jpg', { type: 'image/jpeg' })
    const wrapper = mountPage()
    await flushPromises()

    await wrapper.get('[data-testid="submit"]').trigger('click')
    await flushPromises()

    expect(mocks.createCoin).toHaveBeenCalledWith(expect.objectContaining({
      kind: 'banknote',
      name: '1000 escudos 1996',
      category: 'Modern',
      era: 'modern',
      material: 'Paper',
      serialNumber: 'CH 1234567',
      certifiedGrade: 'PMG 66 EPQ',
      presentation: 'slab',
      references: [{ catalog: 'Pick', number: '188c', volume: undefined, uri: undefined }],
    }))
    expect(mocks.uploadImage).toHaveBeenCalledWith(81, mocks.front, 'obverse', true, false)
    expect(mocks.push).toHaveBeenCalledWith('/coin/81')
  })

  it('searches banknotes in the catalogue and fills the form from the chosen type', async () => {
    const wrapper = mountPage()
    await flushPromises()
    await wrapper.get('[data-testid="banknote-mode-catalog"]').trigger('click')
    const catalog = wrapper.get('[data-testid="catalog"]')
    expect(catalog.attributes('data-category')).toBe('banknote')

    const side = { description: '', lettering: '' }
    const selection: CatalogSelection = {
      type: {
        id: 1234, url: 'https://pt.numista.com/1234', title: '1000 Escudos', issuer: { code: 'portugal', name: 'Portugal' }, rulers: [], mints: [],
        obverse: side, reverse: side, edge: side, references: [{ catalogue: 'P', number: '188c' }], photos: [], composition: 'Papel', printers: ['Valora'],
      },
      wishlist: false,
    }
    wrapper.findComponent(CatalogStub).vm.$emit('use', selection)
    await flushPromises()

    expect(wrapper.get('[data-testid="banknote-from-catalog"]').text()).toContain('1000 Escudos')
    await wrapper.get('[data-testid="submit"]').trigger('click')
    await flushPromises()
    // The form's own values (the stub's) are saved on top of the catalogue's.
    expect(mocks.createCoin).toHaveBeenCalledWith(expect.objectContaining({ kind: 'banknote', printer: 'Valora', country: 'Portugal' }))
  })

  it('says so when the banknote cannot be saved', async () => {
    mocks.createCoin.mockRejectedValue(new Error('offline'))
    const wrapper = mountPage()
    await flushPromises()
    await wrapper.get('[data-testid="submit"]').trigger('click')
    await flushPromises()

    expect(mocks.showAlert).toHaveBeenCalledWith('Não foi possível guardar a nota. Tente de novo.', { title: 'Erro' })
    expect(mocks.push).not.toHaveBeenCalled()
  })
})
