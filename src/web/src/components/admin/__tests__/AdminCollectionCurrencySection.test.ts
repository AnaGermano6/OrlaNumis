import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import AdminCollectionCurrencySection from '../AdminCollectionCurrencySection.vue'
import { collectionCurrency, setCollectionCurrency } from '@/utils/collectionCurrency'

const mocks = vi.hoisted(() => ({ changeCollectionCurrency: vi.fn() }))
vi.mock('@/api/client', () => mocks)

describe('AdminCollectionCurrencySection', () => {
  beforeEach(() => {
    mocks.changeCollectionCurrency.mockReset()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })
  afterEach(() => {
    setCollectionCurrency('USD')
    vi.restoreAllMocks()
  })

  it('switches the collection to euros after confirmation', async () => {
    mocks.changeCollectionCurrency.mockResolvedValue({ data: { currency: 'EUR', previous: 'USD', coinsConverted: 1, exchangeRate: { usdPerEur: 1.12, date: '2026-10-09' } } })
    const wrapper = mount(AdminCollectionCurrencySection)
    const button = wrapper.get('button[type="submit"]')
    expect(button.attributes('disabled')).toBeDefined()

    await wrapper.get('[data-testid="collection-currency"]').setValue('EUR')
    expect(wrapper.text()).toContain("today's ECB rate")
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(window.confirm).toHaveBeenCalled()
    expect(mocks.changeCollectionCurrency).toHaveBeenCalledWith('EUR')
    expect(collectionCurrency.value).toBe('EUR')
    expect(wrapper.get('[role="status"]').text()).toBe('Done: 1 coin converted (1 € = 1.12 US$).')
  })

  it('changes nothing when not confirmed, and reports a failure', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    const wrapper = mount(AdminCollectionCurrencySection)
    await wrapper.get('[data-testid="collection-currency"]').setValue('EUR')
    await wrapper.get('form').trigger('submit')
    expect(mocks.changeCollectionCurrency).not.toHaveBeenCalled()

    vi.spyOn(window, 'confirm').mockReturnValue(true)
    mocks.changeCollectionCurrency.mockRejectedValue(new Error('502'))
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(collectionCurrency.value).toBe('USD')
    expect(wrapper.get('[role="status"]').text()).toContain('Nothing was converted')
  })
})
