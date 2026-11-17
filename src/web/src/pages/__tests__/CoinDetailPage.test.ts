import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import CoinDetailPage from '../CoinDetailPage.vue'
import router from '@/router'
import { buildRomanDenariusCore } from '@/test/fixtures/coins'

const coin = buildRomanDenariusCore()
const fetchCoin = vi.fn()
const routerPush = vi.fn()
const shareCoinCard = vi.fn()
const sharing = ref(false)
const showConfirm = vi.fn()
const route = reactive<{ name: string; params: { id: string }; hash?: string }>({ name: 'coin-detail', params: { id: String(coin.id) } })

vi.mock('@/stores/coins', () => ({
  useCoinsStore: () => ({
    loading: false,
    currentCoin: coin,
    fetchCoin,
  }),
}))

vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-router')>()
  return {
    ...actual,
    useRoute: () => route,
    useRouter: () => ({ push: routerPush }),
  }
})

vi.mock('@/api/client', () => ({
  createCoinReference: vi.fn(),
  deleteCoin: vi.fn(),
  deleteCoinReference: vi.fn(),
  duplicateCoin: vi.fn(),
  getCoinStatus: vi.fn(),
  listCatalogs: vi.fn().mockResolvedValue([]),
  purchaseCoin: vi.fn(),
  swapCoinImageSides: vi.fn().mockResolvedValue({ data: [] }),
  updateCoinReference: vi.fn(),
}))

import { deleteCoin, duplicateCoin, getCoinStatus, purchaseCoin, swapCoinImageSides } from '@/api/client'

vi.mock('@/composables/useDialog', () => ({
  useDialog: () => ({
    showConfirm,
    showAlert: vi.fn(),
  }),
}))

vi.mock('@/composables/useCoinShareCard', () => ({
  useCoinShareCard: () => ({
    sharing,
    shareCoinCard,
  }),
}))

const routerLinkStub = {
  props: ['to'],
  template: '<a :href="to"><slot /></a>',
}

