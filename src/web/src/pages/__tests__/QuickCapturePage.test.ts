import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { withEnglishCopy } from '@/test/i18nSource'
import en from '@/i18n/locales/en'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = withEnglishCopy(fs.readFileSync(path.resolve(__dirname, '../QuickCapturePage.vue'), 'utf8'))

describe('QuickCapturePage', () => {
  it('composes the MVP save workflow and count-exclusion message', () => {
    expect(source).toContain('QuickCaptureForm')
    expect(source).toContain('Draft saved')
    expect(source).toContain("t('quickCapture.draftSavedDetail', { id: lastDraft.id })")
    expect(en.quickCapture.draftSavedDetail).toContain('excluded from collection, wishlist, stats, and health counts')
  })

  it('keeps Quick Capture v1 manual and navigates to drafts without AI intake expansion', () => {
    expect(source).toContain('to="/quick-capture/drafts"')
    expect(source).toContain('aria-label="All captures"')
    expect(source).toContain('List')
    expect(source).toContain('Capture sparse coin details quickly')
    expect(source).not.toContain('createIntakeDraft')
    expect(source).not.toContain('lookupCoin')
    expect(source).not.toContain('agent')
  })
})
