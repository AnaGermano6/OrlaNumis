<template>
  <hr class="my-6 border-0 border-t border-border-subtle" />

  <!-- Collection Health Snapshots -->
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.health.title') }}</h3>
  <p class="mb-4 text-base text-text-secondary">{{ t('admin.schedules.health.intro') }}</p>
  <p v-if="statusLoading" class="mb-4 text-body text-text-muted">{{ t('admin.schedules.health.checking') }}</p>
  <p v-else-if="status" class="mb-4 text-body" :class="status.enabled ? 'text-gain' : 'text-text-muted'">
    {{ t('admin.schedules.serverStatus') }} <strong>{{ status.enabled ? t('admin.oidc.enabled') : t('admin.oidc.disabled') }}</strong>{{ status.enabled ? ` — ${t('admin.schedules.nextRunIn', { value: formatNextRunIn(status.nextRunIn) })}` : '' }}
  </p>
  <div class="mb-4">
    <div class="form-group flex items-center justify-between gap-3">
      <label class="form-label">{{ t('admin.schedules.health.enable') }}</label>
      <BaseToggle
        size="sm"
        :label="t('admin.schedules.health.enable')"
        :model-value="settings.CollectionHealthSnapshotsEnabled === 'true'"
        @update:model-value="settings.CollectionHealthSnapshotsEnabled = $event ? 'true' : 'false'"
      />
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('admin.schedules.startDaily') }}</label>
      <input
        v-model="settings.CollectionHealthSnapshotsStartTime"
        class="form-input w-full max-w-[120px]"
        type="time"
      />
      <span class="form-hint">{{ t('admin.schedules.health.startHint') }}</span>
      <AdminScheduleSummary :start-time="settings.CollectionHealthSnapshotsStartTime" default-start-time="04:30" :zone="settings.ScheduleTimezone" />
    </div>
    <div class="mt-4 flex w-full flex-col gap-3 md:flex-row md:items-center">
      <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="emit('save')">
        {{ settingsSaving ? t('common.saving') : t('admin.schedules.health.save') }}
      </button>
      <span v-if="settingsMsg" class="text-body text-gold md:mr-auto" :class="settingsError ? 'text-loss' : ''">{{ settingsMsg }}</span>
      <button class="btn btn-secondary btn-sm md:ml-auto" :disabled="triggerLoading" @click="triggerManualSnapshots()">
        {{ triggerLoading ? t('common.running') : t('admin.schedules.runNow') }}
      </button>
    </div>
  </div>

  <hr class="my-6 border-0 border-t border-border-subtle" />
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.health.history') }}</h3>
  <div v-if="loading" class="flex justify-center py-8"><div class="spinner"></div></div>
  <div v-else-if="runs.length === 0" class="px-8 py-8 text-center font-sans text-text-muted">{{ t('admin.schedules.health.empty') }}</div>
  <template v-else>
    <div class="overflow-x-auto">
      <table class="data-table text-chip md:table-fixed [&_th]:px-[0.35rem] [&_th]:py-2 md:[&_th]:px-2 md:[&_th]:py-3 [&_td]:px-[0.35rem] [&_td]:py-2 md:[&_td]:px-2 md:[&_td]:py-3">
        <thead>
          <tr>
            <th>{{ t('common.date') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.trigger') }}</th>
            <th>{{ t('common.status') }}</th>
            <th>{{ t('admin.schedules.health.eligible') }}</th>
            <th>{{ t('admin.schedules.health.snapshotted') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.failed') }}</th>
            <th>{{ t('admin.schedules.duration') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="run in runs" :key="run.id">
            <td class="text-body text-text-secondary">{{ formatDate(run.startedAt) }}</td>
            <td class="hidden md:table-cell">{{ runTriggerLabel(run.triggerType) }}</td>
            <td>
              <BaseStatusBadge :tone="run.status === 'error' ? 'error' : run.status === 'success' ? 'success' : 'warning'">{{ runStatusLabel(run.status) }}</BaseStatusBadge>
            </td>
            <td>{{ run.usersEligible }}</td>
            <td class="font-semibold text-gain">{{ run.usersSnapshotted }}</td>
            <td class="hidden font-semibold text-loss md:table-cell">{{ run.usersFailed }}</td>
            <td>{{ formatDuration(run.durationMs) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="mt-4 flex items-center justify-center gap-3">
      <button class="btn btn-secondary btn-sm" :disabled="page <= 1" @click="prevPage()">{{ t('common.prev') }}</button>
      <span class="text-chip text-text-secondary">{{ t('admin.schedules.page', { page }) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="runs.length < 5" @click="nextPage()">{{ t('common.next') }}</button>
    </div>
  </template>
</template>

<script setup lang="ts">
import { BaseStatusBadge, BaseToggle } from '@/components/ui'
import AdminScheduleSummary from '@/components/admin/schedules/AdminScheduleSummary.vue'
import { onMounted, onUnmounted, ref, watch } from 'vue'
import {
  getCollectionHealthSnapshotRuns,
  getCollectionHealthSnapshotStatus,
  triggerCollectionHealthSnapshots,
} from '@/api/client'
import { useRunHistoryPagination } from '@/composables/useRunHistoryPagination'
import type { AppSettings, CollectionHealthSnapshotRun, SchedulerStatus } from '@/types'
import { t, intlLocale } from '@/i18n'
import { runStatusLabel, runTriggerLabel } from '@/utils/format'

const props = defineProps<{
  settings: AppSettings
  settingsSaving: boolean
  settingsMsg: string
  settingsError: boolean
}>()

const emit = defineEmits<{
  save: []
  'update:settingsMsg': [val: string]
  'update:settingsError': [val: boolean]
}>()

const triggerLoading = ref(false)
const status = ref<SchedulerStatus | null>(null)
const statusLoading = ref(false)
const timers: ReturnType<typeof setTimeout>[] = []

const {
  runs,
  page,
  loading,
  loadRuns,
  prevPage,
  nextPage,
} = useRunHistoryPagination<CollectionHealthSnapshotRun>(async (currentPage, limit) => {
  const res = await getCollectionHealthSnapshotRuns(currentPage, limit)
  return res.data ?? {}
})

watch(() => props.settingsSaving, (saving, wasSaving) => {
  if (wasSaving && !saving) {
    loadStatus()
  }
})

async function loadStatus() {
  statusLoading.value = true
  try {
    const res = await getCollectionHealthSnapshotStatus()
    status.value = res.data
  } catch {
    status.value = null
  } finally {
    statusLoading.value = false
  }
}

function formatNextRunIn(nanoseconds: number) {
  const totalMinutes = Math.max(0, Math.round(nanoseconds / 1e9 / 60))
  const hours = Math.floor(totalMinutes / 60)
  const minutes = totalMinutes % 60
  if (hours === 0) return `${minutes}m`
  return `${hours}h ${minutes}m`
}

async function triggerManualSnapshots() {
  triggerLoading.value = true
  emit('update:settingsMsg', '')
  emit('update:settingsError', false)
  try {
    const res = await triggerCollectionHealthSnapshots()
    const { message, users, snapshotsCreated, skipped, errors, durationMs } = res.data
    const parts = [
      snapshotsCreated != null ? t('admin.schedules.health.snapshots', { count: snapshotsCreated }) : null,
      users != null ? t('admin.schedules.health.users', { count: users }) : null,
      skipped != null ? t('admin.schedules.health.skipped', { count: skipped }) : null,
      errors ? t('admin.schedules.health.errors', { count: errors }) : null,
      durationMs != null ? `${(durationMs / 1000).toFixed(1)}s` : null,
    ].filter((part): part is string => part != null)
    emit('update:settingsMsg', message ?? (parts.length ? t('admin.schedules.health.completeWith', { parts: parts.join(', ') }) : t('admin.schedules.health.complete')))
    if (errors) {
      emit('update:settingsError', true)
    }
    timers.push(setTimeout(() => { emit('update:settingsMsg', '') }, 10000))
    timers.push(setTimeout(() => { loadRuns() }, 1000))
    timers.push(setTimeout(() => { loadStatus() }, 1000))
  } catch {
    emit('update:settingsMsg', t('admin.schedules.health.failed'))
    emit('update:settingsError', true)
  } finally {
    triggerLoading.value = false
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString(intlLocale())
}

function formatDuration(ms: number) {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

onMounted(() => {
  loadRuns()
  loadStatus()
})

onUnmounted(() => {
  timers.forEach(clearTimeout)
})
</script>
