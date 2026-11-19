import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import SetDashboardCard from '../SetDashboardCard.vue'
import type { CoinSetSummary } from '@/types'
import { setLocale } from '@/i18n'

function card(set: Partial<CoinSetSummary>) {
  return mount(SetDashboardCard, {
    props: { set: { id: 1, name: 'Set', color: '#c9a84c', setType: 'standard', coinCount: 0, totalValue: 0, ...set } as CoinSetSummary },
    global: { stubs: { AuthenticatedImage: { props: ['mediaPath'], template: '<img class="own" :src="mediaPath" />' } } },
  })
}
const text = (value: string) => value.replace(/\s+/g, ' ').trim()

describe('SetDashboardCard', () => {
  beforeEach(() => setLocale('pt-PT'))
  afterEach(() => setLocale('en'))

  it("shows a joint issue's year apart, its name after it, how many of its coins the collection has of how many, and how far it is", async () => {
    const wrapper = card({
      name: 'Emissão comum de 2007: 50th anniversary of the Treaty of Rome', setType: 'checklist', year: 2007,
      coinCount: 1, total: 13, completionPercentage: 7.69, oneCoin: true, coverImageUrls: ['https://www.ecb.europa.eu/rome/Portugal.jpg'],
    })

    expect(wrapper.get('[data-testid="set-card-year"]').text()).toBe('2007')
    expect(wrapper.get('[data-testid="set-card-title"]').text()).toBe('50th anniversary of the Treaty of Rome')
    expect(wrapper.attributes('title')).toBe('Emissão comum de 2007: 50th anniversary of the Treaty of Rome')
    expect(text(wrapper.get('[data-testid="set-card-count"]').text())).toBe('1 / 13')
    expect(text(wrapper.get('[data-testid="set-card-progress"]').text())).toBe('8%')
    expect(wrapper.get('[data-testid="set-card-progress"] [aria-hidden] span').attributes('style')).toContain('width: 7.69%')
    // One coin, struck by every country: its picture alone.
    expect(wrapper.get('[data-testid="set-card-cover"]').attributes('data-cover')).toBe('coin')
    expect(wrapper.get('[data-testid="set-card-cover"] img').attributes('src')).toBe('https://www.ecb.europa.eu/rome/Portugal.jpg')
    await wrapper.trigger('click')
    expect(wrapper.emitted('click')).toHaveLength(1)
  })

  it("shows another set's whole name, its coins, a photo of one of them and what kind of set it is", () => {
    const wrapper = card({ name: 'Moedas de 1991', setType: 'smart', coinCount: 6 })

    expect(wrapper.find('[data-testid="set-card-year"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="set-card-title"]').text()).toBe('Moedas de 1991')
    expect(wrapper.get('[data-testid="set-card-count"]').findAll(':scope > span').map((part) => part.text())).toEqual(['6', 'Moedas'])
    expect(wrapper.find('[data-testid="set-card-progress"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="set-card-kind"]').text()).toBe('Conjunto inteligente')
    // Six coins, without pictures: three blank coins piled up.
    expect(wrapper.get('[data-testid="set-card-cover"]').attributes('data-cover')).toBe('stack')
    expect(wrapper.findAll('[data-testid="set-card-stack-coin"]')).toHaveLength(3)
    expect(wrapper.findAll('[data-testid="set-card-blank-coin"]')).toHaveLength(3)
  })

  it("piles up three of a set's different coins, the first in front, blank coins for those it lacks", () => {
    const catalogue = card({ name: 'Portugal · 2 €', setType: 'checklist', coinCount: 36, total: 37, oneCoin: false, coverImageUrls: ['a.jpg', 'b.jpg', 'c.jpg'] })
    const coins = catalogue.findAll('[data-testid="set-card-stack-coin"]')
    expect(coins.map((coin) => coin.find('img').attributes('src'))).toEqual(['b.jpg', 'c.jpg', 'a.jpg'])
    expect(coins[2]!.classes()).toContain('z-[3]')

    const own = card({ name: 'Meus', coinCount: 2, coverImages: ['coins/one.jpg'] })
    expect(own.get('[data-testid="set-card-cover"]').attributes('data-cover')).toBe('stack')
    expect(own.findAll('img.own').map((image) => image.attributes('src'))).toEqual(['coins/one.jpg'])
    expect(own.findAll('[data-testid="set-card-blank-coin"]')).toHaveLength(2)

    // A set of one coin shows it alone.
    const single = card({ name: 'Uma', coinCount: 1, coverImages: ['coins/one.jpg'] })
    expect(single.get('[data-testid="set-card-cover"]').attributes('data-cover')).toBe('coin')
  })

  it('rings a set without a picture with its colour; a goal shows how far it is, an agentic set being made does not', () => {
    const empty = card({ name: 'Vazio', color: '#10b981' })
    expect(empty.get('[data-testid="set-card-cover"]').attributes('style')).toContain('#10b981')
    expect(empty.get('[data-testid="set-card-kind"]').text()).toBe('Pronto para moedas')
    expect(text(card({ setType: 'goal', coinCount: 3, completionPercentage: 50 }).get('[data-testid="set-card-progress"]').text())).toBe('50%')
    expect(card({ setType: 'agentic', agenticStatus: 'generating' }).get('[data-testid="set-card-kind"]').text()).toBe('Conjunto agêntico em geração')
  })
})
