<template>
  <section v-if="summary && summary.coins > 0" class="mb-4 rounded-md border border-border-subtle bg-card p-4" aria-labelledby="collection-highlights-title">
    <h2 id="collection-highlights-title" class="mb-3 mt-0 font-display text-lg font-medium text-heading">{{ t('collection.highlights.title') }}</h2>
    <ul class="m-0 grid list-none gap-3 p-0 md:grid-cols-3">
      <li v-for="item in items" :key="item.key" class="flex min-w-0 items-center gap-3 rounded-sm border border-border-subtle bg-input p-3">
        <div class="flex h-16 w-16 shrink-0 items-center justify-center overflow-hidden rounded-sm border border-border-subtle bg-card">
          <AuthenticatedImage v-if="item.coin?.imagePath" :media-path="item.coin.imagePath" :alt="item.coin.name" class="h-full w-full object-contain">
            <template #fallback><Coins :size="24" :stroke-width="1" class="text-text-primary/30" /></template>
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
// The rarest, oldest and most valuable coins of the collection, shown under
// the gallery KPIs. Values are stored in US dollars and shown in euros with
// the same ECB rate as the total-value KPI.
import { computed, type Component } from 'vue'
import { RouterLink } from 'vue-router'
import { Coins, Gem, Hourglass, TrendingUp } from 'lucide-vue-next'
import type { CoinHighlight, CollectionSummary } from '@/api/client'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { t } from '@/i18n'
import { rarityLabel } from '@/utils/coinAttributes'
import { formatCurrency } from '@/utils/format'
import { formatNumistaYear } from '@/utils/numistaCatalog'

const props = defineProps<{
  summary: CollectionSummary | null
}>()

interface HighlightItem {
  key: string
  label: string
  icon: Component
  coin: CoinHighlight | null
  detail: string
  secondary: string
  empty: string
}

function valueText(value: number | null): string {
  if (value == null) return ''
  const rate = props.summary?.exchangeRate
  return rate ? formatCurrency(value / rate.usdPerEur, 'EUR') : formatCurrency(value, 'USD')
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
      empty: t('collection.highlights.noRarity'),
    },
    {
      key: 'oldest',
      label: t('collection.highlights.oldest'),
      icon: Hourglass,
      coin: oldest,
      detail: oldest ? yearText(oldest) : '',
      secondary: oldest ? secondary(oldest, oldest.grade) : '',
      empty: t('collection.highlights.noYear'),
    },
    {
      key: 'mostValuable',
      label: t('collection.highlights.mostValuable'),
      icon: TrendingUp,
      coin: mostValuable,
      detail: mostValuable ? valueText(mostValuable.currentValue) : '',
      secondary: mostValuable ? secondary(mostValuable, yearText(mostValuable)) : '',
      empty: t('collection.highlights.noValue'),
    },
  ]
})
</script>
