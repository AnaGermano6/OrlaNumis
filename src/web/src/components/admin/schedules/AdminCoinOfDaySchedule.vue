<template>
  <hr class="my-6 border-0 border-t border-border-subtle" />

  <!-- Coin of the Day -->
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('featured.title') }}</h3>
  <p class="mb-4 text-base text-text-secondary">{{ t('admin.schedules.cod.intro') }}</p>
  <div class="mb-4">
    <div class="form-group flex items-center justify-between gap-3">
      <label class="form-label">{{ t('admin.schedules.cod.enable') }}</label>
      <BaseToggle
        size="sm"
        :label="t('admin.schedules.cod.enable')"
        :model-value="settings.CoinOfDayEnabled === 'true'"
        @update:model-value="settings.CoinOfDayEnabled = $event ? 'true' : 'false'"
      />
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('admin.schedules.startDaily') }}</label>
      <input
        v-model="settings.CoinOfDayStartTime"
        class="form-input w-full max-w-[120px]"
        type="time"
      />
      <span class="form-hint">{{ t('admin.schedules.cod.startHint') }}</span>
    </div>
    <div class="form-group">
      <label class="form-label" for="coin-of-day-timezone">{{ t('admin.schedules.tz') }}</label>
      <select
        id="coin-of-day-timezone"
        class="form-input w-full max-w-[320px]"
        :value="settings.CoinOfDayTimezone ?? ''"
        @change="settings.CoinOfDayTimezone = ($event.target as HTMLSelectElement).value"
      >
        <option value="">{{ t('admin.schedules.cod.sameAsSchedules', { zone: sharedZoneLabel }) }}</option>
        <option v-for="zone in timezoneOptions" :key="zone" :value="zone">{{ zone }}</option>
      </select>
      <span class="form-hint">{{ scheduleSummary }} {{ t('admin.schedules.cod.applyHint') }}</span>
      <button
        v-if="browserTimezone && settings.CoinOfDayTimezone !== browserTimezone"
        type="button"
        class="mt-1 self-start bg-transparent p-0 text-sm text-gold underline"
        @click="settings.CoinOfDayTimezone = browserTimezone"
      >
        {{ t('admin.schedules.tzUseMine', { zone: browserTimezone }) }}
      </button>
    </div>
    <div class="mt-4 flex w-full flex-col gap-3 md:flex-row md:items-center">
      <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="emit('save')">
        {{ settingsSaving ? t('common.saving') : t('admin.schedules.cod.save') }}
      </button>
      <span v-if="settingsMsg" class="text-body text-gold md:mr-auto" :class="settingsError ? 'text-loss' : ''">{{ settingsMsg }}</span>
      <button class="btn btn-secondary btn-sm md:ml-auto" :disabled="triggerLoading" @click="triggerManualCoinOfDay()">
        {{ triggerLoading ? t('common.running') : t('admin.schedules.runNow') }}
      </button>
    </div>
  </div>

  <hr class="my-6 border-0 border-t border-border-subtle" />
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.cod.history') }}</h3>
  <div v-if="loading" class="flex justify-center py-8"><div class="spinner"></div></div>
  <div v-else-if="runs.length === 0" class="px-8 py-8 text-center font-sans text-text-muted">{{ t('admin.schedules.cod.empty') }}</div>
  <template v-else>
    <div class="overflow-x-auto">
      <table class="data-table text-chip md:table-fixed [&_th]:px-[0.35rem] [&_th]:py-2 md:[&_th]:px-2 md:[&_th]:py-3 [&_td]:px-[0.35rem] [&_td]:py-2 md:[&_td]:px-2 md:[&_td]:py-3">
        <thead>
          <tr>
            <th>{{ t('common.date') }}</th>
            <th>{{ t('common.status') }}</th>
            <th>{{ t('admin.schedules.cod.picked') }}</th>
            <th>{{ t('admin.schedules.skipped') }}</th>
            <th>{{ t('admin.schedules.errors') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.trigger') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="run in runs" :key="run.id">
            <td class="text-body text-text-secondary">{{ formatDate(run.startedAt) }}</td>
            <td>{{ runStatusLabel(run.status) }}</td>
            <td>{{ run.picked }}</td>
            <td>{{ run.skipped }}</td>
            <td>{{ run.errors }}</td>
            <td class="hidden md:table-cell">{{ runTriggerLabel(run.triggerType) }}</td>
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
import { BaseToggle } from '@/components/ui'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { getCoinOfDayRunDetail, getCoinOfDayRuns, triggerCoinOfDayRun } from '@/api/client'
import { useRunHistoryPagination } from '@/composables/useRunHistoryPagination'
import type { AppSettings, CoinOfDayRun } from '@/types'
import { t, intlLocale } from '@/i18n'
import { runStatusLabel, runTriggerLabel } from '@/utils/format'
import { getBrowserTimezone } from '@/composables/usePurchaseReminder'
import { describeSchedule, scheduleZoneLabel, timezoneOptionsWith } from '@/composables/useScheduleTimezone'

const props = defineProps<{
  settings: AppSettings
  settingsSaving: boolean
}>()

const emit = defineEmits<{
  save: []
}>()

const triggerLoading = ref(false)
const browserTimezone = getBrowserTimezone()

const timezoneOptions = computed(() => timezoneOptionsWith(props.settings.CoinOfDayTimezone))
const sharedZoneLabel = computed(() => scheduleZoneLabel(props.settings.ScheduleTimezone))

const scheduleSummary = computed(() => describeSchedule({
  startTime: props.settings.CoinOfDayStartTime,
  defaultStartTime: '07:00',
  zone: props.settings.CoinOfDayTimezone || props.settings.ScheduleTimezone,
}))

const settingsMsg = ref('')
const settingsError = ref(false)
const timers: ReturnType<typeof setTimeout>[] = []
let pollTimer: ReturnType<typeof setInterval> | null = null

const {
  runs,
  page,
  loading,
  loadRuns,
  prevPage,
  nextPage,
} = useRunHistoryPagination<CoinOfDayRun>(async (currentPage, limit) => {
  const res = await getCoinOfDayRuns(currentPage, limit)
  return res.data ?? {}
})

function runIsTerminal(runStatus: string) {
  return runStatus === 'completed' || runStatus === 'failed'
}

function refreshPolling() {
  const hasActive = runs.value.some(run => !runIsTerminal(run.status))
  if (hasActive && !pollTimer) {
    pollTimer = setInterval(() => { loadRuns() }, 5000)
  } else if (!hasActive && pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }

  watch(runs, () => {
    refreshPolling()
  })
}

async function triggerManualCoinOfDay() {
  triggerLoading.value = true
  settingsMsg.value = ''
  settingsError.value = false
  try {
    const res = await triggerCoinOfDayRun()
    const runId = Number(res.data.runId ?? 0)
    settingsMsg.value = runId ? t('admin.schedules.cod.runQueued', { id: runId }) : t('admin.schedules.cod.queued')
    if (runId) {
      const detail = await getCoinOfDayRunDetail(runId)
      const run = detail.data
      if (runIsTerminal(run.status)) {
        settingsMsg.value = t('admin.schedules.cod.summary', { picked: run.picked, skipped: run.skipped }) + (run.errors ? t('admin.schedules.cod.summaryErrors', { count: run.errors }) : '')
        settingsError.value = run.status === 'failed'
      }
    }
    await loadRuns()
    refreshPolling()
    timers.push(setTimeout(() => { settingsMsg.value = '' }, 10000))
  } catch {
    settingsMsg.value = t('admin.schedules.cod.failed')
    settingsError.value = true
  } finally {
    triggerLoading.value = false
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString(intlLocale())
}

onMounted(() => {
  loadRuns()
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  timers.forEach(clearTimeout)
})
</script>
