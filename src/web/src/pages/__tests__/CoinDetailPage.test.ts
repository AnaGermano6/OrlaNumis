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
const route = reactive({ name: 'coin-detail', params: { id: String(coin.id) } })

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
