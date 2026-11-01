<template>
  <BaseModal :open="open" :aria-label="t('collection.bulk.applySet')" max-width="md:max-w-[320px]" @close="$emit('close')">
    <h3 class="mb-3 text-base text-heading">{{ t('collection.bulk.applySet') }}</h3>
    <div v-if="sets.length" class="flex max-h-[300px] flex-col gap-1.5 overflow-y-auto">
      <button
        v-for="set in sets"
        :key="set.id"
        class="flex items-center gap-2 rounded-sm border border-border-subtle bg-surface px-3 py-2 text-left text-body text-text-primary transition-colors hover:border-gold hover:text-gold"
        @click="$emit('select', set.id)"
      >
        <span class="h-3 w-3 shrink-0 rounded-full" :style="{ background: set.color }"></span>
        {{ set.name }}
      </button>
    </div>
    <p v-else class="text-body text-text-muted">{{ t('collection.bulk.noSets') }}</p>
    <button class="btn btn-secondary btn-sm mt-3" @click="$emit('close')">{{ t('common.cancel') }}</button>
  </BaseModal>
</template>

<script setup lang="ts">
import type { CollectionSetOption } from '@/types'
import { t } from '@/i18n'
import BaseModal from '@/components/ui/BaseModal.vue'

defineProps<{
  open: boolean
  sets: CollectionSetOption[]
}>()

defineEmits<{
  select: [setId: number]
  close: []
}>()
</script>
