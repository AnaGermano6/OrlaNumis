<template>
  <section class="rounded-md border border-border-subtle bg-card p-4" aria-labelledby="map-target-title" data-testid="map-target-progress">
    <h2 id="map-target-title" class="m-0 font-display text-lg font-medium text-heading">
      {{ t('map.target.progress', { target: title, have: progress.have.length, total }) }}
    </h2>
    <div class="mt-2 h-2 overflow-hidden rounded-full bg-input" role="progressbar" :aria-valuenow="progress.have.length" aria-valuemin="0" :aria-valuemax="total">
      <div class="h-full rounded-full bg-gold" :style="{ width: `${total ? (progress.have.length / total) * 100 : 0}%` }"></div>
    </div>
    <p v-if="!progress.missing.length" class="mb-0 mt-3 text-body text-text-secondary">{{ t('map.target.complete') }}</p>
    <template v-else>
      <p class="mb-2 mt-3 text-chip text-text-secondary">{{ t('map.target.missing') }}</p>
      <ul class="m-0 flex list-none flex-wrap gap-2 p-0">
        <li v-for="key in sortedMissing" :key="key">
          <button
            type="button"
            class="chip border border-dashed border-[var(--text-warning)]"
            :class="{ 'bg-gold-glow': key === selectedKey }"
            :aria-pressed="key === selectedKey"
            @click="emit('select', key === selectedKey ? null : key)"
          >
            {{ labelFor(key) }}
          </button>
        </li>
      </ul>
    </template>
  </section>
</template>

<script setup lang="ts">
// How much of a target (the euro countries, Europe, the world) the
// collection covers, and the countries still missing, which can be chosen.
import { computed } from 'vue'
import { intlLocale, t } from '@/i18n'
import type { TargetProgress } from '@/utils/collectionMap'

const props = defineProps<{
  title: string
  progress: TargetProgress
  labelFor: (key: string) => string
  selectedKey: string | null
}>()

const emit = defineEmits<{ select: [key: string | null] }>()

const total = computed(() => props.progress.have.length + props.progress.missing.length)
const sortedMissing = computed(() => [...props.progress.missing].sort((a, b) => props.labelFor(a).localeCompare(props.labelFor(b), intlLocale())))
</script>
