<template>
  <BaseButton
    type="button"
    variant="ghost"
    size="sm"
    class="min-h-[44px]"
    :disabled="disabled"
    aria-haspopup="dialog"
    :aria-label="title || label || t('deepAnalysis.title')"
    :title="title"
    @click="$emit('click')"
  >
    <Microscope :size="16" aria-hidden="true" />
    {{ label || t('deepAnalysis.title') }}
  </BaseButton>
</template>

<script setup lang="ts">
import { Microscope } from 'lucide-vue-next'
import BaseButton from '@/components/ui/BaseButton.vue'
import { t } from '@/i18n'

withDefaults(defineProps<{
  label?: string
  disabled?: boolean
  /**
   * Explanatory tooltip shown while `disabled` is true (T088/F6) - e.g.
   * because the user is already at the `MaxActivePerUser` limit - so the
   * user sees why the control is unavailable instead of clicking it and
   * getting an error toast after the request round-trips.
   */
  title?: string
}>(), {
  label: undefined,
  disabled: false,
  title: undefined,
})

defineEmits<{ click: [] }>()
</script>
