import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { config, mount } from '@vue/test-utils'
import { setLocale } from '@/i18n'
import MobileCollectionHeader from '@/components/collection/MobileCollectionHeader.vue'

// The filters sheet (BaseModal) renders in place, where the wrapper finds it.
config.global.stubs = { ...config.global.stubs, teleport: true }

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const headerPath = path.resolve(__dirname, '../collection/MobileCollectionHeader.vue')

describe('MobileCollectionHeader', () => {
  it('keeps select mode in the menu instead of the top row', () => {
    const source = fs.readFileSync(headerPath, 'utf8')

    expect(source).not.toContain('class="pwa-icon-btn"')
    expect(source).toContain(`<span class="section-label mb-0">{{ t('collection.menu.selection') }}</span>`)
    expect(source).toContain("{{ selectMode ? t('collection.menu.exitSelection') : t('collection.menu.enableSelection') }}")
  })

  it('does not launch Mint Map from the collection menu on a phone', () => {
    const source = fs.readFileSync(headerPath, 'utf8')

    expect(source).not.toContain('to="/mint-map"')
    expect(source).not.toContain('Mint Map')
    expect(source).not.toContain('MapPin')
  })

  function mountHeader(props: Record<string, unknown> = {}) {
    setLocale('en')
    return mount(MobileCollectionHeader, {
      props: {
        search: '', selectMode: false, menuOpen: false, selectedDenomination: '', selectedSet: '',
        userSets: [], denominationOptions: [], sortKey: 'updated_desc', viewMode: 'grid', layout: 'grid',
        gridSide: null, activeFilters: 0, total: 26, ...props,
      },
      global: { stubs: { SearchBar: true, DenominationFilter: true, SortSelect: true } },
    })
  }

  it('shows how many filters are on and opens them in a sheet', async () => {
    const wrapper = mountHeader({ activeFilters: 2 })
    expect(wrapper.get('[data-testid="gallery-filters-count"]').text()).toBe('2')
    expect(wrapper.get('[data-testid="gallery-filters-button"]').attributes('aria-label')).toContain('2 filters on')
    expect(wrapper.find('[data-testid="gallery-filters"]').exists()).toBe(false)

    await wrapper.get('[data-testid="gallery-filters-button"]').trigger('click')
    expect(wrapper.emitted('update:menuOpen')?.at(-1)).toEqual([true])

    await wrapper.setProps({ menuOpen: true })
    expect(wrapper.find('[data-testid="gallery-filters"]').exists()).toBe(true)
    await wrapper.get('[data-testid="gallery-filters-done"]').trigger('click')
    expect(wrapper.get('[data-testid="gallery-filters-done"]').text()).toBe('Show 26 coins')
    expect(wrapper.emitted('update:menuOpen')?.at(-1)).toEqual([false])
  })

  it('offers one choice of view: grid, list or swipe', async () => {
    const wrapper = mountHeader({ menuOpen: true })
    expect(wrapper.get('[data-testid="gallery-view-grid"]').attributes('aria-pressed')).toBe('true')

    await wrapper.get('[data-testid="gallery-view-list"]').trigger('click')
    expect(wrapper.emitted('update:viewMode')?.at(-1)).toEqual(['grid'])
    expect(wrapper.emitted('update:layout')?.at(-1)).toEqual(['list'])

    await wrapper.get('[data-testid="gallery-view-swipe"]').trigger('click')
    expect(wrapper.emitted('update:viewMode')?.at(-1)).toEqual(['swipe'])
  })
})
