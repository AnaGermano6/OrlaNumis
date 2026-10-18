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
    <SwipeGallery v-if="isPwa && viewMode === 'swipe' && !selectMode && groupBy === 'none'" :coins="coins" :total="total" :page="page" :per-page="perPage" @page-change="$emit('page-change', $event)" />
    <template v-else>
      <p v-if="truncated" class="mb-3 text-body text-text-secondary" role="status">{{ t('collection.groups.truncated', { shown: coins.length, total }) }}</p>
      <section
        v-for="group in groups"
        :key="group.key"
        class="mb-8 last:mb-0"
        :aria-label="groupBy === 'none' ? undefined : group.label"
      >
        <h2 v-if="groupBy !== 'none'" class="mb-3 flex items-baseline gap-2 border-b border-border-subtle pb-1 font-display text-lg font-medium text-heading">
          <span v-if="groupBy === 'country' && hasCountryFlag(group.label)" class="font-sans" aria-hidden="true" data-testid="group-flag"><CountryFlag :country="group.label" /></span>
          {{ group.label }}
          <span class="font-sans text-chip font-normal text-text-muted">{{ t('collection.groups.count', { count: group.coins.length }) }}</span>
        </h2>
        <CoinListTable
          v-if="layout === 'list'"
          :coins="group.coins"
          :image-side="gridSide"
          :selectable="selectMode"
          :selected-ids="selectedCoinIds"
          @toggle-select="$emit('toggle-coin-select', $event)"
        />
        <div v-else class="coins-grid">
          <CoinCard
            v-for="coin in group.coins"
            :key="coin.id"
            :coin="coin"
            :image-side="gridSide"
            :selectable="selectMode"
            :selected="selectedCoinIds.has(coin.id)"
            @toggle-select="$emit('toggle-coin-select', $event)"
          />
        </div>
      </section>
    </template>
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
import { computed } from 'vue'
import type { Coin, ImageType } from '@/types'
import CoinCard from '@/components/CoinCard.vue'
import SwipeGallery from '@/components/SwipeGallery.vue'
import CoinListTable from '@/components/collection/CoinListTable.vue'
import { groupCoins, type CoinGroupBy } from '@/utils/coinAttributes'
import { t } from '@/i18n'
import CountryFlag from '@/components/CountryFlag.vue'
import { hasCountryFlag } from '@/utils/countryFlags'

const props = withDefaults(defineProps<{
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
  groupBy?: CoinGroupBy
  layout?: 'grid' | 'list'
  /** The grouped view hit the load limit and shows part of the collection. */
  truncated?: boolean
}>(), { groupBy: 'none', layout: 'grid', truncated: false })

const groups = computed(() => groupCoins(props.coins, props.groupBy))

defineEmits<{
  'select-all': []
  'deselect-all': []
  'toggle-coin-select': [coinId: number]
  'page-change': [page: number]
}>()
</script>
