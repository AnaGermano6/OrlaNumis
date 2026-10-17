import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const addCoinPagePath = path.resolve(__dirname, '../AddCoinPage.vue')

describe('AddCoinPage', () => {
  it('builds the create and intake payloads from the same form helper, without an imperial figure', () => {
    const source = fs.readFileSync(addCoinPagePath, 'utf8')

    expect(source).not.toContain('romanImperialFigureId')
    expect(source).toContain('overrides: buildCoinPayload(reviewForm)')
    expect(source).toContain('const payload = buildCoinPayload(form)')
  })
})
