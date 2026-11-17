<template>
  <!-- The page's title names it ("New releases"). -->
  <section class="mb-8" data-testid="coin-releases">
    <p v-if="failed" class="m-0 text-body text-text-secondary" role="alert">{{ t('collectionOverview.releases.failed') }}</p>
    <p v-else-if="!groups.length" class="m-0 text-body text-text-muted">{{ t('common.loading') }}</p>
    <template v-else>
      <!-- The lists as tabs, their filters beside them, and a rule under both. -->
      <div class="mb-4 flex flex-wrap items-center justify-between gap-x-4 gap-y-3 border-b border-border-subtle pb-3">
        <AppTabs :model-value="source" :tabs="tabItems" id-prefix="releases" :label="t('collectionOverview.releases.sources')" @update:model-value="choose" />
        <div class="flex flex-wrap items-center gap-2">
          <select
            v-model="country"
            class="form-select !w-auto"
            :aria-label="t('collectionOverview.releases.country')"
            data-testid="releases-country"
            @change="limit = PAGE"
          >
            <option value="">{{ t('collectionOverview.releases.allCountries', { count: current?.releases.length ?? 0 }) }}</option>
            <option v-for="option in countries" :key="option.key" :value="option.key">{{ withFlagOfCode(option.key, option.name) }} ({{ option.count }})</option>
          </select>
          <select
            v-model="filter"
            class="form-select !w-auto"
            :aria-label="t('collectionOverview.releases.status')"
            data-testid="releases-status"
            @change="limit = PAGE"
          >
            <option v-for="option in FILTERS" :key="option" :value="option">
              {{ t('collectionOverview.releases.statusOption', { status: t(`collectionOverview.releases.filters.${option}`) }) }}
            </option>
          </select>
          <!-- By metal, on both lists alike; and without the special editions. -->
          <select
            v-if="metals.length"
            :value="metal"
            class="form-select !w-auto"
            :aria-label="t('collectionOverview.releases.metal')"
            data-testid="releases-metal"
            @change="chooseMetal(($event.target as HTMLSelectElement).value as CoinReleaseMetal | '')"
          >
            <option value="">{{ metalLabel('') }}</option>
            <option v-for="option in metals" :key="option" :value="option">{{ metalLabel(option) }}</option>
          </select>
          <select
            v-if="hasSpecialEditions"
            :value="noSpecialEditions ? 'none' : ''"
            class="form-select !w-auto"
            :aria-label="t('collectionOverview.releases.editions')"
            :title="t('collectionOverview.releases.noSpecialEditionsHint')"
            data-testid="releases-editions"
            @change="setNoSpecialEditions(($event.target as HTMLSelectElement).value === 'none')"
          >
            <option value="">{{ t('collectionOverview.releases.allEditions') }}</option>
            <option value="none">{{ t('collectionOverview.releases.noSpecialEditions') }}</option>
          </select>
          <!-- Where the list comes from, on demand. -->
          <button
            v-if="note"
            type="button"
            class="inline-flex h-9 w-9 items-center justify-center rounded-full text-text-muted transition-colors hover:text-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
            :class="showNote ? 'text-gold' : ''"
            :aria-label="t('collectionOverview.releases.about')"
            :title="note"
            :aria-expanded="showNote"
            aria-controls="releases-note"
            data-testid="releases-info"
            @click="showNote = !showNote"
          >
            <Info :size="20" aria-hidden="true" />
          </button>
        </div>
      </div>

      <div :id="`releases-panel-${source}`" role="tabpanel" :aria-labelledby="`releases-tab-${source}`">
        <p v-if="note && showNote" id="releases-note" class="mb-3 mt-0 text-chip text-text-muted" data-testid="releases-note">{{ note }}</p>
        <p v-if="current && !current.releases.length" class="m-0 text-body text-text-secondary">{{ t('collectionOverview.releases.unavailable') }}</p>
        <p v-else-if="!filtered.length" class="m-0 text-body text-text-secondary" data-testid="releases-empty">{{ t('collectionOverview.releases.empty') }}</p>
        <ul v-else class="m-0 grid list-none grid-cols-2 gap-3 p-0 md:grid-cols-3 xl:grid-cols-4">
          <CoinReleaseCard v-for="release in filtered.slice(0, limit)" :key="release.source + release.imageUrl" :release="release" addable wishlistable />
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
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
// New and upcoming coins: the €2 coins (the ECB's commemoratives of this year
// and the last, and the mints' €2 coins it does not list yet) and the
// collector coins of the mints of Portugal, France and Austria, by country
// and metal, marked when the collection has them. Upcoming first, then the
// newest.
import { computed, onMounted, ref } from 'vue'
import { Info } from 'lucide-vue-next'
import { getCoinReleases, type CoinReleaseGroup, type CoinReleaseMetal } from '@/api/client'
import AppTabs from '@/components/AppTabs.vue'
import CoinReleaseCard from '@/components/collection/CoinReleaseCard.vue'
import { intlLocale, t } from '@/i18n'
import { isSpecialEdition, JOINT_ISSUE, releaseCountryName } from '@/utils/coinReleases'
import { withFlagOfCode } from '@/utils/countryFlags'

