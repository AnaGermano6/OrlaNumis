<template>
  <div class="mb-6">
    <div class="section-content-card">
      <div class="grid grid-cols-1 gap-2 sm:grid-cols-3 sm:gap-2">
        <button
          class="btn btn-primary btn-sm min-h-11 min-w-0 justify-center whitespace-nowrap leading-5"
          :disabled="busy || !hasObverse || !aiAvailable"
          :aria-label="t('coinDetail.ai.analyzeObverse')"
          :title="!aiAvailable ? aiMessage : !hasObverse ? t('coinDetail.ai.noObverseImage') : ''"
          @click="handleAnalyze('obverse')"
        >
          {{ analyzingSide === 'obverse' ? t('coinDetail.ai.analyzing') : t('common.obverse') }}
        </button>
        <button
          class="btn btn-primary btn-sm min-h-11 min-w-0 justify-center whitespace-nowrap leading-5"
          :disabled="busy || !hasReverse || !aiAvailable"
          :aria-label="t('coinDetail.ai.analyzeReverse')"
          :title="!aiAvailable ? aiMessage : !hasReverse ? t('coinDetail.ai.noReverseImage') : ''"
          @click="handleAnalyze('reverse')"
        >
          {{ analyzingSide === 'reverse' ? t('coinDetail.ai.analyzing') : t('common.reverse') }}
        </button>
        <button
          class="btn btn-primary btn-sm min-h-11 min-w-0 justify-center whitespace-nowrap leading-5"
          :disabled="busy || !canGradeCoin || !aiAvailable"
          :aria-label="t('coinDetail.ai.gradeCoin')"
          :title="gradingDisabledTitle"
          @click="handleGradeCoin"
        >
          {{ grading ? t('coinDetail.ai.grading') : t('coinDetail.ai.grade') }}
        </button>
      </div>
      <p v-if="!aiAvailable" class="mt-2 text-body italic text-bronze">{{ aiMessage || t('coinDetail.ai.unavailable') }}</p>
      <p v-if="jobStatusMessage" class="mt-2 text-body text-gold">{{ jobStatusMessage }}</p>
      <div class="mb-4">
        <div class="grid gap-[0.35rem]">
          <p class="text-body leading-6 text-text-secondary">
            {{ t('coinDetail.ai.gradingDisclaimer') }}
          </p>
          <p v-if="!canGradeCoin" class="text-body leading-6 text-text-muted">
            {{ t('coinDetail.ai.addPhotosFirst') }}
          </p>
          <p v-if="gradingError" class="text-body leading-6 text-bronze">{{ gradingError }}</p>
        </div>
        <div v-if="gradingReport" class="mt-3 border-b border-border-subtle pb-4">
          <h5 class="mb-2 text-label font-semibold uppercase tracking-[0.08em] text-gold">{{ t('coinDetail.ai.gradingReport') }}</h5>
          <div
            class="ai-content text-body leading-[1.7] text-text-secondary [&_h1]:mb-2 [&_h1]:mt-4 [&_h1]:text-gold [&_h2]:mb-2 [&_h2]:mt-4 [&_h2]:text-gold [&_h3]:mb-2 [&_h3]:mt-4 [&_h3]:text-gold [&_ol]:pl-5 [&_strong]:text-text-primary [&_ul]:pl-5"
            v-html="renderedGradingReport"
          ></div>
        </div>
      </div>

      <div v-if="obverseAnalysis" class="mb-5 border-b border-border-subtle pb-4">
        <div class="flex items-center justify-between gap-3">
          <h5 class="mb-0 text-label font-semibold uppercase tracking-[0.08em] text-gold">{{ t('coinDetail.ai.obverseAnalysis') }}</h5>
          <button class="btn btn-ghost btn-xs" @click="handleDeleteAnalysis('obverse')">{{ t('common.remove') }}</button>
        </div>
        <div
          class="ai-content text-body leading-[1.7] text-text-secondary [&_h1]:mb-2 [&_h1]:mt-4 [&_h1]:text-gold [&_h2]:mb-2 [&_h2]:mt-4 [&_h2]:text-gold [&_h3]:mb-2 [&_h3]:mt-4 [&_h3]:text-gold [&_ol]:pl-5 [&_strong]:text-text-primary [&_ul]:pl-5"
          v-html="renderedObverse"
        ></div>
      </div>

      <div v-if="reverseAnalysis" class="mb-5 border-b border-border-subtle pb-4">
        <div class="flex items-center justify-between gap-3">
          <h5 class="mb-0 text-label font-semibold uppercase tracking-[0.08em] text-gold">{{ t('coinDetail.ai.reverseAnalysis') }}</h5>
          <button class="btn btn-ghost btn-xs" @click="handleDeleteAnalysis('reverse')">{{ t('common.remove') }}</button>
        </div>
        <div
          class="ai-content text-body leading-[1.7] text-text-secondary [&_h1]:mb-2 [&_h1]:mt-4 [&_h1]:text-gold [&_h2]:mb-2 [&_h2]:mt-4 [&_h2]:text-gold [&_h3]:mb-2 [&_h3]:mt-4 [&_h3]:text-gold [&_ol]:pl-5 [&_strong]:text-text-primary [&_ul]:pl-5"
          v-html="renderedReverse"
        ></div>
      </div>

      <div v-if="aiAnalysis && !obverseAnalysis && !reverseAnalysis" class="border-b border-border-subtle pb-4">
        <div
          class="ai-content text-body leading-[1.7] text-text-secondary [&_h1]:mb-2 [&_h1]:mt-4 [&_h1]:text-gold [&_h2]:mb-2 [&_h2]:mt-4 [&_h2]:text-gold [&_h3]:mb-2 [&_h3]:mt-4 [&_h3]:text-gold [&_ol]:pl-5 [&_strong]:text-text-primary [&_ul]:pl-5"
          v-html="renderedLegacy"
        ></div>
      </div>

      <p v-if="!obverseAnalysis && !reverseAnalysis && !aiAnalysis && aiAvailable" class="text-body italic text-text-muted">
        {{ t('coinDetail.ai.emptyHint') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { analyzeCoin, deleteAnalysis, formatAgentServiceError, getAIJob, getAIStatus, getCoinAIJobs, gradeCoin } from '@/api/client'
import { useDialog } from '@/composables/useDialog'
import { useNotifications } from '@/composables/useNotifications'
import { renderSafeMarkdown } from '@/composables/useMarkdown'
import { useToast } from '@/composables/useToast'
import type { AIJob, AIJobStartResponse, CoinGradingResult } from '@/types'
import { t } from '@/i18n'

const props = defineProps<{
  coinId: number
  obverseAnalysis?: string | null
  reverseAnalysis?: string | null
  aiAnalysis?: string | null
  hasObverse: boolean
  hasReverse: boolean
}>()

const emit = defineEmits<{
  analysisUpdated: []
}>()

const { showConfirm, showAlert } = useDialog()
const { refresh: refreshNotifications } = useNotifications()
const { showToast } = useToast()
const POLL_INTERVAL_MS = 3_000

const analyzing = ref(false)
const analyzingSide = ref<string | null>(null)
const grading = ref(false)
const gradingReport = ref('')
const gradingError = ref('')
const aiAvailable = ref(true)
const aiMessage = ref('')
const activeJob = ref<AIJob | null>(null)
let pollTimer: ReturnType<typeof setTimeout> | null = null
let unmounted = false

const renderedObverse = computed(() => renderSafeMarkdown(props.obverseAnalysis))
const renderedReverse = computed(() => renderSafeMarkdown(props.reverseAnalysis))
const renderedLegacy = computed(() => renderSafeMarkdown(props.aiAnalysis))
const renderedGradingReport = computed(() => renderSafeMarkdown(gradingReport.value))
const busy = computed(() => analyzing.value || grading.value)
const canGradeCoin = computed(() => props.hasObverse || props.hasReverse)
const gradingDisabledTitle = computed(() => {
  if (!aiAvailable.value) return aiMessage.value
  if (!canGradeCoin.value) return t('coinDetail.ai.addPhotosTitle')
  return ''
})
const jobStatusMessage = computed(() => {
  if (!activeJob.value) return ''
  if (grading.value) {
    const status = activeJob.value.status || 'queued'
    return t('coinDetail.ai.gradingStatus', { status: formatStatus(status) })
  }
  if (!analyzingSide.value) return ''
  const status = activeJob.value.status || 'queued'
  return t('coinDetail.ai.analysisStatus', { side: sideName(analyzingSide.value), status: formatStatus(status) })
})

onMounted(async () => {
  try {
    const res = await getAIStatus()
    aiAvailable.value = res.data.available
    aiMessage.value = res.data.message
  } catch {
    aiAvailable.value = false
    aiMessage.value = t('coinDetail.ai.statusCheckFailed')
  }
  await resumeAnalysisJob()
})

onUnmounted(() => {
  unmounted = true
  clearPollTimer()
})

async function handleAnalyze(side: 'obverse' | 'reverse') {
  clearPollTimer()
  analyzing.value = true
  analyzingSide.value = side
  grading.value = false
  activeJob.value = null
  try {
    const res = await analyzeCoin(props.coinId, side)
    const job = normalizeStartedJob(res.data, side)
    rememberJob(side, job.id)
    showToast(t('coinDetail.ai.analysisQueued', { side: sideName(side) }), 'info')
    await pollAnalysisJob(job.id, side, job)
  } catch (err) {
    const detail = formatAgentServiceError(err, t('errors.agentRetry'))
    await showAlert(t('coinDetail.ai.analysisFailedFor', { side: sideNameLower(side), detail }), { title: t('coinDetail.ai.analysisFailedTitle') })
    analyzing.value = false
    analyzingSide.value = null
  }
}

async function handleGradeCoin() {
  clearPollTimer()
  grading.value = true
  gradingError.value = ''
  gradingReport.value = ''
  analyzing.value = false
  analyzingSide.value = null
  activeJob.value = null
  try {
    const res = await gradeCoin(props.coinId)
    const job = normalizeStartedJob(res.data)
    rememberGradingJob(job.id)
    showToast(t('coinDetail.ai.gradingQueued'), 'info')
    await pollGradingJob(job.id, job)
  } catch (err) {
    const detail = formatAgentServiceError(err, t('coinDetail.ai.gradingStartFailed'))
    gradingError.value = detail
    await showAlert(t('coinDetail.ai.gradingFailedDetail', { detail }), { title: t('coinDetail.ai.gradingFailedTitle') })
    grading.value = false
  }
}

async function handleDeleteAnalysis(side: 'obverse' | 'reverse') {
  if (!await showConfirm(t('coinDetail.ai.confirmDelete', { side: sideNameLower(side) }), { title: t('coinDetail.ai.deleteTitle'), variant: 'danger' })) return
  try {
    await deleteAnalysis(props.coinId, side)
    emit('analysisUpdated')
  } catch {
    await showAlert(t('coinDetail.ai.deleteFailed', { side: sideNameLower(side) }), { title: t('common.error') })
  }
}

async function resumeAnalysisJob() {
  try {
    const res = await getCoinAIJobs(props.coinId, true)
    const jobs = normalizeJobList(res.data)
    const activeGrading = jobs.find((job) => isGradingJob(job) && !isTerminalStatus(job.status))
    if (activeGrading?.id) {
      grading.value = true
      gradingError.value = ''
      await pollGradingJob(activeGrading.id, activeGrading)
      return
    }
    const activeAnalysis = jobs.find((job) => isAnalysisJob(job) && !isTerminalStatus(job.status))
    if (activeAnalysis?.id) {
      const side = activeAnalysis.side === 'reverse' ? 'reverse' : 'obverse'
      analyzing.value = true
      analyzingSide.value = side
      await pollAnalysisJob(activeAnalysis.id, side, activeAnalysis)
      return
    }
  } catch {
    // Stored job IDs below still give navigation recovery a chance.
  }

  for (const side of ['obverse', 'reverse'] as const) {
    const jobId = sessionStorage.getItem(jobStorageKey(side))
    if (!jobId) continue
    try {
      const res = await getAIJob(jobId)
      if (!isAnalysisJob(res.data)) continue
      if (isTerminalStatus(res.data.status)) {
        await finishAnalysisJob(res.data, side)
      } else {
        analyzing.value = true
        analyzingSide.value = side
        await pollAnalysisJob(jobId, side, res.data)
      }
      return
    } catch {
      sessionStorage.removeItem(jobStorageKey(side))
    }
  }

  const gradingJobId = sessionStorage.getItem(gradingJobStorageKey())
  if (gradingJobId) {
    try {
      const res = await getAIJob(gradingJobId)
      if (isGradingJob(res.data)) {
        if (isTerminalStatus(res.data.status)) {
          await finishGradingJob(res.data)
        } else {
          grading.value = true
          gradingError.value = ''
          await pollGradingJob(gradingJobId, res.data)
        }
        return
      }
    } catch {
      sessionStorage.removeItem(gradingJobStorageKey())
    }
  }

  await recoverCompletedGradingJob()
}

async function pollAnalysisJob(jobId: string, side: 'obverse' | 'reverse', knownJob?: AIJob) {
  if (unmounted) return
  if (knownJob) activeJob.value = knownJob
  try {
    const res = await getAIJob(jobId)
    activeJob.value = res.data
    if (isTerminalStatus(res.data.status)) {
      await finishAnalysisJob(res.data, side)
      return
    }
  } catch {
    // Keep polling; transient network errors should not lose the backend job.
  }
  schedulePoll(jobId, side)
}

function schedulePoll(jobId: string, side: 'obverse' | 'reverse') {
  clearPollTimer()
  pollTimer = setTimeout(() => {
    void pollAnalysisJob(jobId, side)
  }, POLL_INTERVAL_MS)
}

async function pollGradingJob(jobId: string, knownJob?: AIJob) {
  if (unmounted) return
  if (knownJob) activeJob.value = knownJob
  try {
    const res = await getAIJob(jobId)
    activeJob.value = res.data
    if (isTerminalStatus(res.data.status)) {
      await finishGradingJob(res.data)
      return
    }
  } catch {
    // Keep polling; transient network errors should not lose the backend job.
  }
  scheduleGradingPoll(jobId)
}

function scheduleGradingPoll(jobId: string) {
  clearPollTimer()
  pollTimer = setTimeout(() => {
    void pollGradingJob(jobId)
  }, POLL_INTERVAL_MS)
}

async function finishAnalysisJob(job: AIJob, side: 'obverse' | 'reverse') {
  clearPollTimer()
  sessionStorage.removeItem(jobStorageKey(side))
  activeJob.value = job
  analyzing.value = false
  analyzingSide.value = null
  if (isFailedStatus(job.status)) {
    const message = job.errorMessage || t('coinDetail.ai.analysisFailedRetry')
    showToast(message, 'error')
    await showAlert(message, { title: t('coinDetail.ai.analysisFailedTitle') })
    return
  }
  activeJob.value = null
  emit('analysisUpdated')
  showToast(t('coinDetail.ai.analysisComplete', { side: sideName(side) }), 'success')
  await refreshNotifications()
}

async function finishGradingJob(job: AIJob) {
  clearPollTimer()
  sessionStorage.removeItem(gradingJobStorageKey())
  activeJob.value = job
  grading.value = false
  if (isFailedStatus(job.status)) {
    gradingError.value = job.errorMessage || t('coinDetail.ai.gradingFailedRetry')
    showToast(gradingError.value, 'error')
    await showAlert(gradingError.value, { title: t('coinDetail.ai.gradingFailedTitle') })
    return
  }

  const parsed = parseGradingResult(job.result)
  if (!parsed?.gradingReport) {
    gradingError.value = t('coinDetail.ai.noGradingReport')
    return
  }

  gradingReport.value = parsed.gradingReport
  activeJob.value = null
  showToast(t('coinDetail.ai.gradingReady'), 'success')
  await refreshNotifications()
}

async function recoverCompletedGradingJob() {
  try {
    const res = await getCoinAIJobs(props.coinId, false)
    const jobs = normalizeJobList(res.data)
    const completedGrading = jobs
      .filter((job) => isGradingJob(job) && isSuccessfulStatus(job.status) && parseGradingResult(job.result)?.gradingReport)
      .sort((a, b) => jobTimestamp(b) - jobTimestamp(a))[0]
    if (!completedGrading) return

    const parsed = parseGradingResult(completedGrading.result)
    if (!parsed?.gradingReport) return

    gradingReport.value = parsed.gradingReport
    gradingError.value = ''
    grading.value = false
    activeJob.value = null
  } catch {
    // Completed-job recovery is best effort; active jobs and explicit actions still work.
  }
}

function clearPollTimer() {
  if (pollTimer) {
    clearTimeout(pollTimer)
    pollTimer = null
  }
}

function normalizeStartedJob(job: AIJobStartResponse, side?: 'obverse' | 'reverse'): AIJob {
  const data = job.job ?? job
  const id = String(('jobId' in data ? data.jobId : data.id) ?? '')
  if (!id) throw new Error('Missing AI job ID')
  return {
    id,
    coinId: data.coinId,
    jobType: data.jobType,
    side: data.side ?? side ?? null,
    status: data.status,
    result: data.result,
    errorMessage: data.errorMessage,
    createdAt: data.createdAt ?? '',
    updatedAt: data.updatedAt ?? '',
    startedAt: data.startedAt,
    completedAt: data.completedAt,
  }
}

function normalizeJobList(data: AIJob[] | { jobs?: AIJob[] }): AIJob[] {
  return Array.isArray(data) ? data : data.jobs ?? []
}

function isAnalysisJob(job: AIJob) {
  return job.coinId === props.coinId && /analy/i.test(job.jobType)
}

function isGradingJob(job: AIJob) {
  return job.coinId === props.coinId && /grading/i.test(job.jobType)
}

function parseGradingResult(result: unknown): CoinGradingResult | null {
  const raw = unwrapGradingResult(result)
  if (!raw || typeof raw !== 'object') return null
  const data = raw as Record<string, unknown>
  const gradingReport = typeof data.gradingReport === 'string'
    ? data.gradingReport
    : typeof data.grading_report === 'string'
      ? data.grading_report
      : ''
  return gradingReport ? { gradingReport } : null
}

function unwrapGradingResult(result: unknown): unknown {
  if (typeof result === 'string') {
    try {
      return unwrapGradingResult(JSON.parse(result))
    } catch {
      return { gradingReport: result }
    }
  }
  if (result && typeof result === 'object') {
    const data = result as Record<string, unknown>
    return data.gradingResult ?? data.result ?? result
  }
  return result
}

function isTerminalStatus(status: string) {
  return ['completed', 'succeeded', 'success', 'failed', 'error', 'cancelled', 'canceled'].includes(status.toLowerCase())
}

function isFailedStatus(status: string) {
  return ['failed', 'error', 'cancelled', 'canceled'].includes(status.toLowerCase())
}

function isSuccessfulStatus(status: string) {
  return ['completed', 'succeeded', 'success'].includes(status.toLowerCase())
}

function jobTimestamp(job: AIJob) {
  return Date.parse(job.completedAt ?? job.updatedAt ?? job.createdAt ?? '') || 0
}

function rememberJob(side: 'obverse' | 'reverse', jobId: string) {
  sessionStorage.setItem(jobStorageKey(side), jobId)
}

function rememberGradingJob(jobId: string) {
  sessionStorage.setItem(gradingJobStorageKey(), jobId)
}

function jobStorageKey(side: 'obverse' | 'reverse') {
  return `aiJob:analysis:${props.coinId}:${side}`
}

function gradingJobStorageKey() {
  return `aiJob:grading:${props.coinId}`
}

function formatStatus(status: string) {
  const normalized = status.toLowerCase()
  if (normalized === 'queued' || normalized === 'pending') return t('jobs.status.queued')
  if (normalized === 'running' || normalized === 'processing') return t('jobs.status.inProgress')
  return normalized
}

// Side name as it opens a sentence ("Obverse analysis…") vs. mid-sentence ("…for obverse").
function sideName(side: string) {
  return side === 'reverse' ? t('coinDetail.ai.side.reverse') : t('coinDetail.ai.side.obverse')
}

function sideNameLower(side: string) {
  return side === 'reverse' ? t('coinDetail.ai.sideLower.reverse') : t('coinDetail.ai.sideLower.obverse')
}
</script>

<style scoped>
/*
 * :deep() audit — markdown-rendered content
 * Target: HTML elements emitted by markdown-it at runtime inside .ai-content.
 * These selectors are intentional: the elements are generated by the markdown
 * renderer after Vue template compilation. Tailwind utilities cannot fully
 * replace this reach because the content is not authored with stable child
 * class names — they are fixed HTML tags (h1, strong, ul, etc.) without classes.
 */
.ai-content :deep(h1),
.ai-content :deep(h2),
.ai-content :deep(h3) {
  color: var(--accent-gold);
  margin-top: 1rem;
  margin-bottom: 0.5rem;
}

.ai-content :deep(strong) {
  color: var(--text-primary);
}

.ai-content :deep(ul),
.ai-content :deep(ol) {
  padding-left: 1.25rem;
}
</style>
