<template>
  <section class="grid min-w-0 gap-2" :aria-label="t('deepAnalysis.coverage.title')">
    <h3 class="m-0 text-lg font-semibold text-text-primary">{{ title || t('deepAnalysis.coverage.title') }}</h3>
    <ul class="m-0 grid gap-2 p-0" style="list-style: none;">
      <li
        v-for="entry in coverage"
        :key="entry.provider"
        class="grid min-w-0 grid-cols-1 items-start gap-1 rounded-sm border border-border-subtle bg-card p-2 sm:grid-cols-[auto_auto_minmax(0,1fr)] sm:items-center sm:gap-3"
      >
        <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-primary">
          {{ providerLabel(entry.provider) }}
        </span>
        <span class="text-sm font-medium" :class="statusClasses(entry.status)">
          {{ statusLabel(entry.status) }}
        </span>
        <span class="min-w-0 break-words text-sm text-text-secondary [overflow-wrap:anywhere]">
          {{ entry.note }}
          <SafeExternalLink
            v-if="entry.linkOut"
            :href="entry.linkOut"
            class="text-gold underline"
          >
            {{ t('deepAnalysis.coverage.viewReference') }}
          </SafeExternalLink>
        </span>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import SafeExternalLink from '@/components/SafeExternalLink.vue'
import type { DeepProviderId, DeepProviderStatus, DeepReportCoverage } from '@/types'
import type { MessageKey } from '@/i18n'
import { t } from '@/i18n'

withDefaults(defineProps<{
  coverage: DeepReportCoverage[]
  title?: string
}>(), {
  title: undefined,
})

const statusLabelKeys: Record<DeepProviderStatus, MessageKey> = {
  pending: 'deepAnalysis.coverageStatus.pending',
  running: 'deepAnalysis.coverageStatus.running',
  contributed: 'deepAnalysis.coverageStatus.contributed',
  no_match: 'deepAnalysis.coverageStatus.noMatch',
  failed: 'deepAnalysis.coverageStatus.failed',
  timed_out: 'deepAnalysis.coverageStatus.timedOut',
  skipped: 'deepAnalysis.coverageStatus.skipped',
  not_automated: 'deepAnalysis.coverageStatus.notAutomated',
  unavailable: 'deepAnalysis.coverageStatus.unavailable',
}

function providerLabel(provider: DeepProviderId): string {
  return provider.toUpperCase()
}

function statusLabel(status: DeepProviderStatus): string {
  const key = statusLabelKeys[status]
  return key ? t(key) : status
}

function statusClasses(status: DeepProviderStatus): string {
  if (status === 'contributed') return 'text-gold'
  if (status === 'failed' || status === 'timed_out') return 'text-byzantine'
  return 'text-text-secondary'
}
</script>
