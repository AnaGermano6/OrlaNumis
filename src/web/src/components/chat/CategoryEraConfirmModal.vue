<template>
  <BaseModal :open="Boolean(request)" :aria-label="request ? t('chat.confirmField', { field: fieldName(request.fieldLabel) }) : ''" max-width="md:max-w-[360px]" layer="dialog" @close="cancel">
    <template v-if="request">
      <h3 class="mb-2 text-base text-heading">{{ t('chat.confirmField', { field: fieldName(request.fieldLabel) }) }}</h3>
      <p class="mb-3 text-body text-text-secondary">
        <I18nT keypath="chat.confirmFieldBody" :params="{ field: fieldName(request.fieldLabel).toLowerCase() }"><template #value>"<strong class="text-text-primary">{{ request.suggestedValue }}</strong>"</template></I18nT>
      </p>
      <div class="flex max-h-[240px] flex-col gap-1.5 overflow-y-auto">
        <button
          v-for="option in request.options"
          :key="option"
          class="rounded-sm border border-border-subtle bg-surface px-3 py-2 text-left text-body text-text-primary transition-colors hover:border-gold hover:text-gold"
          @click="choose(option)"
        >
          {{ optionLabel(request.fieldLabel, option) }}
        </button>
      </div>
      <button
        class="btn btn-secondary btn-sm mt-3 w-full"
        @click="choose(request.suggestedValue)"
      >
        {{ t('chat.keepSuggested', { value: request.suggestedValue }) }}
      </button>
      <button class="btn btn-ghost btn-sm mt-2 w-full" @click="cancel">{{ t('common.cancel') }}</button>
    </template>
  </BaseModal>
</template>

<script setup lang="ts">
import type { CategoryEraConfirmRequest } from '@/composables/useCoinSearchChat'
import I18nT from '@/components/I18nT.vue'
import { t } from '@/i18n'
import BaseModal from '@/components/ui/BaseModal.vue'
import { categoryLabel, eraLabel } from '@/i18n/coinVocabulary'

defineProps<{
  request: CategoryEraConfirmRequest | null
}>()

const emit = defineEmits<{
  choose: [value: string]
  cancel: []
}>()

function fieldName(label: string): string {
  if (label === 'Category') return t('coinFields.category')
  if (label === 'Era') return t('coinFields.era')
  return label
}

// Options are stored values; show them in the current locale.
function optionLabel(fieldLabel: string, option: string): string {
  return fieldLabel === 'Era' ? eraLabel(option) : categoryLabel(option)
}

function choose(value: string) {
  emit('choose', value)
}

function cancel() {
  emit('cancel')
}
</script>
