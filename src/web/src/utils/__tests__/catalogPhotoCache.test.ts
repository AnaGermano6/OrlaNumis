import { beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({ proxyImage: vi.fn() }))
vi.mock('@/api/client', () => mocks)

describe('loadCatalogPhoto', () => {
  beforeEach(() => {
    vi.resetModules()
    mocks.proxyImage.mockReset()
    URL.createObjectURL = vi.fn((blob: Blob) => `blob:${(blob as Blob & { name?: string }).name ?? 'x'}`)
  })

  it('loads at most two photos at a time and reuses finished ones', async () => {
    const { loadCatalogPhoto } = await import('@/utils/catalogPhotoCache')
    const resolvers: Array<() => void> = []
    mocks.proxyImage.mockImplementation(() => new Promise((resolve) => resolvers.push(() => resolve({ data: new Blob(['x']) }))))

    const loads = ['a', 'b', 'c'].map((name) => loadCatalogPhoto(`https://upload.wikimedia.org/${name}.jpg`))
    await Promise.resolve()
    expect(mocks.proxyImage).toHaveBeenCalledTimes(2)

    resolvers[0]!()
    await loads[0]
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(mocks.proxyImage).toHaveBeenCalledTimes(3)

    resolvers[1]!()
    resolvers[2]!()
    await Promise.all(loads)
    await loadCatalogPhoto('https://upload.wikimedia.org/a.jpg')
    expect(mocks.proxyImage).toHaveBeenCalledTimes(3)
  })

  it('forgets failures so a later view can try again', async () => {
    const { loadCatalogPhoto } = await import('@/utils/catalogPhotoCache')
    mocks.proxyImage.mockRejectedValueOnce(new Error('429')).mockResolvedValueOnce({ data: new Blob(['x']) })

    expect(await loadCatalogPhoto('https://upload.wikimedia.org/a.jpg')).toBeNull()
    expect(await loadCatalogPhoto('https://upload.wikimedia.org/a.jpg')).toMatch(/^blob:/)
  })
})
