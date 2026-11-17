<template>
  <section v-if="leaders.length" class="card" aria-labelledby="stats-leaders-title" data-testid="stats-leaders">
    <h2 id="stats-leaders-title" class="mb-5 text-lg">{{ t('stats.leaders.title') }}</h2>
    <ol class="m-0 flex list-none flex-col gap-4 p-0">
      <li v-for="(entry, index) in leaders" :key="entry.key" :data-testid="`leader-${entry.key}`">
        <router-link :to="entry.to" class="group block no-underline" :title="t('stats.leaders.open')">
          <span class="flex items-baseline justify-between gap-3 text-body">
            <span class="min-w-0 truncate text-text-primary group-hover:text-gold">
              <span class="mr-2 text-text-muted">{{ index + 1 }}.</span>{{ entry.name }}
            </span>
            <span class="shrink-0 whitespace-nowrap text-text-secondary">
              {{ t('stats.leaders.count', { percent: formatPercent(entry.owned / entry.total), owned: entry.owned, total: entry.total }) }}
            </span>
          </span>
          <span class="mt-1.5 block h-2 overflow-hidden rounded-full bg-input" aria-hidden="true">
            <span class="block h-full rounded-full bg-gold" :style="{ width: `${(entry.owned / entry.total) * 100}%` }" />
          </span>
        </router-link>
      </li>
    </ol>
  </section>
</template>

<script setup lang="ts">
// The catalogue's countries the collection has most of, as a share of each
// (its €2 commemoratives and collector coins, or Portugal's in escudos): the
// first few, each opening its catalogue.
import { computed } from 'vue'
import type { CatalogueCountry } from '@/api/client'
import { t } from '@/i18n'
import { formatShare } from '@/utils/numberFormat'
import { releaseCountryName } from '@/utils/coinReleases'
import { withFlagOfCode } from '@/utils/countryFlags'

const props = defineProps<{ countries: CatalogueCountry[] }>()

/** The countries shown. */
const SHOWN = 6

const leaders = computed(() => props.countries
  .map((entry) => {
    const total = entry.commemoratives + entry.collector
    const owned = entry.ownedCommemoratives + entry.ownedCollector
    const country = releaseCountryName(entry.country)
    const name = entry.currency ? t('stats.leaders.escudos', { country }) : country
    const query = entry.currency ? `?country=${entry.country}&currency=${entry.currency}` : `?country=${entry.country}`
    return { key: `${entry.country}${entry.currency ?? ''}`, name: withFlagOfCode(entry.country, name), owned, total, to: `/sets/commemoratives${query}` }
  })
  .filter((entry) => entry.total > 0 && entry.owned > 0)
  .sort((a, b) => b.owned / b.total - a.owned / a.total || b.owned - a.owned)
  .slice(0, SHOWN))

function formatPercent(value: number): string {
  return formatShare(value)
}
</script>
