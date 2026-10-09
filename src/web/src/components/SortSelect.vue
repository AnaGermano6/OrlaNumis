<template>
  <div class="relative flex items-center">
    <ArrowUpDown class="pointer-events-none absolute left-2.5 text-text-muted" :size="14" />
    <select
      class="form-select rounded-full py-2 pl-8 text-chip"
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
  </div>
</template>

<script setup lang="ts">
import { ArrowUpDown } from 'lucide-vue-next'
import { t } from '@/i18n'
defineProps<{ modelValue: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
