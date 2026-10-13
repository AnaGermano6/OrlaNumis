<template>
  <article
    v-if="coin"
    class="flex min-w-0 gap-4 rounded-md border border-border-subtle bg-card p-4"
    :aria-labelledby="titleId"
    :data-testid="`daily-${kind}`"
  >
    <RouterLink
      :to="`/coin/${coin.id}`"
      class="flex shrink-0 items-center justify-center overflow-hidden rounded-sm border border-border-subtle bg-input"
      :class="kind === 'banknote' ? 'h-24 w-36' : 'h-28 w-28'"
      :aria-label="coin.name"
    >
      <AuthenticatedImage v-if="photo" :media-path="photo" :alt="coin.name" class="h-full w-full object-contain">
        <template #fallback><component :is="kind === 'banknote' ? Banknote : Coins" :size="32" :stroke-width="1" class="text-text-primary/30" /></template>
      </AuthenticatedImage>
      <component :is="kind === 'banknote' ? Banknote : Coins" v-else :size="32" :stroke-width="1" class="text-text-primary/30" />
    </RouterLink>
    <div class="flex min-w-0 flex-col gap-1">
      <h3 :id="titleId" class="m-0 text-label font-semibold uppercase tracking-[0.08em] text-gold">
        {{ kind === 'banknote' ? t('collectionOverview.daily.banknote') : t('collectionOverview.daily.coin') }}
      </h3>
      <RouterLink :to="`/coin/${coin.id}`" class="font-display text-lg font-medium leading-snug text-text-primary hover:text-gold">{{ coin.name }}</RouterLink>
      <span class="text-body text-text-secondary">{{ details }}</span>
      <p v-if="description" class="m-0 line-clamp-2 text-chip text-text-muted">{{ description }}</p>
    </div>
  </article>
</template>

<script setup lang="ts">
// One coin (or banknote) of the collection for the day: the same all day,
// another tomorrow. Shows nothing when the collection has none of that kind.
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { Banknote, Coins } from 'lucide-vue-next'
import { getDailyCoin } from '@/api/client'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { t } from '@/i18n'
import type { Coin } from '@/types'
import { formatCurrency } from '@/utils/format'
import { formatNumistaYear } from '@/utils/numistaCatalog'

const props = defineProps<{ kind: 'coin' | 'banknote' }>()

const coin = ref<Coin | null>(null)
const titleId = computed(() => `daily-${props.kind}-title`)

// The obverse first: the side a coin is known by.
const photo = computed(() => {
  const images = coin.value?.images ?? []
  return (images.find((image) => image.imageType === 'obverse') ?? images.find((image) => image.isPrimary) ?? images[0])?.filePath
})

const details = computed(() => {
  const item = coin.value
  if (!item) return ''
  const year = item.year != null ? formatNumistaYear(item.year) : item.dateRange
  return [item.country, year, item.grade, item.currentValue != null ? formatCurrency(item.currentValue) : ''].filter(Boolean).join(' · ')
})

const description = computed(() => coin.value?.obverseDescription?.trim() || coin.value?.reverseDescription?.trim() || '')

// The viewer's own day, so the coin changes at their midnight.
function today(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
}

onMounted(async () => {
  try {
    coin.value = (await getDailyCoin(props.kind, today())).data.coin
  } catch {
    // Nothing to show: the card stays out of the page.
  }
})
</script>
