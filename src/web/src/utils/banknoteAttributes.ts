// What a banknote has that a coin does not: what it is printed on, the
// collectors' grading scale for paper money, replacement and specimen notes,
// a grading company's grade ("PMG 66 EPQ") and its Pick catalogue number.
import { t } from '@/i18n'
import type { CoinReference } from '@/types'

/** What banknotes are printed on; stored as the item's material. */
export const BANKNOTE_MATERIALS = ['Paper', 'Polymer', 'Hybrid', 'Other'] as const
export type BanknoteMaterial = (typeof BANKNOTE_MATERIALS)[number]

/** The usual grades of paper money, from best to worst. */
export const BANKNOTE_GRADES = ['UNC', 'AU', 'XF', 'VF', 'F', 'VG', 'G', 'P'] as const

/** A replacement (star) note or a specimen; '' is an ordinary note. */
export const NOTE_VARIANTS = ['', 'replacement', 'specimen'] as const
export type NoteVariant = (typeof NOTE_VARIANTS)[number]

/** Companies that grade and encapsulate banknotes. */
export const GRADING_COMPANIES = ['PMG', 'PCGS Banknote'] as const

/** The catalogue of world banknotes, as registered in the app. */
export const PICK_CATALOG = 'Pick'

export function banknoteMaterialLabel(material: string | undefined): string {
  switch (material) {
    case 'Paper': return t('banknoteForm.materials.paper')
    case 'Polymer': return t('banknoteForm.materials.polymer')
    case 'Hybrid': return t('banknoteForm.materials.hybrid')
    case 'Other': return t('banknoteForm.materials.other')
    default: return material ?? ''
  }
}

/** "UNC — Nova (não circulada)"; another grade as it is written. */
export function banknoteGradeLabel(grade: string | undefined): string {
  const known = BANKNOTE_GRADES.find((option) => option === grade)
  return known ? `${known} — ${t(`banknoteForm.grades.${known}`)}` : grade ?? ''
}

export function noteVariantLabel(variant: string | undefined): string {
  if (variant === 'replacement') return t('banknoteForm.variants.replacement')
  if (variant === 'specimen') return t('banknoteForm.variants.specimen')
  return t('banknoteForm.variants.normal')
}

/** Maps a Numista composition ("Papel", "Polymer", "Hybrid substrate") to a banknote material. */
export function banknoteMaterialFromComposition(composition: string | undefined): BanknoteMaterial {
  const text = (composition ?? '').toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '')
  if (/hybrid|hibrid/.test(text)) return 'Hybrid'
  if (/polym|polim|plastic/.test(text)) return 'Polymer'
  if (/paper|papel|cotton|algodao|linen|linho/.test(text)) return 'Paper'
  return 'Other'
}

/** A grading company's grade, in parts. */
export interface CertifiedGrade {
  company: string
  /** 1 to 70. */
  score: number | null
  /** Exceptional Paper Quality: the paper was never treated. */
  epq: boolean
  /** PMG's star for exceptional eye appeal. */
  star: boolean
}

/** "PMG 66 EPQ ★"; '' without a company or a score. */
export function composeCertifiedGrade(grade: CertifiedGrade): string {
  const company = grade.company.trim()
  if (!company || grade.score == null) return ''
  return [company, String(grade.score), grade.epq ? 'EPQ' : '', grade.star ? '★' : ''].filter(Boolean).join(' ')
}

/** Reads "PMG 66 EPQ ★" back into its parts. */
export function parseCertifiedGrade(text: string | undefined): CertifiedGrade {
  const value = (text ?? '').trim()
  const match = /^(.*?)\s+(\d{1,2})\b(.*)$/.exec(value)
  if (!match) return { company: value, score: null, epq: false, star: false }
  const rest = match[3]!.toUpperCase()
  return { company: match[1]!.trim(), score: Number(match[2]), epq: /\bEPQ\b/.test(rest), star: /★|\*|\bSTAR\b/.test(rest) }
}

/** The item's Pick number ("188a"), or ''. */
export function pickNumber(references: Pick<CoinReference, 'catalog' | 'number'>[] | undefined): string {
  return references?.find((reference) => reference.catalog.toLowerCase() === PICK_CATALOG.toLowerCase())?.number ?? ''
}

/** The references with the Pick number set to number (removed when ''). */
export function withPickNumber<T extends Pick<CoinReference, 'catalog' | 'number'>>(references: T[] | undefined, number: string): T[] {
  const others = (references ?? []).filter((reference) => reference.catalog.toLowerCase() !== PICK_CATALOG.toLowerCase())
  const value = number.trim().replace(/^p\s*#?\s*-?\s*/i, '')
  return value ? [...others, { catalog: PICK_CATALOG, number: value } as T] : others
}
