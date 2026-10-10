<template>
  <section class="mb-4 grid grid-cols-2 gap-3 md:grid-cols-5" :aria-label="t('collection.kpis.label')">
    <component
      :is="kpi.action ? 'button' : 'div'"
      v-for="kpi in kpis"
      :key="kpi.key"
      :type="kpi.action ? 'button' : undefined"
      class="flex min-w-0 flex-col items-center gap-1 rounded-md border border-border-subtle bg-card px-2 py-4 text-center md:px-4"
      :class="[
        kpi.action ? 'cursor-pointer transition-colors hover:border-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2' : '',
        kpi.wide ? 'col-span-2 md:col-span-1' : '',
      ]"
      :title="kpi.hint"
      @click="kpi.action?.()"
    >
      <span class="whitespace-nowrap font-display text-xl font-semibold text-text-primary md:text-2xl">{{ kpi.display }}</span>
      <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ kpi.label }}</span>
      <span v-if="kpi.note" class="text-chip text-text-muted">{{ kpi.note }}</span>
    </component>
  </section>
</template>

<script setup lang="ts">
// Headline numbers of the collection shown at the top of the gallery: coins,
// distinct coin types, countries, coincards and total value in euros. They
// cover the whole collection, not the current filters.
import { computed } from 'vue'
import type { CollectionSummary } from '@/api/client'
import { intlLocale, t } from '@/i18n'
import { formatCurrency } from '@/utils/format'

const props = defineProps<{
  summary: CollectionSummary | null
}>()

const emit = defineEmits<{
  /** The coincard KPI was chosen: show only coins in a coincard. */
  'show-coincards': []
}>()

interface Kpi {
  key: string
  display: string
  label: string
  hint: string
  note: string
  action?: () => void
  /** Takes the whole row on phones, for long amounts. */
  wide?: boolean
}

function count(value: number | undefined): string {
  return value === undefined ? '—' : new Intl.NumberFormat(intlLocale()).format(value)
}

function formatDate(isoDate: string): string {
  const date = new Date(`${isoDate}T00:00:00Z`)
  return Number.isNaN(date.getTime()) ? isoDate : date.toLocaleDateString(intlLocale(), { timeZone: 'UTC' })
}

// A euro collection's total is shown as is. A dollar collection's total is
// converted by the server with the ECB reference rate, which is named here.
function valueKpi(summary: CollectionSummary | null): Kpi {
  const rate = summary?.currency === 'EUR' ? undefined : summary?.exchangeRate
  const euros = summary?.totalValueEur
  const notes = [
    summary && euros == null ? t('collection.kpis.rateUnavailable') : '',
    summary?.unvaluedCoins ? t('collection.kpis.unvalued', { count: summary.unvaluedCoins }) : '',
  ].filter(Boolean)
  const rateText = rate ? new Intl.NumberFormat(intlLocale(), { maximumFractionDigits: 4 }).format(rate.usdPerEur) : ''
  return {
    key: 'value',
    display: euros != null ? formatCurrency(euros, 'EUR') : '—',
    label: t('collection.kpis.value'),
    hint: summary?.currency === 'EUR'
      ? t('collection.kpis.valueHintEuros')
      : rate ? t('collection.kpis.valueHint', { date: formatDate(rate.date), rate: rateText }) : t('collection.kpis.valueHintNoRate'),
    note: notes.length ? notes.join(' · ') : rate ? t('collection.kpis.rateDate', { date: formatDate(rate.date) }) : '',
    wide: true,
  }
}

const kpis = computed<Kpi[]>(() => {
  const summary = props.summary
  return [
    { key: 'coins', display: count(summary?.coins), label: t('collection.kpis.coins'), hint: t('collection.kpis.coinsHint'), note: '' },
    { key: 'types', display: count(summary?.types), label: t('collection.kpis.types'), hint: t('collection.kpis.typesHint'), note: '' },
    {
      key: 'countries',
      display: count(summary?.countries),
      label: t('collection.kpis.countries'),
      hint: t('collection.kpis.countriesHint'),
      note: summary?.withoutCountry ? t('collection.kpis.withoutCountry', { count: summary.withoutCountry }) : '',
    },
    {
      key: 'coincards',
      display: count(summary?.coincards),
      label: t('collection.kpis.coincards'),
      hint: t('collection.kpis.coincardsHint'),
      note: '',
      action: () => emit('show-coincards'),
    },
    valueKpi(summary),
  ]
})
</script>
