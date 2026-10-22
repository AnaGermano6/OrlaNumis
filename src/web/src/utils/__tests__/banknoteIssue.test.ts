import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import type { NumistaIssueSummary } from '@/api/client'
import type { Coin } from '@/types'
import { applyNumistaIssue, issueSignatures, numistaTypeIdOf } from '@/utils/banknoteIssue'
import { numistaIssueDate, numistaIssueLabel, numistaIssuePrintedDate, signerSurname } from '@/utils/numistaCatalog'

const ISSUE: NumistaIssueSummary = {
  id: 542229,
  year: 1978,
  month: 9,
  day: 13,
  mintage: 29645000,
  pick: '176a',
  signatures: [
    { name: 'Emílio Rui da Veiga Peixoto Vilar', title: 'Deputy Governor' },
    { name: 'Abel António Pinto dos Reis', title: 'Administrator' },
  ],
  comment: 'P#176a; Silva# 68 - Signatures: Emílio Rui da Veiga Peixoto Vilar & Abel António Pinto dos Reis',
}

describe('banknote issues', () => {
  afterEach(() => setLocale('en'))

  it('labels an issue by its date, Pick number and signatories instead of its comment', () => {
    setLocale('pt-PT')
    expect(numistaIssueLabel(ISSUE).replace(/\s/g, ' ')).toBe('13/09/1978 · P176a · Vilar / Reis · 29 645 000 exemplares')
    expect(numistaIssueLabel({ id: 3, year: 1986, variant: 'replacement', comment: '1986-06-16 Replacement note with serial # ZA-ZB-ZC' }))
      .toBe('1986 · Nota de substituição (estrela) · 1986-06-16 Replacement note with serial…')
    // A coin's issue reads as before.
    expect(numistaIssueLabel({ id: 102, year: 1968, mintLetter: 'L', comment: 'Lisbon' })).toBe('1968 · L · Lisbon')
  })

  it('writes the date printed on the note in the language of the app', () => {
    setLocale('pt-PT')
    expect(numistaIssuePrintedDate(ISSUE)).toBe('13 de setembro de 1978')
    expect(numistaIssuePrintedDate({ id: 1, year: 1978, month: 9 })).toBe('setembro de 1978')
    expect(numistaIssueDate({ id: 1, year: 1978, month: 9 })).toBe('09/1978')
    expect(numistaIssuePrintedDate({ id: 1, year: 1978 })).toBe('1978')
    expect(numistaIssuePrintedDate({ id: 1 })).toBe('')
    setLocale('en')
    expect(numistaIssuePrintedDate(ISSUE)).toBe('September 13, 1978')
  })

  it('names a signatory by surname', () => {
    expect(signerSurname(' José da Silva Lopes ')).toBe('Lopes')
    expect(signerSurname('Roberto Campos Neto')).toBe('Campos Neto')
    expect(signerSurname('Neto')).toBe('Neto')
  })

  it('fills in a banknote from the issue chosen', () => {
    const form: Partial<Coin> = {
      signatures: 'typed', noteVariant: 'replacement', dateRange: '1978', year: 1978,
      references: [
        { id: 1, coinId: 9, catalog: 'Numista', number: '202243', volume: '', uri: '', createdAt: '', updatedAt: '' },
        { id: 2, coinId: 9, catalog: 'Pick', number: '176', volume: '', uri: '', createdAt: '', updatedAt: '' },
      ],
    }
    setLocale('pt-PT')
    applyNumistaIssue(form, ISSUE)
    expect(form).toMatchObject({
      numistaIssueId: 542229,
      signatures: 'Emílio Rui da Veiga Peixoto Vilar / Abel António Pinto dos Reis',
      noteVariant: '',
      dateRange: '13 de setembro de 1978',
      year: 1978,
    })
    expect(form.references?.map((reference) => `${reference.catalog} ${reference.number}`)).toEqual(['Numista 202243', 'Pick 176a'])

    // What the issue does not say is left as it was.
    applyNumistaIssue(form, { id: 9, variant: 'specimen' })
    expect(form).toMatchObject({ numistaIssueId: 9, signatures: 'Emílio Rui da Veiga Peixoto Vilar / Abel António Pinto dos Reis', noteVariant: 'specimen', year: 1978 })
    expect(numistaTypeIdOf(form.references)).toBe(202243)
  })

  it('keeps the signatures within what a banknote stores', () => {
    const long = issueSignatures({ signatures: Array.from({ length: 20 }, (_, i) => ({ name: `Signatory number ${i}` })) })
    expect(long).toHaveLength(300)
    expect(long.endsWith('…')).toBe(true)
    expect(numistaTypeIdOf([{ catalog: 'Numista', number: 'abc' }])).toBeUndefined()
    expect(numistaTypeIdOf(undefined)).toBeUndefined()
  })
})
