<template>
  <section v-if="total > 0" class="rounded-md border border-border-subtle bg-card p-4 md:px-5" :aria-labelledby="titleId" :data-testid="`${kind}-distribution`">
    <h2 :id="titleId" class="mb-3 mt-0 text-label font-semibold uppercase tracking-[0.08em] text-text-muted">
      {{ kind === 'banknote' ? t('collectionOverview.distribution.banknotesTitle', { count: formatCount(total) }) : t('collectionOverview.distribution.title', { count: formatCount(total) }) }}
    </h2>

    <StackedBar :parts="barParts" :total="total" test-id-prefix="distribution" @choose="choose" />

    <!-- The gallery's other quick filters: legal tender, commemorative, in a coincard. -->
    <p v-if="shortcuts.length" class="m-0 mt-4 flex flex-wrap items-center gap-x-2 gap-y-1 border-t border-border-subtle pt-3 text-chip text-text-muted">
      <span>{{ t('collectionOverview.distribution.showOnly') }}</span>
      <template v-for="(item, index) in shortcuts" :key="item.key">
        <span v-if="index" aria-hidden="true">·</span>
        <button
          type="button"
          class="cursor-pointer rounded-sm font-medium text-gold underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :title="item.hint"
          :data-testid="`distribution-${item.key}`"
          @click="emit('open', item.filter)"
        >
          {{ item.label }}
        </button>
      </template>
    </p>
  </section>
</template>

<script setup lang="ts">
// The coins by Numista's kind of coin, or the banknotes by circulation, as a
// bar: for the coins the three usual kinds (circulating, circulating
// commemorative, collector), for the banknotes those still legal tender and
// those no longer; the rest together. Each different coin or banknote counts
// once, as the galleries list them; a part opens its gallery with that
// filter. Under the coins', those still legal tender, the commemorative ones
// and those in a coincard (each with its quantity), which open the gallery
// likewise.
import { computed, useId } from 'vue'
import type { CollectionSummary } from '@/api/client'
import { t } from '@/i18n'
import StackedBar, { type StackedBarPart } from '@/components/collection/StackedBar.vue'
import { collectionClassificationLabel } from '@/utils/coinAttributes'
import { numberFormat } from '@/utils/numberFormat'

/** The filters a part of the card opens the coin gallery (or the banknotes) with. */
export interface DistributionFilter {
  classification?: string
  circulation?: 'current' | 'demonetized'
  commemorative?: 'true'
  presentation?: 'coincard'
}

const props = withDefaults(defineProps<{
  summary: CollectionSummary | null
  kind?: 'coin' | 'banknote'
}>(), { kind: 'coin' })

const emit = defineEmits<{
  /** A part, or a quick filter, was chosen: show only its coins or banknotes. */
  open: [filter: DistributionFilter]
}>()

const titleId = `distribution-${useId()}`

/** The parts shown on their own, the rest are "others". */
const SHOWN = 3
const COLORS = ['var(--accent-gold)', 'var(--cat-modern)', 'var(--accent-bronze)']
const OTHER_COLOR = 'var(--cat-other)'

interface Part {
  key: string
  label: string
  count: number
  color: string
  hint: string
  /** What it opens; none for the others. */
  filter?: DistributionFilter
}

interface Group {
  key: string
  label: string
  count: number
  filter: DistributionFilter
}

const total = computed(() => props.summary ? props.summary.uniqueCoins ?? props.summary.coins : 0)

// The coins' kinds as the server sorts them (the usual first), or the banknotes still legal tender and no longer.
const groups = computed<Group[]>(() => {
  if (props.kind === 'banknote') {
    const circulation = props.summary?.circulation
    if (!circulation) return []
    return [
      { key: 'current', label: t('collectionOverview.distribution.legalTender'), count: circulation.current, filter: { circulation: 'current' as const } },
      { key: 'demonetized', label: t('collectionOverview.distribution.demonetized'), count: circulation.demonetized, filter: { circulation: 'demonetized' as const } },
    ].filter((group) => group.count > 0)
  }
  return (props.summary?.classifications ?? []).map((kind) => ({ key: kind.key, label: collectionClassificationLabel(kind), count: kind.count, filter: { classification: kind.key } }))
})

const parts = computed<Part[]>(() => {
  const banknotes = props.kind === 'banknote'
  const hint = banknotes ? t('collectionOverview.distribution.openBanknotesHint') : t('collectionOverview.distribution.openHint')
  const shown: Part[] = groups.value.slice(0, SHOWN).map((group, index) => ({ ...group, color: COLORS[index]!, hint }))
  const rest = total.value - shown.reduce((sum, part) => sum + part.count, 0)
  if (rest <= 0) return shown
  // The other parts by name, then those Numista says nothing of.
  const others = groups.value.slice(SHOWN).map((group) => `${group.label}: ${formatCount(group.count)}`)
  const collector = banknotes ? props.summary?.circulation?.nonCirculating ?? 0 : 0
  if (collector > 0) others.push(t('collectionOverview.distribution.nonCirculating', { count: collector }))
  const unknown = rest - collector - groups.value.slice(SHOWN).reduce((sum, group) => sum + group.count, 0)
  if (unknown > 0) others.push(t('collectionOverview.distribution.unknown', { count: unknown }))
  return [...shown, { key: 'other', label: t('collectionOverview.distribution.other'), count: rest, color: OTHER_COLOR, hint: others.join('\n') }]
})

const barParts = computed<StackedBarPart[]>(() => parts.value.map((part) => ({ ...part, clickable: Boolean(part.filter) })))

function choose(key: string) {
  const filter = parts.value.find((part) => part.key === key)?.filter
  if (filter) emit('open', filter)
}

const shortcuts = computed(() => {
  const summary = props.summary
  if (!summary || props.kind === 'banknote') return []
  const items: { key: string; label: string; hint: string; count: number; filter: DistributionFilter }[] = [
    { key: 'current', count: summary.currentCoins ?? 0, label: t('collectionOverview.distribution.current', { count: summary.currentCoins ?? 0 }), hint: t('collectionOverview.distribution.currentHint'), filter: { circulation: 'current' } },
    { key: 'commemorative', count: summary.commemorativeCoins ?? 0, label: t('collectionOverview.distribution.commemorative', { count: summary.commemorativeCoins ?? 0 }), hint: t('collectionOverview.distribution.commemorativeHint'), filter: { commemorative: 'true' } },
    { key: 'coincards', count: summary.coincards ?? 0, label: t('collectionOverview.distribution.coincards', { count: summary.coincards ?? 0 }), hint: t('collectionOverview.distribution.coincardsHint'), filter: { presentation: 'coincard' } },
  ]
  return items.filter((item) => item.count > 0)
})

function formatCount(value: number): string {
  return numberFormat().format(value)
}
</script>
