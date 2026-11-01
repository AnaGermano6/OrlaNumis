import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import DesktopPageContext from '@/components/DesktopPageContext.vue'

let isMobile = false

vi.mock('@/composables/useLayout', async () => {
  const { ref } = await import('vue')
  return { useLayout: () => ({ isMobile: ref(isMobile) }) }
})

describe('DesktopPageContext', () => {
  beforeEach(() => {
    isMobile = false
    document.body.innerHTML = `
      <div id="desktop-page-title"></div>
      <div id="desktop-page-actions"></div>
    `
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('teleports the page title and actions into the desktop shell', () => {
    const wrapper = mount(DesktopPageContext, {
      props: { title: 'Wishlist' },
      slots: {
        default: '<h1>Local title</h1>',
        actions: '<button aria-label="Add to wishlist">Add</button>',
        desktop: '<p>Desktop only</p>',
      },
    })

    expect(document.querySelector('#desktop-page-title')?.textContent).toBe('Wishlist')
    expect(document.querySelector('#desktop-page-actions button')?.getAttribute('aria-label')).toBe('Add to wishlist')
    expect(wrapper.find('.desktop-page-context-local').classes()).not.toContain('is-mobile')

    wrapper.unmount()

    expect(document.querySelector('#desktop-page-title')?.textContent).toBe('')
    expect(document.querySelector('#desktop-page-actions')?.textContent).toBe('')
  })

  it('keeps the title in the page on a phone', () => {
    isMobile = true

    const wrapper = mount(DesktopPageContext, {
      props: { title: 'Wishlist' },
      slots: {
        default: '<h1>Local title</h1>',
        actions: '<button>Add</button>',
        desktop: '<p>Desktop only</p>',
      },
    })

    expect(document.querySelector('#desktop-page-title')?.textContent).toBe('')
    expect(document.querySelector('#desktop-page-actions')?.textContent).toBe('')
    expect(wrapper.find('.desktop-page-context-local').classes()).toContain('is-mobile')
    expect(wrapper.text()).toContain('Local title')
  })
})
