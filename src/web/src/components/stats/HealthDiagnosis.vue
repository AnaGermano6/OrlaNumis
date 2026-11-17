<template>
  <section class="card" aria-labelledby="health-diagnosis-title" data-testid="health-diagnosis">
    <h2 id="health-diagnosis-title" class="mb-5 mt-0 flex items-center gap-2 text-lg">
      {{ t('stats.health.diagnosis') }}
      <span class="inline-flex text-text-muted" :title="weightsHint" :aria-label="weightsHint" role="img" data-testid="health-weights-hint">
        <Info :size="18" aria-hidden="true" />
      </span>
    </h2>

    <div class="grid items-center gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)_auto] lg:divide-x lg:divide-border-subtle">
      <!-- The score as a gauge, out of 100, with its grade. -->
      <div class="flex flex-col items-center justify-center rounded-md bg-gold-glow px-6 py-6" data-testid="health-gauge">
        <svg viewBox="0 0 120 68" class="w-full max-w-56" aria-hidden="true">
          <path :d="ARC" fill="none" stroke="var(--accent-gold-dim)" stroke-width="9" stroke-linecap="round" />
          <path :d="ARC" fill="none" stroke="var(--accent-gold)" stroke-width="9" stroke-linecap="round" :stroke-dasharray="`${(summary.score / 100) * ARC_LENGTH} ${ARC_LENGTH}`" />
        </svg>
        <p class="-mt-10 mb-0 font-display text-2xl font-semibold text-text-primary" data-testid="health-score">{{ t('stats.health.outOf', { score: summary.score }) }}</p>
        <p class="m-0 mt-1 text-label font-semibold uppercase tracking-[0.08em] text-text-secondary">{{ t('stats.health.grade', { grade: summary.grade }) }}</p>
      </div>

      <!-- Each part of the score. -->
      <ul class="m-0 flex list-none flex-col gap-4 p-0 lg:pl-6">
        <li v-for="part in parts" :key="part.key" :data-testid="`health-part-${part.key}`">
          <span class="flex items-baseline justify-between gap-3 text-body">
            <span class="text-text-primary">{{ part.label }}</span>
            <span class="font-semibold text-text-primary">{{ formatShare(part.value / 100) }}</span>
          </span>
          <span class="mt-1.5 block h-2 overflow-hidden rounded-full bg-input" aria-hidden="true">
            <span class="block h-full rounded-full" :style="{ width: `${part.value}%`, background: part.color }" />
          </span>
        </li>
      </ul>

      <!-- What to do about each: the priority list, of that part only. -->
      <ul class="m-0 flex list-none flex-col items-start gap-3 p-0 lg:items-end lg:pl-6">
        <li v-for="part in parts" :key="part.key">
          <button
            type="button"
            class="inline-flex cursor-pointer items-center gap-1.5 rounded-sm text-base text-text-primary hover:text-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
            :class="focused === part.dimension ? 'text-gold' : ''"
            :aria-pressed="focused === part.dimension"
            :data-testid="`health-focus-${part.dimension}`"
            @click="emit('focus', focused === part.dimension ? null : part.dimension)"
          >
            {{ part.action }} <ArrowRight :size="16" aria-hidden="true" />
          </button>
        </li>
      </ul>
    </div>
  </section>
</template>

<script setup lang="ts">
// The collection's health at a glance: its score as a gauge, out of 100, with
// its grade; each part of it (metadata, images, valuations, AI analyses when
// an AI provider counts them); and for each, a way to the coins that miss
// something of it. The ⓘ says how much each part weighs.
import { computed } from 'vue'
import { ArrowRight, Info } from 'lucide-vue-next'
import type { CollectionHealthSummary, HealthChecklistDimension } from '@/types'
import { t } from '@/i18n'
import { formatShare } from '@/utils/numberFormat'

const props = defineProps<{
  summary: CollectionHealthSummary
  /** The part the priority list shows, if one. */
  focused: HealthChecklistDimension | null
}>()

const emit = defineEmits<{ focus: [dimension: HealthChecklistDimension | null] }>()

/** Half a circle, left to right over the top. */
const ARC = 'M 10 60 A 50 50 0 0 1 110 60'
const ARC_LENGTH = Math.PI * 50

const PARTS = [
  { key: 'metadata', dimension: 'metadata', label: () => t('health.dimensions.metadata'), action: () => t('stats.health.focus.metadata'), color: 'var(--cat-extra-4)' },
  { key: 'imageCoverage', dimension: 'images', label: () => t('stats.health.dimensions.imageCoverage'), action: () => t('stats.health.focus.images'), color: 'var(--cat-modern)' },
  { key: 'valuationFreshness', dimension: 'valuation', label: () => t('stats.health.dimensions.valuationFreshness'), action: () => t('stats.health.focus.valuation'), color: 'var(--color-positive)' },
  { key: 'aiCoverage', dimension: 'ai', label: () => t('stats.health.dimensions.aiCoverage'), action: () => t('stats.health.focus.ai'), color: 'var(--cat-roman)' },
] as const

// A part with no weight (AI analyses, while no AI provider is set up) does not count, so it is not shown.
const parts = computed(() => PARTS
  .filter((part) => (props.summary.weights?.[part.key] ?? 1) !== 0)
  .map((part) => ({
    key: part.key,
    dimension: part.dimension as HealthChecklistDimension,
    label: part.label(),
    action: part.action(),
    color: part.color,
    value: props.summary.dimensions?.[part.key] ?? 0,
  })))

const weightsHint = computed(() => t('stats.health.weightsHint', {
  weights: PARTS
    .filter((part) => (props.summary.weights?.[part.key] ?? 0) > 0)
    .map((part) => `${part.label()} ${props.summary.weights[part.key]}%`)
    .join(', '),
}))
</script>
