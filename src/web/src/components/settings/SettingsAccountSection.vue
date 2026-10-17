<template>
  <section class="card text-text-primary">
    <h2 class="mb-5 border-b border-border-subtle pb-3 text-lg text-heading">{{ t('settings.tabs.account') }}</h2>

    <!-- Avatar -->
    <div class="flex items-center justify-between gap-4 border-b border-border-subtle py-3 last:border-0">
      <div class="shrink-0">
        <AuthenticatedImage :media-path="avatarUrl" :alt="t('settings.account.avatar')" class="h-16 w-16 rounded-full border-2 border-gold-dim object-cover" />
      </div>
      <div class="flex flex-wrap gap-2">
        <label class="btn btn-secondary btn-sm cursor-pointer focus-within:outline-2 focus-within:outline-gold focus-within:outline-offset-2">
          {{ t('settings.account.uploadAvatar') }}
          <input type="file" accept="image/*" hidden @change="handleAvatarUpload" />
        </label>
        <button
          v-if="auth.user?.avatarPath"
          class="btn btn-danger btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          @click="handleAvatarDelete"
        >
          {{ t('common.remove') }}
        </button>
      </div>
    </div>

    <div class="flex items-center justify-between gap-4 border-b border-border-subtle py-3 last:border-0">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('auth.username') }}</span>
        <span class="text-base text-text-secondary">{{ auth.user?.username }}</span>
      </div>
    </div>
    <div class="flex items-center justify-between gap-4 border-b border-border-subtle py-3 last:border-0">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('settings.account.role') }}</span>
        <span class="badge text-base text-text-secondary" :class="`badge-${auth.user?.role === 'admin' ? 'roman' : 'modern'}`">
          {{ roleLabel }}
        </span>
      </div>
    </div>

    <!-- Profile / Social Settings -->
    <h3 class="mb-3 mt-5 text-base text-text-secondary">{{ t('settings.account.profile') }}</h3>
    <div class="form-group">
      <label class="form-label">{{ t('auth.email') }}</label>
      <input v-model="profileEmail" type="email" class="form-input" placeholder="you@example.com" />
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('settings.account.bio') }}</label>
      <input v-model="profileBio" class="form-input" :placeholder="t('settings.account.bioPlaceholder')" maxlength="200" />
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('settings.account.zip') }}</label>
      <input v-model="profileZipCode" class="form-input" :placeholder="t('settings.account.zipPlaceholder')" maxlength="10" />
      <span class="mt-1 block text-chip text-text-muted">{{ t('settings.account.zipHint') }}</span>
    </div>

    <button
      class="btn btn-primary btn-sm mt-2 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
      @click="handleSaveProfile"
      :disabled="saveDisabled"
    >
      {{ saveLabel }}
    </button>
    <p v-if="profileMsg" class="mt-2 text-body text-gold" :class="{ 'text-loss': profileError }">{{ profileMsg }}</p>

    <h3 class="mb-3 mt-5 text-base text-text-secondary">{{ t('settings.account.changePassword') }}</h3>
    <form class="max-w-[350px]" @submit.prevent="handleChangePassword">
      <div class="form-group">
        <label class="form-label">{{ t('settings.account.currentPassword') }}</label>
        <input v-model="currentPassword" type="password" class="form-input" required />
      </div>
      <div class="form-group">
        <label class="form-label">{{ t('settings.account.newPassword') }}</label>
        <input v-model="newPassword" type="password" class="form-input" required minlength="6" />
      </div>
      <div class="form-group">
        <label class="form-label">{{ t('settings.account.confirmNewPassword') }}</label>
        <input v-model="confirmPassword" type="password" class="form-input" required />
      </div>
      <p v-if="passwordMsg" class="my-2 text-body text-gold" :class="{ 'text-loss': passwordError }">{{ passwordMsg }}</p>
      <button
        type="submit"
        class="btn btn-primary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        :disabled="passwordLoading"
      >
        {{ passwordLoading ? t('settings.account.changing') : t('settings.account.changePassword') }}
      </button>
    </form>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useSettingsProfile } from '@/composables/useSettingsProfile'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { t } from '@/i18n'

const auth = useAuthStore()

const {
  avatarUrl, handleAvatarUpload, handleAvatarDelete,
  profileEmail, profileBio, profileZipCode,
  profileMsg, profileError, profileSaving,
  handleSaveProfile,
  currentPassword, newPassword, confirmPassword,
  passwordMsg, passwordError, passwordLoading, handleChangePassword,
} = useSettingsProfile()

const roleLabel = computed(() => {
  const role = auth.user?.role
  if (role === 'admin') return t('settings.account.roles.admin')
  if (role === 'user') return t('settings.account.roles.user')
  return role
})

const saveDisabled = computed(() => profileSaving.value)
const saveLabel = computed(() => profileSaving.value ? t('common.saving') : t('settings.account.saveProfile'))
</script>
