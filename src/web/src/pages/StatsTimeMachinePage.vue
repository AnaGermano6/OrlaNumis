<template>
  <div class="container flex flex-col gap-6">
    <DesktopPageContext :title="t('nav.timeMachine')">
      <template #actions>
        <AppIconButton to="/stats" :title="t('stats.backToStats')"><ArrowLeft :size="20" /></AppIconButton>
      </template>
      <header class="page-header flex flex-nowrap items-center justify-between gap-4">
        <div>
          <p class="section-label">{{ t('stats.insights') }}</p>
          <h1>{{ t('nav.timeMachine') }}</h1>
          <p class="mt-[0.35rem] text-base text-text-secondary">
            {{ t('stats.timeMachine.intro') }}
          </p>
        </div>
        <router-link
          class="inline-flex shrink-0 items-center justify-center rounded-sm border border-border-subtle bg-transparent p-[0.4rem] text-text-secondary transition hover:border-border-accent hover:bg-gold-glow hover:text-gold"
          to="/stats"
          :aria-label="t('stats.backToStats')"
        >
          <ArrowLeft :size="20" />
        </router-link>
      </header>
      <template #desktop>
        <p class="text-base text-text-secondary">{{ t('stats.timeMachine.intro') }}</p>
      </template>
    </DesktopPageContext>

    <div v-if="loadingBounds" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="error" class="card flex flex-col gap-2">
      <p class="text-loss">{{ error }}</p>
      <button class="btn btn-secondary self-start" @click="init">{{ t('common.tryAgain') }}</button>
    </div>

    <!-- Nothing is dated, so there is no timeline to scrub. Say so plainly
         rather than rendering an empty slider. -->
    <div v-else-if="!bounds?.hasData" class="card flex flex-col gap-3">
      <h2>{{ t('stats.timeMachine.noDatesTitle') }}</h2>
      <p class="text-text-secondary">
        {{ t('stats.timeMachine.noDatesBody') }}
      </p>
      <router-link class="btn btn-primary self-start" to="/">{{ t('stats.timeMachine.addDates') }}</router-link>
    </div>

    <template v-else>
      <section class="card flex flex-col gap-4">
        <div class="flex flex-wrap items-baseline justify-between gap-3">
          <div>
            <span class="section-label">{{ t('stats.timeMachine.viewing') }}</span>
            <p class="font-display text-2xl font-semibold text-gold">{{ formattedDate }}</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="preset in presets"
              :key="preset.label"
              type="button"
              class="rounded-sm border border-border-subtle px-3 py-1 text-sm text-text-secondary transition hover:border-border-accent hover:text-gold"
              :class="{ 'border-border-accent text-gold': selectedDate === preset.date }"
              @click="selectDate(preset.date)"
            >
              {{ preset.label }}
            </button>
          </div>
        </div>

        <label class="flex flex-col gap-2">
          <span class="sr-only">{{ t('common.date') }}</span>
          <input
            v-model="sliderValue"
            type="range"
            class="w-full accent-[var(--color-gold)]"
            :min="0"
            :max="totalDays"
            step="1"
            :aria-label="t('stats.timeMachine.collectionDate')"
            @input="onScrub"
          />
        </label>

        <div class="flex justify-between text-sm text-text-muted">
          <span>{{ bounds.earliestDate }}</span>
          <span>{{ bounds.latestDate }}</span>
        </div>

        <input
          v-model="selectedDate"
          type="date"
          class="form-input w-auto self-start"
          :min="bounds.earliestDate"
          :max="bounds.latestDate"
          :aria-label="t('stats.timeMachine.jumpToDate')"
          @change="onDateInput"
        />
      </section>

      <div v-if="loadingSnapshot && !snapshot" class="loading-overlay">
        <div class="spinner"></div>
      </div>

      <template v-if="snapshot">
        <div
          class="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(160px,1fr))]"
          :class="{ 'opacity-60': loadingSnapshot }"
        >
          <div class="flex flex-col gap-1 rounded-md border border-border-subtle bg-card p-6 text-center">
            <span class="font-display text-xl font-semibold text-gold">{{ snapshot.coinCount }}</span>
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('stats.timeMachine.coinsOwned') }}</span>
          </div>
          <div class="flex flex-col gap-1 rounded-md border border-border-subtle bg-card p-6 text-center">
            <span class="font-display text-xl font-semibold text-gold">{{ formatCurrency(snapshot.totalValue) }}</span>
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('common.value') }}</span>
          </div>
          <div class="flex flex-col gap-1 rounded-md border border-border-subtle bg-card p-6 text-center">
            <span class="font-display text-xl font-semibold text-text-primary">{{ formatCurrency(snapshot.totalInvested) }}</span>
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('stats.timeMachine.invested') }}</span>
          </div>
          <div class="flex flex-col gap-1 rounded-md border border-border-subtle bg-card p-6 text-center">
            <span
              class="font-display text-xl font-semibold"
              :class="snapshot.unrealizedGain >= 0 ? 'text-gain' : 'text-loss'"
            >
              {{ snapshot.unrealizedGain >= 0 ? '+' : '' }}{{ formatCurrency(snapshot.unrealizedGain) }}
            </span>
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('stats.timeMachine.unrealized') }}</span>
          </div>
          <div class="flex flex-col gap-1 rounded-md border border-border-subtle bg-card p-6 text-center">
            <span class="font-display text-xl font-semibold text-text-primary">{{ snapshot.acquiredInYear }}</span>
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('stats.timeMachine.addedThatYear') }}</span>
          </div>
        </div>

        <!-- Be explicit about what the numbers rest on. Early dates predate the
             valuation scheduler, so most values are purchase prices. -->
        <p v-if="basisNote" class="text-sm text-text-muted">{{ basisNote }}</p>
        <p v-if="snapshot.undatedCoinCount > 0" class="text-sm text-text-muted">
          {{ t('stats.timeMachine.undated', { count: snapshot.undatedCoinCount }) }}
        </p>

        <div v-if="snapshot.coinCount === 0" class="card">
          <p class="text-text-secondary">{{ t('stats.timeMachine.ownedNone') }}</p>
        </div>

        <template v-else>
          <section v-for="group in breakdownGroups" :key="group.title" class="card flex flex-col gap-4">
            <h2>{{ group.title }}</h2>
            <div v-for="entry in group.entries" :key="entry.label" class="flex flex-col gap-1">
              <div class="flex justify-between text-sm">
                <span class="text-text-primary">{{ breakdownLabel(group.key, entry.label) }}</span>
                <span class="text-text-muted">
                  {{ entry.count }} · {{ formatCurrency(entry.value) }}
                </span>
              </div>
              <div class="h-2 w-full overflow-hidden rounded-full bg-input">
                <div
                  class="h-full rounded-full bg-[var(--color-gold)]"
                  :style="{ width: `${percentOf(entry.count, group.entries)}%` }"
                ></div>
              </div>
            </div>
          </section>

          <section class="card flex flex-col gap-3">
            <h2>{{ t('stats.timeMachine.largestHoldings') }}</h2>
            <ol class="flex flex-col gap-2">
              <li
                v-for="(coin, index) in snapshot.topCoins"
                :key="coin.id"
                class="flex items-baseline justify-between gap-3"
              >
                <span class="text-text-primary">
                  <span class="mr-2 text-text-muted">{{ index + 1 }}.</span>
                  <router-link class="hover:text-gold" :to="`/coin/${coin.id}`">{{ coin.name }}</router-link>
                </span>
                <span class="whitespace-nowrap text-text-muted">
                  {{ formatCurrency(coin.value) }}
                  <span v-if="!coin.valueFromHistory" :title="t('stats.timeMachine.purchasePriceNote')">*</span>
                </span>
              </li>
            </ol>
          </section>
        </template>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ArrowLeft } from 'lucide-vue-next'
