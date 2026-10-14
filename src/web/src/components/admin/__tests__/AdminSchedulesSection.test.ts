import { flushPromises, mount } from '@vue/test-utils'
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
  triggerCollectionHealthSnapshots: vi.fn(),
  getCollectionHealthSnapshotRuns: vi.fn(),
  getCollectionHealthSnapshotStatus: vi.fn(),
  triggerCoinOfDayRun: vi.fn(),
  getCoinOfDayRuns: vi.fn(),
  getCoinOfDayRunDetail: vi.fn(),
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
    mocks.getCoinOfDayRuns.mockResolvedValue({ data: { runs: [], total: 0 } })
    mocks.getCollectionHealthSnapshotRuns.mockResolvedValue({ data: { runs: [], total: 0 } })
    mocks.getCollectionHealthSnapshotStatus.mockResolvedValue({
      data: { name: 'collection-health', enabled: true, isRunning: false, nextRunIn: 49320000000000 },
    })
    mocks.getCoinOfDayRunDetail.mockResolvedValue({ data: { id: 1, status: 'queued', picked: 0, skipped: 0, errors: 0 } })
  })

  it('renders every schedule workflow and loads each run history', async () => {
    const wrapper = mount(AdminSchedulesSection, {
      props: buildProps(),
      global: {
        stubs: {
          SafeExternalLink: { template: '<a><slot /></a>' },
        },
      },
    })
    await flushPromises()

    for (const heading of [
      'Wishlist Availability Check',
      'Wishlist Search Alerts',
      'Collection Valuation',
      'Collection Health Snapshots',
      'Coin of the Day',
      'Purchase Reminder Delivery',
    ]) {
      expect(wrapper.text()).toContain(heading)
    }

    expect(mocks.getAvailabilityRuns).toHaveBeenCalled()
    expect(mocks.getAdminWishlistSearchAlertRuns).toHaveBeenCalledWith(1, 5)
    // ParcelApp tracking, the watch-bid digest and the auction alerts were removed.
    for (const removed of ['ParcelApp', 'Auction Watch Bid Digest', 'Auction Ending Alerts', 'Auction Price Alerts']) {
      expect(wrapper.text()).not.toContain(removed)
    }
    expect(mocks.getValuationRuns).toHaveBeenCalled()
    expect(mocks.getCollectionHealthSnapshotRuns).toHaveBeenCalled()
    expect(mocks.getCoinOfDayRuns).toHaveBeenCalled()
  })

  it('keeps each wishlist run history next to its own schedule settings', async () => {
    const wrapper = mount(AdminSchedulesSection, {
      props: buildProps(),
      global: {
        stubs: {
          SafeExternalLink: { template: '<a><slot /></a>' },
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

  it('shows collection health snapshot run history and triggers a manual run', async () => {
    mocks.getCollectionHealthSnapshotRuns.mockResolvedValue({
      data: {
        runs: [{
          id: 5,
          triggerType: 'scheduled',
          status: 'success',
          usersEligible: 3,
          usersSnapshotted: 3,
          usersFailed: 0,
          durationMs: 800,
          startedAt: '2026-07-16T04:30:00Z',
          completedAt: '2026-07-16T04:30:01Z',
          createdAt: '2026-07-16T04:30:00Z',
        }],
        total: 1,
      },
    })
    mocks.triggerCollectionHealthSnapshots.mockResolvedValue({
      data: { message: 'Snapshot run complete', users: 3, snapshotsCreated: 3, skipped: 0, errors: 0, durationMs: 900 },
    })

    const wrapper = mount(AdminSchedulesSection, {
      props: buildProps(),
      global: {
        stubs: {
          SafeExternalLink: { template: '<a><slot /></a>' },
        },
      },
    })
    await flushPromises()

    expect(wrapper.text()).toContain('Collection Health Snapshot Run History')
    expect(mocks.getCollectionHealthSnapshotRuns).toHaveBeenCalled()
    expect(mocks.getCollectionHealthSnapshotStatus).toHaveBeenCalled()
    expect(wrapper.text()).toContain('Server status:')
    expect(wrapper.text()).toContain('Enabled')
    expect(wrapper.text()).toContain('13h 42m')

    // Run Now buttons in template order: Availability(0), Valuation(1),
    // Health(2), CoinOfDay(3).
    const runButtons = wrapper.findAll('button').filter(button => button.text() === 'Run Now')
    await runButtons[2]?.trigger('click')
    await flushPromises()

    expect(mocks.triggerCollectionHealthSnapshots).toHaveBeenCalledTimes(1)
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
    healthSettingsMsg: '',
    healthSettingsError: false,
    valSettingsMsg: '',
    valSettingsError: false,
  }
}
