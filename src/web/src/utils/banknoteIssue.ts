// Which Numista issue a banknote is: told apart by its printed date and its
// signatories, it gives the note its signatures, its own Pick number ("176a")
// and whether it is a replacement or specimen note.
import type { NumistaIssueSummary } from '@/api/client'
import type { Coin, CoinReference } from '@/types'
import { withPickNumber } from '@/utils/banknoteAttributes'
import { numistaIssuePrintedDate } from '@/utils/numistaCatalog'

/** The signatories of an issue as a banknote stores them: "José da Silva Lopes / Abel Reis". */
export function issueSignatures(issue: Pick<NumistaIssueSummary, 'signatures'>): string {
  const names = (issue.signatures ?? []).map((signature) => signature.name.trim()).filter(Boolean).join(' / ')
  return names.length > 300 ? `${names.slice(0, 299)}…` : names
}

/** The Numista type an item is linked to, from its references. */
export function numistaTypeIdOf(references: Pick<CoinReference, 'catalog' | 'number'>[] | undefined): number | undefined {
  const number = Number(references?.find((reference) => reference.catalog.trim().toLowerCase() === 'numista')?.number.trim())
  return Number.isInteger(number) && number > 0 ? number : undefined
}

/** Fills in a banknote from the Numista issue it is; what the issue does not say is left as it was. */
export function applyNumistaIssue(form: Partial<Coin>, issue: NumistaIssueSummary): void {
  form.numistaIssueId = issue.id
  const signatures = issueSignatures(issue)
  if (signatures) form.signatures = signatures
  if (issue.pick) form.references = withPickNumber(form.references, issue.pick)
  form.noteVariant = issue.variant ?? ''
  if (issue.year !== undefined) {
    form.year = issue.year
    form.dateRange = numistaIssuePrintedDate(issue)
  }
}
