<template>
  <div class="container max-w-3xl">
    <DesktopPageContext :title="t('ucoinImport.title')">
      <div class="page-header">
        <h1>{{ t('ucoinImport.title') }}</h1>
      </div>
    </DesktopPageContext>

    <section class="card mb-4 flex flex-col gap-4">
      <p class="m-0 text-body text-text-secondary">{{ t('ucoinImport.intro') }}</p>

      <div v-if="!busy" class="flex flex-wrap items-center gap-3">
        <label class="btn btn-secondary inline-flex cursor-pointer focus-within:outline-2 focus-within:outline-gold focus-within:outline-offset-2">
          <FileUp :size="16" aria-hidden="true" />
          {{ t('ucoinImport.chooseFile') }}
          <input data-testid="ucoin-file" type="file" accept=".csv,.xlsx,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" class="sr-only" @change="onFile" />
        </label>
        <span v-if="importer.fileName.value" class="text-body text-text-muted">{{ importer.fileName.value }}</span>
      </div>

      <p v-if="importer.parseError.value" class="m-0 text-body text-byzantine" role="alert">
        {{ t(importer.parseError.value === 'columns' ? 'ucoinImport.errors.columns' : importer.parseError.value === 'unreadable' ? 'ucoinImport.errors.unreadable' : 'ucoinImport.errors.empty') }}
      </p>
      <p v-if="importer.loadFailed.value" class="m-0 text-body text-byzantine" role="alert">{{ t('ucoinImport.loadFailed') }}</p>

      <div v-if="importer.phase.value === 'ready'" class="flex flex-col gap-3" data-testid="ucoin-preview">
        <ul class="m-0 flex list-none flex-col gap-1 p-0 text-body text-text-primary">
          <li>
            {{ t('ucoinImport.summary.inFile', { count: importer.rows.value.length }) }}<template v-if="importer.pieces.value > importer.rows.value.length"> ({{ t('ucoinImport.summary.pieces', { count: importer.pieces.value }) }})</template>
            · {{ t('ucoinImport.summary.countries', { count: importer.countries.value }) }}
          </li>
          <li v-if="importer.alreadyImported.value">{{ t('ucoinImport.summary.already', { count: importer.alreadyImported.value }) }}</li>
          <li class="font-medium">{{ t('ucoinImport.summary.toAdd', { count: importer.pendingRows.value.length }) }}</li>
        </ul>

        <div v-if="importer.updatable.value.length" class="flex flex-col gap-2 rounded-sm border border-border-subtle p-3 md:flex-row md:items-center md:justify-between" data-testid="ucoin-updatable">
          <p class="m-0 text-body text-text-primary">{{ t('ucoinImport.summary.updatable', { count: importer.updatable.value.length }) }}</p>
          <button class="btn btn-secondary" @click="importer.fillExisting()">{{ t('ucoinImport.updateExisting') }}</button>
        </div>
        <p v-if="importer.progress.toUpdate" class="m-0 text-body text-text-secondary" role="status" data-testid="ucoin-updated">
          {{ t('ucoinImport.done.updated', { count: importer.progress.updated }) }}<template v-if="importer.progress.updateFailed"> · {{ t('ucoinImport.done.updateFailed', { count: importer.progress.updateFailed }) }}</template>
        </p>

        <label class="inline-flex items-start gap-2 text-body text-text-primary">
          <input v-model="withNumista" type="checkbox" class="mt-1" />
          <span>{{ t('ucoinImport.withNumista') }}</span>
        </label>

        <p v-if="withNumista" class="m-0 text-chip text-text-muted">{{ t('ucoinImport.estimate', { count: formatCount(importer.estimatedRequests.value) }) }}</p>

        <div>
          <button v-if="importer.pendingRows.value.length" class="btn btn-primary" @click="importer.start(withNumista)">
            {{ t('ucoinImport.import', { count: importer.pendingRows.value.length }) }}
          </button>
          <p v-else class="m-0 text-body text-text-secondary">{{ t('ucoinImport.nothingNew') }}</p>
        </div>
      </div>
    </section>

    <section
      v-if="!busy && importer.awaitingNumista.value.length && importer.phase.value !== 'ready'"
      class="card mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between"
    >
      <p class="m-0 text-body text-text-primary">{{ t('ucoinImport.awaiting', { count: importer.awaitingNumista.value.length }) }}</p>
      <button class="btn btn-secondary" @click="importer.enrich()">{{ t('ucoinImport.complete') }}</button>
    </section>

    <!-- Coins imported before photos were added on import. -->
    <section
      v-if="!busy && importer.withoutImages.value.length && importer.phase.value !== 'ready' && importer.phase.value !== 'finished'"
      class="card mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between"
      data-testid="ucoin-without-images"
    >
      <p class="m-0 text-body text-text-primary">{{ t('ucoinImport.withoutImages', { count: importer.withoutImages.value.length }) }}</p>
      <button class="btn btn-secondary" @click="importer.addMissingPhotos()">{{ t('ucoinImport.addPhotos') }}</button>
    </section>

    <!-- Coins Numista did not have before: matching improves, so they can be tried again. -->
    <section
      v-if="!busy && importer.earlierNotFound.value && importer.phase.value !== 'ready' && importer.phase.value !== 'finished'"
      class="card mb-4 flex flex-col gap-3"
      data-testid="ucoin-earlier-not-found"
    >
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <p class="m-0 text-body text-text-primary">{{ t('ucoinImport.earlierNotFound', { count: importer.earlierNotFound.value }) }}</p>
        <button class="btn btn-secondary" @click="retryNotFound">{{ t('ucoinImport.done.retryNotFound') }}</button>
      </div>
      <details class="text-body">
        <summary class="cursor-pointer text-text-secondary">{{ t('ucoinImport.done.notFoundList') }}</summary>
        <ul class="mt-2 flex list-none flex-col gap-2 p-0">
          <UcoinNotFoundCoin
            v-for="coin in importer.notFoundList.value"
            :key="coin.id"
            :coin="coin"
            :candidates-for="importer.numistaCandidatesFor"
            :choose-type="importer.chooseNumistaType"
          />
        </ul>
      </details>
    </section>

    <section v-if="busy" class="card mb-4 flex flex-col gap-3" aria-live="polite">
      <p class="m-0 text-body font-medium text-text-primary">{{ progressText }}</p>
      <div
        class="h-2 overflow-hidden rounded-full bg-input"
        role="progressbar"
        :aria-valuenow="progressDone"
        aria-valuemin="0"
        :aria-valuemax="progressTotal"
        :aria-label="progressText"
      >
        <div class="h-full rounded-full bg-gold transition-[width] duration-300" :style="{ width: `${progressPercent}%` }"></div>
      </div>
      <div class="flex flex-wrap items-center justify-between gap-3">
        <span class="text-chip text-text-muted">{{ t('ucoinImport.keepOpen') }}</span>
        <button class="btn btn-secondary btn-sm" @click="importer.stop()">{{ t('ucoinImport.stop') }}</button>
      </div>
    </section>

    <section v-if="importer.phase.value === 'finished'" class="card mb-4 flex flex-col gap-3" data-testid="ucoin-result">
      <h2 class="m-0 text-lg">{{ t(importer.stopped.value ? 'ucoinImport.done.stoppedTitle' : 'ucoinImport.done.title') }}</h2>
      <p v-if="importer.stopMessage.value" class="m-0 text-body text-byzantine" role="alert">{{ importer.stopMessage.value }}</p>
      <p v-else-if="importer.stopped.value" class="m-0 text-body text-text-secondary">{{ t('ucoinImport.done.stoppedByUser') }}</p>
      <ul class="m-0 flex list-none flex-col gap-1 p-0 text-body text-text-primary">
        <li v-if="importer.progress.toCreate">{{ t('ucoinImport.done.created', { count: importer.progress.created }) }}</li>
        <li v-if="importer.progress.createFailed" class="text-byzantine">
          {{ t('ucoinImport.done.createFailed', { count: importer.progress.createFailed, lines: importer.failedLines.value.join(', ') }) }}
        </li>
        <li v-if="importer.progress.toEnrich">{{ t('ucoinImport.done.enriched', { count: importer.progress.enriched }) }}</li>
        <li v-if="importer.progress.failed" class="text-byzantine">{{ t('ucoinImport.done.failed', { count: importer.progress.failed }) }}</li>
        <li v-if="importer.progress.notFound">{{ t('ucoinImport.done.notFound', { count: importer.progress.notFound }) }}</li>
        <li v-if="importer.progress.toPhoto">{{ t('ucoinImport.done.photos', { count: importer.progress.photographed, total: importer.progress.toPhoto }) }}</li>
      </ul>
      <p v-if="importer.progress.toPhoto" class="m-0 text-chip text-text-muted">{{ t('ucoinImport.done.photosNote') }}</p>
      <p v-if="importer.progress.enriched" class="m-0 text-chip text-text-muted">{{ t('ucoinImport.done.valueNote') }}</p>

      <details v-if="importer.notFoundCoins.value.length" class="text-body">
        <summary class="cursor-pointer text-text-secondary">{{ t('ucoinImport.done.notFoundList') }}</summary>
        <ul class="mt-2 flex list-none flex-col gap-2 p-0">
          <UcoinNotFoundCoin
            v-for="coin in importer.notFoundCoins.value"
            :key="coin.id"
            :coin="coin"
            :candidates-for="importer.numistaCandidatesFor"
            :choose-type="importer.chooseNumistaType"
          />
        </ul>
      </details>

      <div class="flex flex-wrap gap-2">
        <RouterLink to="/" class="btn btn-primary">{{ t('ucoinImport.done.viewGallery') }}</RouterLink>
        <button v-if="importer.progress.notFound" class="btn btn-secondary" @click="retryNotFound">{{ t('ucoinImport.done.retryNotFound') }}</button>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
