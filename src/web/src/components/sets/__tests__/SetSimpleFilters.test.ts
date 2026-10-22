import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import SetSimpleFilters from '@/components/sets/SetSimpleFilters.vue'
import { setLocale } from '@/i18n'
import { emptyFilters, filtersToCriteria, type SimpleSetFilters } from '@/utils/setFilters'

const mocks = vi.hoisted(() => ({
  getCoinCountries: vi.fn(),
  getSuggestions: vi.fn(),
  previewSmartSet: vi.fn(),
}))
vi.mock('@/api/client', () => mocks)

const countryOptions = () => [...document.querySelectorAll<HTMLElement>('[data-testid="country-select-option"]')]

// A parent keeping the filters, as v-model does.
function mountFilters(initial: SimpleSetFilters = emptyFilters()) {
  const wrapper = mount(SetSimpleFilters, {
    props: {
      modelValue: initial,
      'onUpdate:modelValue': (value: SimpleSetFilters) => wrapper.setProps({ modelValue: value }),
    },
  })
  return wrapper
}

describe('SetSimpleFilters', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    vi.useFakeTimers()
    vi.clearAllMocks()
    mocks.getCoinCountries.mockResolvedValue({ data: [{ country: 'Portugal', count: 40 }, { country: 'Alemanha', count: 9 }, { country: '', count: 2 }] })
    mocks.getSuggestions.mockResolvedValue({ data: ['1 euro', '2 euro'] })
    mocks.previewSmartSet.mockResolvedValue({ data: { coinIds: [1, 2, 3], coinCount: 3, totalValue: 4.5 } })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it("offers the collection's countries and face values, and counts the coins the filters keep", async () => {
    const wrapper = mountFilters()
    await flushPromises()

    expect(mocks.getCoinCountries).toHaveBeenCalledWith('coin')
    const menu = wrapper.get('#filterCountry')
    expect(menu.attributes('role')).toBe('combobox')
    expect(menu.text()).toBe('Qualquer')
    await menu.trigger('click')
    expect(countryOptions().map((option) => option.textContent?.trim())).toEqual(['Qualquer', '🇩🇪Alemanha', '🇵🇹Portugal'])
    await menu.trigger('keydown', { key: 'Escape' })
    expect(wrapper.findAll('#filterDenominationOptions option').map((option) => option.attributes('value'))).toEqual(['1 euro', '2 euro'])
    expect(mocks.previewSmartSet).toHaveBeenCalledWith(filtersToCriteria(emptyFilters()))
    expect(wrapper.get('[data-testid="filter-preview"]').text()).toContain('3 moedas da coleção')
  })

  it('changes the filters and counts again a moment later', async () => {
    const wrapper = mountFilters()
    await flushPromises()

    await wrapper.get('#filterCountry').trigger('click')
    countryOptions().find((option) => option.textContent?.includes('Portugal'))!.click()
    await flushPromises()
    await wrapper.get('#filterYearFrom').setValue('2002')
    await wrapper.get('#filterYearTo').setValue('2010')
    await wrapper.get('[data-testid="filter-classification-commemorative"]').trigger('click')
    await wrapper.get('#filterPresentation').setValue('coincard')
    expect(wrapper.get('[data-testid="filter-preview"]').text()).toBe('A contar as moedas…')

    mocks.previewSmartSet.mockResolvedValue({ data: { coinIds: [7], coinCount: 1, totalValue: 2 } })
    vi.advanceTimersByTime(300)
    await flushPromises()
    const filters = { ...emptyFilters(), country: 'Portugal', yearFrom: 2002, yearTo: 2010, classification: 'commemorative' as const, presentation: 'coincard' }
    expect(wrapper.props('modelValue')).toEqual(filters)
    // Counted once, after the last change.
    expect(mocks.previewSmartSet).toHaveBeenCalledTimes(2)
    expect(mocks.previewSmartSet).toHaveBeenLastCalledWith(filtersToCriteria(filters))
    expect(wrapper.get('[data-testid="filter-preview"]').text()).toContain('1 moeda da coleção')
    expect(wrapper.get('[data-testid="filter-classification-commemorative"]').attributes('aria-pressed')).toBe('true')
  })

  it('suggests face values as one types', async () => {
    const wrapper = mountFilters()
    await flushPromises()

    mocks.getSuggestions.mockResolvedValue({ data: ['2 euro'] })
    await wrapper.get('#filterDenomination').setValue('2 e')
    vi.advanceTimersByTime(300)
    await flushPromises()
    expect(mocks.getSuggestions).toHaveBeenLastCalledWith('denomination', '2 e')
    expect(wrapper.findAll('#filterDenominationOptions option').map((option) => option.attributes('value'))).toEqual(['2 euro'])
  })

  it('keeps a country chosen before, even when the collection no longer has it', async () => {
    const wrapper = mountFilters({ ...emptyFilters(), country: 'Malta' })
    await flushPromises()

    expect(wrapper.get('#filterCountry').text()).toBe('🇲🇹Malta')
  })

  it('draws the flag of a country that no longer exists', async () => {
    mocks.getCoinCountries.mockResolvedValue({ data: [{ country: 'Jugoslávia', count: 1 }] })
    const wrapper = mountFilters()
    await flushPromises()

    await wrapper.get('#filterCountry').trigger('click')
    const yugoslavia = countryOptions().find((option) => option.textContent?.includes('Jugoslávia'))!
    expect(yugoslavia.querySelector('img[data-testid="country-flag-image"]')).not.toBeNull()
    await wrapper.get('#filterCountry').trigger('keydown', { key: 'Escape' })
  })
})
