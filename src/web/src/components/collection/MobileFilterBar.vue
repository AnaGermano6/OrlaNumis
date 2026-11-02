<template>
  <div class="container-bleed sticky top-[60px] z-(--z-sticky) mb-3 flex items-center gap-2 bg-surface py-2">
    <SearchBar class="!max-w-none flex-1" :model-value="search" :placeholder="placeholder" @update:model-value="$emit('update:search', $event)" />
    <button
      type="button"
      class="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-sm border border-border-subtle bg-card text-text-secondary transition-all hover:border-gold hover:bg-[var(--accent-gold-dim)] hover:text-gold"
      :class="menuOpen || activeFilters ? 'border-gold text-gold' : ''"
      :aria-label="activeFilters ? `${t('collection.menu.openFilters')} · ${t('collection.menu.activeFilters', { count: activeFilters })}` : t('collection.menu.openFilters')"
      :aria-expanded="menuOpen"
      data-testid="gallery-filters-button"
      @click="$emit('update:menuOpen', true)"
    >
      <SlidersHorizontal :size="22" />
      <span
        v-if="activeFilters"
        class="absolute -top-1.5 -right-1.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-gold px-1 text-sm font-semibold text-surface"
        aria-hidden="true"
        data-testid="gallery-filters-count"
      >{{ activeFilters }}</span>
    </button>
    <slot name="actions" />
  </div>

  <BaseModal :open="menuOpen" :aria-labelledby="titleId" :padded="false" testid="gallery-filters" @close="$emit('update:menuOpen', false)">
    <div class="flex items-center justify-between gap-3 border-b border-border-subtle px-5 py-3">
      <h2 :id="titleId" class="m-0 text-md text-heading">{{ t('collection.menu.filters') }}</h2>
      <button
        type="button"
        class="inline-flex h-9 w-9 items-center justify-center rounded-sm text-text-muted transition-colors hover:bg-white/5 hover:text-text-primary"
        :aria-label="t('common.close')"
        @click="$emit('update:menuOpen', false)"
      >
        <X :size="20" />
      </button>
    </div>

    <div class="flex flex-col gap-4 px-5 py-4">
      <slot />
    </div>

    <!-- Back to the list, saying how many the filters keep. -->
    <div class="sticky bottom-0 mt-auto border-t border-border-subtle bg-card px-5 py-3">
      <button type="button" class="btn btn-primary w-full justify-center" data-testid="gallery-filters-done" @click="$emit('update:menuOpen', false)">
        {{ doneLabel }}
      </button>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
// A gallery's top on a phone (the coins' and the banknotes'): the search,
// which stays in sight, and a button that opens the filters and view options
// (the default slot) in a sheet, showing how many filters are on.
import { useId } from 'vue'
import SearchBar from '@/components/SearchBar.vue'
import BaseModal from '@/components/ui/BaseModal.vue'
import { SlidersHorizontal, X } from 'lucide-vue-next'
import { t } from '@/i18n'

withDefaults(defineProps<{
  search: string
  menuOpen: boolean
  /** The sheet's button back to the list, e.g. "Show 26 coins". */
  doneLabel: string
  placeholder?: string
  /** How many filters are on, shown on the button. */
  activeFilters?: number
}>(), { placeholder: undefined, activeFilters: 0 })

defineEmits<{
  'update:search': [value: string]
  'update:menuOpen': [value: boolean]
}>()

const titleId = useId()
</script>