describe('CoinDetailPage', () => {
  beforeEach(() => {
    vi.mocked(getCoinStatus).mockReset()
    vi.mocked(getCoinStatus).mockResolvedValue({ data: { current: null, commemorative: null } } as Awaited<ReturnType<typeof getCoinStatus>>)
    fetchCoin.mockReset()
    routerPush.mockReset()
    shareCoinCard.mockReset()
    shareCoinCard.mockResolvedValue({ mode: 'downloaded' })
    vi.mocked(duplicateCoin).mockReset()
    vi.mocked(deleteCoin).mockReset()
    vi.mocked(purchaseCoin).mockReset()
    vi.mocked(duplicateCoin).mockResolvedValue({ data: { ...coin, id: 314 } })
    vi.mocked(deleteCoin).mockResolvedValue({ data: {} })
    vi.mocked(purchaseCoin).mockResolvedValue({ data: coin })
    showConfirm.mockReset()
    showConfirm.mockResolvedValue(true)
    coin.isWishlist = false
    sharing.value = false
  })

  it('loads the coin again when the address goes to another one (previous and next)', async () => {
    mount(CoinDetailPage, { global: { stubs: pageStubs() } })
    await flushPromises()
    expect(fetchCoin).toHaveBeenCalledWith(coin.id)

    route.params.id = '999'
    await flushPromises()
    expect(fetchCoin).toHaveBeenLastCalledWith(999)
    route.params.id = String(coin.id)
    await flushPromises()
  })

  it('keeps the beta/main two-image hero display and shows the share action', () => {
    const wrapper = mount(CoinDetailPage, {
      global: {
        stubs: pageStubs(),
      },
    })

    // Two-column obverse/reverse hero grid — was `.hero-media-grid`, now a plain
    // Tailwind `grid grid-cols-2` container (only the hero media uses grid-cols-2
    // without a responsive prefix; the metadata section below uses `md:grid-cols-2`).
    expect(wrapper.find('.grid-cols-2').exists()).toBe(true)
    expect(wrapper.find('button[aria-label="Share"]').exists()).toBe(true)
    expect(fetchCoin).toHaveBeenCalledWith(coin.id)
  })

  it('shows both sides with their names, then the name, the key facts in a row and the notes', async () => {
    const wrapper = mount(CoinDetailPage, { global: { stubs: pageStubs() } })
    await flushPromises()

    const summary = wrapper.get('[data-testid="coin-summary"]')
    expect(summary.findAll('figcaption').map((caption) => caption.text())).toEqual(['Obverse', 'Reverse'])
    expect(summary.get('h1').text()).toBe(coin.name)
    const facts = summary.findAll('[data-testid="coin-key-facts"] > div').map((fact) => [fact.get('dt').text(), fact.get('dd').text()])
    expect(facts).toEqual([['Material', 'Silver'], ['Weight', '3.35 g'], ['Diameter', '18 mm'], ['Grade', 'VF']])
    expect(summary.get('[data-testid="coin-notes"]').text()).toBe(coin.notes)
    // The details table has the others, not the key facts again.
    const rows = wrapper.findComponent({ name: 'CoinDetailMetadataTable' }).props('rows') as Array<{ key: string }>
    expect(rows.map((row) => row.key)).not.toContain('material')
    expect(rows.map((row) => row.key)).not.toContain('grade')
    expect(rows.map((row) => row.key)).toContain('quantity')
  })

  it('shows the details, inscriptions, history, references and sets one at a time, the details first', async () => {
    const wrapper = mount(CoinDetailPage, { global: { stubs: pageStubs() } })
    await flushPromises()

    const tabs = () => wrapper.findAll('[role="tab"]')
    expect(tabs().map((tab) => tab.text())).toEqual(['Details', 'Inscription', 'History', 'Catalog References', 'Sets'])
    expect(tabs().map((tab) => tab.attributes('aria-selected'))).toEqual(['true', 'false', 'false', 'false', 'false'])
    const shown = () => wrapper.findAll('[role="tabpanel"]').filter((panel) => (panel.element as HTMLElement).style.display !== 'none').map((panel) => panel.attributes('id'))
    expect(shown()).toEqual(['coin-panel-details'])

    await wrapper.get('[data-testid="coin-tab-history"]').trigger('click')
    expect(shown()).toEqual(['coin-panel-history'])
    // The arrow keys go to the next tab, and round.
    await wrapper.get('[role="tablist"]').trigger('keydown', { key: 'ArrowRight' })
    expect(shown()).toEqual(['coin-panel-references'])
    // The market value goes with the references.
    expect(wrapper.get('#coin-panel-references').findComponent({ name: 'CoinMarketValueSection' }).exists()).toBe(true)
    await wrapper.get('[role="tablist"]').trigger('keydown', { key: 'ArrowRight' })
    expect(shown()).toEqual(['coin-panel-sets'])
    expect(wrapper.get('#coin-panel-sets').findComponent({ name: 'CoinSetsSection' }).exists()).toBe(true)
    await wrapper.get('[role="tablist"]').trigger('keydown', { key: 'ArrowRight' })
    expect(shown()).toEqual(['coin-panel-details'])
  })

  it('lays out a banknote the same way: its front and back, then its material, size and grade', async () => {
    const saved = { kind: coin.kind, material: coin.material, widthMm: coin.widthMm, heightMm: coin.heightMm }
    Object.assign(coin, { kind: 'banknote', material: 'Paper', widthMm: 160, heightMm: 75.5 })
    try {
      const wrapper = mount(CoinDetailPage, { global: { stubs: pageStubs() } })
      await flushPromises()

      const summary = wrapper.get('[data-testid="coin-summary"]')
      expect(summary.get('[data-kind]').attributes('data-kind')).toBe('banknote')
      expect(summary.findAll('figcaption').map((caption) => caption.text())).toEqual(['Front', 'Back'])
      const facts = summary.findAll('[data-testid="coin-key-facts"] > div').map((fact) => fact.get('dt').text())
      expect(facts).toEqual(['Material', 'Size', 'Condition'])
      expect(summary.get('[data-testid="coin-key-facts"]').text()).toContain('160 × 75.5 mm')
      expect(summary.findAll('[role="tab"]').map((tab) => tab.text())).toEqual(['Details', 'Inscription', 'History', 'Catalog References', 'Sets'])
    } finally {
      // The coin is shared with the other tests.
      Object.assign(coin, saved)
    }
  })

  it('opens the catalogue references when the address asks for them', async () => {
    route.hash = '#catalog-references'
    const wrapper = mount(CoinDetailPage, { global: { stubs: pageStubs() } })
    await flushPromises()
    expect(wrapper.get('[data-testid="coin-tab-references"]').attributes('aria-selected')).toBe('true')
    route.hash = undefined
  })

  it('swaps the obverse and reverse images of a coin that has both', async () => {
    const wrapper = mount(CoinDetailPage, { global: { stubs: pageStubs() } })
    fetchCoin.mockClear()

    const swap = wrapper.findAll('button').find((button) => button.text() === 'Swap obverse and reverse')!
    await swap.trigger('click')
    await flushPromises()

    expect(swapCoinImageSides).toHaveBeenCalledWith(coin.id)
    expect(fetchCoin).toHaveBeenCalledWith(coin.id)
  })

  it('shares the currently loaded coin when the Share action is clicked', async () => {
    const wrapper = mount(CoinDetailPage, {
      global: {
        stubs: pageStubs(),
      },
    })

    await wrapper.find('button[aria-label="Share"]').trigger('click')
    await flushPromises()

    expect(shareCoinCard).toHaveBeenCalledWith(coin)
  })

  it('duplicates the loaded coin and navigates to the new detail page', async () => {
    const wrapper = mount(CoinDetailPage, {
      global: {
        stubs: pageStubs(),
      },
    })

    await wrapper.find('button[aria-label="Open overflow actions"]').trigger('click')
    await wrapper.find('button[aria-label="Copy Coin"]').trigger('click')
    await flushPromises()

    expect(duplicateCoin).toHaveBeenCalledWith(coin.id)
    expect(routerPush).toHaveBeenCalledWith('/coin/314')
  })

  it('anchors Catalog References on the existing coin detail route without a standalone Numista destination', () => {
    const stubs = pageStubs()
    delete stubs.CoinReferencesSection
    const wrapper = mount(CoinDetailPage, {
      global: { stubs },
    })

    const heading = wrapper.findAll('h3').find(item => item.text() === 'Catalog References')
    expect(heading).toBeDefined()
    expect(heading!.element.closest('section')?.id).toBe('catalog-references')

    const numistaRoutes = router.getRoutes().filter(route =>
      /numista/i.test(`${route.path} ${String(route.name ?? '')}`),
    )
    expect(numistaRoutes).toHaveLength(0)
  })

  it('marks a wishlist coin as purchased and reloads it', async () => {
    coin.isWishlist = true
    const stubs = pageStubs()
    stubs.PurchaseModal = {
      template: '<button aria-label="Confirm purchase" @click="$emit(\'confirm\', {})">Confirm</button>',
    }
    const wrapper = mount(CoinDetailPage, { global: { stubs } })
    await wrapper.findAll('button').find((button) => button.text().includes('Mark as Purchased'))!.trigger('click')
    await wrapper.get('button[aria-label="Confirm purchase"]').trigger('click')
    await flushPromises()

    expect(purchaseCoin).toHaveBeenCalledWith(coin.id, {})
    expect(fetchCoin).toHaveBeenLastCalledWith(coin.id)
    expect(fetchCoin).toHaveBeenCalledTimes(2)
  })

  it('labels commemorative coins beside their category', async () => {
    vi.mocked(getCoinStatus).mockResolvedValueOnce({ data: { current: true, commemorative: true } } as Awaited<ReturnType<typeof getCoinStatus>>)
    const commemorative = mount(CoinDetailPage, { global: { stubs: pageStubs() } })
    await flushPromises()
    expect(commemorative.get('[data-testid="commemorative-badge"]').text()).toBe('Commemorative')
    commemorative.unmount()

    vi.mocked(getCoinStatus).mockResolvedValueOnce({ data: { current: true, commemorative: false } } as Awaited<ReturnType<typeof getCoinStatus>>)
    const regular = mount(CoinDetailPage, { global: { stubs: pageStubs() } })
    await flushPromises()
    expect(regular.find('[data-testid="commemorative-badge"]').exists()).toBe(false)
    expect(regular.find('.badge').exists()).toBe(true)
  })

  it('deletes the coin, and offers neither selling nor shipment tracking', async () => {
    const wrapper = mount(CoinDetailPage, { global: { stubs: pageStubs() } })

    await wrapper.get('button[aria-label="Delete"]').trigger('click')
    await flushPromises()
    expect(deleteCoin).toHaveBeenCalledWith(coin.id)
    expect(routerPush).toHaveBeenCalledWith('/')

    await wrapper.find('button[aria-label="Open overflow actions"]').trigger('click')
    expect(wrapper.find('button[aria-label="Sell Coin"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('Shipment')
    expect(wrapper.text()).toContain('Activity Journal')
  })
})

function pageStubs() {
  return {
    RouterLink: routerLinkStub,
    PurchaseModal: true,
    ImageLightbox: true,
    CoinSetsSection: true,
    CoinDetailMetadataTable: true,
    CoinListingStatus: true,
    CoinReferencesSection: true,
    AuthenticatedImage: true,
    ArrowLeft: true,
    CircleDollarSign: true,
    Copy: true,
    Pencil: true,
    Share2: true,
    Trash2: true,
  }
}
