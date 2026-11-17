import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import HealthDiagnosis from '../HealthDiagnosis.vue'
import HealthPriorities from '../HealthPriorities.vue'
import { setLocale } from '@/i18n'
import type { CoinHealthItem, CollectionHealthSummary } from '@/types'

const mocks = vi.hoisted(() => ({ getCoinHealthList: vi.fn() }))
vi.mock('@/api/client', () => ({ getCoinHealthList: mocks.getCoinHealthList }))

const RouterLinkStub = defineComponent({
  props: { to: { type: String, required: true } },
  setup: (props, { slots, attrs }) => () => h('a', { ...attrs, href: props.to }, slots.default?.()),
})
const stubs = { RouterLink: RouterLinkStub, 'router-link': RouterLinkStub, AuthenticatedImage: true }

const SUMMARY: CollectionHealthSummary = {
  score: 61,
  grade: 'D',
  eligibleCoinCount: 470,
  weights: { metadata: 45, imageCoverage: 30, valuationFreshness: 25, aiCoverage: 0 },
  dimensions: { metadata: 75, imageCoverage: 25, valuationFreshness: 72, aiCoverage: 0 },
}

function item(coinId: number, title: string, missing: [string, 'high' | 'medium' | 'low', CoinHealthItem['quickActions'][number]][], imagePath?: string): CoinHealthItem {
  return {
    coinId, title, score: 50, grade: 'D', imagePath,
    dimensions: { metadata: 50, imageCoverage: 0, valuationFreshness: 50, aiCoverage: 0 },
    missingItems: missing.map(([key, severity, actionHint]) => ({ key, dimension: key.split('.')[0] as 'metadata', label: key, severity, actionHint })),
    quickActions: [],
  }
}

beforeEach(() => setLocale('pt-PT'))

describe('HealthDiagnosis', () => {
  it('shows the score out of 100 with its grade, each part that counts, and a way to each', async () => {
    const wrapper = mount(HealthDiagnosis, { props: { summary: SUMMARY, focused: null } })

    expect(wrapper.get('[data-testid="health-score"]').text()).toBe('61 / 100')
    expect(wrapper.get('[data-testid="health-gauge"]').text()).toContain('Nota D')
    // The AI analyses weigh nothing: not shown.
    expect(wrapper.findAll('li[data-testid^="health-part-"]').map((part) => part.findAll('span > span').map((cell) => cell.text()).filter(Boolean))).toEqual([
      ['Metadados', '75%'], ['Cobertura de imagens', '25%'], ['Atualidade das avaliações', '72%'],
    ])
    expect(wrapper.get('[data-testid="health-weights-hint"]').attributes('title')).toBe('A nota de 0 a 100 junta as partes, cada uma com o seu peso: Metadados 45%, Cobertura de imagens 30%, Atualidade das avaliações 25%.')

    await wrapper.get('[data-testid="health-focus-images"]').trigger('click')
    expect(wrapper.emitted('focus')).toEqual([['images']])
    await wrapper.setProps({ focused: 'images' })
    await wrapper.get('[data-testid="health-focus-images"]').trigger('click')
    expect(wrapper.emitted('focus')?.[1]).toEqual([null])
  })
})

