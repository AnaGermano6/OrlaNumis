import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { mount } from '@vue/test-utils'
import CoinDetailHeaderActions from '../CoinDetailHeaderActions.vue'

const pushMock = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: pushMock }),
}))

const isPwa = ref(false)

vi.mock('@/composables/usePwa', () => ({
  usePwa: () => ({ isPwa: isPwa.value }),
}))

const routerLinkStub = {
  props: ['to'],
  template: '<a :href="to"><slot /></a>',
}

describe('CoinDetailHeaderActions', () => {
  beforeEach(() => {
    isPwa.value = false
  })

  it('routes back to wishlist gallery for wishlist items', async () => {
    pushMock.mockReset()
    const wrapper = mount(CoinDetailHeaderActions, {
      props: {
        isWishlist: true,
        coinId: 42,
      },
      global: {
        stubs: {
          RouterLink: routerLinkStub,
          ArrowLeft: true,
          CircleDollarSign: true,
          Copy: true,
          Pencil: true,
          Share2: true,
          Trash2: true,
        },
      },
    })

    expect(wrapper.find('button').text()).toBe('Back to the wishlist')
    await wrapper.find('button').trigger('click')
    expect(pushMock).toHaveBeenCalledWith('/wishlist')
  })

  it('routes back to collection gallery for non-wishlist items', async () => {
    pushMock.mockReset()
    const wrapper = mount(CoinDetailHeaderActions, {
      props: {
        isWishlist: false,
        coinId: 42,
      },
      global: {
        stubs: {
          RouterLink: routerLinkStub,
          ArrowLeft: true,
          CircleDollarSign: true,
          Copy: true,
          Pencil: true,
          Share2: true,
          Trash2: true,
        },
      },
    })

    expect(wrapper.find('button').text()).toBe('Back to the coin gallery')
    await wrapper.find('button').trigger('click')
    expect(pushMock).toHaveBeenCalledWith('/')
  })

  it('routes a banknote back to the banknotes', async () => {
    pushMock.mockReset()
    const wrapper = mount(CoinDetailHeaderActions, {
      props: { isWishlist: false, isBanknote: true, coinId: 42 },
      global: { stubs: { RouterLink: routerLinkStub, ArrowLeft: true, CircleDollarSign: true, Copy: true, Pencil: true, Share2: true, Trash2: true } },
    })

    const back = wrapper.find('button')
    expect(back.text()).toBe('Back to the banknote gallery')
    await back.trigger('click')
    expect(pushMock).toHaveBeenCalledWith('/banknotes')

    // Copying a banknote says so.
    await wrapper.find('button[aria-label="Open overflow actions"]').trigger('click')
    const copy = wrapper.find('button[aria-label="Copy banknote"]')
    expect(copy.exists()).toBe(true)
    await copy.trigger('click')
    expect(wrapper.emitted('duplicate')).toHaveLength(1)
  })

  it('keeps primary icons visible and moves copy/details into overflow menu', async () => {
    const wrapper = mount(CoinDetailHeaderActions, {
      props: {
        isWishlist: false,
        coinId: 42,
      },
      global: {
        stubs: {
          RouterLink: routerLinkStub,
          ArrowLeft: true,
          CircleDollarSign: true,
          Copy: true,
          Pencil: true,
          Share2: true,
          Trash2: true,
        },
      },
    })

    expect(wrapper.find('button[aria-label="Share"]').exists()).toBe(true)
    expect(wrapper.find('button[aria-label="Edit"]').exists()).toBe(true)
    expect(wrapper.find('button[aria-label="Delete"]').exists()).toBe(true)
    expect(wrapper.find('button[aria-label="Open overflow actions"]').exists()).toBe(true)

    await wrapper.find('button[aria-label="Share"]').trigger('click')

    expect(wrapper.emitted('share')).toHaveLength(1)

    await wrapper.find('button[aria-label="Open overflow actions"]').trigger('click')
    // Selling and shipment tracking are not used in this collection.
    expect(wrapper.find('button[aria-label="Sell Coin"]').exists()).toBe(false)
    expect(wrapper.find('button[aria-label="Copy Coin"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('Shipment Tracker')
    expect(wrapper.text()).toContain('Activity Journal')
    expect(wrapper.text()).toContain('Value Trend')
  })

  it('disables the share button while a share card is being generated', () => {
    const wrapper = mount(CoinDetailHeaderActions, {
      props: {
        isWishlist: false,
        coinId: 42,
        sharing: true,
      },
      global: {
        stubs: {
          RouterLink: routerLinkStub,
          ArrowLeft: true,
          CircleDollarSign: true,
          Copy: true,
          Pencil: true,
          Share2: true,
          Trash2: true,
        },
      },
    })

    const shareButton = wrapper.find('button[aria-label="Sharing..."]')
    expect(shareButton.attributes('disabled')).toBeDefined()
  })

  it('emits duplicate from overflow and disables copy while pending', async () => {
    const activeWrapper = mount(CoinDetailHeaderActions, {
      props: {
        isWishlist: false,
        coinId: 42,
      },
      global: {
        stubs: {
          RouterLink: routerLinkStub,
          ArrowLeft: true,
          CircleDollarSign: true,
          Copy: true,
          Pencil: true,
          Share2: true,
          Trash2: true,
        },
      },
    })

    await activeWrapper.find('button[aria-label="Open overflow actions"]').trigger('click')
    await activeWrapper.find('button[aria-label="Copy Coin"]').trigger('click')

    expect(activeWrapper.emitted('duplicate')).toHaveLength(1)

    const pendingWrapper = mount(CoinDetailHeaderActions, {
      props: {
        isWishlist: false,
        coinId: 42,
        duplicating: true,
      },
      global: {
        stubs: {
          RouterLink: routerLinkStub,
          ArrowLeft: true,
          CircleDollarSign: true,
          Copy: true,
          Pencil: true,
          Share2: true,
          Trash2: true,
        },
      },
    })

    await pendingWrapper.find('button[aria-label="Open overflow actions"]').trigger('click')
    const duplicateButton = pendingWrapper.find('button[aria-label="Copying coin..."]')
    expect(duplicateButton.attributes('disabled')).toBeDefined()
  })

  it('emits edit from the primary action row', async () => {
    const wrapper = mount(CoinDetailHeaderActions, {
      props: {
        isWishlist: false,
        coinId: 42,
      },
      global: {
        stubs: {
          RouterLink: routerLinkStub,
          ArrowLeft: true,
          CircleDollarSign: true,
          Copy: true,
          Menu: true,
          Pencil: true,
          Share2: true,
          Trash2: true,
        },
      },
    })

    await wrapper.find('button[aria-label="Edit"]').trigger('click')
    expect(wrapper.emitted('edit')).toHaveLength(1)
  })

  it.each([true, false])('moves share into the overflow menu in the PWA (isWishlist: %s)', async (isWishlist) => {
    isPwa.value = true
    const wrapper = mount(CoinDetailHeaderActions, {
      props: { isWishlist, coinId: 42 },
      global: {
        stubs: { RouterLink: routerLinkStub },
      },
    })

    expect(wrapper.find('button[aria-label="Share"]').exists()).toBe(false)
    expect(wrapper.find('button[aria-label="Edit"]').exists()).toBe(true)
    expect(wrapper.find('button[aria-label="Delete"]').exists()).toBe(true)

    await wrapper.find('button[aria-label="Open overflow actions"]').trigger('click')
    await wrapper.get('button[aria-label="Share"]').trigger('click')
    expect(wrapper.emitted('share')).toHaveLength(1)
  })

  it('keeps share on the action row outside the PWA', async () => {
    const wrapper = mount(CoinDetailHeaderActions, {
      props: { isWishlist: true, coinId: 42 },
      global: {
        stubs: { RouterLink: routerLinkStub },
      },
    })

    expect(wrapper.find('button[aria-label="Share"]').exists()).toBe(true)

    await wrapper.find('button[aria-label="Open overflow actions"]').trigger('click')
    expect(wrapper.findAll('button[aria-label="Share"]')).toHaveLength(1)
  })

  it('names in Portuguese the gallery it goes back to', async () => {
    const { setLocale } = await import('@/i18n')
    setLocale('pt-PT')
    const stubs = { RouterLink: routerLinkStub, ArrowLeft: true, CircleDollarSign: true, Copy: true, Pencil: true, Share2: true, Trash2: true }
    const label = (props: { isWishlist: boolean; isBanknote?: boolean }) => mount(CoinDetailHeaderActions, { props: { ...props, coinId: 42 }, global: { stubs } }).find('button').text()

    expect(label({ isWishlist: false })).toBe('Voltar à galeria de moedas')
    expect(label({ isWishlist: false, isBanknote: true })).toBe('Voltar à galeria de notas')
    expect(label({ isWishlist: true, isBanknote: true })).toBe('Voltar à lista de desejos')
    setLocale('en')
  })
})
