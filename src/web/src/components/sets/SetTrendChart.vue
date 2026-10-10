<template>
  <section class="mb-6 rounded-md border border-border-subtle bg-card p-6">
    <div class="mb-4 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <h2>{{ t('coinDetail.sections.valuation.title') }}</h2>
      <select :value="range" class="rounded-sm border border-border-subtle bg-input px-[0.6rem] py-[0.4rem] text-base text-text-primary" @change="$emit('update:range', ($event.target as HTMLSelectElement).value)">
        <option value="1m">{{ t('sets.trend.range1m') }}</option>
        <option value="3m">{{ t('sets.trend.range3m') }}</option>
        <option value="1y">{{ t('sets.trend.range1y') }}</option>
        <option value="all">{{ t('stats.value.timeframeAll') }}</option>
      </select>
    </div>
    <div v-if="snapshots.length" class="flex flex-col gap-2">
      <div
        v-for="snapshot in snapshots"
        :key="snapshot.snapshotDate"
        class="flex justify-between border-b border-border-subtle pb-[0.4rem] text-body text-text-secondary last:border-b-0"
      >
        <span>{{ formatDate(snapshot.snapshotDate) }}</span>
        <strong class="font-semibold text-gold">{{ formatCurrency(snapshot.totalValue) }}</strong>
      </div>
    </div>
    <p v-else class="m-0 text-body text-text-secondary">{{ t('sets.trend.empty') }}</p>
    <div v-if="$slots.actions" class="mt-4">
      <slot name="actions" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { formatCurrency } from '@/utils/format'
import type { CoinSetSnapshot } from '@/types'
import { t, intlLocale } from '@/i18n'

defineProps<{
  snapshots: CoinSetSnapshot[]
  range: string
}>()

defineEmits<{
  'update:range': [value: string]
}>()

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(intlLocale())
}
</script>
