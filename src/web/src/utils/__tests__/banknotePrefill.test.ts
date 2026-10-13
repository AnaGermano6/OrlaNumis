import { beforeEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import type { NumistaCatalogType } from '@/api/client'
import type { CatalogRegistry } from '@/types'
import { buildBanknotePrefill } from '@/utils/banknotePrefill'

const side = { description: '', lettering: '' }
const TYPE: NumistaCatalogType = {
  id: 1234,
  url: 'https://pt.numista.com/1234',
  title: '1000 Escudos (Pedro Álvares Cabral)',
  category: 'banknote',
  issuer: { code: 'portugal', name: 'Portugal' },
  issuingEntity: 'Banco de Portugal',
  minYear: 1996,
  maxYear: 2000,
  denomination: '1000 Escudos',
  rulers: ['Terceira República'],
  composition: 'Papel',
  diameterMm: 160,
  heightMm: 75,
  printers: ['Valora'],
  watermark: 'Pedro Álvares Cabral',
  series: 'Chapa 13',
  mints: [],
  obverse: { description: 'Pedro Álvares Cabral', lettering: 'BANCO DE PORTUGAL\nMIL ESCUDOS' },
  reverse: side,
  edge: side,
  references: [{ catalogue: 'P', number: '188c' }, { catalogue: 'XYZ', number: '9' }],
  photos: [],
}

const REGISTRY = [
  { catalog: 'Numista', volumeRequired: false },
  { catalog: 'Pick', volumeRequired: false },
] as CatalogRegistry[]

describe('buildBanknotePrefill', () => {
  beforeEach(() => setLocale('pt-PT'))

  it('fills a banknote from its Numista type', () => {
    const { fields, references } = buildBanknotePrefill({ type: TYPE, wishlist: false }, { categories: ['Modern', 'Other'], eras: ['modern'], registry: REGISTRY })

    expect(fields).toMatchObject({
      kind: 'banknote',
      name: '1000 Escudos (Pedro Álvares Cabral) – Portugal',
      category: 'Modern',
      era: 'modern',
      denomination: '1000 Escudos',
      ruler: 'Banco de Portugal',
      series: 'Chapa 13',
      dateRange: '1996–2000',
      material: 'Paper',
      widthMm: 160,
      heightMm: 75,
      printer: 'Valora',
      watermark: 'Pedro Álvares Cabral',
      obverseInscription: 'BANCO DE PORTUGAL / MIL ESCUDOS',
    })
    // Numista's "P" is the Pick catalogue; an unknown catalogue goes to the notes.
    expect(references).toEqual([
      { catalog: 'Numista', number: '1234', uri: 'https://pt.numista.com/1234' },
      { catalog: 'Pick', number: '188c' },
    ])
    expect(fields.notes).toContain('XYZ 9')
  })
})
