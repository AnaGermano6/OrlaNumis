// Turns a banknote type picked in the Numista catalogue into Add Banknote
// form values: the catalogue's facts (issuing bank, series, printer, size,
// watermark, Pick number…), not what belongs to the user's own note (its
// serial number, price, date).
import { t } from '@/i18n'
import type { CatalogRegistry, Coin, CoinReferenceInput } from '@/types'
import { formatNumistaYear, numistaYearRange } from '@/utils/numistaCatalog'
import { numistaGradeShort } from '@/utils/numistaGrades'
import { categoryForType, eraForYear, type CatalogPrefill, type CatalogSelection } from '@/utils/numistaCatalogPrefill'
import { PICK_CATALOG, banknoteMaterialFromComposition } from '@/utils/banknoteAttributes'
import { rarityFromMintage } from '@/utils/coinAttributes'

export interface BanknotePrefillOptions {
  categories: string[]
  eras: string[]
  /** Catalogues configured in Admin; other references go into the notes. */
  registry: CatalogRegistry[]
}

function clip(value: string, max: number): string {
  return value.length > max ? `${value.slice(0, max - 1)}…` : value
}

function lettering(value: string | undefined): string {
  return (value ?? '').split('\n').map((line) => line.trim()).filter(Boolean).join(' / ')
}

// Numista writes Pick numbers under "P".
function registeredCatalog(code: string, registry: CatalogRegistry[]): CatalogRegistry | undefined {
  const wanted = /^p(ick)?$/i.test(code.trim()) ? PICK_CATALOG : code.trim()
  return registry.find((entry) => entry.catalog.toLowerCase() === wanted.toLowerCase())
}

export function buildBanknotePrefill(selection: CatalogSelection, options: BanknotePrefillOptions): CatalogPrefill {
  const { type, issue } = selection
  const issueYear = issue?.year
  const year = issueYear ?? type.maxYear ?? type.minYear
  const dateRange = issueYear !== undefined
    ? formatNumistaYear(issueYear)
    : issue && (issue.minYear !== undefined || issue.maxYear !== undefined)
      ? numistaYearRange(issue.minYear, issue.maxYear)
      : numistaYearRange(type.minYear, type.maxYear)

  const references: CoinReferenceInput[] = [{ catalog: 'Numista', number: String(type.id), uri: type.url }]
  const otherReferences: string[] = []
  for (const reference of type.references) {
    const registered = registeredCatalog(reference.catalogue, options.registry)
    if (registered && !registered.volumeRequired && registered.catalog.toLowerCase() !== 'numista') {
      references.push({ catalog: registered.catalog, number: reference.number })
    } else {
      otherReferences.push(`${reference.catalogue} ${reference.number}`)
    }
  }

  const notes = [
    t('catalog.notes.source', { title: type.title, issuer: type.issuer.name, url: type.url }),
    type.composition ? t('catalog.notes.composition', { value: type.composition }) : '',
    otherReferences.length ? t('catalog.notes.otherReferences', { value: otherReferences.join(', ') }) : '',
  ].filter(Boolean).join('\n')

  const nameParts = [type.title, issueYear !== undefined ? formatNumistaYear(issueYear) : ''].filter(Boolean).join(' ')
  const fields: Partial<Coin> = {
    kind: 'banknote',
    name: clip(type.issuer.name ? `${nameParts} – ${type.issuer.name}` : nameParts, 200),
    category: categoryForType(type, options.categories) as Coin['category'],
    era: eraForYear(year, options.eras),
    denomination: clip(type.denomination ?? '', 200),
    // The bank that issued the note (the coin's "ruler" field).
    ruler: clip(type.issuingEntity || type.rulers.join(', '), 200),
    series: clip(type.series ?? '', 200),
    // A starting point from the number printed; the user can change it.
    rarity: rarityFromMintage(issue?.mintage) ?? '',
    dateRange: clip(dateRange, 200),
    year: issueYear ?? (type.minYear !== undefined && type.minYear === type.maxYear ? type.minYear : null),
    country: clip(type.issuer.name, 120),
    material: banknoteMaterialFromComposition(type.composition) as Coin['material'],
    // Numista gives a banknote's width as its size.
    widthMm: type.diameterMm ?? null,
    heightMm: type.heightMm ?? null,
    printer: clip((type.printers ?? []).join(', '), 200),
    watermark: clip(type.watermark ?? '', 500),
    grade: selection.grade ? numistaGradeShort(selection.grade) : '',
    obverseInscription: clip(lettering(type.obverse.lettering), 1000),
    reverseInscription: clip(lettering(type.reverse.lettering), 1000),
    obverseDescription: clip(type.obverse.description ?? '', 2000),
    reverseDescription: clip(type.reverse.description ?? '', 2000),
    currentValue: selection.currentValue === undefined ? undefined : Math.round(selection.currentValue * 100) / 100,
    notes: clip(notes, 5000),
    isWishlist: selection.wishlist,
  }
  return { fields, references }
}
