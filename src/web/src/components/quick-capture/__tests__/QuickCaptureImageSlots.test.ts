import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { withEnglishCopy } from '@/test/i18nSource'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const source = withEnglishCopy(fs.readFileSync(path.resolve(__dirname, '../QuickCaptureImageSlots.vue'), 'utf8'))

describe('QuickCaptureImageSlots', () => {
  it('offers obverse, reverse, detail, camera, and file upload inputs', () => {
    expect(source).toContain('Obverse')
    expect(source).toContain('Reverse')
    expect(source).toContain('Detail photos')
    expect(source).toContain('capture="environment"')
    expect(source).toContain('accept="image/*"')
  })
})
