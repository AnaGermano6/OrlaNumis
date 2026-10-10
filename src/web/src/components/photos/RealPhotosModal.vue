<template>
  <div class="fixed inset-0 z-[1000] flex items-center justify-center bg-overlay p-4" @click.self="$emit('close')" @keydown.esc="$emit('close')">
    <div
      role="dialog"
      :aria-labelledby="titleId"
      aria-modal="true"
      class="card flex max-h-[90vh] w-full max-w-[760px] flex-col !p-0"
      tabindex="-1"
    >
      <div class="flex items-center justify-between gap-3 border-b border-border-subtle px-5 py-4">
        <div class="min-w-0">
          <h3 :id="titleId" class="m-0 font-display text-lg font-medium text-heading">{{ t('photos.title') }}</h3>
          <p class="m-0 truncate text-body text-gold">{{ coinName }}</p>
        </div>
        <button
          type="button"
          class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-sm text-text-muted transition-colors hover:bg-white/5 hover:text-text-primary"
          :title="t('common.close')"
          :aria-label="t('common.close')"
          @click="$emit('close')"
        >
          <X :size="18" />
        </button>
      </div>

      <div class="min-h-0 overflow-y-auto px-5 py-4">
        <p class="mb-3 mt-0 text-chip text-text-muted">{{ t('photos.chooseHint') }}</p>
        <RealPhotoPicker v-model="choice" :initial-query="coinName" :numista-type-id="numistaTypeId" :details="details" />
      </div>

      <div class="flex flex-wrap items-center justify-end gap-2 border-t border-border-subtle px-5 py-3">
        <p v-if="error" class="m-0 mr-auto text-body text-byzantine" role="alert">{{ error }}</p>
        <button type="button" class="btn btn-secondary" :disabled="saving" @click="$emit('close')">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary" :disabled="saving || !hasChoice" @click="save">
          {{ saving ? t('common.saving') : t('photos.save') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// Find real photos for a coin that is already in the collection and save the
// chosen ones as its obverse/reverse images, crediting the authors in its notes.
import { computed, ref, useId } from 'vue'
import { X } from 'lucide-vue-next'
import { importCoinImage } from '@/api/client'
import { t } from '@/i18n'
import RealPhotoPicker from '@/components/photos/RealPhotoPicker.vue'
import { photoCreditLine, type CoinPhotoDetails, type PhotoSide, type RealPhotoChoice } from '@/utils/coinPhotos'

const props = defineProps<{
  coinId: number
  coinName: string
  numistaTypeId: number | null
  /** Year, country… for the ECB image of a €2 commemorative. */
  details?: CoinPhotoDetails
  /** Whether the coin already has a primary image (kept as primary if so). */
  hasPrimaryImage: boolean
}>()

const emit = defineEmits<{
  close: []
  /** At least one photo was saved; the coin should be reloaded. */
  saved: []
}>()

const titleId = useId()
const choice = ref<RealPhotoChoice>({})
const saving = ref(false)
const error = ref('')

const hasChoice = computed(() => Boolean(choice.value.obverse || choice.value.reverse))

async function save() {
  saving.value = true
  error.value = ''
  const refused: string[] = []
  let saved = 0
  try {
    for (const side of ['obverse', 'reverse'] as PhotoSide[]) {
      const photo = choice.value[side]
      if (!photo) continue
      const isPrimary = !props.hasPrimaryImage && (side === 'obverse' || !choice.value.obverse)
      try {
        await importCoinImage(props.coinId, { url: photo.imageUrl, imageType: side, isPrimary, credit: photoCreditLine(side, photo) })
        saved += 1
      } catch {
        refused.push(side === 'obverse' ? t('common.obverse') : t('common.reverse'))
      }
    }
  } finally {
    saving.value = false
  }
  if (saved) emit('saved')
  if (refused.length) {
    // Stay open so the user can pick another photo for the refused side.
    error.value = t('photos.importFailed', { sides: refused.join(', ') })
    return
  }
  emit('close')
}
</script>
