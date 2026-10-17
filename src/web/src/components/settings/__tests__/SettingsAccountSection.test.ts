import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import SettingsAccountSection from '@/components/settings/SettingsAccountSection.vue'
import type { User } from '@/types'

const mockShowConfirm = vi.fn()
const mockUpdateProfile = vi.fn()

const authUser: User = {
  id: 1,
  username: 'collector',
  role: 'user',
  email: 'collector@example.com',
  avatarPath: '',
  bio: '',
  zipCode: '',
  pwaSwipeNavEnabled: false,
}

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: authUser,
  }),
}))

vi.mock('@/api/client', () => ({
  getApiErrorMessage: (error: unknown) => {
    const maybeError = error as { response?: { data?: { error?: string; message?: string } }; message?: string }
    return maybeError.response?.data?.error ?? maybeError.response?.data?.message ?? maybeError.message ?? ''
  },
  changePassword: vi.fn(),
  uploadAvatar: vi.fn(),
  deleteAvatar: vi.fn(),
  updateProfile: (...args: unknown[]) => mockUpdateProfile(...args),
}))

vi.mock('@/composables/useDialog', () => ({
  useDialog: () => ({
    showConfirm: mockShowConfirm,
  }),
}))

function mountSection() {
  return mount(SettingsAccountSection, {
    global: {
      stubs: {
        AuthenticatedImage: { template: '<img alt="Avatar" />' },
        Teleport: true,
      },
    },
  })
}

function buttonByText(wrapper: ReturnType<typeof mount>, text: string) {
  const button = wrapper.findAll('button').find(candidate => candidate.text().includes(text))
  expect(button, `button containing ${text} should exist`).toBeTruthy()
  return button!
}

describe('S2: Account section does not render Swipe Navigation row', () => {
  it('does not render a Swipe Navigation toggle in the Account section', async () => {
    const { flushPromises } = await import('@vue/test-utils')
    const wrapper = mountSection()
    await flushPromises()
    expect(wrapper.text()).not.toContain('Swipe Navigation')
    wrapper.unmount()
  })
})

describe('S5: Account save payload does not contain pwaSwipeNavEnabled', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    authUser.pwaSwipeNavEnabled = true
    mockUpdateProfile.mockResolvedValue({
      data: {
        id: 1, username: 'collector', role: 'user', email: 'collector@example.com',
        avatarPath: '', bio: '', zipCode: '',
        numisBidsUsername: '', numisBidsConfigured: false, cngUsername: '', cngConfigured: false,
        pwaSwipeNavEnabled: true,
      },
    })
  })

  it('Save Profile payload does not contain pwaSwipeNavEnabled', async () => {
    const { flushPromises } = await import('@vue/test-utils')
    const wrapper = mountSection()
    await flushPromises()
    await buttonByText(wrapper, 'Save Profile').trigger('click')
    await flushPromises()
    expect(mockUpdateProfile).toHaveBeenCalled()
    const callArg = mockUpdateProfile.mock.calls[0]?.[0] as Record<string, unknown>
    expect(callArg).not.toHaveProperty('pwaSwipeNavEnabled')
    wrapper.unmount()
  })
})
