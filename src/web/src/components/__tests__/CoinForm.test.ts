import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { reactive, ref } from 'vue'
import CoinForm from '../CoinForm.vue'
import type { Coin } from '@/types'

const apiMocks = vi.hoisted(() => ({
  getStorageLocations: vi.fn(),
  getStorageLocationOccupancy: vi.fn(),
  getMintLocations: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  getStorageLocations: apiMocks.getStorageLocations,
  getStorageLocationOccupancy: apiMocks.getStorageLocationOccupancy,
  getMintLocations: apiMocks.getMintLocations,
}))

vi.mock('@/composables/usePwa', () => ({
  usePwa: () => ({ isPwa: false }),
}))

vi.mock('@/composables/useCoinOptions', () => ({
  useCoinOptions: () => ({
    categoryOptions: ref(['Roman', 'Greek']),
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
    apiMocks.getStorageLocations.mockResolvedValue({
      data: {
        storageLocations: [
          { id: 1, name: 'Safe', type: 'standard', rows: null, columns: null, capacity: 0, occupied: 0 },
          { id: 2, name: 'Cabinet A', type: 'tray', rows: 2, columns: 2, capacity: 4, occupied: 2 },
          { id: 3, name: 'Cabinet B', type: 'tray', rows: 1, columns: 2, capacity: 2, occupied: 0 },
        ],
      },
    })
    apiMocks.getStorageLocationOccupancy.mockResolvedValue({
      data: { id: 2, rows: 2, columns: 2, capacity: 4, occupiedSlots: [2, 4], currentCoinSlot: 4 },
    })
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

  it('shows the Imperial figure picker only for Roman coins, alongside the free-text Ruler field', () => {
    const source = fs.readFileSync(coinFormPath, 'utf8')

    expect(source).toContain('v-if="form.category === \'Roman\'"')
    expect(source).toContain('<ImperialFigurePicker v-model="form.romanImperialFigureId!" />')
    expect(source).toContain(`<AutocompleteInput v-model="form.ruler!" field="ruler" :placeholder="t('coinForm.rulerPlaceholder')" />`)
    expect(source).toContain("import ImperialFigurePicker from '@/components/ImperialFigurePicker.vue'")
    expect(source).toContain("if (category !== 'Roman')")
    expect(source).toContain('props.form.romanImperialFigureId = null')
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
    const form = { name: 'Coin', category: 'Roman', material: 'Silver', weightGrams: 7.128, diameterMm: 25.75 }
    const wrapper = mount(CoinForm, {
      props: { form, submitLabel: 'Save' },
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          AutocompleteInput: true,
          ImperialFigurePicker: true,
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

  it('shows exact coordinates only for trays and keeps the current coin slot available', async () => {
    const form = { name: 'Coin', category: 'Roman', material: 'Silver', storageLocationId: 2, storageSlot: 4 }
    const wrapper = mount(CoinForm, {
      props: { form, submitLabel: 'Save', coinId: 99 },
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          AutocompleteInput: true,
          ImperialFigurePicker: true,
          CreateMintModal: true,
          AuthenticatedImage: true,
          X: true,
          Camera: true,
        },
      },
    })
    await flushPromises()

    expect(apiMocks.getStorageLocationOccupancy).toHaveBeenCalledWith(2, 99)
    expect(wrapper.findAll('[role="gridcell"]')).toHaveLength(4)
    expect(wrapper.get('[aria-label="Row 1, column 2, occupied"]').attributes('disabled')).toBeDefined()
    expect(wrapper.get('[aria-label="Row 2, column 2, selected"]').attributes('disabled')).toBeUndefined()
  })

  it('resets stale slots on location changes and requires a new tray slot before submit', async () => {
    const form = { name: 'Coin', category: 'Roman', material: 'Silver', storageLocationId: 2, storageSlot: 4 }
    const wrapper = mount(CoinForm, {
      props: { form, submitLabel: 'Save', coinId: 99 },
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          AutocompleteInput: true,
          ImperialFigurePicker: true,
          CreateMintModal: true,
          AuthenticatedImage: true,
          X: true,
          Camera: true,
        },
      },
    })
    await flushPromises()

    apiMocks.getStorageLocationOccupancy.mockResolvedValueOnce({
      data: { id: 3, rows: 1, columns: 2, capacity: 2, occupiedSlots: [], currentCoinSlot: null },
    })
    const storageSelect = wrapper.findAll('select.form-select').find((select) => select.text().includes('Cabinet B'))
    await storageSelect!.setValue('3')
    await flushPromises()

    expect(form.storageSlot).toBeNull()
    await wrapper.find('form').trigger('submit')
    expect(wrapper.text()).toContain('Choose an exact tray slot before saving.')
    expect(wrapper.emitted('submit')).toBeUndefined()

    await wrapper.get('[aria-label="Row 1, column 1, available"]').trigger('click')
    await wrapper.find('form').trigger('submit')
    expect(form.storageSlot).toBe(1)
    expect(wrapper.emitted('submit')).toHaveLength(1)
  })

  it('refreshes stale occupancy after a conflict without losing other form data', async () => {
    const form = {
      name: 'Preserved name',
      category: 'Roman',
      material: 'Silver',
      notes: 'Preserved notes',
      storageLocationId: 2,
      storageSlot: 1,
    }
    const wrapper = mount(CoinForm, {
      props: { form, submitLabel: 'Save', coinId: 99 },
      global: {
        stubs: {
          RouterLink: { template: '<a><slot /></a>' },
          AutocompleteInput: true,
          ImperialFigurePicker: true,
          CreateMintModal: true,
          AuthenticatedImage: true,
          X: true,
          Camera: true,
        },
      },
    })
    await flushPromises()

    apiMocks.getStorageLocationOccupancy.mockRejectedValueOnce(new Error('stale'))
    await (wrapper.vm as unknown as { refreshStorageOccupancy: () => Promise<void> }).refreshStorageOccupancy()
    await flushPromises()
    expect(wrapper.text()).toContain('Tray occupancy changed or could not be loaded.')

    apiMocks.getStorageLocationOccupancy.mockResolvedValueOnce({
      data: { id: 2, rows: 2, columns: 2, capacity: 4, occupiedSlots: [1, 2], currentCoinSlot: null },
    })
    await wrapper.findAll('button').find((button) => button.text() === 'Retry')!.trigger('click')
    await flushPromises()

    expect(form.name).toBe('Preserved name')
    expect(form.notes).toBe('Preserved notes')
    expect(form.storageSlot).toBe(1)
    expect(wrapper.get('[aria-label="Row 1, column 1, occupied"]').attributes('disabled')).toBeDefined()
  })

  it('takes a typed year, clears it to an empty value and shows the year it would read from the date', async () => {
    apiMocks.getStorageLocations.mockResolvedValue({ data: [] })
    apiMocks.getMintLocations.mockResolvedValue({ data: [] })
    const form = reactive<Partial<Coin>>({ name: 'Denarius', category: 'Roman', material: 'Silver', dateRange: 'c. 46-44 BC', year: null })
    const wrapper = mount(CoinForm, {
      props: { form, submitLabel: 'Save' },
      global: { stubs: { RouterLink: { template: '<a><slot /></a>' }, AutocompleteInput: true, ImperialFigurePicker: true, CreateMintModal: true, AuthenticatedImage: true } },
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
