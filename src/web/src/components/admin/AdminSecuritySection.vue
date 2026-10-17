<template>
  <div class="flex min-w-0 flex-col gap-4">
    <section class="admin-section card min-w-0">
      <div class="mb-4 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p class="section-label">{{ t('admin.security.eyebrow') }}</p>
          <h2 class="m-0 text-xl font-medium">{{ t('admin.security.title') }}</h2>
        </div>
        <button class="btn btn-secondary btn-sm" :disabled="loading" @click="loadAll">
          {{ loading ? t('common.loading') : t('admin.security.refresh') }}
        </button>
      </div>

      <div
        v-if="error"
        class="mb-4 flex items-start gap-2 rounded-sm border border-status-error-border bg-status-error-bg p-3 text-body text-loss"
        role="alert"
      >
        <AlertTriangle :size="18" />
        <span>{{ error }}</span>
      </div>

      <div class="grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(150px,1fr))]">
        <div v-for="card in summaryCards" :key="card.label" class="flex flex-col gap-[0.35rem] rounded-sm border border-border-subtle bg-input p-3">
          <span class="section-label">{{ card.label }}</span>
          <span class="font-display text-xl font-semibold text-gold">{{ card.value }}</span>
        </div>
      </div>
    </section>

    <section class="admin-section card min-w-0">
      <div class="mb-4 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p class="section-label">{{ t('admin.security.exposureEyebrow') }}</p>
          <h2 class="m-0 text-xl font-medium">{{ t('admin.security.exposureTitle') }}</h2>
        </div>
        <button class="btn btn-secondary btn-sm" :disabled="exposureLoading" @click="loadExposure">
          {{ exposureLoading ? t('common.checking') : t('admin.security.runCheck') }}
        </button>
      </div>

      <div v-if="exposure" class="flex flex-col gap-3">
        <p v-if="exposure.publicIp" class="m-0 text-base text-text-secondary">
          <I18nT keypath="admin.security.ipSeen"><template #ip><strong class="text-gold">{{ exposure.publicIp }}</strong></template></I18nT>
        </p>
        <div
          v-for="check in exposureChecks"
          :key="check.label"
          class="grid items-start gap-3 rounded-sm border border-border-subtle bg-input p-3 [grid-template-columns:max-content_minmax(0,1fr)]"
          :class="!check.ok ? 'border-border-accent' : ''"
        >
          <span class="chip-sm inline-flex min-w-16 justify-center">{{ check.ok ? t('common.ok') : t('admin.security.review') }}</span>
          <div class="min-w-0">
            <strong class="block leading-[1.3]">{{ check.label }}</strong>
            <p class="mt-1 break-words text-base text-text-secondary">{{ check.message }}</p>
          </div>
        </div>
        <div
          v-if="exposure.warnings?.length"
          class="flex items-start gap-2 rounded-sm border border-border-accent bg-[var(--accent-gold-glow)] p-3 text-body text-warning"
        >
          <AlertTriangle :size="18" />
          <ul class="m-0 list-disc pl-4">
            <li v-for="warning in exposure.warnings" :key="warning">{{ warning }}</li>
          </ul>
        </div>
      </div>
      <p v-else class="text-body text-text-muted">{{ t('admin.security.runHint') }}</p>
    </section>

    <section class="admin-section card min-w-0">
      <div class="mb-4 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p class="section-label">{{ t('admin.security.eventsEyebrow') }}</p>
          <h2 class="m-0 text-xl font-medium">{{ t('admin.security.eventsTitle') }}</h2>
        </div>
        <button class="btn btn-secondary btn-sm" :disabled="eventsLoading" @click="loadEvents">
          {{ eventsLoading ? t('common.loading') : t('admin.security.applyFilters') }}
        </button>
      </div>

      <form class="mb-4 grid grid-cols-1 gap-2 md:grid-cols-2 xl:grid-cols-4" @submit.prevent="loadEvents">
        <input v-model="filters.type" class="form-input min-w-0 w-full" :placeholder="t('admin.security.type')" />
        <select v-model="filters.outcome" class="form-select min-w-0 w-full">
          <option value="">{{ t('admin.security.allOutcomes') }}</option>
          <option value="success">{{ t('admin.security.outcomes.success') }}</option>
          <option value="failure">{{ t('admin.security.outcomes.failure') }}</option>
          <option value="blocked">{{ t('admin.security.outcomes.blocked') }}</option>
        </select>
        <select v-model="filters.severity" class="form-select min-w-0 w-full">
          <option value="">{{ t('admin.security.allSeverities') }}</option>
          <option value="info">{{ t('admin.security.severities.info') }}</option>
          <option value="warning">{{ t('admin.security.severities.warning') }}</option>
          <option value="critical">{{ t('admin.security.severities.critical') }}</option>
        </select>
        <input v-model="filters.username" class="form-input min-w-0 w-full" :placeholder="t('admin.security.user')" autocomplete="off" />
        <input v-model="filters.ip" class="form-input min-w-0 w-full" :placeholder="t('admin.security.ip')" autocomplete="off" />
        <input v-model="filters.since" class="form-input min-w-0 w-full appearance-none overflow-hidden" type="date" :aria-label="t('admin.security.sinceDate')" />
        <select v-model.number="filters.limit" class="form-select min-w-0 w-full">
          <option :value="25">25</option>
          <option :value="50">50</option>
          <option :value="100">100</option>
          <option :value="250">250</option>
        </select>
        <button class="btn btn-primary btn-sm w-full justify-center" type="submit" :disabled="eventsLoading">{{ t('admin.security.filter') }}</button>
      </form>

      <div class="w-full max-w-full min-w-0 overflow-x-auto">
        <table class="data-table min-w-[44rem] table-fixed [&_th]:px-2 [&_th]:py-3 [&_td]:px-2 [&_td]:py-3 [&_td]:align-top">
          <thead>
            <tr>
              <th>{{ t('admin.security.time') }}</th>
              <th>{{ t('admin.security.type') }}</th>
              <th>{{ t('admin.security.severity') }}</th>
              <th>{{ t('admin.security.outcome') }}</th>
              <th>{{ t('admin.security.user') }}</th>
              <th>{{ t('admin.security.ip') }}</th>
              <th>{{ t('admin.security.message') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="events.length === 0">
              <td colspan="7" class="p-6 text-center text-body text-text-muted">{{ t('admin.security.noEvents') }}</td>
            </tr>
            <tr v-for="event in events" :key="event.id">
              <td class="w-32 max-w-32 break-words text-body text-text-muted">{{ formatDateTime(event.timestamp) }}</td>
              <td>{{ event.type }}</td>
              <td><span class="chip-sm">{{ event.severity }}</span></td>
              <td>{{ event.outcome ?? '—' }}</td>
              <td>{{ event.username ?? '—' }}</td>
              <td>{{ event.ip ?? '—' }}</td>
              <td>{{ event.message ?? '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="admin-section card min-w-0">
      <div class="mb-4 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p class="section-label">{{ t('admin.security.lockoutsEyebrow') }}</p>
          <h2 class="m-0 text-xl font-medium">{{ t('admin.security.lockedTitle') }}</h2>
        </div>
      </div>
      <div v-if="lockedUsers.length === 0" class="text-body text-text-muted">
        {{ t('admin.security.noLocked') }}
      </div>
      <div v-else class="flex flex-col gap-3">
        <div
          v-for="user in lockedUsers"
          :key="user.id"
          class="flex flex-col gap-3 rounded-sm border border-border-subtle bg-input p-3 sm:flex-row sm:items-start sm:justify-between"
        >
          <div>
            <strong>{{ user.username }}</strong>
            <p class="mt-1 break-words text-base text-text-secondary">{{ t('admin.security.lockedUntil', { date: formatDateTime(user.lockedUntil ?? '') }) }}</p>
          </div>
          <button class="btn btn-secondary btn-sm" :disabled="unlockingUserId === user.id" @click="unlock(user.id)">
            {{ unlockingUserId === user.id ? t('admin.security.unlocking') : t('admin.security.unlock') }}
          </button>
        </div>
      </div>
    </section>
  </div>
</template>



<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { AlertTriangle } from 'lucide-vue-next'
import {
  getSecurityEvents,
  getSecurityExposureCheck,
  getSecuritySummary,
  unlockUser,
} from '@/api/client'
import type { SecurityEvent, SecurityEventFilters, SecurityExposureCheck, SecuritySummary, UserInfo } from '@/types'
import I18nT from '@/components/I18nT.vue'
import { t, intlLocale } from '@/i18n'

const props = defineProps<{
  users: UserInfo[]
  registrationMode?: string
}>()

const emit = defineEmits<{
  unlocked: [userId: number]
}>()

const loading = ref(false)
const eventsLoading = ref(false)
const exposureLoading = ref(false)
const error = ref('')
const summary = ref<SecuritySummary>({
  failedLogins: 0,
  lockedAccounts: 0,
  recentEvents: 0,
})
const events = ref<SecurityEvent[]>([])
const exposure = ref<SecurityExposureCheck | null>(null)
const unlockingUserId = ref<number | null>(null)

const filters = reactive<SecurityEventFilters>({
  type: '',
  severity: '',
  username: '',
  ip: '',
  outcome: '',
  since: '',
  limit: 50,
})

const summaryCards = computed(() => [
  { label: t('admin.security.cards.failedLogins'), value: summary.value.failedLogins },
  { label: t('admin.security.cards.lockedAccounts'), value: summary.value.lockedAccounts },
  { label: t('admin.security.cards.recentEvents'), value: summary.value.recentEvents },
])

const lockedUsers = computed(() =>
  props.users.filter((user) => {
    if (!user.lockedUntil) return false
    return new Date(user.lockedUntil).getTime() > Date.now()
  }),
)

const exposureChecks = computed(() => {
  const current = exposure.value
  if (!current) return []
  const config = current.config
  const registrationMode = config?.registrationMode ?? props.registrationMode
  return [
    buildExposureRow(t('admin.security.checks.proxy'), current.proxy ?? config?.trustedProxiesConfigured ?? !hasWarning(current, 'Trusted proxies'), current.proxyWarning, t('admin.security.checks.proxyOk')),
    buildExposureRow('CORS', current.cors ?? !hasWarning(current, 'CORS'), current.corsWarning, t('admin.security.checks.corsOk')),
    buildExposureRow(t('admin.security.checks.publicUrl'), current.publicAppUrl ?? current.publicAppURL ?? Boolean(config?.publicAppURL && !hasWarning(current, 'PublicAppURL')), current.publicAppUrlWarning, t('admin.security.checks.publicUrlOk')),
    buildExposureRow(t('admin.security.checks.registration'), current.registration ?? (registrationMode ? registrationMode === 'closed' : !hasWarning(current, 'Registration')), current.registrationWarning, t('admin.security.checks.registrationMode', { mode: registrationMode || t('admin.security.checks.backendDefault') })),
    buildExposureRow(t('admin.security.checks.agentToken'), current.agentToken ?? config?.agentInternalTokenSet, current.agentTokenWarning, t('admin.security.checks.agentTokenOk')),
  ]
})

function buildExposureRow(label: string, ok: boolean | undefined, warning: string | undefined, fallback: string) {
  return {
    label,
    ok: ok !== false,
    message: warning || fallback,
  }
}

function hasWarning(current: SecurityExposureCheck, token: string) {
  return (current.warnings ?? []).some((warning) => warning.toLowerCase().includes(token.toLowerCase()))
}

onMounted(() => {
  void loadAll()
})

async function loadAll() {
  loading.value = true
  error.value = ''
  try {
    await Promise.all([loadSummary(), loadEvents(), loadExposure()])
  } catch {
    error.value = t('admin.security.loadFailed')
  } finally {
    loading.value = false
  }
}

async function loadSummary() {
  const res = await getSecuritySummary()
  summary.value = normalizeSummary(res.data)
}

async function loadEvents() {
  eventsLoading.value = true
  try {
    const params = compactFilters(filters)
    const res = await getSecurityEvents(params)
    const rawEvents = Array.isArray(res.data) ? res.data : (res.data.events ?? [])
    events.value = rawEvents.map(normalizeEvent)
  } finally {
    eventsLoading.value = false
  }
}

async function loadExposure() {
  exposureLoading.value = true
  try {
    const res = await getSecurityExposureCheck()
    exposure.value = res.data
  } finally {
    exposureLoading.value = false
  }
}

async function unlock(userId: number) {
  unlockingUserId.value = userId
  error.value = ''
  try {
    await unlockUser(userId)
    emit('unlocked', userId)
    await loadSummary()
  } catch {
    error.value = t('admin.security.unlockFailed')
  } finally {
    unlockingUserId.value = null
  }
}

function compactFilters(source: SecurityEventFilters): SecurityEventFilters {
  const params: SecurityEventFilters = {}
  if (source.type) params.type = source.type
  if (source.severity) params.severity = source.severity
  if (source.username) params.username = source.username
  if (source.ip) params.clientIp = source.ip
  if (source.outcome) params.outcome = source.outcome
  if (source.since) params.since = new Date(`${source.since}T00:00:00`).toISOString()
  if (source.limit) params.limit = source.limit
  return params
}

function normalizeSummary(data: SecuritySummary | Record<string, unknown>): SecuritySummary {
  const root = data as Record<string, unknown>
  const raw = (typeof root.summary === 'object' && root.summary !== null ? root.summary : root) as Record<string, unknown>
  return {
    failedLogins: asNumber(raw.failedLogins ?? raw.loginFailures ?? raw.failed_login_count ?? raw.failedLoginCount),
    lockedAccounts: asNumber(raw.lockedAccounts ?? raw.locked_account_count ?? raw.lockedAccountCount),
    recentEvents: asNumber(raw.recentEvents ?? raw.recent_security_events ?? raw.recentEventCount),
  }
}

function normalizeEvent(event: SecurityEvent): SecurityEvent {
  return {
    ...event,
    timestamp: event.timestamp ?? event.createdAt ?? '',
    ip: event.ip ?? event.clientIp ?? null,
    severity: event.severity ?? '—',
  }
}

function asNumber(value: unknown) {
  return typeof value === 'number' ? value : Number(value ?? 0)
}

function formatDateTime(value: string) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString(intlLocale())
}
</script>
