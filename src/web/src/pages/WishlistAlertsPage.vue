<template>
  <div class="container">
    <DesktopPageContext :title="t('wishlistAlerts.title')">
      <template #actions>
        <AppIconButton :title="t('wishlistAlerts.newAlert')" @click="startCreate">
          <Search :size="20" />
        </AppIconButton>
      </template>
      <div class="page-header">
        <div>
          <h1>{{ t('wishlistAlerts.title') }}</h1>
          <p class="mt-1 text-body text-text-muted">{{ t('wishlistAlerts.intro') }}</p>
        </div>
        <div v-if="isPwa" class="pwa-actions">
          <button class="pwa-icon-btn focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" type="button" :title="t('wishlistAlerts.newAlert')" @click="startCreate">
            <Search :size="22" />
          </button>
        </div>
        <button v-else class="btn btn-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" type="button" @click="startCreate">
          <Search :size="16" /> {{ t('wishlistAlerts.newAlert') }}
        </button>
      </div>
      <template #desktop>
        <p class="mb-4 text-body text-text-muted">{{ t('wishlistAlerts.intro') }}</p>
      </template>
    </DesktopPageContext>

    <p v-if="error" class="mb-4 text-body text-bronze">{{ error }}</p>
    <div v-if="loading" class="loading-overlay"><div class="spinner"></div></div>
    <div v-else-if="!alerts.length" class="empty-state">
      <h3>{{ t('wishlistAlerts.empty.title') }}</h3>
      <p>{{ t('wishlistAlerts.empty.body') }}</p>
      <button
        class="btn btn-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
        type="button"
        @click="startCreate"
      >
        <Search :size="16" /> {{ t('wishlistAlerts.create') }}
      </button>
    </div>

    <div
      v-else
      class="grid items-start gap-4"
      :class="{ 'md:grid-cols-[minmax(260px,_0.9fr)_minmax(0,_1.6fr)]': selectedAlert }"
    >
      <aside class="grid gap-4" :aria-label="t('wishlistAlerts.listLabel')">
        <article
          v-for="alert in alerts"
          :key="alert.id"
          :class="[
            'rounded-md border bg-card p-4',
            selectedAlert?.id === alert.id ? 'border-gold shadow-glow' : 'border-border-subtle',
          ]"
        >
          <button
            class="w-full rounded-sm bg-transparent p-0 text-left text-inherit focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
            type="button"
            @click="selectAlert(alert)"
          >
            <h2>{{ alert.name }}</h2>
            <AlertCriteriaSummary :alert="alert" />
            <p class="mt-1 text-body text-text-muted">{{ alert.cadence === 'manual' ? t('wishlistAlerts.manualCadence') : t('wishlistAlerts.autoCadence', { cadence: alertCadenceLabel(alert.cadence).toLowerCase() }) }}</p>
          </button>
          <div class="mt-4 flex flex-wrap items-start justify-between gap-3">
            <button class="btn btn-secondary btn-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" type="button" @click="edit(alert)"><Pencil :size="14" /> {{ t('common.edit') }}</button>
            <button class="btn btn-secondary btn-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" type="button" @click="toggle(alert)">{{ alert.isActive ? t('common.disable') : t('common.enable') }}</button>
            <button class="btn btn-danger btn-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" type="button" @click="remove(alert)"><Trash2 :size="14" /> {{ t('common.delete') }}</button>
          </div>
        </article>
      </aside>

      <main v-if="selectedAlert" class="grid gap-4 border-t border-border-accent pt-4 md:border-t-0 md:border-l md:pl-4 md:pt-0">
        <section class="flex flex-wrap items-start justify-between gap-3 rounded-md border border-border-subtle bg-card p-4">
          <div>
            <p class="section-label">{{ t('wishlistAlerts.selected') }}</p>
            <h2>{{ selectedAlert.name }}</h2>
            <p class="mt-1 text-body text-text-muted">{{ t('wishlistAlerts.queueHint') }}</p>
          </div>
          <button
            class="btn btn-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
            type="button"
            :disabled="running || !selectedAlert.isActive"
            @click="runNow"
          >
            <Play :size="16" /> {{ running ? t('wishlistAlerts.running') : t('wishlistAlerts.runNow') }}
          </button>
        </section>
        <p v-if="!selectedAlert.isActive" class="m-0 text-body text-text-muted">{{ t('wishlistAlerts.enableFirst') }}</p>
        <p v-if="runMessage" class="m-0 text-body text-text-muted">{{ runMessage }}</p>

        <AlertRunHistory
          :runs="runs"
          :selected-run="selectedRun"
          :selected-run-id="selectedRun?.id ?? null"
          :loading="runsLoading"
          :error="runsError"
          @select="loadRunDetail"
          @refresh="loadRuns"
        />

        <section class="grid gap-4 rounded-md border border-border-subtle bg-card p-4">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h3>{{ t('wishlistAlerts.candidateReview') }}</h3>
              <p class="mt-1 text-body text-text-muted">{{ t('wishlistAlerts.candidateReviewHint') }}</p>
            </div>
            <div class="flex flex-wrap gap-3">
              <select
                v-model="candidateState"
                class="form-select min-w-[12rem] focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
                @change="loadCandidates"
              >
                <option value="">{{ t('wishlistAlerts.filterActiveAndReview') }}</option>
                <option value="active">{{ candidateStateLabel('active') }}</option>
                <option value="needs_review">{{ candidateStateLabel('needs_review') }}</option>
                <option value="dismissed">{{ candidateStateLabel('dismissed') }}</option>
                <option value="converted">{{ candidateStateLabel('converted') }}</option>
                <option value="suppressed">{{ candidateStateLabel('suppressed') }}</option>
              </select>
              <select
                v-model="provenanceStatus"
                class="form-select min-w-[12rem] focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
                @change="loadCandidates"
              >
                <option value="">{{ t('wishlistAlerts.anyProvenance') }}</option>
                <option value="verified">{{ provenanceStatusLabel('verified') }}</option>
                <option value="partial">{{ provenanceStatusLabel('partial') }}</option>
                <option value="unverified">{{ provenanceStatusLabel('unverified') }}</option>
              </select>
            </div>
          </div>

          <p v-if="candidatesError" class="text-body text-bronze">{{ candidatesError }}</p>
          <div v-if="candidatesLoading" class="text-body text-text-muted">{{ t('wishlistAlerts.loadingCandidates') }}</div>
          <div v-else-if="!candidates.length" class="rounded-sm border border-dashed border-border-subtle p-4 text-center text-body text-text-muted">{{ t('wishlistAlerts.noCandidates') }}</div>
          <div v-else class="grid gap-4">
            <CandidateReviewCard
              v-for="candidate in candidates"
              :key="candidate.id"
              :candidate="candidate"
              :busy="candidateBusyId === candidate.id"
              :duplicate-warnings="duplicateWarnings[candidate.id] ?? []"
              @dismiss="dismissCandidate"
              @restore="restoreCandidate"
              @convert="convertCandidate"
            />
          </div>

          <details class="rounded-md border border-border-subtle bg-card p-4">
            <summary class="cursor-pointer text-gold focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]">{{ t('wishlistAlerts.adjustCriteria') }}</summary>
            <p class="mt-2 text-body text-text-muted">{{ t('wishlistAlerts.adjustHint') }}</p>
            <AlertForm class="mt-3" :alert="selectedAlert" :saving="adjusting" @save="adjustCriteria" @cancel="noop" />
          </details>
        </section>
      </main>
    </div>

    <section v-if="creating || editing" class="mt-4 rounded-md border border-border-subtle bg-card p-4">
      <h2>{{ editing ? t('wishlistAlerts.edit') : t('wishlistAlerts.create') }}</h2>
      <AlertForm :alert="editing" :saving="saving" @save="save" @cancel="closeEditor" />
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { Search, Pencil, Trash2, Play } from 'lucide-vue-next'
import AlertCriteriaSummary from '@/components/wishlist-alerts/AlertCriteriaSummary.vue'
import AlertForm from '@/components/wishlist-alerts/AlertForm.vue'
import AlertRunHistory from '@/components/wishlist-alerts/AlertRunHistory.vue'
import CandidateReviewCard from '@/components/wishlist-alerts/CandidateReviewCard.vue'
import { usePwa } from '@/composables/usePwa'
import {
  adjustWishlistSearchAlertCriteria,
  convertWishlistSearchAlertCandidate,
  createWishlistSearchAlert,
  deleteWishlistSearchAlert,
  dismissWishlistSearchAlertCandidate,
  getApiErrorMessage,
  getWishlistSearchAlertRun,
  listWishlistSearchAlertCandidates,
  listWishlistSearchAlertRuns,
  listWishlistSearchAlerts,
  restoreWishlistSearchAlertCandidate,
  runWishlistSearchAlert,
  updateWishlistSearchAlert,
} from '@/api/client'
import type { AlertCandidate, AlertCandidateState, AlertRun, CandidateDismissalReason, CandidateProvenanceStatus, CoinMutationPayload, WishlistSearchAlert, WishlistSearchAlertInput } from '@/types'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { alertCadenceLabel, alertRunStatusLabel, candidateStateLabel, provenanceStatusLabel } from '@/utils/format'
import { t } from '@/i18n'

