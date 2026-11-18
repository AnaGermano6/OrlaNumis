<template>
  <div class="flex h-11 shrink-0 rounded-md border border-border-subtle bg-input p-1" role="group" :aria-label="label">
    <button
      v-for="option in options"
      :key="option.value"
      type="button"
      class="inline-flex min-w-0 flex-1 cursor-pointer items-center justify-center gap-2 whitespace-nowrap rounded-sm px-4 text-body transition-colors focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
      :class="option.value === modelValue ? 'bg-gold font-semibold text-surface' : 'font-medium text-text-secondary hover:bg-card-hover hover:text-text-primary'"
      :aria-pressed="option.value === modelValue"
      :data-testid="testIdPrefix ? `${testIdPrefix}-${option.value}` : undefined"
      @click="emit('update:modelValue', option.value === modelValue && clearable ? null : option.value)"
    >
      <component :is="option.icon" v-if="option.icon" :size="16" aria-hidden="true" />
      {{ option.label }}
    </button>
  </div>
</template>

<script setup lang="ts" generic="T extends string">
// A row of buttons, one of them on: how a list shows (grid or list), which
// side of the coins. A clearable one turns off when its button is pressed again.
import type { Component } from 'vue'

withDefaults(defineProps<{
  options: { value: T; label: string; icon?: Component }[]
  modelValue: T | null
  /** The group's name, for screen readers. */
  label: string
  /** Pressing the button that is on turns it off. */
  clearable?: boolean
  testIdPrefix?: string
}>(), { clearable: false, testIdPrefix: undefined })

const emit = defineEmits<{ 'update:modelValue': [value: T | null] }>()
</script>
