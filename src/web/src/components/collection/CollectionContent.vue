<template>
  <div v-if="loading" class="loading-overlay">
    <div class="spinner"></div>
    <p>{{ t('collection.loading') }}</p>
  </div>

  <template v-else-if="coins.length">
    <div v-if="selectMode" class="mb-3 flex items-center gap-2">
      <button class="btn btn-sm btn-secondary" @click="$emit('select-all')">{{ t('collection.selectAll') }}</button>
      <button class="btn btn-sm btn-secondary" @click="$emit('deselect-all')">{{ t('collection.deselectAll') }}</button>
      <span class="ml-2 text-body text-text-secondary">{{ t('collection.selectedCount', { count: selectedCount }) }}</span>
    </div>
    <SwipeGallery v-if="isPwa && viewMode === 'swipe' && !selectMode" :coins="coins" :total="total" :page="page" :per-page="perPage" @page-change="$emit('page-change', $event)" />
    <div v-else class="coins-grid">
      <CoinCard
        v-for="coin in coins"
        :key="coin.id"
        :coin="coin"
        :image-side="gridSide"
        :selectable="selectMode"
        :selected="selectedCoinIds.has(coin.id)"
        @toggle-select="$emit('toggle-coin-select', $event)"
      />
    </div>
  </template>

  <div v-else class="empty-state">
    <h3>{{ hasFilters ? t('collection.empty.noMatchTitle') : t('collection.empty.emptyTitle') }}</h3>
    <p>{{ hasFilters ? t('collection.empty.noMatchBody') : t('collection.empty.emptyBody') }}</p>
    <router-link v-if="!hasFilters" to="/add" class="btn btn-primary mt-4">
      {{ t('collection.empty.addFirst') }}
    </router-link>
  </div>
</template>

<script setup lang="ts">
import type { Coin, ImageType } from '@/types'
import CoinCard from '@/components/CoinCard.vue'
import SwipeGallery from '@/components/SwipeGallery.vue'
import { t } from '@/i18n'

defineProps<{
  loading: boolean
  coins: Coin[]
  selectMode: boolean
  selectedCoinIds: Set<number>
  selectedCount: number
  isPwa: boolean
  viewMode: 'grid' | 'swipe'
  gridSide: ImageType | null
  hasFilters: boolean
  total: number
  page: number
  perPage: number
}>()

defineEmits<{
  'select-all': []
  'deselect-all': []
  'toggle-coin-select': [coinId: number]
  'page-change': [page: number]
}>()
</script>