const alerts = ref<WishlistSearchAlert[]>([])
const selectedAlertId = ref<number | null>(null)
const loading = ref(false)
const saving = ref(false)
const running = ref(false)
const adjusting = ref(false)
const error = ref('')
const runMessage = ref('')
const editing = ref<WishlistSearchAlert | null>(null)
const creating = ref(false)
const runs = ref<AlertRun[]>([])
const selectedRun = ref<AlertRun | null>(null)
const runsLoading = ref(false)
const runsError = ref('')
const candidates = ref<AlertCandidate[]>([])
const candidatesLoading = ref(false)
const candidatesError = ref('')
const candidateState = ref<AlertCandidateState | ''>('')
const provenanceStatus = ref<CandidateProvenanceStatus | ''>('')
const candidateBusyId = ref<number | null>(null)
const duplicateWarnings = ref<Record<number, string[]>>({})
const { isPwa } = usePwa()
const RUN_POLL_INTERVAL_MS = 3000
let runPollTimer: ReturnType<typeof setTimeout> | null = null

const selectedAlert = computed(() => alerts.value.find((alert) => alert.id === selectedAlertId.value) ?? null)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await listWishlistSearchAlerts({ page: 1, limit: 100 })
    alerts.value = res.data.alerts
    if (selectedAlertId.value && !alerts.value.some((alert) => alert.id === selectedAlertId.value)) clearSelection()
    if (selectedAlertId.value) await loadSelectedData()
    else clearReviewState()
  } catch (err) {
    error.value = getApiErrorMessage(err) || t('wishlistAlerts.errors.load')
  } finally {
    loading.value = false
  }
}

