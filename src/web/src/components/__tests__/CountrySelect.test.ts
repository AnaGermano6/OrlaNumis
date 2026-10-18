import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import CountrySelect from '../CountrySelect.vue'

const OPTIONS = [
  { value: '', label: 'Todos os países' },
  { value: 'Alemanha', label: 'Alemanha (23)', country: 'Alemanha' },
  { value: 'Jugoslávia', label: 'Jugoslávia (1)', country: 'Jugoslávia' },
  { value: 'Portugal', label: 'Portugal (120)', country: 'Portugal' },
]

const list = () => document.querySelector<HTMLElement>('[data-testid="country-select-list"]')
const options = () => [...document.querySelectorAll<HTMLElement>('[data-testid="country-select-option"]')]

function render(modelValue = '') {
  return mount(CountrySelect, { props: { modelValue, options: OPTIONS, label: 'País' }, attachTo: document.body })
}

describe('CountrySelect', () => {
  afterEach(() => { document.body.innerHTML = '' })

  it('shows the chosen country with its flag, and works as a combobox for screen readers', async () => {
    const wrapper = render('Portugal')
    const button = wrapper.get('button')
    expect(button.text()).toBe('🇵🇹Portugal (120)')
    expect(button.attributes()).toMatchObject({ role: 'combobox', 'aria-haspopup': 'listbox', 'aria-expanded': 'false', 'aria-label': 'País' })

    await button.trigger('click')
    expect(button.attributes('aria-expanded')).toBe('true')
    expect(list()!.getAttribute('role')).toBe('listbox')
    expect(options().map((option) => option.getAttribute('aria-selected'))).toEqual(['false', 'false', 'false', 'true'])
    // The chosen one is active at first.
    expect(button.attributes('aria-activedescendant')).toBe(options()[3]!.id)
    wrapper.unmount()
  })

  it('moves with the arrows, Home and End, jumps to a typed letter, chooses with Enter and closes with Escape', async () => {
    const wrapper = render()
    const button = wrapper.get('button')

    await button.trigger('keydown', { key: 'ArrowDown' })
    expect(list()).not.toBeNull()
    await button.trigger('keydown', { key: 'ArrowDown' })
    expect(button.attributes('aria-activedescendant')).toBe(options()[1]!.id)
    await button.trigger('keydown', { key: 'End' })
    expect(button.attributes('aria-activedescendant')).toBe(options()[3]!.id)
    await button.trigger('keydown', { key: 'Home' })
    await button.trigger('keydown', { key: 'j' })
    expect(button.attributes('aria-activedescendant')).toBe(options()[2]!.id)
    await button.trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('update:modelValue')).toEqual([['Jugoslávia']])
    expect(list()).toBeNull()

    await button.trigger('click')
    await button.trigger('keydown', { key: 'Escape' })
    expect(list()).toBeNull()
    expect(wrapper.emitted('update:modelValue')).toHaveLength(1)
    wrapper.unmount()
  })

  it('closes on a click elsewhere, and shows drawn flags for countries that no longer exist', async () => {
    const wrapper = render()
    await wrapper.get('button').trigger('click')
    expect(options()[2]!.querySelector('img[data-testid="country-flag-image"]')).not.toBeNull()
    expect(options()[0]!.querySelector('img, span[aria-hidden]')).toBeNull()

    document.body.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(list()).toBeNull()
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    wrapper.unmount()
  })
})
