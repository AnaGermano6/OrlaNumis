<template>
  <div class="overflow-x-auto rounded-md border border-border-subtle bg-card">
    <table class="data-table text-body">
      <thead>
        <tr>
          <th v-if="selectable" class="w-10 px-3 py-2"><span class="sr-only">{{ t('collection.list.select') }}</span></th>
          <th class="w-14 px-3 py-2"><span class="sr-only">{{ t('collection.list.photo') }}</span></th>
          <th class="px-3 py-2">{{ t('coinFields.name') }}</th>
          <th class="hidden px-3 py-2 md:table-cell">{{ t('coinFields.country') }}</th>
          <th class="hidden px-3 py-2 md:table-cell">{{ t('collection.list.year') }}</th>
          <th class="hidden px-3 py-2 lg:table-cell">{{ t('coinFields.grade') }}</th>
          <th class="hidden px-3 py-2 lg:table-cell">{{ t('coinFields.rarity') }}</th>
          <th class="hidden px-3 py-2 lg:table-cell">{{ t('coinFields.presentation') }}</th>
          <th class="hidden whitespace-nowrap px-3 py-2 !text-center sm:table-cell">{{ t('coinFields.quantity') }}</th>
          <th class="hidden whitespace-nowrap px-3 py-2 !text-right sm:table-cell">{{ t('coinFields.currentValue') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="coin in coins"
          :key="coin.id"
          class="hover:bg-card-hover [&:last-child>td]:border-b-0"
          :class="{ 'bg-gold-glow': selectable && selectedIds.has(coin.id) }"
        >
          <td v-if="selectable" class="px-3 py-2">
            <input
              type="checkbox"
              class="h-4 w-4 accent-[var(--accent-gold)]"
              :checked="selectedIds.has(coin.id)"
              :aria-label="t('collection.list.selectCoin', { name: coin.name })"
              @change="$emit('toggle-select', coin.id)"
            />
          </td>
          <td class="px-3 py-2">
            <div class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-sm border border-border-subtle bg-input">
              <AuthenticatedImage v-if="thumbnail(coin)" :media-path="thumbnail(coin)" :alt="coin.name" class="h-full w-full object-contain" loading="lazy">
                <template #fallback><Coins :size="20" :stroke-width="1" class="text-text-primary/30" /></template>
              </AuthenticatedImage>
              <Coins v-else :size="20" :stroke-width="1" class="text-text-primary/30" />
            </div>
          </td>
          <td class="min-w-[12rem] px-3 py-2">
            <RouterLink :to="`/coin/${coin.id}`" class="font-medium text-text-primary hover:text-gold">{{ coin.name }}</RouterLink>
            <span class="block text-chip text-text-secondary md:hidden">{{ [coin.country, yearText(coin), quantityOf(coin) > 1 ? t('collection.list.quantityShort', { count: quantityOf(coin) }) : ''].filter(Boolean).join(' · ') }}</span>
          </td>
          <td class="hidden px-3 py-2 text-text-secondary md:table-cell">{{ coin.country || '—' }}</td>
          <td class="hidden whitespace-nowrap px-3 py-2 text-text-secondary md:table-cell">{{ yearText(coin) || '—' }}</td>
          <td class="hidden px-3 py-2 text-text-secondary lg:table-cell">{{ coin.grade || '—' }}</td>
          <td class="hidden px-3 py-2 text-text-secondary lg:table-cell">{{ rarityLabel(coin.rarity) || '—' }}</td>
          <td class="hidden px-3 py-2 text-text-secondary lg:table-cell">{{ presentationLabel(coin.presentation) || '—' }}</td>
          <td class="hidden px-3 py-2 !text-center tabular-nums text-text-primary sm:table-cell" data-testid="list-quantity">{{ quantityOf(coin) }}</td>
          <td class="hidden whitespace-nowrap px-3 py-2 !text-right text-text-primary sm:table-cell">{{ coin.currentValue != null ? formatCurrency(coin.currentValue) : '—' }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
// The collection as a compact table: one row per coin with the details used
// to browse by country, year, rarity, presentation and how many are held.
import { RouterLink } from 'vue-router'
import { Coins } from 'lucide-vue-next'
import type { Coin, ImageType } from '@/types'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { t } from '@/i18n'
import { presentationLabel, rarityLabel } from '@/utils/coinAttributes'
import { formatCurrency } from '@/utils/format'
import { formatNumistaYear } from '@/utils/numistaCatalog'

const props = withDefaults(defineProps<{
  coins: Coin[]
  imageSide?: ImageType | null
  selectable?: boolean
  selectedIds?: Set<number>
}>(), { imageSide: null, selectable: false, selectedIds: () => new Set<number>() })

defineEmits<{
  'toggle-select': [coinId: number]
}>()

function thumbnail(coin: Coin): string | undefined {
  const images = coin.images ?? []
  const bySide = props.imageSide ? images.find((image) => image.imageType === props.imageSide) : undefined
  return (bySide ?? images.find((image) => image.isPrimary) ?? images.find((image) => image.imageType === 'obverse') ?? images[0])?.filePath
}

// Coins saved before the field existed count as one.
function quantityOf(coin: Coin): number {
  return coin.quantity && coin.quantity > 0 ? coin.quantity : 1
}

function yearText(coin: Coin): string {
  return coin.year !== null && coin.year !== undefined ? formatNumistaYear(coin.year) : coin.dateRange ?? ''
}
</script>
