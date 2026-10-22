import { t, type MessageKey } from '@/i18n'

/** Numista's grade scale, lowest to highest, as returned by its prices endpoint. */
export const NUMISTA_GRADES = ['g', 'vg', 'f', 'vf', 'xf', 'au', 'unc'] as const
export type NumistaGrade = (typeof NUMISTA_GRADES)[number]

const GRADE_NAME_KEYS: Record<NumistaGrade, MessageKey> = {
  g: 'numistaGrades.g',
  vg: 'numistaGrades.vg',
  f: 'numistaGrades.f',
  vf: 'numistaGrades.vf',
  xf: 'numistaGrades.xf',
  au: 'numistaGrades.au',
  unc: 'numistaGrades.unc',
}

function isNumistaGrade(value: string): value is NumistaGrade {
  return (NUMISTA_GRADES as readonly string[]).includes(value)
}

/** International abbreviation (G … UNC), the same in every locale. */
export function numistaGradeShort(grade: string): string {
  return grade.toUpperCase()
}

/** Descriptive name of a Numista grade in the current locale. */
export function numistaGradeName(grade: string): string {
  return isNumistaGrade(grade) ? t(GRADE_NAME_KEYS[grade]) : grade.toUpperCase()
}

// Words and abbreviations collectors use in the free-text grade field, checked
// from the highest grade down so "Good VF" reads as VF and "About Unc" as AU.
// Portuguese terms follow the commonly cited equivalences (REG≈G, BC≈F,
// MBC≈VF, Bela≈XF, Soberba/FDC≈UNC).
const GRADE_PATTERNS: [NumistaGrade, RegExp][] = [
  ['au', /\b(ABOUT|ALMOST) UNC|\b(A|CH )?AU\b|\bAUNC\b/],
  ['unc', /\b(UNC|UNCIRCULATED|MS|BU|BNC|FDC|SOB|SOBERBA|SOBERBO|FLOR DE CUNHO)\b/],
  ['xf', /\bEXTREMELY FINE\b|\b(A|G|CH )?(XF|EF)\b|\bBEL[AO]\b/],
  ['vf', /\bVERY FINE\b|\b(A|G|CH )?VF\b|\bMBC\b/],
  ['vg', /\bVERY GOOD\b|\b(A|G)?VG\b/],
  ['f', /\bFINE\b|\b(A|G|CH )?F\b|\bBC\b/],
  ['g', /\bGOOD\b|\b(A)?G\b|\bREG(ULAR)?\b|\bMC\b/],
]

/** Maps a coin's free-text grade ("Ch XF", "aEF", "MBC", "MS 63") to Numista's scale, or null. */
export function numistaGradeForCoinGrade(grade: string | null | undefined): NumistaGrade | null {
  const normalized = (grade ?? '')
    .toUpperCase()
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .replace(/[^A-Z0-9]+/g, ' ')
    .trim()
  if (!normalized) return null
  for (const [numistaGrade, pattern] of GRADE_PATTERNS) {
    if (pattern.test(normalized)) return numistaGrade
  }
  return null
}
