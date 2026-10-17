<template>
  <div class="relative min-h-screen flex items-center justify-center p-8 bg-[radial-gradient(ellipse_at_top,var(--bg-secondary)_0%,var(--bg-primary)_70%)]">
    <LanguageSwitcher class="absolute top-[calc(1rem+env(safe-area-inset-top))] right-4" />
    <div class="w-full max-w-[400px] text-center">
      <img
        :src="coinLogoSrc"
        :alt="t('app.brandFull')"
        class="w-20 h-20 rounded-full object-cover border-[3px] border-gold-dim mb-6 shadow-[0_0_30px_var(--accent-gold-glow)] mx-auto block"
      />
      <h1 class="mb-1">{{ t('app.brandFull') }}</h1>
      <p class="text-text-secondary mb-8 text-base">{{ t('auth.login.subtitle') }}</p>
      <form @submit.prevent="handleLogin" class="text-left">
        <div class="form-group">
          <label class="form-label">{{ t('auth.username') }}</label>
          <input v-model="username" class="form-input" required autocomplete="username" />
        </div>
        <div class="form-group">
          <label class="form-label">{{ t('auth.password') }}</label>
          <input v-model="password" type="password" class="form-input" required autocomplete="current-password" />
        </div>
        <p v-if="error" class="text-loss text-body mb-2">{{ error }}</p>
        <button
          type="submit"
          class="btn btn-primary w-full justify-center py-3 mt-2"
          :disabled="loading"
        >
          {{ loading ? t('auth.login.signingIn') : t('auth.login.signIn') }}
        </button>
      </form>
      <p class="mt-6 text-body text-text-secondary">
        {{ t('auth.login.noAccount') }} <router-link to="/register">{{ t('auth.login.createOne') }}</router-link>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import { t } from '@/i18n'

const router = useRouter()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const coinLogoSrc = '/coin-logo.jpg'
let retryTimer: ReturnType<typeof setInterval> | null = null

onUnmounted(() => {
  clearRetryTimer()
})

async function handleLogin() {
  error.value = ''
  loading.value = true
  const trimmedUsername = username.value.trim()
  try {
    await auth.doLogin(trimmedUsername, password.value)
    router.push('/')
  } catch (err: unknown) {
    if (!handleRateLimitError(err)) {
      error.value = t('auth.login.invalidCredentials')
    }
  } finally {
    loading.value = false
  }
}

function handleRateLimitError(err: unknown) {
  const response = getErrorResponse(err)
  if (response?.status !== 429) return false

  const retryAfter = getRetryAfterSeconds(response.headers)
  if (retryAfter > 0) {
    startRetryCountdown(retryAfter)
  } else {
    error.value = t('auth.login.tooManyAttempts')
  }
  return true
}

function getErrorResponse(err: unknown): { status?: number; headers?: Record<string, unknown> } | null {
  if (typeof err !== 'object' || err === null || !('response' in err)) return null
  const response = (err as { response?: unknown }).response
  if (typeof response !== 'object' || response === null) return null
  return response as { status?: number; headers?: Record<string, unknown> }
}

function getRetryAfterSeconds(headers: Record<string, unknown> | undefined) {
  const raw = headers?.['retry-after'] ?? headers?.['Retry-After']
  if (typeof raw !== 'string') return 0
  const seconds = Number(raw)
  if (Number.isFinite(seconds)) return Math.max(0, Math.ceil(seconds))
  const retryAt = new Date(raw).getTime()
  if (Number.isNaN(retryAt)) return 0
  return Math.max(0, Math.ceil((retryAt - Date.now()) / 1000))
}

function startRetryCountdown(seconds: number) {
  clearRetryTimer()
  let remaining = seconds
  error.value = formatRateLimitMessage(remaining)
  retryTimer = setInterval(() => {
    remaining -= 1
    if (remaining <= 0) {
      clearRetryTimer()
      error.value = t('auth.login.tooManyAttempts')
      return
    }
    error.value = formatRateLimitMessage(remaining)
  }, 1000)
}

function formatRateLimitMessage(seconds: number) {
  return t('auth.login.retryIn', { count: seconds })
}

function clearRetryTimer() {
  if (!retryTimer) return
  clearInterval(retryTimer)
  retryTimer = null
}
</script>
