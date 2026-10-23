<template>
  <section class="flex flex-col gap-4" data-testid="admin-updates">
    <div class="card">
      <h3 class="mb-1 mt-0 text-lg text-heading">{{ t('admin.updates.title') }}</h3>
      <p class="mb-4 mt-0 text-body text-text-secondary">{{ t('admin.updates.intro') }}</p>

      <p v-if="loading && !status" class="m-0 text-body text-text-muted">{{ t('common.loading') }}</p>
      <template v-else-if="status">
        <dl class="m-0 grid gap-3 sm:grid-cols-2">
          <div>
            <dt class="section-label mb-1">{{ t('admin.updates.current') }}</dt>
            <dd class="m-0 font-display text-xl text-text-primary" data-testid="updates-current">{{ status.current }}</dd>
          </div>
          <div>
            <dt class="section-label mb-1">{{ t('admin.updates.latest') }}</dt>
            <dd class="m-0 font-display text-xl text-text-primary" data-testid="updates-latest">{{ status.latest || '—' }}</dd>
          </div>
        </dl>
        <p class="mb-0 mt-3 text-body" :class="status.updateAvailable ? 'text-gold' : 'text-text-secondary'" role="status" data-testid="updates-message">{{ message }}</p>
        <p v-if="status.checkedAt" class="mb-0 mt-1 text-chip text-text-muted">{{ t('admin.updates.checkedAt', { time: formatTime(status.checkedAt) }) }}</p>
        <div class="mt-4 flex flex-wrap gap-2">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="loading || status.status === 'disabled'" data-testid="updates-check" @click="refresh(true)">
            <RefreshCw :size="14" aria-hidden="true" /> {{ loading ? t('admin.updates.checking') : t('admin.updates.check') }}
          </button>
        </div>
      </template>
      <p v-if="failed" class="mb-0 mt-3 text-body text-loss" role="alert">{{ t('admin.updates.loadFailed') }}</p>
    </div>

    <!-- What the new version brings -->
    <div v-if="status?.updateAvailable && status.notes" class="card" data-testid="updates-notes">
      <h3 class="mb-3 mt-0 text-lg text-heading">{{ t('admin.updates.notesTitle', { version: status.latest ?? '' }) }}</h3>
      <div
        class="markdown-rendered break-words text-body leading-[1.5] text-text-primary [&_h3]:mb-1 [&_h3]:mt-3 [&_h3]:text-base [&_h3]:font-semibold [&_li]:mb-1 [&_p]:mb-2 [&_ul]:mb-2 [&_ul]:list-disc [&_ul]:pl-5"
        v-html="notesHtml"
      ></div>
      <a v-if="status.releaseUrl" :href="status.releaseUrl" target="_blank" rel="noopener noreferrer" class="mt-3 inline-block text-body text-gold underline-offset-2 hover:underline">{{ t('admin.updates.onGitHub') }}</a>
    </div>

    <!-- How to update -->
    <div v-if="status?.updateAvailable" class="card" data-testid="updates-how">
      <h3 class="mb-2 mt-0 text-lg text-heading">{{ t('admin.updates.howTitle', { version: status.latest ?? '' }) }}</h3>
      <template v-if="status.watchtower">
        <p class="mb-3 mt-0 text-body text-text-secondary">{{ t('admin.updates.watchtowerIntro') }}</p>
        <button type="button" class="btn btn-primary" :disabled="updating" data-testid="updates-apply" @click="apply">
          <ArrowUpCircle :size="16" aria-hidden="true" /> {{ updating ? t('admin.updates.updating') : t('admin.updates.apply', { version: status.latest ?? '' }) }}
        </button>
        <p v-if="applyMessage" class="mb-0 mt-3 text-body" :class="applyFailed ? 'text-loss' : 'text-text-secondary'" role="status" data-testid="updates-apply-message">{{ applyMessage }}</p>
      </template>
      <template v-else>
        <p class="mb-3 mt-0 text-body text-text-secondary">
          {{ t('admin.updates.backupFirst') }}
          <RouterLink :to="{ path: '/settings', query: { tab: 'backups' } }" class="text-gold underline-offset-2 hover:underline">{{ t('admin.updates.backups') }}</RouterLink>
        </p>
        <dl class="m-0 flex flex-col gap-3">
          <div v-for="way in WAYS" :key="way.key">
            <dt class="text-body font-medium text-text-primary">{{ t(`admin.updates.ways.${way.key}.title`) }}</dt>
            <dd class="m-0 mt-1 text-body text-text-secondary">
              {{ t(`admin.updates.ways.${way.key}.body`) }}
              <code v-if="way.command" class="mt-1 block overflow-x-auto rounded-sm bg-input px-3 py-2 font-mono text-chip text-text-primary">{{ way.command }}</code>
            </dd>
          </div>
        </dl>
      </template>
    </div>

    <!-- Settings -->
    <form class="card" data-testid="updates-settings" @submit.prevent="save">
      <h3 class="mb-3 mt-0 text-lg text-heading">{{ t('admin.updates.settingsTitle') }}</h3>
      <label class="mb-4 flex items-center gap-2 text-body text-text-primary">
        <input v-model="checkEnabled" type="checkbox" class="h-4 w-4 accent-[var(--accent-gold)]" data-testid="updates-enabled" />
        {{ t('admin.updates.enabled') }}
      </label>
      <p class="mb-3 mt-0 text-body text-text-secondary">{{ t('admin.updates.watchtowerHelp') }}</p>
      <pre class="mb-4 mt-0 overflow-x-auto rounded-sm bg-input px-3 py-2 font-mono text-chip text-text-primary">{{ WATCHTOWER_COMPOSE }}</pre>
      <div class="grid gap-3 md:grid-cols-2">
        <div class="form-group mb-0 min-w-0">
          <label class="form-label" for="watchtower-url">{{ t('admin.updates.watchtowerUrl') }}</label>
          <input id="watchtower-url" v-model="watchtowerUrl" class="form-input" type="url" placeholder="http://watchtower:8080" maxlength="300" data-testid="updates-watchtower-url" />
        </div>
        <div class="form-group mb-0 min-w-0">
          <label class="form-label" for="watchtower-token">{{ t('admin.updates.watchtowerToken') }}</label>
          <input id="watchtower-token" v-model="watchtowerToken" class="form-input" type="password" autocomplete="off" maxlength="300" data-testid="updates-watchtower-token" />
        </div>
      </div>
      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button type="submit" class="btn btn-primary btn-sm" :disabled="saving" data-testid="updates-save">{{ saving ? t('common.saving') : t('common.save') }}</button>
        <span v-if="saved" class="text-chip text-text-muted" role="status">{{ t('admin.updates.saved') }}</span>
        <span v-if="saveFailed" class="text-chip text-loss" role="alert">{{ t('admin.updates.saveFailed') }}</span>
      </div>
    </form>
  </section>
