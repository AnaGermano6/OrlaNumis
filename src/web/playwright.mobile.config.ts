import { defineConfig, devices } from '@playwright/test'

// Phone-size checks of a running Orla Numis: no page wider than the screen,
// and the gallery's filters reachable to their end and closed again. Needs an
// app (MOBILE_BASE_URL, e.g. http://localhost:18080) and an account of its
// own on it (MOBILE_USERNAME, MOBILE_PASSWORD); never a real collection's.
// It signs in once and the phones reuse that session.
// Run with: npx playwright test -c playwright.mobile.config.ts
export const MOBILE_SESSION = './.artifacts/mobile/session.json'

export default defineConfig({
  testDir: './e2e/mobile',
  outputDir: './.artifacts/mobile/playwright-output',
  workers: 1,
  fullyParallel: false,
  reporter: [['list']],
  use: {
    baseURL: process.env.MOBILE_BASE_URL,
    trace: 'off',
    video: 'off',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'sign-in', testMatch: /sign-in\.setup\.ts/, use: { browserName: 'chromium' } },
    // The narrowest phone in use, and a common one.
    {
      name: 'phone-320',
      dependencies: ['sign-in'],
      use: { ...devices['iPhone SE'], browserName: 'chromium', viewport: { width: 320, height: 568 }, storageState: MOBILE_SESSION },
    },
    {
      name: 'phone-375',
      dependencies: ['sign-in'],
      use: { ...devices['iPhone 13 Mini'], browserName: 'chromium', storageState: MOBILE_SESSION },
    },
  ],
})
