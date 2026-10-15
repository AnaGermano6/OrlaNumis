import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import en from '@/i18n/locales/en'

// Every message the app ships is one some screen shows: a message no source
// file names (outside the locales and the tests) is left over from a removed
// feature, and goes with it.

const src = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..')

type Tree = { [key: string]: unknown }

// A plural message ({ one, other }) is one key.
function isPlural(value: unknown): boolean {
  return !!value && typeof value === 'object' && Object.keys(value).sort().join(',') === 'one,other'
}

function keys(tree: Tree, prefix = ''): string[] {
  return Object.entries(tree).flatMap(([key, value]) => {
    const name = prefix ? `${prefix}.${key}` : key
    return typeof value === 'string' || isPlural(value) ? [name] : keys(value as Tree, name)
  })
}

function sources(dir: string): string[] {
  return fs.readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const file = path.join(dir, entry.name)
    if (entry.isDirectory()) return entry.name === '__tests__' || file.endsWith(path.join('i18n', 'locales')) ? [] : sources(file)
    return /\.(ts|vue)$/.test(entry.name) ? [fs.readFileSync(file, 'utf8')] : []
  })
}

describe('messages', () => {
  it('are all used somewhere', () => {
    const text = sources(src).join('\n')
    // Keys built at run time name their group: t(`nav.${item}`).
    const prefixes = [...text.matchAll(/`([A-Za-z0-9_.]+)\.\$\{/g)].map((match) => `${match[1]}.`)
    const unused = keys(en as unknown as Tree).filter((key) => !text.includes(key) && !prefixes.some((prefix) => key.startsWith(prefix)))
    expect(unused).toEqual([])
  })
})
