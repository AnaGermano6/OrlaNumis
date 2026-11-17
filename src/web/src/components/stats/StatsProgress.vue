<template>
  <section v-if="commemoratives.length || circulation.length" class="card" aria-labelledby="stats-progress-title" data-testid="stats-progress">
    <div class="mb-5 flex flex-wrap items-center justify-between gap-3">
      <h2 id="stats-progress-title" class="m-0 text-lg">{{ t('stats.progress.title') }}</h2>
      <div class="flex flex-wrap items-center gap-4">
        <!-- One country for every chart, or all of them. -->
        <select
          v-if="countryOptions.length > 1"
          v-model="country"
          class="form-select !w-auto max-w-full py-1.5 text-body"
          :aria-label="t('stats.progress.country')"
          data-testid="progress-country"
        >
          <option value="">{{ t('stats.progress.allCountries') }}</option>
          <option v-for="option in countryOptions" :key="option.code" :value="option.code">{{ option.label }}</option>
        </select>
        <router-link to="/sets/commemoratives" class="text-body text-gold underline-offset-2 hover:underline">{{ t('stats.progress.open') }}</router-link>
      </div>
    </div>

    <div class="flex flex-col gap-6 divide-y divide-border-subtle [&>*+*]:pt-6">
      <StatsCompletionChart
        id="commemoratives"
        :title="t('stats.progress.commemoratives')"
        :countries="commemoratives"
        :years="catalogue?.commemorativeYears"
        :link-for="commemorativesLink"
        :country="country"
      />
      <StatsCompletionChart id="circulation" :title="t('stats.progress.circulation')" :countries="circulation" :country="country" />
    </div>
  </section>
</template>

<script setup lang="ts">
// How much of the euro coins the collection has: the euro area's €2
// commemoratives (whose rings open their catalogue) and, under them, the
// coins of circulation, the eight of each year of each country; of every
// country, or of the one chosen for both.
import { computed, ref } from 'vue'
import type { CatalogueCountry, CatalogueYear } from '@/api/client'
import StatsCompletionChart, { type CompletionCountry } from '@/components/stats/StatsCompletionChart.vue'
import { intlLocale, t } from '@/i18n'
import { JOINT_ISSUE, releaseCountryName } from '@/utils/coinReleases'
import { withFlagOfCode } from '@/utils/countryFlags'

const props = defineProps<{
  /** The catalogue's countries and its years; null while it loads, or without it. */
  catalogue: {
    countries: CatalogueCountry[]
    commemorativeYears?: CatalogueYear[]
    circulation?: { country: string; years: CatalogueYear[] }[]
  } | null
}>()

const commemoratives = computed<CompletionCountry[]>(() => (props.catalogue?.countries ?? [])
  .filter((entry) => !entry.currency && entry.commemoratives > 0)
  .map((entry) => ({
    code: entry.country || JOINT_ISSUE,
    owned: entry.ownedCommemoratives,
    total: entry.commemoratives,
    years: entry.commemorativeYears ?? [],
  })))

const circulation = computed<CompletionCountry[]>(() => (props.catalogue?.circulation ?? []).map((entry) => ({
  code: entry.country,
  owned: entry.years.reduce((sum, year) => sum + year.owned, 0),
  total: entry.years.reduce((sum, year) => sum + year.total, 0),
  years: entry.years,
})))

/** The country every chart shows (an ISO code, or the euro area's), or '' for all. */
const country = ref('')

// The countries of either chart, by name, with their flags.
const countryOptions = computed(() => [...new Set([...commemoratives.value, ...circulation.value].filter((entry) => entry.total > 0).map((entry) => entry.code))]
  .map((code) => ({ code, name: releaseCountryName(code), label: withFlagOfCode(code, releaseCountryName(code)) }))
  .sort((a, b) => a.name.localeCompare(b.name, intlLocale())))

// Every country's catalogue, or one's (the joint issues are in every country's).
function commemorativesLink(code: string): string {
  return code && code !== JOINT_ISSUE ? `/sets/commemoratives?country=${code}` : '/sets/commemoratives'
}
</script>
