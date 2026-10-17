<template>
  <section class="card mb-4" data-testid="full-backup">
    <h2 class="mb-3 border-b border-border-subtle pb-3 text-lg">{{ t('settings.backups.full.title') }}</h2>
    <p class="mb-2 mt-0 text-sm text-text-muted">{{ t('settings.backups.full.intro') }}</p>
    <p v-if="state" class="mb-1 mt-0 text-sm text-text-secondary" data-testid="full-backup-schedule">
      {{ state.enabled ? t('settings.backups.full.scheduleOn', { next: formatDate(nextRun(state)), keep: state.keep }) : t('settings.backups.full.scheduleOff') }}
      <RouterLink :to="{ path: '/admin', query: { tab: 'schedules' } }" class="text-gold underline-offset-2 hover:underline">{{ t('settings.backups.full.scheduleChange') }}</RouterLink>
    </p>
    <p v-if="state?.folder" class="mb-3 mt-0 break-all text-sm text-text-muted" data-testid="full-backup-folder">{{ t('settings.backups.full.folder', { folder: state.folder }) }}</p>

    <p
      v-if="state?.lastRestore"
      class="mb-3 mt-0 rounded-md border border-border-subtle p-3 text-sm"
      :class="state.lastRestore.ok ? 'text-text-secondary' : 'text-loss'"
      :role="state.lastRestore.ok ? 'status' : 'alert'"
      data-testid="full-backup-last-restore"
    >
      {{ state.lastRestore.ok
        ? t('settings.backups.full.restored', { date: formatDate(state.lastRestore.at) })
        : t('settings.backups.full.restoreFailed', { date: formatDate(state.lastRestore.at), error: state.lastRestore.error ?? '' }) }}
    </p>

    <div class="mb-3 flex flex-wrap items-center gap-2">
      <button type="button" class="btn btn-primary btn-sm" :disabled="busy || state?.running" data-testid="full-backup-create" @click="create">
        <DatabaseBackup :size="16" /> {{ state?.running ? t('settings.backups.full.creating') : t('settings.backups.full.create') }}
      </button>
      <label class="btn btn-secondary btn-sm inline-flex cursor-pointer focus-within:outline-2 focus-within:outline-gold focus-within:outline-offset-2" :class="{ 'pointer-events-none opacity-55': busy }">
        <Upload :size="16" /> {{ t('settings.backups.full.import') }}
        <input type="file" accept=".zip,application/zip" hidden data-testid="full-backup-import" :disabled="busy" @change="importFile" />
      </label>
    </div>

    <p v-if="restarting" class="mb-3 mt-0 text-body text-gold" role="status" data-testid="full-backup-restarting">{{ t('settings.backups.full.restarting') }}</p>
    <p v-else-if="progress" class="mb-3 mt-0 text-body text-text-secondary" role="status">{{ progress }}</p>
    <p v-if="message" class="mb-3 mt-0 text-body" :class="failed ? 'text-loss' : 'text-gold'" :role="failed ? 'alert' : 'status'" data-testid="full-backup-message">{{ message }}</p>

    <p v-if="loadFailed" class="m-0 text-body text-loss" role="alert">{{ t('settings.backups.full.loadFailed') }}</p>
    <p v-else-if="state && !state.backups.length" class="m-0 text-body text-text-muted" data-testid="full-backup-empty">{{ t('settings.backups.full.empty') }}</p>
    <ul v-else-if="state" class="m-0 list-none p-0">
      <li
        v-for="backup in state.backups"
        :key="backup.name"
        class="flex flex-col gap-2 border-t border-border-subtle py-3 first:border-t-0 md:flex-row md:items-center md:justify-between"
        data-testid="full-backup-item"
      >
        <span class="min-w-0">
          <span class="flex flex-wrap items-center gap-2">
            <span class="font-medium text-text-primary">{{ formatDate(backup.createdAt) }}</span>
            <span class="rounded-full border border-border-subtle px-2 py-[0.1rem] text-chip text-text-secondary">{{ t(`settings.backups.full.kinds.${backup.kind}`) }}</span>
          </span>
          <span class="block text-sm text-text-muted">
            {{ formatSize(backup.size) }} · {{ t('settings.backups.full.coins', { count: backup.coins }) }} · {{ t('settings.backups.full.files', { count: backup.files }) }}
          </span>
        </span>
        <span class="flex shrink-0 flex-wrap gap-2">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="busy" data-testid="full-backup-download" @click="download(backup)">
            <Download :size="14" /> {{ t('settings.backups.full.download') }}
          </button>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="busy || state.running" data-testid="full-backup-restore" @click="restore(backup)">
            <RotateCcw :size="14" /> {{ t('settings.backups.full.restore') }}
          </button>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="busy" :aria-label="t('settings.backups.full.deleteLabel', { date: formatDate(backup.createdAt) })" data-testid="full-backup-delete" @click="remove(backup)">
            <Trash2 :size="14" />
          </button>
        </span>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
