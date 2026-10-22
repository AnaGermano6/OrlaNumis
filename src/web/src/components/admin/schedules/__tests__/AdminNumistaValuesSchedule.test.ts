import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import AdminNumistaValuesSchedule from '../AdminNumistaValuesSchedule.vue'
import type { NumistaValueStatus } from '@/types'
import { setLocale } from '@/i18n'

const mocks = vi.hoisted(() => ({
  getNumistaValuesStatus: vi.fn(),
  runNumistaValuesNow: vi.fn(),
}))

vi.mock('@/api/client', () => mocks)

function status(overrides: Partial<NumistaValueStatus> = {}): NumistaValueStatus {
  return {
    enabled: true,
    isRunning: false,
    nextRunIn: '5h12m0s',
    usage: { month: '2026-11', requests: 400, limit: 2000, pauseAt: 1800, backgroundPaused: false },
    quotaLevel: 'ok',
    resumesOn: '2026-12-01',
    dailyBudget: 40,
    eligible: 192,
    due: 150,
    withoutGrade: 269,
    currency: 'EUR',
    runs: [
      { id: 2, triggerType: 'scheduled', status: 'budget_spent', startedAt: '2026-11-02T05:00:00Z', completedAt: '2026-11-02T05:01:00Z', budget: 40, requests: 40, checked: 25, updated: 20, unchanged: 3, noPrice: 2, failed: 0 },
      { id: 1, triggerType: 'manual', status: 'unconfigured', startedAt: '2026-11-01T09:00:00Z', completedAt: '2026-11-01T09:00:01Z', budget: 40, requests: 1, checked: 0, updated: 0, unchanged: 0, noPrice: 0, failed: 0 },
    ],
    ...overrides,
  }
}

function mountCard(settings: Record<string, string> = { NumistaValuesEnabled: 'true', NumistaValuesStartTime: '05:00', NumistaValuesDailyRequests: '40' }) {
  return mount(AdminNumistaValuesSchedule, {
    props: { settings, settingsSaving: false },
    global: { stubs: { BaseToggle: true } },
  })
}

describe('AdminNumistaValuesSchedule', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useRealTimers()
    setLocale('pt-PT')
    mocks.getNumistaValuesStatus.mockResolvedValue({ data: status() })
    mocks.runNumistaValuesNow.mockResolvedValue({ data: { message: 'Batch started' } })
  })

  afterEach(() => setLocale('en'))

  it('shows the quota, the next batch, the coins covered and the latest batches', async () => {
    const wrapper = mountCard()
    await flushPromises()

    const box = wrapper.get('[data-testid="numista-values-status"]').text()
    expect(box).toContain('400 de 2000 pedidos')
    expect(box).toContain('Os lotes param sozinhos aos 1800 pedidos')
    expect(box).toContain('Próximo lote: até 40 pedidos · daqui a 5 h 12 min')
    expect(box).toContain('192 moedas abrangidas · 150 por verificar')
    expect(box).toContain('269 moedas ligadas ao Numista não têm um grau com preço')
    const runs = wrapper.get('[data-testid="numista-values-runs"]').text()
    expect(runs).toContain('orçamento do dia gasto')
    expect(runs).toContain('sem chave do Numista')
    expect(runs).toContain('40 / 40')
  })

  it('warns when the quota runs low and when it is spent', async () => {
    mocks.getNumistaValuesStatus.mockResolvedValue({ data: status({ quotaLevel: 'low', usage: { month: '2026-11', requests: 1600, limit: 2000, pauseAt: 1800, backgroundPaused: false } }) })
    const low = mountCard()
    await flushPromises()
    expect(low.get('[data-testid="numista-values-quota-message"]').text()).toBe('A quota está a acabar: os lotes ficam mais pequenos e param aos 1800 pedidos.')

    mocks.getNumistaValuesStatus.mockResolvedValue({
      data: status({ quotaLevel: 'paused', dailyBudget: 0, usage: { month: '2026-11', requests: 1810, limit: 2000, pauseAt: 1800, backgroundPaused: true } }),
    })
    const paused = mountCard()
    await flushPromises()
    expect(paused.get('[data-testid="numista-values-quota-message"]').text()).toContain('voltam a ser atualizados a partir de 1 de dezembro')
    expect(paused.text()).not.toContain('Próximo lote')
  })

  it('says when the daily batch is off', async () => {
    mocks.getNumistaValuesStatus.mockResolvedValue({ data: status({ enabled: false, nextRunIn: '' }) })
    const wrapper = mountCard({ NumistaValuesEnabled: 'false' })
    await flushPromises()
    expect(wrapper.text()).toContain('O lote diário está desligado.')
  })

  it('starts a batch now and follows it until it ends', async () => {
    vi.useFakeTimers()
    const wrapper = mountCard()
    await flushPromises()

    mocks.getNumistaValuesStatus.mockResolvedValue({ data: status({ isRunning: true }) })
    await wrapper.get('[data-testid="numista-values-run"]').trigger('click')
    await flushPromises()
    expect(mocks.runNumistaValuesNow).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('Lote iniciado.')
    expect(wrapper.get('[data-testid="numista-values-run"]').attributes('disabled')).toBeDefined()

    mocks.getNumistaValuesStatus.mockResolvedValue({ data: status() })
    await vi.advanceTimersByTimeAsync(4000)
    await flushPromises()
    expect(mocks.getNumistaValuesStatus).toHaveBeenCalledTimes(2)
    expect(wrapper.get('[data-testid="numista-values-run"]').attributes('disabled')).toBeUndefined()

    await vi.advanceTimersByTimeAsync(8000)
    expect(mocks.getNumistaValuesStatus).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })

  it('says when a batch is already running', async () => {
    mocks.runNumistaValuesNow.mockRejectedValue({ response: { status: 409 } })
    const wrapper = mountCard()
    await flushPromises()
    await wrapper.get('[data-testid="numista-values-run"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Já está a correr um lote.')
  })
})
