// A set shown as an album: a place for each coin, filled by a coin of the
// collection or waiting for one.
import type { ChecklistDefinition, ChecklistSlot, Coin, CoinSetCompletion } from '@/types'
import { t } from '@/i18n'
import { formatFaceValue, releaseCountryName } from '@/utils/coinReleases'

export interface AlbumSlot {
  key: string
  title: string
  /** Under the title: a face value, or the coin filling a challenge's slot. */
  detail?: string
  year?: number | null
  /** The country's name, to group by. */
  country?: string
  owned: boolean
  /** Wanted: a goal set's coin on the wishlist. */
  wishlist?: boolean
  /** The collection's coin: its page and photo (an authenticated media path). */
  coinId?: number
  coinImage?: string
  /** A picture of the coin to find, and the page that describes it. */
  imageUrl?: string
  pageUrl?: string
  /** A slot the page acts on itself (an agentic set's, to assign a coin). */
  targetId?: number
}

/** The photo a coin is known by: its obverse, else its primary, else its first. */
export function coinPhoto(coin: Pick<Coin, 'images'>): string | undefined {
  const images = coin.images ?? []
  return (images.find((image) => image.imageType === 'obverse') ?? images.find((image) => image.isPrimary) ?? images[0])?.filePath
}

/** "1 cêntimo", "50 cêntimos", "2 €". */
export function euroCoinName(value: number): string {
  return value < 1 ? t('sets.album.cents', { count: Math.round(value * 100) }) : formatFaceValue(value)
}

/** A checklist slot as the album shows it. */
export function checklistAlbumSlot(slot: ChecklistSlot, definition: ChecklistDefinition): AlbumSlot {
  const album: AlbumSlot = {
    key: slot.key,
    title: slot.title ?? '',
    year: slot.year || null,
    owned: slot.owned,
    coinId: slot.coinId,
    coinImage: slot.coinImage,
    imageUrl: slot.imageUrl,
    pageUrl: slot.pageUrl,
  }
  const country = slot.country ? releaseCountryName(slot.country) : ''
  if (definition.source === 'catalogue') {
    album.detail = (slot.faceValues ?? (slot.faceValue ? [slot.faceValue] : [])).map((value) => formatFaceValue(value, slot.currency)).join(' · ')
  } else if (definition.challenge === 'annual-series') {
    album.title = euroCoinName(slot.faceValue ?? 0)
    album.detail = slot.coinName
  } else {
    // A country's slot: the country (no grouping by it), and the coin filling it.
    album.title = country
    album.detail = definition.challenge === 'joint-issue' ? slot.coinName : (slot.coinName ?? t('sets.album.anyEuroCoin'))
  }
  return album
}

/** A coin of a standard, smart or goal set: a goal set's wishlist coins are still to find. */
export function coinAlbumSlot(coin: Coin, goal: boolean): AlbumSlot {
  const wishlist = goal && coin.isWishlist
  return {
    key: `coin:${coin.id}`,
    title: coin.name,
    detail: coin.denomination || undefined,
    year: coin.year ?? null,
    country: coin.country || undefined,
    owned: !wishlist,
    wishlist,
    coinId: coin.id,
    coinImage: coinPhoto(coin),
  }
}

/** An agentic set's slots: its targets, each filled by the coin assigned to it. */
export function agenticAlbumSlots(completion: CoinSetCompletion | null): AlbumSlot[] {
  const matches = completion?.targetMatches?.length
    ? completion.targetMatches
    : (completion?.targets ?? completion?.missingTargets ?? []).map((target) => ({ target, coin: null }))
  return matches.map(({ target, coin }) => ({
    key: `target:${target.id}`,
    title: coin?.name ?? target.label,
    detail: coin ? target.label : undefined,
    year: target.year ?? coin?.year ?? null,
    country: target.country ?? coin?.country ?? undefined,
    owned: Boolean(coin),
    coinId: coin?.id,
    coinImage: coin ? coinPhoto(coin) : undefined,
    targetId: target.id,
  }))
}

/** What a checklist set lists, in a line: "Catálogo · Portugal · 2 €". */
export function checklistDescription(definition: ChecklistDefinition): string {
  if (definition.source === 'catalogue') {
    const parts = [t('sets.album.fromCatalogue'), releaseCountryName(definition.country ?? '')]
    if (definition.currency === 'PTE') parts.push(t('sets.catalogue.currencies.PTE'))
    if (definition.faceValue) parts.push(formatFaceValue(definition.faceValue, definition.currency || undefined))
    return parts.join(' · ')
  }
  switch (definition.challenge) {
    case 'euro-countries': return t('sets.challenges.euroCountries.title')
    case 'joint-issue': return t('sets.challenges.jointIssue.description', { year: definition.year ?? '' })
    case 'annual-series': return t('sets.challenges.annualSeries.named', { country: releaseCountryName(definition.country ?? ''), year: definition.year ?? '' })
    default: return ''
  }
}
