import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import StatsRulers from '../StatsRulers.vue'
import { setLocale } from '@/i18n'
import { useCoinsStore } from '@/stores/coins'

const push = vi.hoisted(() => vi.fn())
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))

const RULERS = [
  { ruler: 'Terceira República', country: 'Portugal', count: 119 },
  { ruler: 'República', country: 'França', count: 55 },
  { ruler: 'Isabel II', country: 'Reino Unido', count: 47 },
  { ruler: 'Segunda República', country: 'Portugal', count: 46 },
  { ruler: 'Segunda República', country: 'Espanha', count: 35 },
  { ruler: 'República Federal', country: 'Alemanha', count: 29 },
  { ruler: 'Quinta República', country: 'França', count: 26 },
  { ruler: 'Guilherme Alexandre', country: 'Países Baixos', count: 12 },
  { ruler: 'João Carlos I', country: 'Espanha', count: 11 },
  { ruler: 'República Federativa do Brasil', country: 'Brasil', count: 6 },
  { ruler: 'Juliana', country: 'Países Baixos', count: 3 },
]

function names(wrapper: ReturnType<typeof mount>) {
  return wrapper.findAll('li').map((item) => item.get('.truncate').text())
}

describe('StatsRulers', () => {
  beforeEach(() => {
    setLocale('pt-PT')
    setActivePinia(createPinia())
    push.mockReset()
  })

  it('lists the first ten rulers, a name two countries share named with each, with their coins and share', () => {
    const wrapper = mount(StatsRulers, { props: { rulers: RULERS, total: 476 } })

    expect(names(wrapper)).toEqual([
      'Terceira República', 'República', 'Isabel II', 'Segunda República (PT)', 'Segunda República (ES)',
      'República Federal', 'Quinta República', 'Guilherme Alexandre', 'João Carlos I', 'República Federativa do Brasil',
    ])
    const first = wrapper.get('li')
    expect(first.text()).toContain('119')
    expect(first.text()).toContain('25%')
    // The others a click away.
    const toggle = wrapper.get('[data-testid="rulers-toggle"]')
    expect(toggle.text()).toBe('Ver todos os 11 governantes')
  })

  it('shows every ruler, the monarchs or the regimes', async () => {
    const wrapper = mount(StatsRulers, { props: { rulers: RULERS, total: 476 } })

    await wrapper.get('[data-testid="rulers-kind-monarchs"]').trigger('click')
    expect(names(wrapper)).toEqual(['Isabel II', 'Guilherme Alexandre', 'João Carlos I', 'Juliana'])
    expect(wrapper.get('[data-testid="rulers-kind-monarchs"]').attributes('aria-checked')).toBe('true')

    await wrapper.get('[data-testid="rulers-kind-regimes"]').trigger('click')
    expect(names(wrapper)).toEqual(['Terceira República', 'República', 'Segunda República (PT)', 'Segunda República (ES)', 'República Federal', 'Quinta República', 'República Federativa do Brasil'])
    expect(wrapper.find('[data-testid="rulers-toggle"]').exists()).toBe(false)
  })

  it('opens the gallery with the coins of a ruler of its country', async () => {
    const store = useCoinsStore()
    const wrapper = mount(StatsRulers, { props: { rulers: RULERS, total: 476 } })

    await wrapper.findAll('li')[4]!.get('button').trigger('click')
    expect(store.searchQuery).toBe('Segunda República')
    expect(store.selectedCountry).toBe('Espanha')
    expect(push).toHaveBeenCalledWith('/')
  })

  it('says so when there is no ruler of a kind', async () => {
    const wrapper = mount(StatsRulers, { props: { rulers: [{ ruler: 'Terceira República', country: 'Portugal', count: 3 }], total: 3 } })
    await wrapper.get('[data-testid="rulers-kind-monarchs"]').trigger('click')
    expect(wrapper.get('[data-testid="rulers-empty"]').text()).toBe('Nenhum governante deste tipo.')
  })
})
