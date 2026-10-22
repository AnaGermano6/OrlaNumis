<template>
  <section class="card p-6" aria-labelledby="numista-fill-title" data-testid="numista-fill">
    <div class="mb-3 flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
      <div>
        <p class="section-label">{{ t('admin.numistaFill.eyebrow') }}</p>
        <h2 id="numista-fill-title" class="m-0 text-xl font-medium text-heading">{{ t('admin.numistaFill.title') }}</h2>
      </div>
      <button
        type="button"
        class="btn btn-secondary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        :disabled="loading || applying"
        @click="load"
      >
        {{ t('admin.numistaFill.refresh') }}
      </button>
    </div>

    <p class="mb-5 text-base text-text-secondary">{{ t('admin.numistaFill.intro') }}</p>

    <p v-if="loading" class="text-body text-text-secondary" role="status">{{ t('admin.numistaFill.loading') }}</p>
    <p v-else-if="loadError" class="text-body text-byzantine" role="alert">{{ loadError }}</p>

    <template v-else-if="preview">
      <p v-if="!preview.coins && !preview.corrections.length" class="text-body text-text-secondary" data-testid="numista-fill-nothing">
        {{ t('admin.numistaFill.nothingToDo') }}
      </p>

      <div v-if="preview.coins" class="rounded-md border border-border-subtle bg-card-hover p-4" data-testid="numista-fill-summary">
        <h3 class="m-0 text-lg font-medium text-heading">{{ t('admin.numistaFill.coinsToFill', { count: preview.coins }) }}</h3>
        <dl class="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-5">
          <div v-for="field in fieldCounts" :key="field.label" class="rounded-sm bg-input px-3 py-2">
            <dt class="text-body text-text-secondary">{{ field.label }}</dt>
            <dd class="m-0 text-lg font-medium tabular-nums" :class="field.count ? 'text-text-primary' : 'text-text-secondary'">{{ field.count }}</dd>
          </div>
        </dl>

        <template v-if="materialCounts.length">
          <h4 class="mb-2 mt-4 text-body font-medium text-text-secondary">{{ t('admin.numistaFill.byMaterial') }}</h4>
          <ul class="m-0 flex list-none flex-wrap gap-[0.35rem] p-0">
            <li v-for="item in materialCounts" :key="item.material" class="chip-sm">
              {{ materialLabel(item.material) }} <span class="tabular-nums text-text-secondary">{{ item.count }}</span>
            </li>
          </ul>
        </template>

        <details class="mt-4">
          <summary class="cursor-pointer text-body text-gold">{{ t('admin.numistaFill.showChanges', { count: preview.changes.length }) }}</summary>
          <ul class="m-0 mt-2 flex max-h-80 list-none flex-col gap-1 overflow-y-auto p-0 pr-1" data-testid="numista-fill-changes">
            <li v-for="change in preview.changes" :key="change.coinId" class="rounded-sm bg-input px-3 py-2">
              <div class="flex flex-wrap items-baseline gap-x-2">
                <strong class="text-text-primary [overflow-wrap:anywhere]">{{ change.name }}</strong>
                <span v-if="change.isWishlist" class="chip-sm">{{ t('admin.numistaFill.wishlist') }}</span>
                <span class="text-body text-text-secondary">{{ changeSummary(change) }}</span>
              </div>
              <p v-if="change.composition" class="m-0 text-sm text-text-secondary [overflow-wrap:anywhere]">
                {{ t('admin.numistaFill.numistaSays', { composition: change.composition }) }}
              </p>
            </li>
          </ul>
        </details>
      </div>

      <p v-if="preview.unmatched" class="mt-3 text-body text-text-secondary">{{ t('admin.numistaFill.unmatched', { count: preview.unmatched }) }}</p>
      <p v-if="preview.notCached" class="mt-3 text-body text-text-secondary">{{ t('admin.numistaFill.notCached', { count: preview.notCached }) }}</p>

      <div v-if="preview.corrections.length" class="mt-4 rounded-md border border-border-subtle bg-card-hover p-4" data-testid="numista-fill-corrections">
        <h3 class="m-0 text-lg font-medium text-heading">{{ t('admin.numistaFill.correctionsTitle') }}</h3>
        <p class="mt-1 text-body text-text-secondary">{{ t('admin.numistaFill.correctionsIntro', { count: preview.corrections.length }) }}</p>
        <details class="mt-3">
          <summary class="cursor-pointer text-body text-gold">{{ t('admin.numistaFill.showChanges', { count: preview.corrections.length }) }}</summary>
          <ul class="m-0 mt-2 flex max-h-80 list-none flex-col gap-1 overflow-y-auto p-0 pr-1">
            <li v-for="correction in preview.corrections" :key="correction.coinId" class="rounded-sm bg-input px-3 py-2">
              <div class="flex flex-wrap items-baseline gap-x-2">
                <strong class="text-text-primary [overflow-wrap:anywhere]">{{ correction.name }}</strong>
                <span v-if="correction.isWishlist" class="chip-sm">{{ t('admin.numistaFill.wishlist') }}</span>
                <span class="text-body text-text-secondary">{{ materialLabel(correction.previousMaterial) }} → {{ materialLabel(correction.material) }}</span>
              </div>
              <p v-if="correction.composition" class="m-0 text-sm text-text-secondary [overflow-wrap:anywhere]">
                {{ t('admin.numistaFill.numistaSays', { composition: correction.composition }) }}
              </p>
            </li>
          </ul>
        </details>
        <label class="mt-3 flex items-center gap-2 text-body text-text-primary">
          <input v-model="withCorrections" type="checkbox" class="h-4 w-4 accent-[var(--accent-gold)]" :disabled="applying" data-testid="numista-fill-corrections-toggle" />
          {{ t('admin.numistaFill.correctionsToggle', { count: preview.corrections.length }) }}
        </label>
      </div>

      <div v-if="toSave" class="mt-5 flex flex-wrap items-center gap-3">
        <button
          type="button"
          class="btn btn-primary focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :disabled="applying"
          data-testid="numista-fill-apply"
          @click="apply"
        >
          {{ applying ? t('admin.numistaFill.applying') : t('admin.numistaFill.apply', { count: toSave }) }}
        </button>
      </div>
    </template>

    <p v-if="message" class="mt-3 text-body" :class="messageError ? 'text-byzantine' : 'text-gold'" role="status">{{ message }}</p>
  </section>
