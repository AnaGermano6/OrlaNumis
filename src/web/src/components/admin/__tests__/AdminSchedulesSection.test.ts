import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AdminSchedulesSection from '../AdminSchedulesSection.vue'

const mocks = vi.hoisted(() => ({
  getAvailabilityRuns: vi.fn(),
  getAvailabilityCycles: vi.fn(),
  getAvailabilityCycleDetail: vi.fn(),
  getAvailabilityRunDetail: vi.fn(),
  triggerAvailabilityCheck: vi.fn(),
  getAdminWishlistSearchAlertRuns: vi.fn(),
  getValuationRuns: vi.fn(),
  getValuationRunDetail: vi.fn(),
  triggerValuation: vi.fn(),
  cancelValuationRun: vi.fn(),
}))

vi.mock('@/api/client', () => mocks)

vi.mock('@/composables/useSafeExternalLink', () => ({
  sanitizeExternalUrl: (url: string | null | undefined) => url ?? null,
}))

describe('AdminSchedulesSection', () => {
  beforeEach(() => {
    Object.values(mocks).forEach(mock => mock.mockReset())
    mocks.getAvailabilityRuns.mockResolvedValue({ data: { runs: [], total: 0 } })
    mocks.getAvailabilityCycles.mockResolvedValue({ data: { cycles: [], total: 0, page: 1, limit: 5 } })
    mocks.getAvailabilityCycleDetail.mockResolvedValue({ data: { children: [] } })
    mocks.getAdminWishlistSearchAlertRuns.mockResolvedValue({ data: { runs: [], total: 0, page: 1, limit: 5 } })
    mocks.getValuationRuns.mockResolvedValue({ data: { runs: [], total: 0 } })
  })

  it('renders every schedule workflow and loads each run history', async () => {
    const wrapper = mount(AdminSchedulesSection, {
      props: buildProps(),
      global: {
        stubs: {
          SafeExternalLink: { template: '<a><slot /></a>' },
          RouterLink: RouterLinkStub,
        },
      },
    })
    await flushPromises()

    for (const heading of [
      'Wishlist Availability Check',
      'Wishlist Search Alerts',
      'Collection Valuation',
      'Automatic Full Backup',
    ]) {
      expect(wrapper.text()).toContain(heading)
    }

    expect(mocks.getAvailabilityRuns).toHaveBeenCalled()
    expect(mocks.getAdminWishlistSearchAlertRuns).toHaveBeenCalledWith(1, 5)
    // These schedules were removed.
    for (const removed of ['ParcelApp', 'Auction Watch Bid Digest', 'Auction Ending Alerts', 'Auction Price Alerts', 'Collection Health Snapshots', 'Coin of the Day', 'Purchase Reminder Delivery']) {
      expect(wrapper.text()).not.toContain(removed)
    }
    expect(mocks.getValuationRuns).toHaveBeenCalled()
  })

  it('keeps each wishlist run history next to its own schedule settings', async () => {
    const wrapper = mount(AdminSchedulesSection, {
      props: buildProps(),
      global: {
        stubs: {
          SafeExternalLink: { template: '<a><slot /></a>' },
          RouterLink: RouterLinkStub,
        },
      },
    })
    await flushPromises()

    const text = wrapper.text()
    const availabilitySettings = text.indexOf('Wishlist Availability Check')
    const availabilityHistory = text.indexOf('Availability Run History')
    const searchSettings = text.indexOf('Wishlist Search Alerts')
    const searchHistory = text.indexOf('Wishlist Search Alert Run History')
    const valuationSettings = text.indexOf('Collection Valuation')

    expect(availabilitySettings).toBeLessThan(availabilityHistory)
    expect(availabilityHistory).toBeLessThan(searchSettings)
    expect(searchSettings).toBeLessThan(searchHistory)
    expect(searchHistory).toBeLessThan(valuationSettings)
  })
})

function buildProps() {
  return {
    settings: {
      WishlistCheckEnabled: 'true',
      WishlistCheckStartTime: '08:00',
      WishlistCheckInterval: '10080',
      WishlistSearchAlertsCheckEnabled: 'true',
      WishlistSearchAlertsCheckStartTime: '08:00',
    },
    settingsSaving: false,
    availSettingsMsg: '',
    availSettingsError: false,
    valSettingsMsg: '',
    valSettingsError: false,
  }
}
