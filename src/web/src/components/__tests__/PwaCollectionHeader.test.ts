import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const headerPath = path.resolve(__dirname, '../collection/PwaCollectionHeader.vue')

describe('PwaCollectionHeader', () => {
  it('keeps select mode in the menu instead of the top row', () => {
    const source = fs.readFileSync(headerPath, 'utf8')

    expect(source).not.toContain('class="pwa-icon-btn"')
    expect(source).toContain(`<span class="section-label mb-0">{{ t('collection.menu.selection') }}</span>`)
    expect(source).toContain("{{ selectMode ? t('collection.menu.exitSelection') : t('collection.menu.enableSelection') }}")
  })

  it('does not launch Mint Map from the PWA collection menu', () => {
    const source = fs.readFileSync(headerPath, 'utf8')

    expect(source).not.toContain('to="/mint-map"')
    expect(source).not.toContain('Mint Map')
    expect(source).not.toContain('MapPin')
  })
})
