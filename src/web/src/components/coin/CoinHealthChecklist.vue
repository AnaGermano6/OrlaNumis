<template>
  <div class="card !p-4">
    <div class="mb-4">
      <div
        class="inline-flex items-center gap-2 rounded-sm border px-4 py-2 text-body font-semibold"
        :class="{
          'border-status-success-border bg-status-success-bg text-gain': grade.toLowerCase() === 'a',
          'border-sky-500/30 bg-sky-500/15 text-sky-400': grade.toLowerCase() === 'b',
          'border-status-warning-border bg-status-warning-bg text-warning': grade.toLowerCase() === 'c',
          'border-orange-500/30 bg-orange-500/15 text-orange-400': grade.toLowerCase() === 'd',
          'border-status-error-border bg-status-error-bg text-loss': grade.toLowerCase() === 'f',
        }"
      >
        <span class="section-label !mb-0 opacity-80">{{ t('health.score') }}</span>
        <span class="text-lg font-bold">{{ score }}</span>
        <span class="text-label font-semibold uppercase tracking-[0.08em]">{{ grade }}</span>
      </div>
    </div>

    <div
      v-if="missingItems.length === 0"
      class="flex items-center gap-2 rounded-sm border border-status-success-border bg-status-success-bg px-3 py-3 text-body font-medium text-gain"
    >
      <CircleCheck :size="20" />
      <span>{{ t('health.allPassed') }}</span>
    </div>

    <div v-else class="flex flex-col gap-3">
      <div class="section-label !mb-0 flex items-center gap-2 text-text-secondary">
        <AlertCircle :size="16" />
        {{ t('collection.needsAttention.title') }} ({{ missingItems.length }})
      </div>

      <div class="flex flex-col gap-2">
        <div
          v-for="(item, idx) in missingItems"
          :key="idx"
          class="flex items-center gap-2 rounded-sm border border-border-subtle bg-input p-[0.6rem] md:gap-3 md:p-3"
          :class="{
            'border-l-[3px] border-l-loss': item.severity === 'high',
            'border-l-[3px] border-l-warning': item.severity === 'medium',
            'border-l-[3px] border-l-sky-400': item.severity === 'low',
          }"
        >
          <div
            class="shrink-0 text-text-muted"
            :class="{
              'text-loss': item.severity === 'high',
              'text-warning': item.severity === 'medium',
              'text-sky-400': item.severity === 'low',
            }"
          >
            <component :is="getSeverityIcon(item.severity)" :size="14" />
          </div>
          <div class="flex flex-1 flex-col gap-[0.15rem]">
            <div class="text-chip font-medium text-text-primary md:text-body">{{ item.label }}</div>
            <div class="text-sm text-text-muted">{{ formatDimension(item.dimension) }}</div>
          </div>
          <button
            v-if="item.actionHint"
            class="btn-xs btn-ghost shrink-0"
            @click="handleQuickAction(item.actionHint)"
          >
            {{ formatQuickAction(item.actionHint) }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { AlertCircle, CircleCheck, AlertTriangle, Info, AlertOctagon } from 'lucide-vue-next'
import type { HealthGrade, MissingChecklistItem, HealthQuickAction } from '@/types'
import { t } from '@/i18n'

defineProps<{
  score: number
  grade: HealthGrade
  missingItems: MissingChecklistItem[]
}>()

const emit = defineEmits<{
  quickAction: [action: HealthQuickAction]
}>()

function getSeverityIcon(severity: string) {
  switch (severity) {
    case 'high':
      return AlertOctagon
    case 'medium':
      return AlertTriangle
    default:
      return Info
  }
}

function formatDimension(dimension: string): string {
  const labels: Record<string, string> = {
    metadata: t('health.dimensions.metadata'),
    images: t('health.dimensions.images'),
    valuation: t('health.dimensions.valuation'),
    ai: t('coinDetail.sections.analysis.title'),
  }
  return labels[dimension] || dimension
}

function formatQuickAction(action: HealthQuickAction): string {
  const labels: Record<HealthQuickAction, string> = {
    edit_metadata: t('collection.needsAttention.actions.edit'),
    upload_images: t('collection.needsAttention.actions.upload'),
    run_valuation: t('collection.needsAttention.actions.valuate'),
    run_ai_analysis: t('collection.needsAttention.actions.analyze'),
  }
  return labels[action] || action
}

function handleQuickAction(action: HealthQuickAction) {
  emit('quickAction', action)
}
</script>
