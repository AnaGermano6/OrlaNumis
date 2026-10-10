// Short texts translated on the server (MyMemory, free), e.g. a coin's
// subject in Portuguese to compare with Numista's English titles.
import { api } from '@/api/http'

export interface TranslationResult {
  status: 'ok' | 'unavailable'
  text?: string
}

export type TranslationLanguage = 'pt' | 'en' | 'fr' | 'es' | 'de' | 'it' | 'nl'

export const translateText = (q: string, from: TranslationLanguage, to: TranslationLanguage = 'en') =>
  api.get<TranslationResult>('/translate', { params: { q, from, to } })
