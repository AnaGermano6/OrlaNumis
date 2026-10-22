import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import HelpSection from '../HelpSection.vue'
import en from '@/i18n/locales/en'
import { withEnglishCopy } from '@/test/i18nSource'
import { setLocale } from '@/i18n'

const srcDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..')

function readSource(relativePath: string): string {
  return fs.readFileSync(path.join(srcDir, relativePath), 'utf8')
}

vi.mock('vue-router', () => ({
  useRoute: () => ({ query: {} }),
}))

describe('HelpSection', () => {
  it('ships a Portuguese guide with the same structure as the English one', () => {
    const english = mount(HelpSection)
    setLocale('pt-PT')
    try {
      const portuguese = mount(HelpSection)
      expect(portuguese.text()).toContain('Guia de iniciação ao colecionismo de moedas')
      for (const selector of ['details', 'summary', 'h4', 'table', 'tr', 'li', 'pre', 'a', 'code']) {
        expect(portuguese.findAll(selector).length, selector).toBe(english.findAll(selector).length)
      }
    } finally {
      setLocale('en')
    }
  })

  it('documents current app help for stats and capture, without the removed features', () => {
    const wrapper = mount(HelpSection)
    const text = wrapper.text()

    expect(text).toContain('Stats Views')
    // The value details, investment breakdown and Notes pages, auction
    // tracking, the calendar and shipments were removed.
    // So were notifications, reminders, tags, Quick Access, the Emperor tracker,
    // Coin of the Day, API keys and the external tool server, OIDC and passkeys,
    // and the AI set proposals.
    for (const removed of ['Value Details', 'Investment Breakdown', 'Save to Notes', 'Calendar', 'Shipment', 'auction alerts', 'notification', 'Purchase reminders', 'Custom tags', 'Quick Access', 'Emperor Tracker', 'Coin of the Day', 'X-API-Key', 'External Tool Server', 'OIDC', 'passkey', 'Museum Tray', 'Agentic', 'proposal']) {
      expect(text).not.toContain(removed)
    }
    expect(text).toContain('Quick Capture, Coin Lookup, and image upload flows start the camera only after you tap')
    expect(text).toContain('Catalogue, challenge and goal sets show how far along you are')
  })

  it('explains Coin Copilot and keeps the Coin Agent for when Copilot is off', () => {
    const wrapper = mount(HelpSection)
    const copilot = wrapper.get('[data-testid="help-coin-copilot"]')
    const text = wrapper.text()
    const progress = withEnglishCopy(readSource('components/chat/CopilotRunProgress.vue'))

    expect(copilot.text()).toBe('Coin Copilot (Beta)')
    expect(progress).toContain('Beta')
    expect(progress).toContain('Coin Copilot')
    expect(text).toContain('Beta · Coin Copilot')
    expect(text).toContain('step checklist')
    expect(text).toContain('Source evidence')
    expect(text).toContain('clarifying question')
    expect(text).toContain('5 listings by default')
    expect(text).toContain('up to 10')
    expect(text).toContain('Coin show search is not available in Copilot yet')
    expect(text).toContain('How Coin Copilot Works')
    expect(text).toContain('When Coin Copilot is off, open the chat')
  })

  // The help section's value is entirely in its instructions being followable. These pin the
  // navigation it names to the navigation that exists, so a settings-tab rename or a moved
  // entry point fails here instead of quietly sending users to the wrong screen.
  it('sends users to the settings tab that actually holds import and export', () => {
    const text = mount(HelpSection).text()
    const settingsPage = readSource('pages/SettingsPage.vue')
    const backups = readSource('components/settings/SettingsBackupsSection.vue')

    expect(settingsPage).toContain("backups: 'settings.tabs.backups'")
    expect(en.settings.tabs.backups).toBe('Backups')
    expect(backups).toContain("t('settings.backups.csvTemplate')")
    expect(en.settings.backups.csvTemplate).toBe('CSV Template')
    expect(text).toContain('Settings → Backups')
    expect(text).toContain('Import the CSV from Settings → Backups → Import')
    expect(text).not.toContain('Settings → Data → Import')
  })

  it('has no Connections tab: its integrations were removed', () => {
    const text = mount(HelpSection).text()
    const settingsPage = readSource('pages/SettingsPage.vue')

    expect(settingsPage).not.toContain("'connections'")
    expect(settingsPage).not.toContain('SettingsConnectionsSection')
    expect('connections' in en.settings.tabs).toBe(false)
    expect(text).not.toContain('Settings → Connections')
    expect(text).not.toContain('NumisBids')
    expect(text).not.toContain('ParcelApp')
    expect(text).not.toContain('Pushover')
  })

  it('names entry points that exist in the sidebar', () => {
    const text = mount(HelpSection).text()
    const app = readSource('App.vue')

    // The agent is a sidebar item, not the retired Wish List → Find Coins button.
    expect(app).toContain("id: 'agent'")
    expect(text).toContain('open the chat from Agent in the sidebar')
    expect(text).not.toContain('Find Coins')

    // Quick Capture and Coin Lookup are one merged entry point (see AppNavigation.test.ts).
    expect(app).not.toContain("to: '/quick-capture'")
    expect(text).toContain('Both intake flows live under Identify Coin')

    for (const view of ['Time Machine', 'Analysis History']) {
      expect(Object.values(en.nav)).toContain(view)
      expect(text).toContain(view)
    }
    // Showcases, followers and storage locations were removed.
    for (const removed of ['Showcases', 'Followers', 'Storage locations']) expect(text).not.toContain(removed)
  })

  it('lists every settings tab in its orientation table', () => {
    const text = mount(HelpSection).text()
    const settingsPage = readSource('pages/SettingsPage.vue')

    const tabKeys = [...settingsPage.matchAll(/^ {2}([a-z]+): 'settings\.tabs\.[a-z]+',$/gm)]
      .map(match => match[1] as keyof typeof en.settings.tabs)
    const documented = new Set(tabKeys.map(key => en.settings.tabs[key]))

    expect(documented.size).toBeGreaterThan(5)
    for (const label of documented) {
      expect(text).toContain(label)
    }
  })

  it('keeps the CSV documentation in step with the importer and its template', () => {
    const text = mount(HelpSection).text()
    const backups = readSource('components/settings/SettingsBackupsSection.vue')

    for (const column of ['purchaseLocation', 'vendorSku', 'vendorInvoice', 'rarityRating']) {
      expect(backups).toContain(column)
      expect(text).toContain(column)
    }
    // Coins are not marked as sold or private: the importer reads no such columns, and the help lists none.
    for (const column of ['isSold', 'soldPrice', 'soldDate', 'soldTo', 'isPrivate']) {
      expect(backups).not.toContain(`['${column}'`)
      expect(text).not.toContain(column)
    }

    // era is an admin-configured enum; free-text reign dates there do not match era filters,
    // so neither the template nor the help example may demonstrate that.
    expect(backups).not.toContain("'27 BC - 14 AD', 'Rome'")
    expect(text).not.toContain('Augustus,27 BC - 14 AD')
  })
})
