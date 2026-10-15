<template>
  <div
    v-if="total > 0 || summary"
    class="mt-3 flex flex-wrap items-center justify-between gap-3 rounded-md border border-border-subtle bg-card px-4 py-3"
    data-testid="missing-photos"
  >
    <div class="min-w-0 flex-1">
      <p v-if="total > 0" class="m-0 font-medium text-text-primary">{{ t('collectionOverview.photos.missing', { count: total }) }}</p>
      <p v-if="running" class="m-0 text-body text-text-secondary" role="status">{{ t('collectionOverview.photos.progress', { checked, total: runTotal }) }}</p>
      <p v-else-if="summary" class="m-0 text-body text-text-secondary" role="status" data-testid="missing-photos-summary">{{ summary }}</p>
      <p v-else-if="failed" class="m-0 text-body text-loss" role="alert">{{ t('collectionOverview.photos.failed') }}</p>
      <p v-else class="m-0 text-body text-text-secondary">{{ t('collectionOverview.photos.body') }}</p>
    </div>
    <button v-if="total > 0" type="button" class="btn btn-secondary btn-sm" :disabled="running" data-testid="missing-photos-start" @click="lookForPhotos">
      <ImageIcon :size="16" /> {{ t('collectionOverview.photos.start') }}
    </button>
  </div>
</template>

<script setup lang="ts">
// The coins without any photo, and a way to give them their official images
// (the ECB's, the mints'), a batch at a time so the page tells how far it is.
import { onMounted, ref } from 'vue'
import { Image as ImageIcon } from 'lucide-vue-next'
import { addOfficialPhotosBatch, getCoinsWithoutPhotos } from '@/api/client'
import { t } from '@/i18n'

const total = ref(0)
const running = ref(false)
const checked = ref(0)
const runTotal = ref(0)
const summary = ref('')
const failed = ref(false)

async function loadCount() {
  try {
    total.value = (await getCoinsWithoutPhotos()).data.count
  } catch {
    total.value = 0
  }
}

async function lookForPhotos() {
  running.value = true
  failed.value = false
  summary.value = ''
  checked.value = 0
  runTotal.value = total.value
  let afterId = 0
  let withPhotos = 0
  try {
    for (;;) {
      const { data } = await addOfficialPhotosBatch(afterId)
      checked.value += data.checked
      withPhotos += data.withPhotos
      afterId = data.nextAfterId
      if (data.done || data.checked === 0) break
    }
    await loadCount()
    summary.value = [
      t('collectionOverview.photos.done', { count: withPhotos }),
      total.value > 0 ? t('collectionOverview.photos.left', { count: total.value }) : '',
    ].filter(Boolean).join(' ')
  } catch {
    failed.value = true
    await loadCount()
  } finally {
    running.value = false
  }
}

onMounted(loadCount)
</script>
