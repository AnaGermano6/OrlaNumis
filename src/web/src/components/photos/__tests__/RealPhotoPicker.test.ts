import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RealPhotoPicker from '../RealPhotoPicker.vue'
import { setLocale } from '@/i18n'
import type { CoinPhotoCandidate } from '@/api/client'
import type { RealPhotoChoice } from '@/utils/coinPhotos'

const mocks = vi.hoisted(() => ({ searchCoinPhotos: vi.fn(), loadCatalogPhoto: vi.fn() }))
vi.mock('@/api/client', () => ({ searchCoinPhotos: mocks.searchCoinPhotos }))
vi.mock('@/utils/catalogPhotoCache', () => ({ loadCatalogPhoto: mocks.loadCatalogPhoto }))

const COMMONS: CoinPhotoCandidate = {
  source: 'wikimedia',
  thumbnail: 'https://upload.wikimedia.org/wikipedia/commons/thumb/b/bb/Obv.jpg/400px-Obv.jpg',
  imageUrl: 'https://upload.wikimedia.org/wikipedia/commons/b/bb/Obv.jpg',
  pageUrl: 'https://commons.wikimedia.org/wiki/File:Obv.jpg',
  title: '50 Escudos 1968 obverse',
  author: 'Jane Doe',
  license: 'CC BY-SA 4.0',
}
const NUMISTA: CoinPhotoCandidate = {
  source: 'numista', side: 'obverse',
  thumbnail: 'https://en.numista.com/catalogue/photos/portugal/1-180.jpg',
  imageUrl: 'https://en.numista.com/catalogue/photos/portugal/1-original.jpg',
}

describe('RealPhotoPicker', () => {
  beforeEach(() => {
    mocks.searchCoinPhotos.mockReset()
    mocks.loadCatalogPhoto.mockReset()
    mocks.loadCatalogPhoto.mockResolvedValue('blob:photo')
    setLocale('en')
  })

  it('searches on open and lets each photo be picked for one side', async () => {
    mocks.searchCoinPhotos.mockResolvedValue({ data: { status: 'ok', query: '50 Escudos 1968', photos: [NUMISTA, COMMONS] } })
    const updates: RealPhotoChoice[] = []
    const wrapper = mount(RealPhotoPicker, {
      props: { initialQuery: '50 Escudos 1968', numistaTypeId: 4930, modelValue: {}, 'onUpdate:modelValue': (value: RealPhotoChoice) => { updates.push(value) } },
    })
    await flushPromises()

    expect(mocks.searchCoinPhotos).toHaveBeenCalledWith({ q: '50 Escudos 1968', numistaTypeId: 4930, lang: 'en' })
    const cards = wrapper.findAll('li')
    expect(cards).toHaveLength(2)
    expect(cards[1]!.text()).toContain('© Jane Doe · CC BY-SA 4.0 · Wikimedia Commons')
    expect(mocks.loadCatalogPhoto).toHaveBeenCalledWith(COMMONS.thumbnail)

    await cards[1]!.findAll('button')[0]!.trigger('click')
    expect(updates.at(-1)).toEqual({ obverse: COMMONS })

    await wrapper.setProps({ modelValue: { obverse: COMMONS } })
    expect(wrapper.findAll('li')[1]!.findAll('button')[0]!.attributes('aria-pressed')).toBe('true')

    await wrapper.findAll('li')[1]!.findAll('button')[1]!.trigger('click')
    expect(updates.at(-1)).toEqual({ reverse: COMMONS })
  })

  it('drops photos that cannot be shown and unselects them', async () => {
    mocks.searchCoinPhotos.mockResolvedValue({ data: { status: 'ok', query: 'q', photos: [COMMONS] } })
    mocks.loadCatalogPhoto.mockResolvedValue(null)
    const updates: RealPhotoChoice[] = []
    const wrapper = mount(RealPhotoPicker, {
      props: { initialQuery: 'q', modelValue: { reverse: COMMONS }, 'onUpdate:modelValue': (value: RealPhotoChoice) => { updates.push(value) } },
    })
    await flushPromises()

    expect(wrapper.findAll('li')).toHaveLength(0)
    expect(updates.at(-1)).toEqual({})
    expect(wrapper.text()).toContain('No photos found')
  })

  it('searches again with edited words and reports failures', async () => {
    mocks.searchCoinPhotos.mockResolvedValueOnce({ data: { status: 'ok', query: 'q', photos: [] } })
    mocks.searchCoinPhotos.mockRejectedValueOnce(new Error('offline'))
    const wrapper = mount(RealPhotoPicker, { props: { initialQuery: 'q', modelValue: {} } })
    await flushPromises()

    await wrapper.get('input').setValue('denarius augustus')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(mocks.searchCoinPhotos).toHaveBeenLastCalledWith({ q: 'denarius augustus', numistaTypeId: undefined, lang: 'en' })
    expect(wrapper.text()).toContain('Photos could not be searched right now.')
  })
})
