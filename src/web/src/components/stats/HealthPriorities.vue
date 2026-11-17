<template>
  <section ref="root" class="card" aria-labelledby="health-priorities-title" data-testid="health-priorities">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-wrap items-center gap-3">
        <h2 id="health-priorities-title" class="m-0 text-lg">{{ t('stats.health.priorities') }}</h2>
        <!-- Of one part only: a way back to all of them. -->
        <button
          v-if="dimension"
          type="button"
          class="chip active inline-flex items-center gap-1"
          :aria-label="t('collection.activeFilters.remove', { filter: dimensionLabel })"
          data-testid="health-priorities-dimension"
          @click="emit('clear-dimension')"
        >
          {{ dimensionLabel }} <X :size="12" aria-hidden="true" />
        </button>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <!-- Of one country, and of one thing missing; each choice with how many coins it would show. -->
        <select
          v-if="countryOptions.length"
          v-model="country"
          class="form-select !w-auto max-w-full py-1.5 text-body"
          :aria-label="t('stats.health.countryFilter')"
          data-testid="health-priorities-country"
        >
          <option value="">{{ t('stats.health.allCountries') }}</option>
          <option v-for="option in countryOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
        </select>
        <select
          v-if="missingOptions.length"
          v-model="missing"
          class="form-select !w-auto max-w-full py-1.5 text-body"
          :aria-label="t('stats.health.missingFilter')"
          data-testid="health-priorities-missing"
        >
          <option value="">{{ t('stats.health.allMissing') }}</option>
          <option v-for="option in missingOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
        </select>
        <button
          v-if="coins.length"
          type="button"
          class="btn btn-secondary btn-sm"
          :aria-pressed="reviewing"
          data-testid="health-review-toggle"
          @click="toggleReview"
        >
          {{ reviewing ? t('stats.health.reviewExit') : t('stats.health.review') }}
        </button>
      </div>
    </div>

    <p v-if="failed" class="m-0 text-body text-text-secondary" role="alert">{{ t('stats.health.prioritiesFailed') }}</p>
    <p v-else-if="loaded && !coins.length" class="m-0 text-body text-text-secondary" data-testid="health-priorities-empty">
      {{ dimension ? t('stats.health.allDoneDimension') : t('stats.health.allDone') }}
    </p>

    <!-- Quick review: one coin at a time, everything it misses, and on to the next. -->
    <div v-else-if="reviewing && current" class="flex flex-col gap-5 md:flex-row md:items-start" data-testid="health-review">
      <CoinThumb :coin="current" class="h-40 w-40 shrink-0 self-center md:self-start" />
      <div class="min-w-0 flex-1">
        <p class="m-0 text-chip text-text-muted" data-testid="health-review-position">{{ t('stats.health.reviewPosition', { index: index + 1, total }) }}</p>
        <router-link :to="`/coin/${current.coinId}`" class="mt-1 block text-md font-semibold text-text-primary no-underline hover:text-gold">{{ current.title }}</router-link>
        <ul class="m-0 mt-3 flex list-none flex-col gap-2 p-0">
          <li v-for="item in current.missingItems" :key="item.key" class="flex flex-wrap items-center justify-between gap-3">
            <span class="rounded-full px-3 py-1 text-chip font-medium" :class="SEVERITY[item.severity]">{{ missingLabel(item) }}</span>
            <router-link :to="actionRoute(current.coinId, item.actionHint)" class="btn btn-primary btn-sm">{{ actionLabel(item) }}</router-link>
          </li>
        </ul>
        <div class="mt-5 flex items-center gap-2">
          <button type="button" class="btn btn-secondary btn-sm" :disabled="index === 0" data-testid="health-review-previous" @click="index--">{{ t('common.previous') }}</button>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="index >= total - 1" data-testid="health-review-next" @click="next">{{ t('common.next') }}</button>
        </div>
      </div>
    </div>

    <template v-else>
      <ul class="m-0 flex list-none flex-col divide-y divide-border-subtle p-0">
        <li v-for="entry in coins" :key="entry.coinId" class="flex flex-wrap items-center gap-4 py-3" :data-testid="`health-priority-${entry.coinId}`">
          <CoinThumb :coin="entry" class="h-14 w-14 shrink-0" />
          <router-link :to="`/coin/${entry.coinId}`" class="min-w-0 flex-1 basis-48 font-medium text-text-primary no-underline hover:text-gold">{{ entry.title }}</router-link>
          <span class="flex flex-wrap items-center justify-end gap-3">
            <span class="rounded-full px-3 py-1 text-chip font-medium" :class="SEVERITY[entry.missingItems[0]!.severity]" data-testid="health-priority-issue">
              {{ missingLabel(entry.missingItems[0]!) }}
            </span>
            <span v-if="entry.missingItems.length > 1" class="text-chip text-text-muted" :title="entry.missingItems.slice(1).map(missingLabel).join('\n')">
              {{ t('stats.health.moreIssues', { count: entry.missingItems.length - 1 }) }}
            </span>
            <router-link :to="actionRoute(entry.coinId, entry.missingItems[0]!.actionHint)" class="btn btn-primary btn-sm" data-testid="health-priority-action">
              {{ actionLabel(entry.missingItems[0]!) }}
            </router-link>
          </span>
        </li>
      </ul>
      <button
        v-if="coins.length < total"
        type="button"
        class="mt-3 rounded-sm text-body text-gold underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        data-testid="health-priorities-more"
        @click="loadMore"
      >
        {{ t('stats.health.more', { count: total - coins.length }) }}
      </button>
    </template>
  </section>
