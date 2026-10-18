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
    expect(wrapper.text()).toContain(`© ${new Date().getFullYear()} Orla Numis`)
  })

  it('says the project it is based on, with a link to it', () => {
    setLocale('pt-PT')
    const credits = mount(AppFooter).get('[data-testid="app-credits"]')

    expect(credits.text().replace(/\s+/g, ' ')).toBe(`© ${new Date().getFullYear()} Orla Numis · Baseada no projeto Aurearia, de Brian DeNicola (licença MIT)`)
    expect(credits.get('a').attributes()).toMatchObject({ href: 'https://github.com/briandenicola/Aurearia', target: '_blank', rel: 'noopener noreferrer' })
  })
})
