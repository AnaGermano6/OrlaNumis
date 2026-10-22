<template>
  <PullToRefresh :on-refresh="handleRefresh">
    <div class="container">
      <DesktopPageContext :title="t('stats.collectionStats')">
        <div class="page-header">
          <h1>{{ t('stats.collectionStats') }}</h1>
        </div>
      </DesktopPageContext>

      <div v-if="!stats" class="loading-overlay">
        <div class="spinner"></div>
      </div>

      <div v-else class="flex flex-col gap-6">
        <div id="summary">
          <StatsSummaryCards :stats="stats" />
        </div>

        <StatsBarChart
          v-if="monetarySystems.bars.length"
          :title="t('stats.charts.byMonetarySystem')"
          :items="monetarySystems.bars"
          :fill-class="() => 'fill-era'"
          :wide="true"
          data-testid="stats-monetary-systems"
        >
          <template #label="{ item }">
            <span :title="monetarySystemBarLabel(item.label, monetarySystems.others)">{{ monetarySystemBarLabel(item.label, monetarySystems.others) }}</span>
          </template>
        </StatsBarChart>

        <!-- Categories and eras only tell something when the collection has more than one. -->
        <StatsBarChart
          v-if="categoryItems.length > 1"
          :title="t('stats.charts.byCategory')"
          :items="categoryItems"
          :fill-class="(label: string) => label"
        >
          <template #label="{ item }">
            <span
              class="badge"
              :style="{
                backgroundColor: colorForLabelBackground(item.label),
                color: colorForLabel(item.label),
                border: `1px solid ${colorForLabelBackground(item.label, 30)}`,
              }"
            >{{ categoryLabel(item.label) }}</span>
          </template>
        </StatsBarChart>

        <StatsBarChart
          :title="t('stats.charts.byMaterial')"
          :items="materialItems"
          :fill-class="() => 'fill-material'"
        >
          <template #label="{ item }">
            <span :class="materialClass(item.label)">{{ materialLabel(item.label) }}</span>
          </template>
        </StatsBarChart>

        <StatsBarChart
          v-if="stats.byGrade?.length"
          :title="t('stats.charts.byGrade')"
          :items="gradeItems"
          :fill-class="() => 'fill-grade'"
        />

        <StatsBarChart
          v-if="eraItems.length > 1"
          :title="t('stats.charts.byEra')"
          :items="eraItems"
          :fill-class="() => 'fill-era'"
        >
          <template #label="{ item }">{{ eraLabel(item.label) }}</template>
        </StatsBarChart>

        <StatsBarChart
          v-if="stats.byRuler?.length"
          :title="t('stats.charts.topRulers')"
          :items="rulerItems"
          :fill-class="() => 'fill-ruler'"
          :wide="true"
        />

        <StatsBarChart
          v-if="stats.byPriceRange?.length"
          :title="t('stats.charts.priceRanges')"
          :items="priceRangeItems"
          :fill-class="() => 'fill-price'"
        />

        <StatsHeatMap v-if="categoryItems.length > 1 || eraItems.length > 1" ref="heatMapRef" />
      </div>
    </div>
  </PullToRefresh>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useCoinsStore } from '@/stores/coins'
import PullToRefresh from '@/components/PullToRefresh.vue'
import StatsSummaryCards from '@/components/stats/StatsSummaryCards.vue'
import StatsBarChart from '@/components/stats/StatsBarChart.vue'
import type { BarItem } from '@/components/stats/StatsBarChart.vue'
import StatsHeatMap from '@/components/stats/StatsHeatMap.vue'
import { colorForLabel, colorForLabelBackground } from '@/utils/categoryColor'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import { materialClass, materialLabel } from '@/utils/format'
import { monetarySystemBarLabel, monetarySystemBars } from '@/utils/monetarySystems'
import { t } from '@/i18n'
import { priceRangeLabel } from '@/utils/priceRanges'
import { categoryLabel, eraLabel } from '@/i18n/coinVocabulary'

const store = useCoinsStore()
const stats = computed(() => store.stats)
const heatMapRef = ref<InstanceType<typeof StatsHeatMap>>()

const priceRangeOrder = ['Under $50', '$50 - $200', '$200 - $500', '$500 - $1K', '$1K+']

const categoryItems = computed<BarItem[]>(() =>
  stats.value?.byCategory.map((c) => ({ label: c.category, count: c.count })) ?? [],
)
const materialItems = computed<BarItem[]>(() =>
  stats.value?.byMaterial.map((m) => ({ label: m.material, count: m.count })) ?? [],
)
const gradeItems = computed<BarItem[]>(() =>
  stats.value?.byGrade?.map((g) => ({ label: g.grade, count: g.count })) ?? [],
)
const monetarySystems = computed(() => monetarySystemBars(stats.value?.byMonetarySystem))
const eraItems = computed<BarItem[]>(() =>
  stats.value?.byEra?.map((e) => ({ label: e.era, count: e.count })) ?? [],
)
const rulerItems = computed<BarItem[]>(() =>
  stats.value?.byRuler?.map((r) => ({ label: r.ruler, count: r.count })) ?? [],
)
const priceRangeItems = computed<BarItem[]>(() => {
  if (!stats.value?.byPriceRange) return []
  return [...stats.value.byPriceRange]
    .sort((a, b) => priceRangeOrder.indexOf(a.range) - priceRangeOrder.indexOf(b.range))
    .map((p) => ({ label: priceRangeLabel(p.range), count: p.count }))
})

async function handleRefresh() {
  await Promise.all([
    store.fetchStats(),
    heatMapRef.value?.fetchDistribution(),
  ])
}

onMounted(() => {
  store.fetchStats()
  if (!store.coins.length) store.fetchCoins()
})
</script>
