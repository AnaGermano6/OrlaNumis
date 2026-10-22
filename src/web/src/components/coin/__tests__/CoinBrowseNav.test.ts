import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import CoinBrowseNav from '@/components/coin/CoinBrowseNav.vue'
import { rememberBrowseSequence } from '@/utils/browseSequence'
import { setLocale } from '@/i18n'

const mocks = vi.hoisted(() => ({ replace: vi.fn() }))
vi.mock('vue-router', () => ({ useRouter: () => ({ replace: mocks.replace }) }))

function mountNav(coinId: number, isBanknote = false) {
  return mount(CoinBrowseNav, { props: { coinId, isBanknote }, attachTo: document.body })
}

function press(key: string, init: KeyboardEventInit = {}, target: EventTarget = window) {
  target.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true, ...init }))
}

describe('CoinBrowseNav', () => {
  beforeEach(() => {
    sessionStorage.clear()
    vi.clearAllMocks()
    setLocale('pt-PT')
    rememberBrowseSequence('gallery', [10, 11, 12])
  })

  afterEach(() => {
    setLocale('en')
    document.body.innerHTML = ''
  })

  it("shows the coin's place and goes to the previous and the next one", async () => {
    const wrapper = mountNav(11)
    expect(wrapper.get('[data-testid="coin-browse-position"]').text()).toBe('2 de 3')
    expect(wrapper.get('[data-testid="coin-browse-previous"]').attributes('aria-label')).toBe('Moeda anterior')

    await wrapper.get('[data-testid="coin-browse-next"]').trigger('click')
    expect(mocks.replace).toHaveBeenLastCalledWith('/coin/12')
    await wrapper.get('[data-testid="coin-browse-previous"]').trigger('click')
    expect(mocks.replace).toHaveBeenLastCalledWith('/coin/10')
    wrapper.unmount()
  })

  it('stops at the ends of the list', () => {
    const first = mountNav(10)
    expect(first.get('[data-testid="coin-browse-previous"]').attributes('disabled')).toBeDefined()
    expect(first.get('[data-testid="coin-browse-next"]').attributes('disabled')).toBeUndefined()
    first.unmount()
    const last = mountNav(12, true)
    expect(last.get('[data-testid="coin-browse-next"]').attributes('disabled')).toBeDefined()
    expect(last.get('[data-testid="coin-browse-next"]').attributes('aria-label')).toBe('Nota seguinte')
    last.unmount()
  })

  it('is not there for a coin opened from somewhere else', () => {
    const wrapper = mountNav(99)
    expect(wrapper.find('[data-testid="coin-browse-nav"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('goes with the arrow keys, but not while typing, with a modifier or a dialog open', () => {
    const wrapper = mountNav(11)
    press('ArrowRight')
    expect(mocks.replace).toHaveBeenLastCalledWith('/coin/12')
    press('ArrowLeft')
    expect(mocks.replace).toHaveBeenLastCalledWith('/coin/10')
    expect(mocks.replace).toHaveBeenCalledTimes(2)

    press('ArrowLeft', { altKey: true })
    const input = document.createElement('input')
    document.body.appendChild(input)
    press('ArrowRight', {}, input)
    const dialog = document.createElement('div')
    dialog.setAttribute('role', 'dialog')
    document.body.appendChild(dialog)
    press('ArrowRight')
    expect(mocks.replace).toHaveBeenCalledTimes(2)

    dialog.remove()
    wrapper.unmount()
    press('ArrowRight')
    expect(mocks.replace).toHaveBeenCalledTimes(2)
  })
})
