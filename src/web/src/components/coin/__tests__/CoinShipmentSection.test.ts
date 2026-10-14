import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CoinShipmentSection from '../CoinShipmentSection.vue'
import { getCoinShipment, upsertCoinShipment } from '@/api/client'
import type { Shipment } from '@/types'

vi.mock('@/api/client', () => ({
  getCoinShipment: vi.fn(),
  upsertCoinShipment: vi.fn(),
  deleteCoinShipment: vi.fn(),
  setCoinShipmentManualOverride: vi.fn(),
}))

function makeShipment(overrides: Partial<Shipment> = {}): Shipment {
  return {
    id: 1,
    userId: 1,
    coinId: 42,
    carrier: 'other',
    manualCarrierName: 'CTT',
    trackingNumber: 'TRK123',
    currentStatus: 'in_transit',
    currentStatusSource: 'manual',
    notes: '',
    manualOverrideEnabled: true,
    manualOverrideStatus: 'in_transit',
    manualOverrideNote: '',
    manualOverrideUpdatedAt: null,
    lastSyncedAt: null,
    lastSyncError: '',
    estimatedDeliveryAt: null,
    deliveredAt: null,
    events: [],
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

describe('CoinShipmentSection', () => {
  beforeEach(() => {
    vi.mocked(getCoinShipment).mockReset()
    vi.mocked(upsertCoinShipment).mockReset()
  })

  it("shows a shipment by its carrier's name, with no ParcelApp check", async () => {
    vi.mocked(getCoinShipment).mockResolvedValue({ data: { shipment: makeShipment(), trackingUrl: '' } } as Awaited<ReturnType<typeof getCoinShipment>>)
    const wrapper = mount(CoinShipmentSection, { props: { coinId: 42 } })
    await flushPromises()

    expect(wrapper.text()).toContain('CTT')
    expect(wrapper.text()).toContain('TRK123')
    expect(wrapper.text()).not.toContain('ParcelApp')
  })

  it("records a new shipment with its carrier's name", async () => {
    vi.mocked(getCoinShipment).mockRejectedValue({ response: { data: { error: 'Shipment not found' } } })
    vi.mocked(upsertCoinShipment).mockResolvedValue({ data: { shipment: makeShipment(), trackingUrl: '' } } as Awaited<ReturnType<typeof upsertCoinShipment>>)
    const wrapper = mount(CoinShipmentSection, { props: { coinId: 42 } })
    await flushPromises()

    const save = () => wrapper.findAll('button').find((button) => button.text() === 'Save Shipment')!.trigger('click')
    const inputs = wrapper.findAll('input')
    await inputs[1]!.setValue('TRK123')
    await save()
    expect(upsertCoinShipment).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('The carrier is required')

    await inputs[0]!.setValue(' CTT ')
    await save()
    await flushPromises()
    expect(upsertCoinShipment).toHaveBeenCalledWith(42, { carrier: 'other', trackingNumber: 'TRK123', notes: '', manualCarrierName: 'CTT' })
    expect(wrapper.text()).toContain('Shipment saved')
  })
})
