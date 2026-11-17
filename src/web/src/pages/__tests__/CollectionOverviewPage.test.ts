import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import CollectionOverviewPage from '../CollectionOverviewPage.vue'
import { setLocale } from '@/i18n'

describe('CollectionOverviewPage', () => {
  it('is the new releases page: its title, the notices and the releases; the collection at a glance is in the statistics', () => {
    setLocale('pt-PT')
    const wrapper = mount(CollectionOverviewPage, { global: { stubs: { CoinReleases: true, NumistaValuesNotice: true, UpdateNotice: true } } })

    expect(wrapper.get('h1').text()).toBe('Novos lançamentos')
    expect(wrapper.find('update-notice-stub').exists()).toBe(true)
    expect(wrapper.find('numista-values-notice-stub').exists()).toBe(true)
    expect(wrapper.find('coin-releases-stub').exists()).toBe(true)
    expect(wrapper.find('[role="tablist"]').exists()).toBe(false)
  })
})
