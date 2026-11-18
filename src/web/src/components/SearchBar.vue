<template>
  <div class="relative flex w-full max-w-[400px] items-center">
    <Search class="pointer-events-none absolute left-3 text-text-muted" :size="16" />
    <input
      type="text"
      class="form-input !py-2 !pr-10 !pl-10 text-body placeholder:text-text-muted"
      :class="box ? 'h-11 !rounded-md' : 'h-10 rounded-full'"
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      :placeholder="placeholder ?? t('collection.searchPlaceholder')"
    />
    <button v-if="modelValue" class="absolute right-2 p-1 text-text-muted transition-colors hover:text-text-primary" @click="$emit('update:modelValue', '')"><X :size="14" /></button>
  </div>
</template>

<script setup lang="ts">
import { Search, X } from 'lucide-vue-next'
import { t } from '@/i18n'
defineProps<{
  modelValue: string
  placeholder?: string
  /** A box with rounded corners, as tall as the toggles beside it, rather than a pill. */
  box?: boolean
}>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
