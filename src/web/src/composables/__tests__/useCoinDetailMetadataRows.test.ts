import { afterEach, describe, expect, it } from 'vitest'
import { useCoinDetailMetadataRows } from '../useCoinDetailMetadataRows'
import { setLocale } from '@/i18n'
import type { Coin } from '@/types'

const coin = { id: 1, name: '2 euro 2024', country: 'Portugal', storageLocation: { id: 3, name: 'Gaveta 1' } } as unknown as Coin

function circulation(status: Parameters<typeof useCoinDetailMetadataRows>[1]) {
  return useCoinDetailMetadataRows(coin, status).rows.value.find((row) => row.key === 'circulation')?.value
}

describe('useCoinDetailMetadataRows', () => {
  afterEach(() => setLocale('en'))

  it('leaves out where the coin is stored', () => {
    const rows = useCoinDetailMetadataRows(coin).rows.value
    expect(rows.map((row) => row.key)).not.toContain('storageLocation')
    expect(rows.map((row) => row.value)).not.toContain('Gaveta 1')
  })

  it('tells whether the coin is current and whether it is commemorative', () => {
    setLocale('pt-PT')
    expect(circulation({ current: true, commemorative: true, commemoratedTopic: 'Jogos Olímpicos' })).toBe('Corrente · Comemorativa')
    expect(circulation({ current: true, commemorative: false })).toBe('Corrente')
    expect(circulation({ current: false, demonetizationDate: '2002-02-28', commemorative: false })).toBe('Desmonetizada em 28/02/2002')
    expect(circulation({ current: false, demonetizationDate: '2002', commemorative: true })).toBe('Desmonetizada em 2002 · Comemorativa')
    expect(circulation({ current: false, commemorative: null })).toBe('Desmonetizada')
    // A collector coin (Numista: non-circulating) is legal tender, but not "current".
    expect(circulation({ current: true, nonCirculating: true, commemorative: true })).toBe('Não circulante · Comemorativa')
    expect(circulation({ current: false, nonCirculating: true, demonetizationDate: '2002-02-28', commemorative: false })).toBe('Não circulante · Desmonetizada em 28/02/2002')
  })

  it('shows no row when nothing is known', () => {
    expect(circulation({ current: null, commemorative: null })).toBeUndefined()
    expect(circulation(null)).toBeUndefined()
  })

  it("shows a banknote's own details instead of a coin's", () => {
    setLocale('pt-PT')
    const banknote = {
      id: 2, name: '1000 escudos', kind: 'banknote', country: 'Portugal', material: 'Paper', grade: 'XF', mint: 'Lisboa', weightGrams: 1, diameterMm: 160,
      series: 'Chapa 13', serialNumber: 'CH 1234567', signatures: 'A / B', noteVariant: 'replacement', printer: 'Valora', widthMm: 160, heightMm: 75,
      watermark: 'Cabral', certifiedGrade: 'PMG 66 EPQ', certificationNumber: '8076543-012', references: [{ catalog: 'Pick', number: '188c' }],
    } as unknown as Coin
    const rows = Object.fromEntries(useCoinDetailMetadataRows(banknote).rows.value.map((row) => [row.key, row.value]))
    expect(rows).toMatchObject({
      series: 'Chapa 13', pick: 'P188c', serialNumber: 'CH 1234567', signatures: 'A / B', noteVariant: 'Nota de substituição (estrela)',
      printer: 'Valora', dimensions: '160 × 75 mm', watermark: 'Cabral', certifiedGrade: 'PMG 66 EPQ', certificationNumber: '8076543-012',
      material: 'Papel', grade: 'XF — Soberba',
    })
    expect(rows).not.toHaveProperty('mint')
    expect(rows).not.toHaveProperty('weight')
    expect(rows).not.toHaveProperty('diameter')
  })
})
