import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CatalogPhoto from '../CatalogPhoto.vue'

const mocks = vi.hoisted(() => ({ loadCatalogPhoto: vi.fn() }))
vi.mock('@/utils/catalogPhotoCache', () => mocks)

const URL_A = 'https://en.numista.com/catalogue/photos/portugal/a-180.jpg'

describe('CatalogPhoto', () => {
  beforeEach(() => mocks.loadCatalogPhoto.mockReset())

  it('loads the photo straight from Numista with its credit', () => {
    const wrapper = mount(CatalogPhoto, { props: { url: URL_A, alt: 'Obverse', credit: '© Ana Example' } })

    expect(wrapper.get('img').attributes('src')).toBe(URL_A)
    expect(wrapper.get('figcaption').text()).toBe('© Ana Example')
    expect(mocks.loadCatalogPhoto).not.toHaveBeenCalled()
  })

  it('falls back to the image proxy when Numista refuses the photo', async () => {
    mocks.loadCatalogPhoto.mockResolvedValue('blob:proxied')
    const wrapper = mount(CatalogPhoto, { props: { url: URL_A, alt: 'Obverse' } })

    await wrapper.get('img').trigger('error')
    await flushPromises()

    expect(mocks.loadCatalogPhoto).toHaveBeenCalledWith(URL_A)
    expect(wrapper.get('img').attributes('src')).toBe('blob:proxied')
  })

  it('leaves the photo out when neither Numista nor the proxy can serve it', async () => {
    mocks.loadCatalogPhoto.mockResolvedValue(null)
    const wrapper = mount(CatalogPhoto, { props: { url: URL_A, alt: 'Obverse', credit: '© Ana Example' } })

    await wrapper.get('img').trigger('error')
    await flushPromises()

    expect(wrapper.find('figure').exists()).toBe(false)
    expect(wrapper.emitted('unavailable')).toHaveLength(1)
  })

  it('renders nothing without a photo', () => {
    expect(mount(CatalogPhoto, { props: { alt: 'Obverse' } }).find('figure').exists()).toBe(false)
  })
})