async function loadSelectedData() {
  await Promise.all([loadRuns(), loadCandidates()])
}

function startCreate() { creating.value = true; editing.value = null }
function edit(alert: WishlistSearchAlert) { editing.value = alert; creating.value = false }
function closeEditor() { creating.value = false; editing.value = null }
function noop() {}

function clearReviewState() {
  clearRunPollTimer()
  running.value = false
  selectedRun.value = null
  runs.value = []
  runsError.value = ''
  candidates.value = []
  candidatesError.value = ''
  duplicateWarnings.value = {}
  runMessage.value = ''
}

function clearSelection() {
  selectedAlertId.value = null
  clearReviewState()
}

async function selectAlert(alert: WishlistSearchAlert) {
  clearRunPollTimer()
  selectedAlertId.value = alert.id
  clearReviewState()
  await loadSelectedData()
}

async function save(input: WishlistSearchAlertInput) {
  saving.value = true
  error.value = ''
  try {
    if (editing.value) await updateWishlistSearchAlert(editing.value.id, input)
    else await createWishlistSearchAlert(input)
    closeEditor()
    await load()
  } catch (err) {
    error.value = getApiErrorMessage(err) || t('wishlistAlerts.errors.save')
  } finally {
    saving.value = false
  }
}

async function toggle(alert: WishlistSearchAlert) {
  await updateWishlistSearchAlert(alert.id, {
    name: alert.name,
    criteria: toInputCriteria(alert),
    cadence: alert.cadence,
    isActive: !alert.isActive,
  })
  await load()
}

