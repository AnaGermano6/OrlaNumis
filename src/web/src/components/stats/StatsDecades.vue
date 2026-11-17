<template>
  <section v-if="bars.length" class="card flex flex-col" aria-labelledby="stats-decades-title" data-testid="stats-decades">
    <h2 id="stats-decades-title" class="mb-5 text-lg">{{ t('stats.decades.title') }}</h2>
    <!-- As tall as the card beside it lets it be; every decade a scroll away, the latest first in view. -->
    <div ref="scroller" class="flex min-h-44 flex-1 flex-col overflow-x-auto pb-2" data-testid="decades-scroller">
      <ol class="m-0 flex w-max min-w-full flex-1 list-none items-end gap-2 p-0">
        <li
          v-for="bar in bars"
          :key="bar.key"
          class="flex h-full w-10 shrink-0 flex-col items-center justify-end gap-1"
          :title="bar.hint"
          :data-testid="`decade-${bar.key}`"
        >
          <span class="text-micro font-semibold text-text-secondary">{{ bar.count }}</span>
          <span class="block w-full rounded-t-sm bg-gold" :style="{ height: `${Math.max(2, (bar.count / max) * 100)}%` }" aria-hidden="true" />
          <span class="sr-only">{{ bar.hint }}</span>
        </li>
      </ol>
      <div class="mt-1.5 flex w-max min-w-full gap-2 border-t border-border-subtle pt-1.5 text-micro text-text-muted" aria-hidden="true">
        <span v-for="bar in bars" :key="bar.key" class="w-10 shrink-0 text-center">{{ bar.label }}</span>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
// The collection's coins by the decade they were issued in, oldest first:
// those before 1900 together, then each decade, scrolled to the latest.
import { computed, ref } from 'vue'
import type { StatsResponse } from '@/types'
import { useScrollToEnd } from '@/composables/useScrollToEnd'
import { t } from '@/i18n'

const props = defineProps<{ byDecade?: StatsResponse['byDecade'] }>()

/** Decades before this one are counted together. */
const FROM = 1900

const bars = computed(() => {
  const decades = props.byDecade ?? []
  const before = decades.filter((entry) => entry.decade < FROM).reduce((sum, entry) => sum + entry.count, 0)
  const list = before ? [{ key: 'before', label: t('stats.decades.beforeShort', { year: FROM }), count: before, hint: t('stats.decades.beforeHint', { year: FROM, count: before }) }] : []
  return list.concat(decades.filter((entry) => entry.decade >= FROM).map((entry) => ({
    key: String(entry.decade),
    label: t('stats.decades.decade', { decade: entry.decade }),
    count: entry.count,
    hint: t('stats.decades.hint', { from: entry.decade, to: entry.decade + 9, count: entry.count }),
  })))
})

const max = computed(() => Math.max(1, ...bars.value.map((bar) => bar.count)))

const scroller = ref<HTMLElement | null>(null)
// The latest decades first in view; the earlier ones a scroll away.
useScrollToEnd(scroller, () => bars.value.length)
</script>
