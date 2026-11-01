import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import en from '@/i18n/locales/en'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const srcRoot = path.resolve(__dirname, '..', '..')

describe('MintMap navigation entry points', () => {
  it('moves mint map and timeline under Stats with legacy redirects', () => {
    const routerSource = fs.readFileSync(path.resolve(srcRoot, 'router/index.ts'), 'utf8')

    expect(routerSource).toContain("path: '/stats/mint-map'")
    expect(routerSource).toContain("name: 'stats-mint-map'")
    expect(routerSource).toContain("path: '/stats/timeline'")
    expect(routerSource).toContain("name: 'stats-timeline'")
    expect(routerSource).toContain("path: '/stats/health'")
    expect(routerSource).toContain("name: 'stats-health'")
    // The value details page was removed.
    expect(routerSource).not.toContain("path: '/stats/value-trends'")
    expect(routerSource).toContain("path: '/stats/distribution'")
    expect(routerSource).toContain("name: 'stats-distribution'")
    expect(routerSource).toContain("path: '/mint-map'")
    expect(routerSource).toContain("redirect: '/stats/mint-map'")
    expect(routerSource).toContain("path: '/timeline'")
    expect(routerSource).toContain("redirect: '/stats/timeline'")
  })

  it('keeps stats subviews nested under the Stats sidebar item', () => {
    const appSource = fs.readFileSync(path.resolve(srcRoot, 'App.vue'), 'utf8')

    expect(appSource).toContain("id: 'stats'")
    expect(appSource).toContain("labelKey: 'nav.stats'")
    expect(appSource).toContain("labelKey: 'nav.timeline', to: '/stats/timeline'")
    expect(appSource).toContain("labelKey: 'nav.map', to: '/stats/mint-map'")
    expect(appSource).toContain("labelKey: 'nav.health', to: '/stats/health'")
    expect(appSource).not.toContain("to: '/stats/value-trends'")
    expect([en.nav.stats, en.nav.timeline, en.nav.map, en.nav.health])
      .toEqual(['Stats', 'Timeline', 'Map', 'Health'])
    expect(appSource).not.toContain("id: 'timeline'")
    expect(appSource).not.toContain("labelKey: 'nav.collectionDistribution'")
    expect(appSource).not.toContain('#collection-health')
    expect(appSource).not.toContain('#value-over-time')
  })

  it('sidebar and overlay z-index stack above Leaflet map controls (fixes #294)', () => {
    // Leaflet's map controls sit at z-index 1000 at most. The drawer and its
    // backdrop are on the sidebar layer of the scale (the drawer later in the
    // page, so above its backdrop), which must clear them.
    const appSource = fs.readFileSync(path.resolve(srcRoot, 'App.vue'), 'utf8')
    const css = fs.readFileSync(path.resolve(srcRoot, 'assets', 'styles', 'main.css'), 'utf8')

    expect(appSource).toMatch(/<div v-if="sidebarOpen" class="[^"]*z-\(--z-sidebar\)[^"]*bg-black\/50"/)
    expect(appSource).toMatch(/<aside v-if="sidebarOpen" class="[^"]*z-\(--z-sidebar\)[^"]*"/)
    const sidebarZ = Number(css.match(/--z-sidebar:\s*(\d+);/)?.[1])
    const overlayZ = Number(css.match(/--z-overlay:\s*(\d+);/)?.[1])
    expect(sidebarZ).toBeGreaterThan(1000)
    expect(overlayZ).toBeGreaterThan(1000)
  })
})
