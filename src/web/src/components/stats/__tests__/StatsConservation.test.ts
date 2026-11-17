import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import StatsConservation from '../StatsConservation.vue'
import { setLocale } from '@/i18n'
import { useCoinsStore } from '@/stores/coins'

const push = vi.hoisted(() => vi.fn())
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))

describe('StatsConservation', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    setActivePinia(createPinia())
  })

  it('shows every grade, best first, with its pieces, share and meaning, then those without a grade', () => {
    const wrapper = mount(StatsConservation, { props: { total: 300, byGrade: [
      { grade: 'UNC', count: 80 }, { grade: 'G', count: 60 }, { grade: 'VF', count: 20 }, { grade: 'XF', count: 14 }, { grade: 'PRF', count: 6 }, { grade: 'Bom', count: 1 },
    ] } })

    const grades = wrapper.findAll('li').map((item) => item.attributes('data-testid'))
    expect(grades).toEqual(['PRF', 'UNC', 'XF', 'VF', 'G', 'Bom'].map((grade) => `conservation-${grade}`))
    const lines = (grade: string) => wrapper.get(`[data-testid="conservation-${grade}"]`).findAll(':scope > span').map((line) => line.text().replace(/\s+/g, ' '))
    expect(lines('UNC')).toEqual(['UNC', '80 peças · 44%', 'Sem circulação (Soberba, FDC)'])
    // A grade written otherwise has no meaning to show.
    expect(lines('Bom')).toEqual(['Bom', '1 peça · 0,6%'])
    expect(wrapper.get('[data-testid="conservation-ungraded"]').text().replace(/\s+/g, ' ')).toBe('119 peças sem estado registado — Classificar agora')
  })

  it('opens the gallery with only the coins without a grade, to grade them', async () => {
    const store = useCoinsStore()
    store.selectedCountry = 'Portugal'
    const wrapper = mount(StatsConservation, { props: { total: 10, byGrade: [{ grade: 'UNC', count: 4 }] } })

    await wrapper.get('[data-testid="conservation-grade-now"]').trigger('click')
    expect(store.selectedUngraded).toBe(true)
    expect(store.selectedCountry).toBe('')
    expect(push).toHaveBeenCalledWith('/')
  })

  it('stays out of the way when no piece has a grade', () => {
    const wrapper = mount(StatsConservation, { props: { total: 5, byGrade: [] } })
    expect(wrapper.find('[data-testid="stats-conservation"]').exists()).toBe(false)
  })
})