async function remove(alert: WishlistSearchAlert) {
  if (!confirm(t('wishlistAlerts.confirmDelete', { name: alert.name }))) return
  await deleteWishlistSearchAlert(alert.id)
  if (selectedAlertId.value === alert.id) clearSelection()
  await load()
}

async function runNow() {
  if (!selectedAlert.value) return
  running.value = true
  runMessage.value = ''
  error.value = ''
  let queued = false
  try {
    const alertId = selectedAlert.value.id
    const res = await runWishlistSearchAlert(selectedAlert.value.id, 20)
    queued = true
    runMessage.value = t('wishlistAlerts.runQueued')
    await Promise.all([load(), loadRuns()])
    if (res.data.runId) await pollAlertRun(alertId, res.data.runId)
  } catch (err) {
    error.value = getApiErrorMessage(err) || t('wishlistAlerts.errors.discovery')
    running.value = false
  } finally {
    if (!queued) running.value = false
  }
}

async function loadRuns() {
  if (!selectedAlertId.value) return
  runsLoading.value = true
  runsError.value = ''
  try {
    const res = await listWishlistSearchAlertRuns(selectedAlertId.value, { page: 1, limit: 20 })
    runs.value = res.data.runs
    if (!selectedRun.value && runs.value[0]) {
      const run = await loadRunDetail(runs.value[0].id)
      if (run && !isTerminalRunStatus(run.status)) {
        running.value = true
        scheduleRunPoll(run.alertId, run.id)
      }
    }
  } catch (err) {
    runsError.value = getApiErrorMessage(err) || t('wishlistAlerts.errors.runs')
  } finally {
    runsLoading.value = false
  }
}

async function loadRunDetail(runId: number) {
  if (!selectedAlertId.value) return null
  const res = await getWishlistSearchAlertRun(selectedAlertId.value, runId)
  selectedRun.value = res.data
  return res.data
}

async function loadCandidates() {
  if (!selectedAlertId.value) return
  candidatesLoading.value = true
  candidatesError.value = ''
  try {
    const res = await listWishlistSearchAlertCandidates(selectedAlertId.value, { state: candidateState.value, provenanceStatus: provenanceStatus.value, page: 1, limit: 50 })
    candidates.value = res.data.candidates
  } catch (err) {
    candidatesError.value = getApiErrorMessage(err) || t('wishlistAlerts.errors.candidates')
  } finally {
    candidatesLoading.value = false
  }
}

async function dismissCandidate(candidate: AlertCandidate, reason: CandidateDismissalReason, notes: string) {
  if (!selectedAlertId.value) return
  candidateBusyId.value = candidate.id
  try {
    await dismissWishlistSearchAlertCandidate(selectedAlertId.value, candidate.id, { reason, notes })
    await loadCandidates()
  } catch (err) {
    candidatesError.value = getApiErrorMessage(err) || t('wishlistAlerts.errors.dismiss')
  } finally {
    candidateBusyId.value = null
  }
}

async function restoreCandidate(candidate: AlertCandidate) {
  if (!selectedAlertId.value) return
  candidateBusyId.value = candidate.id
  try {
    await restoreWishlistSearchAlertCandidate(selectedAlertId.value, candidate.id)
    await loadCandidates()
  } catch (err) {
    candidatesError.value = getApiErrorMessage(err) || t('wishlistAlerts.errors.restore')
  } finally {
    candidateBusyId.value = null
  }
}

