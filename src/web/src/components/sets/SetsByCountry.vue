<template>
  <section aria-labelledby="sets-by-country-title" data-testid="sets-by-country">
    <h2 id="sets-by-country-title" class="sr-only">{{ t('sets.byCountry.title') }}</h2>
    <p class="mb-4 mt-0 text-base text-text-secondary">{{ t('sets.byCountry.intro') }}</p>
    <p v-if="!countries.length" class="text-body text-text-muted" data-testid="sets-by-country-empty">{{ t('sets.byCountry.empty') }}</p>
    <p v-else-if="!shown.length" class="text-body text-text-muted" data-testid="sets-by-country-no-match">{{ t('sets.byCountry.noMatch') }}</p>

    <!-- Two columns on a wide screen, each as tall as its own cards (a country opened grows alone). -->
    <div v-else class="grid items-start gap-4 lg:grid-cols-2">
      <ul v-for="(column, index) in columns" :key="index" class="m-0 flex list-none flex-col gap-4 p-0">
        <li
          v-for="entry in column"
          :key="entry.country"
          class="card flex flex-col gap-4 !p-5 max-sm:gap-3 max-sm:!p-4"
          :data-testid="`sets-country-${entry.country}`"
        >
          <div>
            <div class="flex min-w-0 flex-col gap-2">
              <!-- Its name and its sets; at the right, how far it is, and its sets shown or hidden. -->
              <div class="flex items-start justify-between gap-3">
                <div class="flex min-w-0 flex-col items-start gap-2">
                  <!-- Its flag beside its name, as in the coin gallery. -->
                  <h3 class="m-0 truncate font-sans text-2xl font-bold leading-tight text-text-primary max-sm:text-lg">
                    <span v-if="flagOfCode(entry.country)" class="mr-2" aria-hidden="true" data-testid="sets-country-flag">{{ flagOfCode(entry.country) }}</span>{{ releaseCountryName(entry.country) }}
                  </h3>
                  <span class="rounded-full bg-gold px-3 py-0.5 text-body font-semibold text-surface" data-testid="sets-country-sets">{{ t('sets.byCountry.sets', { count: entry.sets.length }) }}</span>
                </div>
                <div class="flex shrink-0 items-center gap-3">
                  <span class="font-sans text-4xl font-bold leading-none text-text-primary max-sm:text-2xl" data-testid="sets-country-percent">{{ formatShare(entry.total ? entry.owned / entry.total : 0) }}</span>
                  <button
                    type="button"
                    class="flex h-10 w-10 items-center justify-center rounded-full border-2 border-gold text-gold transition-colors hover:bg-gold hover:text-surface focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2 max-sm:h-8 max-sm:w-8"
                    :aria-expanded="open.has(entry.country)"
                    :aria-label="t(open.has(entry.country) ? 'sets.byCountry.hideSets' : 'sets.byCountry.showSets', { country: releaseCountryName(entry.country) })"
                    data-testid="sets-country-toggle"
                    @click="toggle(entry.country)"
                  >
                    <ChevronRight :size="20" class="transition-transform" :class="open.has(entry.country) ? 'rotate-90' : ''" aria-hidden="true" />
                  </button>
                </div>
              </div>
              <!-- How many of its coins the collection has, and how many it misses. -->
              <div class="flex flex-wrap items-baseline justify-between gap-x-3 text-md text-text-primary max-sm:text-body">
                <span data-testid="sets-country-coins">{{ formatNumber(entry.owned) }} / {{ formatNumber(entry.total) }} <strong>{{ t('sets.byCountry.coins', { count: entry.total }) }}</strong></span>
                <span class="text-text-secondary" data-testid="sets-country-missing">{{ entry.missing ? t('sets.byCountry.missing', { count: entry.missing }) : t('sets.byCountry.complete') }}</span>
              </div>
              <span v-if="entry.partial" class="text-chip text-text-muted">{{ t('sets.byCountry.partial') }}</span>
            </div>
          </div>

          <span class="block h-2 overflow-hidden rounded-full bg-input" aria-hidden="true">
            <span class="block h-full rounded-full bg-gold" :style="{ width: `${percent(entry.owned, entry.total)}%` }"></span>
          </span>

          <!-- Its sets, each with a picture of its coins, its name, its year or coins, and whether it is complete. -->
          <ul v-if="open.has(entry.country)" class="m-0 grid list-none grid-cols-[repeat(auto-fill,minmax(10rem,1fr))] gap-x-4 gap-y-3 border-t border-border-subtle p-0 pt-4" data-testid="sets-country-list">
            <li v-for="set in entry.sets" :key="set.id" class="min-w-0">
              <RouterLink :to="`/sets/${set.id}`" class="flex min-w-0 items-center gap-2 rounded-sm no-underline hover:text-gold focus-visible:outline-2 focus-visible:outline-gold" :title="set.name" :data-testid="`sets-country-set-${set.id}`">
                <span class="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full bg-input">
                  <AuthenticatedImage v-if="set.cover?.path" :media-path="set.cover.path" alt="" loading="lazy" class="h-full w-full object-cover">
                    <template #fallback><span class="h-3 w-3 rounded-full" :style="{ backgroundColor: set.color }"></span></template>
                  </AuthenticatedImage>
                  <img v-else-if="set.cover?.url" :src="set.cover.url" alt="" loading="lazy" class="h-full w-full object-cover" />
                  <span v-else class="h-3 w-3 rounded-full" :style="{ backgroundColor: set.color }"></span>
                </span>
                <span class="min-w-0 flex-1">
                  <span class="block truncate text-body font-semibold text-text-primary">{{ set.title }}</span>
                  <span class="block truncate text-chip text-text-secondary">{{ set.year ?? t('sets.byCountry.setCoins', { owned: formatNumber(set.owned), total: formatNumber(set.total) }) }}</span>
                </span>
                <CircleCheck v-if="set.owned >= set.total" :size="20" class="shrink-0 text-gain" :aria-label="t('sets.byCountry.complete')" data-testid="sets-country-set-complete" />
                <span v-else class="h-4 w-4 shrink-0 rounded-full border-2 border-loss" :title="t('sets.byCountry.missing', { count: set.total - set.owned })" data-testid="sets-country-set-missing"></span>
              </RouterLink>
            </li>
          </ul>
        </li>
      </ul>
    </div>
    <p v-if="otherSets" class="mb-0 mt-4 text-chip text-text-muted" data-testid="sets-by-country-others">{{ t('sets.byCountry.others', { count: otherSets }) }}</p>
  </section>
