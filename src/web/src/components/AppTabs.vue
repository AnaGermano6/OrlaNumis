<template>
  <!-- The rule under the tabs is a shadow: a border would make the bar scroll by a pixel. -->
  <div
    ref="root"
    class="flex overflow-x-auto shadow-[inset_0_-1px_0_var(--border-subtle)]"
    role="tablist"
    :aria-label="label"
    @keydown="onKeydown"
  >
    <button
      v-for="item in tabs"
      :id="`${idPrefix}-tab-${item.id}`"
      :key="item.id"
      type="button"
      role="tab"
      class="shrink-0 whitespace-nowrap border-b-2 px-4 py-3 text-base transition-colors"
      :class="[
        stretch ? 'md:flex-1' : '',
        item.id === modelValue ? 'border-gold font-medium text-text-primary' : 'border-transparent text-text-secondary hover:text-text-primary',
      ]"
      :aria-selected="item.id === modelValue"
      :aria-controls="`${idPrefix}-panel-${item.id}`"
      :tabindex="item.id === modelValue ? 0 : -1"
      :data-testid="`${idPrefix}-tab-${item.id}`"
      @click="emit('update:modelValue', item.id)"
    >
      {{ item.label }}
    </button>
  </div>
</template>

<script setup lang="ts" generic="T extends string">
// A row of tabs, one panel showing at a time. Each tab is
// "<idPrefix>-tab-<id>" and names its panel, "<idPrefix>-panel-<id>", which
// the page lays out. The arrow keys, Home and End go from tab to tab.
import { ref } from 'vue'

const props = withDefaults(defineProps<{
  tabs: { id: T; label: string }[]
  modelValue: T
  /** Starts the tabs' and the panels' ids. */
  idPrefix: string
  /** The tabs' name, for screen readers. */
  label: string
  /** The tabs share the row's width on a wider screen. */
  stretch?: boolean
}>(), { stretch: false })

const emit = defineEmits<{ 'update:modelValue': [id: T] }>()

const root = ref<HTMLElement | null>(null)

function onKeydown(event: KeyboardEvent) {
  const ids = props.tabs.map((item) => item.id)
  const at = ids.indexOf(props.modelValue)
  const index = event.key === 'ArrowRight' ? (at + 1) % ids.length
    : event.key === 'ArrowLeft' ? (at - 1 + ids.length) % ids.length
      : event.key === 'Home' ? 0
        : event.key === 'End' ? ids.length - 1
          : -1
  if (index < 0 || !ids.length) return
  event.preventDefault()
  const next = ids[index]!
  emit('update:modelValue', next)
  root.value?.querySelector<HTMLElement>(`[data-testid="${props.idPrefix}-tab-${next}"]`)?.focus()
}
</script>
