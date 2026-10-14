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

    <div class="flex items-center justify-between gap-4 border-b border-border-subtle py-3 last:border-0">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('settings.account.publicCollection') }}</span>
        <span class="text-sm text-text-muted">{{ t('settings.account.publicHint') }}</span>
      </div>
      <BaseToggle
        :label="t('settings.account.publicCollection')"
        :model-value="profilePublic"
        @change="onPublicToggle"
      />
    </div>
    <div class="flex items-center justify-between gap-4 border-b border-border-subtle py-3 last:border-0">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('featured.title') }}</span>
        <span class="text-sm text-text-muted">{{ t('settings.account.coinOfDayHint') }}</span>
      </div>
      <BaseToggle
        :label="t('featured.title')"
        v-model="coinOfDayEnabled"
      />
    </div>
    <div v-if="coinOfDayEnabled" class="flex items-center justify-between gap-4 border-b border-border-subtle py-3 pl-4 last:border-0">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-sm font-medium">{{ t('settings.account.includeWishlist') }}</span>
        <span class="text-xs text-text-muted">{{ t('settings.account.includeWishlistHint') }}</span>
      </div>
      <BaseToggle
        size="sm"
        :label="t('settings.account.includeWishlist')"
        v-model="coinOfDayIncludeWishlist"
      />
    </div>
    <div class="flex items-center justify-between gap-4 border-b border-border-subtle py-3 last:border-0">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('coinForm.emperorTracker') }}</span>
        <span class="text-sm text-text-muted">{{ t('settings.account.emperorHint') }}</span>
      </div>
      <BaseToggle
        :label="t('coinForm.emperorTracker')"
        v-model="emperorTrackerEnabled"
      />
    </div>
    <template v-if="emperorTrackerEnabled">
      <div class="flex items-center justify-between gap-4 border-b border-border-subtle py-3 pl-4 last:border-0">
        <div class="flex flex-col gap-[0.15rem]">
          <span class="text-sm font-medium">{{ t('settings.account.trackUsurpers') }}</span>
          <span class="text-xs text-text-muted">{{ t('settings.account.trackUsurpersHint') }}</span>
        </div>
        <BaseToggle
          size="sm"
          :label="t('settings.account.trackUsurpers')"
          v-model="emperorTrackerShowUsurpers"
        />
      </div>
      <div class="flex items-center justify-between gap-4 border-b border-border-subtle py-3 pl-4 last:border-0">
        <div class="flex flex-col gap-[0.15rem]">
          <span class="text-sm font-medium">{{ t('settings.account.trackEmpresses') }}</span>
          <span class="text-xs text-text-muted">{{ t('settings.account.trackEmpressesHint') }}</span>
        </div>
        <BaseToggle
          size="sm"
          :label="t('settings.account.trackEmpresses')"
          v-model="emperorTrackerShowEmpresses"
        />
      </div>
      <div class="flex items-center justify-between gap-4 border-b border-border-subtle py-3 pl-4 last:border-0">
        <div class="flex flex-col gap-[0.15rem]">
          <span class="text-sm font-medium">{{ t('settings.account.trackOther') }}</span>
          <span class="text-xs text-text-muted">{{ t('settings.account.trackOtherHint') }}</span>
        </div>
        <BaseToggle
          size="sm"
          :label="t('settings.account.trackOther')"
          v-model="emperorTrackerShowOtherFigures"
        />
      </div>
    </template>

    <button
      class="btn btn-primary btn-sm mt-2 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
      @click="handleSaveProfile"
      :disabled="saveDisabled"
    >
      {{ saveLabel }}
    </button>
    <p v-if="profileMsg" class="mt-2 text-body text-gold" :class="{ 'text-loss': profileError }">{{ profileMsg }}</p>

    <!-- Privacy Warning Modal -->
    <Teleport to="body">
      <div v-if="showPrivacyWarning" class="fixed inset-0 z-[1000] flex items-start justify-center bg-[var(--overlay-dark)] p-4 pt-[15vh]" @click.self="cancelGoPrivate">
        <div class="card w-full max-w-[440px]">
          <div class="border-b border-border-subtle px-5 py-4">
            <h2 class="m-0 flex items-center gap-2 text-base text-heading">
              {{ t('settings.account.makePrivateTitle') }}
            </h2>
          </div>
          <div class="p-5">
            <p class="mb-3 mt-0 leading-6 text-text-secondary">
              <I18nT keypath="settings.account.makePrivateBody">
                <template #strong><strong class="text-text-primary">{{ t('settings.account.removeFollowers') }}</strong></template>
              </I18nT>
            </p>
            <p class="mb-4 mt-0 leading-6 text-text-secondary">
              {{ t('settings.account.hiddenFromSearch') }}
            </p>
            <div class="flex justify-end gap-3">
              <button class="btn btn-secondary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" @click="cancelGoPrivate">{{ t('common.cancel') }}</button>
              <button class="btn btn-danger btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" @click="confirmGoPrivate">{{ t('settings.account.makePrivate') }}</button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

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

    <h3 class="mb-3 mt-5 text-base text-text-secondary">{{ t('settings.account.providers') }}</h3>
    <p class="mt-2 text-sm text-text-muted">
      {{ t('settings.account.providersHint') }}
    </p>

    <div v-if="oidcMsg" class="my-2 text-body text-gold" :class="{ 'text-loss': oidcError }" role="status">
      {{ oidcMsg }}
    </div>

    <div v-if="oidcLoading" class="my-2 text-body text-gold">
      {{ t('settings.account.loadingProviders') }}
    </div>
    <div v-else>
      <div v-if="oidcIdentities.length" class="mt-4 flex flex-col gap-3">
        <div
          v-for="identity in oidcIdentities"
          :key="identity.id"
          class="flex flex-col items-stretch gap-3 rounded-sm border border-border-subtle bg-input p-3 md:flex-row md:items-start md:justify-between"
        >
          <div class="flex min-w-0 flex-col gap-[0.2rem]">
            <div class="flex flex-wrap items-center gap-[0.35rem]">
              <span class="text-base font-medium text-text-primary">{{ identity.providerDisplayName }}</span>
              <span
                class="chip-sm"
                :class="identity.emailVerified ? 'border-[var(--color-positive)] text-gain' : 'border-[var(--color-negative)] text-loss'"
              >
                {{ identity.emailVerified ? t('settings.account.emailVerified') : t('settings.account.emailUnverified') }}
              </span>
            </div>
            <span class="text-sm text-text-muted [overflow-wrap:anywhere]">{{ t('settings.account.issuer', { value: identity.issuer }) }}</span>
            <span class="text-sm text-text-muted [overflow-wrap:anywhere]">{{ t('settings.account.subject', { value: identity.subjectPreview }) }}</span>
            <span class="text-sm text-text-muted [overflow-wrap:anywhere]">{{ t('settings.account.emailValue', { value: identity.email }) }}</span>
            <span class="text-sm text-text-muted [overflow-wrap:anywhere]">
              {{ t('settings.account.linked', { date: formatDateTime(identity.createdAt) }) }}
              <template v-if="identity.lastLoginAt"> · {{ t('settings.account.lastLogin', { date: formatDateTime(identity.lastLoginAt) }) }}</template>
            </span>
          </div>
          <button
            class="btn btn-danger btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
            :disabled="unlinkingIdentityId === identity.id"
            @click="handleUnlinkIdentity(identity.id, identity.providerDisplayName)"
          >
            {{ unlinkingIdentityId === identity.id ? t('settings.account.unlinking') : t('settings.account.unlink') }}
          </button>
        </div>
      </div>
      <p v-else class="mt-2 text-sm text-text-muted">{{ t('settings.account.noProviders') }}</p>

      <div v-if="linkableProviders.length" class="mt-3 flex flex-wrap gap-2">
        <button
          v-for="provider in linkableProviders"
          :key="provider.id"
          type="button"
          class="btn btn-secondary btn-sm inline-flex items-center gap-[0.35rem] focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :disabled="linkingProviderId === provider.id"
          @click="handleLinkProvider(provider.id, provider.displayName)"
        >
          <LinkIcon :size="16" aria-hidden="true" />
          {{ linkingProviderId === provider.id ? t('settings.account.starting') : t('settings.account.linkProvider', { name: provider.displayName }) }}
        </button>
      </div>
      <p v-else-if="!oidcProviders.length" class="mt-2 text-sm text-text-muted">
        {{ t('settings.account.noOidc') }}
      </p>
    </div>

    <template v-if="supportsWebAuthn">
      <h3 class="mb-3 mt-5 text-base text-text-secondary">{{ t('settings.account.biometric') }}</h3>
      <p class="mb-3 text-sm text-text-muted">
        {{ t('settings.account.biometricHint') }}
      </p>

      <button
        class="btn btn-primary btn-sm inline-flex items-center gap-[0.35rem] focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        :disabled="registeringCredential"
        @click="handleRegisterCredential"
      >
        <LockKeyhole v-if="!registeringCredential" :size="16" aria-hidden="true" />
        {{ registeringCredential ? t('settings.account.registering') : t('settings.account.registerBiometric') }}
      </button>
      <p v-if="credentialMsg" class="mt-2 text-body text-gold" :class="{ 'text-loss': credentialError }">{{ credentialMsg }}</p>

      <div v-if="webauthnCredentials.length" class="mt-4 flex flex-col gap-2">
        <div
          v-for="cred in webauthnCredentials"
          :key="cred.id"
          class="flex items-center justify-between gap-3 border-b border-border-subtle py-[0.6rem] last:border-0"
        >
          <div class="flex min-w-0 flex-col gap-[0.1rem]">
            <span class="text-base font-medium">{{ cred.name }}</span>
            <span class="text-sm text-text-muted">{{ t('settings.account.registered', { date: formatDate(cred.createdAt) }) }}</span>
          </div>
          <button class="btn btn-danger btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" @click="handleDeleteCredential(cred.id)">
            {{ t('common.remove') }}
          </button>
        </div>
      </div>
      <p v-else-if="!registeringCredential" class="mt-2 text-sm text-text-muted">{{ t('settings.account.noBiometric') }}</p>
    </template>
  </section>
