import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import CountryFlag from '../CountryFlag.vue'

describe('CountryFlag', () => {
  it("shows a living country's emoji, a drawing for one that no longer exists, else nothing", () => {
    expect(mount(CountryFlag, { props: { country: 'Portugal' } }).text()).toBe('🇵🇹')

    const yugoslavia = mount(CountryFlag, { props: { country: 'Jugoslávia' } })
    const image = yugoslavia.get('[data-testid="country-flag-image"]')
    expect(image.attributes('src')).toMatch(/^data:image\/svg\+xml,/)
    expect(image.attributes('alt')).toBe('')
    expect(image.attributes('aria-hidden')).toBe('true')

    expect(mount(CountryFlag, { props: { country: 'Atlântida' } }).html()).toBe('<!--v-if-->')
  })
})
