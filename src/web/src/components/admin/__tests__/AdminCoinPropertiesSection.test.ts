import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AdminCoinPropertiesSection from '../AdminCoinPropertiesSection.vue'

const mocks = vi.hoisted(() => ({
  getMintLocations: vi.fn(),
  adminCreateMintLocation: vi.fn(),
  adminUpdateMintLocation: vi.fn(),
  adminDeleteMintLocation: vi.fn(),
}))

vi.mock('@/api/client', () => mocks)

describe('AdminCoinPropertiesSection - Portuguese names', () => {
  beforeEach(() => {
    Object.values(mocks).forEach(mock => mock.mockReset())
    mocks.getMintLocations.mockResolvedValue({ data: [] })
  })

  it('edits per-option Portuguese names as Value=Name lines for configured options only', async () => {
    const wrapper = mount(AdminCoinPropertiesSection, {
      props: {
        categoryOptions: 'Roman\nCeltic',
        eraOptions: 'ancient',
        categoryLabelsPt: 'Celtic=Celta\nRemoved=Removida',
        eraLabelsPt: '',
        saving: false,
        msg: '',
        error: false,
      },
    })
    await flushPromises()

    const roman = wrapper.get('input[aria-label="Portuguese name for Roman"]')
    const celtic = wrapper.get('input[aria-label="Portuguese name for Celtic"]')
    expect(roman.attributes('placeholder')).toBe('Romana')
    expect((celtic.element as HTMLInputElement).value).toBe('Celta')
    expect(wrapper.get('input[aria-label="Portuguese name for ancient"]').attributes('placeholder')).toBe('antiga')

    await roman.setValue('Romana imperial')

    const updates = wrapper.emitted('update:categoryLabelsPt') ?? []
    expect(updates.at(-1)).toEqual(['Roman=Romana imperial\nCeltic=Celta'])
  })
})
