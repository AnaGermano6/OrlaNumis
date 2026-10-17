import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AxiosError, AxiosHeaders } from 'axios'
import SettingsFullBackup from '../SettingsFullBackup.vue'
import { setLocale } from '@/i18n'
import type { BackupInfo, BackupsResponse } from '@/api/client'

const mocks = vi.hoisted(() => ({
  getBackups: vi.fn(),
  createBackup: vi.fn(),
  downloadBackup: vi.fn(),
  deleteBackup: vi.fn(),
  restoreBackup: vi.fn(),
  uploadBackup: vi.fn(),
  showConfirm: vi.fn(),
  waitForRestart: vi.fn(),
  reloadPage: vi.fn(),
}))
vi.mock('@/api/client', () => ({
  getBackups: mocks.getBackups,
  createBackup: mocks.createBackup,
  downloadBackup: mocks.downloadBackup,
  deleteBackup: mocks.deleteBackup,
  restoreBackup: mocks.restoreBackup,
  uploadBackup: mocks.uploadBackup,
}))
vi.mock('@/composables/useDialog', () => ({ useDialog: () => ({ showConfirm: mocks.showConfirm }) }))
vi.mock('@/utils/appRestart', () => ({ waitForRestart: mocks.waitForRestart, reloadPage: mocks.reloadPage }))

function backup(fields: Partial<BackupInfo>): BackupInfo {
  return { name: 'aurearia-2026-10-16-030000-auto.zip', kind: 'auto', createdAt: '2026-10-16T03:00:00Z', size: 231_400_000, users: 1, coins: 452, files: 2043, ...fields }
}

function response(fields: Partial<BackupsResponse> = {}): { data: BackupsResponse } {
  return { data: {
    backups: [
      backup({}),
      backup({ name: 'aurearia-2026-10-15-120000-manual.zip', kind: 'manual', createdAt: '2026-10-15T12:00:00Z', coins: 1, files: 1, size: 1_000_000 }),
    ],
    running: false, enabled: true, startTime: '03:00', keep: 7, nextRunIn: 3600, folder: '~/Aurearia-backups/aurearia-dev',
    ...fields,
  } }
}

function axiosError(status: number) {
  return new AxiosError('failed', String(status), undefined, undefined, {
    status, statusText: '', headers: {}, config: { headers: new AxiosHeaders() }, data: {},
  })
}

function mountSection() {
  return mount(SettingsFullBackup, { global: { stubs: { RouterLink: RouterLinkStub } } })
}

