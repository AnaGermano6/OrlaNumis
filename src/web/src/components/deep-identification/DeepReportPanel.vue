<template>
  <section class="grid min-w-0 gap-4 overflow-hidden" :aria-label="t('deepAnalysis.report.label')">
    <div v-if="report.partialSuccess" class="rounded-sm border border-byzantine bg-card p-3" role="status">
      <p class="m-0 text-sm font-semibold text-byzantine">{{ t('deepAnalysis.report.partial') }}</p>
      <p class="m-0 text-sm text-text-secondary">
        {{ t('deepAnalysis.report.partialHint') }}
      </p>
    </div>

    <div
      v-if="report.quickLookupOutcome === 'unavailable'"
      class="rounded-sm border border-warning/30 bg-[color-mix(in_srgb,var(--text-warning)_10%,transparent)] p-3"
      role="status"
    >
      <p class="m-0 text-sm font-semibold text-warning">{{ t('deepAnalysis.report.quickIncomplete') }}</p>
      <p class="m-0 text-sm text-text-secondary">
        {{ t('deepAnalysis.report.quickIncompleteHint') }}
      </p>
    </div>
    <p v-else-if="report.quickLookupOutcome === 'no_data'" class="m-0 text-sm text-text-muted">
      {{ t('deepAnalysis.report.quickNoData') }}
    </p>

    <div class="grid min-w-0 gap-2">
      <h3 class="m-0 text-lg font-semibold text-text-primary">{{ t('deepAnalysis.report.narrative') }}</h3>
      <p class="m-0 whitespace-pre-line break-words text-body text-text-secondary [overflow-wrap:anywhere]">{{ safeNarrative }}</p>
    </div>

    <DeepProviderCoverageList :coverage="report.coverage" />

    <div v-if="report.face_analyses?.length" class="grid min-w-0 gap-2">
      <h3 class="m-0 text-lg font-semibold text-text-primary">{{ t('deepAnalysis.report.faceAnalysisTitle') }}</h3>
      <details
        v-for="face in report.face_analyses"
        :key="face.role"
        class="min-w-0 overflow-hidden rounded-sm border border-border-subtle bg-card p-3"
      >
        <summary class="cursor-pointer select-none text-label font-semibold uppercase tracking-[0.08em] text-text-muted">
          {{ t('deepAnalysis.report.faceAnalysis', { role: faceRoleLabel(face.role) }) }}
        </summary>
        <p
          v-if="face.status === 'completed'"
          class="mb-0 mt-3 whitespace-pre-line break-words text-sm text-text-secondary [overflow-wrap:anywhere]"
        >
          {{ face.narrative }}
        </p>
        <p v-else class="mb-0 mt-3 text-sm text-text-secondary">
          {{ face.limitation || t('deepAnalysis.report.faceFailed') }}
        </p>
      </details>
    </div>

    <div v-if="report.disagreements?.length" class="grid gap-2">
      <h3 class="m-0 text-lg font-semibold text-text-primary">{{ t('deepAnalysis.report.disagreements') }}</h3>
      <ul class="m-0 grid gap-3 p-0" style="list-style: none;">
        <li
          v-for="disagreement in report.disagreements"
          :key="disagreement.field"
          class="grid min-w-0 gap-1 rounded-sm border border-border-subtle bg-card p-3"
        >
          <p class="m-0 text-sm font-semibold text-text-primary">
            {{ disagreement.field }}
            <span class="text-text-secondary">({{ disagreement.resolution }})</span>
          </p>
          <ul class="m-0 grid gap-1 p-0" style="list-style: none;">
            <li v-for="(claim, index) in disagreement.claims" :key="index" class="break-words text-sm text-text-secondary [overflow-wrap:anywhere]">
              {{ claim.value }}
              <a v-if="claim.citation" :href="claim.citation" target="_blank" rel="noopener noreferrer" class="text-gold underline">{{ t('deepAnalysis.report.source') }}</a>
              <span v-else class="chip-sm" :title="t('deepAnalysis.report.fromImagesTitle')">
                {{ t('deepAnalysis.report.fromImages') }}
              </span>
            </li>
          </ul>
        </li>
      </ul>
    </div>

    <details
      v-if="report.image_hypothesis"
      class="min-w-0 overflow-hidden rounded-sm border border-border-subtle bg-card p-3"
    >
      <summary class="cursor-pointer select-none text-label font-semibold uppercase tracking-[0.08em] text-text-muted">
        {{ t('deepAnalysis.report.hypothesis') }}
      </summary>

      <div class="mt-3 grid min-w-0 gap-3">
        <p v-if="!report.image_hypothesis.legible" class="m-0 text-sm text-text-secondary">
          {{ t('deepAnalysis.report.notLegible') }}
        </p>

        <template v-else>
          <p v-if="!hypothesisFieldEntries.length" class="m-0 text-sm text-text-muted">
            {{ t('deepAnalysis.report.noFindings') }}
          </p>
          <ul v-else class="m-0 grid gap-2 p-0" style="list-style: none;">
            <li
              v-for="entry in hypothesisFieldEntries"
              :key="entry.name"
              class="grid min-w-0 gap-1"
            >
              <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ formatFieldName(entry.name) }}</span>
              <p class="m-0 flex flex-wrap items-center gap-2 break-words text-sm text-text-secondary [overflow-wrap:anywhere]">
                <span class="text-text-primary">{{ entry.value }}</span>
                <span class="chip-sm">{{ t('deepAnalysis.report.confidence', { value: Math.round(entry.confidence * 100) }) }}</span>
              </p>
            </li>
          </ul>

          <p v-if="report.image_hypothesis.observations" class="m-0 break-words text-sm text-text-secondary [overflow-wrap:anywhere]">
            {{ report.image_hypothesis.observations }}
          </p>
        </template>
      </div>
    </details>

    <div v-if="report.unresolvedQuestions?.length" class="grid gap-2">
      <h3 class="m-0 text-lg font-semibold text-text-primary">{{ t('deepAnalysis.report.unresolved') }}</h3>
      <ul class="m-0 grid gap-1 pl-5">
        <li v-for="(question, index) in report.unresolvedQuestions" :key="index" class="text-sm text-text-secondary">
          {{ question }}
        </li>
      </ul>
    </div>

    <div v-if="report.attributions?.length" class="grid gap-2">
      <h3 class="m-0 text-lg font-semibold text-text-primary">{{ t('deepAnalysis.report.attribution') }}</h3>
      <ul class="m-0 grid gap-1 p-0" style="list-style: none;">
        <li
          v-for="entry in report.attributions"
          :key="entry.provider"
          class="min-w-0 overflow-hidden rounded-sm border border-border-subtle bg-card p-2"
        >
          <p class="m-0 break-words text-sm text-text-secondary [overflow-wrap:anywhere]">
            {{ entry.text }}
            <SafeExternalLink
              v-if="entry.identifier"
              :href="entry.identifier"
              class="text-gold underline"
            >
              {{ t('deepAnalysis.report.source') }}
            </SafeExternalLink>
          </p>
        </li>
      </ul>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { CoinHypothesis, DeepReport } from '@/types'
