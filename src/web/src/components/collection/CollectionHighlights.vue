<template>
  <section v-if="summary && summary.coins > 0" class="mb-4 rounded-md border border-border-subtle bg-card p-4" :aria-labelledby="titleId">
    <h2 :id="titleId" class="mb-3 mt-0 font-display text-lg font-medium text-heading">{{ title ?? t('collection.highlights.title') }}</h2>
    <ul class="m-0 grid list-none gap-3 p-0 md:grid-cols-3">
      <li v-for="item in items" :key="item.key" class="flex min-w-0 items-center gap-3 rounded-sm border border-border-subtle bg-input p-3">
        <div class="flex h-16 w-16 shrink-0 items-center justify-center overflow-hidden rounded-sm border border-border-subtle bg-card">
          <AuthenticatedImage v-if="item.coin?.imagePath" :media-path="item.coin.imagePath" :alt="item.coin.name" class="h-full w-full object-contain">
            <template #fallback><component :is="kind === 'banknote' ? Banknote : Coins" :size="24" :stroke-width="1" class="text-text-primary/30" /></template>
          </AuthenticatedImage>
          <component :is="item.icon" v-else :size="24" :stroke-width="1.5" class="text-gold" />
        </div>
        <div class="min-w-0">
          <span class="block text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ item.label }}</span>
          <template v-if="item.coin">
            <RouterLink :to="`/coin/${item.coin.id}`" class="block truncate font-medium text-text-primary hover:text-gold" :title="item.coin.name">{{ item.coin.name }}</RouterLink>
            <span class="block truncate text-body text-gold">{{ item.detail }}</span>
            <span v-if="item.secondary" class="block truncate text-chip text-text-muted">{{ item.secondary }}</span>
          </template>
          <span v-else class="block text-chip text-text-muted">{{ item.empty }}</span>
        </div>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
// The rarest, oldest and most valuable coins (or banknotes) of the
// collection, with values in the currency every page uses.
import { computed, type Component } from 'vue'
import { RouterLink } from 'vue-router'
import { Banknote, Coins, Gem, Hourglass, TrendingUp } from 'lucide-vue-next'
import type { CoinHighlight, CollectionSummary } from '@/api/client'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { t } from '@/i18n'
import { rarityLabel } from '@/utils/coinAttributes'
import { formatCurrency } from '@/utils/format'
import { formatNumistaYear } from '@/utils/numistaCatalog'

const props = withDefaults(defineProps<{
  summary: CollectionSummary | null
  /** The banknotes' highlights say "banknote" where the coins' say "coin". */
  kind?: 'coin' | 'banknote'
  title?: string
}>(), { kind: 'coin', title: undefined })

const titleId = computed(() => `${props.kind}-highlights-title`)

interface HighlightItem {
  key: string
  label: string
  icon: Component
  coin: CoinHighlight | null
  detail: string
  secondary: string
  empty: string
}

// In the currency every page uses, like the value KPI.
function valueText(value: number | null): string {
  return value == null ? '' : formatCurrency(value)
}

function yearText(coin: CoinHighlight): string {
  return coin.year != null ? formatNumistaYear(coin.year) : coin.dateRange
}

function secondary(coin: CoinHighlight, ...parts: string[]): string {
  return [coin.country, ...parts].filter(Boolean).join(' · ')
}

const items = computed<HighlightItem[]>(() => {
  const highlights = props.summary?.highlights
  const rarest = highlights?.rarest ?? null
  const oldest = highlights?.oldest ?? null
  const mostValuable = highlights?.mostValuable ?? null
  return [
    {
      key: 'rarest',
      label: t('collection.highlights.rarest'),
      icon: Gem,
      coin: rarest,
      detail: rarest ? rarityLabel(rarest.rarity) : '',
      secondary: rarest ? secondary(rarest, yearText(rarest)) : '',
      empty: props.kind === 'banknote' ? t('collectionOverview.banknoteHighlights.noRarity') : t('collection.highlights.noRarity'),
    },
    {
      key: 'oldest',
      label: t('collection.highlights.oldest'),
      icon: Hourglass,
      coin: oldest,
      detail: oldest ? yearText(oldest) : '',
      secondary: oldest ? secondary(oldest, oldest.grade) : '',
      empty: props.kind === 'banknote' ? t('collectionOverview.banknoteHighlights.noYear') : t('collection.highlights.noYear'),
    },
    {
      key: 'mostValuable',
      label: t('collection.highlights.mostValuable'),
      icon: TrendingUp,
      coin: mostValuable,
      detail: mostValuable ? valueText(mostValuable.currentValue) : '',
      secondary: mostValuable ? secondary(mostValuable, yearText(mostValuable)) : '',
      empty: props.kind === 'banknote' ? t('collectionOverview.banknoteHighlights.noValue') : t('collection.highlights.noValue'),
    },
  ]
})
</script>
