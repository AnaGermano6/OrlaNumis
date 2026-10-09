<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-[300] flex items-center justify-center bg-overlay-60 px-4" @click="$emit('close')">
      <div class="w-full max-w-[320px] rounded-md border border-border-subtle bg-card p-6 shadow-[0_12px_40px_var(--overlay-50)]" @click.stop>
        <h3 class="mb-3 text-base text-heading">{{ t('collection.bulk.assignLocation') }}</h3>
        <div v-if="locations.length" class="flex max-h-[300px] flex-col gap-1.5 overflow-y-auto">
          <button
            class="flex items-center gap-2 rounded-sm border border-border-subtle bg-surface px-3 py-2 text-left text-body italic text-text-secondary transition-colors hover:border-gold hover:text-gold"
            @click="$emit('select', null)"
          >
            <span class="flex shrink-0 items-center justify-center">—</span>
            {{ t('collection.bulk.noLocation') }}
          </button>
          <button
            v-for="location in locations"
            :key="location.id"
            class="flex items-center gap-2 rounded-sm border border-border-subtle bg-surface px-3 py-2 text-left text-body text-text-primary transition-colors hover:border-gold hover:text-gold"
            :disabled="location.type === 'tray'"
            @click="$emit('select', location.id)"
          >
            <span class="flex shrink-0 items-center justify-center"><MapPin :size="14" /></span>
            <span>
              {{ location.name }}
              <small v-if="location.type === 'tray'" class="block text-text-muted">
                {{ t('collection.bulk.trayNeedsSlot') }}
              </small>
            </span>
          </button>
        </div>
        <p v-else class="text-body text-text-muted">{{ t('collection.bulk.noLocations') }}</p>
        <button class="btn btn-secondary btn-sm mt-3" @click="$emit('close')">{{ t('common.cancel') }}</button>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { MapPin } from 'lucide-vue-next'
import type { StorageLocation } from '@/types'
import { t } from '@/i18n'

defineProps<{
  open: boolean
  locations: StorageLocation[]
}>()

defineEmits<{
  select: [locationId: number | null]
  close: []
}>()
</script>
