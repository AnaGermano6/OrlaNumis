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

vi.mock('@/composables/useLayout', async () => {
  const { ref } = await import('vue')
  return { useLayout: () => ({ isMobile: ref(false) }) }
})

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

  it('has no card for the catalogue of commemorative coins (the menu and the statistics open it)', async () => {
    const wrapper = mount(SetsPage, { global: { stubs: { Teleport: true } } })
    await flushPromises()

    expect(wrapper.find('[data-testid="sets-catalogue-link"]').exists()).toBe(false)
    expect(wrapper.find('a[href="/sets/commemoratives"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('Commemorative catalogue')
  })

  it('shows how far a checklist set is on its card', async () => {
    mockGetSets.mockResolvedValue({ data: { sets: [
      { id: 3, name: 'Portugal · 2 €', color: '#c9a84c', setType: 'checklist', coinCount: 12, totalValue: 30, completionPercentage: 33.3 },
    ] } })
    const wrapper = mount(SetsPage, { global: { stubs: { Teleport: true } } })
    await flushPromises()

    expect(wrapper.get('[data-testid="set-card-progress"]').text()).toContain('33%')
    expect(wrapper.get('[data-testid="set-card-count"]').text()).toContain('12')
  })

  it('shows all the sets, two in a row on a wide screen, or those to complete, or the complete ones', async () => {
    mockGetSets.mockResolvedValue({ data: { sets: [
      { id: 1, name: 'Done', color: '#c9a84c', setType: 'checklist', coinCount: 3, total: 3, totalValue: 6, completionPercentage: 100 },
      { id: 2, name: 'Halfway', color: '#c9a84c', setType: 'checklist', coinCount: 1, total: 2, totalValue: 2, completionPercentage: 50 },
      { id: 3, name: 'Mine', color: '#c9a84c', setType: 'standard', coinCount: 4, totalValue: 8 },
    ] } })
    const wrapper = mount(SetsPage, { global: { stubs: { Teleport: true } } })
    await flushPromises()

    const titles = () => wrapper.findAll('[data-testid="set-card-title"]').map((title) => title.text())
    expect(wrapper.get('[data-testid="sets-cards"]').classes()).toContain('lg:grid-cols-2')
    const status = wrapper.get('[data-testid="sets-status"]')
    expect(status.findAll('option').map((option) => option.text())).toEqual(['All', 'To complete', 'Complete'])
    expect(titles()).toEqual(['Done', 'Halfway', 'Mine'])
    await status.setValue('incomplete')
    expect(titles()).toEqual(['Halfway', 'Mine'])
    await status.setValue('complete')
    expect(titles()).toEqual(['Done'])

    mockGetSets.mockResolvedValue({ data: { sets: [{ id: 3, name: 'Mine', color: '#c9a84c', setType: 'standard', coinCount: 4, totalValue: 8 }] } })
    const none = mount(SetsPage, { global: { stubs: { Teleport: true } } })
    await flushPromises()
    await none.get('[data-testid="sets-status"]').setValue('complete')
    expect(none.get('[data-testid="sets-status-none"]').text()).toBe('No set matches this choice.')
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
    expect(wrapper.find('[data-testid="sets-country-search"]').exists()).toBe(false)
    await wrapper.get('[data-testid="sets-view-countries"]').trigger('click')
    expect(wrapper.get('[data-testid="sets-country-PT"]').text()).toContain('24 missing')
    expect(localStorage.getItem('sets:view')).toBe('countries')
    // By country: a country searched for and their order, in place of the sets' choice.
    expect(wrapper.find('[data-testid="sets-status"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="sets-country-sort"]').findAll('option').map((option) => option.text())).toEqual(['Most progress', 'Least progress', 'Most missing', 'Name'])
    await wrapper.get('[data-testid="sets-country-search"] input').setValue('spain')
    expect(wrapper.find('[data-testid="sets-country-PT"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="sets-by-country-no-match"]').text()).toBe('No country matches the search.')

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