import DeepProviderCoverageList from './DeepProviderCoverageList.vue'
import SafeExternalLink from '@/components/SafeExternalLink.vue'
import { formatFieldName } from '@/utils/format'
import { t } from '@/i18n'

const props = defineProps<{ report: DeepReport }>()

// Display order mirrors contracts/vision-hypothesis.md §1 — every coin-field
// vocabulary key the hypothesis can carry, excluding the `observations`/
// `legible` metadata rendered separately below.
type HypothesisFieldName = Exclude<keyof CoinHypothesis, 'observations' | 'legible'>
const hypothesisFieldOrder: HypothesisFieldName[] = [
  'ruler',
  'denomination',
  'material',
  'mint',
  'dateRange',
  'era',
  'obverseInscription',
  'reverseInscription',
  'obverseDescription',
  'reverseDescription',
  'diameterMm',
  'weightGrams',
  'notes',
  'coin_type',
]

const hypothesisFieldEntries = computed(() => {
  const hypothesis = props.report.image_hypothesis
  if (!hypothesis) return []
  return hypothesisFieldOrder.flatMap((name) => {
    const field = hypothesis[name]
    return field ? [{ name, value: field.value, confidence: field.confidence }] : []
  })
})

function faceRoleLabel(role: string): string {
  if (role === 'obverse') return t('deepAnalysis.report.roles.obverse')
  if (role === 'reverse') return t('deepAnalysis.report.roles.reverse')
  return role
}

const safeNarrative = computed(() => {
  const narrative = props.report.narrative?.trim() ?? ''
  const transportMarkers = [
    /['"]type['"]\s*:\s*['"]thinking['"]/i,
    /['"]signature['"]\s*:/i,
    /data:image\/[^;]+;base64,/i,
    /[A-Za-z0-9+/]{500,}={0,2}/,
  ]
  if (!narrative || transportMarkers.some((marker) => marker.test(narrative))) {
    return t('deepAnalysis.report.unsafeNarrative')
  }
  return narrative
})

</script>
