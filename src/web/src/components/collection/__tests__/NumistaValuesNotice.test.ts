import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import NumistaValuesNotice from '@/components/collection/NumistaValuesNotice.vue'
import { useAuthStore } from '@/stores/auth'
import type { NumistaValueStatus, UserInfo } from '@/types'
import { setLocale } from '@/i18n'

const mocks = vi.hoisted(() => ({ getNumistaValuesStatus: vi.fn(), onTokenRefreshed: vi.fn() }))

vi.mock('@/api/client', () => mocks)

function status(overrides: Partial<NumistaValueStatus> = {}): NumistaValueStatus {
  return {
    enabled: true,
    isRunning: false,
    nextRunIn: '5h0m0s',
    usage: { month: '2026-11', requests: 1600, limit: 2000, pauseAt: 1800, backgroundPaused: false },
    quotaLevel: 'low',
    resumesOn: '2026-12-01',
    dailyBudget: 10,
    eligible: 192,
    due: 10,
    withoutGrade: 0,
    currency: 'EUR',
    runs: [],
    ...overrides,
  }
}

function signIn(role: string) {
  useAuthStore().user = { id: 1, username: 'ana', role } as UserInfo
}

function mountNotice() {
  return mount(NumistaValuesNotice, { global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } } })
}

describe('NumistaValuesNotice', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    setLocale('pt-PT')
    mocks.getNumistaValuesStatus.mockResolvedValue({ data: status() })
  })

  afterEach(() => setLocale('en'))

  it('warns an admin that the quota is running low', async () => {
    signIn('admin')
    const wrapper = mountNotice()
    await flushPromises()
    expect(wrapper.get('[data-testid="numista-values-notice"]').text()).toContain('A quota do Numista está a acabar (1600 de 2000 pedidos)')
    expect(wrapper.text()).toContain('Ver o agendamento')
  })

  it('says until when the values wait once the quota is spent', async () => {
    signIn('admin')
    mocks.getNumistaValuesStatus.mockResolvedValue({ data: status({ quotaLevel: 'paused' }) })
    const wrapper = mountNotice()
    await flushPromises()
    expect(wrapper.text()).toContain('a partir de 1 de dezembro')
  })

  it('stays hidden while the quota is fine or the batch is off', async () => {
    signIn('admin')
    mocks.getNumistaValuesStatus.mockResolvedValue({ data: status({ quotaLevel: 'ok' }) })
    expect(mountNotice().find('[data-testid="numista-values-notice"]').exists()).toBe(false)
    await flushPromises()

    mocks.getNumistaValuesStatus.mockResolvedValue({ data: status({ enabled: false }) })
    const off = mountNotice()
    await flushPromises()
    expect(off.find('[data-testid="numista-values-notice"]').exists()).toBe(false)
  })

  it('asks nothing for someone who is not an admin', async () => {
    signIn('user')
    const wrapper = mountNotice()
    await flushPromises()
    expect(mocks.getNumistaValuesStatus).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="numista-values-notice"]').exists()).toBe(false)
  })
})
