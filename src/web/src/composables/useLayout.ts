import { readonly, ref } from 'vue'

// A phone: a narrow screen, or a touch screen of little height (a phone
// turned sideways). The app is laid out for it the same way in the browser
// and installed (PWA); being installed only decides what is the app's own
// (the install prompt).
const MOBILE_QUERY = '(max-width: 767px), (pointer: coarse) and (max-height: 500px)'

const media = typeof window !== 'undefined' && typeof window.matchMedia === 'function'
  ? window.matchMedia(MOBILE_QUERY)
  : null
const isMobile = ref(media?.matches ?? false)
media?.addEventListener?.('change', (event) => {
  isMobile.value = event.matches
})

/** Whether the app is laid out for a phone; follows the screen as it turns. */
export function useLayout() {
  return { isMobile: readonly(isMobile) }
}
