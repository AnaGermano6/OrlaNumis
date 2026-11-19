import { createPinia, setActivePinia } from 'pinia'
import { describe, expect, it, vi } from 'vitest'

// The pages themselves are not needed: only where the router goes.
vi.mock('@/pages/CollectionPage.vue', () => ({ default: { template: '<div />' } }))
vi.mock('@/pages/CollectionOverviewPage.vue', () => ({ default: { template: '<div />' } }))

describe('the page the app opens on', () => {
  it('is the new releases when the app is opened at its address alone; the gallery opens from within', async () => {
    setActivePinia(createPinia())
    const { useAuthStore } = await import('@/stores/auth')
    useAuthStore().token = 'token'
    const { default: router } = await import('@/router')

    await router.push('/')
    expect(router.currentRoute.value.name).toBe('collection-overview')
    expect(router.currentRoute.value.path).toBe('/collection')

    await router.push('/')
    expect(router.currentRoute.value.name).toBe('collection')
  })
})
