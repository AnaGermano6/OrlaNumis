import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import CategoryEraConfirmModal from '../chat/CategoryEraConfirmModal.vue'

const css = fs.readFileSync(path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..', 'assets', 'styles', 'main.css'), 'utf8')
/** A layer of the scale in main.css (--z-chat…). */
const layer = (name: string) => Number(css.match(new RegExp(`${name}:\\s*(\\d+);`))?.[1])


describe('CategoryEraConfirmModal', () => {
  it('renders nothing when request is null', () => {
    const wrapper = mount(CategoryEraConfirmModal, {
      props: { request: null },
      global: { stubs: { Teleport: true } },
    })
    expect(wrapper.find('[class*="fixed"]').exists()).toBe(false)
  })

  it('overlay z-index exceeds CoinSearchChat sidebar z-index so the dialog is selectable', () => {
    // The chat sits on the chat layer; the confirmation must open above it.
    // jsdom cannot measure stacking, so the layer and the scale are asserted.
    const wrapper = mount(CategoryEraConfirmModal, {
      props: {
        request: {
          fieldLabel: 'Era',
          suggestedValue: 'AD190',
          options: ['Roman Imperial', 'Roman Republic'],
        },
      },
      global: { stubs: { Teleport: true } },
    })

    // It opens on the dialog layer, which the scale puts above the chat's.
    const overlay = wrapper.find('.base-modal')
    expect(overlay.exists()).toBe(true)
    expect(overlay.classes()).toContain('z-(--z-dialog)')
    expect(layer('--z-dialog')).toBeGreaterThan(layer('--z-chat'))
  })

  it('emits choose with the selected option', async () => {
    const wrapper = mount(CategoryEraConfirmModal, {
      props: {
        request: {
          fieldLabel: 'Era',
          suggestedValue: 'AD190',
          options: ['Roman Imperial', 'Roman Republic'],
        },
      },
      global: { stubs: { Teleport: true } },
    })

    const buttons = wrapper.findAll('button')
    const romanImperial = buttons.find((b) => b.text() === 'Roman Imperial')
    expect(romanImperial).toBeTruthy()
    await romanImperial!.trigger('click')
    expect(wrapper.emitted('choose')).toEqual([['Roman Imperial']])
  })

  it('emits cancel when clicking the Cancel button', async () => {
    const wrapper = mount(CategoryEraConfirmModal, {
      props: {
        request: {
          fieldLabel: 'Era',
          suggestedValue: 'AD190',
          options: ['Roman Imperial'],
        },
      },
      global: { stubs: { Teleport: true } },
    })

    const cancelBtn = wrapper.findAll('button').find((b) => b.text() === 'Cancel')
    expect(cancelBtn).toBeTruthy()
    await cancelBtn!.trigger('click')
    expect(wrapper.emitted('cancel')).toBeTruthy()
  })
})