</template>

<script setup lang="ts">
// Completes the coins' material, weight, diameter and coin card from the
// Numista types the app has cached: shows what would change, and saves it
// once confirmed. Materials that differ from Numista change only when the
// admin ticks the corrections.
import { computed, onMounted, ref } from 'vue'
import { applyNumistaFill, previewNumistaFill, type CoinNumistaFillChange, type CoinNumistaFillResult } from '@/api/client'
import { useDialog } from '@/composables/useDialog'
import { intlLocale, t } from '@/i18n'
import { presentationLabel } from '@/utils/coinAttributes'
import { materialLabel } from '@/utils/format'

const { showConfirm } = useDialog()

const preview = ref<CoinNumistaFillResult | null>(null)
const loading = ref(false)
const loadError = ref('')
const applying = ref(false)
const withCorrections = ref(false)
const message = ref('')
const messageError = ref(false)

const fieldCounts = computed(() => {
  const result = preview.value
  if (!result) return []
  return [
    { label: t('admin.numistaFill.fieldMaterial'), count: result.materials },
    { label: t('admin.numistaFill.fieldWeight'), count: result.weights },
    { label: t('admin.numistaFill.fieldDiameter'), count: result.diameters },
    { label: t('admin.numistaFill.fieldCoincard'), count: result.presentations },
    { label: t('coinFields.monetarySystem'), count: result.monetarySystems ?? 0 },
  ]
})

const materialCounts = computed(() =>
  Object.entries(preview.value?.byMaterial ?? {})
    .map(([material, count]) => ({ material, count }))
    .sort((a, b) => b.count - a.count || materialLabel(a.material).localeCompare(materialLabel(b.material), intlLocale())),
)

// The coins a save changes: those filled, and the corrected ones when ticked.
const toSave = computed(() => {
  const result = preview.value
  if (!result) return 0
  const ids = new Set(result.changes.map((change) => change.coinId))
  if (withCorrections.value) for (const correction of result.corrections) ids.add(correction.coinId)
  return ids.size
})

function number(value: number): string {
  return value.toLocaleString(intlLocale(), { maximumFractionDigits: 2 })
}

function changeSummary(change: CoinNumistaFillChange): string {
  const parts: string[] = []
  if (change.material) parts.push(materialLabel(change.material))
  if (change.weightGrams !== undefined) parts.push(`${number(change.weightGrams)} g`)
  if (change.diameterMm !== undefined) parts.push(`${number(change.diameterMm)} mm`)
  if (change.presentation) parts.push(presentationLabel(change.presentation))
  if (change.monetarySystem) parts.push(change.monetarySystem)
  return parts.join(' · ')
}

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    preview.value = (await previewNumistaFill()).data
  } catch {
    loadError.value = t('admin.numistaFill.loadFailed')
  } finally {
    loading.value = false
  }
}

async function apply() {
  const count = toSave.value
  if (!count || applying.value) return
  const confirmed = await showConfirm(t('admin.numistaFill.confirmMessage', { count }), {
    title: t('admin.numistaFill.confirmTitle'),
    confirmLabel: t('admin.numistaFill.confirmLabel'),
  })
  if (!confirmed) return
  applying.value = true
  message.value = ''
  try {
    await applyNumistaFill(withCorrections.value)
    messageError.value = false
    message.value = t('admin.numistaFill.done', { count })
    withCorrections.value = false
    await load()
  } catch {
    messageError.value = true
    message.value = t('admin.numistaFill.applyFailed')
  } finally {
    applying.value = false
  }
}

onMounted(load)
</script>
