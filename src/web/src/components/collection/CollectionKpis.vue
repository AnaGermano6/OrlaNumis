<template>
  <section class="mb-4 grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-6" :aria-label="t('collection.kpis.label')">
    <component
      :is="kpi.action ? 'button' : 'div'"
      v-for="kpi in kpis"
      :key="kpi.key"
      :type="kpi.action ? 'button' : undefined"
      class="flex min-w-0 flex-col items-center gap-1 rounded-md border border-border-subtle bg-card px-2 py-4 text-center md:px-4"
      :class="[
        kpi.action ? 'cursor-pointer transition-colors hover:border-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2' : '',
        kpi.wide ? 'col-span-2 md:col-span-1' : '',
        kpi.active ? 'border-gold' : '',
      ]"
      :title="kpi.hint"
      :aria-pressed="kpi.active"
      :data-testid="`kpi-${kpi.key}`"
      @click="kpi.action?.()"
    >
      <span class="whitespace-nowrap font-display text-xl font-semibold text-text-primary md:text-2xl">{{ kpi.display }}</span>
      <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ kpi.label }}</span>
      <span v-if="kpi.note" class="text-chip text-text-muted">{{ kpi.note }}</span>
      <button
        v-if="kpi.key === 'countries' && summary?.extinctCountries"
        type="button"
        class="rounded-sm text-chip font-medium text-gold underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        :title="t('collection.kpis.extinctCountriesHint')"
        data-testid="kpi-extinct-countries"
        @click="emit('show-extinct')"
      >
        {{ t('collection.kpis.extinctCountries', { count: summary.extinctCountries }) }}
      </button>
    </component>
  </section>
</template>

<script setup lang="ts">
// Headline numbers of the collection's coins: coins, coins still legal
// tender, commemorative coins, countries (and how many no longer exist),
// coincards and total value in the currency every page uses. They cover the whole
// collection, not the current filters.
import { computed } from 'vue'
import type { CollectionSummary } from '@/api/client'
import { intlLocale, t } from '@/i18n'
import { collectionCurrency, currencyName } from '@/utils/collectionCurrency'
import { formatCurrency } from '@/utils/format'

const props = defineProps<{
  summary: CollectionSummary | null
  /** The gallery shows only the coins still legal tender. */
  currentOnly?: boolean
  /** The gallery shows only the commemorative coins. */
  commemorativeOnly?: boolean
}>()

const emit = defineEmits<{
  /** The coincard KPI was chosen: show only coins in a coincard. */
  'show-coincards': []
  /** The extinct countries line was chosen: show only their coins. */
  'show-extinct': []
  /** Show only coins still legal tender, or every coin again. */
  'toggle-current': []
  /** Show only commemorative coins, or every coin again. */
  'toggle-commemorative': []
}>()

interface Kpi {
  key: string
  display: string
  label: string
  hint: string
  note: string
  action?: () => void
  /** For a card that filters the gallery: whether it shows only these coins. */
  active?: boolean
  /** Takes the whole row on phones, for long amounts. */
  wide?: boolean
}


function count(value: number | undefined): string {
  return value === undefined ? '—' : new Intl.NumberFormat(intlLocale()).format(value)
}

// The total is in the currency every page uses (Settings › Appearance).
function valueKpi(summary: CollectionSummary | null): Kpi {
  return {
    key: 'value',
    display: summary ? formatCurrency(summary.totalValue) : '—',
    label: t('collection.kpis.value'),
    hint: t('collection.kpis.valueHintIn', { currency: currencyName(collectionCurrency.value) }),
    note: summary?.unvaluedCoins ? t('collection.kpis.unvalued', { count: summary.unvaluedCoins }) : '',
    wide: true,
  }
}

const kpis = computed<Kpi[]>(() => {
  const summary = props.summary
  return [
    { key: 'coins', display: count(summary?.coins), label: t('collection.kpis.coins'), hint: t('collection.kpis.coinsHint'), note: '' },
    {
      key: 'current',
      display: count(summary ? summary.currentCoins ?? 0 : undefined),
      label: t('collection.kpis.current'),
      hint: t('collection.kpis.currentHint'),
      note: '',
      action: () => emit('toggle-current'),
      active: !!props.currentOnly,
    },
    {
      key: 'commemorative',
      display: count(summary ? summary.commemorativeCoins ?? 0 : undefined),
      label: t('collection.kpis.commemorative'),
      hint: t('collection.kpis.commemorativeHint'),
      note: '',
      action: () => emit('toggle-commemorative'),
      active: !!props.commemorativeOnly,
    },
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
      wide: true,
    },
    valueKpi(summary),
  ]
})
</script>
