import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import AdminNumistaFillSection from '@/components/admin/AdminNumistaFillSection.vue'
import type { CoinNumistaFillResult } from '@/api/client'
import { setLocale } from '@/i18n'

const mocks = vi.hoisted(() => ({
  previewNumistaFill: vi.fn(),
  applyNumistaFill: vi.fn(),
  showConfirm: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  previewNumistaFill: () => mocks.previewNumistaFill(),
  applyNumistaFill: (corrections: boolean) => mocks.applyNumistaFill(corrections),
}))

vi.mock('@/composables/useDialog', () => ({
  useDialog: () => ({ showConfirm: mocks.showConfirm }),
}))

function result(overrides: Partial<CoinNumistaFillResult> = {}): CoinNumistaFillResult {
  return {
    applied: false,
    coins: 2,
    materials: 2,
    weights: 1,
    diameters: 1,
    presentations: 1,
    byMaterial: { 'Copper-nickel': 1, Bimetallic: 1 },
    unmatched: 1,
    notCached: 0,
    changes: [
      { coinId: 1, name: '2.50 Escudos', isWishlist: false, material: 'Copper-nickel', composition: 'Cuproníquel', weightGrams: 3.5, diameterMm: 20 },
      { coinId: 2, name: '2 Euro 2007', isWishlist: true, material: 'Bimetallic', composition: 'Bimetálica', presentation: 'coincard' },
    ],
    corrections: [
      { coinId: 3, name: '1 Escudo', isWishlist: false, material: 'Copper', previousMaterial: 'Silver', composition: 'Cobre banhado em prata' },
    ],
    corrected: 0,
    ...overrides,
  }
}

describe('AdminNumistaFillSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setLocale('pt-PT')
    mocks.previewNumistaFill.mockResolvedValue({ data: result() })
    mocks.applyNumistaFill.mockResolvedValue({ data: { ...result(), applied: true } })
    mocks.showConfirm.mockResolvedValue(true)
  })

  afterEach(() => setLocale('en'))

  it('shows what the fill would change, in Portuguese, before anything is saved', async () => {
    const wrapper = mount(AdminNumistaFillSection)
    await flushPromises()

    const summary = wrapper.get('[data-testid="numista-fill-summary"]').text()
    expect(summary).toContain('2 moedas a completar')
    expect(summary).toContain('Cuproníquel')
    expect(summary).toContain('Bimetálica')
    const changes = wrapper.get('[data-testid="numista-fill-changes"]').text()
    expect(changes).toContain('2.50 Escudos')
    expect(changes).toContain('3,5 g · 20 mm')
    expect(changes).toContain('Numista: Cuproníquel')
    expect(changes).toContain('Lista de desejos')
    expect(changes).toContain('Coincard')
    expect(wrapper.text()).toContain('1 moeda tem uma composição sem material correspondente')
    const corrections = wrapper.get('[data-testid="numista-fill-corrections"]').text()
    expect(corrections).toContain('Prata → Cobre')
    expect(wrapper.get('[data-testid="numista-fill-apply"]').text()).toBe('Guardar em 2 moedas')
    expect(mocks.applyNumistaFill).not.toHaveBeenCalled()
  })

  it('saves only the fill unless the corrections are ticked, after confirming', async () => {
    const wrapper = mount(AdminNumistaFillSection)
    await flushPromises()

    await wrapper.get('[data-testid="numista-fill-apply"]').trigger('click')
    await flushPromises()
    expect(mocks.showConfirm).toHaveBeenCalledWith(expect.stringContaining('2 moedas'), expect.anything())
    expect(mocks.applyNumistaFill).toHaveBeenCalledWith(false)
    expect(wrapper.text()).toContain('Feito: 2 moedas alteradas.')
    expect(mocks.previewNumistaFill).toHaveBeenCalledTimes(2)

    await wrapper.get('[data-testid="numista-fill-corrections-toggle"]').setValue(true)
    expect(wrapper.get('[data-testid="numista-fill-apply"]').text()).toBe('Guardar em 3 moedas')
    await wrapper.get('[data-testid="numista-fill-apply"]').trigger('click')
    await flushPromises()
    expect(mocks.applyNumistaFill).toHaveBeenLastCalledWith(true)
  })

  it('saves nothing when the confirmation is cancelled', async () => {
    mocks.showConfirm.mockResolvedValue(false)
    const wrapper = mount(AdminNumistaFillSection)
    await flushPromises()

    await wrapper.get('[data-testid="numista-fill-apply"]').trigger('click')
    await flushPromises()
    expect(mocks.applyNumistaFill).not.toHaveBeenCalled()
  })

  it('says when there is nothing to fill', async () => {
    mocks.previewNumistaFill.mockResolvedValue({
      data: result({ coins: 0, materials: 0, weights: 0, diameters: 0, presentations: 0, byMaterial: {}, unmatched: 0, changes: [], corrections: [] }),
    })
    const wrapper = mount(AdminNumistaFillSection)
    await flushPromises()

    expect(wrapper.find('[data-testid="numista-fill-nothing"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="numista-fill-apply"]').exists()).toBe(false)
  })

  it('says when the coins cannot be checked', async () => {
    mocks.previewNumistaFill.mockRejectedValue(new Error('offline'))
    const wrapper = mount(AdminNumistaFillSection)
    await flushPromises()

    expect(wrapper.text()).toContain('Não foi possível verificar as moedas.')
  })
})