</template>

<script setup lang="ts">
import { BaseToggle } from '@/components/ui'
import { computed, ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import {
  webauthnRegisterBegin, webauthnRegisterFinish,
  webauthnListCredentials, webauthnDeleteCredential,
  deleteOIDCIdentity, getApiErrorMessage, getOIDCIdentities, getOIDCPublicProviders, startOIDCLink,
} from '@/api/client'
import { useDialog } from '@/composables/useDialog'
import { useSettingsProfile } from '@/composables/useSettingsProfile'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import I18nT from '@/components/I18nT.vue'
import { t, intlLocale } from '@/i18n'
import type { OIDCLinkedIdentity, OIDCPublicProvider, WebAuthnCredentialInfo } from '@/types'
import { Link as LinkIcon, LockKeyhole } from 'lucide-vue-next'

const auth = useAuthStore()
const { showConfirm } = useDialog()

const {
  avatarUrl, handleAvatarUpload, handleAvatarDelete,
  profileEmail, profileBio, profileZipCode,
  profilePublic, profileMsg, profileError, profileSaving,
  showPrivacyWarning, onPublicToggle, confirmGoPrivate, cancelGoPrivate,
  handleSaveProfile, coinOfDayEnabled,
  coinOfDayIncludeWishlist,
  emperorTrackerEnabled, emperorTrackerShowUsurpers, emperorTrackerShowEmpresses, emperorTrackerShowOtherFigures,
  currentPassword, newPassword, confirmPassword,
  passwordMsg, passwordError, passwordLoading, handleChangePassword,
} = useSettingsProfile()

// WebAuthn Biometric
const supportsWebAuthn = !!window.PublicKeyCredential
const webauthnCredentials = ref<WebAuthnCredentialInfo[]>([])
const registeringCredential = ref(false)
const credentialMsg = ref('')
const credentialError = ref(false)
const oidcIdentities = ref<OIDCLinkedIdentity[]>([])
const oidcProviders = ref<OIDCPublicProvider[]>([])
const oidcLoading = ref(true)
const oidcMsg = ref('')
const oidcError = ref(false)
const linkingProviderId = ref<number | null>(null)
const unlinkingIdentityId = ref<number | null>(null)

const linkableProviders = computed(() => {
  const linkedProviderIds = new Set(oidcIdentities.value.map(identity => identity.providerId))
  return oidcProviders.value.filter(provider => !linkedProviderIds.has(provider.id))
})

const roleLabel = computed(() => {
  const role = auth.user?.role
  if (role === 'admin') return t('settings.account.roles.admin')
  if (role === 'user') return t('settings.account.roles.user')
  return role
})

const saveDisabled = computed(() => profileSaving.value)
const saveLabel = computed(() => profileSaving.value ? t('common.saving') : t('settings.account.saveProfile'))

async function loadOIDCAccounts() {
  oidcLoading.value = true
  try {
    const [identitiesResponse, providersResponse] = await Promise.all([
      getOIDCIdentities(),
      getOIDCPublicProviders(),
    ])
    oidcIdentities.value = identitiesResponse.data.identities ?? []
    oidcProviders.value = providersResponse.data.providers ?? []
  } catch (error: unknown) {
    oidcMsg.value = getApiErrorMessage(error) || t('settings.account.providersLoadFailed')
    oidcError.value = true
  } finally {
    oidcLoading.value = false
  }
}

async function handleLinkProvider(providerId: number, displayName: string) {
  oidcMsg.value = ''
  oidcError.value = false
  linkingProviderId.value = providerId
  try {
    const response = await startOIDCLink(providerId, {
      redirectPath: '/settings?tab=account',
      callbackPath: `/settings/oidc/link/callback/${providerId}`,
    })
    const authorizationUrl = response.data.authorizationUrl
    if (!authorizationUrl) {
      oidcMsg.value = t('settings.account.noAuthUrl', { name: displayName })
      oidcError.value = true
      return
    }
    window.location.assign(authorizationUrl)
  } catch (error: unknown) {
    oidcMsg.value = mapOIDCAccountError(error, 'link')
    oidcError.value = true
  } finally {
    linkingProviderId.value = null
  }
}

async function handleUnlinkIdentity(identityId: number, displayName: string) {
  const confirmed = await showConfirm(
    t('settings.account.confirmUnlink', { name: displayName }),
    { title: t('settings.account.unlinkTitle'), variant: 'danger' },
  )
  if (!confirmed) return

  oidcMsg.value = ''
  oidcError.value = false
  unlinkingIdentityId.value = identityId
  try {
    await deleteOIDCIdentity(identityId)
    await loadOIDCAccounts()
    oidcMsg.value = t('settings.account.unlinked', { name: displayName })
  } catch (error: unknown) {
    oidcMsg.value = mapOIDCAccountError(error, 'unlink')
    oidcError.value = true
  } finally {
    unlinkingIdentityId.value = null
  }
}

function mapOIDCAccountError(error: unknown, action: 'link' | 'unlink') {
  const response = getErrorResponse(error)
  const message = getApiErrorMessage(error)
  const normalized = message.toLowerCase()

  if (response?.status === 409 && action === 'link') {
    if (normalized.includes('another user') || normalized.includes('already linked')) {
      return t('settings.account.errors.linkedElsewhere')
    }
    return t('settings.account.errors.cannotLink')
  }

  if (response?.status === 409 && action === 'unlink') {
    return t('settings.account.errors.lastMethod')
  }

  if (response?.status === 404) {
    return t('settings.account.errors.notFound')
  }

  if (normalized.includes('state') || normalized.includes('claims') || response?.status === 400) {
    return t('settings.account.errors.validation')
  }

  if (normalized.includes('configuration') || normalized.includes('discovery') || response?.status === 500) {
    return t('settings.account.errors.misconfigured')
  }

  return message || (action === 'link' ? t('settings.account.errors.linkFailed') : t('settings.account.errors.unlinkFailed'))
}

function getErrorResponse(error: unknown): { status?: number } | null {
  if (typeof error !== 'object' || error === null || !('response' in error)) return null
  const response = (error as { response?: unknown }).response
  if (typeof response !== 'object' || response === null) return null
  return response as { status?: number }
}

async function loadCredentials() {
  try {
    const res = await webauthnListCredentials()
    webauthnCredentials.value = res.data
  } catch {
    // silently fail
  }
}

function base64urlToBuffer(base64url: string): ArrayBuffer {
  const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/')
  const pad = base64.length % 4 === 0 ? '' : '='.repeat(4 - (base64.length % 4))
  const binary = atob(base64 + pad)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
  return bytes.buffer
}

async function handleRegisterCredential() {
  registeringCredential.value = true
  credentialMsg.value = ''
  credentialError.value = false

  try {
    const beginRes = await webauthnRegisterBegin()
    const options = beginRes.data
    const publicKey = options.publicKey
    if (!publicKey?.challenge || !publicKey.user?.id) {
      throw new Error(t('settings.account.errors.challenge'))
    }

    const publicKeyOptions: PublicKeyCredentialCreationOptions = {
      challenge: base64urlToBuffer(publicKey.challenge),
      rp: publicKey.rp,
      user: {
        id: base64urlToBuffer(publicKey.user.id),
        name: publicKey.user.name,
        displayName: publicKey.user.displayName,
      },
      pubKeyCredParams: publicKey.pubKeyCredParams,
      timeout: publicKey.timeout || 60000,
      authenticatorSelection: publicKey.authenticatorSelection,
      attestation: publicKey.attestation || 'none',
      excludeCredentials: (publicKey.excludeCredentials || []).map((c: { id: string; type: string; transports?: string[] }) => ({
        id: base64urlToBuffer(c.id),
        type: c.type,
        transports: c.transports,
      })),
    }

    const credential = await navigator.credentials.create({
      publicKey: publicKeyOptions,
    }) as PublicKeyCredential

    await webauthnRegisterFinish(credential)

    credentialMsg.value = t('settings.account.credentialRegistered')
    await loadCredentials()
  } catch (e: unknown) {
    credentialMsg.value = e instanceof Error ? e.message : t('settings.account.registrationFailed')
    credentialError.value = true
  } finally {
    registeringCredential.value = false
  }
}

async function handleDeleteCredential(id: number) {
  if (!await showConfirm(t('settings.account.confirmRemoveCredential'), { title: t('settings.account.removeCredentialTitle') })) return
  try {
    await webauthnDeleteCredential(id)
    await loadCredentials()
  } catch {
    credentialMsg.value = t('settings.account.removeCredentialFailed')
    credentialError.value = true
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString(intlLocale(), {
    year: 'numeric', month: 'short', day: 'numeric',
  })
}

function formatDateTime(dateStr: string) {
  return new Date(dateStr).toLocaleString(intlLocale(), {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

onMounted(() => {
  if (supportsWebAuthn) loadCredentials()
  void loadOIDCAccounts()
})

defineExpose({ loadCredentials, loadOIDCAccounts })
</script>
