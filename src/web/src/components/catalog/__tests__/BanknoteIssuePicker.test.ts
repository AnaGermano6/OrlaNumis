import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { reactive } from 'vue'
import BanknoteIssuePicker from '../BanknoteIssuePicker.vue'
import { setLocale } from '@/i18n'
import type { NumistaIssueSummary } from '@/api/client'
import type { Coin } from '@/types'

const mocks = vi.hoisted(() => ({ getNumistaTypeIssues: vi.fn() }))
vi.mock('@/api/client', () => mocks)

const ISSUES: NumistaIssueSummary[] = [
  { id: 1, year: 1978, month: 9, day: 13, pick: '176a', signatures: [{ name: 'Emílio Vilar' }, { name: 'Abel Reis' }] },
  { id: 2, year: 1978, month: 9, day: 13, pick: '176a', signatures: [{ name: 'Emílio Vilar' }, { name: 'Joaquim Mestre' }] },
  { id: 3, year: 1978, pick: '176s', variant: 'specimen' },
]

function reference(catalog: string, number: string) {
  return { id: 0, coinId: 7, catalog, number, volume: '', uri: '', createdAt: '', updatedAt: '' }
}

function linkedNote(extra: Partial<Coin> = {}): Partial<Coin> {
  return reactive<Partial<Coin>>({ kind: 'banknote', references: [reference('Numista', '202243'), reference('Pick', '176')], signatures: '', ...extra })
}

describe('BanknoteIssuePicker', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    setLocale('pt-PT')
    mocks.getNumistaTypeIssues.mockResolvedValue({ data: { status: 'ok', issues: ISSUES, stale: false } })
  })

  afterEach(() => setLocale('en'))

  it('shows nothing for a note not linked to Numista', () => {
    const wrapper = mount(BanknoteIssuePicker, { props: { form: reactive<Partial<Coin>>({ references: [] }) } })
    expect(wrapper.find('[data-testid="banknote-issue"]').exists()).toBe(false)
  })

  it('loads the issues only when asked, and fills the note from the one chosen', async () => {
    const form = linkedNote()
    const wrapper = mount(BanknoteIssuePicker, { props: { form } })
    expect(mocks.getNumistaTypeIssues).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Escolher a emissão')

    await wrapper.find('[data-testid="banknote-issue-load"]').trigger('click')
    await flushPromises()
    expect(mocks.getNumistaTypeIssues).toHaveBeenCalledWith(202243)
    const options = wrapper.findAll('[data-testid="banknote-issue-select"] option')
    expect(options.map((option) => option.text())).toEqual([
      'Escolha a emissão…',
      '13/09/1978 · P176a · Vilar / Reis',
      '13/09/1978 · P176a · Vilar / Mestre',
      '1978 · P176s · Espécime',
    ])

    await wrapper.find('[data-testid="banknote-issue-select"]').setValue('2')
    expect(form).toMatchObject({ numistaIssueId: 2, signatures: 'Emílio Vilar / Joaquim Mestre', dateRange: '13 de setembro de 1978', noteVariant: '' })
    expect(form.references?.find((item) => item.catalog === 'Pick')?.number).toBe('176a')
    expect(wrapper.find('[data-testid="banknote-issue-hint"]').text()).toContain('Preenchido a partir da emissão escolhida')
  })

  it('shows the issue already chosen', async () => {
    const wrapper = mount(BanknoteIssuePicker, { props: { form: linkedNote({ numistaIssueId: 3 }) } })
    expect(wrapper.text()).toContain('Mudar a emissão')
    await wrapper.find('[data-testid="banknote-issue-load"]').trigger('click')
    await flushPromises()
    expect((wrapper.find('[data-testid="banknote-issue-select"]').element as HTMLSelectElement).value).toBe('3')
  })

  it('says when Numista has no issues or cannot be reached', async () => {
    mocks.getNumistaTypeIssues.mockResolvedValueOnce({ data: { status: 'no_issues', issues: [], stale: false } })
    const wrapper = mount(BanknoteIssuePicker, { props: { form: linkedNote() } })
    await wrapper.find('[data-testid="banknote-issue-load"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('O Numista não indica emissões desta nota.')

    mocks.getNumistaTypeIssues.mockResolvedValueOnce({ data: { status: 'quota_limited', issues: [], stale: false } })
    await wrapper.find('[data-testid="banknote-issue-load"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('limite mensal de pedidos ao Numista')

    mocks.getNumistaTypeIssues.mockRejectedValueOnce(new Error('offline'))
    await wrapper.find('[data-testid="banknote-issue-load"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Não foi possível carregar as emissões do Numista')
  })
})
