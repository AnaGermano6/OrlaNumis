// Reads the first worksheet of an .xlsx file into rows of cell texts. An
// .xlsx is a zip of XML parts: the zip is read here and its parts inflated
// with the browser's own DecompressionStream, so no library is needed. Only
// what a data export uses is supported: shared and inline strings, numbers
// (as written, e.g. dates stay Excel serial numbers) and booleans.

interface ZipEntry {
  method: number
  compressedSize: number
  localHeader: number
}

const END_OF_CENTRAL_DIRECTORY = 0x06054b50
const CENTRAL_FILE_HEADER = 0x02014b50
const LOCAL_FILE_HEADER = 0x04034b50

export class XlsxError extends Error {}

function zipEntries(bytes: Uint8Array): Map<string, ZipEntry> {
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength)
  let end = -1
  // The directory's end record is last, after a comment of up to 64 KiB.
  for (let at = bytes.length - 22; at >= Math.max(0, bytes.length - 22 - 0xffff); at -= 1) {
    if (view.getUint32(at, true) === END_OF_CENTRAL_DIRECTORY) {
      end = at
      break
    }
  }
  if (end < 0) throw new XlsxError('not a zip file')
  const count = view.getUint16(end + 10, true)
  let at = view.getUint32(end + 16, true)
  const names = new TextDecoder()
  const entries = new Map<string, ZipEntry>()
  for (let index = 0; index < count; index += 1) {
    if (at + 46 > bytes.length || view.getUint32(at, true) !== CENTRAL_FILE_HEADER) throw new XlsxError('broken zip directory')
    const nameLength = view.getUint16(at + 28, true)
    const name = names.decode(bytes.subarray(at + 46, at + 46 + nameLength))
    entries.set(name, {
      method: view.getUint16(at + 10, true),
      compressedSize: view.getUint32(at + 20, true),
      localHeader: view.getUint32(at + 42, true),
    })
    at += 46 + nameLength + view.getUint16(at + 30, true) + view.getUint16(at + 32, true)
  }
  return entries
}

async function inflate(data: Uint8Array): Promise<Uint8Array> {
  const stream = new ReadableStream<Uint8Array>({
    start(controller) {
      controller.enqueue(data)
      controller.close()
    },
  }).pipeThrough(new DecompressionStream('deflate-raw') as unknown as ReadableWritablePair<Uint8Array, Uint8Array>)
  const reader = stream.getReader()
  const chunks: Uint8Array[] = []
  let length = 0
  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    chunks.push(value)
    length += value.length
  }
  const out = new Uint8Array(length)
  let offset = 0
  for (const chunk of chunks) {
    out.set(chunk, offset)
    offset += chunk.length
  }
  return out
}

async function readPart(bytes: Uint8Array, entries: Map<string, ZipEntry>, name: string): Promise<string | undefined> {
  const entry = entries.get(name)
  if (!entry) return undefined
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength)
  if (view.getUint32(entry.localHeader, true) !== LOCAL_FILE_HEADER) throw new XlsxError(`broken zip entry ${name}`)
  const start = entry.localHeader + 30 + view.getUint16(entry.localHeader + 26, true) + view.getUint16(entry.localHeader + 28, true)
  const data = bytes.subarray(start, start + entry.compressedSize)
  if (entry.method === 0) return new TextDecoder().decode(data)
  if (entry.method === 8) return new TextDecoder().decode(await inflate(data))
  throw new XlsxError(`unsupported compression in ${name}`)
}

function parseXml(text: string): Document {
  const doc = new DOMParser().parseFromString(text, 'application/xml')
  if (doc.getElementsByTagName('parsererror').length) throw new XlsxError('broken XML')
  return doc
}

function elements(parent: Document | Element, localName: string): Element[] {
  return Array.from(parent.getElementsByTagNameNS('*', localName))
}

/** Excel writes some characters as _xHHHH_ ("_x000D_" is a carriage return). */
function unescapeXlsx(value: string): string {
  return value.replace(/_x([0-9A-Fa-f]{4})_/g, (_, code: string) => String.fromCharCode(Number.parseInt(code, 16)))
}

function text(parent: Element): string {
  // Rich text is several <t> runs; phonetic hints (<rPh>) are not part of it.
  return unescapeXlsx(elements(parent, 't').filter((node) => node.parentElement?.localName !== 'rPh').map((node) => node.textContent ?? '').join(''))
}

/** "B12" → 1 (column B). */
function columnIndex(reference: string): number {
  const letters = /^[A-Z]+/.exec(reference)?.[0] ?? ''
  let index = 0
  for (const letter of letters) index = index * 26 + letter.charCodeAt(0) - 64
  return index - 1
}

/** The path of the workbook's first worksheet. */
async function firstSheetPath(bytes: Uint8Array, entries: Map<string, ZipEntry>): Promise<string> {
  const workbook = await readPart(bytes, entries, 'xl/workbook.xml')
  const rels = await readPart(bytes, entries, 'xl/_rels/workbook.xml.rels')
  if (workbook && rels) {
    const sheet = elements(parseXml(workbook), 'sheet')[0]
    const id = sheet && Array.from(sheet.attributes).find((attribute) => attribute.localName === 'id')?.value
    const target = id && elements(parseXml(rels), 'Relationship').find((relation) => relation.getAttribute('Id') === id)?.getAttribute('Target')
    if (target) return target.startsWith('/') ? target.slice(1) : `xl/${target}`
  }
  return 'xl/worksheets/sheet1.xml'
}

/** The first worksheet's rows, each a list of cell texts ('' for empty cells). */
export async function readXlsxRows(buffer: ArrayBuffer): Promise<string[][]> {
  const bytes = new Uint8Array(buffer)
  const entries = zipEntries(bytes)
  const sharedXml = await readPart(bytes, entries, 'xl/sharedStrings.xml')
  const shared = sharedXml ? elements(parseXml(sharedXml), 'si').map(text) : []
  const sheetXml = await readPart(bytes, entries, await firstSheetPath(bytes, entries))
  if (!sheetXml) throw new XlsxError('no worksheet')

  const rows: string[][] = []
  for (const row of elements(parseXml(sheetXml), 'row')) {
    const cells: string[] = []
    for (const cell of elements(row, 'c')) {
      const reference = cell.getAttribute('r')
      const index = reference ? columnIndex(reference) : cells.length
      const type = cell.getAttribute('t')
      const value = elements(cell, 'v')[0]?.textContent ?? ''
      let content: string
      if (type === 's') content = shared[Number(value)] ?? ''
      else if (type === 'inlineStr') content = elements(cell, 'is')[0] ? text(elements(cell, 'is')[0]!) : ''
      else if (type === 'b') content = value === '1' ? 'TRUE' : 'FALSE'
      else content = value
      while (cells.length < index) cells.push('')
      cells[index] = content
    }
    rows.push(cells)
  }
  return rows
}

/** An Excel date serial number (days since 1899-12-30) as YYYY-MM-DD, or undefined. */
export function excelSerialDate(value: string): string | undefined {
  const serial = Number(value)
  // 1 = 1900-01-01; anything before 1950 or after 2100 is not a purchase date.
  if (!Number.isFinite(serial) || serial < 18264 || serial > 73051) return undefined
  const date = new Date(Date.UTC(1899, 11, 30) + Math.floor(serial) * 86_400_000)
  return date.toISOString().slice(0, 10)
}
