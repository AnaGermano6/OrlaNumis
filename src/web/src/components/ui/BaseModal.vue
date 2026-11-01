<template>
  <Teleport to="body">
    <Transition name="base-modal">
      <div
        v-if="open"
        class="base-modal fixed inset-0 flex justify-center bg-overlay-60 backdrop-blur-[2px]"
        :class="[layer === 'dialog' ? 'z-(--z-dialog)' : 'z-(--z-overlay)', sheet ? 'items-end md:items-center md:p-4' : 'items-center p-4']"
        @click.self="dismiss"
      >
        <div
          ref="panel"
          :role="alert ? 'alertdialog' : 'dialog'"
          aria-modal="true"
          :aria-label="ariaLabel || undefined"
          :aria-labelledby="ariaLabelledby || undefined"
          tabindex="-1"
          class="base-modal-panel flex w-full min-w-0 flex-col overflow-y-auto overscroll-contain border border-border-subtle bg-card shadow-card outline-none"
          :class="[maxWidth, sheet ? 'is-sheet' : 'is-centered', { 'is-padded': padded }]"
          :data-testid="testid || undefined"
        >
          <slot :close="dismiss" />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
// Every dialog and sheet of the app. On a phone it rises from the bottom
// (a sheet) at most 90% of the screen high, scrolling on its own; on a wider
// screen it is centred. While open the page behind does not scroll, Escape
// and a tap outside close it, and the agent's button stays hidden.
import { nextTick, ref, toRef, watch } from 'vue'
import { useOverlayLayer } from '@/composables/useOverlayLayer'

const props = withDefaults(defineProps<{
  open: boolean
  ariaLabel?: string
  ariaLabelledby?: string
  /** Width on a wider screen, e.g. 'md:max-w-[440px]'. */
  maxWidth?: string
  /** Rises from the bottom on a phone; false keeps it centred (a short confirmation). */
  sheet?: boolean
  padded?: boolean
  /** 'dialog' opens over the other overlays (the chat, another dialog). */
  layer?: 'overlay' | 'dialog'
  /** False when it must not close by Escape or a tap outside (work in progress). */
  dismissible?: boolean
  /** A confirmation that interrupts (role alertdialog). */
  alert?: boolean
  testid?: string
}>(), {
  ariaLabel: '',
  ariaLabelledby: '',
  maxWidth: 'md:max-w-[480px]',
  sheet: true,
  padded: true,
  layer: 'overlay',
  dismissible: true,
  alert: false,
  testid: '',
})

const emit = defineEmits<{ close: [] }>()

const panel = ref<HTMLElement | null>(null)
let returnFocus: HTMLElement | null = null

function dismiss() {
  if (props.dismissible) emit('close')
}

useOverlayLayer(toRef(props, 'open'), dismiss)

watch(() => props.open, async (open) => {
  if (open) {
    returnFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
    await nextTick()
    if (panel.value && !panel.value.contains(document.activeElement)) panel.value.focus({ preventScroll: true })
  } else if (returnFocus?.isConnected) {
    returnFocus.focus({ preventScroll: true })
    returnFocus = null
  }
}, { immediate: true })
</script>

<style scoped>
.base-modal-panel.is-sheet {
  max-height: 90dvh;
  border-bottom: 0;
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
  padding-bottom: env(safe-area-inset-bottom);
}

.base-modal-panel.is-sheet.is-padded {
  padding: 1.5rem 1.25rem calc(1.5rem + env(safe-area-inset-bottom));
}

.base-modal-panel.is-centered {
  max-height: 90dvh;
  border-radius: var(--radius-md);
}

.base-modal-panel.is-centered.is-padded {
  padding: 1.5rem;
}

@media (min-width: 768px) {
  .base-modal-panel.is-sheet {
    max-height: 85dvh;
    border-bottom: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding-bottom: 0;
  }

  .base-modal-panel.is-sheet.is-padded {
    padding: 1.5rem;
  }
}

.base-modal-enter-active,
.base-modal-leave-active {
  transition: opacity 0.18s ease;
}

.base-modal-enter-active .base-modal-panel,
.base-modal-leave-active .base-modal-panel {
  transition: transform 0.22s ease;
}

.base-modal-enter-from,
.base-modal-leave-to {
  opacity: 0;
}

.base-modal-enter-from .base-modal-panel.is-sheet,
.base-modal-leave-to .base-modal-panel.is-sheet {
  transform: translateY(24px);
}

@media (prefers-reduced-motion: reduce) {
  .base-modal-enter-active,
  .base-modal-leave-active,
  .base-modal-enter-active .base-modal-panel,
  .base-modal-leave-active .base-modal-panel {
    transition: none;
  }
}
</style>
