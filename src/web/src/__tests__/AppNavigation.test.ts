import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter, type Router } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import en from '@/i18n/locales/en'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const appPath = path.resolve(__dirname, '../App.vue')

const mockIsPwa = vi.hoisted(() => ({ value: false }))
const mockGetSets = vi.hoisted(() => vi.fn())

vi.mock('@/composables/usePwa', () => ({
  usePwa: () => ({ isPwa: mockIsPwa.value }),
}))

vi.mock('sortablejs', () => ({
  default: { create: vi.fn(() => ({ destroy: vi.fn(), toArray: () => [] })) },
}))

vi.mock('virtual:pwa-register', () => ({
  registerSW: () => vi.fn(async () => undefined),
}))

// Public-dir asset referenced by App.vue's brand <img>; the Vitest module
// runner cannot resolve root-absolute public assets.
vi.mock('/coin-logo.jpg', () => ({ default: '/coin-logo.jpg' }))

vi.mock('@/api/client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/api/client')>()
  return {
    ...actual,
    getSets: (...args: unknown[]) => mockGetSets(...args),
    getMe: vi.fn(async () => ({ data: { id: 1, emailMissing: false, createdAt: '2020-01-01T00:00:00Z' } })),
    getCoinOptions: vi.fn(async () => ({ data: { categories: [], eras: [], labels: {} } })),
    updateProfile: vi.fn(async () => ({ data: {} })),
  }
})

describe('App sidebar navigation', () => {
  it('offers the overview, the gallery and the banknotes under Collection, without the removed tray views', () => {
    const source = fs.readFileSync(appPath, 'utf8')
    const routerSource = fs.readFileSync(path.resolve(__dirname, '../router/index.ts'), 'utf8')

    expect(source).toContain("labelKey: 'nav.collectionOverview', to: '/collection'")
    expect(source).toContain("labelKey: 'nav.gallery', to: '/'")
    expect(source).toContain("labelKey: 'nav.banknotes', to: '/banknotes'")
    expect(routerSource).toContain("path: '/collection'")
    expect(source).not.toContain("to: '/tray'")
    expect(source).not.toContain("to: '/storage-trays'")
    // Old links open the gallery.
    expect(routerSource).toContain("{ path: '/:pathMatch(.*)*', name: 'not-found', redirect: '/' }")
  })

  it('renders Stats as a collapsible parent with dedicated route children', () => {
    const source = fs.readFileSync(appPath, 'utf8')

    expect(source).toContain("labelKey: 'nav.stats'")
    expect(source).toContain("labelKey: 'nav.timeline', to: '/stats/timeline'")
    expect(source).toContain("labelKey: 'nav.map', to: '/stats/mint-map'")
    expect(source).toContain("labelKey: 'nav.health', to: '/stats/health'")
    // The value details page was removed.
    expect(source).not.toContain("to: '/stats/value-trends'")
    expect([en.nav.stats, en.nav.timeline, en.nav.map, en.nav.health])
      .toEqual(['Stats', 'Timeline', 'Map', 'Health'])
    expect(source).not.toContain("id: 'stats-emperors'")
    expect(source).not.toContain("to: '/stats/emperors'")
    expect(source).not.toContain("id: 'timeline'")
    expect(source).not.toContain("labelKey: 'nav.collectionDistribution'")
    expect(source).not.toContain('#collection-health')
    expect(source).not.toContain('#value-over-time')
  })

  it('renders Sets as a collapsible parent, without the removed Emperors, pins, notifications or Quick Access', () => {
    const source = fs.readFileSync(appPath, 'utf8')

    expect(source).toContain("labelKey: 'nav.sets'")
    expect(source).toContain("labelKey: 'nav.mySets', to: '/sets'")
    expect([en.nav.sets, en.nav.mySets]).toEqual(['Sets', 'My Sets'])
    expect(source).not.toContain('emperor')
    expect(source).not.toContain('sets-pinned-')
    expect(source).not.toContain('/notifications')
    expect(source).not.toContain('/quick-access')
  })

  it('uses Identify Coin as the single merged quick capture entry point', () => {
    const source = fs.readFileSync(appPath, 'utf8')

    expect(source).not.toContain("id: 'quick-capture'")
    expect(source).not.toContain("labelKey: 'nav.quickCapture'")
    expect(source).not.toContain("to: '/quick-capture'")
    expect(source).toContain("id: 'lookup'")
    expect(source).toContain("labelKey: 'nav.identifyCoin'")
    expect(source).toContain("to: '/lookup'")
    expect(source).toContain("id: 'add-coin'")
    expect(source).toContain("labelKey: 'nav.addCoin'")
    expect([en.nav.identifyCoin, en.nav.addCoin]).toEqual(['Identify Coin', 'Add Coin'])
  })

  it('preserves AI intake identification navigation without adding Quick Capture AI expansion', () => {
    const source = fs.readFileSync(appPath, 'utf8')
    const routerSource = fs.readFileSync(path.resolve(__dirname, '../router/index.ts'), 'utf8')

    expect(source).toContain("id: 'lookup'")
    expect(source).toContain("labelKey: 'nav.identifyCoin'")
    expect(source).toContain("to: '/lookup'")
    expect(routerSource).toContain("path: '/lookup'")
    expect(routerSource).not.toContain('/quick-capture/intake')
    expect(routerSource).not.toContain('/quick-capture/ai')
  })

  it('has no Getting Started link: the guide opens from the first-login prompt and Settings → Help', () => {
    const source = fs.readFileSync(appPath, 'utf8')

    expect(source).not.toContain('gettingStarted')
    // The first-login prompt still opens the guide.
    expect(source).toContain('@click="openOnboardingGuide"')
    expect(source).toContain("query: { tab: 'help' }")
  })

  it('has no shipment tracker, auctions or calendar: their old addresses lead elsewhere', () => {
    const source = fs.readFileSync(appPath, 'utf8')
    const routerSource = fs.readFileSync(path.resolve(__dirname, '../router/index.ts'), 'utf8')

    for (const id of ['shipments', 'auctions', 'calendar']) expect(source).not.toContain(`id: '${id}'`)
    expect(routerSource).not.toContain('CoinDetailShipmentPage')
    expect(routerSource).not.toContain('CalendarPage')
    for (const path of ['/coin/:id/shipment', '/auctions', '/calendar']) expect(routerSource).not.toContain(`path: '${path}'`)
    expect(routerSource).toContain("{ path: '/:pathMatch(.*)*', name: 'not-found', redirect: '/' }")
  })

  it('keeps the same app-wide Agent drawer entry points for legacy chat and Coin Copilot', () => {
    const source = fs.readFileSync(appPath, 'utf8')

    expect(source).toContain('{ id: \'agent\', labelKey: \'nav.agent\'')
    expect(source).toContain('handleAgentFabClick')
    expect(source).toContain('<CoinSearchChat v-if="showChat"')
    expect(source.match(/<CoinSearchChat v-if="showChat"/g) ?? []).toHaveLength(1)
  })
})

