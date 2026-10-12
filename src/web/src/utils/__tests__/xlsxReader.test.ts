import { describe, expect, it } from 'vitest'
import { excelSerialDate, readXlsxRows } from '@/utils/xlsxReader'
import { buildXlsx, buildZip } from './xlsxFixture'

const MAIN = 'http://schemas.openxmlformats.org/spreadsheetml/2006/main'

describe('readXlsxRows', () => {
  it('reads shared strings, numbers and gaps from the first worksheet, deflated or stored', async () => {
    const rows = await readXlsxRows(await buildXlsx([
      ['País', 'Ano', '', 'Compra'],
      ['Alemanha', 2026, 'D', 46174],
      ['Portugal', 1986],
    ]))
    expect(rows).toEqual([
      ['País', 'Ano', '', 'Compra'],
      ['Alemanha', '2026', 'D', '46174'],
      ['Portugal', '1986'],
    ])
  })

  it('reads rich and inline strings and undoes Excel\'s _xHHHH_ escapes', async () => {
    const sheet = `<worksheet xmlns="${MAIN}"><sheetData>
      <row r="1"><c r="A1" t="s"><v>0</v></c><c r="C1" t="inlineStr"><is><t>inline</t></is></c><c r="D1" t="b"><v>1</v></c></row>
    </sheetData></worksheet>`
    const shared = `<sst xmlns="${MAIN}"><si><r><t>Jefferson </t></r><r><t>Nickel_x000D_</t></r><rPh><t>ignored</t></rPh></si></sst>`
    const rows = await readXlsxRows(await buildZip({ 'xl/worksheets/sheet1.xml': sheet, 'xl/sharedStrings.xml': shared }, ['xl/sharedStrings.xml']))
    expect(rows).toEqual([['Jefferson Nickel\r', '', 'inline', 'TRUE']])
  })

  it('refuses what is not a workbook', async () => {
    await expect(readXlsxRows(new TextEncoder().encode('País,Ano\nPortugal,1986').buffer)).rejects.toThrow('not a zip file')
  })
})

describe('excelSerialDate', () => {
  it('turns Excel day numbers into dates and ignores what cannot be a purchase date', () => {
    expect(excelSerialDate('46174')).toBe('2026-06-01')
    expect(excelSerialDate('45931')).toBe('2025-10-01')
    expect(excelSerialDate('0')).toBeUndefined()
    expect(excelSerialDate('abc')).toBeUndefined()
  })
})
