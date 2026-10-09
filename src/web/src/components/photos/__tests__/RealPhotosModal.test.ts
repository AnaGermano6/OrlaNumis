import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import RealPhotosModal from '../RealPhotosModal.vue'
import type { CoinPhotoCandidate } from '@/api/client'

const mocks = vi.hoisted(() => ({ importCoinImage: vi.fn() }))
vi.mock('@/api/client', () => ({ importCoinImage: mocks.importCoinImage }))

const PickerStub = defineComponent({
  name: 'RealPhotoPicker',
  props: { modelValue: Object, initialQuery: String, numistaTypeId: Number },
  emits: ['update:modelValue'],
  template: '<div />',
})

const OBVERSE: CoinPhotoCandidate = { source: 'wikimedia', thumbnail: 't1', imageUrl: 'https://upload.wikimedia.org/wikipedia/commons/a/aa/O.jpg', author: 'Jane', license: 'CC0', pageUrl: 'https://commons.wikimedia.org/wiki/File:O.jpg' }
const REVERSE: CoinPhotoCandidate = { source: 'numista', thumbnail: 't2', imageUrl: 'https://en.numista.com/catalogue/photos/x/r.jpg' }

function mountModal(hasPrimaryImage = false) {
  return mount(RealPhotosModal, {
    props: { coinId: 9, coinName: '50 Escudos 1968', numistaTypeId: 4930, hasPrimaryImage },
    global: { stubs: { RealPhotoPicker: PickerStub } },
  })
}

function saveButton(wrapper: ReturnType<typeof mountModal>) {
  return wrapper.findAll('button').find((button) => button.text() === 'Save photos')!
}

describe('RealPhotosModal', () => {
  beforeEach(() => mocks.importCoinImage.mockReset())

  it('imports the chosen sides with credit and closes', async () => {
    mocks.importCoinImage.mockResolvedValue({ data: {} })
    const wrapper = mountModal()
    expect(wrapper.getComponent(PickerStub).props('initialQuery')).toBe('50 Escudos 1968')
    expect(saveButton(wrapper).attributes('disabled')).toBeDefined()

    wrapper.getComponent(PickerStub).vm.$emit('update:modelValue', { obverse: OBVERSE, reverse: REVERSE })
    await flushPromises()
    await saveButton(wrapper).trigger('click')
    await flushPromises()

    expect(mocks.importCoinImage).toHaveBeenNthCalledWith(1, 9, {
      url: OBVERSE.imageUrl, imageType: 'obverse', isPrimary: true,
      credit: 'Obverse photo: Jane, CC0, Wikimedia Commons – https://commons.wikimedia.org/wiki/File:O.jpg',
    })
    expect(mocks.importCoinImage).toHaveBeenNthCalledWith(2, 9, {
      url: REVERSE.imageUrl, imageType: 'reverse', isPrimary: false, credit: 'Reverse photo: Numista',
    })
    expect(wrapper.emitted('saved')).toHaveLength(1)
    expect(wrapper.emitted('close')).toHaveLength(1)
  })

  it('keeps the coin primary image and stays open when a source refuses a photo', async () => {
    mocks.importCoinImage.mockResolvedValueOnce({ data: {} }).mockRejectedValueOnce(new Error('refused'))
    const wrapper = mountModal(true)

    wrapper.getComponent(PickerStub).vm.$emit('update:modelValue', { obverse: OBVERSE, reverse: REVERSE })
    await flushPromises()
    await saveButton(wrapper).trigger('click')
    await flushPromises()

    expect(mocks.importCoinImage.mock.calls[0]![1].isPrimary).toBe(false)
    expect(wrapper.emitted('saved')).toHaveLength(1)
    expect(wrapper.emitted('close')).toBeUndefined()
    expect(wrapper.get('[role="alert"]').text()).toContain('The source refused the photo (Reverse)')
  })
})
