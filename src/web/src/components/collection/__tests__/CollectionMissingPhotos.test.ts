import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CollectionMissingPhotos from '../CollectionMissingPhotos.vue'

const mocks = vi.hoisted(() => ({ getCoinsWithoutPhotos: vi.fn(), addOfficialPhotosBatch: vi.fn() }))
vi.mock('@/api/client', () => mocks)

describe('CollectionMissingPhotos', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('looks for the official photos a batch at a time, and says how it went', async () => {
    mocks.getCoinsWithoutPhotos.mockResolvedValueOnce({ data: { count: 3 } }).mockResolvedValueOnce({ data: { count: 1 } })
    mocks.addOfficialPhotosBatch
      .mockResolvedValueOnce({ data: { checked: 2, withPhotos: 1, images: 2, nextAfterId: 7, done: false } })
      .mockResolvedValueOnce({ data: { checked: 1, withPhotos: 1, images: 2, nextAfterId: 9, done: true } })
    const wrapper = mount(CollectionMissingPhotos)
    await flushPromises()

    expect(wrapper.text()).toContain('3 coins without a photo')
    await wrapper.get('[data-testid="missing-photos-start"]').trigger('click')
    await flushPromises()

    expect(mocks.addOfficialPhotosBatch.mock.calls).toEqual([[0], [7]])
    expect(wrapper.get('[data-testid="missing-photos-summary"]').text()).toBe('Official photos added to 2 coins. 1 has no official photo.')
    expect(wrapper.text()).toContain('1 coin without a photo')
  })

  it('shows nothing when every coin has a photo', async () => {
    mocks.getCoinsWithoutPhotos.mockResolvedValue({ data: { count: 0 } })
    const wrapper = mount(CollectionMissingPhotos)
    await flushPromises()

    expect(wrapper.find('[data-testid="missing-photos"]').exists()).toBe(false)
  })

  it('says so when the search fails', async () => {
    mocks.getCoinsWithoutPhotos.mockResolvedValue({ data: { count: 2 } })
    mocks.addOfficialPhotosBatch.mockRejectedValue(new Error('offline'))
    const wrapper = mount(CollectionMissingPhotos)
    await flushPromises()
    await wrapper.get('[data-testid="missing-photos-start"]').trigger('click')
    await flushPromises()

    expect(wrapper.get('[role="alert"]').text()).toBe('The photos could not be looked for. Try again.')
  })
})
