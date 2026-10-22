<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-[300] flex items-center justify-center bg-overlay-60 px-4" @click="close">
      <form
        ref="dialog"
        class="w-full max-w-[420px] rounded-md border border-border-subtle bg-card p-6 shadow-[0_12px_40px_var(--overlay-50)]"
        role="dialog"
        aria-modal="true"
        aria-labelledby="export-title"
        data-testid="export-dialog"
        @click.stop
        @keydown.esc="close"
        @submit.prevent="download"
      >
        <h3 id="export-title" class="mb-1 mt-0 text-lg font-medium text-heading">{{ t('collection.export.title') }}</h3>
        <p class="mb-4 mt-0 text-body text-text-secondary">{{ t('collection.export.intro') }}</p>

        <fieldset class="mb-4 border-0 p-0">
          <legend class="form-label mb-2">{{ t('collection.export.what') }}</legend>
          <div class="flex flex-col gap-2">
            <label class="flex cursor-pointer items-center gap-2 text-body text-text-primary">
              <input v-model="scope" type="radio" name="export-scope" value="shown" class="accent-[var(--accent-gold)]" data-testid="export-scope-shown" />
              {{ shownLabel }}
            </label>
            <label class="flex cursor-pointer items-center gap-2 text-body text-text-primary">
              <input v-model="scope" type="radio" name="export-scope" value="all" class="accent-[var(--accent-gold)]" data-testid="export-scope-all" />
              {{ t('collection.export.everything') }}
            </label>
          </div>
        </fieldset>

        <fieldset class="mb-4 border-0 p-0">
          <legend class="form-label mb-2">{{ t('collection.export.format') }}</legend>
          <div class="flex flex-col gap-2">
            <label class="flex cursor-pointer items-center gap-2 text-body text-text-primary">
              <input v-model="format" type="radio" name="export-format" value="xlsx" class="accent-[var(--accent-gold)]" data-testid="export-format-xlsx" />
              {{ t('collection.export.xlsx') }}
            </label>
            <label class="flex cursor-pointer items-center gap-2 text-body text-text-primary">
              <input v-model="format" type="radio" name="export-format" value="csv" class="accent-[var(--accent-gold)]" data-testid="export-format-csv" />
              {{ t('collection.export.csv') }}
            </label>
          </div>
        </fieldset>

        <p v-if="failed" class="mb-3 mt-0 text-body text-loss" role="alert">{{ t('collection.export.failed') }}</p>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="close">{{ t('common.cancel') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="downloading" data-testid="export-download">
            <Download :size="16" aria-hidden="true" /> {{ downloading ? t('collection.export.downloading') : t('collection.export.download') }}
          </button>
        </div>
      </form>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
// Downloads a gallery's coins (those its filters keep, in its order) or the
// whole collection, coins and banknotes, as an Excel workbook or a CSV.
import { computed, nextTick, ref, watch } from 'vue'
import { Download } from 'lucide-vue-next'
import { exportCoins } from '@/api/client'
import { locale, t } from '@/i18n'

const props = withDefaults(defineProps<{
  open: boolean
  /** The gallery's: its coins, or its banknotes. */
  kind?: 'coin' | 'banknote'
  /** The gallery's filters and order, as it asks the server for its list. */
  params: Record<string, unknown>
  /** How many the gallery shows. */
  count: number
  /** Whether a filter or a search is on. */
  filtered?: boolean
}>(), { kind: 'coin', filtered: false })

const emit = defineEmits<{ close: [] }>()

const dialog = ref<HTMLElement | null>(null)
const scope = ref<'shown' | 'all'>('shown')
const format = ref<'xlsx' | 'csv'>('xlsx')
const downloading = ref(false)
const failed = ref(false)

const shownLabel = computed(() => {
  const count = props.count
  if (props.filtered) return t(props.kind === 'banknote' ? 'collection.export.filteredBanknotes' : 'collection.export.filteredCoins', { count })
  return t(props.kind === 'banknote' ? 'collection.export.allBanknotes' : 'collection.export.allCoins', { count })
})

watch(() => props.open, async (isOpen) => {
  if (!isOpen) return
  scope.value = 'shown'
  failed.value = false
  await nextTick()
  dialog.value?.querySelector<HTMLElement>('input:checked')?.focus()
})

function close() {
  if (!downloading.value) emit('close')
}

// The whole collection keeps the gallery's order, without its filters.
function requestParams(): Record<string, unknown> {
  if (scope.value === 'shown') return props.params
  const { sort, order, seed } = props.params
  return { wishlist: 'false', sort, order, seed }
}

async function download() {
  if (downloading.value) return
  downloading.value = true
  failed.value = false
  try {
    const { blob, filename } = await exportCoins(requestParams(), format.value, locale.value.startsWith('pt') ? 'pt' : 'en')
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
