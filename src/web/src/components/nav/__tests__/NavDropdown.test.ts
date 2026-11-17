import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { reactive } from 'vue'
import NavDropdown from '../NavDropdown.vue'

const route = reactive({ fullPath: '/' })
vi.mock('vue-router', () => ({ useRoute: () => route }))

function mountDropdown(hover: boolean) {
  return mount(NavDropdown, {
    attachTo: document.body,
    props: { label: 'Sets', hover },
    slots: { button: 'Sets', default: '<a href="/sets">My sets</a>' },
  })
}

describe('NavDropdown', () => {
  it('opens as the pointer passes, and a click then keeps it open', async () => {
    const wrapper = mountDropdown(true)
    await wrapper.trigger('mouseenter')
    expect(wrapper.find('a').exists()).toBe(true)
    await wrapper.get('button').trigger('click')
    expect(wrapper.find('a').exists()).toBe(true)
    await wrapper.trigger('mouseleave')
    expect(wrapper.find('a').exists()).toBe(false)
    wrapper.unmount()
  })

  it('opens and closes on a click, and closes on Escape, a click elsewhere or another page', async () => {
    const wrapper = mountDropdown(false)
    const button = wrapper.get('button')
    await button.trigger('click')
    expect(button.attributes('aria-expanded')).toBe('true')
    await button.trigger('click')
    expect(button.attributes('aria-expanded')).toBe('false')

    await button.trigger('click')
    await wrapper.trigger('keydown', { key: 'Escape' })
    expect(wrapper.find('a').exists()).toBe(false)

    await button.trigger('click')
    document.body.click()
    await wrapper.vm.$nextTick()
    expect(wrapper.find('a').exists()).toBe(false)

    await button.trigger('click')
    route.fullPath = '/sets'
    await wrapper.vm.$nextTick()
    expect(wrapper.find('a').exists()).toBe(false)
    wrapper.unmount()
  })
})
