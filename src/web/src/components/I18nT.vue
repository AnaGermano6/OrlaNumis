<template>
  <template v-for="(part, index) in parts" :key="index">
    <slot v-if="part.slot" :name="part.slot" />
    <template v-else>{{ part.text }}</template>
  </template>
</template>

<script setup lang="ts">
import { computed, useSlots } from 'vue'
import { t, type MessageKey, type MessageParams } from '@/i18n'

// Renders a translated sentence whose {placeholders} are filled by same-named
// slots, so links, <strong> and icons can sit inside text whose word order
// differs between languages:
//   <I18nT keypath="pwa.install.iosSafari"><template #share><strong>…</strong></template></I18nT>
const props = defineProps<{ keypath: MessageKey; params?: MessageParams }>()
const slots = useSlots()

const parts = computed(() =>
  t(props.keypath, props.params)
    .split(/(\{\w+\})/)
    .filter(Boolean)
    .map((chunk) => {
      const name = chunk.match(/^\{(\w+)\}$/)?.[1]
      return name && slots[name] ? { slot: name } : { text: chunk }
    }),
)
</script>
