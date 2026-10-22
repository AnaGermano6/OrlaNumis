<template>
  <section aria-labelledby="sets-by-country-title" data-testid="sets-by-country">
    <h2 id="sets-by-country-title" class="sr-only">{{ t('sets.byCountry.title') }}</h2>
    <p class="mb-3 mt-0 text-body text-text-secondary">{{ t('sets.byCountry.intro') }}</p>
    <p v-if="!countries.length" class="text-body text-text-muted" data-testid="sets-by-country-empty">{{ t('sets.byCountry.empty') }}</p>
    <ul v-else class="m-0 flex list-none flex-col gap-2 p-0">
      <li v-for="entry in countries" :key="entry.country">
        <details class="group rounded-md border border-border-subtle bg-card" :data-testid="`sets-country-${entry.country}`">
          <summary class="flex cursor-pointer list-none items-center gap-3 rounded-md px-4 py-3 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2 [&::-webkit-details-marker]:hidden">
            <span class="min-w-0 flex-1">
              <span class="flex items-baseline justify-between gap-3">
                <span class="truncate font-medium text-text-primary">
                  <span v-if="flagOfCode(entry.country)" class="mr-1.5" aria-hidden="true">{{ flagOfCode(entry.country) }}</span>{{ releaseCountryName(entry.country) }}
                </span>
                <span class="shrink-0 text-chip font-semibold" :class="entry.missing ? 'text-text-secondary' : 'text-gold'" data-testid="sets-country-missing">
                  {{ entry.missing ? t('sets.byCountry.missing', { count: entry.missing }) : t('sets.byCountry.complete') }}
                </span>
              </span>
              <span class="mt-1 block text-chip text-text-muted">
                {{ t('sets.byCountry.progress', { owned: entry.owned, total: entry.total }) }} · {{ t('sets.byCountry.sets', { count: entry.sets.length }) }}<template v-if="entry.partial"> · {{ t('sets.byCountry.partial') }}</template>
              </span>
              <span class="mt-2 block h-1.5 overflow-hidden rounded-full bg-input" aria-hidden="true">
                <span class="block h-full rounded-full bg-gold" :style="{ width: `${percent(entry.owned, entry.total)}%` }"></span>
              </span>
            </span>
            <ChevronDown :size="16" class="shrink-0 text-text-muted transition-transform group-open:rotate-180" aria-hidden="true" />
          </summary>
          <ul class="m-0 list-none border-t border-border-subtle px-4 py-1">
            <li v-for="set in entry.sets" :key="set.id">
              <RouterLink :to="`/sets/${set.id}`" class="flex items-center gap-3 rounded-sm py-2 no-underline hover:text-gold focus-visible:outline-2 focus-visible:outline-gold" :data-testid="`sets-country-set-${set.id}`">
                <span class="h-3 w-3 shrink-0 rounded-full" :style="{ backgroundColor: set.color }" aria-hidden="true"></span>
                <span class="min-w-0 flex-1 truncate text-body text-text-primary">{{ set.name }}</span>
                <span class="shrink-0 text-chip text-text-secondary">
                  {{ set.owned === set.total ? t('sets.byCountry.complete') : t('sets.byCountry.missingOf', { count: set.total - set.owned, total: set.total }) }}
                </span>
              </RouterLink>
            </li>
          </ul>
        </details>
      </li>
    </ul>
    <p v-if="otherSets" class="mb-0 mt-3 text-chip text-text-muted" data-testid="sets-by-country-others">{{ t('sets.byCountry.others', { count: otherSets }) }}</p>
  </section>
</template>

<script setup lang="ts">
// The sets' completion country by country: for each country, the places its
// checklist sets have and how many are still missing, set by set. The other
// sets (by filters, free, goal) have no places and are only counted below.
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { ChevronDown } from 'lucide-vue-next'
import { t } from '@/i18n'
import { normalizeCoinSetType, type CoinSetSummary } from '@/types'
import { releaseCountryName } from '@/utils/coinReleases'
import { flagOfCode } from '@/utils/countryFlags'
import { setsByCountry } from '@/utils/setCountries'

const props = defineProps<{ sets: CoinSetSummary[] }>()

const countries = computed(() => setsByCountry(props.sets))
// A checklist whose list is still being read has no places yet, but is one.
const otherSets = computed(() => props.sets.filter((set) => normalizeCoinSetType(set.setType) !== 'checklist').length)

function percent(owned: number, total: number): number {
  return total ? Math.round((owned / total) * 100) : 0
}
</script>
