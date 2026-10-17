// Waiting for the server to restart (after a backup is restored), then
// opening the page again.
import { API_BASE } from '@/api/http'

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

async function serverUp(): Promise<boolean> {
  try {
    return (await fetch(`${API_BASE}/healthz`, { cache: 'no-store' })).ok
  } catch {
    return false
  }
}

/** Resolves once the server has gone down and is up again (or after 20 s up, or 2 minutes in all). */
export async function waitForRestart(): Promise<void> {
  const started = Date.now()
  let wentDown = false
  while (Date.now() - started < 120_000) {
    await sleep(2000)
    if (!(await serverUp())) wentDown = true
    else if (wentDown || Date.now() - started > 20_000) return
  }
}

export function reloadPage(): void {
  window.location.reload()
}
