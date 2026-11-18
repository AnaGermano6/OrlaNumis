import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import StatsConservation from '../StatsConservation.vue'
import { setLocale } from '@/i18n'

describe('StatsConservation', () => {
  beforeEach(() => setLocale('pt-PT'))

  it('shows every grade, best first, with its pieces, share and meaning, then those without a grade', () => {
    const wrapper = mount(StatsConservation, { props: { id: 'all', total: 300, byGrade: [
      { grade: 'UNC', count: 80 }, { grade: 'G', count: 60 }, { grade: 'VF', count: 20 }, { grade: 'XF', count: 14 }, { grade: 'PRF', count: 6 }, { grade: 'Bom', count: 1 },
    ] } })

    const grades = wrapper.findAll('li').map((item) => item.attributes('data-testid'))
    expect(grades).toEqual(['PRF', 'UNC', 'XF', 'VF', 'G', 'Bom'].map((grade) => `conservation-${grade}`))
    const lines = (grade: string) => wrapper.get(`[data-testid="conservation-${grade}"]`).findAll(':scope > span').map((line) => line.text().replace(/\s+/g, ' '))
    expect(lines('UNC')).toEqual(['UNC', '80 peças · 44%', 'Sem circulação (Soberba, FDC)'])
    // A grade written otherwise has no meaning to show.
    expect(lines('Bom')).toEqual(['Bom', '1 peça · 0,6%'])
    expect(wrapper.get('[data-testid="conservation-ungraded"]').text().replace(/\s+/g, ' ')).toBe('119 peças sem estado registado — Classificar agora')
    // Its title names it, with an id of its own on a page with several.
    expect(wrapper.get('section').attributes('aria-labelledby')).toBe('all-conservation-title')
    expect(wrapper.get('h2').attributes('id')).toBe('all-conservation-title')
  })

  it('asks for the pieces without a grade, to grade them', async () => {
    const wrapper = mount(StatsConservation, { props: { id: 'coins', total: 10, byGrade: [{ grade: 'UNC', count: 4 }] } })

    await wrapper.get('[data-testid="conservation-grade-now"]').trigger('click')
    expect(wrapper.emitted('grade-now')).toHaveLength(1)
  })

  it('stays out of the way when no piece has a grade', () => {
    const wrapper = mount(StatsConservation, { props: { id: 'banknotes', total: 5, byGrade: [] } })
    expect(wrapper.find('[data-testid="stats-conservation"]').exists()).toBe(false)
  })
})