const FILTERS = ['all', 'upcoming', 'missing'] as const
type Filter = (typeof FILTERS)[number]
const METALS: CoinReleaseMetal[] = ['gold', 'silver', 'base']
/** Cards shown at first, and added by "Show more". */
const PAGE = 8

const groups = ref<CoinReleaseGroup[]>([])
const failed = ref(false)
const source = ref<CoinReleaseGroup['source']>('ecb')
const filter = ref<Filter>('all')
/** An ISO code, JOINT_ISSUE, or '' for every country. */
const country = ref('')
/** One metal, or '' for every one. */
const metal = ref<CoinReleaseMetal | ''>('')
/** Leaves out the special editions: gold coins, and coins in proof or FDC. */
const noSpecialEditions = ref(false)
const limit = ref(PAGE)
/** Where the list comes from, shown under the filters on demand. */
const showNote = ref(false)

const current = computed(() => groups.value.find((group) => group.source === source.value))
const filtered = computed(() => (current.value?.releases ?? []).filter((release) =>
  (!country.value || (release.country || JOINT_ISSUE) === country.value)
  && (filter.value === 'upcoming' ? release.upcoming : filter.value === 'missing' ? !release.owned : true)
  && (!metal.value || !!release.metals?.includes(metal.value))
  && (!noSpecialEditions.value || !isSpecialEdition(release))))

// The metals of the list shown, in a fixed order.
const metals = computed(() => METALS.filter((option) => (current.value?.releases ?? []).some((release) => release.metals?.includes(option))))

const hasSpecialEditions = computed(() => (current.value?.releases ?? []).some(isSpecialEdition))

function metalLabel(option: CoinReleaseMetal | ''): string {
  return t('collectionOverview.releases.metalOption', { metal: t(`collectionOverview.releases.metals.${option || 'all'}`) })
}

// Gold and "without special editions" exclude each other.
function chooseMetal(option: CoinReleaseMetal | '') {
  metal.value = option
  if (option === 'gold') noSpecialEditions.value = false
  limit.value = PAGE
}

function setNoSpecialEditions(value: boolean) {
  noSpecialEditions.value = value
  if (value && metal.value === 'gold') metal.value = ''
  limit.value = PAGE
}

// The countries of the list shown, by name, with their flags and how many coins each has.
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
  metal.value = ''
  limit.value = PAGE
}

const tabItems = computed(() => groups.value.map((group) => ({ id: group.source, label: groupLabel(group) })))

function groupLabel(group: CoinReleaseGroup): string {
  return group.source === 'ecb' ? t('collectionOverview.releases.ecb') : t('collectionOverview.releases.mints')
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