</template>

<script setup lang="ts">
// The coins to fix first: those that miss something, the most urgent first
// (of every part of the score, or of the one chosen; of every country or of
// one; missing anything or one thing), each with what it misses most and the
// way to fix it. Quick review shows one coin at a time,
// with all it misses, and goes on to the next.
import { computed, defineComponent, h, ref, watch } from 'vue'
import { Coins, X } from 'lucide-vue-next'
import { getCoinHealthList } from '@/api/client'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import type { CoinHealthItem, HealthChecklistDimension, HealthQuickAction, MissingChecklistItem } from '@/types'
import { intlLocale, t, type MessageKey } from '@/i18n'
import { withCountryFlag } from '@/utils/countryFlags'

const props = defineProps<{ dimension: HealthChecklistDimension | null }>()
const emit = defineEmits<{ 'clear-dimension': [] }>()

/** Coins loaded at a time. */
const PAGE = 10

const SEVERITY: Record<string, string> = {
  high: 'bg-status-error-bg text-status-error-fg',
  medium: 'bg-status-warning-bg text-status-warning-fg',
  low: 'bg-status-info-bg text-status-info-fg',
}

const MISSING: Record<string, MessageKey> = {
  'metadata.denomination': 'stats.health.missing.denomination',
  'metadata.country': 'stats.health.missing.country',
  'metadata.monetarySystem': 'stats.health.missing.monetarySystem',
  'metadata.mint': 'stats.health.missing.mint',
  'metadata.category': 'stats.health.missing.category',
  'metadata.material': 'stats.health.missing.material',
  'metadata.grade': 'stats.health.missing.grade',
  'images.any': 'stats.health.missing.noImage',
  'images.coverage': 'stats.health.missing.oneSide',
  'valuation.value': 'stats.health.missing.noValue',
  'valuation.freshness': 'stats.health.missing.oldValue',
  'ai.analysis': 'stats.health.missing.noAnalysis',
  'ai.coverage': 'stats.health.missing.halfAnalysis',
}

const DIMENSIONS: Record<HealthChecklistDimension, MessageKey> = {
  metadata: 'health.dimensions.metadata',
  images: 'stats.health.dimensions.imageCoverage',
  valuation: 'stats.health.dimensions.valuationFreshness',
  ai: 'stats.health.dimensions.aiCoverage',
}

// A coin's photo, or a coin drawn where it has none.
const CoinThumb = defineComponent({
  props: { coin: { type: Object as () => CoinHealthItem, required: true } },
  setup: (thumb) => () => h('div', { class: 'flex items-center justify-center overflow-hidden rounded-full border border-border-subtle bg-input' }, [
    thumb.coin.imagePath
      ? h(AuthenticatedImage, { mediaPath: thumb.coin.imagePath, alt: thumb.coin.title, class: 'h-full w-full object-cover', loading: 'lazy' }, { fallback: () => h(Coins, { size: 20, strokeWidth: 1, class: 'text-text-primary/30' }) })
      : h(Coins, { size: 20, strokeWidth: 1, class: 'text-text-primary/30', 'aria-hidden': 'true' }),
  ]),
})

