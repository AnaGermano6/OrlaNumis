<template>
  <section v-if="rulers.length" class="card" aria-labelledby="stats-rulers-title" data-testid="stats-rulers">
    <div class="mb-5 flex flex-wrap items-center justify-between gap-3">
      <h2 id="stats-rulers-title" class="m-0 text-lg">{{ t('stats.charts.topRulers') }}</h2>
      <!-- Every ruler, the monarchs or the regimes (republics, states, federations). -->
      <div class="flex rounded-full border border-border-subtle bg-input p-1" role="radiogroup" :aria-label="t('stats.rulers.kind')">
        <button
          v-for="option in KINDS"
          :key="option"
          type="button"
          role="radio"
          class="cursor-pointer rounded-full px-4 py-1.5 text-body transition-colors focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="option === kind ? 'bg-gold font-semibold text-surface' : 'text-text-secondary hover:text-text-primary'"
          :aria-checked="option === kind"
          :data-testid="`rulers-kind-${option}`"
          @click="choose(option)"
        >
          {{ t(`stats.rulers.kinds.${option}`) }}
        </button>
      </div>
    </div>

    <ol v-if="shown.length" class="m-0 flex list-none flex-col gap-1 rounded-md border border-border-subtle bg-input p-2 sm:p-3" data-testid="rulers-list">
      <li v-for="entry in shown" :key="entry.key" :data-testid="`ruler-${entry.key}`">
        <!-- Each opens the gallery with its coins: of its country, found by its name. -->
        <button
          type="button"
          class="group flex w-full cursor-pointer flex-wrap items-center gap-x-4 gap-y-1.5 rounded-sm px-2 py-1.5 text-left transition-colors hover:bg-card focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :title="t('stats.rulers.open', { ruler: entry.label })"
          @click="openGallery({ country: entry.country || undefined, search: entry.ruler })"
        >
          <span class="flex min-w-0 basis-full items-center gap-2.5 sm:basis-60">
            <span class="flex h-6 w-6 shrink-0 items-center justify-center overflow-hidden rounded-full text-base" aria-hidden="true"><CountryFlag :country="entry.country" /></span>
            <span class="truncate text-base text-text-primary">{{ entry.label }}</span>
          </span>
          <span class="block h-3.5 min-w-16 flex-1 overflow-hidden rounded-full bg-card" aria-hidden="true">
            <span class="block h-full rounded-full" :style="{ width: `${Math.max(2, (entry.count / max) * 100)}%`, background: 'linear-gradient(90deg, var(--text-muted), var(--accent-gold))' }" />
          </span>
          <span class="w-10 shrink-0 text-right text-base font-semibold text-text-primary">{{ formatNumber(entry.count) }}</span>
          <span class="flex w-16 shrink-0 items-center justify-end gap-1 text-body text-text-muted">
            {{ formatShare(entry.count / total) }}
            <ArrowRight :size="14" class="opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100" aria-hidden="true" />
          </span>
        </button>
      </li>
    </ol>
    <p v-else class="m-0 text-body text-text-secondary" data-testid="rulers-empty">{{ t('stats.rulers.none') }}</p>

    <!-- The first ten; the others a click away. -->
    <div v-if="filtered.length > SHOWN" class="mt-3 text-center">
      <button
        type="button"
        class="inline-flex cursor-pointer items-center gap-1 rounded-sm text-body text-gold underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        :aria-expanded="expanded"
        data-testid="rulers-toggle"
        @click="expanded = !expanded"
      >
        {{ expanded ? t('stats.rulers.showLess') : t('stats.rulers.showAll', { count: filtered.length }) }}
        <ArrowRight v-if="!expanded" :size="14" aria-hidden="true" />
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
// The rulers the collection has most coins of, each of its country (with its
// flag; the same name in two countries named with them, "Segunda República
// (PT)"), its coins and their share of the collection; every ruler, only the
// monarchs, or only the regimes. Each opens the gallery with its coins. The
// first ten, the others a click away.
import { computed, ref } from 'vue'
import { ArrowRight } from 'lucide-vue-next'
import CountryFlag from '@/components/CountryFlag.vue'
import { useOpenGallery } from '@/composables/useOpenGallery'
import { t } from '@/i18n'
import { countryFlagCode } from '@/utils/countryFlags'
import { formatNumber, formatShare } from '@/utils/numberFormat'

const props = defineProps<{
  rulers: { ruler: string; country: string; count: number }[]
  /** Every coin of the collection, for each ruler's share. */
  total: number
}>()

const KINDS = ['all', 'monarchs', 'regimes'] as const
type Kind = (typeof KINDS)[number]
/** Shown before "Ver todos". */
const SHOWN = 10

/** A republic, a state, a federation, a union, an occupation…: not a person. */
const REGIME = /rep[úu]blic|republik|r[ée]publique|\bestado|\bstate|f[ée]d[ée]ra|uni[ãa]o|\bunion|confedera|commonwealth|territ|col[óo]nia|colony|protetorado|protectorate|administra|governo|government|mandat|ocupa|occupation|regime|junta|provis|revolu|sovi[ée]t|socialis|popular|democr|reg[êe]ncia|regency|zona\b|\bzone/i

const kind = ref<Kind>('all')
const expanded = ref(false)
const openGallery = useOpenGallery()

function isRegime(ruler: string): boolean {
  return REGIME.test(ruler)
}

// A name two countries share is named with each country's code.
const named = computed(() => {
  const countries = new Map<string, Set<string>>()
  for (const entry of props.rulers) countries.set(entry.ruler, (countries.get(entry.ruler) ?? new Set()).add(entry.country))
  return props.rulers.map((entry) => {
    const shared = (countries.get(entry.ruler)?.size ?? 0) > 1
    const code = countryFlagCode(entry.country) ?? entry.country
    return { ...entry, key: `${entry.ruler}|${entry.country}`, label: shared && code ? `${entry.ruler} (${code})` : entry.ruler }
  })
})

const filtered = computed(() => named.value.filter((entry) => kind.value === 'all' || (kind.value === 'regimes') === isRegime(entry.ruler)))
const shown = computed(() => (expanded.value ? filtered.value : filtered.value.slice(0, SHOWN)))
const max = computed(() => Math.max(1, ...filtered.value.map((entry) => entry.count)))

function choose(option: Kind) {
  kind.value = option
  expanded.value = false
}
</script>