describe('HealthPriorities', () => {
  beforeEach(() => {
    mocks.getCoinHealthList.mockReset()
    mocks.getCoinHealthList.mockResolvedValue({ data: {
      coins: [
        item(1, '2 euro 2026 – 70 Anos da Fundação Calouste Gulbenkian', [['images.any', 'high', 'upload_images'], ['metadata.mint', 'medium', 'edit_metadata']]),
        item(2, '5 euro – Albarnaz', [['metadata.grade', 'low', 'edit_metadata']], 'coins/albarnaz.jpg'),
        item(3, '10 euro – Amadeo de Souza-Cardoso', [['valuation.freshness', 'medium', 'run_valuation']]),
      ],
      pagination: { page: 1, limit: 10, total: 12 },
      facets: {
        countries: [{ value: 'Portugal', count: 9 }, { value: '', count: 2 }, { value: 'Espanha', count: 1 }],
        missing: [{ value: 'images.any', count: 7 }, { value: 'metadata.grade', count: 4 }],
      },
    } })
  })

  it('narrows to one country and to one thing missing, each choice with its count', async () => {
    const wrapper = mount(HealthPriorities, { props: { dimension: null }, global: { stubs } })
    await flushPromises()

    const countries = wrapper.get('[data-testid="health-priorities-country"]')
    // The countries by name (those without one are not a choice), with their flags.
    expect(countries.findAll('option').map((option) => option.text())).toEqual(['Todos os países', '🇪🇸 Espanha (1)', '🇵🇹 Portugal (9)'])
    const missing = wrapper.get('[data-testid="health-priorities-missing"]')
    expect(missing.findAll('option').map((option) => option.text())).toEqual(['Tudo o que falta', 'Sem fotografia (7)', 'Estado de conservação em falta (4)'])

    await countries.setValue('Portugal')
    await flushPromises()
    expect(mocks.getCoinHealthList).toHaveBeenLastCalledWith(expect.objectContaining({ scope: 'priority', country: 'Portugal', missing: undefined, page: 1 }))
    await missing.setValue('metadata.grade')
    await flushPromises()
    expect(mocks.getCoinHealthList).toHaveBeenLastCalledWith(expect.objectContaining({ country: 'Portugal', missing: 'metadata.grade', page: 1 }))
  })

  it('lists the coins to fix first, each with what it misses most and the way to fix it', async () => {
    const wrapper = mount(HealthPriorities, { props: { dimension: null }, global: { stubs } })
    await flushPromises()

    expect(mocks.getCoinHealthList).toHaveBeenCalledWith({ scope: 'priority', dimension: undefined, page: 1, limit: 10 })
    const rows = wrapper.findAll('li[data-testid^="health-priority-"]')
    expect(rows.map((row) => row.get('[data-testid="health-priority-issue"]').text())).toEqual(['Sem fotografia', 'Estado de conservação em falta', 'Avaliação desatualizada'])
    expect(rows.map((row) => row.get('[data-testid="health-priority-action"]').text())).toEqual(['Adicionar foto', 'Classificar', 'Atualizar'])
    expect(rows.map((row) => row.get('[data-testid="health-priority-action"]').attributes('href'))).toEqual(['/edit/1', '/edit/2', '/coin/3/actions'])
    expect(rows[0]!.text()).toContain('+1')
    expect(wrapper.get('[data-testid="health-priorities-more"]').text()).toBe('Mostrar mais (9)')
  })

  it('keeps to one part when asked, with a way back to all', async () => {
    const wrapper = mount(HealthPriorities, { props: { dimension: null }, global: { stubs } })
    await flushPromises()
    await wrapper.setProps({ dimension: 'images' })
    await flushPromises()

    expect(mocks.getCoinHealthList).toHaveBeenLastCalledWith({ scope: 'priority', dimension: 'images', page: 1, limit: 10 })
    const chip = wrapper.get('[data-testid="health-priorities-dimension"]')
    expect(chip.text()).toBe('Cobertura de imagens')
    await chip.trigger('click')
    expect(wrapper.emitted('clear-dimension')).toHaveLength(1)
  })

  it('reviews one coin at a time, with everything it misses', async () => {
    const wrapper = mount(HealthPriorities, { props: { dimension: null }, global: { stubs } })
    await flushPromises()

    await wrapper.get('[data-testid="health-review-toggle"]').trigger('click')
    const review = () => wrapper.get('[data-testid="health-review"]')
    expect(review().get('[data-testid="health-review-position"]').text()).toBe('1 de 12')
    expect(review().findAll('li').map((entry) => [entry.get('span').text(), entry.get('a').text()])).toEqual([['Sem fotografia', 'Adicionar foto'], ['Casa da moeda em falta', 'Completar']])
    await wrapper.get('[data-testid="health-review-next"]').trigger('click')
    await flushPromises()
    expect(review().get('[data-testid="health-review-position"]').text()).toBe('2 de 12')
    expect(review().text()).toContain('Albarnaz')
  })

  it('says when there is nothing to fix', async () => {
    mocks.getCoinHealthList.mockResolvedValue({ data: { coins: [], pagination: { page: 1, limit: 10, total: 0 } } })
    const wrapper = mount(HealthPriorities, { props: { dimension: null }, global: { stubs } })
    await flushPromises()
    expect(wrapper.get('[data-testid="health-priorities-empty"]').text()).toBe('Nada a corrigir: todas as moedas têm o que precisam.')
  })
})
