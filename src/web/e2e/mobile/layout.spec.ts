import { expect, test, type Page } from '@playwright/test'

// See playwright.mobile.config.ts: a running app, and the session signed in
// by sign-in.setup.ts.
const PAGES = ['/collection', '/', '/banknotes', '/wishlist', '/sets', '/stats', '/settings', '/add', '/lookup']

/** How far the page reaches past the screen's right edge, in pixels. */
function overflow(page: Page) {
  return page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth)
}

test.describe('on a phone', () => {
  for (const path of PAGES) {
    test(`${path} is no wider than the screen`, async ({ page }) => {
      await page.goto(path)
      await page.waitForLoadState('networkidle')
      // Signed in, on a phone's screen: not a pass for the wrong reason.
      await expect(page).not.toHaveURL(/\/login/)
      expect(await page.evaluate(() => document.documentElement.clientWidth)).toBeLessThanOrEqual(430)
      expect(await overflow(page)).toBeLessThanOrEqual(0)
    })
  }

  test('the gallery opens its filters in a sheet that scrolls to its end and closes', async ({ page }) => {
    await page.goto('/')
    await page.getByTestId('gallery-filters-button').click()

    const sheet = page.getByTestId('gallery-filters')
    await expect(sheet).toBeVisible()
    // The search bar stays where it sticks while the sheet is open.
    await expect(page.getByTestId('gallery-filters-button')).toBeInViewport()
    // The page behind holds still while the sheet is open.
    expect(await page.evaluate(() => document.documentElement.style.overflow)).toBe('hidden')

    // The last option can be reached, and the sheet fits the screen.
    await sheet.evaluate((panel) => panel.scrollTo(0, panel.scrollHeight))
    await expect(sheet.getByRole('group', { name: /face/i })).toBeInViewport()
    // Measured once it has risen into place (it slides up as it opens).
    const viewportHeight = page.viewportSize()!.height
    await expect.poll(async () => {
      const box = await sheet.boundingBox()
      return box ? box.y + box.height : Number.POSITIVE_INFINITY
    }).toBeLessThanOrEqual(viewportHeight + 1)

    await page.getByTestId('gallery-filters-done').click()
    await expect(sheet).toBeHidden()
    expect(await page.evaluate(() => document.documentElement.style.overflow)).toBe('')
  })
})