describe('SettingsFullBackup', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    vi.clearAllMocks()
    mocks.getBackups.mockResolvedValue(response())
  })

  it('lists the backups with what they hold, and the schedule', async () => {
    const wrapper = mountSection()
    await flushPromises()

    expect(wrapper.get('[data-testid="full-backup-schedule"]').text()).toMatch(/^Feita sozinha todos os dias, a próxima a .+; guardam-se as últimas 7\./)
    expect(wrapper.get('[data-testid="full-backup-folder"]').text()).toBe('Pasta: ~/Aurearia-backups/aurearia-dev')
    const items = wrapper.findAll('[data-testid="full-backup-item"]')
    expect(items).toHaveLength(2)
    expect(items[0].text()).toContain('Automática')
    expect(items[0].text()).toContain('231,4 MB · 452 moedas · 2043 ficheiros')
    expect(items[1].text()).toContain('Manual')
    expect(items[1].text()).toContain('1 moeda · 1 ficheiro')
  })

  it('makes a backup now, and says when it is made', async () => {
    vi.useFakeTimers()
    try {
      mocks.createBackup.mockResolvedValue({})
      const wrapper = mountSection()
      await flushPromises()
      mocks.getBackups.mockResolvedValueOnce(response({ running: true }))
      await wrapper.get('[data-testid="full-backup-create"]').trigger('click')
      await flushPromises()
      expect(mocks.createBackup).toHaveBeenCalled()
      expect(wrapper.get('[data-testid="full-backup-create"]').text()).toContain('A fazer a cópia…')
      expect(wrapper.get('[data-testid="full-backup-create"]').attributes('disabled')).toBeDefined()

      await vi.advanceTimersByTimeAsync(3000)
      await flushPromises()
      expect(wrapper.get('[data-testid="full-backup-message"]').text()).toBe('Cópia feita.')
      expect(wrapper.get('[data-testid="full-backup-create"]').attributes('disabled')).toBeUndefined()
    } finally {
      vi.useRealTimers()
    }
  })

  it('restores a backup once confirmed, then waits for the app to restart', async () => {
    mocks.showConfirm.mockResolvedValueOnce(false).mockResolvedValueOnce(true)
    mocks.restoreBackup.mockResolvedValue({ data: { restarting: true } })
    let restarted: () => void = () => {}
    mocks.waitForRestart.mockReturnValue(new Promise<void>((resolve) => { restarted = resolve }))
    const wrapper = mountSection()
    await flushPromises()

    const restoreButton = () => wrapper.findAll('[data-testid="full-backup-restore"]')[1]
    await restoreButton().trigger('click')
    await flushPromises()
    expect(mocks.showConfirm).toHaveBeenCalledWith(expect.stringContaining('O que mudou depois perde-se'), expect.objectContaining({ variant: 'danger' }))
    expect(mocks.restoreBackup).not.toHaveBeenCalled()

    await restoreButton().trigger('click')
    await flushPromises()
    expect(mocks.restoreBackup).toHaveBeenCalledWith('aurearia-2026-10-15-120000-manual.zip')
    expect(wrapper.get('[data-testid="full-backup-restarting"]').text()).toContain('a aplicação está a reiniciar')
    expect(mocks.reloadPage).not.toHaveBeenCalled()
    restarted()
    await flushPromises()
    expect(mocks.reloadPage).toHaveBeenCalled()
  })

  it('says why a backup could not be restored', async () => {
    mocks.showConfirm.mockResolvedValue(true)
    mocks.restoreBackup.mockRejectedValue(axiosError(400))
    const wrapper = mountSection()
    await flushPromises()
    await wrapper.findAll('[data-testid="full-backup-restore"]')[0].trigger('click')
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toBe('Não foi possível restaurar a cópia: está danificada, ou não é uma cópia da aplicação.')
    expect(mocks.waitForRestart).not.toHaveBeenCalled()
  })

  it('imports a backup file, and tells a file that is not one', async () => {
    const wrapper = mountSection()
    await flushPromises()
    const input = wrapper.get('[data-testid="full-backup-import"]')
    const file = new File(['PK'], 'aurearia.zip', { type: 'application/zip' })
    Object.defineProperty(input.element, 'files', { value: [file], configurable: true })

    mocks.uploadBackup.mockResolvedValueOnce({ data: backup({ kind: 'uploaded' }) })
    await input.trigger('change')
    await flushPromises()
    expect(mocks.uploadBackup).toHaveBeenCalledWith(file, expect.any(Function))
    expect(wrapper.get('[data-testid="full-backup-message"]').text()).toBe('Cópia importada: pode restaurá-la na lista abaixo.')

    mocks.uploadBackup.mockRejectedValueOnce(axiosError(400))
    await input.trigger('change')
    await flushPromises()
    expect(wrapper.get('[data-testid="full-backup-message"]').text()).toBe('Este ficheiro não é uma cópia de segurança da aplicação, ou está danificado.')
  })

  it('deletes a backup once confirmed', async () => {
    mocks.showConfirm.mockResolvedValue(true)
    mocks.deleteBackup.mockResolvedValue({})
    const wrapper = mountSection()
    await flushPromises()
    await wrapper.findAll('[data-testid="full-backup-delete"]')[1].trigger('click')
    await flushPromises()
    expect(mocks.deleteBackup).toHaveBeenCalledWith('aurearia-2026-10-15-120000-manual.zip')
    expect(mocks.getBackups).toHaveBeenCalledTimes(2)
  })
})
