import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import DesktopPageContext from '@/components/DesktopPageContext.vue'

let isMobile = false

vi.mock('@/composables/useLayout', async () => {
  const { ref } = await import('vue')
  return { useLayout: () => ({ isMobile: ref(isMobile) }) }
})

const SLOTS = {
  default: '<h1>Local title</h1>',
  actions: '<button aria-label="Add to wishlist">Add</button>',
  desktop: '<p>Desktop only</p>',
}

describe('DesktopPageContext', () => {
  it('shows the page title and actions at the top of the page on a wider screen', () => {
    isMobile = false
    const wrapper = mount(DesktopPageContext, { props: { title: 'Wishlist' }, slots: SLOTS })

    const header = wrapper.get('[data-testid="page-context"]')
    expect(header.get('h1').text()).toBe('Wishlist')
    expect(header.get('button').attributes('aria-label')).toBe('Add to wishlist')
    expect(wrapper.text()).toContain('Desktop only')
    // The page's own header is the phone's.
    expect(wrapper.find('.desktop-page-context-local').classes()).not.toContain('is-mobile')
  })

  it('keeps the page its own header on a phone', () => {
    isMobile = true
    const wrapper = mount(DesktopPageContext, { props: { title: 'Wishlist' }, slots: SLOTS })

    expect(wrapper.find('[data-testid="page-context"]').exists()).toBe(false)
    expect(wrapper.find('.desktop-page-context-local').classes()).toContain('is-mobile')
    expect(wrapper.text()).toContain('Local title')
  })
})
