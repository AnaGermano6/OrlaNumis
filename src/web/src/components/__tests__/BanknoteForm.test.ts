import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'
import { beforeEach, describe, expect, it } from 'vitest'
import BanknoteForm from '../BanknoteForm.vue'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'

const stubs = {
  RouterLink: { template: '<a><slot /></a>' },
  AutocompleteInput: true,
  AuthenticatedImage: true,
  X: true,
  Camera: true,
}

function mountForm(form: Partial<Coin>, coinId?: number) {
  return mount(BanknoteForm, { props: { form, submitLabel: 'Guardar', coinId }, global: { stubs } })
}

describe('BanknoteForm', () => {
  beforeEach(() => setLocale('pt-PT'))

  it('asks what makes a banknote, not what makes a coin', async () => {
    const form = reactive<Partial<Coin>>({ kind: 'banknote', name: '1000 escudos', material: 'Paper', references: [] })
    const wrapper = mountForm(form)
    await flushPromises()

    const text = wrapper.text()
    for (const label of ['Banco emissor', 'Valor facial', 'Série ou emissão', 'Data impressa na nota', 'Número Pick', 'Número de série', 'Assinaturas', 'Tipo de nota', 'Largura (mm)', 'Altura (mm)', 'Impressor', 'Marca de água', 'Foto da frente', 'Foto do verso']) {
      expect(text).toContain(label)
    }
    // No coin fields, and the kind is only offered when editing.
    for (const label of ['Peso', 'Diâmetro', 'Casa da moeda']) expect(text).not.toContain(label)
    expect(wrapper.find('[data-testid="banknote-kind"]').exists()).toBe(false)
    expect(wrapper.get<HTMLSelectElement>('[data-testid="banknote-material"]').findAll('option').map((option) => option.text()))
      .toEqual(['Papel', 'Polímero', 'Híbrido (papel e polímero)', 'Outro'])
  })

  it('keeps the Pick number as a reference and the certificate grade as one text', async () => {
    const form = reactive<Partial<Coin>>({ kind: 'banknote', name: 'Nota', references: [], certifiedGrade: '' })
    const wrapper = mountForm(form)
    await flushPromises()

    await wrapper.get('[data-testid="banknote-pick"]').setValue('P188c')
    await wrapper.get('[data-testid="banknote-pick"]').trigger('change')
    expect(form.references).toEqual([{ catalog: 'Pick', number: '188c' }])

    await wrapper.get('[data-testid="banknote-company"]').setValue('PMG')
    await wrapper.get('[data-testid="banknote-score"]').setValue('66')
    await wrapper.get('[data-testid="banknote-epq"]').setValue(true)
    expect(form.certifiedGrade).toBe('PMG 66 EPQ')

    await wrapper.get('[data-testid="banknote-variant"]').setValue('replacement')
    expect(form.noteVariant).toBe('replacement')
    await wrapper.get('[data-testid="banknote-grade"]').setValue('UNC')
    expect(form.grade).toBe('UNC')
  })

  it('reads a saved grading company grade and lets an edited banknote become a coin', async () => {
    const form = reactive<Partial<Coin>>({ id: 7, kind: 'banknote', name: 'Nota', certifiedGrade: 'PMG 64 EPQ', references: [{ id: 1, coinId: 7, catalog: 'Pick', number: '190', volume: '', uri: '', createdAt: '', updatedAt: '' }] })
    const wrapper = mountForm(form, 7)
    await flushPromises()

    expect(wrapper.get<HTMLInputElement>('[data-testid="banknote-score"]').element.value).toBe('64')
    expect(wrapper.get<HTMLInputElement>('[data-testid="banknote-pick"]').element.value).toBe('190')
    await wrapper.get('[data-testid="banknote-kind"]').setValue('coin')
    expect(form.kind).toBe('coin')
  })
})
