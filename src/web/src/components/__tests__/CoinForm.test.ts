import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { reactive, ref } from 'vue'
import CoinForm from '../CoinForm.vue'
import type { Coin } from '@/types'

const apiMocks = vi.hoisted(() => ({
  getMintLocations: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  getMintLocations: apiMocks.getMintLocations,
}))

vi.mock('@/composables/usePwa', () => ({
  usePwa: () => ({ isPwa: false }),
}))

vi.mock('@/composables/useCoinOptions', () => ({
  useCoinOptions: () => ({
    categoryOptions: ref(['Modern', 'Other']),
    eraOptions: ref(['ancient']),
    materialOptions: ref(['Silver']),
    loadOptions: vi.fn(),
  }),
}))

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const coinFormPath = path.resolve(__dirname, '../CoinForm.vue')

describe('CoinForm', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    apiMocks.getMintLocations.mockResolvedValue({ data: { mintLocations: [] } })
  })

  it('renders section titles inside the form sections with larger heading styles', () => {
    const source = fs.readFileSync(coinFormPath, 'utf8')
    const mainCssPath = path.resolve(__dirname, '../../assets/styles/main.css')
    const mainCss = fs.readFileSync(mainCssPath, 'utf8')

    // Section titles are <h2>s (not <legend>s) styled with the Tailwind
    // `text-lg` utility instead of a dedicated `.form-section-title` class.
    expect(source).toContain(`<h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('coinForm.sections.basic') }}</h2>`)
    expect(source).not.toContain('<legend>Basic Information</legend>')
    // `text-lg` resolves to the same 1.2rem the old CSS block hard-coded, and
    // `mb-4` is the Tailwind 1rem spacing step used for the old bottom margin.
    expect(mainCss).toMatch(/--text-lg:\s*1\.2rem/)
  })

  it('allows purchase form fields to shrink within grid columns', () => {
    const source = fs.readFileSync(coinFormPath, 'utf8')

    expect(source).toContain('<input v-model="form.purchaseDate" class="form-input" type="date" />')
    // `.form-group { min-width: 0 }` is now the `min-w-0` utility applied
    // directly alongside `form-group` on each field wrapper.
    expect(source).toContain('class="form-group min-w-0"')
  })

  it('keeps a current custom era selectable in the edit form', () => {
    const source = fs.readFileSync(coinFormPath, 'utf8')

    expect(source).toContain('v-for="era in displayedEraOptions"')
    expect(source).toContain('const displayedEraOptions = computed(() => {')
    expect(source).toContain('return [currentEra, ...eraOptions.value]')
  })

  it('keeps the free-text Ruler field, without the removed Imperial figure picker', () => {
    const source = fs.readFileSync(coinFormPath, 'utf8')

    expect(source).toContain(`<AutocompleteInput v-model="form.ruler!" field="ruler" :placeholder="t('coinForm.rulerPlaceholder')" />`)
    expect(source).not.toContain('ImperialFigurePicker')
    expect(source).not.toContain('romanImperialFigureId')
  })

  it('replaces the free-text mint input with a managed dropdown offering Unknown, My Mints, Mints, and Create new', () => {
    const source = fs.readFileSync(coinFormPath, 'utf8')

    expect(source).not.toContain('<input v-model="form.mint" class="form-input" placeholder="e.g. Rome" />')
    expect(source).toContain('v-model="mintLocationIdModel"')
    expect(source).toContain(`<option value="">{{ t('coinForm.mintUnknown') }}</option>`)
    expect(source).toContain(`:label="t('coinForm.myMints')"`)
    expect(source).toContain(`:label="t('coinForm.mints')"`)
    expect(source).toContain(`<option value="__create__">{{ t('coinForm.createMint') }}</option>`)
    expect(source).toContain("import CreateMintModal from '@/components/CreateMintModal.vue'")
  })

  it('nudges the user to link a legacy free-text mint that has no matching location, without blocking the form', () => {
    const source = fs.readFileSync(coinFormPath, 'utf8')

    expect(source).toContain('form.mint && !form.mintLocationId')
    expect(source).toContain("t('coinForm.unlinkedMint'")
  })

  it('opens the create-mint modal from the dropdown and adopts the newly created mint', () => {
    const source = fs.readFileSync(coinFormPath, 'utf8')

    expect(source).toContain("value === '__create__'")
    expect(source).toContain('showCreateMintModal.value = true')
    expect(source).toContain('function onMintCreated(mintLocation: MintLocation)')
    expect(source).toContain('props.form.mintLocationId = mintLocation.id')
  })

  // Catalogue measurements have more decimals than a fixed step allows
  // (25.75 mm for a 2 euro coin, 7.128 g), and the browser would refuse to save.
  it('accepts weights and diameters with any number of decimals', async () => {
    const form = { name: 'Coin', category: 'Modern', material: 'Silver', weightGrams: 7.128, diameterMm: 25.75 }
    const wrapper = mount(CoinForm, {
      props: { form, submitLabel: 'Save' },
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          AutocompleteInput: true,
          CreateMintModal: true,
          AuthenticatedImage: true,
          X: true,
          Camera: true,
        },
      },
    })
    await flushPromises()

    const inputs = ['weightGrams', 'diameterMm'].map((field) =>
      wrapper.findAll('input[type="number"]').find((input) => (input.element as HTMLInputElement).valueAsNumber === form[field as 'weightGrams' | 'diameterMm'])!)
    for (const input of inputs) {
      expect(input.attributes('step')).toBe('any')
      expect((input.element as HTMLInputElement).validity.valid).toBe(true)
    }
  })

  it('chooses the grade from a list that says what each means, keeping one written otherwise', async () => {
    const { setLocale } = await import('@/i18n')
    setLocale('pt-PT')
    const form = reactive<Partial<Coin>>({ name: 'Coin', category: 'Modern', material: 'Silver', grade: 'S' })
    const wrapper = mount(CoinForm, {
      props: { form, submitLabel: 'Save' },
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' }, AutocompleteInput: true, CreateMintModal: true, AuthenticatedImage: true, X: true, Camera: true } },
    })
    await flushPromises()

    const select = wrapper.get('[data-testid="coin-grade"]')
    expect(select.element.tagName).toBe('SELECT')
    expect(wrapper.get('label[for="coin-grade"]').text()).toBe('Grau')
    // The list says what each grade means: no guide besides it.
    expect(wrapper.find('[data-testid="grade-guide"]').exists()).toBe(false)
    expect(select.findAll('optgroup').map((group) => group.attributes('label'))).toEqual(['Conservação', 'Acabamentos de coleção'])
    expect(select.findAll('option').map((option) => option.text())).toEqual([
      'Não indicada', 'S',
      'UNC — Sem circulação (Soberba, FDC)', 'AU — Quase sem circulação', 'XF — Extremamente bem conservada (Bela)', 'VF — Muito bem conservada (MBC)',
      'F — Bem conservada (BC)', 'VG — Entre regular e bem conservada', 'G — Regular (REG)',
      'BNC — Brilhante não circulada', 'PRF — Proof',
    ])
    expect((select.element as HTMLSelectElement).value).toBe('S')

    await select.setValue('VF')
    expect(form.grade).toBe('VF')
    // The written grade is no longer offered once another is chosen.
    expect(select.findAll('option').map((option) => option.text())).not.toContain('S')
    await select.setValue('')
    expect(form.grade).toBe('')
    setLocale('en')
  })

  it('sets the circulation by hand, or leaves it to Numista', async () => {
    const form = reactive<Partial<Coin>>({ name: 'Coin', category: 'Modern', material: 'Silver', circulation: '' })
    const wrapper = mount(CoinForm, {
      props: { form, submitLabel: 'Save', coinId: 99 },
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          AutocompleteInput: true,
          CreateMintModal: true,
          AuthenticatedImage: true,
          X: true,
          Camera: true,
        },
      },
    })
    await flushPromises()

    const select = wrapper.get('#coin-circulation')
    expect(select.findAll('option').map((option) => option.text())).toEqual([
      'Automatic (as Numista says)',
      'Current',
      'Current · Commemorative',
      'Demonetized',
      'Demonetized · Commemorative',
      'Non-circulating',
      'Non-circulating · Commemorative',
    ])
    await select.setValue('current_commemorative')
    expect(form.circulation).toBe('current_commemorative')
    await select.setValue('')
    expect(form.circulation).toBe('')
  })

  it('says whether the item is a coin or a banknote, a coin when it does not say', async () => {
    const form = reactive<Partial<Coin>>({ name: '1000 escudos', category: 'Modern', material: 'Other' })
    const wrapper = mount(CoinForm, {
      props: { form, submitLabel: 'Save', coinId: 99 },
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          AutocompleteInput: true,
          CreateMintModal: true,
          AuthenticatedImage: true,
          X: true,
          Camera: true,
        },
      },
    })
    await flushPromises()

    const select = wrapper.get<HTMLSelectElement>('[data-testid="coin-kind"]')
    expect(select.element.value).toBe('coin')
    expect(select.findAll('option').map((option) => option.text())).toEqual(['Coin', 'Banknote'])
    await select.setValue('banknote')
    expect(form.kind).toBe('banknote')
  })

  it('takes a typed year, clears it to an empty value and shows the year it would read from the date', async () => {
    apiMocks.getMintLocations.mockResolvedValue({ data: [] })
    const form = reactive<Partial<Coin>>({ name: 'Denarius', category: 'Roman', material: 'Silver', dateRange: 'c. 46-44 BC', year: null })
    const wrapper = mount(CoinForm, {
      props: { form, submitLabel: 'Save' },
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' }, AutocompleteInput: true, CreateMintModal: true, AuthenticatedImage: true } },
    })
    await flushPromises()

    const year = wrapper.get('#coin-year')
    expect(year.attributes('placeholder')).toBe('Automatic: 46 BC')

    await year.setValue('-44')
    expect(form.year).toBe(-44)
    await year.setValue('1968.7')
    expect(form.year).toBe(1968)
    await year.setValue('0')
    expect(form.year).toBeNull()
    await year.setValue('')
    expect(form.year).toBeNull()

    form.dateRange = 'Undated'
    await flushPromises()
    expect(year.attributes('placeholder')).toBe('e.g. 1968')
  })
})
