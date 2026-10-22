import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ExportDialog from '@/components/collection/ExportDialog.vue'
import { setLocale } from '@/i18n'

const mocks = vi.hoisted(() => ({ exportCoins: vi.fn() }))
vi.mock('@/api/client', () => mocks)

const PARAMS = { kind: 'coin', country: 'Portugal', wishlist: 'false', sort: 'name', order: 'asc', seed: undefined }

function mountDialog(props: Record<string, unknown> = {}) {
  return mount(ExportDialog, { props: { open: true, params: PARAMS, count: 3, filtered: true, ...props }, attachTo: document.body })
}

const field = <T extends HTMLElement>(selector: string) => document.querySelector<T>(selector)!
const text = () => document.querySelector('[data-testid="export-dialog"]')?.textContent ?? ''

describe('ExportDialog', () => {
  let clicked: string[]

  beforeEach(() => {
    vi.clearAllMocks()
    setLocale('pt-PT')
    clicked = []
    URL.createObjectURL = vi.fn(() => 'blob:export')
    URL.revokeObjectURL = vi.fn()
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(function (this: HTMLAnchorElement) { clicked.push(this.download) })
    mocks.exportCoins.mockResolvedValue({ blob: new Blob(['x']), filename: 'colecao-2026-10-21.xlsx' })
  })

  afterEach(() => {
    setLocale('en')
    vi.restoreAllMocks()
    document.body.innerHTML = ''
  })

  it("exports the gallery's filtered coins as an Excel workbook by default", async () => {
    const wrapper = mountDialog()
    await flushPromises()
    expect(text()).toContain('Só as 3 moedas filtradas')
    expect(field<HTMLInputElement>('[data-testid="export-scope-shown"]').checked).toBe(true)
    expect(field<HTMLInputElement>('[data-testid="export-format-xlsx"]').checked).toBe(true)

    field<HTMLButtonElement>('[data-testid="export-download"]').click()
    await flushPromises()
    expect(mocks.exportCoins).toHaveBeenCalledWith(PARAMS, 'xlsx', 'pt')
    expect(clicked).toEqual(['colecao-2026-10-21.xlsx'])
    expect(wrapper.emitted('close')).toHaveLength(1)
    wrapper.unmount()
  })

  it('exports the whole collection as a CSV, in the gallery order, without its filters', async () => {
    const wrapper = mountDialog({ filtered: false })
    await flushPromises()
    expect(text()).toContain('As 3 moedas da galeria')
    field<HTMLInputElement>('[data-testid="export-scope-all"]').click()
    field<HTMLInputElement>('[data-testid="export-format-csv"]').click()
    field<HTMLButtonElement>('[data-testid="export-download"]').click()
    await flushPromises()
    expect(mocks.exportCoins).toHaveBeenCalledWith({ wishlist: 'false', sort: 'name', order: 'asc', seed: undefined }, 'csv', 'pt')
    wrapper.unmount()
  })

  it('names the banknotes in the banknote gallery, in English when the app is', async () => {
    setLocale('en')
    const wrapper = mountDialog({ kind: 'banknote', count: 1, filtered: true })
    await flushPromises()
    expect(text()).toContain('Only the filtered banknote')
    field<HTMLButtonElement>('[data-testid="export-download"]').click()
    await flushPromises()
    expect(mocks.exportCoins).toHaveBeenCalledWith(PARAMS, 'xlsx', 'en')
    wrapper.unmount()
  })

  it('says when the spreadsheet could not be made, and stays open', async () => {
    mocks.exportCoins.mockRejectedValue(new Error('offline'))
    const wrapper = mountDialog()
    await flushPromises()
    field<HTMLButtonElement>('[data-testid="export-download"]').click()
    await flushPromises()
    expect(text()).toContain('Não foi possível preparar a folha de cálculo.')
    expect(wrapper.emitted('close')).toBeUndefined()
    wrapper.unmount()
  })
})