async function convertCandidate(candidate: AlertCandidate, coin: CoinMutationPayload, acknowledgeDuplicateWarning: boolean) {
  if (!selectedAlertId.value) return
  candidateBusyId.value = candidate.id
  candidatesError.value = ''
  try {
    const res = await convertWishlistSearchAlertCandidate(selectedAlertId.value, candidate.id, { coin, acknowledgeDuplicateWarning })
    duplicateWarnings.value[candidate.id] = res.data.warnings ?? []
    await loadCandidates()
  } catch (err) {
    const response = (err as { response?: { status?: number; data?: { warnings?: string[] } } }).response
    if (response?.status === 409 && response.data?.warnings?.length) duplicateWarnings.value[candidate.id] = response.data.warnings
    candidatesError.value = getApiErrorMessage(err) || t('wishlistAlerts.errors.convert')
  } finally {
    candidateBusyId.value = null
  }
}

async function adjustCriteria(input: WishlistSearchAlertInput) {
  if (!selectedAlertId.value) return
  adjusting.value = true
  try {
    const candidateIds = candidates.value.slice(0, 20).map((candidate) => candidate.id)
    await adjustWishlistSearchAlertCriteria(selectedAlertId.value, { candidateIds, criteria: input.criteria })
    await load()
  } catch (err) {
    error.value = getApiErrorMessage(err) || t('wishlistAlerts.errors.adjust')
  } finally {
    adjusting.value = false
  }
}

function toInputCriteria(alert: WishlistSearchAlert): WishlistSearchAlertInput['criteria'] {
  return {
    rulerOrIssuer: alert.rulerOrIssuer,
    coinType: alert.coinType,
    dateFrom: alert.dateFrom,
    dateTo: alert.dateTo,
    mint: alert.mint,
    mintLocationId: alert.mintLocationId,
    material: alert.material,
    gradeOrCondition: alert.gradeOrCondition,
    priceMin: alert.priceMin,
    priceMax: alert.priceMax,
    currency: alert.currency,
    dealerPreference: alert.dealerPreference,
    sourceFilters: [...alert.sourceFilters],
    keywords: alert.keywords,
    notes: alert.notes,
  }
}

function runResultMessage(status: string, count: number, duplicates: number) {
  if (status === 'queued' || status === 'running') return t('wishlistAlerts.result.processing')
  if (status === 'failed') return t('wishlistAlerts.result.failed')
  if (status === 'rate_limited') return t('wishlistAlerts.result.rateLimited')
  if (count === 0) return t('wishlistAlerts.result.none')
  return t('wishlistAlerts.result.summary', { status: alertRunStatusLabel(status), count, duplicates })
}

async function pollAlertRun(alertId: number, runId: number) {
  if (selectedAlertId.value !== alertId) return
  try {
    const res = await getWishlistSearchAlertRun(alertId, runId)
    selectedRun.value = res.data
    if (isTerminalRunStatus(res.data.status)) {
      await finishAlertRun(res.data)
      return
    }
  } catch {
    // Keep polling; transient network errors should not lose the backend-owned run.
  }
  scheduleRunPoll(alertId, runId)
}

function scheduleRunPoll(alertId: number, runId: number) {
  clearRunPollTimer()
  runPollTimer = setTimeout(() => {
    void pollAlertRun(alertId, runId)
  }, RUN_POLL_INTERVAL_MS)
}

async function finishAlertRun(run: AlertRun) {
  clearRunPollTimer()
  running.value = false
  runMessage.value = runResultMessage(run.status, run.resultCount, run.duplicateCount)
  await Promise.all([load(), loadRuns(), loadCandidates()])
  await loadRunDetail(run.id)
}

function isTerminalRunStatus(status: AlertRun['status']) {
  return ['completed', 'failed', 'partial', 'rate_limited', 'cancelled'].includes(status)
}

function clearRunPollTimer() {
  if (runPollTimer) {
    clearTimeout(runPollTimer)
    runPollTimer = null
  }
}

onMounted(load)
onBeforeUnmount(clearRunPollTimer)
</script>
