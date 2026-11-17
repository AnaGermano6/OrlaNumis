import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { DOMWrapper, shallowMount } from '@vue/test-utils'
import type { Coin } from '@/types'
import WishlistPage from '../WishlistPage.vue'

// The page with its dialogs, which open at the end of the document.
const page = () => new DOMWrapper(document.body)

const mockStore = {
  loading: false,
  coins: [] as Coin[],
  total: 0,
  fetchCoins: vi.fn(),
  selectedRepeated: false,
}
let mockIsMobile = false

vi.mock('@/stores/coins', () => ({
  useCoinsStore: () => mockStore,
}))

vi.mock('@/composables/useLayout', async () => {
  const { ref } = await import('vue')
  return { useLayout: () => ({ isMobile: ref(mockIsMobile) }) }
})

const mockPush = vi.hoisted(() => vi.fn())
vi.mock('vue-router', async (importOriginal) => ({
  ...await importOriginal<typeof import('vue-router')>(),
  useRouter: () => ({ push: mockPush }),
}))

vi.mock('@/api/client', () => ({
  purchaseCoin: vi.fn(),
  checkWishlistAvailability: vi.fn(),
  updateListingStatus: vi.fn(),
}))

function createCoin(id: number): Coin {
  return {
    id,
    name: `Coin ${id}`,
    category: 'Roman',
    denomination: 'Denarius',
    ruler: 'Philip I',
    era: 'Roman Empire',
    mint: 'Rome',
    material: 'Silver',
    weightGrams: null,
    diameterMm: null,
    grade: '',
    obverseInscription: '',
    reverseInscription: '',
    obverseDescription: '',
    reverseDescription: '',
    rarityRating: '',
    purchasePrice: 800,
    currentValue: null,
    purchaseDate: null,
    purchaseLocation: '',
    notes: '',
    aiAnalysis: '',
    obverseAnalysis: '',
    reverseAnalysis: '',
    referenceUrl: '',
    referenceText: '',
    isWishlist: true,
    listingStatus: 'available',
    listingCheckedAt: null,
    listingCheckReason: '',
    userId: 1,
    images: [],
    createdAt: '',
    updatedAt: '',
  }
}

const routerLinkStub = {
  props: ['to'],
  template: '<a :href="to" :title="$attrs.title"><slot /></a>',
}

function mountPage() {
  return shallowMount(WishlistPage, {
    global: {
      stubs: {
        RouterLink: routerLinkStub,
        DesktopPageContext: false,
        AppIconButton: false,
        Teleport: false,
        BaseModal: false,
      },
    },
  })
}

describe('WishlistPage', () => {
  beforeEach(() => {
    mockStore.loading = false
    mockStore.coins = []
    mockStore.total = 0
    mockStore.fetchCoins.mockReset()
    mockIsMobile = false
    document.body.innerHTML = `
      <div id="desktop-page-title"></div>
      <div id="desktop-page-actions"></div>
    `
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('does not show the empty state when wishlist coins are present on a single page', () => {
    mockStore.coins = [createCoin(1)]
    mockStore.total = 1

    const wrapper = mountPage()

    expect(mockStore.fetchCoins).toHaveBeenCalledWith({ wishlist: 'true', sort: 'updated_at', order: 'desc', page: 1 })
    expect(wrapper.find('.coins-grid').exists()).toBe(true)
    expect(wrapper.find('.empty-state').exists()).toBe(false)
    expect(wrapper.find('.pagination').exists()).toBe(false)
  })

  it('continues to fetch only wishlist coins and never quick-capture drafts', () => {
    mountPage()

    expect(mockStore.fetchCoins).toHaveBeenCalledWith({ wishlist: 'true', sort: 'updated_at', order: 'desc', page: 1 })
  })

  it('shows the empty state when no wishlist coins are present', () => {
    const wrapper = mountPage()

    expect(wrapper.find('.coins-grid').exists()).toBe(false)
    expect(wrapper.find('.empty-state').exists()).toBe(true)
    const finderLink = wrapper.find('a[title="Add Wish List Finder Agent"]')
    expect(finderLink.exists()).toBe(true)
    expect(finderLink.attributes('href')).toBe('/wishlist/search-alerts')
    expect(finderLink.text()).toContain('Add Wish List Finder Agent')
  })

  it('opens the desktop add dialog with URL and image-analysis choices', async () => {
    const wrapper = mountPage()

    const addButton = wrapper.find('[data-testid="page-context"] button[title="Add to wishlist"]')
    expect(addButton.exists()).toBe(true)
    await addButton.trigger('click')
    await wrapper.vm.$nextTick()

    expect(page().get('[role="dialog"]').text()).toContain('Add from URL')
    expect(page().get('[role="dialog"]').text()).toContain('Add from image analysis')
    expect(page().get('a[href="/lookup"]').text()).toContain('Add from image analysis')
  })

  it('shows the desktop search alerts icon when wishlist coins are present', () => {
    mockStore.coins = [createCoin(1)]
    mockStore.total = 1

    const wrapper = mountPage()

    const finderLink = wrapper.find('[data-testid="page-context"] a[href="/wishlist/search-alerts"]')
    expect(finderLink.exists()).toBe(true)
    expect(finderLink.attributes('title')).toBe('Search Alerts')
  })

  it('opens the add dialog from the PWA plus icon', async () => {
    mockIsMobile = true

    const wrapper = mountPage()

    await wrapper.get('button[aria-label="Add to wishlist"]').trigger('click')

    expect(page().get('[role="dialog"]').text()).toContain('Add from URL')
    expect(page().get('[role="dialog"]').text()).toContain('Add from image analysis')
  })

  it('shows the finder agent icon in PWA mode when wishlist coins are present', () => {
    mockIsMobile = true
    mockStore.coins = [createCoin(1)]
    mockStore.total = 1

    const wrapper = mountPage()

    const finderLink = wrapper.find('.pwa-actions a[href="/wishlist/search-alerts"]')
    expect(finderLink.exists()).toBe(true)
    expect(finderLink.attributes('href')).toBe('/wishlist/search-alerts')
    expect(finderLink.text()).not.toContain('Search Alerts')
  })


  it('has a section for the coins and banknotes held more than once, which opens them and shares them', async () => {
    const wrapper = mountPage()

    const section = wrapper.get('[data-testid="wishlist-trades"]')
    expect(section.get('h2').text()).toBe('Duplicates to trade')
    const lists = section.findAllComponents({ name: 'CollectionRepeated' })
    expect(lists.map((list) => [list.props('kind'), list.props('title')])).toEqual([['coin', 'Coins'], ['banknote', 'Banknotes']])

    // The coins in the gallery with only the duplicates; the banknotes in theirs.
    lists[0]!.vm.$emit('open')
    expect(mockStore.selectedRepeated).toBe(true)
    expect(mockPush).toHaveBeenLastCalledWith('/')
    lists[1]!.vm.$emit('open')
    expect(mockPush).toHaveBeenLastCalledWith({ name: 'banknotes', query: { repeated: 'true' } })

    // Sharing them starts on the duplicates; the page's own share, on the wishlist.
    lists[0]!.vm.$emit('share')
    await wrapper.vm.$nextTick()
    expect(wrapper.findComponent({ name: 'ShareListDialog' }).props()).toMatchObject({ open: true, initialList: 'trades' })
  })
})
