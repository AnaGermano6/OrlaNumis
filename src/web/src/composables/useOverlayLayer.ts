import { computed, onBeforeUnmount, ref, watch, type Ref } from 'vue'

// The overlays open (dialogs, sheets, the chat, the image viewer), the last
// opened last. While one is open the page behind does not scroll, Escape
// closes the topmost one only, and the agent's button stays hidden.
const stack = ref<symbol[]>([])
// Only the root element is held: were the body held too, it would become
// the scrolling box, and bars that stick to the top would stop sticking.
let saved: string | null = null

/** Whether any overlay is open. */
export const overlayOpen = computed(() => stack.value.length > 0)

function lockScroll() {
  if (saved !== null || typeof document === 'undefined') return
  saved = document.documentElement.style.overflow
  document.documentElement.style.overflow = 'hidden'
}

function unlockScroll() {
  if (saved === null || typeof document === 'undefined') return
  document.documentElement.style.overflow = saved
  saved = null
}

/**
 * Registers an overlay while `open` is true. `onEscape`, when given, is
 * called on Escape if this overlay is the topmost one.
 */
export function useOverlayLayer(open: Ref<boolean>, onEscape?: () => void) {
  const id = Symbol('overlay')
  const isTop = () => stack.value[stack.value.length - 1] === id

  function add() {
    if (stack.value.includes(id)) return
    stack.value = [...stack.value, id]
    lockScroll()
  }

  function remove() {
    if (!stack.value.includes(id)) return
    stack.value = stack.value.filter((entry) => entry !== id)
    if (!stack.value.length) unlockScroll()
  }

  function onKeydown(event: KeyboardEvent) {
    if (event.key !== 'Escape' || !onEscape || !isTop()) return
    event.stopPropagation()
    onEscape()
  }

  watch(open, (value) => (value ? add() : remove()), { immediate: true })
  if (typeof document !== 'undefined') document.addEventListener('keydown', onKeydown)
  onBeforeUnmount(() => {
    remove()
    if (typeof document !== 'undefined') document.removeEventListener('keydown', onKeydown)
  })

  return { isTop }
}
