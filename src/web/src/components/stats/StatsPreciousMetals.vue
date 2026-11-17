<template>
  <section v-if="metals" class="card" aria-labelledby="stats-metals-title" data-testid="stats-metals">
    <h2 id="stats-metals-title" class="mb-5 text-lg">{{ t('stats.metals.title') }}</h2>
    <div class="grid gap-5 sm:grid-cols-2 sm:divide-x sm:divide-border-subtle">
      <div v-for="metal in list" :key="metal.key" class="flex min-w-0 flex-col gap-1 sm:px-6 sm:first:pl-0" :data-testid="`metal-${metal.key}`">
        <span class="flex items-center gap-2 text-body text-text-secondary">
          <span class="h-3 w-5 shrink-0 rounded-sm border border-border-subtle" :style="{ background: metal.color }" aria-hidden="true" />
          {{ metal.label }}
        </span>
        <span class="font-display text-2xl font-semibold text-text-primary" data-testid="metal-fine">{{ grams(metal.total.fineGrams) }}</span>
        <span class="text-chip text-text-muted">
          <template v-if="metal.total.pieces">{{ t('stats.metals.gross', { weight: grams(metal.total.grossGrams), pieces: t('stats.metals.pieces', { count: metal.total.pieces }) }) }}</template>
          <template v-else>{{ metal.none }}</template>
        </span>
        <span v-if="metal.total.unknownFineness" class="text-chip text-text-muted" data-testid="metal-unknown">
          {{ t('stats.metals.unknownFineness', { count: metal.total.unknownFineness }) }}
        </span>
      </div>
    </div>

    <!-- The coins' other materials, most first, each with how many coins are of it. -->
    <div v-if="others.length" class="mt-6 border-t border-border-subtle pt-5" data-testid="metal-others">
      <h3 class="mb-3 mt-0 font-sans text-body font-semibold text-text-secondary">{{ t('stats.metals.others') }}</h3>
      <ul class="m-0 grid list-none gap-x-8 gap-y-3 p-0 sm:grid-cols-2 lg:grid-cols-3">
        <li v-for="entry in others" :key="entry.material" :title="t('stats.metals.otherHint', { count: entry.count })" :data-testid="`material-${entry.material}`">
          <span class="flex items-baseline justify-between gap-3 text-body">
            <span class="min-w-0 truncate text-text-primary">{{ materialLabel(entry.material) }}</span>
            <span class="shrink-0 font-semibold text-text-secondary">{{ entry.count }}</span>
          </span>
          <span class="mt-1 block h-1.5 overflow-hidden rounded-full bg-input" aria-hidden="true">
            <span class="block h-full rounded-full bg-gold" :style="{ width: `${(entry.count / maxOther) * 100}%` }" />
          </span>
        </li>
      </ul>
    </div>
  </section>
</template>

<script setup lang="ts">
// The collection's gold and silver: the pure metal (the weight times the
// fineness its Numista type gives), then the whole weight of those coins and
// how many they are; coins whose fineness nobody says count in the whole
// weight only. Plated coins are not of the metal they are plated with.
// Under them, the coins' other materials, most first.
import { computed } from 'vue'
import type { MetalTotal } from '@/types'
import { intlLocale, t } from '@/i18n'
import { materialLabel } from '@/utils/format'
import { numberFormat } from '@/utils/numberFormat'

const props = defineProps<{
  metals?: { gold: MetalTotal; silver: MetalTotal }
  /** How many coins of each material the collection has. */
  byMaterial?: { material: string; count: number }[]
}>()

/** Gold and silver have cards of their own; paper and polymer are banknotes'. */
const NOT_OTHER = new Set(['Gold', 'Silver', 'Paper', 'Polymer'])

const others = computed(() => (props.byMaterial ?? [])
  .filter((entry) => entry.count > 0 && entry.material && !NOT_OTHER.has(entry.material))
  .sort((a, b) => b.count - a.count || materialLabel(a.material).localeCompare(materialLabel(b.material), intlLocale())))

const maxOther = computed(() => Math.max(1, ...others.value.map((entry) => entry.count)))

const list = computed(() => props.metals ? [
  { key: 'gold', label: t('stats.metals.gold'), color: 'var(--mat-gold)', total: props.metals.gold, none: t('stats.metals.noGold') },
  { key: 'silver', label: t('stats.metals.silver'), color: 'var(--mat-silver)', total: props.metals.silver, none: t('stats.metals.noSilver') },
] : [])

function grams(value: number): string {
  return t('stats.metals.grams', { value: numberFormat({ maximumFractionDigits: 1, minimumFractionDigits: value > 0 && value < 10 ? 1 : 0 }).format(value) })
}
</script>
