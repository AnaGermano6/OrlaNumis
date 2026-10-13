import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import SettingsCurrencySetting from '../SettingsCurrencySetting.vue'
import { collectionCurrency, setCollectionCurrency } from '@/utils/collectionCurrency'

const mocks = vi.hoisted(() => ({ changeCollectionCurrency: vi.fn(), auth: { isAdmin: true } }))
vi.mock('@/api/client', () => ({ changeCollectionCurrency: mocks.changeCollectionCurrency }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => mocks.auth }))

describe('SettingsCurrencySetting', () => {
  beforeEach(() => {
    mocks.auth.isAdmin = true
    mocks.changeCollectionCurrency.mockReset()
    vi.spyOn(window, 'confirm').mockReturnValue(true)
  })
  afterEach(() => {
    setCollectionCurrency('USD')
    vi.restoreAllMocks()
  })

  it('switches the collection to euros after confirmation', async () => {
    mocks.changeCollectionCurrency.mockResolvedValue({ data: { currency: 'EUR', previous: 'USD', coinsConverted: 1, factor: 0.8929, rateDate: '2026-10-09' } })
    const wrapper = mount(SettingsCurrencySetting)
    const button = wrapper.get('button[type="submit"]')
    expect(button.attributes('disabled')).toBeDefined()

    await wrapper.get('[data-testid="collection-currency"]').setValue('EUR')
    expect(wrapper.text()).toContain("today's ECB rate")
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(window.confirm).toHaveBeenCalled()
    expect(mocks.changeCollectionCurrency).toHaveBeenCalledWith('EUR')
    expect(collectionCurrency.value).toBe('EUR')
    expect(wrapper.get('[role="status"]').text()).toBe('Done: 1 coin converted (1 USD = 0.8929 EUR, ECB rates of 10/9/2026).')
  })

  it('offers the other currencies of the ECB rates, and converts between any two', async () => {
    setCollectionCurrency('EUR')
    mocks.changeCollectionCurrency.mockResolvedValue({ data: { currency: 'GBP', previous: 'EUR', coinsConverted: 3, factor: 0.8625, rateDate: '2026-10-09' } })
    const wrapper = mount(SettingsCurrencySetting)
    const options = wrapper.findAll('[data-testid="collection-currency"] option').map((option) => option.attributes('value'))
    expect(options.slice(0, 4)).toEqual(['EUR', 'USD', 'GBP', 'CHF'])
    expect(options).toEqual(expect.arrayContaining(['JPY', 'BRL', 'SEK', 'PLN']))

    await wrapper.get('[data-testid="collection-currency"]').setValue('GBP')
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(mocks.changeCollectionCurrency).toHaveBeenCalledWith('GBP')
    expect(collectionCurrency.value).toBe('GBP')
    expect(wrapper.get('[role="status"]').text()).toContain('3 coins converted (1 EUR = 0.8625 GBP')
  })

  it('changes nothing when not confirmed, and reports a failure', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    const wrapper = mount(SettingsCurrencySetting)
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

  it('shows the currency, without changing it, to a user who is not an administrator', () => {
    setCollectionCurrency('EUR')
    mocks.auth.isAdmin = false
    const wrapper = mount(SettingsCurrencySetting)

    expect(wrapper.find('form').exists()).toBe(false)
    expect(wrapper.text()).toContain('Euro (€)')
    expect(wrapper.text()).toContain('Only an administrator can change it.')
  })
})