// Imports a collection exported from uCoin.net and completes it from Numista.
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'
import { FileUp } from 'lucide-vue-next'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import UcoinNotFoundCoin from '@/components/ucoin/UcoinNotFoundCoin.vue'
import { useUcoinImport } from '@/composables/useUcoinImport'
import { intlLocale, t } from '@/i18n'

const importer = useUcoinImport()
const withNumista = ref(true)


const busy = computed(() => ['updating', 'creating', 'enriching', 'photos'].includes(importer.phase.value))

const progressTotal = computed(() => {
  if (importer.phase.value === 'updating') return importer.progress.toUpdate
  if (importer.phase.value === 'creating') return importer.progress.toCreate
  return importer.phase.value === 'photos' ? importer.progress.toPhoto : importer.progress.toEnrich
})
const progressDone = computed(() => {
  if (importer.phase.value === 'updating') return importer.progress.updated + importer.progress.updateFailed
  if (importer.phase.value === 'creating') return importer.progress.created + importer.progress.createFailed
  if (importer.phase.value === 'photos') return importer.progress.photoDone
  return importer.progress.enriched + importer.progress.notFound + importer.progress.failed
})
const progressPercent = computed(() => progressTotal.value ? Math.min(100, Math.round((progressDone.value / progressTotal.value) * 100)) : 0)
const progressKey = computed(() => {
  if (importer.phase.value === 'creating') return 'ucoinImport.creating' as const
  return importer.phase.value === 'photos' ? 'ucoinImport.addingPhotos' as const : 'ucoinImport.enriching' as const
})
const progressText = computed(() => (importer.phase.value === 'updating'
  ? t('ucoinImport.progress.updating', { done: progressDone.value, count: progressTotal.value })
  : t(progressKey.value, { done: progressDone.value, total: progressTotal.value })))

function formatCount(value: number): string {
  return new Intl.NumberFormat(intlLocale()).format(value)
}

async function onFile(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (file) await importer.readFile(file)
}

function retryNotFound() {
  importer.retryNotFound()
  void importer.enrich()
}

// Leaving mid-import would stop it halfway; ask first.
function warnBeforeUnload(event: Event) {
  if (!busy.value) return
  event.preventDefault()
}

onBeforeRouteLeave(() => !busy.value || window.confirm(t('ucoinImport.leaveConfirm')))

onMounted(() => {
  window.addEventListener('beforeunload', warnBeforeUnload)
  void importer.init()
})

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', warnBeforeUnload)
  importer.stop()
})
</script>
