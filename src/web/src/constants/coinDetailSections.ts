// T001: Coin detail section constants for #219
// These constants define the standard section links shown on the coin detail overview page.
// title/description are getters so they follow the current UI locale wherever they are read.
import { t } from '@/i18n'

export interface CoinDetailSection {
  id: 'shipment' | 'journal' | 'notes' | 'actions' | 'analysis' | 'health' | 'valuation'
  title: string
  description: string
  route: (coinId: number) => string
}

export const COIN_DETAIL_SECTIONS: Record<string, CoinDetailSection> = {
  shipment: {
    id: 'shipment',
    get title() { return t('coinDetail.sections.shipment.title') },
    get description() { return t('coinDetail.sections.shipment.description') },
    route: (coinId: number) => `/coin/${coinId}/shipment`,
  },
  journal: {
    id: 'journal',
    get title() { return t('coinDetail.sections.journal.title') },
    get description() { return t('coinDetail.sections.journal.description') },
    route: (coinId: number) => `/coin/${coinId}/journal`,
  },
  health: {
    id: 'health',
    get title() { return t('coinDetail.sections.health.title') },
    get description() { return t('coinDetail.sections.health.description') },
    route: (coinId: number) => `/coin/${coinId}/health`,
  },
  notes: {
    id: 'notes',
    get title() { return t('coinDetail.sections.notes.title') },
    get description() { return t('coinDetail.sections.notes.description') },
    route: (coinId: number) => `/coin/${coinId}/notes`,
  },
  actions: {
    id: 'actions',
    get title() { return t('coinDetail.sections.actions.title') },
    get description() { return t('coinDetail.sections.actions.description') },
    route: (coinId: number) => `/coin/${coinId}/actions`,
  },
  analysis: {
    id: 'analysis',
    get title() { return t('coinDetail.sections.analysis.title') },
    get description() { return t('coinDetail.sections.analysis.description') },
    route: (coinId: number) => `/coin/${coinId}/analysis`,
  },
  valuation: {
    id: 'valuation',
    get title() { return t('coinDetail.sections.valuation.title') },
    get description() { return t('coinDetail.sections.valuation.description') },
    route: (coinId: number) => `/coin/${coinId}/valuation`,
  },
}

export const SECTION_ORDER: CoinDetailSection['id'][] = ['shipment', 'journal', 'health', 'notes', 'actions', 'analysis', 'valuation']
