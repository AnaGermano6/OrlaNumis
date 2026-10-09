<template>
  <div class="container">
    <DesktopPageContext :title="t('availability.title')">
      <template #actions>
        <AppIconButton to="/wishlist" :title="t('availability.back')">
          <ArrowLeft :size="20" />
        </AppIconButton>
      </template>
      <header class="page-header">
        <h1>{{ t('availability.title') }}</h1>
        <RouterLink class="btn btn-secondary btn-sm" to="/wishlist">
          <ArrowLeft :size="16" /> {{ t('availability.back') }}
        </RouterLink>
      </header>
    </DesktopPageContext>

    <div v-if="unauthorized" class="empty-state card">
      <h3>{{ t('availability.signInTitle') }}</h3>
      <p>{{ t('availability.signInBody') }}</p>
    </div>

    <!-- Deep-linked single-run detail (e.g. from a notification) -->
    <template v-else-if="routeRunId !== null">
      <div v-if="detailLoading" class="loading-overlay"><div class="spinner"></div></div>
      <div v-else-if="detailError" class="empty-state card">
        <h3>{{ t('availability.runNotFound') }}</h3>
        <p>{{ detailError }}</p>
      </div>
      <section v-else-if="detailRun" class="card">
        <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div>
            <span class="section-label mb-1 block">{{ t('availability.runNumber', { id: detailRun.id }) }}</span>
            <strong class="text-lg text-text-primary">{{ formatDate(detailRun.startedAt) }}</strong>
          </div>
          <span class="chip-sm" :class="statusClass(detailRun.status)">{{ runStatusLabel(detailRun.status) }}</span>
        </div>

        <div class="mb-4 flex flex-wrap items-center gap-4 text-body text-text-secondary">
          <span>{{ t('wishlist.check.checked', { count: detailRun.coinsChecked }) }}</span>
          <span class="font-semibold text-gain">{{ t('wishlist.check.available', { count: detailRun.available }) }}</span>
          <span class="font-semibold text-loss">{{ t('wishlist.check.unavailable', { count: detailRun.unavailable }) }}</span>
          <span class="font-semibold text-warning">{{ t('wishlist.check.unknown', { count: detailRun.unknown }) }}</span>
          <span v-if="detailRun.failMessage" class="text-loss">{{ detailRun.failMessage }}</span>
        </div>

        <p v-if="!detailRun.results || detailRun.results.length === 0" class="py-4 text-center text-text-muted">{{ t('availability.noResults') }}</p>
        <div v-else class="overflow-x-auto">
          <table class="data-table text-body [&_th]:px-2 [&_th]:py-2 [&_td]:px-2 [&_td]:py-2">
            <thead class="hidden md:table-header-group">
              <tr>
                <th>{{ t('availability.coin') }}</th>
                <th>{{ t('common.status') }}</th>
                <th>{{ t('availability.reason') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="result in detailRun.results" :key="result.id" class="grid grid-cols-[minmax(50vw,1fr)_auto] md:table-row">
                <td class="min-w-0 !border-b-0 md:!border-b">
                  <SafeExternalLink v-if="safeResultUrl(result.url)" :href="result.url" target="_blank" rel="noopener" class="block min-w-0 text-gold no-underline hover:underline md:inline" :title="result.url">
                    <span class="block truncate md:inline md:overflow-visible md:whitespace-normal" :title="result.coinName">{{ result.coinName }}</span>
                  </SafeExternalLink>
                  <span v-else class="block truncate md:inline md:overflow-visible md:whitespace-normal" :title="result.coinName">{{ result.coinName }}</span>
                  <span v-if="resultHost(result.url)" class="mt-0.5 block text-sm text-text-muted">{{ resultHost(result.url) }}</span>
                </td>
                <td class="!border-b-0 md:!border-b"><span class="chip-sm" :class="resultStatusClass(result.status)">{{ resultStatusLabel(result.status) }}</span></td>
                <td class="col-span-2 whitespace-normal italic text-text-secondary md:table-cell md:max-w-[220px] md:overflow-hidden md:text-ellipsis md:not-italic md:whitespace-nowrap">{{ result.reason || '--' }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <RouterLink class="btn btn-secondary btn-sm mt-4" to="/wishlist/availability-runs">{{ t('availability.backToRuns') }}</RouterLink>
      </section>
    </template>

    <template v-else>
      <div v-if="loading" class="loading-overlay"><div class="spinner"></div></div>
      <div v-else-if="loadError" class="empty-state card">
        <h3>{{ t('availability.loadFailedTitle') }}</h3>
        <p>{{ loadError }}</p>
      </div>
      <div v-else-if="runs.length === 0" class="empty-state card">
        <h3>{{ t('availability.empty.title') }}</h3>
        <p>{{ t('availability.empty.body') }}</p>
      </div>

      <template v-else>
        <div class="flex flex-col gap-2">
          <div
            v-for="run in runs"
            :key="run.id"
            class="card !p-0 cursor-pointer !px-4 !py-3 transition-colors hover:bg-card-hover"
            @click="toggleExpand(run.id)"
          >
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div class="min-w-0">
                <span class="section-label mb-0.5 block">{{ triggerLabel(run.triggerType) }}</span>
                <strong class="text-body text-text-primary">{{ formatDate(run.startedAt) }}</strong>
              </div>
              <div class="flex flex-wrap items-center gap-3 text-body text-text-secondary">
                <span>{{ t('wishlist.check.checked', { count: run.coinsChecked }) }}</span>
                <span class="font-semibold text-loss">{{ t('availability.unavailShort', { count: run.unavailable }) }}</span>
                <span class="chip-sm" :class="statusClass(run.status)">{{ runStatusLabel(run.status) }}</span>
              </div>
            </div>

            <div v-if="expandedRunId === run.id" class="mt-3 border-t border-border-subtle pt-3" @click.stop>
              <div v-if="expandedLoading" class="flex justify-center py-4"><div class="spinner"></div></div>
              <template v-else>
                <div class="mb-3 flex flex-wrap items-center gap-4 text-body text-text-secondary">
                  <span class="font-semibold text-gain">{{ t('wishlist.check.available', { count: run.available }) }}</span>
                  <span class="font-semibold text-warning">{{ t('wishlist.check.unknown', { count: run.unknown }) }}</span>
                  <span v-if="run.failMessage" class="text-loss">{{ run.failMessage }}</span>
                </div>
                <p v-if="expandedResults.length === 0" class="py-4 text-center text-text-muted">{{ t('availability.noResults') }}</p>
                <div v-else class="overflow-x-auto">
                  <table class="data-table text-body [&_th]:px-2 [&_th]:py-2 [&_td]:px-2 [&_td]:py-2">
                    <thead class="hidden md:table-header-group">
                      <tr>
                        <th>{{ t('availability.coin') }}</th>
                        <th>{{ t('common.status') }}</th>
                        <th>{{ t('availability.reason') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="result in expandedResults" :key="result.id" class="grid grid-cols-[minmax(50vw,1fr)_auto] md:table-row">
                        <td class="min-w-0 !border-b-0 md:!border-b">
                          <SafeExternalLink v-if="safeResultUrl(result.url)" :href="result.url" target="_blank" rel="noopener" class="block min-w-0 text-gold no-underline hover:underline md:inline" :title="result.url" @click.stop>
                            <span class="block truncate md:inline md:overflow-visible md:whitespace-normal" :title="result.coinName">{{ result.coinName }}</span>
                          </SafeExternalLink>
                          <span v-else class="block truncate md:inline md:overflow-visible md:whitespace-normal" :title="result.coinName">{{ result.coinName }}</span>
                          <span v-if="resultHost(result.url)" class="mt-0.5 block text-sm text-text-muted">{{ resultHost(result.url) }}</span>
                        </td>
                        <td class="!border-b-0 md:!border-b"><span class="chip-sm" :class="resultStatusClass(result.status)">{{ resultStatusLabel(result.status) }}</span></td>
                        <td class="col-span-2 whitespace-normal italic text-text-secondary md:table-cell md:max-w-[220px] md:overflow-hidden md:text-ellipsis md:not-italic md:whitespace-nowrap">{{ result.reason || '--' }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </template>
            </div>
          </div>
        </div>

        <div class="mt-6 flex items-center justify-center gap-4">
          <button class="btn btn-secondary btn-sm" :disabled="page <= 1" @click="prevPage()">&larr; {{ t('common.previous') }}</button>
          <span class="text-body text-text-secondary">{{ t('availability.page', { page }) }}</span>
          <button class="btn btn-secondary btn-sm" :disabled="page * limit >= total" @click="nextPage()">{{ t('common.next') }} &rarr;</button>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowLeft } from 'lucide-vue-next'
import { getMyAvailabilityRunDetail, listMyAvailabilityRuns } from '@/api/client'
import SafeExternalLink from '@/components/SafeExternalLink.vue'
import { sanitizeExternalUrl } from '@/composables/useSafeExternalLink'
import type { AvailabilityResult, AvailabilityRun } from '@/types'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { intlLocale, t } from '@/i18n'

const route = useRoute()

const runs = ref<AvailabilityRun[]>([])
const total = ref(0)
const page = ref(1)
const limit = 20
const loading = ref(false)
const loadError = ref('')
const unauthorized = ref(false)

const expandedRunId = ref<number | null>(null)
const expandedResults = ref<AvailabilityResult[]>([])
const expandedLoading = ref(false)

const routeRunId = computed(() => {
  const raw = route.params.id
  const id = Array.isArray(raw) ? raw[0] : raw
  const parsed = id ? Number(id) : NaN
  return Number.isFinite(parsed) ? parsed : null
})
const detailRun = ref<AvailabilityRun | null>(null)
const detailLoading = ref(false)
const detailError = ref('')

function isUnauthorizedError(err: unknown): boolean {
  return (err as { response?: { status?: number } })?.response?.status === 401
}

function statusClass(status: string): string {
  if (status === 'queued') return 'text-gold'
  if (status === 'running') return 'text-bronze'
  if (status === 'failed' || status === 'partial_failure') return 'text-loss'
  return 'text-gain'
}

function runStatusLabel(status: string): string {
  switch (status) {
    case 'completed': return t('availability.runStatus.done')
    case 'queued': return t('availability.runStatus.queued')
    case 'running': return t('availability.runStatus.running')
    case 'failed': return t('availability.runStatus.failed')
    case 'partial_failure': return t('availability.runStatus.partialFailure')
    default: return status
  }
}

function resultStatusLabel(status: string): string {
  switch (status) {
    case 'available': return t('availability.resultStatus.available')
    case 'unavailable': return t('availability.resultStatus.unavailable')
    case 'unknown': return t('availability.resultStatus.unknown')
    default: return status
  }
}

function triggerLabel(trigger: string): string {
  if (trigger === 'scheduled') return t('availability.trigger.scheduled')
  if (trigger === 'manual') return t('availability.trigger.manual')
  return trigger
}

function resultStatusClass(status: string): string {
  if (status === 'available') return 'text-gain'
  if (status === 'unavailable') return 'text-loss'
  return 'text-warning'
}

function safeResultUrl(url: string | null | undefined): string | null {
  return sanitizeExternalUrl(url)
}

function resultHost(url: string | null | undefined): string | null {
  const safe = sanitizeExternalUrl(url)
  if (!safe) return null
  try {
    return new URL(safe).hostname.replace(/^www\./, '')
  } catch {
    return null
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString(intlLocale())
}

async function loadRuns() {
  loading.value = true
  loadError.value = ''
  try {
    const res = await listMyAvailabilityRuns(page.value, limit)
    runs.value = res.data?.runs ?? []
    total.value = res.data?.total ?? 0
  } catch (err) {
    if (isUnauthorizedError(err)) {
      unauthorized.value = true
    } else {
      loadError.value = t('availability.loadFailed')
    }
    runs.value = []
    total.value = 0
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

async function toggleExpand(runId: number) {
  if (expandedRunId.value === runId) {
    expandedRunId.value = null
    expandedResults.value = []
    return
  }
  expandedRunId.value = runId
  expandedResults.value = []
  expandedLoading.value = true
  try {
    const res = await getMyAvailabilityRunDetail(runId)
    expandedResults.value = res.data?.results ?? []
  } catch (err) {
    if (isUnauthorizedError(err)) unauthorized.value = true
    expandedResults.value = []
  } finally {
    expandedLoading.value = false
  }
}

async function loadDetail(runId: number) {
  detailLoading.value = true
  detailError.value = ''
  try {
    const res = await getMyAvailabilityRunDetail(runId)
    detailRun.value = res.data ?? null
  } catch (err) {
    if (isUnauthorizedError(err)) {
      unauthorized.value = true
    } else {
      detailError.value = t('availability.detailFailed')
    }
    detailRun.value = null
  } finally {
    detailLoading.value = false
  }
}

onMounted(async () => {
  if (routeRunId.value !== null) {
    await loadDetail(routeRunId.value)
    return
  }
  await loadRuns()
})
</script>
