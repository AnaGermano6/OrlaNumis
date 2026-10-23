import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import CollectionHealthScorecard from '../CollectionHealthScorecard.vue'
import { setLocale } from '@/i18n'
import type { CollectionHealthSummary } from '@/types'

function summary(aiWeight: number): CollectionHealthSummary {
  return {
    score: 72,
    grade: 'C',
    eligibleCoinCount: 3,
    weights: aiWeight
      ? { metadata: 40, imageCoverage: 20, valuationFreshness: 20, aiCoverage: 20 }
      : { metadata: 50, imageCoverage: 25, valuationFreshness: 25, aiCoverage: 0 },
    dimensions: { metadata: 89, imageCoverage: 25, valuationFreshness: 85, aiCoverage: 0 },
  } as CollectionHealthSummary
}

describe('CollectionHealthScorecard', () => {
  beforeEach(() => setLocale('pt-PT'))
  afterEach(() => setLocale('en'))

  it('leaves out the AI part while it does not count', () => {
    const wrapper = mount(CollectionHealthScorecard, { props: { summary: summary(0) } })
    expect(wrapper.text()).not.toContain('Cobertura de IA')
    expect(wrapper.text()).toContain('Metadados: 50%')
    expect(wrapper.text()).toContain('Cobertura de imagens: 25%')
  })

  it('shows it once an AI provider makes it count', () => {
    const wrapper = mount(CollectionHealthScorecard, { props: { summary: summary(20) } })
    expect(wrapper.text()).toContain('Cobertura de IA: 20%')
  })
})