// Behavioural coverage for the mounted sidebar and title bar.
describe('App sidebar', () => {
  let router: Router
  let wrapper: VueWrapper | null = null

  async function mountApp(initialPath = '/') {
    document.getElementById('desktop-page-title')?.remove()
    document.getElementById('desktop-page-actions')?.remove()

    const DesktopPageContext = (await import('../components/DesktopPageContext.vue')).default
    const ContextPage = {
      components: { DesktopPageContext },
      template: `
        <DesktopPageContext title="Wishlist">
          <template #actions><button aria-label="Add to wishlist">Add</button></template>
        </DesktopPageContext>
        <div>Context page</div>
      `,
    }
    router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/', name: 'collection', component: { template: '<div />' } },
        { path: '/sets', name: 'sets', component: { template: '<div />' } },
        { path: '/sets/:id', name: 'set-detail', component: { template: '<div />' } },
        { path: '/context', name: 'context', component: ContextPage },
        { path: '/plain', name: 'plain', component: { template: '<div>Plain page</div>' } },
        { path: '/login', name: 'login', component: { template: '<div />' } },
        { path: '/:pathMatch(.*)*', name: 'catch-all', component: { template: '<div />' } },
      ],
    })
    await router.push(initialPath)
    await router.isReady()

    const App = (await import('../App.vue')).default
    wrapper = mount(App, {
      attachTo: document.body,
      global: {
        plugins: [router],
        stubs: {
          CoinSearchChat: true,
          AppDialog: true,
          AppToasts: true,
          PwaInstallPrompt: true,
          PwaUpdateBanner: true,
        },
      },
    })
    await flushPromises()
    return wrapper
  }

  async function openSidebar() {
    await wrapper!.find('nav button').trigger('click')
    await flushPromises()
  }

  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    localStorage.setItem('token', 'test-token')
    localStorage.setItem('user', JSON.stringify({ id: 1, username: 'tester', role: 'user' }))
    setActivePinia(createPinia())
    mockIsPwa.value = false
    mockGetSets.mockReset()
  })

  afterEach(() => {
    wrapper?.unmount()
    wrapper = null
    document.body.innerHTML = ''
  })

  it('renders page-owned desktop context and clears it on navigation', async () => {
    await mountApp('/context')

    expect(wrapper!.get('#desktop-page-title').text()).toBe('Wishlist')
    expect(wrapper!.get('#desktop-page-actions [aria-label="Add to wishlist"]').text()).toBe('Add')

    await router.push('/plain')
    await flushPromises()

    expect(wrapper!.get('#desktop-page-title').text()).toBe('')
    expect(wrapper!.get('#desktop-page-actions').text()).toBe('')
  })

  it('keeps the Sets parent collapsed on load and lists its views once opened', async () => {
    await mountApp()
    await openSidebar()

    expect(wrapper!.find('aside').exists()).toBe(true)
    expect(wrapper!.find('[aria-label="Sets views"]').exists()).toBe(false)

    const setsParent = wrapper!.findAll('.sidebar-link').find((el) => el.text() === 'Sets')
    await setsParent!.trigger('click')
    await flushPromises()

    const links = wrapper!.get('[aria-label="Sets views"]').findAll('a')
    expect(links.map((l) => l.text())).toEqual(['My Sets', 'Commemorative catalogue'])
    expect(links.map((l) => l.attributes('href'))).toEqual(['/sets', '/sets/commemoratives'])
    expect(mockGetSets).not.toHaveBeenCalled()
  })

  it('has no notification bell or Quick Access link', async () => {
    await mountApp()

    expect(wrapper!.find('nav a[aria-label="Notifications"]').exists()).toBe(false)
    expect(wrapper!.find('nav a[aria-label="Quick Access"]').exists()).toBe(false)
  })

  it('signs out to the login page from the sidebar', async () => {
    await mountApp()
    await openSidebar()

    const logoutButton = wrapper!.findAll('.sidebar-link').find((el) => el.text() === 'Logout')
    await logoutButton!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.path).toBe('/login')
    expect(useAuthStore().user).toBeNull()
  })
})
