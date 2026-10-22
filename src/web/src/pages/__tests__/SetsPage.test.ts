import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import SetsPage from '@/pages/SetsPage.vue'

const mockGetSets = vi.fn()
const mockCreateSetApi = vi.fn()
const mockCreateSetFromCsv = vi.fn()
const mockShowAlert = vi.fn()
const mockPush = vi.fn()

vi.mock('@/api/client', () => ({
  getApiErrorMessage: (error: unknown) => {
    const err = error as { response?: { data?: { error?: string } } }
    return err.response?.data?.error ?? ''
  },
  getSets: (...args: unknown[]) => mockGetSets(...args),
  createSet: (...args: unknown[]) => mockCreateSetApi(...args),
  createSetFromCsv: (...args: unknown[]) => mockCreateSetFromCsv(...args),
  getJointIssues: () => Promise.resolve({ data: { issues: [] } }),
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush }),
  RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' },
}))

vi.mock('@/composables/usePwa', () => ({
  usePwa: () => ({ isPwa: false }),
}))

vi.mock('@/composables/useDialog', () => ({
  useDialog: () => ({
    showAlert: (...args: unknown[]) => mockShowAlert(...args),
    showConfirm: vi.fn(),
  }),
}))

describe('SetsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetSets.mockResolvedValue({ data: { sets: [] } })
  })

  it('links to the catalogue of commemorative coins', async () => {
    const wrapper = mount(SetsPage, { global: { stubs: { Teleport: true } } })
    await flushPromises()

    const link = wrapper.get('[data-testid="sets-catalogue-link"]')
    expect(link.attributes('href')).toBe('/sets/commemoratives')
    expect(link.text()).toContain('Commemorative catalogue')
  })

  it('shows how far a checklist set is on its card', async () => {
    mockGetSets.mockResolvedValue({ data: { sets: [
      { id: 3, name: 'Portugal · 2 €', color: '#c9a84c', setType: 'checklist', coinCount: 12, totalValue: 30, completionPercentage: 33.3 },
    ] } })
    const wrapper = mount(SetsPage, { global: { stubs: { Teleport: true } } })
    await flushPromises()

    expect(wrapper.text()).toContain('Checklist · 33% complete')
    expect(wrapper.text()).toContain('12')
  })

  it('shows the places of the sets country by country, and remembers the choice', async () => {
    localStorage.removeItem('sets:view')
    mockGetSets.mockResolvedValue({ data: { sets: [
      { id: 3, name: 'Portugal · 2 €', color: '#c9a84c', setType: 'checklist', coinCount: 12, totalValue: 30, completionPercentage: 33.3, countries: [{ country: 'PT', owned: 12, total: 36 }] },
      { id: 4, name: 'Mine', color: '#c9a84c', setType: 'standard', coinCount: 2, totalValue: 0 },
    ] } })
    const wrapper = mount(SetsPage, { global: { stubs: { Teleport: true } } })
    await flushPromises()

    expect(wrapper.get('[data-testid="sets-view-sets"]').attributes('aria-pressed')).toBe('true')
    expect(wrapper.find('[data-testid="sets-by-country"]').exists()).toBe(false)
    await wrapper.get('[data-testid="sets-view-countries"]').trigger('click')
    expect(wrapper.get('[data-testid="sets-country-PT"]').text()).toContain('24 missing')
    expect(localStorage.getItem('sets:view')).toBe('countries')

    const again = mount(SetsPage, { global: { stubs: { Teleport: true } } })
    await flushPromises()
    expect(again.find('[data-testid="sets-by-country"]').exists()).toBe(true)
    localStorage.removeItem('sets:view')
  })

  it('offers the view by country only when a set has places by country', async () => {
    mockGetSets.mockResolvedValue({ data: { sets: [{ id: 4, name: 'Mine', color: '#c9a84c', setType: 'standard', coinCount: 2, totalValue: 0 }] } })
    const wrapper = mount(SetsPage, { global: { stubs: { Teleport: true } } })
    await flushPromises()
    expect(wrapper.find('[data-testid="sets-view"]').exists()).toBe(false)
  })

  it('creates a challenge set and opens it', async () => {
    mockCreateSetApi.mockResolvedValue({ data: { id: 42 } })
    const wrapper = mount(SetsPage, { global: { stubs: { Teleport: true } } })
    await flushPromises()

    await wrapper.find('.btn-primary').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="set-path-challenge"]').trigger('click')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(mockCreateSetApi).toHaveBeenCalledWith(expect.objectContaining({
      name: 'A coin of each euro country',
      setType: 'checklist',
      checklist: { source: 'challenge', challenge: 'euro-countries' },
    }))
    expect(mockPush).toHaveBeenCalledWith({ name: 'set-detail', params: { id: 42 } })
    expect(wrapper.find('.card').exists()).toBe(false)
  })

  it('shows the server error when a set cannot be created', async () => {
    mockCreateSetApi.mockRejectedValue({ response: { data: { error: 'Name is required' } } })
    const wrapper = mount(SetsPage, { global: { stubs: { Teleport: true } } })
    await flushPromises()

    await wrapper.find('.btn-primary').trigger('click')
    await flushPromises()
    await wrapper.get('[data-testid="set-path-manual"]').trigger('click')
    await wrapper.find('#setName').setValue('Silver Quarters')
    await wrapper.find('form').trigger('submit.prevent')
    await flushPromises()

    expect(mockShowAlert).toHaveBeenCalledWith('Name is required', { title: 'Failed to Create Set' })
    expect(mockPush).not.toHaveBeenCalled()
  })
})
