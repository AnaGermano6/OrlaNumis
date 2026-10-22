<template>
  <hr class="my-6 border-0 border-t border-border-subtle" />

  <!-- Coin values from Numista's prices -->
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.numistaValues.title') }}</h3>
  <p class="mb-4 text-base text-text-secondary">{{ t('admin.schedules.numistaValues.intro') }}</p>
  <div class="mb-4" data-testid="numista-values-schedule">
    <div class="form-group flex items-center justify-between gap-3">
      <label class="form-label" for="numista-values-enabled">{{ t('admin.schedules.numistaValues.enable') }}</label>
      <BaseToggle
        id="numista-values-enabled"
        size="sm"
        :label="t('admin.schedules.numistaValues.enable')"
        :model-value="settings.NumistaValuesEnabled === 'true'"
        @update:model-value="settings.NumistaValuesEnabled = $event ? 'true' : 'false'"
      />
    </div>
    <div class="form-group">
      <label class="form-label" for="numista-values-start-time">{{ t('admin.schedules.startDaily') }}</label>
      <input
        id="numista-values-start-time"
        v-model="settings.NumistaValuesStartTime"
        class="form-input w-full max-w-[120px]"
        type="time"
        aria-describedby="numista-values-start-time-hint"
      />
      <span id="numista-values-start-time-hint" class="form-hint">{{ t('admin.schedules.numistaValues.startHint') }}</span>
      <AdminScheduleSummary :start-time="settings.NumistaValuesStartTime" default-start-time="05:00" :zone="settings.ScheduleTimezone" />
    </div>
    <div class="form-group">
      <label class="form-label" for="numista-values-daily">{{ t('admin.schedules.numistaValues.dailyRequests') }}</label>
      <input
        id="numista-values-daily"
        v-model="settings.NumistaValuesDailyRequests"
        class="form-input w-full max-w-[120px]"
        type="number"
        min="1"
        max="500"
        aria-describedby="numista-values-daily-hint"
      />
      <span id="numista-values-daily-hint" class="form-hint">{{ t('admin.schedules.numistaValues.dailyRequestsHint') }}</span>
    </div>

    <p v-if="loadError" class="text-body text-loss" role="alert">{{ t('admin.schedules.numistaValues.loadFailed') }}</p>
    <div
      v-else-if="status"
      class="mt-2 rounded-sm border bg-card p-4"
      :class="status.quotaLevel === 'ok' ? 'border-border-subtle' : 'border-status-warning-border'"
      data-testid="numista-values-status"
    >
      <div class="flex flex-wrap items-baseline justify-between gap-2">
        <span class="text-body font-medium text-text-primary">{{ t('admin.schedules.numistaValues.quota') }}</span>
        <span class="text-body tabular-nums text-text-secondary">
          {{ t('admin.schedules.numistaValues.quotaCount', { requests: formatCount(status.usage.requests), limit: formatCount(status.usage.limit) }) }}
        </span>
      </div>
      <div
        class="mt-2 h-2 overflow-hidden rounded-full bg-surface"
        role="progressbar"
        :aria-label="t('admin.schedules.numistaValues.quota')"
        aria-valuemin="0"
        :aria-valuemax="status.usage.limit"
        :aria-valuenow="status.usage.requests"
      >
        <div class="h-full rounded-full" :class="quotaBarClass" :style="{ width: `${quotaPercent}%` }"></div>
      </div>
      <p class="mb-0 mt-2 text-sm" :class="status.quotaLevel === 'ok' ? 'text-text-secondary' : 'text-loss'" :role="status.quotaLevel === 'ok' ? undefined : 'status'" data-testid="numista-values-quota-message">
        {{ quotaMessage }}
      </p>
      <ul class="m-0 mt-3 flex list-none flex-col gap-1 p-0 text-body text-text-secondary">
        <li v-if="status.enabled && status.quotaLevel !== 'paused'">
          {{ t('admin.schedules.numistaValues.nextBatch', { count: status.dailyBudget }) }}<template v-if="nextRunText"> · {{ t('admin.schedules.numistaValues.nextRunIn', { time: nextRunText }) }}</template>
        </li>
        <li v-else-if="!status.enabled">{{ t('admin.schedules.numistaValues.off') }}</li>
        <li>
          {{ t('admin.schedules.numistaValues.coins', { count: status.eligible }) }} · {{ t('admin.schedules.numistaValues.due', { count: status.due }) }}
        </li>
        <li v-if="status.withoutGrade">{{ t('admin.schedules.numistaValues.withoutGrade', { count: status.withoutGrade }) }}</li>
      </ul>
    </div>

    <div class="mt-4 flex w-full flex-col gap-3 md:flex-row md:items-center">
      <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="emit('save')">
        {{ settingsSaving ? t('common.saving') : t('admin.schedules.numistaValues.save') }}
      </button>
      <span v-if="message" class="text-body md:mr-auto" :class="messageError ? 'text-loss' : 'text-gold'" role="status">{{ message }}</span>
      <button
        class="btn btn-secondary btn-sm md:ml-auto"
        :disabled="starting || status?.isRunning"
        data-testid="numista-values-run"
        @click="runNow"
      >
        {{ starting ? t('common.starting') : t('admin.schedules.runNow') }}
      </button>
    </div>
  </div>

  <h4 class="mb-3 text-base font-medium text-heading">{{ t('admin.schedules.numistaValues.history') }}</h4>
  <p v-if="!status?.runs.length" class="text-body text-text-muted">{{ t('admin.schedules.numistaValues.empty') }}</p>
  <div v-else class="overflow-x-auto">
    <table class="data-table text-chip [&_td]:px-[0.35rem] [&_td]:py-2 [&_th]:px-[0.35rem] [&_th]:py-2 md:[&_td]:px-2 md:[&_th]:px-2" data-testid="numista-values-runs">
      <thead>
        <tr>
          <th>{{ t('common.date') }}</th>
          <th class="hidden md:table-cell">{{ t('admin.schedules.trigger') }}</th>
          <th>{{ t('common.status') }}</th>
          <th>{{ t('admin.schedules.numistaValues.updated') }}</th>
          <th class="hidden md:table-cell">{{ t('admin.schedules.numistaValues.unchanged') }}</th>
          <th class="hidden md:table-cell">{{ t('admin.schedules.numistaValues.noPrice') }}</th>
          <th class="hidden md:table-cell">{{ t('admin.schedules.numistaValues.failed') }}</th>
          <th>{{ t('admin.schedules.numistaValues.requests') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="run in status.runs" :key="run.id">
          <td class="text-body text-text-secondary">{{ formatDateTime(run.startedAt) }}</td>
          <td class="hidden md:table-cell">{{ runTriggerLabel(run.triggerType) }}</td>
          <td><BaseStatusBadge :tone="runTone(run.status)">{{ runStatusText(run.status) }}</BaseStatusBadge></td>
          <td class="font-semibold text-gain">{{ run.updated }}</td>
          <td class="hidden md:table-cell">{{ run.unchanged }}</td>
          <td class="hidden md:table-cell text-warning">{{ run.noPrice }}</td>
          <td class="hidden md:table-cell text-loss">{{ run.failed }}</td>
          <td class="tabular-nums">{{ run.requests }} / {{ run.budget }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
// The daily batch of coin values from Numista's prices: its settings, how the
// month's quota stands (with a warning when it runs low or is spent), the
// coins it covers, and the latest batches.
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { BaseStatusBadge, BaseToggle, type StatusTone } from '@/components/ui'
import AdminScheduleSummary from '@/components/admin/schedules/AdminScheduleSummary.vue'
import { getNumistaValuesStatus, runNumistaValuesNow } from '@/api/client'
import type { AppSettings, NumistaValueStatus } from '@/types'
import { intlLocale, t, type MessageKey } from '@/i18n'
import { runTriggerLabel } from '@/utils/format'

defineProps<{
  settings: AppSettings
  settingsSaving: boolean
}>()

const emit = defineEmits<{
  save: []
}>()

const status = ref<NumistaValueStatus | null>(null)
const loadError = ref(false)
const starting = ref(false)
const message = ref('')
const messageError = ref(false)
let pollTimer: ReturnType<typeof setInterval> | undefined

function formatCount(value: number): string {
  return value.toLocaleString(intlLocale())
}

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString(intlLocale(), { dateStyle: 'short', timeStyle: 'short' })
}

const quotaPercent = computed(() => {
  const usage = status.value?.usage
  return usage && usage.limit > 0 ? Math.min(100, (usage.requests / usage.limit) * 100) : 0
})

const quotaBarClass = computed(() => {
  switch (status.value?.quotaLevel) {
    case 'paused': return 'bg-[var(--color-negative)]'
    case 'low': return 'bg-[var(--status-warning-fg)]'
    default: return 'bg-gold'
  }
})

const quotaMessage = computed(() => {
  const value = status.value
  if (!value) return ''
  const pauseAt = formatCount(value.usage.pauseAt)
  if (value.quotaLevel === 'paused') {
    const date = new Date(`${value.resumesOn}T00:00:00`).toLocaleDateString(intlLocale(), { day: 'numeric', month: 'long' })
    return t('admin.schedules.numistaValues.quotaPaused', { date })
  }
  if (value.quotaLevel === 'low') return t('admin.schedules.numistaValues.quotaLow', { pauseAt })
  return t('admin.schedules.numistaValues.quotaOk', { pauseAt })
})

// "5h3m0s" → "5 h 3 min"; "" when there is no next run.
const nextRunText = computed(() => {
  const raw = status.value?.nextRunIn ?? ''
  const hours = Number(/(\d+)h/.exec(raw)?.[1] ?? 0)
  const minutes = Number(/(\d+)m(?!s)/.exec(raw)?.[1] ?? 0)
  if (!raw) return ''
  if (hours) return `${hours} h ${minutes} min`
  return `${Math.max(1, minutes)} min`
})

const RUN_STATUS_KEYS: Record<string, MessageKey> = {
  running: 'admin.schedules.numistaValues.statuses.running',
  completed: 'admin.schedules.numistaValues.statuses.completed',
  budget_spent: 'admin.schedules.numistaValues.statuses.budgetSpent',
  paused: 'admin.schedules.numistaValues.statuses.paused',
  unconfigured: 'admin.schedules.numistaValues.statuses.unconfigured',
  failed: 'admin.schedules.numistaValues.statuses.failed',
}

function runStatusText(value: string): string {
  const key = RUN_STATUS_KEYS[value]
  return key ? t(key) : value
}

function runTone(value: string): StatusTone {
  switch (value) {
    case 'running': return 'info'
    case 'completed':
    case 'budget_spent': return 'success'
    case 'failed':
    case 'unconfigured': return 'error'
    default: return 'warning'
  }
}

async function load() {
  try {
    status.value = (await getNumistaValuesStatus()).data
    loadError.value = false
  } catch {
    loadError.value = true
  }
  // While a batch runs, check again every few seconds.
  if (status.value?.isRunning && !pollTimer) {
    pollTimer = setInterval(() => { void load() }, 4000)
  } else if (!status.value?.isRunning && pollTimer) {
    clearInterval(pollTimer)
    pollTimer = undefined
  }
}

async function runNow() {
  starting.value = true
  message.value = ''
  try {
    await runNumistaValuesNow()
    messageError.value = false
    message.value = t('admin.schedules.numistaValues.started')
    if (status.value) status.value = { ...status.value, isRunning: true }
    if (!pollTimer) pollTimer = setInterval(() => { void load() }, 4000)
  } catch (error) {
    messageError.value = true
    const conflict = (error as { response?: { status?: number } })?.response?.status === 409
    message.value = t(conflict ? 'admin.schedules.numistaValues.alreadyRunning' : 'admin.schedules.numistaValues.startFailed')
  } finally {
    starting.value = false
  }
}

onMounted(() => { void load() })
onUnmounted(() => { if (pollTimer) clearInterval(pollTimer) })
</script>
