// What a collection health checklist item asks for, in the app's language;
// the server's checklist keys ("metadata.grade") name them.
import { t } from '@/i18n'

const ITEM_MESSAGES = {
  'metadata.denomination': 'health.items.denomination',
  'metadata.country': 'health.items.country',
  'metadata.monetarySystem': 'health.items.monetarySystem',
  'metadata.mint': 'health.items.mint',
  'metadata.category': 'health.items.category',
  'metadata.material': 'health.items.material',
  'metadata.grade': 'health.items.grade',
  'images.any': 'health.items.anyImage',
  'images.coverage': 'health.items.bothImages',
  'valuation.value': 'health.items.value',
  'valuation.freshness': 'health.items.recentValue',
  'ai.analysis': 'health.items.aiAnalysis',
  'ai.coverage': 'health.items.aiCoverage',
} as const

/** The item's label; an item the app does not know keeps the server's label, or its key. */
export function healthItemLabel(key: string, fallback = ''): string {
  const message = ITEM_MESSAGES[key as keyof typeof ITEM_MESSAGES]
  return message ? t(message) : fallback || key
}
