import { afterEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import BaseModal from '../BaseModal.vue'
import { overlayOpen } from '@/composables/useOverlayLayer'

function mountModal(props: Record<string, unknown> = {}) {
  return mount(BaseModal, {
    props: { open: true, ariaLabel: 'Export', ...props },
    slots: { default: '<p>Content</p>' },
    attachTo: document.body,
  })
}

function escape() {
  document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
}

describe('BaseModal', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('shows its content as a dialog and stops the page behind from scrolling', async () => {
    const wrapper = mountModal()
    await nextTick()

    const dialog = document.querySelector('[role="dialog"]')
    expect(dialog?.getAttribute('aria-modal')).toBe('true')
    expect(dialog?.getAttribute('aria-label')).toBe('Export')
    expect(dialog?.textContent).toContain('Content')
    expect(document.documentElement.style.overflow).toBe('hidden')
    expect(overlayOpen.value).toBe(true)

    await wrapper.setProps({ open: false })
    expect(document.documentElement.style.overflow).toBe('')
    expect(overlayOpen.value).toBe(false)
    wrapper.unmount()
  })

  it('closes on Escape and on a tap outside, the topmost dialog only', async () => {
    const below = mountModal()
    const above = mountModal({ layer: 'dialog' })
    await nextTick()

    escape()
    expect(above.emitted('close')).toHaveLength(1)
    expect(below.emitted('close')).toBeUndefined()

    above.unmount()
    escape()
    expect(below.emitted('close')).toHaveLength(1)

    ;(document.querySelector('.base-modal') as HTMLElement).click()
    expect(below.emitted('close')).toHaveLength(2)
    below.unmount()
    expect(overlayOpen.value).toBe(false)
  })

  it('stays open while it must not be dismissed', async () => {
    const wrapper = mountModal({ dismissible: false })
    await nextTick()

    escape()
    ;(document.querySelector('.base-modal') as HTMLElement).click()
    expect(wrapper.emitted('close')).toBeUndefined()
    wrapper.unmount()
  })

  it('rises from the bottom as a sheet, or stays centred', async () => {
    const sheet = mountModal()
    await nextTick()
    expect(document.querySelector('.base-modal-panel')?.classList.contains('is-sheet')).toBe(true)
    sheet.unmount()

    const centred = mountModal({ sheet: false })
    await nextTick()
    expect(document.querySelector('.base-modal-panel')?.classList.contains('is-centered')).toBe(true)
    centred.unmount()
  })
})
