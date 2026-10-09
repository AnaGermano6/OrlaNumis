<template>
  <hr class="my-6 border-0 border-t border-border-subtle" />

  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.wishlist.title') }}</h3>
  <p class="mb-4 text-base text-text-secondary">{{ t('admin.schedules.wishlist.intro') }}</p>
  <div class="mb-4">
    <div class="form-group flex items-center justify-between gap-3">
      <label class="form-label">{{ t('admin.schedules.enableChecks') }}</label>
      <BaseToggle
        size="sm"
        :label="t('admin.schedules.enableChecks')"
        :model-value="settings.WishlistSearchAlertsCheckEnabled === 'true'"
        @update:model-value="settings.WishlistSearchAlertsCheckEnabled = $event ? 'true' : 'false'"
      />
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('admin.schedules.startAnchor') }}</label>
      <input
        v-model="settings.WishlistSearchAlertsCheckStartTime"
        class="form-input w-full max-w-[120px]"
        type="time"
      />
      <span class="form-hint">{{ t('admin.schedules.wishlist.startHint') }}</span>
      <AdminScheduleSummary :start-time="settings.WishlistSearchAlertsCheckStartTime" default-start-time="03:00" :zone="settings.ScheduleTimezone" />
    </div>
    <div class="mt-4 flex w-full flex-col gap-3 md:flex-row md:items-center">
      <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="emit('save')">
        {{ settingsSaving ? t('common.saving') : t('admin.schedules.saveSchedule') }}
      </button>
    </div>
  </div>

  <hr class="my-6 border-0 border-t border-border-subtle" />
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.wishlist.history') }}</h3>

  <div v-if="loading" class="flex justify-center py-8"><div class="spinner"></div></div>
  <div v-else-if="historyError" class="px-8 py-8 text-center font-sans text-loss" role="alert">{{ historyError }}</div>
  <div v-else-if="runs.length === 0" class="px-8 py-8 text-center font-sans text-text-muted">{{ t('admin.schedules.wishlist.empty') }}</div>
  <template v-else>
    <div class="overflow-x-auto">
      <table class="data-table text-body [&_th]:px-2 [&_th]:py-3 [&_td]:px-2 [&_td]:py-3">
        <thead>
          <tr>
            <th>{{ t('common.date') }}</th>
            <th>{{ t('admin.schedules.wishlist.alert') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.wishlist.owner') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.trigger') }}</th>
            <th>{{ t('common.status') }}</th>
            <th>{{ t('admin.schedules.wishlist.results') }}</th>
            <th>{{ t('admin.schedules.wishlist.new') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.wishlist.duplicates') }}</th>
            <th>{{ t('admin.schedules.duration') }}</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="run in runs" :key="run.id">
            <tr>
              <td class="text-text-secondary">{{ formatDate(run.startedAt) }}</td>
              <td class="max-w-[180px] overflow-hidden text-ellipsis whitespace-nowrap" :title="run.alertName || t('admin.schedules.wishlist.alertNumber', { id: run.alertId })">
                {{ run.alertName || t('admin.schedules.wishlist.alertNumber', { id: run.alertId }) }}
              </td>
              <td class="hidden md:table-cell">{{ run.userName || t('admin.schedules.wishlist.userNumber', { id: run.userId }) }}</td>
              <td class="hidden md:table-cell">{{ runTriggerLabel(run.triggerType) }}</td>
              <td><span class="chip-sm" :class="statusClass(run.status)">{{ alertRunStatusLabel(run.status) }}</span></td>
              <td>{{ run.resultCount }}</td>
              <td class="font-semibold text-gain">{{ run.newCount }}</td>
              <td class="hidden md:table-cell">{{ run.duplicateCount }}</td>
              <td>{{ formatDuration(run.durationMs) }}</td>
            </tr>
            <tr v-if="run.errorMessage || run.partialWarnings.length > 0">
              <td colspan="9" class="italic text-text-secondary">
                {{ t('admin.schedules.runNumber', { id: run.id }) }}: {{ run.errorMessage || run.partialWarnings.join('; ') }}
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <div class="mt-4 flex items-center justify-center gap-3">
      <button class="btn btn-secondary btn-sm" :disabled="page <= 1" @click="prevPage()">{{ t('common.prev') }}</button>
      <span class="text-body text-text-secondary">{{ t('admin.schedules.page', { page }) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="page * pageLimit >= total" @click="nextPage()">{{ t('common.next') }}</button>
    </div>
  </template>
</template>

<script setup lang="ts">
import { BaseToggle } from '@/components/ui'
import AdminScheduleSummary from '@/components/admin/schedules/AdminScheduleSummary.vue'
import { onMounted, ref } from 'vue'
import { getAdminWishlistSearchAlertRuns } from '@/api/client'
import type { AdminWishlistSearchAlertRun, AppSettings } from '@/types'
import { t, intlLocale } from '@/i18n'
import { alertRunStatusLabel, runTriggerLabel } from '@/utils/format'

defineProps<{
  settings: AppSettings
  settingsSaving: boolean
}>()

const emit = defineEmits<{ save: [] }>()
const pageLimit = 5
const runs = ref<AdminWishlistSearchAlertRun[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const historyError = ref('')

async function loadRuns() {
  loading.value = true
  historyError.value = ''
  try {
    const response = await getAdminWishlistSearchAlertRuns(page.value, pageLimit)
    runs.value = response.data?.runs ?? []
    total.value = response.data?.total ?? 0
  } catch {
    runs.value = []
    total.value = 0
    historyError.value = t('admin.schedules.wishlist.loadFailed')
  } finally {
    loading.value = false
  }
}

async function prevPage() {
  page.value = Math.max(1, page.value - 1)
  await loadRuns()
}

async function nextPage() {
  page.value++
  await loadRuns()
}

function statusClass(status: AdminWishlistSearchAlertRun['status']): string {
  if (status === 'queued') return 'text-gold'
  if (status === 'running') return 'text-bronze'
  if (status === 'failed' || status === 'cancelled' || status === 'rate_limited') return 'text-loss'
  if (status === 'partial') return 'text-warning'
  return 'text-gain'
}

function formatDate(date: string): string {
  return new Date(date).toLocaleString(intlLocale())
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(1)}s`
}

onMounted(loadRuns)
</script>
