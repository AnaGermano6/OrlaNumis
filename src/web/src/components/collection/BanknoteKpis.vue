<template>
  <section class="mb-4 grid grid-cols-2 gap-3 md:grid-cols-3" :aria-label="t('banknotes.kpis.label')">
    <div
      v-for="kpi in kpis"
      :key="kpi.key"
      class="flex min-w-0 flex-col items-center gap-1 rounded-md border border-border-subtle bg-card px-2 py-4 text-center md:px-4"
      :class="kpi.key === 'value' ? 'col-span-2 md:col-span-1' : ''"
      :data-testid="`banknotes-kpi-${kpi.key}`"
    >
      <span class="whitespace-nowrap font-display text-xl font-semibold text-text-primary md:text-2xl">{{ kpi.display }}</span>
      <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ kpi.label }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
// Headline numbers of the banknotes: how many, of how many countries, and
// their total value in the currency every page uses.
import { computed } from 'vue'
import type { CollectionSummary } from '@/api/client'
import { intlLocale, t } from '@/i18n'
import { formatCurrency } from '@/utils/format'

const props = defineProps<{ summary: CollectionSummary | null }>()

function count(value: number | undefined): string {
  return value === undefined ? '—' : new Intl.NumberFormat(intlLocale()).format(value)
}

const kpis = computed(() => [
  { key: 'banknotes', display: count(props.summary?.coins), label: t('banknotes.kpis.banknotes') },
  { key: 'countries', display: count(props.summary?.countries), label: t('banknotes.kpis.countries') },
  {
    key: 'value',
    display: props.summary ? formatCurrency(props.summary.totalValue) : '—',
    label: t('banknotes.kpis.value'),
  },
])
</script>
