<template>
  <section class="mb-4 grid grid-cols-2 gap-3 md:grid-cols-3" :aria-label="t('banknotes.kpis.label')">
    <!-- The counts, then the amounts. -->
    <div
      v-for="kpi in kpis"
      :key="kpi.key"
      class="flex min-w-0 flex-col items-center gap-1 rounded-md border border-border-subtle bg-card px-2 py-4 text-center md:px-4"
      :title="kpi.hint || undefined"
      :data-testid="`banknotes-kpi-${kpi.key}`"
    >
      <!-- A long amount gets a smaller size on phones, to fit half the row. -->
      <span class="whitespace-nowrap font-display font-semibold text-text-primary md:text-2xl" :class="kpi.display.length > 10 ? 'text-base' : 'text-xl'">{{ kpi.display }}</span>
      <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ kpi.label }}</span>
      <span v-if="kpi.note" class="text-chip text-text-muted">{{ kpi.note }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
// Headline numbers of the banknotes: how many (counting quantities, and each
// different note once), of how many countries, their total value and face
// value (each note's times its quantity) in the currency every page uses, and
// how much more than their face value they are worth.
import { computed } from 'vue'
import type { CollectionSummary } from '@/api/client'
import { intlLocale, t } from '@/i18n'
import { faceValueKpis } from '@/utils/faceValueKpis'
import { formatCurrency } from '@/utils/format'

const props = defineProps<{ summary: CollectionSummary | null }>()

function count(value: number | undefined): string {
  return value === undefined ? '—' : new Intl.NumberFormat(intlLocale()).format(value)
}

const kpis = computed(() => [
  { key: 'banknotes', display: count(props.summary?.coins), label: t('banknotes.kpis.banknotes'), hint: '', note: '' },
  { key: 'unique', display: count(props.summary ? props.summary.uniqueCoins ?? props.summary.coins : undefined), label: t('banknotes.kpis.unique'), hint: '', note: '' },
  { key: 'countries', display: count(props.summary?.countries), label: t('banknotes.kpis.countries'), hint: '', note: '' },
  {
    key: 'value',
    display: props.summary ? formatCurrency(props.summary.totalValue) : '—',
    label: t('banknotes.kpis.value'),
    hint: '',
    note: '',
  },
  ...faceValueKpis(props.summary),
])
</script>
