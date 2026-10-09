import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { getCoinOptions } from '@/api/client'
import { setLocale } from '@/i18n'
import {
  builtInCoinOptionLabel,
  categoryLabel,
  coinVocabulary,
  eraLabel,
  loadCoinVocabulary,
  resetCoinVocabulary,
} from '@/i18n/coinVocabulary'
import { COIN_ERAS } from '@/types'
import { formatOptionLabelMap, parseOptionLabelMap } from '@/utils/options'

vi.mock('@/api/client', () => ({
  getCoinOptions: vi.fn(),
}))

type CoinOptionsResult = Awaited<ReturnType<typeof getCoinOptions>>

function mockOptions(data: CoinOptionsResult['data']) {
  vi.mocked(getCoinOptions).mockResolvedValue({ data } as CoinOptionsResult)
}

describe('coin vocabulary', () => {
  beforeEach(() => {
    vi.mocked(getCoinOptions).mockReset()
    resetCoinVocabulary()
  })

  afterEach(() => {
    setLocale('en')
  })

  it('shows stored values unchanged in English', () => {
    setLocale('en')
    expect(categoryLabel('Roman')).toBe('Roman')
    expect(eraLabel('ancient')).toBe('ancient')
  })

  it('translates built-in and common values in Portuguese, following the stored first-letter case', () => {
    setLocale('pt-PT')
    expect(categoryLabel('Roman')).toBe('Romana')
    expect(categoryLabel('Other')).toBe('Outra')
    expect(categoryLabel('Roman Provincial')).toBe('Romana provincial')
    expect(eraLabel('ancient')).toBe('antiga')
    expect(eraLabel('Medieval')).toBe('Medieval')
    expect(categoryLabel('Unlisted Mint')).toBe('Unlisted Mint')
    expect(categoryLabel(undefined)).toBe('')
    expect(builtInCoinOptionLabel('Celtic')).toBe('Celta')
    expect(builtInCoinOptionLabel('Unlisted Mint')).toBeUndefined()
  })

  it('prefers the admin-defined names, matched per list and case-insensitively', async () => {
    mockOptions({
      categories: ['Roman', 'Celtic coins'],
      eras: ['ancient', 'Republican'],
      labels: {
        'pt-PT': {
          categories: { Roman: 'Romanas', 'Celtic coins': 'Moedas celtas' },
          eras: { republican: 'República' },
        },
      },
    })
    await loadCoinVocabulary()
    setLocale('pt-PT')

    expect(coinVocabulary.categories).toEqual(['Roman', 'Celtic coins'])
    expect(coinVocabulary.eras).toEqual(['ancient', 'Republican'])
    expect(categoryLabel('Roman')).toBe('Romanas')
    expect(categoryLabel('Celtic coins')).toBe('Moedas celtas')
    expect(eraLabel('Republican')).toBe('República')
    // An era name does not leak into categories.
    expect(categoryLabel('Republican')).toBe('Republicana')
    expect(eraLabel('ancient')).toBe('antiga')

    setLocale('en')
    expect(categoryLabel('Roman')).toBe('Roman')
  })

  it('keeps the defaults when the options cannot be loaded', async () => {
    vi.mocked(getCoinOptions).mockRejectedValue(new Error('offline'))
    await expect(loadCoinVocabulary()).resolves.toBeUndefined()
    expect(coinVocabulary.eras).toEqual(COIN_ERAS)
    expect(coinVocabulary.loaded).toBe(false)
  })

  it('shares one request between concurrent loads', async () => {
    mockOptions({ categories: [], eras: [], labels: {} })
    await Promise.all([loadCoinVocabulary(), loadCoinVocabulary()])
    expect(getCoinOptions).toHaveBeenCalledTimes(1)
  })
})

describe('option label maps', () => {
  it('parses Value=Name lines like the API does', () => {
    expect(parseOptionLabelMap(' Roman = Romana \n\nno separator\n=Orphan\nBlank=  \nA=B=C\nRoman=Romanas')).toEqual({
      Roman: 'Romanas',
      A: 'B=C',
    })
    expect(parseOptionLabelMap(undefined)).toEqual({})
  })

  it('formats only configured options with a name, in option order', () => {
    expect(formatOptionLabelMap(['Greek', 'Roman', 'Other'], { Roman: ' Romana ', Greek: 'Grega', Gone: 'X', Other: ' ' }))
      .toBe('Greek=Grega\nRoman=Romana')
  })
})
