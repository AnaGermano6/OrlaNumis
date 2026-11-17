<template>
  <div>
    <!-- The legend says the same in words: the bar is only its picture. -->
    <div v-if="separated" class="flex h-3 gap-1" aria-hidden="true">
      <span v-for="part in parts" :key="part.key" class="h-full min-w-2 rounded-full" :style="{ flexGrow: part.count, flexBasis: 0, background: part.color }" />
    </div>
    <div v-else class="flex h-3 gap-0.5 overflow-hidden rounded-full bg-input" aria-hidden="true">
      <span v-for="part in parts" :key="part.key" class="h-full" :style="{ width: `${(part.count / total) * 100}%`, background: part.color }" />
    </div>

    <ul v-if="legend" class="m-0 mt-3 flex list-none flex-wrap gap-x-5 gap-y-2 p-0 text-body">
      <li v-for="part in parts" :key="part.key" :data-testid="`${testIdPrefix}-${part.key}`">
        <component
          :is="part.clickable ? 'button' : 'span'"
          :type="part.clickable ? 'button' : undefined"
          class="inline-flex items-center gap-1.5 rounded-sm text-text-secondary"
          :class="part.clickable ? 'cursor-pointer hover:text-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2' : ''"
          :title="part.hint || undefined"
          @click="part.clickable && emit('choose', part.key)"
        >
          <span class="h-2.5 w-2.5 shrink-0 rounded-full" :style="{ background: part.color }" aria-hidden="true" />
          <span><strong class="font-semibold text-text-primary">{{ formatCount(part.count) }}</strong> {{ part.label }} <span class="text-text-muted">({{ formatPercent(part.count / total) }})</span></span>
        </component>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
// A bar split into parts, with a legend: each part's count, name and share of
// the total. A part can be chosen (to open what it counts).
import { formatShare, numberFormat } from '@/utils/numberFormat'

export interface StackedBarPart {
  key: string
  label: string
  count: number
  /** A colour of the theme: "var(--accent-gold)". */
  color: string
  hint?: string
  clickable?: boolean
}

withDefaults(defineProps<{
  parts: StackedBarPart[]
  total: number
  /** Starts each legend entry's test id: "<prefix>-<key>". */
  testIdPrefix?: string
  /** Shows the legend under the bar (a card can lay out its own). */
  legend?: boolean
  /** Each part a rounded bar of its own, apart from the next; the smallest still seen. */
  separated?: boolean
}>(), { testIdPrefix: 'part', legend: true, separated: false })

const emit = defineEmits<{ choose: [key: string] }>()

function formatCount(value: number): string {
  return numberFormat().format(value)
}

function formatPercent(value: number): string {
  return formatShare(value)
}
</script>
