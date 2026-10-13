import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import QuantityInput from '../QuantityInput.vue'

function render(modelValue: number | null | undefined) {
  return mount(QuantityInput, { props: { modelValue, id: 'quantity' } })
}

describe('QuantityInput', () => {
  it('starts at one and steps up and down, never below one', async () => {
    const wrapper = render(undefined)
    const input = wrapper.get<HTMLInputElement>('[data-testid="quantity-input"]')
    expect(input.element.value).toBe('1')
    expect(input.attributes('id')).toBe('quantity')
    expect(wrapper.get('[data-testid="quantity-less"]').attributes('disabled')).toBeDefined()

    await wrapper.get('[data-testid="quantity-more"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([[2]])

    await wrapper.setProps({ modelValue: 3 })
    await wrapper.get('[data-testid="quantity-less"]').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([2])
  })

  it('takes a typed whole number and puts back anything else', async () => {
    const wrapper = render(2)
    const input = wrapper.get<HTMLInputElement>('[data-testid="quantity-input"]')

    await input.setValue('12')
    expect(wrapper.emitted('update:modelValue')).toEqual([[12]])

    await input.setValue('0')
    await input.trigger('change')
    expect(wrapper.emitted('update:modelValue')).toEqual([[12]])
    expect(input.element.value).toBe('2')

    await input.setValue('1.5')
    await input.trigger('change')
    expect(wrapper.emitted('update:modelValue')).toEqual([[12]])
    expect(input.element.value).toBe('2')
  })
})
