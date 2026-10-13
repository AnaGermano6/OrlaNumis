import { beforeEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import {
  banknoteGradeLabel,
  banknoteMaterialFromComposition,
  composeCertifiedGrade,
  noteVariantLabel,
  parseCertifiedGrade,
  pickNumber,
  withPickNumber,
} from '@/utils/banknoteAttributes'

describe('banknoteAttributes', () => {
  beforeEach(() => setLocale('pt-PT'))

  it('writes a grading company grade as one text and reads it back', () => {
    const grade = { company: 'PMG', score: 66, epq: true, star: false }
    expect(composeCertifiedGrade(grade)).toBe('PMG 66 EPQ')
    expect(parseCertifiedGrade('PMG 66 EPQ')).toEqual(grade)
    expect(composeCertifiedGrade({ company: 'PMG', score: 70, epq: true, star: true })).toBe('PMG 70 EPQ ★')
    expect(parseCertifiedGrade('PCGS Banknote 64 PPQ')).toEqual({ company: 'PCGS Banknote', score: 64, epq: false, star: false })
    // Without a score there is no grade yet.
    expect(composeCertifiedGrade({ company: 'PMG', score: null, epq: true, star: false })).toBe('')
    expect(parseCertifiedGrade('')).toEqual({ company: '', score: null, epq: false, star: false })
  })

  it('keeps the Pick number among the references', () => {
    const references = [{ catalog: 'Numista', number: '1234' }]
    const withPick = withPickNumber(references, 'P# 188c')
    expect(withPick).toEqual([{ catalog: 'Numista', number: '1234' }, { catalog: 'Pick', number: '188c' }])
    expect(pickNumber(withPick)).toBe('188c')
    expect(pickNumber(withPickNumber(withPick, '189'))).toBe('189')
    expect(withPickNumber(withPick, ' ')).toEqual(references)
    expect(pickNumber(undefined)).toBe('')
  })

  it('names materials, grades and variants of banknotes', () => {
    expect(banknoteMaterialFromComposition('Papel')).toBe('Paper')
    expect(banknoteMaterialFromComposition('Polymer')).toBe('Polymer')
    expect(banknoteMaterialFromComposition('Hybrid (paper/polymer)')).toBe('Hybrid')
    expect(banknoteMaterialFromComposition(undefined)).toBe('Other')
    expect(banknoteGradeLabel('XF')).toBe('XF — Soberba')
    expect(banknoteGradeLabel('MBC')).toBe('MBC')
    expect(noteVariantLabel('replacement')).toBe('Nota de substituição (estrela)')
    expect(noteVariantLabel('')).toBe('Normal')
  })
})
