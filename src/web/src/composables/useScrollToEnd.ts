import { nextTick, onBeforeUnmount, watch, type Ref } from 'vue'

/**
 * Keeps a strip that scrolls sideways at its end (the latest years in view)
 * as it is laid out and resized, until the user scrolls away from the end;
 * when `trigger` changes (another country), back to the end.
 */
export function useScrollToEnd(target: Ref<HTMLElement | null>, trigger: () => unknown) {
  let pinned = true
  let observer: ResizeObserver | undefined

  function toEnd() {
    const element = target.value
    if (element && pinned) element.scrollLeft = element.scrollWidth
  }

  function onScroll() {
    const element = target.value
    if (element) pinned = element.scrollLeft + element.clientWidth >= element.scrollWidth - 2
  }

  watch(target, (element, previous) => {
    previous?.removeEventListener('scroll', onScroll)
    observer?.disconnect()
    if (!element) return
    element.addEventListener('scroll', onScroll, { passive: true })
    if (typeof ResizeObserver !== 'undefined') {
      observer = new ResizeObserver(toEnd)
      observer.observe(element)
      if (element.firstElementChild) observer.observe(element.firstElementChild)
    }
    toEnd()
  }, { flush: 'post' })

  watch(trigger, async () => {
    pinned = true
    await nextTick()
    toEnd()
  }, { flush: 'post' })

  onBeforeUnmount(() => {
    observer?.disconnect()
    target.value?.removeEventListener('scroll', onScroll)
  })
}
