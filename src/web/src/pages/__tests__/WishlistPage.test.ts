import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import type { Coin } from '@/types'
import WishlistPage from '../WishlistPage.vue'

const mockStore = {
  loading: false,
  coins: [] as Coin[],
  total: 0,
  fetchCoins: vi.fn(),
}
let mockIsPwa = false

vi.mock('@/stores/coins', () => ({
  useCoinsStore: () => mockStore,
}))

vi.mock('@/composables/usePwa', () => ({
  usePwa: () => ({
    isPwa: mockIsPwa,
  }),
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
    mockIsPwa = false
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

    const addButton = document.querySelector<HTMLButtonElement>('#desktop-page-actions button[title="Add to wishlist"]')
    expect(addButton).not.toBeNull()
    addButton?.click()
    await wrapper.vm.$nextTick()

    expect(wrapper.get('[role="dialog"]').text()).toContain('Add from URL')
    expect(wrapper.get('[role="dialog"]').text()).toContain('Add from image analysis')
    expect(wrapper.get('a[href="/lookup"]').text()).toContain('Add from image analysis')
  })

  it('shows the desktop search alerts icon when wishlist coins are present', () => {
    mockStore.coins = [createCoin(1)]
    mockStore.total = 1

    mountPage()

    const finderLink = document.querySelector<HTMLAnchorElement>('#desktop-page-actions a[href="/wishlist/search-alerts"]')
    expect(finderLink).not.toBeNull()
    expect(finderLink?.getAttribute('href')).toBe('/wishlist/search-alerts')
    expect(finderLink?.getAttribute('title')).toBe('Search Alerts')
  })

  it('opens the add dialog from the PWA plus icon', async () => {
    mockIsPwa = true

    const wrapper = mountPage()

    await wrapper.get('button[aria-label="Add to wishlist"]').trigger('click')

    expect(wrapper.get('[role="dialog"]').text()).toContain('Add from URL')
    expect(wrapper.get('[role="dialog"]').text()).toContain('Add from image analysis')
  })

  it('shows the finder agent icon in PWA mode when wishlist coins are present', () => {
    mockIsPwa = true
    mockStore.coins = [createCoin(1)]
    mockStore.total = 1

    const wrapper = mountPage()

    const finderLink = wrapper.find('.pwa-actions a[href="/wishlist/search-alerts"]')
    expect(finderLink.exists()).toBe(true)
    expect(finderLink.attributes('href')).toBe('/wishlist/search-alerts')
    expect(finderLink.text()).not.toContain('Search Alerts')
  })

})
