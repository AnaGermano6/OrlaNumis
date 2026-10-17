<template>
  <Transition
    enter-active-class="transition-all duration-300 ease-out"
    enter-from-class="translate-y-5 opacity-0"
    leave-active-class="transition-all duration-300 ease-in"
    leave-to-class="translate-y-5 opacity-0"
  >
    <div v-if="visible" class="fixed bottom-[calc(1rem+env(safe-area-inset-bottom))] left-1/2 z-[200] flex w-[calc(100vw-2rem)] max-w-[calc(100vw-2rem)] -translate-x-1/2 flex-col items-stretch gap-[0.6rem] rounded-md border border-gold-dim bg-card px-3 py-3 whitespace-normal shadow-[0_8px_30px_var(--overlay-50)] min-[641px]:bottom-6 min-[641px]:w-auto min-[641px]:max-w-[calc(100vw-2rem)] min-[641px]:flex-row min-[641px]:items-center min-[641px]:gap-4 min-[641px]:px-5 min-[641px]:whitespace-nowrap">
      <span class="text-center text-body font-medium text-text-secondary min-[641px]:text-left">{{ t('collection.bulk.selected', { count: selectedCount }) }}</span>
      <div class="grid grid-cols-2 gap-2 min-[641px]:flex min-[641px]:flex-wrap min-[641px]:justify-center min-[641px]:gap-2">
        <button class="btn btn-ghost btn-sm min-w-0 justify-center" @click="$emit('set')">
          <FolderPlus :size="16" /> {{ t('collection.bulk.addToSet') }}
        </button>
        <button class="btn btn-ghost btn-sm min-w-0 justify-center hover:border-loss hover:text-loss" @click="$emit('delete')">
          <Trash2 :size="16" /> {{ t('common.delete') }}
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { FolderPlus, Trash2 } from 'lucide-vue-next'
import { t } from '@/i18n'

defineProps<{
  visible: boolean
  selectedCount: number
}>()

defineEmits<{
  set: []
  delete: []
}>()
</script>
