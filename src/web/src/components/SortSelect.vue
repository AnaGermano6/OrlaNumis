<template>
  <select
    class="form-select h-10 w-full bg-card py-0 text-body transition-colors hover:border-border-accent"
    :value="modelValue"
    @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
  >
    <option value="updated_at_desc">{{ t('collection.sort.lastUpdated') }}</option>
    <option value="created_at_desc">{{ t('collection.sort.newest') }}</option>
    <option value="created_at_asc">{{ t('collection.sort.oldest') }}</option>
    <option value="current_value_desc">{{ t('collection.sort.priceHighLow') }}</option>
    <option value="current_value_asc">{{ t('collection.sort.priceLowHigh') }}</option>
    <option value="random_desc">{{ t('collection.sort.random') }}</option>
    <option value="needs_attention">{{ t('collection.sort.needsAttention') }}</option>
  </select>
</template>

<script setup lang="ts">
// The gallery's order, the same size as its filters.
import { t } from '@/i18n'
defineProps<{ modelValue: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