import { getTimeMachineBounds, getTimeMachineSnapshot, getApiErrorMessage } from '@/api/client'
import { formatCurrency, materialLabel } from '@/utils/format'
import { t, intlLocale } from '@/i18n'
import { categoryLabel, eraLabel } from '@/i18n/coinVocabulary'
import type { TimeMachineBounds, TimeMachineBreakdownEntry, TimeMachineSnapshot } from '@/types'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'

const MS_PER_DAY = 86_400_000
/** Debounce window for scrubbing, so dragging the slider does not fire a request per pixel. */
const SCRUB_DEBOUNCE_MS = 250

const bounds = ref<TimeMachineBounds | null>(null)
const snapshot = ref<TimeMachineSnapshot | null>(null)
const selectedDate = ref('')
const sliderValue = ref(0)
const loadingBounds = ref(true)
const loadingSnapshot = ref(false)
const error = ref('')

let debounceTimer: ReturnType<typeof setTimeout> | null = null
/** Guards against an earlier, slower request overwriting a newer snapshot. */
let requestVersion = 0

function toDate(value: string): Date {
  return new Date(`${value}T00:00:00Z`)
}

function toISO(date: Date): string {
  return date.toISOString().slice(0, 10)
}

const totalDays = computed(() => {
  if (!bounds.value) return 0
  const span = toDate(bounds.value.latestDate).getTime() - toDate(bounds.value.earliestDate).getTime()
  return Math.max(0, Math.round(span / MS_PER_DAY))
})

