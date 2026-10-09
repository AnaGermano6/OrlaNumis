<template>
  <form class="card grid gap-4" @submit.prevent="saveDraft">
    <QuickCaptureImageSlots
      v-model:obverse-image="obverseImage"
      v-model:reverse-image="reverseImage"
      v-model:detail-images="detailImages"
    />

    <div class="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(220px,1fr))] max-[600px]:grid-cols-1">
      <label class="form-group flex flex-col gap-1.5">
        <span class="section-label">{{ t('quickCapture.fields.workingTitle') }}</span>
        <input v-model="workingTitle" class="form-input" type="text" maxlength="200" :placeholder="t('quickCapture.fields.workingTitlePlaceholder')">
      </label>
      <label class="form-group flex flex-col gap-1.5">
        <span class="section-label">{{ t('quickCapture.fields.dateRange') }}</span>
        <input v-model="dateRange" class="form-input" type="text" placeholder="c. 330-335">
      </label>
      <label class="form-group flex flex-col gap-1.5">
        <span class="section-label">{{ t('quickCapture.fields.era') }}</span>
        <input v-model="era" class="form-input" type="text" placeholder="ancient">
      </label>
      <label class="form-group flex flex-col gap-1.5">
        <span class="section-label">{{ t('quickCapture.fields.acquisitionSource') }}</span>
        <input v-model="acquisitionSource" class="form-input" type="text" :placeholder="t('quickCapture.fields.acquisitionPlaceholder')">
      </label>
      <label class="form-group flex flex-col gap-1.5">
        <span class="section-label">{{ t('quickCapture.fields.purchasePrice') }}</span>
        <input v-model.number="purchasePrice" class="form-input" type="number" min="0" step="0.01">
      </label>
      <label class="form-group col-span-full flex flex-col gap-1.5">
        <span class="section-label">{{ t('quickCapture.fields.notes') }}</span>
        <textarea v-model="notes" class="form-input min-h-28 resize-y leading-[1.5]" rows="4" :placeholder="t('quickCapture.fields.notesPlaceholder')"></textarea>
      </label>
    </div>

    <p class="m-0 text-body text-text-secondary">{{ t('quickCapture.formHint') }}</p>
    <p v-if="error" class="text-base text-warning">{{ error }}</p>
    <p v-if="savedMessage" class="text-base text-text-secondary">{{ savedMessage }}</p>
    <button type="submit" class="btn btn-primary" :disabled="saving || !canSave">
      {{ saving ? t('common.saving') : t('quickCapture.saveDraft') }}
    </button>
  </form>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { createQuickCaptureDraft, getApiErrorMessage } from '@/api/client'
import type { QuickCaptureDraft } from '@/types'
import QuickCaptureImageSlots from './QuickCaptureImageSlots.vue'
import { t } from '@/i18n'

const emit = defineEmits<{
  saved: [draft: QuickCaptureDraft]
}>()

const workingTitle = ref('')
const dateRange = ref('')
const era = ref('')
const acquisitionSource = ref('')
const purchasePrice = ref<number | null>(null)
const notes = ref('')
const obverseImage = ref<File | null>(null)
const reverseImage = ref<File | null>(null)
const detailImages = ref<File[]>([])
const saving = ref(false)
const error = ref('')
const savedMessage = ref('')

const canSave = computed(() =>
  workingTitle.value.trim() !== '' ||
  notes.value.trim() !== '' ||
  Boolean(obverseImage.value || reverseImage.value || detailImages.value.length),
)

async function saveDraft() {
  if (!canSave.value || saving.value) return
  saving.value = true
  error.value = ''
  savedMessage.value = ''
  try {
    const response = await createQuickCaptureDraft({
      workingTitle: workingTitle.value,
      dateRange: dateRange.value,
      era: era.value,
      acquisitionSource: acquisitionSource.value,
      purchasePrice: purchasePrice.value,
      notes: notes.value,
      obverseImage: obverseImage.value,
      reverseImage: reverseImage.value,
      detailImages: detailImages.value,
    })
    savedMessage.value = t('quickCapture.savedActive')
    emit('saved', response.data)
  } catch (err) {
    error.value = getApiErrorMessage(err) || t('quickCapture.saveFailed')
  } finally {
    saving.value = false
  }
}
</script>
