import { afterEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/i18n'
import { MATERIALS } from '@/types'
import { materialClass, materialLabel } from '@/utils/format'

describe('coin materials', () => {
  afterEach(() => setLocale('en'))

  it('names every material in Portuguese', () => {
    setLocale('pt-PT')
    expect(MATERIALS.map(materialLabel)).toEqual([
      'Ouro', 'Prata', 'Bilhão', 'Eletro', 'Cobre', 'Bronze', 'Bronze-alumínio', 'Latão', 'Latão de níquel',
      'Ouro nórdico', 'Cuproníquel', 'Níquel', 'Bimetálica', 'Alumínio', 'Zinco', 'Aço inoxidável', 'Aço revestido', 'Outro',
    ])
  })

  it('names every material in English', () => {
    for (const material of MATERIALS) expect(materialLabel(material)).toBe(material)
  })

  it('gives each material one CSS class, without spaces', () => {
    expect(materialClass('Gold')).toBe('material-gold')
    expect(materialClass('Nickel brass')).toBe('material-nickel-brass')
    expect(materialClass('Copper-nickel')).toBe('material-copper-nickel')
    expect(materialClass('Stainless steel')).toBe('material-stainless-steel')
    for (const material of MATERIALS) expect(materialClass(material)).not.toMatch(/\s/)
  })
})
