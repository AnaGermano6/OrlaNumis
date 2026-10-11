<template>
  <div class="container">
    <div class="form-wrapper">
      <DesktopPageContext :title="t('editCoin.title')">
        <div class="page-header">
          <h1>{{ t('editCoin.title') }}</h1>
        </div>
      </DesktopPageContext>
      <div v-if="loading" class="loading-overlay">
        <div class="spinner"></div>
      </div>
      <CoinForm v-else ref="coinFormRef" :form="form" :coin-id="form.id" :submit-label="t('common.saveChanges')" :loading="saving" @submit="handleSubmit" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CoinForm from '@/components/CoinForm.vue'
import { getCoin, updateCoin, uploadImage, deleteImage, extractText } from '@/api/client'
import type { Coin } from '@/types'
import { useDialog } from '@/composables/useDialog'
import { useQuickAccess } from '@/composables/useQuickAccess'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import { t } from '@/i18n'

const { showAlert } = useDialog()
const { refresh: refreshQuickAccess } = useQuickAccess()
const route = useRoute()
const router = useRouter()
const loading = ref(true)
const saving = ref(false)
const coinFormRef = ref<InstanceType<typeof CoinForm> | null>(null)

const form = reactive<Partial<Coin>>({})

onMounted(async () => {
  const id = Number(route.params['id'])
  try {
    const res = await getCoin(id)
    Object.assign(form, res.data)
    if (form.purchaseDate) {
      form.purchaseDate = form.purchaseDate.substring(0, 10)
    }
  } catch {
    await showAlert(t('editCoin.loadFailed'), { title: t('common.error') })
    router.push('/')
  } finally {
    loading.value = false
  }
})

async function handleSubmit() {
  saving.value = true
  // Tracks how far the save got, so a failure reports what actually happened.
  // The coin's fields and its images are saved by separate requests, and the
  // image ones can fail on their own — telling the user the whole edit was
  // lost when only the upload failed sends them back to redo work that saved.
  let coinSaved = false
  let cleanupPending = false
  try {
    await updateCoin(form.id!, form)
    coinSaved = true

    const formComp = coinFormRef.value
    const coinId = form.id!
    async function removeImage(imageId: number) {
      const response = await deleteImage(coinId, imageId)
      cleanupPending ||= response.data.cleanupPending === true
    }

    // Replace a side by uploading the new image FIRST, then deleting the one
    // it supersedes. Deleting first means a failed upload leaves the coin with
    // no image at all, which is how a dropped upload turned into lost data.
    // There is no unique constraint on (coin_id, image_type), so the coin
    // holding both briefly is fine, and uploading the obverse as primary
    // clears the old primary flag before the old row goes away.
    async function replaceSide(
      imageType: 'obverse' | 'reverse',
      file: File,
      isPrimary: boolean,
      removedId: number | null,
    ) {
      await uploadImage(coinId, file, imageType, isPrimary)
      const superseded = removedId ?? form.images?.find((i) => i.imageType === imageType)?.id
      if (superseded) {
        await removeImage(superseded)
      }
    }

    if (formComp?.obverseFile) {
      await replaceSide('obverse', formComp.obverseFile, true, formComp.removedObverseId)
    } else if (formComp?.removedObverseId) {
      await removeImage(formComp.removedObverseId)
    }

    if (formComp?.reverseFile) {
      await replaceSide('reverse', formComp.reverseFile, false, formComp.removedReverseId)
    } else if (formComp?.removedReverseId) {
      await removeImage(formComp.removedReverseId)
    }

    // Extract text from store card if uploaded
    if (formComp?.cardFile) {
      try {
        const res = await extractText(formComp.cardFile)
        const extractedText = res.data.text
        if (extractedText) {
          const existingNotes = form.notes || ''
          const updatedNotes = existingNotes
            ? `${existingNotes}\n\n--- Store Card ---\n${extractedText}`
            : `--- Store Card ---\n${extractedText}`
          await updateCoin(coinId, { notes: updatedNotes })
        }
      } catch {
        console.warn('Card text extraction failed – coin saved without card notes')
      }
    }

    await refreshQuickAccess()
    if (cleanupPending) {
      await showAlert(t('editCoin.cleanupPending'), { title: t('editCoin.cleanupPendingTitle') })
    }
    router.back()
  } catch (error: unknown) {
    const code = (error as { response?: { data?: { code?: string } } })?.response?.data?.code
    if (!coinSaved && code === 'slot_occupied') {
      await showAlert(t('addCoin.slotTaken'), { title: t('addCoin.slotUnavailable') })
      return
    }
    await showAlert(
      coinSaved
        ? t('editCoin.imageNotUploaded')
        : t('editCoin.updateFailed'),
      { title: t('common.error') },
    )
  } finally {
    saving.value = false
  }
}
</script>
