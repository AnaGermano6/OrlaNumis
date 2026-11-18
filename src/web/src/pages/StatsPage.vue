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
        <!-- A line between the tabs above (each ending with its pieces' condition)
             and the rest: how much of the catalogue the collection has, and on. -->
        <hr class="m-0 border-0 border-t border-border-accent" data-testid="stats-divider" />
        <StatsProgress :catalogue="catalogue" />

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

        <StatsRulers :rulers="rulers" :total="stats.totalCoins" />

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
import StatsCountryLeaders from '@/components/stats/StatsCountryLeaders.vue'
import StatsDecades from '@/components/stats/StatsDecades.vue'
import StatsPreciousMetals from '@/components/stats/StatsPreciousMetals.vue'
import StatsProgress from '@/components/stats/StatsProgress.vue'
import StatsRulers from '@/components/stats/StatsRulers.vue'
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
// Every ruler with its country; from an older server, its first ten without.
const rulers = computed(() => stats.value?.rulers ?? stats.value?.byRuler?.map((entry) => ({ ...entry, country: '' })) ?? [])

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
