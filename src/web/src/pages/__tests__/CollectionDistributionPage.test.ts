import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const pagePath = path.resolve(__dirname, '../CollectionDistributionPage.vue')

describe('CollectionDistributionPage', () => {
  it('owns distribution content without health and value trends sections', () => {
    const source = fs.readFileSync(pagePath, 'utf8')

    expect(source).toContain('StatsHeatMap')
    expect(source).toContain('StatsBarChart')
    expect(source).not.toContain('StatsValueOverTime')
    expect(source).not.toContain('id="collection-health"')
    expect(source).not.toContain('id="value-over-time"')
    expect(source).not.toContain('to="/mint-map"')
  })

  it('shows amounts in the currency every page uses, never a written dollar sign', () => {
    const timeline = fs.readFileSync(path.resolve(__dirname, '../TimelinePage.vue'), 'utf8')
    expect(timeline).not.toContain('${{')
    expect(timeline).toContain('formatCurrency(coin.purchasePrice)')
    expect(fs.readFileSync(pagePath, 'utf8')).toContain('priceRangeLabel(p.range)')
  })
})
