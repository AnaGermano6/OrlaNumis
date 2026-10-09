import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import NumistaCatalogBrowser from '../NumistaCatalogBrowser.vue'
import { setLocale } from '@/i18n'
import type { CoinMarketValue, NumistaCatalogType } from '@/api/client'

const mocks = vi.hoisted(() => ({
  searchNumistaIssuers: vi.fn(),
  searchNumistaCatalog: vi.fn(),
  getNumistaCatalogType: vi.fn(),
  getNumistaTypeMarketValue: vi.fn(),
  searchCoinPhotos: vi.fn(),
  proxyImage: vi.fn(),
}))

vi.mock('@/api/client', () => mocks)

const TYPE: NumistaCatalogType = {
  id: 5,
  url: 'https://en.numista.com/catalogue/pieces5.html',
  title: '50 Escudos',
  issuer: { code: 'portugal', name: 'Portugal' },
  minYear: 1968,
  maxYear: 1972,
  denomination: '50 Escudos',
  rulers: [],
  composition: 'Silver (.650)',
  weightGrams: 18,
  diameterMm: 34,
  mints: ['Lisbon'],
  obverse: { description: 'Shield', lettering: 'REPUBLICA\nPORTUGUESA' },
  reverse: {},
  edge: {},
  references: [{ catalogue: 'KM', number: '599' }],
  photos: [{ side: 'obverse', thumbnail: 'https://en.numista.com/o.jpg', copyright: 'Ana Example', license: 'CC BY-NC' }],
}

function marketValue(currency: string, price: number): CoinMarketValue {
  return {
    status: 'ok',
    numistaTypeId: 5,
    issues: [],
    selectedIssueId: 102,
    currency,
    prices: [{ grade: 'f', price: price / 2 }, { grade: 'vf', price }],
    stale: false,
    catalogPhotos: [],
  }
}

async function wait(ms: number) {
  await new Promise((resolve) => setTimeout(resolve, ms))
  await flushPromises()
}

async function searchAndOpenFirstResult() {
  const wrapper = mount(NumistaCatalogBrowser)
  await wrapper.get('#catalog-query').setValue('50 escudos')
  await wrapper.get('#catalog-year').setValue('1968')
  await wrapper.get('form').trigger('submit')
  await flushPromises()
  await wrapper.get('ul li button').trigger('click')
  await flushPromises()
  return wrapper
}

function buttonWithText(wrapper: ReturnType<typeof mount>, text: string) {
  const button = wrapper.findAll('button').find((item) => item.text().includes(text))
  if (!button) throw new Error(`No button with text ${text}`)
  return button
}

