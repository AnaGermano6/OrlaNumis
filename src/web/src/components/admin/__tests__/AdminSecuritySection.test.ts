import { flushPromises, mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AdminSecuritySection from '@/components/admin/AdminSecuritySection.vue'
import type { UserInfo } from '@/types'

const mockGetSecuritySummary = vi.fn()
const mockGetSecurityEvents = vi.fn()
const mockGetSecurityExposureCheck = vi.fn()
const mockUnlockUser = vi.fn()

vi.mock('@/api/client', () => ({
  getSecuritySummary: () => mockGetSecuritySummary(),
  getSecurityEvents: (filters?: Record<string, unknown>) => mockGetSecurityEvents(filters),
  getSecurityExposureCheck: () => mockGetSecurityExposureCheck(),
  unlockUser: (id: number) => mockUnlockUser(id),
}))

const baseUser: UserInfo = {
  id: 1,
  username: 'admin',
  role: 'admin',
  email: 'admin@example.com',
  avatarPath: '',
  bio: '',
  zipCode: '',
  emailMissing: false,
  numisBidsUsername: '',
  numisBidsConfigured: false,
  createdAt: '2026-01-01T00:00:00Z',
}

function mountSection(users: UserInfo[] = []) {
  return mount(AdminSecuritySection, {
    props: {
      users,
      registrationMode: 'invite',
    },
  })
}

describe('AdminSecuritySection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetSecuritySummary.mockResolvedValue({
      data: { failedLogins: 3, lockedAccounts: 1, recentEvents: 7 },
    })
    mockGetSecurityEvents.mockResolvedValue({
      data: {
        events: [
          {
            id: 10,
            timestamp: '2026-06-19T12:00:00Z',
            type: 'login',
            severity: 'warning',
            outcome: 'failure',
            username: 'brian',
            ip: '203.0.113.5',
            message: 'Failed login',
          },
        ],
      },
    })
    mockGetSecurityExposureCheck.mockResolvedValue({
      data: {
        publicIp: '198.51.100.4',
        proxy: true,
        cors: false,
        corsWarning: 'CORS allows a broad origin',
        publicAppUrl: true,
        registration: true,
        agentToken: true,
      },
    })
    mockUnlockUser.mockResolvedValue({ data: {} })
  })

  it('shows loading state while security data loads', async () => {
    mockGetSecuritySummary.mockReturnValue(new Promise(() => {}))

    const wrapper = mountSection()
    await nextTick()

    expect(wrapper.text()).toContain('Loading...')
  })

  it('loads summary, exposure and events', async () => {
    const wrapper = mountSection([{ ...baseUser, id: 2, username: 'locked', lockedUntil: '2999-01-01T00:00:00Z' }])
    await flushPromises()

    expect(wrapper.text()).toContain('Failed Logins')
    expect(wrapper.text()).toContain('3')
    expect(wrapper.text()).toContain('API sees your IP as')
    expect(wrapper.text()).toContain('Failed login')
    expect(wrapper.text()).toContain('locked')
  })

  it('normalizes backend security response shapes from the Go admin API', async () => {
    mockGetSecuritySummary.mockResolvedValue({
      data: {
        summary: {
          loginFailures: 5,
        },
        backupStatus: 'restore drill complete',
      },
    })
    mockGetSecurityEvents.mockResolvedValue({
      data: {
        events: [
          {
            id: 11,
            createdAt: '2026-06-19T12:00:00Z',
            type: 'password_login_failure',
            username: 'alice',
            clientIp: '198.51.100.10',
            message: 'failed password login',
          },
        ],
      },
    })
    const wrapper = mountSection()
    await flushPromises()

    expect(wrapper.text()).toContain('5')
    expect(wrapper.text()).toContain('198.51.100.10')
  })

  it('renders the empty events state', async () => {
    mockGetSecurityEvents.mockResolvedValue({ data: { events: [] } })

    const wrapper = mountSection()
    await flushPromises()

    expect(wrapper.text()).toContain('No security events match the current filters.')
  })

  it('shows an error when loading fails', async () => {
    mockGetSecuritySummary.mockRejectedValue(new Error('backend unavailable'))

    const wrapper = mountSection()
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to load security data')
  })

  it('applies security event filters', async () => {
    const wrapper = mountSection()
    await flushPromises()

    await wrapper.find('input[placeholder="Type"]').setValue('login')
    await wrapper.find('input[placeholder="User"]').setValue('brian')
    await wrapper.find('input[placeholder="IP"]').setValue('203.0.113.5')
    await wrapper.find('input[type="date"]').setValue('2026-06-19')
    await wrapper.findAll('form')[0].trigger('submit.prevent')
    await flushPromises()

    expect(mockGetSecurityEvents).toHaveBeenLastCalledWith(expect.objectContaining({
      type: 'login',
      username: 'brian',
      clientIp: '203.0.113.5',
      since: expect.any(String),
      limit: 50,
    }))
  })

  it('marks security event dates with narrow-safe hooks', async () => {
    const wrapper = mountSection()
    await flushPromises()

    const dateInput = wrapper.find('input[type="date"]')
    expect(dateInput.classes()).toContain('appearance-none')
    expect(dateInput.classes()).toContain('overflow-hidden')

    const eventsTable = wrapper.findAll('table')[0]
    expect(eventsTable.exists()).toBe(true)
    expect(eventsTable.find('tbody td').text()).toBeTruthy()
  })

})
