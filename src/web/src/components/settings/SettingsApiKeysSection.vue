<template>
  <section class="card">
    <h2 class="text-xl font-medium mb-5 pb-3 border-b border-border-subtle">{{ t('settings.tabs.apikeys') }}</h2>
    <p class="text-sm text-text-muted mb-4">
      <I18nT keypath="settings.apiKeys.intro">
        <template #header><code>X-API-Key</code></template>
      </I18nT>
    </p>

    <!-- Generate form -->
    <div class="flex flex-wrap items-center gap-3 mb-3">
      <input
        v-model="apiKeyName"
        type="text"
        class="form-input flex-1 min-w-[200px]"
        :placeholder="t('settings.apiKeys.namePlaceholder')"
        :disabled="generatingKey"
      />
      <div class="flex items-center gap-[0.35rem]">
        <button
          type="button"
          class="chip"
          :class="{ active: apiKeyScope === 'read' }"
          :disabled="generatingKey"
          @click="apiKeyScope = 'read'"
        >
          {{ t('settings.apiKeys.scopes.read') }}
        </button>
        <button
          type="button"
          class="chip"
          :class="{ active: apiKeyScope === 'read,copilot' }"
          :disabled="generatingKey"
          @click="apiKeyScope = 'read,copilot'"
        >
          {{ t('settings.apiKeys.scopes.readCopilot') }}
        </button>
        <button
          type="button"
          class="chip"
          :class="{ active: apiKeyScope === 'read,write' }"
          :disabled="generatingKey"
          @click="apiKeyScope = 'read,write'"
        >
          {{ t('settings.apiKeys.scopes.readWrite') }}
        </button>
      </div>
      <button
        class="btn btn-primary btn-sm inline-flex items-center gap-[0.35rem]"
        :disabled="!apiKeyName.trim() || generatingKey"
        @click="handleGenerateKey"
      >
        <KeyRound :size="14" />
        {{ generatingKey ? t('settings.apiKeys.generating') : t('settings.apiKeys.generate') }}
      </button>
    </div>

    <!-- Newly generated key reveal -->
    <div
      v-if="newlyGeneratedKey"
      class="bg-surface border border-gold-dim rounded-sm px-4 py-3 mb-3"
    >
      <p class="text-chip text-gold font-medium mb-2">
        {{ t('settings.apiKeys.copyNow') }}
      </p>
      <div class="flex items-center gap-2">
        <code class="flex-1 text-chip bg-card px-[0.6rem] py-[0.4rem] rounded-sm break-all select-all">
          {{ newlyGeneratedKey }}
        </code>
        <button
          class="btn btn-secondary btn-sm inline-flex items-center gap-[0.35rem]"
          @click="copyKey"
        >
          <Check v-if="keyCopied" :size="14" />
          <Clipboard v-else :size="14" />
          {{ keyCopied ? t('common.copied') : t('common.copy') }}
        </button>
      </div>
    </div>

    <p
      v-if="apiKeyMsg"
      class="text-body my-2"
      :class="apiKeyError ? 'text-byzantine' : 'text-gold'"
    >
      {{ apiKeyMsg }}
    </p>

    <!-- Key list -->
    <div v-if="apiKeys.length" class="flex flex-col gap-2 mt-4">
      <div
        v-for="key in apiKeys"
        :key="key.id"
        class="flex justify-between items-center py-[0.6rem] border-b border-border-subtle last:border-0 gap-3"
        :class="{ 'opacity-50': key.revokedAt }"
      >
        <div class="flex flex-col gap-[0.1rem] min-w-0">
          <div class="flex items-center gap-2">
            <span class="text-base font-medium">{{ key.name }}</span>
            <span
              class="chip-sm shrink-0"
              :class="key.capabilities !== 'read'
                ? 'bg-gold-glow text-gold border border-gold-dim'
                : 'bg-input text-text-secondary border border-border-subtle'"
            >
              {{ capabilityLabel(key.capabilities) }}
            </span>
          </div>
          <span class="text-sm text-text-muted">
            ...{{ key.keyPrefix }}
            · {{ t('settings.apiKeys.created', { date: formatDate(key.createdAt) }) }}
            <template v-if="key.lastUsedAt"> · {{ t('settings.apiKeys.lastUsed', { date: formatDate(key.lastUsedAt) }) }}</template>
          </span>
        </div>
        <span
          v-if="key.revokedAt"
          class="chip-sm shrink-0 bg-surface text-text-muted"
        >{{ t('settings.apiKeys.revoked') }}</span>
        <button
          v-else
          class="btn btn-danger btn-sm"
          @click="handleRevokeKey(key.id)"
        >
          {{ t('settings.apiKeys.revoke') }}
        </button>
      </div>
    </div>
    <p v-else-if="!generatingKey" class="text-sm text-text-muted mt-2">{{ t('settings.apiKeys.empty') }}</p>
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Check, Clipboard, KeyRound } from 'lucide-vue-next'
import { generateApiKey, listApiKeys, revokeApiKey } from '@/api/client'
import type { ApiKey, ApiKeyCapability } from '@/types'
import I18nT from '@/components/I18nT.vue'
import { t, intlLocale } from '@/i18n'

