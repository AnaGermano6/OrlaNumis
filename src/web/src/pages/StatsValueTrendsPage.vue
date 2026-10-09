<template>
  <PullToRefresh :on-refresh="handleRefresh">
    <div class="container flex flex-col gap-6">
      <DesktopPageContext :title="t('nav.valueDetails')">
        <template #actions>
          <AppIconButton to="/stats" :title="t('stats.backToStats')"><ArrowLeft :size="20" /></AppIconButton>
        </template>
        <header class="page-header flex flex-nowrap items-center justify-between gap-4">
          <div>
            <p class="section-label">{{ t('stats.insights') }}</p>
            <h1>{{ t('nav.valueDetails') }}</h1>
            <p class="mt-[0.35rem] text-base text-text-secondary">{{ t('stats.value.intro') }}</p>
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
          <p class="text-base text-text-secondary">{{ t('stats.value.intro') }}</p>
        </template>
      </DesktopPageContext>

      <div v-if="isLoading" class="loading-overlay">
        <div class="spinner"></div>
      </div>

      <template v-else>
        <div v-if="stats" class="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(160px,1fr))]">
          <div class="flex flex-col gap-1 rounded-md border border-border-subtle bg-card p-6 text-center">
            <span class="font-display text-xl font-semibold text-gold">{{ formatCurrency(stats.values.totalCurrentValue) }}</span>
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('stats.value.totalValue') }}</span>
          </div>
          <div class="flex flex-col gap-1 rounded-md border border-border-subtle bg-card p-6 text-center">
            <span class="font-display text-xl font-semibold text-text-primary">{{ formatCurrency(stats.values.totalPurchasePrice) }}</span>
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('stats.value.totalInvested') }}</span>
          </div>
          <div class="flex flex-col gap-1 rounded-md border border-border-subtle bg-card p-6 text-center">
            <span
              class="font-display text-xl font-semibold"
              :class="netGainLoss >= 0 ? 'text-gain' : 'text-loss'"
            >
              {{ netGainLoss >= 0 ? '+' : '' }}{{ formatCurrency(netGainLoss) }}
            </span>
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('stats.value.netGainLoss') }}</span>
          </div>
          <div v-if="stats.values.totalPurchasePrice" class="flex flex-col gap-1 rounded-md border border-border-subtle bg-card p-6 text-center">
            <span
              class="font-display text-xl font-semibold"
              :class="roi >= 0 ? 'text-gain' : 'text-loss'"
            >
              {{ roi >= 0 ? '+' : '' }}{{ roi.toFixed(1) }}%
            </span>
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('stats.value.roi') }}</span>
          </div>
        </div>

        <div v-if="stats" class="card flex flex-col gap-5">
          <h2>{{ t('stats.value.averages') }}</h2>
          <div class="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(200px,1fr))]">
            <div class="flex flex-col gap-1">
              <span class="text-sm text-text-muted">{{ t('stats.value.avgPurchase') }}</span>
              <span class="text-lg font-semibold text-text-primary">{{ formatCurrency(stats.values.avgPurchasePrice) }}</span>
            </div>
            <div class="flex flex-col gap-1">
              <span class="text-sm text-text-muted">{{ t('stats.value.avgCurrent') }}</span>
              <span class="text-lg font-semibold text-gold">{{ formatCurrency(stats.values.avgCurrentValue) }}</span>
            </div>
          </div>
        </div>

        <div class="flex flex-col gap-3">
          <div class="flex flex-wrap items-center justify-between gap-4">
            <span class="section-label">{{ t('stats.value.overTime') }}</span>
            <div class="flex gap-[0.35rem]">
              <button
                v-for="tf in timeframes"
                :key="tf.label"
                class="chip chip-sm"
                :class="{ 'border-gold bg-gold-dim text-gold': selectedDays === tf.days }"
                type="button"
                @click="selectedDays = tf.days"
              >
                {{ tf.label }}
              </button>
            </div>
          </div>
          <StatsValueOverTime :history="filteredHistory" />
        </div>

        <StatsPurchasesPerYear />
      </template>
    </div>
  </PullToRefresh>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowLeft } from 'lucide-vue-next'
import { useCoinsStore } from '@/stores/coins'
import PullToRefresh from '@/components/PullToRefresh.vue'
import StatsValueOverTime from '@/components/stats/StatsValueOverTime.vue'
import StatsPurchasesPerYear from '@/components/stats/StatsPurchasesPerYear.vue'
import { formatCurrency } from '@/utils/format'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { t } from '@/i18n'

const store = useCoinsStore()
const isLoading = ref(true)
const stats = computed(() => store.stats)

const timeframes = computed(() => [
  { label: t('stats.value.timeframeAll'), days: 0 },
  { label: t('stats.value.timeframe1y'), days: 365 },
  { label: '6M', days: 180 },
  { label: '3M', days: 90 },
])
const selectedDays = ref(0)

const filteredHistory = computed(() => {
  if (!selectedDays.value) return store.valueHistory
  const cutoff = new Date()
  cutoff.setDate(cutoff.getDate() - selectedDays.value)
  return store.valueHistory.filter((s) => new Date(s.recordedAt) >= cutoff)
})

const roi = computed(() => {
  if (!stats.value?.values.totalPurchasePrice) return 0
  return (
    ((stats.value.values.totalCurrentValue - stats.value.values.totalPurchasePrice) /
      stats.value.values.totalPurchasePrice) *
    100
  )
})

const netGainLoss = computed(() => {
  if (!stats.value) return 0
  return stats.value.values.totalCurrentValue - stats.value.values.totalPurchasePrice
})

async function handleRefresh() {
  isLoading.value = true
  await Promise.all([store.fetchValueHistory(), store.fetchStats()])
  isLoading.value = false
}

onMounted(async () => {
  isLoading.value = true
  await Promise.all([store.fetchValueHistory(), store.fetchStats()])
  isLoading.value = false
})
</script>
