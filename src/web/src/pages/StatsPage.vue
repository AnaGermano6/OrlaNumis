<template>
  <PullToRefresh :on-refresh="handleRefresh">
    <div class="container flex flex-col gap-6">
      <DesktopPageContext :title="t('nav.stats')">
        <header class="page-header mb-0">
          <h1>{{ t('nav.stats') }}</h1>
        </header>
      </DesktopPageContext>

      <!-- First the collection at a glance: the whole of it, its coins, its banknotes. -->
      <CollectionOverviewTabs ref="overviewRef" />

      <div v-if="!stats" class="loading-overlay">
        <div class="spinner"></div>
      </div>

      <template v-else>
        <!-- How much of the catalogue the collection has; under it, its pieces' condition. -->
        <StatsProgress :catalogue="catalogue" />
        <StatsConservation :by-grade="stats.byGrade" :total="stats.totalCoins" />

        <!-- Its gold and silver; then, side by side and as tall as each other, the countries it has most of and its decades. -->
        <StatsPreciousMetals :metals="stats.preciousMetals" :by-material="stats.byMaterial" />
        <div class="grid gap-6 lg:grid-cols-2">
          <StatsCountryLeaders class="min-w-0" :countries="catalogue?.countries ?? []" />
          <StatsDecades class="min-w-0" :by-decade="stats.byDecade" />
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
          :title="t('stats.charts.byCoinType')"
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
          v-if="eraItems.length > 1"
          :title="t('stats.charts.byEra')"
          :items="eraItems"
          :fill-class="() => 'fill-era'"
        >
          <template #label="{ item }">{{ eraLabel(item.label) }}</template>
        </StatsBarChart>

        <StatsBarChart
          v-if="rulerItems.length"
          :title="t('stats.charts.topRulers')"
          :items="rulerItems"
          :fill-class="() => 'fill-ruler'"
          :wide="true"
        />

        <StatsHeatMap v-if="categoryItems.length > 1 || eraItems.length > 1" ref="heatMapRef" />
      </template>
    </div>
  </PullToRefresh>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useCoinsStore } from '@/stores/coins'
import PullToRefresh from '@/components/PullToRefresh.vue'
import CollectionOverviewTabs from '@/components/collection/CollectionOverviewTabs.vue'
import StatsConservation from '@/components/stats/StatsConservation.vue'
import StatsCountryLeaders from '@/components/stats/StatsCountryLeaders.vue'
import StatsDecades from '@/components/stats/StatsDecades.vue'
import StatsPreciousMetals from '@/components/stats/StatsPreciousMetals.vue'
import StatsProgress from '@/components/stats/StatsProgress.vue'
import { getCoinCatalogueCountries, type CatalogueCountry, type CatalogueYear } from '@/api/client'
import StatsBarChart from '@/components/stats/StatsBarChart.vue'
import type { BarItem } from '@/components/stats/StatsBarChart.vue'
import StatsHeatMap from '@/components/stats/StatsHeatMap.vue'
import { colorForLabel, colorForLabelBackground } from '@/utils/categoryColor'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import { t } from '@/i18n'
import { categoryLabel, eraLabel } from '@/i18n/coinVocabulary'
import { monetarySystemBarLabel, monetarySystemBars } from '@/utils/monetarySystems'

const store = useCoinsStore()
const stats = computed(() => store.stats)
const heatMapRef = ref<InstanceType<typeof StatsHeatMap>>()
const overviewRef = ref<InstanceType<typeof CollectionOverviewTabs>>()
// The catalogue's countries and years, for the progress and the countries' cards.
const catalogue = ref<{ countries: CatalogueCountry[]; commemorativeYears?: CatalogueYear[]; circulation?: { country: string; years: CatalogueYear[] }[] } | null>(null)

async function loadCatalogue() {
  try {
    catalogue.value = (await getCoinCatalogueCountries()).data
  } catch {
    // Without the catalogue, its cards stay out of the way.
  }
}

const categoryItems = computed<BarItem[]>(() =>
  stats.value?.byCategory.map((c) => ({ label: c.category, count: c.count })) ?? [],
)
const monetarySystems = computed(() => monetarySystemBars(stats.value?.byMonetarySystem))
const eraItems = computed<BarItem[]>(() =>
  stats.value?.byEra?.map((e) => ({ label: e.era, count: e.count })) ?? [],
)
const rulerItems = computed<BarItem[]>(() =>
  stats.value?.byRuler?.map((r) => ({ label: r.ruler, count: r.count })) ?? [],
)

async function handleRefresh() {
  await Promise.all([
    store.fetchStats(),
    heatMapRef.value?.fetchDistribution(),
    overviewRef.value?.load(),
    loadCatalogue(),
  ])
}

onMounted(() => {
  store.fetchStats()
  void loadCatalogue()
})
</script>
