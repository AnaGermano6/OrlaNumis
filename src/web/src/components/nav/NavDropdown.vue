<template>
  <div ref="root" class="relative" @mouseenter="hover && show()" @mouseleave="hover && hide()" @keydown.esc="hide(true)">
    <button
      ref="button"
      type="button"
      :class="buttonClass"
      :aria-expanded="open"
      :aria-controls="panelId"
      :aria-label="label"
      :title="label"
      @click="hover || !open ? show() : hide()"
    >
      <slot name="button" />
    </button>
    <!-- The gap above the list is part of it, so the pointer can cross it. -->
    <div v-if="open" :id="panelId" class="absolute top-full z-(--z-popover) pt-1" :class="align === 'right' ? 'right-0' : 'left-0'">
      <div class="min-w-[13rem] rounded-md border border-border-subtle bg-card py-1 shadow-card" @click="hide()">
        <slot />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// A button of the top bar that opens a list under it: on a click, or as the
// pointer passes over it (then a click keeps it open: the pointer opened it
// on its way). A click elsewhere, Escape or another page closes it.
import { onBeforeUnmount, ref, useId, watch } from 'vue'
import { useRoute } from 'vue-router'

withDefaults(defineProps<{
  /** The button's name, for screen readers and as its tooltip. */
  label: string
  buttonClass?: string | string[] | Record<string, boolean>
  /** Opens as the pointer passes over it. */
  hover?: boolean
  align?: 'left' | 'right'
}>(), { buttonClass: '', hover: false, align: 'left' })

const root = ref<HTMLElement | null>(null)
const button = ref<HTMLElement | null>(null)
const open = ref(false)
const panelId = `nav-dropdown-${useId()}`
const route = useRoute()

function onDocumentClick(event: MouseEvent) {
  if (root.value && !root.value.contains(event.target as Node)) hide()
}

function show() {
  if (open.value) return
  open.value = true
  document.addEventListener('click', onDocumentClick)
}

function hide(focusButton = false) {
  if (!open.value) return
  open.value = false
  document.removeEventListener('click', onDocumentClick)
  if (focusButton) button.value?.focus()
}

watch(() => route.fullPath, () => hide())
onBeforeUnmount(() => document.removeEventListener('click', onDocumentClick))
</script>