</template>

<script setup lang="ts">
// Tells the admin which version runs and whether a newer one is out (from the
// project's GitHub releases), with its notes, and how to update: with one
// click through Watchtower when it is set up (a backup is made first), or
// with the commands of Docker Compose, Unraid and docker run.
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowUpCircle, RefreshCw } from 'lucide-vue-next'
import { applyUpdate, checkForUpdates, getAppSettings, getUpdateStatus, updateAppSettings, type UpdateStatus } from '@/api/client'
import { useDialog } from '@/composables/useDialog'
import { renderSafeMarkdown } from '@/composables/useMarkdown'
import { intlLocale, t } from '@/i18n'

const WAYS = [
  { key: 'compose', command: 'docker compose pull && docker compose up -d' },
  { key: 'unraid', command: '' },
  { key: 'run', command: 'docker pull anagermano/orlanumis:latest' },
] as const

const WATCHTOWER_COMPOSE = `  watchtower:
    image: nickfedor/watchtower:1.23.0
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      WATCHTOWER_HTTP_API_ENDPOINTS: update
      WATCHTOWER_HTTP_API_TOKEN: <a long random token>
      WATCHTOWER_CLEANUP: "true"
    restart: unless-stopped`

// While the container is replaced, the page asks every few seconds until
// the new version answers, then reloads; after a while it says so.
const POLL_MS = 5000
const POLL_LIMIT_MS = 10 * 60 * 1000

