<template>
  <div class="flex shrink items-center overflow-y-auto p-4 text-center text-text-secondary">
    <div class="flex w-full flex-col items-center gap-3">
      <Bot :size="32" />
      <p class="max-w-[300px] leading-relaxed">
        {{ t('chat.intro.lead') }}
      </p>
      <p class="m-0 max-w-[340px] text-center text-sm leading-relaxed text-text-muted">
        {{ t('chat.intro.sources') }}
      </p>
      <div class="mt-2 flex w-full flex-col gap-2">
        <button
          class="w-full rounded-sm border border-border-subtle bg-card px-3 py-2 text-left text-chip text-text-secondary transition-colors hover:border-border-accent hover:text-gold"
          @click="$emit('send', t('chat.intro.caesarPrompt'))"
        >
          {{ t('chat.intro.caesar') }}
        </button>
        <button
          class="w-full rounded-sm border border-border-subtle bg-card px-3 py-2 text-left text-chip text-text-secondary transition-colors hover:border-border-accent hover:text-gold"
          @click="$emit('send', t('chat.intro.solidiPrompt'))"
        >
          {{ t('chat.intro.solidi') }}
        </button>
        <button
          class="w-full rounded-sm border border-border-subtle bg-card px-3 py-2 text-left text-chip text-text-secondary transition-colors hover:border-border-accent hover:text-gold"
          @click="$emit('send', t('chat.intro.missingEraPrompt'))"
        >
          {{ t('chat.intro.missingEra') }}
        </button>
        <button
          class="w-full rounded-sm border border-border-subtle bg-card px-3 py-2 text-left text-chip text-text-secondary transition-colors hover:border-border-accent hover:text-gold"
          @click="$emit('send', t('chat.intro.gapsPrompt'))"
        >
          {{ t('chat.intro.gaps') }}
        </button>
        <button
          class="w-full rounded-sm border border-border-subtle bg-card px-3 py-2 text-left text-chip text-text-secondary transition-colors hover:border-border-accent hover:text-gold"
          @click="$emit('send', t('chat.intro.owlPrompt'))"
        >
          {{ t('chat.intro.owl') }}
        </button>
        <button
          class="w-full rounded-sm border border-border-subtle bg-card px-3 py-2 text-left text-chip text-text-secondary transition-colors hover:border-border-accent hover:text-gold"
          @click="$emit('send-portfolio')"
        >
          {{ t('chat.intro.portfolio') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Bot } from 'lucide-vue-next'
import { t } from '@/i18n'

defineEmits<{
  send: [message: string]
  'send-portfolio': []
}>()
</script>