const apiKeys = ref<ApiKey[]>([])
const apiKeyName = ref('')
const apiKeyScope = ref<ApiKeyCapability>('read')
const newlyGeneratedKey = ref('')
const keyCopied = ref(false)
const generatingKey = ref(false)
const apiKeyMsg = ref('')
const apiKeyError = ref(false)

async function loadApiKeys() {
  try {
    const res = await listApiKeys()
    apiKeys.value = res.data
  } catch {
    // silently fail on load
  }
}

async function handleGenerateKey() {
  if (!apiKeyName.value.trim()) return

  generatingKey.value = true
  apiKeyMsg.value = ''
  apiKeyError.value = false
  newlyGeneratedKey.value = ''
  keyCopied.value = false

  try {
    const res = await generateApiKey(apiKeyName.value.trim(), apiKeyScope.value)
    newlyGeneratedKey.value = res.data.key
    apiKeyName.value = ''
    apiKeyScope.value = 'read'
    await loadApiKeys()
  } catch {
    apiKeyMsg.value = t('settings.apiKeys.generateFailed')
    apiKeyError.value = true
  } finally {
    generatingKey.value = false
  }
}

async function copyKey() {
  try {
    await navigator.clipboard.writeText(newlyGeneratedKey.value)
    keyCopied.value = true
    setTimeout(() => { keyCopied.value = false }, 3000)
  } catch {
    const textarea = document.createElement('textarea')
    textarea.value = newlyGeneratedKey.value
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    keyCopied.value = true
    setTimeout(() => { keyCopied.value = false }, 3000)
  }
}

async function handleRevokeKey(id: number) {
  apiKeyMsg.value = ''
  apiKeyError.value = false
  try {
    await revokeApiKey(id)
    await loadApiKeys()
    newlyGeneratedKey.value = ''
  } catch {
    apiKeyMsg.value = t('settings.apiKeys.revokeFailed')
    apiKeyError.value = true
  }
}

function capabilityLabel(capabilities: string): string {
  switch (capabilities) {
    case 'read,copilot':
      return t('settings.apiKeys.scopes.readCopilot')
    case 'read,write':
      return t('settings.apiKeys.scopes.readWrite')
    case 'read,write,copilot':
      return t('settings.apiKeys.scopes.readWriteCopilot')
    default:
      return t('settings.apiKeys.scopes.read')
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString(intlLocale(), {
    year: 'numeric', month: 'short', day: 'numeric',
  })
}

onMounted(() => {
  loadApiKeys()
})

defineExpose({ loadApiKeys })
</script>