const { showConfirm } = useDialog()
const status = ref<UpdateStatus | null>(null)
const loading = ref(false)
const failed = ref(false)
const updating = ref(false)
const applyMessage = ref('')
const applyFailed = ref(false)
const checkEnabled = ref(true)
const watchtowerUrl = ref('')
const watchtowerToken = ref('')
const saving = ref(false)
const saved = ref(false)
const saveFailed = ref(false)
let pollTimer: ReturnType<typeof setTimeout> | undefined

const notesHtml = computed(() => renderSafeMarkdown(status.value?.notes))

const message = computed(() => {
  const value = status.value
  if (!value) return ''
  if (value.status === 'disabled') return t('admin.updates.disabled')
  if (value.updateAvailable) {
    return value.publishedAt
      ? t('admin.updates.availableOn', { version: value.latest ?? '', date: formatDate(value.publishedAt) })
      : t('admin.updates.available', { version: value.latest ?? '' })
  }
  if (value.status === 'unavailable') return t('admin.updates.unavailable')
  return t('admin.updates.upToDate')
})

function formatDate(value: string): string {
  return new Date(value).toLocaleDateString(intlLocale(), { day: 'numeric', month: 'long', year: 'numeric' })
}

function formatTime(value: string): string {
  return new Date(value).toLocaleString(intlLocale(), { dateStyle: 'medium', timeStyle: 'short' })
}

async function refresh(force = false) {
  loading.value = true
  failed.value = false
  try {
    status.value = (force ? await checkForUpdates() : await getUpdateStatus()).data
  } catch {
    failed.value = true
  } finally {
    loading.value = false
  }
}

async function loadSettings() {
  try {
    const settings = (await getAppSettings()).data
    checkEnabled.value = settings.UpdateCheckEnabled !== 'false'
    watchtowerUrl.value = settings.WatchtowerURL ?? ''
    watchtowerToken.value = settings.WatchtowerToken ?? ''
  } catch {
    // The defaults stay shown.
  }
}

async function save() {
  saving.value = true
  saved.value = false
  saveFailed.value = false
  try {
    await updateAppSettings([
      { key: 'UpdateCheckEnabled', value: checkEnabled.value ? 'true' : 'false' },
      { key: 'WatchtowerURL', value: watchtowerUrl.value.trim() },
      { key: 'WatchtowerToken', value: watchtowerToken.value.trim() },
    ])
    saved.value = true
    await refresh(false)
  } catch {
    saveFailed.value = true
  } finally {
    saving.value = false
  }
}

async function apply() {
  const version = status.value?.latest ?? ''
  const confirmed = await showConfirm(t('admin.updates.confirm', { version }), { title: t('admin.updates.confirmTitle') })
  if (!confirmed) return
  updating.value = true
  applyFailed.value = false
  applyMessage.value = t('admin.updates.started', { version })
  try {
    await applyUpdate()
  } catch {
    updating.value = false
    applyFailed.value = true
    applyMessage.value = t('admin.updates.applyFailed')
    return
  }
  waitForVersion(version, Date.now())
}

function waitForVersion(version: string, since: number) {
  pollTimer = setTimeout(async () => {
    try {
      const current = (await getUpdateStatus()).data.current
      if (current === version) {
        window.location.reload()
        return
      }
    } catch {
      // The app is restarting.
    }
    if (Date.now() - since > POLL_LIMIT_MS) {
      updating.value = false
      applyMessage.value = t('admin.updates.slow')
      return
    }
    waitForVersion(version, since)
  }, POLL_MS)
}

onMounted(() => {
  void refresh(false)
  void loadSettings()
})

onUnmounted(() => {
  if (pollTimer) clearTimeout(pollTimer)
})
</script>
