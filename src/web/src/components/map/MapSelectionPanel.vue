<template>
  <section class="min-w-0 rounded-md border border-border-subtle bg-card p-4" aria-live="polite" data-testid="map-country-panel">
    <p v-if="!title" class="m-0 text-body text-text-secondary">{{ t('map.country.choose') }}</p>
    <template v-else>
      <div class="mb-3 flex flex-wrap items-start justify-between gap-3">
        <div class="min-w-0">
          <p v-if="kicker" class="section-label mb-1">{{ kicker }}</p>
          <h2 class="m-0 font-display text-xl font-medium text-heading">{{ title }}</h2>
          <p v-if="details.length" class="mb-0 mt-1 flex flex-wrap gap-x-3 text-chip text-text-secondary">
            <span v-for="detail in details" :key="detail">{{ detail }}</span>
          </p>
          <p v-if="note" class="mb-0 mt-1 text-chip text-text-muted">{{ note }}</p>
        </div>
        <div class="flex flex-wrap gap-2">
          <button v-if="galleryCountry && coins.length" type="button" class="btn btn-primary btn-sm" @click="emit('show-in-gallery', galleryCountry)">{{ t('map.country.showInGallery') }}</button>
          <button type="button" class="btn btn-secondary btn-sm" @click="emit('clear')">{{ t('map.country.clear') }}</button>
        </div>
      </div>
      <template v-if="coins.length">
        <label class="mb-3 flex items-center gap-2 whitespace-nowrap text-chip text-text-secondary">
          {{ t('map.country.sortBy') }}
          <select v-model="sortBy" class="form-select h-[34px] bg-card text-body" style="width: auto" data-testid="map-sort">
            <option value="year">{{ t('map.country.sortYear') }}</option>
            <option value="name">{{ t('map.country.sortName') }}</option>
            <option value="value">{{ t('map.country.sortValue') }}</option>
          </select>
        </label>
        <CoinListTable :coins="sorted" />
      </template>
      <p v-else-if="emptyText" class="m-0 text-body text-text-secondary">{{ emptyText }}</p>
    </template>
  </section>
</template>

<script setup lang="ts">
// The coins of the place (or mint) chosen on the map, sortable by year,
// name or value, with the way to the gallery filtered to that country.
import { computed, ref } from 'vue'
import type { Coin } from '@/types'
import CoinListTable from '@/components/collection/CoinListTable.vue'
import { intlLocale, t } from '@/i18n'

const props = withDefaults(defineProps<{
  /** Empty: nothing chosen. */
  title: string
  kicker?: string
  details?: string[]
  note?: string
  coins: Coin[]
  /** Shown when there are no coins. */
  emptyText?: string
  /** The stored country name to filter the gallery by, if any. */
  galleryCountry?: string | null
}>(), { kicker: '', details: () => [], note: '', emptyText: '', galleryCountry: null })

const emit = defineEmits<{
  clear: []
  'show-in-gallery': [country: string]
}>()

const sortBy = ref<'year' | 'name' | 'value'>('year')

const sorted = computed(() => {
  const list = [...props.coins]
  const byName = (a: Coin, b: Coin) => a.name.localeCompare(b.name, intlLocale())
  if (sortBy.value === 'name') return list.sort(byName)
  if (sortBy.value === 'value') return list.sort((a, b) => (b.currentValue ?? -1) - (a.currentValue ?? -1) || byName(a, b))
  return list.sort((a, b) => (a.year ?? Infinity) - (b.year ?? Infinity) || byName(a, b))
})
</script>
