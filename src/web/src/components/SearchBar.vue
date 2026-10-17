<template>
  <div class="relative flex w-full max-w-[400px] items-center">
    <Search class="pointer-events-none absolute left-3 text-text-muted" :size="16" />
    <input
      type="text"
      class="form-input h-10 rounded-full !py-2 !pr-10 !pl-10 text-body placeholder:text-text-muted"
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
defineProps<{ modelValue: string; placeholder?: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
