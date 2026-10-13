// The photos of an item's two sides picked in a form: a new file for each
// side (with a preview) or the removal of the one it has, for the page to
// upload or delete once the item is saved.
import { computed, onBeforeUnmount, ref, type Ref } from 'vue'
import type { Coin } from '@/types'

export type PhotoSide = 'obverse' | 'reverse'

export function useSidePhotoInputs(form: () => Partial<Coin>) {
  const files: Record<PhotoSide, Ref<File | null>> = { obverse: ref(null), reverse: ref(null) }
  const previews: Record<PhotoSide, Ref<string | null>> = { obverse: ref(null), reverse: ref(null) }
  const removedIds: Record<PhotoSide, Ref<number | null>> = { obverse: ref(null), reverse: ref(null) }
  const inputs: Record<PhotoSide, Ref<HTMLInputElement | null>> = { obverse: ref(null), reverse: ref(null) }

  function existing(side: PhotoSide) {
    return computed(() => {
      if (removedIds[side].value) return null
      return form().images?.find((image) => image.imageType === side)?.filePath ?? null
    })
  }

  function pick(side: PhotoSide, event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0]
    if (!file) return
    if (previews[side].value) URL.revokeObjectURL(previews[side].value!)
    files[side].value = file
    previews[side].value = URL.createObjectURL(file)
  }

  function clear(side: PhotoSide) {
    const current = form().images?.find((image) => image.imageType === side)
    if (current) removedIds[side].value = current.id
    if (previews[side].value) URL.revokeObjectURL(previews[side].value!)
    files[side].value = null
    previews[side].value = null
    if (inputs[side].value) inputs[side].value!.value = ''
  }

  onBeforeUnmount(() => {
    for (const side of ['obverse', 'reverse'] as const) {
      if (previews[side].value) URL.revokeObjectURL(previews[side].value!)
    }
  })

  return {
    obverseFile: files.obverse,
    reverseFile: files.reverse,
    obversePreview: previews.obverse,
    reversePreview: previews.reverse,
    removedObverseId: removedIds.obverse,
    removedReverseId: removedIds.reverse,
    obverseInput: inputs.obverse,
    reverseInput: inputs.reverse,
    existingObverse: existing('obverse'),
    existingReverse: existing('reverse'),
    pick,
    clear,
  }
}
