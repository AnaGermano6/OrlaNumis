import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import SetAlbum from '@/components/sets/SetAlbum.vue'
import { setLocale } from '@/i18n'
import type { AlbumSlot } from '@/utils/setAlbum'

const stubs = {
  RouterLink: { props: ['to'], template: '<a data-router-link :href="`/coin/${to.params.id}`"><slot /></a>' },
  AuthenticatedImage: { props: ['mediaPath'], template: '<img :data-media="mediaPath" />' },
}

const SLOTS: AlbumSlot[] = [
  { key: 'a', title: 'Rotary', detail: '2 €', year: 2026, country: 'Portugal', owned: false, imageUrl: 'https://ecb/rotary.jpg', pageUrl: 'https://ecb/rotary' },
  { key: 'b', title: 'Gago Coutinho', year: 2022, country: 'Portugal', owned: true, coinId: 5, coinImage: 'coins/gago.jpg' },
  { key: 'c', title: 'Erasmus', year: 2022, country: 'Áustria', owned: true, coinId: 6 },
  { key: 'd', title: 'Ainda sem nada', owned: false },
]

function titles(wrapper: ReturnType<typeof mount>): string[] {
  return wrapper.findAll('[data-testid="album-title"]').map((node) => node.text())
}

describe('SetAlbum', () => {
  beforeEach(() => {
    setLocale('pt-PT')
  })

  it("shows how far the set is, the collection's coins with their photo and the missing ones dashed", () => {
    const wrapper = mount(SetAlbum, { props: { slots: SLOTS }, global: { stubs } })

    expect(wrapper.get('[data-testid="album-progress"]').text()).toBe('Tem 2 de 4')
    expect(wrapper.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('50')
    const slots = wrapper.findAll('[data-testid="album-slot"]')
    expect(slots.map((slot) => slot.attributes('data-owned'))).toEqual(['false', 'true', 'true', 'false'])
    expect(slots[0]!.classes()).toContain('border-dashed')
    expect(slots[0]!.get('[data-testid="album-missing"]').text()).toBe('Em falta')
    // The coin to find links to its source, its picture faded.
    expect(slots[0]!.get('a').attributes('href')).toBe('https://ecb/rotary')
    expect(slots[0]!.get('img').classes()).toContain('grayscale')
    // The collection's coin opens its page, with its own photo.
    expect(slots[1]!.get('[data-testid="album-owned"]').text()).toBe('Na coleção')
    expect(slots[1]!.get('a').attributes('href')).toBe('/coin/5')
    expect(slots[1]!.get('img').attributes('data-media')).toBe('coins/gago.jpg')
    expect(slots[3]!.find('[data-testid="album-no-image"]').exists()).toBe(true)
  })

  it('filters what the set has or lacks', async () => {
    const wrapper = mount(SetAlbum, { props: { slots: SLOTS }, global: { stubs } })

    expect(wrapper.findAll('[data-testid^="album-show-"]').map((node) => node.text().replace(/\s+/g, ' '))).toEqual(['Todas · 4', 'Na coleção · 2', 'Em falta · 2'])
    await wrapper.get('[data-testid="album-show-missing"]').trigger('click')
    expect(titles(wrapper)).toEqual(['Rotary', 'Ainda sem nada'])
    await wrapper.get('[data-testid="album-show-owned"]').trigger('click')
    expect(titles(wrapper)).toEqual(['Gago Coutinho', 'Erasmus'])
  })

  it('groups by year, newest first, or by country, with the places without one last', async () => {
    const wrapper = mount(SetAlbum, { props: { slots: SLOTS, defaultGroup: 'year' }, global: { stubs } })

    const sections = () => wrapper.findAll('[data-testid="album-section"]').map((node) => node.text().replace(/\s+/g, ' '))
    expect(sections()).toEqual(['2026 · 0 de 1', '2022 · 2 de 2', 'Sem informação · 0 de 1'])
    await wrapper.get('[data-testid="album-group-country"]').trigger('click')
    expect(sections()).toEqual(['Áustria · 1 de 1', 'Portugal · 1 de 2', 'Sem informação · 0 de 1'])
    await wrapper.get('[data-testid="album-group-none"]').trigger('click')
    expect(sections()).toEqual([])
    expect(titles(wrapper)).toEqual(['Rotary', 'Gago Coutinho', 'Erasmus', 'Ainda sem nada'])
  })

  it('offers no filters where there is nothing to tell apart', () => {
    const complete = SLOTS.filter((slot) => slot.owned).map((slot) => ({ ...slot, year: 2022, country: 'Portugal' }))
    const wrapper = mount(SetAlbum, { props: { slots: complete, defaultGroup: 'year' }, global: { stubs } })

    expect(wrapper.find('[data-testid^="album-show-"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid^="album-group-"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="album-section"]').exists()).toBe(false)
    expect(wrapper.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('100')
  })

  it('hands the places of a set the page acts on to the page', async () => {
    const wrapper = mount(SetAlbum, { props: { slots: SLOTS, selectable: true }, global: { stubs } })

    expect(wrapper.find('a').exists()).toBe(false)
    await wrapper.findAll('[data-testid="album-slot"] button')[1]!.trigger('click')
    expect(wrapper.emitted('select')).toEqual([[SLOTS[1]]])
  })
})
