import type en from './locales/en'

/** A message whose wording depends on a `count` param (Intl.PluralRules categories). */
export interface PluralMessage {
  one: string
  other: string
}

export type MessageParams = Record<string, string | number>

/** The English catalog is the source of truth; every other locale must match its shape. */
export type Messages = typeof en

type KeyPaths<T> = {
  [K in keyof T & string]: T[K] extends string
    ? K
    : T[K] extends PluralMessage
      ? K
      : `${K}.${KeyPaths<T[K]>}`
}[keyof T & string]

/** Every valid dotted message key, e.g. `nav.collection`. Unknown keys fail type-check. */
export type MessageKey = KeyPaths<Messages>
