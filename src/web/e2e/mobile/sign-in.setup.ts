import { mkdirSync } from 'node:fs'
import { dirname } from 'node:path'
import { expect, test as setup } from '@playwright/test'
import { MOBILE_SESSION } from '../../playwright.mobile.config'

// Signs in once (the app limits sign-in attempts); the phones reuse the session.
setup('sign in', async ({ page }) => {
  const username = process.env.MOBILE_USERNAME ?? ''
  const password = process.env.MOBILE_PASSWORD ?? ''
  setup.skip(!username || !password, 'MOBILE_USERNAME and MOBILE_PASSWORD are not set')

  await page.goto('/login')
  await page.locator('input[autocomplete="username"]').fill(username)
  await page.locator('input[autocomplete="current-password"]').fill(password)
  await page.locator('form button[type="submit"]').click()
  await page.waitForURL((url) => !url.pathname.startsWith('/login'))
  // The first start's prompts are not under test: marked as seen for the
  // session (the getting-started one is kept per user).
  await page.evaluate(() => {
    const user = JSON.parse(localStorage.getItem('user') ?? '{}') as { id?: number }
    if (user.id) localStorage.setItem(`onboardingPromptSeen:${user.id}`, 'true')
    localStorage.setItem('pwa-install-dismissed', 'true')
  })
  await expect(page).not.toHaveURL(/\/login/)

  mkdirSync(dirname(MOBILE_SESSION), { recursive: true })
  await page.context().storageState({ path: MOBILE_SESSION })
})
