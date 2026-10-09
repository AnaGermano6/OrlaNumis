import { proxyImage } from '@/api/client'

// Blob URLs of photos fetched through the image proxy, shared for the session
// so search results, previews and coin pages reuse them.
const blobUrls = new Map<string, Promise<string | null>>()

// Photo sources (Wikimedia in particular) rate-limit bursts, so only a couple
// of proxied photos load at a time.
const MAX_CONCURRENT = 2
let active = 0
const waiting: Array<() => void> = []

async function withSlot<T>(task: () => Promise<T>): Promise<T> {
  if (active >= MAX_CONCURRENT) await new Promise<void>((resolve) => waiting.push(resolve))
  active += 1
  try {
    return await task()
  } finally {
    active -= 1
    waiting.shift()?.()
  }
}

/** Loads an external https image through the app's proxy; null when it fails. */
export function loadCatalogPhoto(url: string): Promise<string | null> {
  let pending = blobUrls.get(url)
  if (!pending) {
    pending = withSlot(() => proxyImage(url))
      .then((res) => URL.createObjectURL(res.data as Blob))
      .catch(() => {
        blobUrls.delete(url)
        return null
      })
    blobUrls.set(url, pending)
  }
  return pending
}
