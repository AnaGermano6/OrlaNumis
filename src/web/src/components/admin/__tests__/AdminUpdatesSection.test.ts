import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import AdminUpdatesSection from '../AdminUpdatesSection.vue'
import { setLocale } from '@/i18n'

const mocks = vi.hoisted(() => ({
  getUpdateStatus: vi.fn(),
  checkForUpdates: vi.fn(),
  applyUpdate: vi.fn(),
  getAppSettings: vi.fn(),
  updateAppSettings: vi.fn(),
  showConfirm: vi.fn(),
}))
vi.mock('@/api/client', () => mocks)
vi.mock('@/composables/useDialog', () => ({ useDialog: () => ({ showConfirm: mocks.showConfirm }) }))
vi.mock('vue-router', () => ({ RouterLink: { props: ['to'], template: '<a><slot /></a>' } }))

const STATUS = {
  current: '1.0.0', latest: '1.1.0', updateAvailable: true, status: 'ok', watchtower: false, image: 'anagermano/orlanumis',
  notes: '### Collection\n- Periods <script>alert(1)</script>', releaseUrl: 'https://github.com/AnaGermano6/OrlaNumis/releases/tag/v1.1.0',
  publishedAt: '2026-11-02T10:00:00Z', checkedAt: '2026-11-03T09:00:00Z',
}

describe('AdminUpdatesSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setLocale('pt-PT')
    mocks.getUpdateStatus.mockResolvedValue({ data: STATUS })
    mocks.getAppSettings.mockResolvedValue({ data: { UpdateCheckEnabled: 'true', WatchtowerURL: '', WatchtowerToken: '' } })
  })
  afterEach(() => setLocale('en'))

  it('shows the new version, its notes (sanitised) and how to update without Watchtower', async () => {
    const wrapper = mount(AdminUpdatesSection)
    await flushPromises()

    expect(wrapper.get('[data-testid="updates-current"]').text()).toBe('1.0.0')
    expect(wrapper.get('[data-testid="updates-latest"]').text()).toBe('1.1.0')
    expect(wrapper.get('[data-testid="updates-message"]').text()).toBe('Há uma versão nova: 1.1.0, publicada a 2 de novembro de 2026.')
    const notes = wrapper.get('[data-testid="updates-notes"]')
    expect(notes.find('h3 + div h3').text()).toBe('Collection')
    expect(notes.html()).not.toContain('<script>')
    const how = wrapper.get('[data-testid="updates-how"]')
    expect(how.text()).toContain('docker compose pull && docker compose up -d')
    expect(how.text()).toContain('Apply Update')
    expect(wrapper.find('[data-testid="updates-apply"]').exists()).toBe(false)
  })

  it('updates through Watchtower after confirming, then waits for the new version', async () => {
    vi.useFakeTimers()
    mocks.getUpdateStatus.mockResolvedValue({ data: { ...STATUS, watchtower: true } })
    mocks.showConfirm.mockResolvedValue(true)
    mocks.applyUpdate.mockResolvedValue({ data: { message: 'Update started' } })
    const reload = vi.fn()
    Object.defineProperty(window, 'location', { value: { ...window.location, reload }, writable: true })
    const wrapper = mount(AdminUpdatesSection)
    await flushPromises()

    await wrapper.get('[data-testid="updates-apply"]').trigger('click')
    await flushPromises()
    expect(mocks.applyUpdate).toHaveBeenCalledTimes(1)
    expect(wrapper.get('[data-testid="updates-apply-message"]').text()).toContain('o Watchtower está a atualizar para a 1.1.0')

    mocks.getUpdateStatus.mockResolvedValue({ data: { ...STATUS, current: '1.1.0', updateAvailable: false } })
    await vi.advanceTimersByTimeAsync(5000)
    expect(reload).toHaveBeenCalled()
    wrapper.unmount()
    vi.useRealTimers()
  })

  it('saves the settings: the check and Watchtower', async () => {
    mocks.updateAppSettings.mockResolvedValue({})
    const wrapper = mount(AdminUpdatesSection)
    await flushPromises()
    await wrapper.get('[data-testid="updates-watchtower-url"]').setValue(' http://watchtower:8080 ')
    await wrapper.get('[data-testid="updates-watchtower-token"]').setValue('token')
    await wrapper.get('[data-testid="updates-enabled"]').setValue(false)
    await wrapper.get('[data-testid="updates-settings"]').trigger('submit')
    await flushPromises()
    expect(mocks.updateAppSettings).toHaveBeenCalledWith([
      { key: 'UpdateCheckEnabled', value: 'false' },
      { key: 'WatchtowerURL', value: 'http://watchtower:8080' },
      { key: 'WatchtowerToken', value: 'token' },
    ])
  })
})