const formattedDate = computed(() => {
  if (!selectedDate.value) return ''
  return toDate(selectedDate.value).toLocaleDateString(intlLocale(), {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    timeZone: 'UTC',
  })
})

const basisNote = computed(() => {
  const basis = snapshot.value?.valueBasis
  if (!basis) return ''
  const total = basis.fromValuationHistory + basis.fromPurchasePrice
  if (total === 0 || basis.fromPurchasePrice === 0) return ''
  if (basis.fromValuationHistory === 0) {
    return t('stats.timeMachine.allPurchasePrice')
  }
  return t('stats.timeMachine.partialBasis', { fallback: basis.fromPurchasePrice, total })
})

const breakdownGroups = computed(() => {
  if (!snapshot.value) return []
  return [
    { key: 'category', title: t('stats.charts.byCategory'), entries: snapshot.value.byCategory },
    { key: 'material', title: t('stats.charts.byMaterial'), entries: snapshot.value.byMaterial },
    { key: 'era', title: t('stats.charts.byEra'), entries: snapshot.value.byEra },
  ].filter((group) => group.entries.length > 0)
})

function breakdownLabel(key: string, label: string): string {
  if (key === 'material') return materialLabel(label)
  if (key === 'category') return categoryLabel(label)
  if (key === 'era') return eraLabel(label)
  return label
}

const presets = computed(() => {
  if (!bounds.value?.hasData) return []
  const latest = toDate(bounds.value.latestDate)
  const earliest = toDate(bounds.value.earliestDate)

  const back = (years: number) => {
    const d = new Date(latest)
    d.setUTCFullYear(d.getUTCFullYear() - years)
    return d < earliest ? null : toISO(d)
  }

  return [
    { label: t('stats.timeMachine.presets.today'), date: bounds.value.latestDate },
    { label: t('stats.timeMachine.presets.oneYear'), date: back(1) },
    { label: t('stats.timeMachine.presets.threeYears'), date: back(3) },
    { label: t('stats.timeMachine.presets.fiveYears'), date: back(5) },
    { label: t('stats.timeMachine.presets.beginning'), date: bounds.value.earliestDate },
  ].filter((preset): preset is { label: string; date: string } => preset.date !== null)
})

function percentOf(count: number, entries: TimeMachineBreakdownEntry[]): number {
  const max = Math.max(...entries.map((entry) => entry.count), 1)
  return Math.round((count / max) * 100)
}

function dateFromSlider(days: number): string {
  if (!bounds.value) return ''
  const start = toDate(bounds.value.earliestDate)
  start.setUTCDate(start.getUTCDate() + days)
  return toISO(start)
}

function sliderFromDate(value: string): number {
  if (!bounds.value) return 0
  const span = toDate(value).getTime() - toDate(bounds.value.earliestDate).getTime()
  return Math.max(0, Math.min(totalDays.value, Math.round(span / MS_PER_DAY)))
}

function onScrub() {
  selectedDate.value = dateFromSlider(Number(sliderValue.value))
}

function onDateInput() {
  sliderValue.value = sliderFromDate(selectedDate.value)
}

function selectDate(date: string) {
  selectedDate.value = date
  sliderValue.value = sliderFromDate(date)
}

async function loadSnapshot(date: string) {
  if (!date) return
  const version = ++requestVersion
  loadingSnapshot.value = true
  try {
    const res = await getTimeMachineSnapshot(date)
    // A slower earlier request must not clobber a newer one.
    if (version !== requestVersion) return
    snapshot.value = res.data
    error.value = ''
  } catch (err) {
    if (version !== requestVersion) return
    error.value = getApiErrorMessage(err) || t('stats.timeMachine.loadSnapshotFailed')
  } finally {
    if (version === requestVersion) loadingSnapshot.value = false
  }
}

watch(selectedDate, (date) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => void loadSnapshot(date), SCRUB_DEBOUNCE_MS)
})

async function init() {
  loadingBounds.value = true
  error.value = ''
  try {
    const res = await getTimeMachineBounds()
    bounds.value = res.data
    if (res.data.hasData) {
      selectDate(res.data.latestDate)
      await loadSnapshot(res.data.latestDate)
    }
  } catch (err) {
    error.value = getApiErrorMessage(err) || t('stats.timeMachine.loadFailed')
  } finally {
    loadingBounds.value = false
  }
}

onMounted(init)
</script>