// The full backup, for administrators: the whole database and every photo
// in one zip, made every day by itself or now; downloaded, imported from a
// file, and restored (the app then restarts and the page opens again).
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { DatabaseBackup, Download, RotateCcw, Trash2, Upload } from 'lucide-vue-next'
import axios from 'axios'
import {
  createBackup, deleteBackup, downloadBackup, getBackups, restoreBackup, uploadBackup,
  type BackupInfo, type BackupsResponse,
} from '@/api/client'
import { useDialog } from '@/composables/useDialog'
import { intlLocale, t } from '@/i18n'
import { reloadPage, waitForRestart } from '@/utils/appRestart'

const { showConfirm } = useDialog()

const state = ref<BackupsResponse | null>(null)
const loadFailed = ref(false)
/** A download, import, restore or deletion under way. */
const busy = ref(false)
const progress = ref('')
const message = ref('')
const failed = ref(false)
const restarting = ref(false)
let poll: ReturnType<typeof setTimeout> | undefined

function formatDate(value: string): string {
  return new Intl.DateTimeFormat(intlLocale(), { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

// The next automatic backup, in the viewer's time (the schedule's time is the server's zone's).
function nextRun(response: BackupsResponse): string {
  const minute = 60_000
  return new Date(Math.round((Date.now() + response.nextRunIn * 1000) / minute) * minute).toISOString()
}

function formatSize(bytes: number): string {
  return `${(bytes / 1_000_000).toLocaleString(intlLocale(), { maximumFractionDigits: 1 })} MB`
}

function say(text: string, error = false) {
  message.value = text
  failed.value = error
}

async function load() {
  try {
    state.value = (await getBackups()).data
    loadFailed.value = false
  } catch {
    loadFailed.value = true
  }
  clearTimeout(poll)
  // While a backup is made, the list again in a moment.
  if (state.value?.running) poll = setTimeout(() => void load().then(afterPoll), 3000)
}

let waitingForCreate = false
function afterPoll() {
  if (waitingForCreate && state.value && !state.value.running) {
    waitingForCreate = false
    say(t('settings.backups.full.created'))
  }
}

async function create() {
  say('')
  try {
    await createBackup()
    waitingForCreate = true
    await load()
    afterPoll()
  } catch {
    say(t('settings.backups.full.createFailed'), true)
  }
}

async function download(backup: BackupInfo) {
  busy.value = true
  say('')
  try {
    const { data } = await downloadBackup(backup.name, (fraction) => {
      progress.value = t('settings.backups.full.downloading', { percent: Math.round(fraction * 100) })
    })
    const url = URL.createObjectURL(data)
    const link = document.createElement('a')
    link.href = url
    link.download = backup.name
    link.click()
    URL.revokeObjectURL(url)
  } catch {
    say(t('settings.backups.full.downloadFailed'), true)
  } finally {
    progress.value = ''
    busy.value = false
  }
}

async function importFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  busy.value = true
  say('')
  try {
    await uploadBackup(file, (fraction) => {
      progress.value = t('settings.backups.full.uploading', { percent: Math.round(fraction * 100) })
    })
    say(t('settings.backups.full.imported'))
    await load()
  } catch (error) {
    const status = axios.isAxiosError(error) ? error.response?.status : undefined
    say(status === 413 ? t('settings.backups.full.tooLarge') : status === 400 ? t('settings.backups.full.notABackup') : t('settings.backups.full.importFailed'), true)
  } finally {
    progress.value = ''
    busy.value = false
  }
}

async function restore(backup: BackupInfo) {
  const confirmed = await showConfirm(t('settings.backups.full.restoreConfirm', { date: formatDate(backup.createdAt) }), {
    title: t('settings.backups.full.restoreTitle'),
    confirmLabel: t('settings.backups.full.restore'),
    variant: 'danger',
  })
  if (!confirmed) return
  busy.value = true
  say('')
  progress.value = t('settings.backups.full.preparing')
  try {
    await restoreBackup(backup.name)
    progress.value = ''
    restarting.value = true
    await waitForRestart()
    reloadPage()
  } catch (error) {
    const status = axios.isAxiosError(error) ? error.response?.status : undefined
    say(status === 409 ? t('settings.backups.full.busyNow') : t('settings.backups.full.restoreNotPossible'), true)
    progress.value = ''
    busy.value = false
  }
}

async function remove(backup: BackupInfo) {
  const confirmed = await showConfirm(t('settings.backups.full.deleteConfirm', { date: formatDate(backup.createdAt) }), {
    title: t('settings.backups.full.deleteTitle'),
    confirmLabel: t('common.delete'),
    variant: 'danger',
  })
  if (!confirmed) return
  busy.value = true
  say('')
  try {
    await deleteBackup(backup.name)
    await load()
  } catch {
    say(t('settings.backups.full.deleteFailed'), true)
  } finally {
    busy.value = false
  }
}

onMounted(load)
onBeforeUnmount(() => clearTimeout(poll))
</script>
