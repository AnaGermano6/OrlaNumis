import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import AppFooter from '../AppFooter.vue'
import { setLocale } from '@/i18n'

describe('AppFooter', () => {
  afterEach(() => setLocale('en'))

  it('names the app, its version, where its data comes from and the year', () => {
    setLocale('pt-PT')
    const wrapper = mount(AppFooter)

    expect(wrapper.get('footer').attributes('aria-label')).toBe('Rodapé')
    expect(wrapper.get('[data-testid="app-version"]').text()).toMatch(/^Versão \S+$/)
    expect(wrapper.text()).toContain('Dados do catálogo: Numista')
    expect(wrapper.text()).toContain('Fotos oficiais: BCE, INCM, Monnaie de Paris e Münze Österreich')
    expect(wrapper.text()).toContain(`© ${new Date().getFullYear()} Aurearia`)
  })
})
