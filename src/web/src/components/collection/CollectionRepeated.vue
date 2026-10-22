<template>
  <section class="mb-4 rounded-md border border-border-subtle bg-card p-4" :aria-labelledby="titleId" :data-testid="`repeated-${kind}`">
    <div class="mb-3 flex flex-wrap items-baseline justify-between gap-2">
      <h3 :id="titleId" class="m-0 font-display text-lg font-medium text-heading">{{ t('collectionOverview.repeated.title') }}</h3>
      <span class="flex flex-wrap items-center gap-3">
        <span v-if="repeatedCount > 0" class="text-body text-text-secondary" data-testid="repeated-counts">
          {{ countLine }}
        </span>
        <!-- The gallery (the coins' or the banknotes'), showing only these. -->
        <button v-if="coins.length" type="button" class="btn btn-secondary btn-sm" data-testid="repeated-open" @click="emit('open')">
          <LayoutGrid :size="14" aria-hidden="true" /> {{ t('collectionOverview.repeated.open') }}
        </button>
        <!-- A file of the duplicates (and what is wanted) for other collectors. -->
        <button v-if="coins.length" type="button" class="btn btn-secondary btn-sm" data-testid="repeated-share" @click="emit('share')">
          <Share2 :size="14" aria-hidden="true" /> {{ t('collectionOverview.repeated.share') }}
        </button>
      </span>
    </div>

    <p v-if="failed" class="m-0 text-body text-text-secondary" role="alert">{{ t('collectionOverview.repeated.failed') }}</p>
    <p v-else-if="loaded && coins.length === 0" class="m-0 text-body text-text-secondary" data-testid="repeated-empty">
      {{ kind === 'banknote' ? t('collectionOverview.repeated.emptyBanknotes') : t('collectionOverview.repeated.emptyCoins') }}
    </p>
    <template v-else-if="coins.length">
      <ul class="m-0 grid list-none gap-2 p-0 md:grid-cols-2">
        <li v-for="coin in shown" :key="coin.id" class="flex min-w-0 items-center gap-3 rounded-sm border border-border-subtle bg-input p-2" data-testid="repeated-item">
          <div class="flex h-12 w-12 shrink-0 items-center justify-center overflow-hidden rounded-sm border border-border-subtle bg-card">
            <AuthenticatedImage v-if="thumbnail(coin)" :media-path="thumbnail(coin)" :alt="coin.name" class="h-full w-full object-contain" loading="lazy">
              <template #fallback><component :is="kind === 'banknote' ? Banknote : Coins" :size="20" :stroke-width="1" class="text-text-primary/30" /></template>
            </AuthenticatedImage>
            <component :is="kind === 'banknote' ? Banknote : Coins" v-else :size="20" :stroke-width="1" class="text-text-primary/30" />
          </div>
          <div class="min-w-0 flex-1">
            <RouterLink :to="`/coin/${coin.id}`" class="block truncate font-medium text-text-primary hover:text-gold" :title="coin.name">{{ coin.name }}</RouterLink>
            <span class="block truncate text-chip text-text-muted">{{ [coin.country, yearText(coin)].filter(Boolean).join(' · ') }}</span>
          </div>
          <div class="shrink-0 text-right">
            <span class="block font-display text-lg font-semibold text-text-primary tabular-nums">{{ t('collection.list.quantityShort', { count: coin.quantity }) }}</span>
            <span class="block text-chip text-gold">{{ t('collectionOverview.repeated.spare', { count: coin.quantity - 1 }) }}</span>
          </div>
        </li>
      </ul>
      <button
        v-if="coins.length > PREVIEW"
        type="button"
        class="mt-3 rounded-sm text-body text-gold underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        data-testid="repeated-toggle"
        @click="expanded = !expanded"
      >
        {{ expanded ? t('collectionOverview.repeated.showLess') : t('collectionOverview.repeated.showAll', { count: coins.length }) }}
      </button>
    </template>
  </section>
</template>

<script setup lang="ts">
// The coins (or banknotes) held more than once, most copies first, with how
// many copies are left to trade.
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { Banknote, Coins, LayoutGrid, Share2 } from 'lucide-vue-next'
import { getCoins, type CollectionSummary } from '@/api/client'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { t } from '@/i18n'
import type { Coin } from '@/types'
import { formatNumistaYear } from '@/utils/numistaCatalog'

const props = defineProps<{
  kind: 'coin' | 'banknote'
  summary: CollectionSummary | null
}>()

const emit = defineEmits<{ open: []; share: [] }>()

/** Shown before "Show all". */
const PREVIEW = 6

const coins = ref<Coin[]>([])
const loaded = ref(false)
const failed = ref(false)
const expanded = ref(false)

const titleId = computed(() => `${props.kind}-repeated-title`)
const shown = computed(() => (expanded.value ? coins.value : coins.value.slice(0, PREVIEW)))
const repeatedCount = computed(() => props.summary?.repeatedCoins ?? coins.value.length)
const countLine = computed(() => {
  const spares = props.summary?.spareCoins ?? coins.value.reduce((sum, coin) => sum + coin.quantity - 1, 0)
  const repeated = props.kind === 'banknote'
    ? t('collectionOverview.repeated.banknotes', { count: repeatedCount.value })
    : t('collectionOverview.repeated.coins', { count: repeatedCount.value })
  return `${repeated} · ${t('collectionOverview.repeated.spares', { count: spares })}`
})

function thumbnail(coin: Coin): string | undefined {
  const images = coin.images ?? []
  return (images.find((image) => image.isPrimary) ?? images.find((image) => image.imageType === 'obverse') ?? images[0])?.filePath
}

function yearText(coin: Coin): string {
  return coin.year != null ? formatNumistaYear(coin.year) : coin.dateRange ?? ''
}

onMounted(async () => {
  try {
    const { data } = await getCoins({ kind: props.kind, repeated: 'true', wishlist: 'false', sort: 'quantity', order: 'desc', limit: 100 })
    coins.value = data.coins
  } catch {
    failed.value = true
  } finally {
    loaded.value = true
  }
})
</script>
