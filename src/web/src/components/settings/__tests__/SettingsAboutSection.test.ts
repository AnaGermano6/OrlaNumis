import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import SettingsAboutSection from '../SettingsAboutSection.vue'
import { setLocale } from '@/i18n'

describe('SettingsAboutSection', () => {
  beforeEach(() => setLocale('pt-PT'))

  it('credits the catalogue data, the official photos and the project the app is based on', () => {
    const wrapper = mount(SettingsAboutSection)

    expect(wrapper.get('h2').text()).toBe('Sobre')
    expect(wrapper.get('[data-testid="about-version"]').text()).toMatch(/^Versão /)
    expect(wrapper.text()).toContain('Dados do catálogo: Numista')
    expect(wrapper.text()).toContain('Fotos oficiais: BCE, INCM, Monnaie de Paris e Münze Österreich')
    const credits = wrapper.get('[data-testid="about-credits"]')
    expect(credits.text()).toBe('Baseada no projeto Aurearia, de Brian DeNicola (licença MIT)')
    expect(credits.get('a').attributes()).toMatchObject({ href: 'https://github.com/briandenicola/Aurearia', target: '_blank' })
    expect(wrapper.findAll('a').map((link) => link.attributes('href'))).toContain('https://en.numista.com/')
  })
})
