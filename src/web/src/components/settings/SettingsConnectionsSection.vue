<template>
  <section class="card text-text-primary">
    <h2 class="mb-5 border-b border-border-subtle pb-3 text-lg text-heading">{{ t('settings.tabs.connections') }}</h2>

    <h3 class="mb-3 text-base text-text-secondary">{{ t('settings.connections.numisbids.title') }}</h3>
    <p class="mb-3 text-sm text-text-muted">
      {{ t('settings.connections.numisbids.intro') }}
    </p>
    <div class="form-group">
      <label class="form-label">{{ t('settings.connections.numisbids.username') }}</label>
      <input v-model="nbUsername" class="form-input" :placeholder="t('settings.connections.numisbids.usernamePlaceholder')" autocomplete="off" />
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('settings.connections.numisbids.password') }}</label>
      <input v-model="nbPassword" type="password" class="form-input" :placeholder="t('settings.connections.numisbids.passwordPlaceholder')" autocomplete="new-password" />
      <span class="mt-1 block text-chip text-text-muted">{{ t('settings.connections.numisbids.passwordHint') }}</span>
    </div>
    <div
      v-if="nbValidating"
      class="mt-1 rounded-sm border border-[color-mix(in_srgb,var(--text-warning)_20%,transparent)] bg-[color-mix(in_srgb,var(--text-warning)_10%,transparent)] px-3 py-[0.4rem] text-chip text-warning"
    >
      {{ t('settings.connections.numisbids.validating') }}
    </div>
    <div
      v-else-if="nbValidationError"
      class="mt-1 rounded-sm border border-[color-mix(in_srgb,var(--color-negative)_20%,transparent)] bg-[color-mix(in_srgb,var(--color-negative)_10%,transparent)] px-3 py-[0.4rem] text-chip text-loss"
    >
      {{ nbValidationError }}
    </div>
    <div
      v-else-if="auth.user?.numisBidsConfigured"
      class="mt-1 rounded-sm border border-[color-mix(in_srgb,var(--color-positive)_20%,transparent)] bg-[color-mix(in_srgb,var(--color-positive)_10%,transparent)] px-3 py-[0.4rem] text-chip text-gain"
    >
      {{ t('settings.connections.numisbids.connected') }}
    </div>

    <h3 class="mb-3 mt-5 text-base text-text-secondary">{{ t('settings.connections.cng.title') }}</h3>
    <p class="mb-3 text-sm text-text-muted">
      {{ t('settings.connections.cng.intro') }}
    </p>
    <div class="form-group">
      <label class="form-label">{{ t('settings.connections.cng.username') }}</label>
      <input v-model="cngUsername" class="form-input" :placeholder="t('settings.connections.cng.usernamePlaceholder')" autocomplete="off" />
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('settings.connections.cng.password') }}</label>
      <input v-model="cngPassword" type="password" class="form-input" :placeholder="t('settings.connections.cng.passwordPlaceholder')" autocomplete="new-password" />
      <span class="mt-1 block text-chip text-text-muted">{{ t('settings.connections.cng.passwordHint') }}</span>
    </div>
    <div
      v-if="cngValidating"
      class="mt-1 rounded-sm border border-[color-mix(in_srgb,var(--text-warning)_20%,transparent)] bg-[color-mix(in_srgb,var(--text-warning)_10%,transparent)] px-3 py-[0.4rem] text-chip text-warning"
    >
      {{ t('settings.connections.cng.validating') }}
    </div>
    <div
      v-else-if="cngValidationError"
      class="mt-1 rounded-sm border border-[color-mix(in_srgb,var(--color-negative)_20%,transparent)] bg-[color-mix(in_srgb,var(--color-negative)_10%,transparent)] px-3 py-[0.4rem] text-chip text-loss"
    >
      {{ cngValidationError }}
    </div>
    <div
      v-else-if="auth.user?.cngConfigured"
      class="mt-1 rounded-sm border border-[color-mix(in_srgb,var(--color-positive)_20%,transparent)] bg-[color-mix(in_srgb,var(--color-positive)_10%,transparent)] px-3 py-[0.4rem] text-chip text-gain"
    >
      {{ t('settings.connections.cng.connected') }}
    </div>

    <h3 class="mb-3 mt-5 text-base text-text-secondary">{{ t('settings.connections.parcel.title') }}</h3>
    <p class="mb-3 text-sm text-text-muted">
      {{ t('settings.connections.parcel.intro') }}
    </p>
    <div class="form-group">
      <label class="form-label">{{ t('settings.connections.parcel.key') }}</label>
      <input v-model="parcelAppAPIKey" type="password" class="form-input" :placeholder="t('settings.connections.parcel.keyPlaceholder')" autocomplete="new-password" />
      <span class="mt-1 block text-chip text-text-muted">{{ t('settings.connections.parcel.keyHint') }}</span>
    </div>
    <div
      v-if="auth.user?.parcelAppConfigured"
      class="mt-1 rounded-sm border border-[color-mix(in_srgb,var(--color-positive)_20%,transparent)] bg-[color-mix(in_srgb,var(--color-positive)_10%,transparent)] px-3 py-[0.4rem] text-chip text-gain"
    >
      {{ t('settings.connections.parcel.connected') }}
    </div>

    <h3 class="mb-3 mt-5 text-base text-text-secondary">{{ t('settings.connections.pushover.title') }}</h3>
    <p class="mb-3 text-sm text-text-muted">
      {{ t('settings.connections.pushover.intro') }}
    </p>
    <div class="form-group">
      <label class="form-label">{{ t('settings.connections.pushover.key') }}</label>
      <input v-model="pushoverKey" type="password" class="form-input" :placeholder="t('settings.connections.pushover.keyPlaceholder')" autocomplete="off" />
      <span class="mt-1 block text-chip text-text-muted">{{ t('settings.connections.pushover.keyHint') }}</span>
    </div>
    <div
      v-if="auth.user?.pushoverEnabled"
      class="mb-2 mt-1 rounded-sm border border-[color-mix(in_srgb,var(--color-positive)_20%,transparent)] bg-[color-mix(in_srgb,var(--color-positive)_10%,transparent)] px-3 py-[0.4rem] text-chip text-gain"
    >
      {{ t('settings.connections.pushover.active') }}
    </div>
    <button
      class="btn btn-secondary btn-sm mb-1 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
      :disabled="pushoverTesting || !auth.user?.pushoverEnabled"
      @click="handleTestPushover"
    >
      {{ pushoverTesting ? t('settings.connections.pushover.sending') : t('settings.connections.pushover.test') }}
    </button>
    <p
      v-if="pushoverTestMsg"
      class="mt-1 text-body text-gold"
      :class="{ 'text-loss': pushoverTestError }"
    >
      {{ pushoverTestMsg }}
    </p>

    <button
      class="btn btn-primary btn-sm mt-2 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
      @click="handleSaveProfile"
      :disabled="profileSaving || nbValidating || cngValidating"
    >
      {{ nbValidating || cngValidating ? t('settings.connections.validating') : profileSaving ? t('common.saving') : t('settings.connections.save') }}
    </button>
    <p v-if="profileMsg" class="mt-2 text-body text-gold" :class="{ 'text-loss': profileError }">{{ profileMsg }}</p>
  </section>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useSettingsProfile } from '@/composables/useSettingsProfile'
import { t } from '@/i18n'

const auth = useAuthStore()
const {
  nbUsername, nbPassword, cngUsername, cngPassword, parcelAppAPIKey, pushoverKey,
  pushoverTesting, pushoverTestMsg, pushoverTestError, handleTestPushover,
  profileMsg, profileError, profileSaving,
  nbValidating, nbValidationError, cngValidating, cngValidationError, handleSaveProfile,
} = useSettingsProfile()
</script>
