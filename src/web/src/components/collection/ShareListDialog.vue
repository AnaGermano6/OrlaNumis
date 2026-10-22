<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-[300] flex items-center justify-center bg-overlay-60 px-4" @click="close">
      <form
        ref="dialog"
        class="w-full max-w-[440px] rounded-md border border-border-subtle bg-card p-6 shadow-[0_12px_40px_var(--overlay-50)]"
        role="dialog"
        aria-modal="true"
        aria-labelledby="share-list-title"
        data-testid="share-list-dialog"
        @click.stop
        @keydown.esc="close"
        @submit.prevent="download"
      >
        <h3 id="share-list-title" class="mb-1 mt-0 text-lg font-medium text-heading">{{ t('collection.shareList.title') }}</h3>
        <p class="mb-4 mt-0 text-body text-text-secondary">{{ t('collection.shareList.intro') }}</p>

        <fieldset class="mb-4 border-0 p-0">
          <legend class="form-label mb-2">{{ t('collection.shareList.what') }}</legend>
          <div class="flex flex-col gap-2">
            <label v-for="option in LISTS" :key="option" class="flex cursor-pointer items-center gap-2 text-body text-text-primary">
              <input v-model="list" type="radio" name="share-list" :value="option" class="accent-[var(--accent-gold)]" :data-testid="`share-list-${option}`" />
              {{ t(`collection.shareList.lists.${option}`) }}
            </label>
          </div>
        </fieldset>

        <fieldset class="mb-4 border-0 p-0">
          <legend class="form-label mb-2">{{ t('collection.shareList.format') }}</legend>
          <div class="flex flex-wrap gap-4">
            <label class="flex cursor-pointer items-center gap-2 text-body text-text-primary">
              <input v-model="format" type="radio" name="share-format" value="pdf" class="accent-[var(--accent-gold)]" data-testid="share-format-pdf" />
              {{ t('collection.shareList.pdf') }}
            </label>
            <label class="flex cursor-pointer items-center gap-2 text-body text-text-primary">
              <input v-model="format" type="radio" name="share-format" value="csv" class="accent-[var(--accent-gold)]" data-testid="share-format-csv" />
              {{ t('collection.shareList.csv') }}
            </label>
          </div>
        </fieldset>

        <label class="mb-4 flex cursor-pointer items-center gap-2 text-body text-text-primary">
          <input v-model="includeValues" type="checkbox" class="h-4 w-4 accent-[var(--accent-gold)]" data-testid="share-values" />
          {{ t('collection.shareList.values') }}
        </label>

        <div class="form-group mb-4">
          <label for="share-contact" class="form-label mb-1 block">{{ t('collection.shareList.contact') }}</label>
          <input id="share-contact" v-model="contact" type="text" maxlength="120" class="form-input w-full" autocomplete="email" :placeholder="t('collection.shareList.contactPlaceholder')" aria-describedby="share-contact-hint" />
          <span id="share-contact-hint" class="form-hint">{{ t('collection.shareList.contactHint') }}</span>
        </div>

        <p v-if="failed" class="mb-3 mt-0 text-body text-loss" role="alert">{{ t('collection.shareList.failed') }}</p>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="close">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="downloading" data-testid="share-download">
            <Download :size="16" aria-hidden="true" /> {{ downloading ? t('collection.shareList.downloading') : t('collection.shareList.download') }}
          </button>
        </div>
      </form>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
// Makes the file of the duplicates to trade, the wishlist, or both, to send to
// other collectors: a PDF to print or a CSV for a spreadsheet, in the app's
// language, with the values only when asked, and a contact printed at the top.
import { nextTick, ref, watch } from 'vue'
import { Download } from 'lucide-vue-next'
import { downloadShareList, type ShareListRequest } from '@/api/client'
import { locale, t } from '@/i18n'

const LISTS = ['trades', 'wishlist', 'both'] as const

const props = withDefaults(defineProps<{
  open: boolean
  /** The list chosen when it opens. */
  initialList?: ShareListRequest['list']
}>(), { initialList: 'trades' })

const emit = defineEmits<{ close: [] }>()

const dialog = ref<HTMLElement | null>(null)
const list = ref<ShareListRequest['list']>(props.initialList)
const format = ref<ShareListRequest['format']>('pdf')
const includeValues = ref(false)
const contact = ref('')
const downloading = ref(false)
const failed = ref(false)

watch(() => props.open, async (isOpen) => {
  if (!isOpen) return
  list.value = props.initialList
  failed.value = false
  await nextTick()
  dialog.value?.querySelector<HTMLElement>('input:checked')?.focus()
})

function close() {
  if (!downloading.value) emit('close')
}

async function download() {
  if (downloading.value) return
  downloading.value = true
  failed.value = false
  try {
    const { blob, filename } = await downloadShareList({
      list: list.value,
      format: format.value,
      lang: locale.value.startsWith('pt') ? 'pt' : 'en',
      includeValues: includeValues.value,
      contact: contact.value.trim(),
    })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    URL.revokeObjectURL(url)
    downloading.value = false
    emit('close')
  } catch {
    failed.value = true
    downloading.value = false
  }
}
</script>
