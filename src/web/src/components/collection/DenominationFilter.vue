<template>
  <select :value="modelValue" data-testid="denomination-filter" @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)" class="form-select h-10 w-full bg-card py-0 text-body transition-colors hover:border-border-accent">
    <option value="">{{ t('collection.allDenominations') }}</option>
    <option v-for="option in denominations" :key="option.denomination" :value="option.denomination">{{ option.denomination }} ({{ option.count }})</option>
  </select>
</template>

<script setup lang="ts">
// The gallery's face value filter: the collection's face values, the most
// common first, with how many coins have each.
import type { CoinDenominationCount } from '@/api/client'
import { t } from '@/i18n'

defineProps<{ modelValue: string; denominations: CoinDenominationCount[] }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
