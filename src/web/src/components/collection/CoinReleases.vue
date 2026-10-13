<template>
  <section class="mb-8" aria-labelledby="releases-title" data-testid="coin-releases">
    <div class="mb-1 flex flex-wrap items-baseline justify-between gap-2">
      <h2 id="releases-title" class="m-0 font-display text-lg font-medium text-heading">{{ t('collectionOverview.releases.title') }}</h2>
    </div>
    <p class="mb-3 mt-0 text-body text-text-secondary">{{ t('collectionOverview.releases.intro') }}</p>

    <p v-if="failed" class="m-0 text-body text-text-secondary" role="alert">{{ t('collectionOverview.releases.failed') }}</p>
    <p v-else-if="!groups.length" class="m-0 text-body text-text-muted">{{ t('common.loading') }}</p>
    <template v-else>
      <div class="mb-3 flex flex-wrap items-center gap-2" role="tablist" :aria-label="t('collectionOverview.releases.sources')">
        <button
          v-for="group in groups"
          :key="group.source"
          type="button"
          role="tab"
          class="chip border"
          :class="group.source === source ? 'border-gold text-gold' : 'border-border-subtle'"
          :aria-selected="group.source === source"
          :data-testid="`releases-tab-${group.source}`"
          @click="choose(group.source)"
        >
          {{ groupLabel(group) }}
        </button>
        <span class="mx-1 h-5 w-px bg-border-subtle" aria-hidden="true"></span>
        <select
          v-model="country"
          class="form-select w-auto"
          :aria-label="t('collectionOverview.releases.country')"
          data-testid="releases-country"
          @change="limit = PAGE"
        >
          <option value="">{{ t('collectionOverview.releases.allCountries', { count: current?.releases.length ?? 0 }) }}</option>
          <option v-for="option in countries" :key="option.key" :value="option.key">{{ option.name }} ({{ option.count }})</option>
        </select>
        <button
          v-for="option in FILTERS"
          :key="option"
          type="button"
          class="chip border"
          :class="option === filter ? 'border-gold text-gold' : 'border-border-subtle'"
          :aria-pressed="option === filter"
          :data-testid="`releases-filter-${option}`"
          @click="filter = option; limit = PAGE"
        >
          {{ t(`collectionOverview.releases.filters.${option}`) }}
        </button>
      </div>

      <p v-if="note" class="mb-3 mt-0 text-chip text-text-muted" data-testid="releases-note">{{ note }}</p>
      <p v-if="current && !current.releases.length" class="m-0 text-body text-text-secondary">{{ t('collectionOverview.releases.unavailable') }}</p>
      <p v-else-if="!filtered.length" class="m-0 text-body text-text-secondary" data-testid="releases-empty">{{ t('collectionOverview.releases.empty') }}</p>
      <ul v-else class="m-0 grid list-none grid-cols-2 gap-3 p-0 md:grid-cols-3 xl:grid-cols-4">
        <CoinReleaseCard v-for="release in filtered.slice(0, limit)" :key="release.source + release.imageUrl" :release="release" />
      </ul>
      <button
        v-if="filtered.length > limit"
        type="button"
        class="mt-3 rounded-sm text-body text-gold underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        data-testid="releases-more"
        @click="limit += PAGE"
      >
        {{ t('collectionOverview.releases.more', { count: filtered.length - limit }) }}
      </button>
    </template>
  </section>
</template>

<script setup lang="ts">
// New and upcoming coins: the €2 commemoratives of the ECB's latest list and
// the collector coins of the mints of Portugal, France and Austria, by
// country, marked when the collection has them. Upcoming first, then the newest.
import { computed, onMounted, ref } from 'vue'
import { getCoinReleases, type CoinReleaseGroup } from '@/api/client'
import CoinReleaseCard from '@/components/collection/CoinReleaseCard.vue'
import { intlLocale, t } from '@/i18n'
import { JOINT_ISSUE, releaseCountryName } from '@/utils/coinReleases'

const FILTERS = ['all', 'upcoming', 'missing'] as const
type Filter = (typeof FILTERS)[number]
/** Cards shown at first, and added by "Show more". */
const PAGE = 8

const groups = ref<CoinReleaseGroup[]>([])
const failed = ref(false)
const source = ref<CoinReleaseGroup['source']>('ecb')
const filter = ref<Filter>('all')
/** An ISO code, JOINT_ISSUE, or '' for every country. */
const country = ref('')
const limit = ref(PAGE)

const current = computed(() => groups.value.find((group) => group.source === source.value))
const filtered = computed(() => (current.value?.releases ?? []).filter((release) =>
  (!country.value || (release.country || JOINT_ISSUE) === country.value)
  && (filter.value === 'upcoming' ? release.upcoming : filter.value === 'missing' ? !release.owned : true)))

// The countries of the list shown, by name, with how many coins each has.
const countries = computed(() => {
  const counts = new Map<string, number>()
  for (const release of current.value?.releases ?? []) {
    const key = release.country || JOINT_ISSUE
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }
  return [...counts].map(([key, count]) => ({ key, count, name: releaseCountryName(key) }))
    .sort((a, b) => a.name.localeCompare(b.name, intlLocale()))
})

// The ECB's descriptions are in English and its list may be of last year;
// some mints give the year of a coin, not its month.
const note = computed(() => {
  const group = current.value
  if (group?.source === 'mints') return group.releases.length ? t('collectionOverview.releases.mintsNote') : ''
  if (group?.source !== 'ecb' || !group.year) return ''
  const thisYear = new Date().getFullYear()
  return group.year < thisYear
    ? t('collectionOverview.releases.ecbOlder', { year: group.year, current: thisYear })
    : t('collectionOverview.releases.ecbNote')
})

function choose(next: CoinReleaseGroup['source']) {
  source.value = next
  country.value = ''
  limit.value = PAGE
}

function groupLabel(group: CoinReleaseGroup): string {
  return group.source === 'ecb'
    ? t('collectionOverview.releases.ecb', { year: group.year || '' })
    : t('collectionOverview.releases.mints')
}

onMounted(async () => {
  try {
    groups.value = (await getCoinReleases()).data.groups
    // Start with a list that has something to show.
    if (!current.value?.releases.length) source.value = groups.value.find((group) => group.releases.length)?.source ?? source.value
  } catch {
    failed.value = true
  }
})
</script>
