import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import { numistaGradeForCoinGrade, numistaGradeName, numistaGradeShort } from '@/utils/numistaGrades'

describe('numistaGradeForCoinGrade', () => {
  it.each([
    ['VF', 'vf'],
    ['Choice XF', 'xf'],
    ['Ch XF 4/5 4/5', 'xf'],
    ['aEF', 'xf'],
    ['gVF', 'vf'],
    ['Good VF', 'vf'],
    ['Very Fine', 'vf'],
    ['Fine', 'f'],
    ['F 12', 'f'],
    ['VG', 'vg'],
    ['Very Good', 'vg'],
    ['Good', 'g'],
    ['AU 58', 'au'],
    ['About Unc', 'au'],
    ['MS 63', 'unc'],
    ['UNC', 'unc'],
    ['BNC', 'unc'],
    ['VF/XF', 'xf'],
    ['MBC', 'vf'],
    ['BC', 'f'],
    ['Bela', 'xf'],
    ['FDC', 'unc'],
    ['Flor de cunho', 'unc'],
    ['REG', 'g'],
  ])('maps %s to %s', (grade, expected) => {
    expect(numistaGradeForCoinGrade(grade)).toBe(expected)
  })

  it('returns null for empty or unknown grades', () => {
    expect(numistaGradeForCoinGrade('')).toBeNull()
    expect(numistaGradeForCoinGrade(undefined)).toBeNull()
    expect(numistaGradeForCoinGrade('Holed, cleaned')).toBeNull()
  })
})

describe('numista grade labels', () => {
  afterEach(() => setLocale('en'))

  it('keeps the international abbreviation and translates the name', () => {
    expect(numistaGradeShort('xf')).toBe('XF')
    expect(numistaGradeName('xf')).toBe('Extremely Fine')
    setLocale('pt-PT')
    expect(numistaGradeShort('xf')).toBe('XF')
    expect(numistaGradeName('vf')).toBe('Muito bem conservada')
    expect(numistaGradeName('unknown')).toBe('UNKNOWN')
  })
})