</template>

<script setup lang="ts">
// The sets' completion country by country: a card for each country, with
// its flag, how many of its coins the sets list and the collection has, and
// how far it is; opened, its sets one by one. A country can be searched for,
// and the countries put in order of progress, of what they miss or of name.
// The other sets (by filters, free, goal) have no places and are only
// counted below.
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { ChevronRight, CircleCheck } from 'lucide-vue-next'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { intlLocale, t } from '@/i18n'
import { normalizeCoinSetType, type CoinSetSummary } from '@/types'
import { releaseCountryName } from '@/utils/coinReleases'
import { flagOfCode } from '@/utils/countryFlags'
import { formatNumber, formatShare } from '@/utils/numberFormat'
import { setsByCountry, type CountryCompletion } from '@/utils/setCountries'

export type CountrySort = 'progress' | 'progressAsc' | 'missing' | 'name'

const props = withDefaults(defineProps<{
  sets: CoinSetSummary[]
  /** Part of a country's name. */
  search?: string
  sort?: CountrySort
}>(), { search: '', sort: 'progress' })

const countries = computed(() => setsByCountry(props.sets))
// A checklist whose list is still being read has no places yet, but is one.
const otherSets = computed(() => props.sets.filter((set) => normalizeCoinSetType(set.setType) !== 'checklist').length)

function fold(value: string): string {
  return value.toLocaleLowerCase(intlLocale()).normalize('NFD').replace(/[̀-ͯ]/g, '')
}

function share(entry: CountryCompletion): number {
  return entry.total ? entry.owned / entry.total : 0
}

const shown = computed(() => {
  const query = fold(props.search.trim())
  const collator = new Intl.Collator(intlLocale())
  const byName = (a: CountryCompletion, b: CountryCompletion) => collator.compare(releaseCountryName(a.country), releaseCountryName(b.country))
  const order: Record<CountrySort, (a: CountryCompletion, b: CountryCompletion) => number> = {
    progress: (a, b) => share(b) - share(a) || byName(a, b),
    progressAsc: (a, b) => share(a) - share(b) || byName(a, b),
    missing: (a, b) => b.missing - a.missing || byName(a, b),
    name: byName,
  }
  return countries.value
    .filter((entry) => !query || fold(releaseCountryName(entry.country)).includes(query) || fold(entry.country) === query)
    .sort(order[props.sort])
})

// On a wide screen, two columns read across (the first, third… country on
// the left), each as tall as its own cards.
const wide = ref(false)
let media: ReturnType<typeof window.matchMedia> | undefined
function onMedia() {
  wide.value = Boolean(media?.matches)
}
onMounted(() => {
  media = typeof window.matchMedia === 'function' ? window.matchMedia('(min-width: 1024px)') : undefined
  onMedia()
  media?.addEventListener('change', onMedia)
})
onBeforeUnmount(() => media?.removeEventListener('change', onMedia))

const columns = computed(() => (wide.value
  ? [shown.value.filter((_, index) => index % 2 === 0), shown.value.filter((_, index) => index % 2 === 1)]
  : [shown.value]))

// The countries whose sets are shown.
const open = reactive(new Set<string>())
function toggle(country: string) {
  if (open.has(country)) open.delete(country)
  else open.add(country)
}

function percent(owned: number, total: number): number {
  return total ? Math.round((owned / total) * 100) : 0
}
</script>
