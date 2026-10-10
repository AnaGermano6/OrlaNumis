<template>
  <hr class="my-6 border-0 border-t border-border-subtle" />

  <!-- Collection Valuation -->
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.valuation.title') }}</h3>
  <div class="mb-4">
    <div class="form-group flex items-center justify-between gap-3">
      <label class="form-label">{{ t('admin.schedules.valuation.enable') }}</label>
      <BaseToggle
        size="sm"
        :label="t('admin.schedules.valuation.enable')"
        :model-value="settings.ValuationCheckEnabled === 'true'"
        @update:model-value="settings.ValuationCheckEnabled = $event ? 'true' : 'false'"
      />
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('admin.schedules.startAnchor') }}</label>
      <input
        v-model="settings.ValuationCheckStartTime"
        class="form-input w-full max-w-[120px]"
        type="time"
      />
      <span class="form-hint">{{ t('admin.schedules.valuation.startHint') }}</span>
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('admin.schedules.valuation.repeatDays') }}</label>
      <input
        v-model="settings.ValuationCheckIntervalDays"
        class="form-input w-full max-w-[120px]"
        type="number"
        min="1"
        step="1"
      />
      <span class="form-hint">{{ t('admin.schedules.valuation.repeatHint') }}</span>
      <AdminScheduleSummary :start-time="settings.ValuationCheckStartTime" default-start-time="03:00" :interval-days="settings.ValuationCheckInterval" :zone="settings.ScheduleTimezone" />
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('admin.schedules.valuation.maxCoins') }}</label>
      <input
        v-model="settings.ValuationMaxCoins"
        class="form-input w-full max-w-[120px]"
        type="number"
        min="1"
        step="10"
      />
      <span class="form-hint">{{ t('admin.schedules.valuation.maxCoinsHint') }}</span>
    </div>
    <div class="mt-4 flex w-full flex-col gap-3 md:flex-row md:items-center">
      <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="emit('save')">
        {{ settingsSaving ? t('common.saving') : t('admin.schedules.valuation.save') }}
      </button>
      <span v-if="settingsMsg" class="text-body text-gold md:mr-auto" :class="settingsError ? 'text-loss' : ''">{{ settingsMsg }}</span>
      <button class="btn btn-secondary btn-sm md:ml-auto" :disabled="triggerLoading" @click="triggerManualValuation()">
        {{ triggerLoading ? t('common.starting') : t('admin.schedules.runNow') }}
      </button>
    </div>
  </div>

  <hr class="my-6 border-0 border-t border-border-subtle" />
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.valuation.history') }}</h3>

  <div v-if="loading" class="flex justify-center py-8"><div class="spinner"></div></div>
  <div v-else-if="runs.length === 0" class="px-8 py-8 text-center font-sans text-text-muted">{{ t('admin.schedules.valuation.empty') }}</div>
  <template v-else>
    <div class="overflow-x-auto">
      <table class="data-table text-chip md:table-fixed [&_th]:px-[0.35rem] [&_th]:py-2 md:[&_th]:px-2 md:[&_th]:py-3 [&_td]:px-[0.35rem] [&_td]:py-2 md:[&_td]:px-2 md:[&_td]:py-3">
        <thead>
          <tr>
            <th>{{ t('common.date') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.trigger') }}</th>
            <th>{{ t('common.status') }}</th>
            <th>{{ t('admin.schedules.valuation.checked') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.valuation.updated') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.skipped') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.errors') }}</th>
            <th>{{ t('admin.schedules.duration') }}</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="run in runs" :key="run.id">
            <tr class="cursor-pointer transition-colors hover:bg-surface" :class="{ 'bg-surface': expandedRunId === run.id }" @click="toggleRunDetail(run.id)">
              <td class="text-body text-text-secondary">{{ formatDate(run.startedAt) }}</td>
              <td class="hidden md:table-cell">{{ runTriggerLabel(run.triggerType) }}</td>
              <td>
                <BaseStatusBadge :tone="run.status === 'running' ? 'info' : run.status === 'completed' ? 'success' : run.status === 'failed' ? 'error' : 'warning'">{{ runStatusLabel(run.status) }}</BaseStatusBadge>
                <span v-if="run.status === 'running' && run.totalCoins > 0" class="ml-[0.35rem] text-label font-medium text-text-secondary">
                  {{ run.coinsChecked + run.coinsSkipped + run.errors }} / {{ run.totalCoins }}
                </span>
                <button v-if="run.status === 'running'" class="ml-[0.4rem] rounded-full border border-status-error-border bg-transparent px-[0.4rem] py-[0.1rem] text-micro text-loss transition-colors hover:bg-status-error-bg" @click.stop="cancelRun(run.id)">{{ t('common.cancel') }}</button>
              </td>
              <td>{{ run.coinsChecked }}</td>
              <td class="hidden font-semibold text-gain md:table-cell">{{ run.coinsUpdated }}</td>
              <td class="hidden font-semibold text-warning md:table-cell">{{ run.coinsSkipped }}</td>
              <td class="hidden font-semibold text-loss md:table-cell">{{ run.errors }}</td>
              <td>{{ formatDuration(run.durationMs) }}</td>
            </tr>
            <tr v-if="expandedRunId === run.id && expandedResults" class="bg-surface-secondary">
              <td :colspan="colspan">
                <div v-if="expandedLoading" class="flex justify-center py-8"><div class="spinner"></div></div>
                <div v-else-if="expandedResults.length" class="overflow-x-auto">
                  <table class="data-table text-chip md:table-fixed [&_th]:px-2 [&_th]:py-[0.4rem] [&_td]:px-2 [&_td]:py-[0.4rem] [&_td]:overflow-hidden [&_td]:text-ellipsis [&_td]:whitespace-nowrap">
                    <thead>
                      <tr>
                        <th>{{ t('admin.schedules.valuation.coin') }}</th>
                        <th>{{ t('admin.schedules.valuation.previous') }}</th>
                        <th>{{ t('admin.schedules.valuation.estimated') }}</th>
                        <th>{{ t('admin.schedules.valuation.confidence') }}</th>
                        <th>{{ t('common.status') }}</th>
                        <th>{{ t('admin.schedules.valuation.explanation') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="result in expandedResults" :key="result.id">
                        <td>{{ result.coinName }}</td>
                        <td>{{ result.previousValue != null ? formatCurrency(result.previousValue) : '--' }}</td>
                        <td class="font-semibold text-gold">{{ result.estimatedValue > 0 ? formatCurrency(result.estimatedValue) : '--' }}</td>
                        <td>
                          <span v-if="result.confidence" class="inline-block rounded-sm px-[0.3rem] py-[0.1rem] text-label font-semibold" :class="result.confidence === 'high' ? 'bg-status-success-bg text-confidence-high' : result.confidence === 'medium' ? 'bg-status-warning-bg text-confidence-medium' : 'bg-status-error-bg text-confidence-low'">{{ confidenceLabel(result.confidence) }}</span>
                          <span v-else>--</span>
                        </td>
                        <td>
                          <BaseStatusBadge :tone="result.status === 'success' ? 'success' : result.status === 'skipped' ? 'neutral' : 'error'">{{ runStatusLabel(result.status) }}</BaseStatusBadge>
                        </td>
                        <td class="max-w-[200px] overflow-hidden text-ellipsis whitespace-nowrap">
                          <div v-if="result.changeExplanation" class="mb-[0.35rem] font-medium text-gold">{{ result.changeExplanation }}</div>
                          <div>{{ result.reasoning || result.errorMessage || '--' }}</div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <p v-else class="px-8 py-8 text-center font-sans text-text-muted">{{ t('admin.schedules.valuation.noResults') }}</p>
              </td>
            </tr>
          </template>
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
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { cancelValuationRun, getValuationRunDetail, getValuationRuns, triggerValuation } from '@/api/client'
import { useRunHistoryPagination } from '@/composables/useRunHistoryPagination'
import type { AppSettings, ValuationRun } from '@/types'
import { t, intlLocale } from '@/i18n'
import { confidenceLabel, formatCurrency, runStatusLabel, runTriggerLabel } from '@/utils/format'

defineProps<{
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

const isMobile = ref(window.innerWidth <= 600)
const colspan = computed(() => isMobile.value ? 4 : 8)
const triggerLoading = ref(false)
const expandedRunId = ref<number | null>(null)
const expandedResults = ref<ValuationRun['results']>(undefined)
const expandedLoading = ref(false)
const timers: ReturnType<typeof setTimeout>[] = []
let pollTimer: ReturnType<typeof setInterval> | null = null

const {
  runs,
  page,
  loading,
  loadRuns: loadRunsBase,
  prevPage,
  nextPage,
} = useRunHistoryPagination<ValuationRun>(async (currentPage, limit) => {
  const res = await getValuationRuns(currentPage, limit)
  return res.data ?? {}
})

function onResize() {
  isMobile.value = window.innerWidth <= 600
}

async function loadRuns() {
  try {
    await loadRunsBase()
    const hasRunning = runs.value.some(run => run.status === 'running')
    if (hasRunning && !pollTimer) {
      pollTimer = setInterval(() => { loadRuns() }, 5000)
    } else if (!hasRunning && pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  } catch { /* ignore */ }
}

async function toggleRunDetail(runId: number) {
  if (expandedRunId.value === runId) {
    expandedRunId.value = null
    expandedResults.value = undefined
    return
  }
  expandedRunId.value = runId
  expandedResults.value = []
  expandedLoading.value = true
  try {
    const res = await getValuationRunDetail(runId)
    expandedResults.value = res.data.results ?? []
  } catch {
    expandedResults.value = []
  } finally {
    expandedLoading.value = false
  }
}

async function triggerManualValuation() {
  triggerLoading.value = true
  emit('update:settingsMsg', '')
  emit('update:settingsError', false)
  try {
    await triggerValuation()
    emit('update:settingsMsg', t('admin.schedules.valuation.started'))
    timers.push(setTimeout(() => { emit('update:settingsMsg', '') }, 10000))
    timers.push(setTimeout(() => { loadRuns() }, 2000))
  } catch {
    emit('update:settingsMsg', t('admin.schedules.valuation.failed'))
    emit('update:settingsError', true)
  } finally {
    triggerLoading.value = false
  }
}

async function cancelRun(runId: number) {
  try {
    await cancelValuationRun(runId)
    emit('update:settingsMsg', t('admin.schedules.valuation.cancelRequested'))
    timers.push(setTimeout(() => { emit('update:settingsMsg', '') }, 5000))
    timers.push(setTimeout(() => { loadRuns() }, 1000))
  } catch {
    emit('update:settingsMsg', t('admin.schedules.valuation.cancelFailed'))
    emit('update:settingsError', true)
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
  window.addEventListener('resize', onResize)
  loadRuns()
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  if (pollTimer) clearInterval(pollTimer)
  timers.forEach(clearTimeout)
})
</script>
