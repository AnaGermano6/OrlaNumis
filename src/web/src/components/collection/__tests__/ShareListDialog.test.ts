import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ShareListDialog from '@/components/collection/ShareListDialog.vue'
import { setLocale } from '@/i18n'

const mocks = vi.hoisted(() => ({ downloadShareList: vi.fn() }))
vi.mock('@/api/client', () => mocks)

function mountDialog(initialList: 'trades' | 'wishlist' | 'both' = 'trades') {
  return mount(ShareListDialog, { props: { open: true, initialList }, attachTo: document.body })
}

const dialog = () => document.querySelector<HTMLElement>('[data-testid="share-list-dialog"]')
const field = <T extends HTMLElement>(selector: string) => document.querySelector<T>(selector)!

describe('ShareListDialog', () => {
  let clicked: string[]

  beforeEach(() => {
    vi.clearAllMocks()
    setLocale('pt-PT')
    clicked = []
    URL.createObjectURL = vi.fn(() => 'blob:share')
    URL.revokeObjectURL = vi.fn()
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(function (this: HTMLAnchorElement) { clicked.push(this.download) })
    mocks.downloadShareList.mockResolvedValue({ blob: new Blob(['x']), filename: 'procuro-2026-10-21.csv' })
  })

  afterEach(() => {
    setLocale('en')
    vi.restoreAllMocks()
    document.body.innerHTML = ''
  })

  it('opens on the list it was asked for, and downloads the file chosen', async () => {
    const wrapper = mountDialog('wishlist')
    await flushPromises()
    expect(dialog()?.getAttribute('aria-modal')).toBe('true')
    expect(field<HTMLInputElement>('[data-testid="share-list-wishlist"]').checked).toBe(true)

    field<HTMLInputElement>('[data-testid="share-format-csv"]').click()
    field<HTMLInputElement>('[data-testid="share-values"]').click()
    const contact = field<HTMLInputElement>('#share-contact')
    contact.value = '  ana@exemplo.pt '
    contact.dispatchEvent(new Event('input'))
    field<HTMLButtonElement>('[data-testid="share-download"]').click()
    await flushPromises()

    expect(mocks.downloadShareList).toHaveBeenCalledWith({ list: 'wishlist', format: 'csv', lang: 'pt', includeValues: true, contact: 'ana@exemplo.pt' })
    expect(clicked).toEqual(['procuro-2026-10-21.csv'])
    expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:share')
    expect(wrapper.emitted('close')).toHaveLength(1)
    wrapper.unmount()
  })

  it('asks for a PDF of the trades without values by default, in English when the app is', async () => {
    setLocale('en')
    const wrapper = mountDialog()
    await flushPromises()
    field<HTMLButtonElement>('[data-testid="share-download"]').click()
    await flushPromises()
    expect(mocks.downloadShareList).toHaveBeenCalledWith({ list: 'trades', format: 'pdf', lang: 'en', includeValues: false, contact: '' })
    wrapper.unmount()
  })

  it('says when the list could not be made, and stays open', async () => {
    mocks.downloadShareList.mockRejectedValue(new Error('offline'))
    const wrapper = mountDialog()
    await flushPromises()
    field<HTMLButtonElement>('[data-testid="share-download"]').click()
    await flushPromises()
    expect(dialog()?.textContent).toContain('Não foi possível preparar a lista.')
    expect(wrapper.emitted('close')).toBeUndefined()
    wrapper.unmount()
  })

  it('closes with Escape or Cancel', async () => {
    const wrapper = mountDialog()
    await flushPromises()
    dialog()!.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    expect(wrapper.emitted('close')).toHaveLength(1)
    ;[...document.querySelectorAll<HTMLButtonElement>('button')].find((button) => button.textContent?.trim() === 'Cancelar')!.click()
    expect(wrapper.emitted('close')).toHaveLength(2)
    wrapper.unmount()
  })
})
