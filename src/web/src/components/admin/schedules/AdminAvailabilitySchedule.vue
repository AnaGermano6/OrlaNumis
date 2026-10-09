<template>
  <!-- Wishlist Availability Check -->
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.availability.title') }}</h3>
  <p class="mb-4 text-base text-text-secondary">{{ t('admin.schedules.availability.intro') }}</p>
  <div class="mb-4">
    <div class="form-group flex items-center justify-between gap-3">
      <label class="form-label">{{ t('admin.schedules.enableChecks') }}</label>
      <BaseToggle
        size="sm"
        :label="t('admin.schedules.enableChecks')"
        :model-value="settings.WishlistCheckEnabled === 'true'"
        @update:model-value="settings.WishlistCheckEnabled = $event ? 'true' : 'false'"
      />
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('admin.schedules.startAnchor') }}</label>
      <input
        v-model="settings.WishlistCheckStartTime"
        class="form-input w-full max-w-[120px]"
        type="time"
      />
      <span class="form-hint">{{ t('admin.schedules.availability.startHint') }}</span>
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('admin.schedules.repeatMinutes') }}</label>
      <input
        v-model="settings.WishlistCheckInterval"
        class="form-input w-full max-w-[120px]"
        type="number"
        min="5"
        step="5"
      />
      <span class="form-hint">{{ t('admin.schedules.availability.intervalHint') }}</span>
      <AdminScheduleSummary :start-time="settings.WishlistCheckStartTime" default-start-time="02:00" :interval-minutes="settings.WishlistCheckInterval" :zone="settings.ScheduleTimezone" />
    </div>
    <div class="mt-4 flex w-full flex-col gap-3 md:flex-row md:items-center">
      <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="emit('save')">
        {{ settingsSaving ? t('common.saving') : t('admin.schedules.saveSchedule') }}
      </button>
      <span v-if="settingsMsg" class="text-body text-gold md:mr-auto" :class="settingsError ? 'text-loss' : ''">{{ settingsMsg }}</span>
      <button class="btn btn-secondary btn-sm md:ml-auto" :disabled="triggerLoading" @click="triggerManualAvailabilityCheck()">
        {{ triggerLoading ? t('admin.schedules.availability.queuing') : t('admin.schedules.runNow') }}
      </button>
    </div>
  </div>

  <hr class="my-6 border-0 border-t border-border-subtle" />
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.availability.history') }}</h3>

  <div v-if="cyclesLoading" class="flex justify-center py-8"><div class="spinner"></div></div>
  <div v-else-if="cycles.length === 0" class="px-8 py-8 text-center font-sans text-text-muted">{{ t('admin.schedules.availability.empty') }}</div>
  <template v-else>
    <div class="overflow-x-auto">
      <table class="data-table text-chip md:table-fixed [&_th]:px-[0.35rem] [&_th]:py-2 md:[&_th]:px-2 md:[&_th]:py-3 [&_td]:px-[0.35rem] [&_td]:py-2 md:[&_td]:px-2 md:[&_td]:py-3">
        <thead>
          <tr>
            <th>{{ t('common.date') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.trigger') }}</th>
            <th>{{ t('common.status') }}</th>
            <th>{{ t('admin.schedules.availability.total') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.availability.queued') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.availability.running') }}</th>
            <th>{{ t('admin.schedules.availability.done') }}</th>
            <th>{{ t('admin.schedules.failed') }}</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="cycle in cycles" :key="cycle.id">
            <tr class="cursor-pointer transition-colors hover:bg-surface" :class="{ 'bg-surface': expandedCycleId === cycle.id }" @click="toggleCycleDetail(cycle.id)">
              <td class="text-body text-text-secondary">{{ formatDate(cycle.startedAt) }}</td>
              <td class="hidden md:table-cell">{{ runTriggerLabel(cycle.triggerType) }}</td>
              <td>
                <span class="chip-sm" :class="statusClass(cycle.status)">{{ availabilityStatusLabel(cycle.status) }}</span>
              </td>
              <td>{{ cycle.totalChildren }}</td>
              <td class="hidden md:table-cell">{{ cycle.queuedChildren }}</td>
              <td class="hidden md:table-cell">{{ cycle.runningChildren }}</td>
              <td class="font-semibold text-gain">{{ cycle.completedChildren }}</td>
              <td class="font-semibold text-loss">{{ cycle.failedChildren }}</td>
            </tr>
            <tr v-if="expandedCycleId === cycle.id" class="bg-surface-secondary">
              <td :colspan="cyclesColspan">
                <div v-if="expandedChildrenLoading" class="flex justify-center py-8"><div class="spinner"></div></div>
                <div v-else-if="expandedChildren.length" class="overflow-x-auto">
                  <table class="data-table text-chip md:table-fixed [&_th]:px-2 [&_th]:py-[0.4rem] [&_td]:px-2 [&_td]:py-[0.4rem] [&_td]:overflow-hidden [&_td]:text-ellipsis [&_td]:whitespace-nowrap">
                    <thead>
                      <tr>
                        <th>{{ t('admin.security.user') }}</th>
                        <th>{{ t('common.status') }}</th>
                        <th>{{ t('admin.schedules.valuation.checked') }}</th>
                        <th>{{ t('admin.schedules.availability.avail') }}</th>
                        <th>{{ t('admin.schedules.availability.unavail') }}</th>
                        <th>{{ t('admin.schedules.availability.unknown') }}</th>
                        <th>{{ t('admin.schedules.errors') }}</th>
                        <th>{{ t('admin.schedules.availability.failure') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="child in expandedChildren" :key="child.id">
                        <td>{{ child.userName ?? t('admin.schedules.wishlist.userNumber', { id: child.userId }) }}</td>
                        <td><span class="chip-sm" :class="statusClass(child.status)">{{ availabilityStatusLabel(child.status) }}</span></td>
                        <td>{{ child.coinsChecked }}</td>
                        <td>{{ child.available }}</td>
                        <td>{{ child.unavailable }}</td>
                        <td>{{ child.unknown }}</td>
                        <td>{{ child.errors }}</td>
                        <td class="max-w-[200px] overflow-hidden text-ellipsis whitespace-nowrap">{{ child.failMessage || '--' }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <p v-else class="px-8 py-8 text-center font-sans text-text-muted">{{ t('admin.schedules.availability.noChildren') }}</p>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <div class="mt-4 flex items-center justify-center gap-3">
      <button class="btn btn-secondary btn-sm" :disabled="cyclesPage <= 1" @click="prevCyclesPage()">{{ t('common.prev') }}</button>
      <span class="text-chip text-text-secondary">{{ t('admin.schedules.page', { page: cyclesPage }) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="cycles.length < 5" @click="nextCyclesPage()">{{ t('common.next') }}</button>
    </div>
  </template>

  <hr class="my-6 border-0 border-t border-border-subtle" />
  <h4 class="section-label mb-4 flex items-center gap-2">
    {{ t('admin.schedules.availability.legacyRuns') }}
    <span class="chip-sm">{{ t('admin.schedules.availability.legacy') }}</span>
  </h4>

  <div v-if="legacyLoading" class="flex justify-center py-8"><div class="spinner"></div></div>
  <div v-else-if="legacyRuns.length === 0" class="px-8 py-8 text-center font-sans text-text-muted">{{ t('admin.schedules.availability.noLegacy') }}</div>
  <template v-else>
    <div class="overflow-x-auto">
      <table class="data-table text-chip md:table-fixed [&_th]:px-[0.35rem] [&_th]:py-2 md:[&_th]:px-2 md:[&_th]:py-3 [&_td]:px-[0.35rem] [&_td]:py-2 md:[&_td]:px-2 md:[&_td]:py-3">
        <thead>
          <tr>
            <th>{{ t('common.date') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.trigger') }}</th>
            <th class="hidden md:table-cell">{{ t('common.status') }}</th>
            <th>{{ t('admin.schedules.valuation.checked') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.availability.avail') }}</th>
            <th>{{ t('admin.schedules.availability.unavail') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.availability.unknown') }}</th>
            <th class="hidden md:table-cell">{{ t('admin.schedules.errors') }}</th>
            <th>{{ t('admin.schedules.duration') }}</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="run in legacyRuns" :key="run.id">
            <tr class="cursor-pointer transition-colors hover:bg-surface" :class="{ 'bg-surface': expandedRunId === run.id }" @click="toggleRunDetail(run.id)">
              <td class="text-body text-text-secondary">
                {{ formatDate(run.startedAt) }}
                <span class="chip-sm ml-2">{{ t('admin.schedules.availability.legacy') }}</span>
              </td>
              <td class="hidden md:table-cell">{{ runTriggerLabel(run.triggerType) }}</td>
              <td class="hidden md:table-cell">
                <span class="chip-sm" :class="statusClass(run.status)">{{ availabilityStatusLabel(run.status) }}</span>
              </td>
              <td>{{ run.coinsChecked }}</td>
              <td class="hidden font-semibold text-gain md:table-cell">{{ run.available }}</td>
              <td class="font-semibold text-loss">{{ run.unavailable }}</td>
              <td class="hidden font-semibold text-warning md:table-cell">{{ run.unknown }}</td>
              <td class="hidden md:table-cell">{{ run.errors }}</td>
              <td>{{ formatDuration(run.durationMs) }}</td>
            </tr>
            <tr v-if="expandedRunId === run.id && expandedResults" class="bg-surface-secondary">
              <td :colspan="legacyColspan">
                <div v-if="expandedLoading" class="flex justify-center py-8"><div class="spinner"></div></div>
                <div v-else-if="expandedResults.length" class="overflow-x-auto">
                  <table class="data-table text-chip md:table-fixed [&_th]:px-2 [&_th]:py-[0.4rem] [&_td]:px-2 [&_td]:py-[0.4rem] [&_td]:overflow-hidden [&_td]:text-ellipsis [&_td]:whitespace-nowrap">
                    <thead>
                      <tr>
                        <th>{{ t('admin.schedules.valuation.coin') }}</th>
                        <th>URL</th>
                        <th>{{ t('common.status') }}</th>
                        <th>{{ t('admin.schedules.availability.reason') }}</th>
                        <th>HTTP</th>
                        <th>{{ t('admin.schedules.availability.agent') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="result in expandedResults" :key="result.id">
                        <td>{{ result.coinName }}</td>
                        <td>
                          <SafeExternalLink
                            v-if="safeRunUrl(result.url)"
                            :href="result.url"
                            target="_blank"
                            rel="noopener"
                            class="text-gold no-underline hover:underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                            @click.stop
                          >
                            {{ truncateUrl(result.url) }}
                          </SafeExternalLink>
                          <span v-else class="text-text-muted">--</span>
                        </td>
                        <td>
                          <BaseStatusBadge :tone="result.status === 'available' ? 'success' : result.status === 'unavailable' ? 'error' : 'warning'">{{ resultStatusLabel(result.status) }}</BaseStatusBadge>
                        </td>
                        <td class="max-w-[200px] overflow-hidden text-ellipsis whitespace-nowrap">{{ result.reason || '--' }}</td>
                        <td>{{ result.httpStatus ?? '--' }}</td>
                        <td>{{ result.agentUsed ? t('common.yes') : t('common.no') }}</td>
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
      <button class="btn btn-secondary btn-sm" :disabled="legacyPage <= 1" @click="prevLegacyPage()">{{ t('common.prev') }}</button>
      <span class="text-chip text-text-secondary">{{ t('admin.schedules.page', { page: legacyPage }) }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="legacyRuns.length < 5" @click="nextLegacyPage()">{{ t('common.next') }}</button>
    </div>
  </template>
</template>

<script setup lang="ts">
import { BaseStatusBadge, BaseToggle } from '@/components/ui'
import AdminScheduleSummary from '@/components/admin/schedules/AdminScheduleSummary.vue'
import { computed, onMounted, onUnmounted, ref } from 'vue'
import {
  getAvailabilityCycleDetail,
  getAvailabilityCycles,
  getAvailabilityRunDetail,
  getAvailabilityRuns,
  triggerAvailabilityCheck,
} from '@/api/client'
import SafeExternalLink from '@/components/SafeExternalLink.vue'
import { sanitizeExternalUrl } from '@/composables/useSafeExternalLink'
import type { AppSettings, AvailabilityCycle, AvailabilityRun } from '@/types'
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

const isMobile = ref(window.innerWidth <= 600)
const cyclesColspan = computed(() => isMobile.value ? 5 : 8)
const legacyColspan = computed(() => isMobile.value ? 4 : 9)
const triggerLoading = ref(false)
const timers: ReturnType<typeof setTimeout>[] = []
let pollTimer: ReturnType<typeof setInterval> | null = null
const CYCLES_LIMIT = 5
const LEGACY_LIMIT = 5

// Parent availability cycles (new Feature 353 roll-up rows)
const cycles = ref<AvailabilityCycle[]>([])
const cyclesPage = ref(1)
const cyclesLoading = ref(false)
const expandedCycleId = ref<number | null>(null)
const expandedChildren = ref<AvailabilityRun[]>([])
const expandedChildrenLoading = ref(false)

// Legacy UserID=0 aggregate rows, retained read-only from /admin/availability-runs
const legacyRuns = ref<AvailabilityRun[]>([])
const legacyPage = ref(1)
const legacyLoading = ref(false)
const expandedRunId = ref<number | null>(null)
const expandedResults = ref<AvailabilityRun['results']>(undefined)
const expandedLoading = ref(false)

function safeRunUrl(url: string | null | undefined): string | null {
  return sanitizeExternalUrl(url)
}

function onResize() {
  isMobile.value = window.innerWidth <= 600
}

function availabilityStatusLabel(status: string): string {
  return runStatusLabel(status === 'completed' ? 'done' : status)
}

function resultStatusLabel(status: string): string {
  if (status === 'available') return t('availability.resultStatus.available')
  if (status === 'unavailable') return t('availability.resultStatus.unavailable')
  if (status === 'unknown') return t('availability.resultStatus.unknown')
  return status
}

function statusClass(status: string): string {
  if (status === 'queued') return 'text-gold'
  if (status === 'running') return 'text-bronze'
  if (status === 'failed' || status === 'partial_failure') return 'text-loss'
  return 'text-gain'
}

async function loadCycles() {
  cyclesLoading.value = true
  try {
    const res = await getAvailabilityCycles(cyclesPage.value, CYCLES_LIMIT)
    cycles.value = res.data?.cycles ?? []
  } catch {
    cycles.value = []
  } finally {
    cyclesLoading.value = false
  }
}

async function prevCyclesPage() {
  cyclesPage.value = Math.max(1, cyclesPage.value - 1)
  await loadCycles()
}

async function nextCyclesPage() {
  cyclesPage.value++
  await loadCycles()
}

async function toggleCycleDetail(cycleId: number) {
  if (expandedCycleId.value === cycleId) {
    expandedCycleId.value = null
    expandedChildren.value = []
    return
  }
  expandedCycleId.value = cycleId
  expandedChildren.value = []
  expandedChildrenLoading.value = true
  try {
    const res = await getAvailabilityCycleDetail(cycleId)
    expandedChildren.value = res.data?.children ?? []
  } catch {
    expandedChildren.value = []
  } finally {
    expandedChildrenLoading.value = false
  }
}

async function loadLegacyRuns() {
  legacyLoading.value = true
  try {
    const res = await getAvailabilityRuns(legacyPage.value, LEGACY_LIMIT)
    legacyRuns.value = (res.data?.runs ?? []).filter((run) => run.userId === 0)
  } catch {
    legacyRuns.value = []
  } finally {
    legacyLoading.value = false
  }
}

async function prevLegacyPage() {
  legacyPage.value = Math.max(1, legacyPage.value - 1)
  await loadLegacyRuns()
}

async function nextLegacyPage() {
  legacyPage.value++
  await loadLegacyRuns()
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
    const res = await getAvailabilityRunDetail(runId)
    expandedResults.value = res.data.results ?? []
  } catch {
    expandedResults.value = []
  } finally {
    expandedLoading.value = false
  }
}

async function loadHistoryWithPoll() {
  try {
    await Promise.all([loadCycles(), loadLegacyRuns()])
    const hasActive = cycles.value.some(cycle => cycle.status === 'queued' || cycle.status === 'running')
    if (hasActive && !pollTimer) {
      pollTimer = setInterval(() => { loadHistoryWithPoll() }, 4000)
    } else if (!hasActive && pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  } catch { /* ignore */ }
}

async function triggerManualAvailabilityCheck() {
  triggerLoading.value = true
  emit('update:settingsMsg', '')
  emit('update:settingsError', false)
  try {
    const res = await triggerAvailabilityCheck()
    emit('update:settingsMsg', t('admin.schedules.availability.cycleQueued', { id: res.data.cycleId }))
    timers.push(setTimeout(() => { emit('update:settingsMsg', '') }, 12000))
    timers.push(setTimeout(() => { loadHistoryWithPoll() }, 1000))
  } catch (err: unknown) {
    const status = (err as { response?: { status?: number } })?.response?.status
    emit('update:settingsMsg', status === 409
      ? t('admin.schedules.availability.inProgress')
      : t('admin.schedules.availability.queueFailed'))
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

function truncateUrl(url: string) {
  try {
    const parsed = new URL(url)
    const path = parsed.pathname.length > 20 ? parsed.pathname.substring(0, 17) + '...' : parsed.pathname
    return parsed.hostname + path
  } catch {
    if (url.length <= 35) return url
    return url.substring(0, 32) + '...'
  }
}

onMounted(() => {
  window.addEventListener('resize', onResize)
  loadHistoryWithPoll()
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  if (pollTimer) clearInterval(pollTimer)
  timers.forEach(clearTimeout)
})
</script>
