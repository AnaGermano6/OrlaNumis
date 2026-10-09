<template>
  <hr class="my-6 border-0 border-t border-border-subtle" />

  <!-- Auction Watch Bid Digest -->
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.digest.title') }}</h3>
  <p class="mb-4 text-base text-text-secondary">{{ t('admin.schedules.digest.intro') }}</p>
  <div class="mb-4">
    <div class="form-group flex items-center justify-between gap-3">
      <label class="form-label">{{ t('admin.schedules.digest.enable') }}</label>
      <BaseToggle
        size="sm"
        :label="t('admin.schedules.digest.enable')"
        :model-value="settings.AuctionWatchBidDigestEnabled === 'true'"
        @update:model-value="settings.AuctionWatchBidDigestEnabled = $event ? 'true' : 'false'"
      />
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('admin.schedules.startAnchor') }}</label>
      <input
        v-model="settings.AuctionWatchBidDigestStartTime"
        class="form-input w-full max-w-[120px]"
        type="time"
      />
      <span class="form-hint">{{ t('admin.schedules.digest.startHint') }}</span>
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('admin.schedules.repeatMinutes') }}</label>
      <input
        v-model="settings.AuctionWatchBidDigestInterval"
        class="form-input w-full max-w-[120px]"
        type="number"
        min="60"
        step="60"
      />
      <span class="form-hint">{{ t('admin.schedules.digest.intervalHint') }}</span>
      <AdminScheduleSummary :start-time="settings.AuctionWatchBidDigestStartTime" default-start-time="08:00" :interval-minutes="settings.AuctionWatchBidDigestInterval" :zone="settings.ScheduleTimezone" />
    </div>
    <div class="mt-4 flex w-full flex-col gap-3 md:flex-row md:items-center">
      <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="emit('save')">
        {{ settingsSaving ? t('common.saving') : t('admin.schedules.digest.save') }}
      </button>
      <span v-if="settingsMsg" class="text-body text-gold md:mr-auto" :class="settingsError ? 'text-loss' : ''">{{ settingsMsg }}</span>
      <button class="btn btn-secondary btn-sm md:ml-auto" :disabled="triggerLoading" @click="triggerManualDigest()">
        {{ triggerLoading ? t('common.starting') : t('admin.schedules.runNow') }}
      </button>
    </div>
  </div>

  <hr class="my-6 border-0 border-t border-border-subtle" />
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.digest.history') }}</h3>

  <div v-if="loading" class="flex justify-center py-8"><div class="spinner"></div></div>
  <div v-else-if="runs.length === 0" class="px-8 py-8 text-center font-sans text-text-muted">{{ t('admin.schedules.digest.empty') }}</div>
  <template v-else>
    <div class="overflow-x-auto">
      <table class="data-table text-chip md:table-fixed [&_th]:px-[0.35rem] [&_th]:py-2 md:[&_th]:px-2 md:[&_th]:py-3 [&_td]:px-[0.35rem] [&_td]:py-2 md:[&_td]:px-2 md:[&_td]:py-3">
        <thead>
          <tr>
            <th>{{ t('common.date') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.trigger') }}</th>
            <th>{{ t('admin.schedules.lots') }}</th>
            <th>{{ t('admin.schedules.digest.digests') }}</th>
            <th class="hidden md:table-cell">{{ t('common.status') }}</th>
            <th>{{ t('admin.schedules.duration') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="run in runs" :key="run.id">
            <td class="text-body text-text-secondary">{{ formatDate(run.startedAt) }}</td>
            <td class="hidden md:table-cell">
              <BaseStatusBadge :tone="run.triggerType === 'manual' ? 'error' : 'warning'">
                {{ runTriggerLabel(run.triggerType) }}
              </BaseStatusBadge>
            </td>
            <td>{{ run.lotsChecked }}</td>
            <td class="font-semibold text-gain">{{ run.digestsSent }}</td>
            <td class="hidden md:table-cell">
              <BaseStatusBadge :tone="run.status === 'error' ? 'error' : (run.status === 'success' ? 'success' : 'warning')">
                {{ runStatusLabel(run.status) }}
              </BaseStatusBadge>
            </td>
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
import { onMounted, onUnmounted, ref } from 'vue'
import { getAuctionWatchBidDigestRuns, triggerAuctionWatchBidDigest } from '@/api/client'
import { useRunHistoryPagination } from '@/composables/useRunHistoryPagination'
import type { AppSettings, AuctionWatchBidDigestRun } from '@/types'
import { t, intlLocale } from '@/i18n'
import { runStatusLabel, runTriggerLabel } from '@/utils/format'

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

const triggerLoading = ref(false)
const timers: ReturnType<typeof setTimeout>[] = []

const {
  runs,
  page,
  loading,
  loadRuns,
  prevPage,
  nextPage,
} = useRunHistoryPagination<AuctionWatchBidDigestRun>(async (currentPage, limit) => {
  const res = await getAuctionWatchBidDigestRuns(currentPage, limit)
  return res.data ?? {}
})

async function triggerManualDigest() {
  triggerLoading.value = true
  emit('update:settingsMsg', '')
  emit('update:settingsError', false)
  try {
    const res = await triggerAuctionWatchBidDigest()
    emit('update:settingsMsg', res.data.message ?? t('admin.schedules.digest.started'))
    timers.push(setTimeout(() => { emit('update:settingsMsg', '') }, 10000))
    timers.push(setTimeout(() => { loadRuns() }, 2000))
  } catch {
    emit('update:settingsMsg', t('admin.schedules.digest.failed'))
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
})

onUnmounted(() => {
  timers.forEach(clearTimeout)
})
</script>