describe('NumistaCatalogBrowser', () => {
  beforeEach(() => {
    Object.values(mocks).forEach((mock) => mock.mockReset())
    localStorage.clear()
    setLocale('en')
    URL.createObjectURL = vi.fn(() => 'blob:catalog-photo')
    mocks.proxyImage.mockResolvedValue({ data: new Blob(['x']) })
    mocks.searchCoinPhotos.mockResolvedValue({ data: { status: 'ok', query: '', photos: [] } })
    mocks.searchNumistaCatalog.mockResolvedValue({
      data: { status: 'ok', count: 45, page: 1, types: [{ id: 5, title: '50 Escudos', issuer: { code: 'portugal', name: 'Portugal' }, minYear: 1968, maxYear: 1972 }] },
    })
    mocks.getNumistaCatalogType.mockResolvedValue({
      data: { status: 'ok', type: TYPE, issues: [{ id: 101, year: 1967 }, { id: 102, year: 1968, mintage: 1000000 }] },
    })
  })

  afterEach(() => setLocale('en'))

  it('suggests countries and searches within the chosen one', async () => {
    mocks.searchNumistaIssuers.mockResolvedValue({ data: { status: 'ok', issuers: [{ code: 'portugal', name: 'Portugal' }, { code: 'portuguese_india', name: 'Portuguese India', parent: 'India' }] } })
    const wrapper = mount(NumistaCatalogBrowser)

    await wrapper.get('#catalog-country').setValue('port')
    await wait(300)

    expect(mocks.searchNumistaIssuers).toHaveBeenCalledWith('port', 'en')
    const options = wrapper.findAll('[role="option"]')
    expect(options.map((option) => option.text())).toEqual(['Portugal', 'Portuguese India · India'])
    await options[0]!.trigger('mousedown')
    expect(wrapper.text()).toContain('Portugal')
    expect(wrapper.find('#catalog-country').exists()).toBe(false)

    await wrapper.get('#catalog-year').setValue('-44')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(mocks.searchNumistaCatalog).toHaveBeenCalledWith({ q: undefined, issuer: 'portugal', year: -44, page: 1, lang: 'en' })
    expect(wrapper.text()).toContain('45 coins found')
    expect(wrapper.text()).toContain('50 Escudos')
    expect(wrapper.text()).toContain('Portugal · 1968–1972')

    await buttonWithText(wrapper, 'Show more').trigger('click')
    await flushPromises()
    expect(mocks.searchNumistaCatalog).toHaveBeenLastCalledWith(expect.objectContaining({ page: 2 }))
  })

  it('previews a type and hands back the chosen issue, grade and USD price', async () => {
    mocks.getNumistaTypeMarketValue.mockResolvedValue({ data: marketValue('USD', 30) })
    const wrapper = await searchAndOpenFirstResult()

    expect(mocks.getNumistaCatalogType).toHaveBeenCalledWith(5, 'en')
    expect(mocks.getNumistaTypeMarketValue).toHaveBeenCalledWith(5, { currency: 'USD', issueId: undefined, year: 1968 })
    expect(wrapper.text()).toContain('Silver (.650)')
    expect(wrapper.text()).toContain('KM 599')
    expect(wrapper.text()).toContain('© Ana Example · CC BY-NC')
    expect((wrapper.get('select').element as HTMLSelectElement).value).toBe('102')

    const vf = wrapper.findAll('[aria-pressed]').find((card) => card.text().startsWith('VF'))!
    await vf.trigger('click')
    expect(vf.attributes('aria-pressed')).toBe('true')

    await buttonWithText(wrapper, 'Add to collection').trigger('click')
    await flushPromises()

    const [selection] = wrapper.emitted('use')![0] as [Record<string, unknown>]
    expect(selection).toMatchObject({ grade: 'vf', currentValue: 30, wishlist: false, issue: { id: 102 }, photos: {} })
    expect(mocks.searchCoinPhotos).toHaveBeenCalledWith({ q: '50 Escudos 1968', numistaTypeId: 5, lang: 'en' })
    expect((selection.type as NumistaCatalogType).id).toBe(5)
  })

  it('uses the USD price for the current value when euros are shown, and can add to the wishlist', async () => {
    setLocale('pt-PT')
    mocks.getNumistaTypeMarketValue.mockImplementation((_id: number, params: { currency: string }) =>
      Promise.resolve({ data: params.currency === 'USD' ? marketValue('USD', 33) : marketValue('EUR', 30) }))
    const wrapper = await searchAndOpenFirstResult()

    const vf = wrapper.findAll('[aria-pressed]').find((card) => card.text().startsWith('VF'))!
    await vf.trigger('click')
    await buttonWithText(wrapper, 'Quero ter').trigger('click')
    await flushPromises()

    expect(mocks.getNumistaTypeMarketValue).toHaveBeenLastCalledWith(5, { currency: 'USD', issueId: 102 })
    const [selection] = wrapper.emitted('use')![0] as [Record<string, unknown>]
    expect(selection).toMatchObject({ grade: 'vf', currentValue: 33, wishlist: true })
  })

  it('explains provider problems instead of results', async () => {
    mocks.searchNumistaCatalog.mockResolvedValue({ data: { status: 'unconfigured', count: 0, page: 1, types: [] } })
    const wrapper = mount(NumistaCatalogBrowser)

    await wrapper.get('#catalog-query').setValue('denarius')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Add a Numista API key in Admin')
  })
})
