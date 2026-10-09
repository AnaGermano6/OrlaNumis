<template>
  <select :value="modelValue" @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)" class="form-select h-[38px] min-w-0 flex-1 bg-card px-[0.6rem] py-[0.45rem] text-body transition-colors hover:border-border-accent">
    <option value="">{{ t('collection.allCategories') }}</option>
    <option v-for="cat in categoryOptions" :key="cat" :value="cat">{{ categoryLabel(cat) }}</option>
  </select>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useCoinOptions } from '@/composables/useCoinOptions'
import { t } from '@/i18n'
import { categoryLabel } from '@/i18n/coinVocabulary'

defineProps<{ modelValue: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()

const { categoryOptions, loadOptions } = useCoinOptions()
onMounted(loadOptions)
</script>