const coins = ref<CoinHealthItem[]>([])
/** The country chosen, as written on the coins, or '' for any. */
const country = ref('')
/** The missing item chosen (its key), or '' for any. */
const missing = ref('')
const facets = ref<{ countries: { value: string; count: number }[]; missing: { value: string; count: number }[] }>({ countries: [], missing: [] })

// The countries by name, with their flags and how many coins each would show; the chosen one even with none.
const countryOptions = computed(() => {
  const options = facets.value.countries.filter((facet) => facet.value)
  if (country.value && !options.some((facet) => facet.value === country.value)) options.push({ value: country.value, count: 0 })
  return options
    .sort((a, b) => a.value.localeCompare(b.value, intlLocale()))
    .map((facet) => ({ value: facet.value, label: `${withCountryFlag(facet.value)} (${facet.count})` }))
})

// What is missing, the most coins first; the chosen one even with none.
const missingOptions = computed(() => {
  const options = [...facets.value.missing]
  if (missing.value && !options.some((facet) => facet.value === missing.value)) options.push({ value: missing.value, count: 0 })
  return options.map((facet) => ({ value: facet.value, label: `${missingLabel({ key: facet.value, label: facet.value } as MissingChecklistItem)} (${facet.count})` }))
})
const total = ref(0)
const page = ref(0)
const loaded = ref(false)
const failed = ref(false)
const reviewing = ref(false)
const index = ref(0)
const root = ref<HTMLElement | null>(null)

const current = computed(() => coins.value[index.value])
const dimensionLabel = computed(() => (props.dimension ? t(DIMENSIONS[props.dimension]) : ''))

function missingLabel(item: MissingChecklistItem): string {
  const key = MISSING[item.key]
  return key ? t(key) : item.label
}

// The way to fix it: the photo and the grade among the coin's fields.
function actionLabel(item: MissingChecklistItem): string {
  if (item.key === 'metadata.grade') return t('stats.health.fix.grade')
  if (item.key === 'valuation.freshness') return t('stats.health.fix.update')
  switch (item.actionHint) {
    case 'upload_images': return t('stats.health.fix.photo')
    case 'run_valuation': return t('stats.health.fix.valuate')
    case 'run_ai_analysis': return t('stats.health.fix.analyze')
    default: return t('stats.health.fix.complete')
  }
}

function actionRoute(coinId: number, action: HealthQuickAction): string {
  switch (action) {
    case 'run_valuation': return `/coin/${coinId}/actions`
    case 'run_ai_analysis': return `/coin/${coinId}/analysis`
    default: return `/edit/${coinId}`
  }
}

async function load(nextPage: number) {
  try {
    const { data } = await getCoinHealthList({
      scope: 'priority',
      dimension: props.dimension ?? undefined,
      country: country.value || undefined,
      missing: missing.value || undefined,
      page: nextPage,
      limit: PAGE,
    })
    coins.value = nextPage === 1 ? data.coins : [...coins.value, ...data.coins]
    if (data.facets) facets.value = data.facets
    total.value = data.pagination.total
    page.value = nextPage
    failed.value = false
  } catch {
    failed.value = true
  } finally {
    loaded.value = true
  }
}

function loadMore() {
  void load(page.value + 1)
}

async function next() {
  if (index.value + 1 >= coins.value.length && coins.value.length < total.value) await load(page.value + 1)
  if (index.value + 1 < coins.value.length) index.value++
}

function toggleReview() {
  reviewing.value = !reviewing.value
  index.value = 0
}

// Another country or thing missing: from the first coin again.
watch([country, missing], () => {
  index.value = 0
  void load(1)
})

// Another part: from the first coin again, and into view.
watch(() => props.dimension, () => {
  index.value = 0
  void load(1)
  root.value?.scrollIntoView?.({ behavior: 'smooth', block: 'start' })
})

void load(1)
defineExpose({ load: () => load(1) })
</script>
