<template>
  <div v-if="supported" class="grid gap-1.5">
    <BaseButton
      type="button"
      variant="secondary"
      size="sm"
      class="min-h-11 justify-self-start"
      :disabled="disabled || status === 'requesting'"
      :aria-label="buttonLabel"
      :aria-pressed="status !== 'idle'"
      @click="toggle"
    >
      <Square v-if="status !== 'idle'" :size="16" aria-hidden="true" />
      <Mic v-else :size="16" aria-hidden="true" />
      {{ visibleLabel }}
    </BaseButton>
    <p class="m-0 text-sm text-text-muted">
      {{ t('voice.privacy') }}
    </p>
    <p v-if="error" class="m-0 text-sm text-loss" role="alert">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, toRef } from 'vue'
import { Mic, Square } from 'lucide-vue-next'
import BaseButton from '@/components/ui/BaseButton.vue'
import { useVoiceDictation } from '@/composables/useVoiceDictation'
import { t } from '@/i18n'

const props = withDefaults(defineProps<{
  disabled?: boolean
}>(), {
  disabled: false,
})

const emit = defineEmits<{
  transcript: [text: string]
}>()

const {
  supported,
  status,
  error,
  start,
  stop,
} = useVoiceDictation({
  disabled: toRef(props, 'disabled'),
  onFinalText: text => emit('transcript', text),
})

const buttonLabel = computed(() =>
  status.value === 'idle' ? t('voice.start') : t('voice.stop'))
const visibleLabel = computed(() => {
  if (status.value === 'requesting') return t('voice.requesting')
  if (status.value === 'listening') return t('voice.listening')
  if (status.value === 'stopping') return t('voice.stopping')
  return t('voice.dictate')
})

function toggle() {
  if (status.value === 'idle') return start()
  else stop()
}
</script>
