<template>
  <div v-if="progress" class="flex min-w-0 flex-col gap-4" :data-testid="`progress-${id}`">
    <!-- Its title and how much of it the collection has, of every country or of the card's. -->
    <div class="min-w-0">
      <h3 class="m-0 font-sans text-base font-semibold text-text-primary">{{ title }}</h3>
      <p class="m-0 mt-0.5 text-body text-text-secondary" data-testid="progress-count">{{ t('stats.progress.count', { owned: formatCount(progress.owned), total: formatCount(progress.total) }) }}</p>
    </div>

    <div class="flex min-w-0 flex-col gap-6 md:flex-row md:items-center">
      <!-- As a ring (that opens their catalogue, when they have one). -->
      <component
        :is="progress.to ? 'router-link' : 'div'"
        :to="progress.to"
        class="relative block h-40 w-40 shrink-0 self-center rounded-full focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2 md:self-auto"
        :aria-label="progress.to ? t('stats.progress.ringLabel', { name: title, percent: formatPercent(progress.owned / progress.total) }) : undefined"
        data-testid="progress-ring"
      >
        <svg viewBox="0 0 100 100" class="h-full w-full -rotate-90" aria-hidden="true">
          <circle cx="50" cy="50" :r="RADIUS" fill="none" stroke="var(--bg-input)" stroke-width="10" />
          <circle
            cx="50"
            cy="50"
            :r="RADIUS"
            fill="none"
            stroke="var(--accent-gold)"
            stroke-width="10"
            stroke-linecap="round"
            :stroke-dasharray="`${(progress.owned / progress.total) * CIRCUMFERENCE} ${CIRCUMFERENCE}`"
          />
        </svg>
        <span class="absolute inset-0 flex items-center justify-center font-display text-3xl font-semibold text-text-primary" data-testid="progress-percent">
          {{ formatPercent(progress.owned / progress.total) }}
        </span>
      </component>

      <!-- Every year side by side, scrolled to the latest: a bar's height is what there is of it, its gold what the collection has. -->
      <div v-if="progress.bars.length > 1" ref="scroller" class="min-w-0 flex-1 overflow-x-auto pb-2" data-testid="progress-scroller">
        <ul class="m-0 flex h-28 w-max list-none items-end gap-2 p-0" :aria-label="t('stats.progress.byYear')">
          <li
            v-for="bar in progress.bars"
            :key="bar.year"
            class="flex h-full w-9 shrink-0 flex-col justify-end"
            :title="bar.hint"
            :data-testid="`progress-bar-${bar.year}`"
          >
            <span class="relative block w-full overflow-hidden rounded-t-sm bg-input" :style="{ height: `${Math.max(4, (bar.total / maxBar) * 100)}%` }">
              <span class="absolute inset-x-0 bottom-0 bg-gold" :style="{ height: `${bar.total ? (bar.owned / bar.total) * 100 : 0}%` }" />
            </span>
            <span class="sr-only">{{ bar.hint }}</span>
          </li>
        </ul>
        <div class="mt-1 flex w-max gap-2 text-micro text-text-muted" aria-hidden="true">
          <span v-for="bar in progress.bars" :key="bar.year" class="w-9 shrink-0 text-center">{{ bar.year }}</span>
        </div>
      </div>
    </div>
  </div>
  <!-- The country chosen has none of these. -->
  <div v-else-if="country" class="min-w-0" :data-testid="`progress-${id}`">
    <h3 class="m-0 font-sans text-base font-semibold text-text-primary">{{ title }}</h3>
    <p class="m-0 mt-0.5 text-body text-text-muted" data-testid="progress-none">{{ t('stats.progress.none') }}</p>
  </div>
</template>

<script setup lang="ts">
// How much of a kind of coin the collection has, of every country or of the
// one the card shows:
// a ring, and every year side by side (scrolled to the latest, the earlier
// ones a scroll away), a bar's height how many coins there are of the year,
// its gold how many of them the collection has.
import { computed, ref } from 'vue'
import { useScrollToEnd } from '@/composables/useScrollToEnd'
import { t } from '@/i18n'
import { formatShare, numberFormat } from '@/utils/numberFormat'

export interface CompletionYear {
  year: number
  total: number
  owned: number
}

export interface CompletionCountry {
  /** An ISO code, or the euro area's for a joint issue. */
  code: string
  owned: number
  total: number
  years: CompletionYear[]
}

const props = defineProps<{
  /** Tells the charts apart: "commemoratives", "circulation". */
  id: string
  title: string
  countries: CompletionCountry[]
  /** Every country's years together; else the countries' are added up. */
  years?: CompletionYear[]
  /** Where a country's (or, for '', every country's) ring leads; none when it leads nowhere. */
  linkFor?: (code: string) => string | undefined
  /** The country shown (an ISO code), or '' for every country. */
  country?: string
}>()

const RADIUS = 40
const CIRCUMFERENCE = 2 * Math.PI * RADIUS

const country = computed(() => props.country ?? '')
const scroller = ref<HTMLElement | null>(null)

function addedUp(countries: CompletionCountry[]): CompletionYear[] {
  const byYear = new Map<number, CompletionYear>()
  for (const entry of countries) {
    for (const year of entry.years) {
      const sum = byYear.get(year.year) ?? { year: year.year, total: 0, owned: 0 }
      byYear.set(year.year, { year: year.year, total: sum.total + year.total, owned: sum.owned + year.owned })
    }
  }
  return [...byYear.values()].sort((a, b) => a.year - b.year)
}

// Every country's, or the one chosen.
const progress = computed(() => {
  const chosen = props.countries.find((entry) => entry.code === country.value)
  // A country with none of these: nothing to show but that.
  if (country.value && !chosen) return null
  const owned = chosen ? chosen.owned : props.countries.reduce((sum, entry) => sum + entry.owned, 0)
  const total = chosen ? chosen.total : props.countries.reduce((sum, entry) => sum + entry.total, 0)
  if (!total) return null
  const years = chosen ? chosen.years : props.years ?? addedUp(props.countries)
  return {
    owned,
    total,
    to: props.linkFor?.(chosen ? chosen.code : ''),
    bars: years.map((year) => ({
      ...year,
      hint: t('stats.progress.barHint', { period: String(year.year), owned: formatCount(year.owned), total: formatCount(year.total) }),
    })),
  }
})

const maxBar = computed(() => Math.max(1, ...(progress.value?.bars ?? []).map((bar) => bar.total)))

// The latest years first in view; the earlier ones a scroll away.
useScrollToEnd(scroller, () => country.value)

function formatCount(value: number): string {
  return numberFormat().format(value)
}

function formatPercent(value: number): string {
  return formatShare(value)
}
</script>
