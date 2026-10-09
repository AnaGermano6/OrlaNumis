<template>
  <div class="container mint-map-page">
    <DesktopPageContext :title="t('map.title')">
      <template #actions>
        <AppIconButton to="/stats" :title="t('stats.backToStats')"><ArrowLeft :size="20" /></AppIconButton>
      </template>
      <header class="page-header mint-map-header">
        <div>
          <p class="section-label">{{ t('stats.insights') }}</p>
          <h1>{{ t('map.title') }}</h1>
        </div>
        <router-link class="back-button" to="/stats" :aria-label="t('stats.backToStats')">
          <ArrowLeft :size="20" />
        </router-link>
      </header>
    </DesktopPageContext>

    <div v-if="loading" class="loading-card card" role="status">
      <div class="spinner"></div>
      <span>{{ t('map.loading') }}</span>
    </div>

    <div v-else-if="errorMessage" class="state-card card" role="alert">
      <h2>{{ t('map.unavailable') }}</h2>
      <p>{{ errorMessage }}</p>
      <button class="btn btn-primary" type="button" @click="loadMapData">{{ t('common.tryAgain') }}</button>
    </div>

    <div v-else-if="!mintLocations.length" class="state-card card" role="alert">
      <h2>{{ t('map.noLocations') }}</h2>
      <p>{{ t('map.noLocationsBody') }}</p>
    </div>

    <div v-else-if="!collectionCoins.length" class="state-card card">
      <h2>{{ t('map.noCoins') }}</h2>
      <p>{{ t('map.noCoinsBody') }}</p>
      <router-link class="btn btn-primary" to="/add">{{ t('nav.addCoin') }}</router-link>
    </div>

    <template v-else>
      <section class="summary-row" :aria-label="t('map.summary')">
        <span class="summary-label">{{ t('map.mappedCoins') }}</span>
        <strong class="mapped-count">{{ mappedCoinCount }}</strong>
      </section>

      <div class="map-layout" data-testid="map-layout-enlarged">
        <MintListPanel
          :groups="panelGroups"
          :selected-mint-id="selectedMintId"
          @select-mint="selectMint"
        />

        <MintMapLeaflet
          :groups="aggregation.matched"
          :selected-mint-id="selectedMintId"
          @select-mint="selectMint"
        />
      </div>

      <MintCoinDrawer
        :open="selectedGroup !== null"
        :group="selectedGroup"
        @close="selectedMintId = null"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowLeft } from 'lucide-vue-next'
import { getCoins, getMintLocations, type MintLocationsResponse } from '@/api/client'
import MintMapLeaflet from '@/components/map/MintMapLeaflet.vue'
import MintCoinDrawer from '@/components/map/MintCoinDrawer.vue'
import MintListPanel from '@/components/map/MintListPanel.vue'
import { UNKNOWN_MINT_ID, groupCoinsByMint, type MintGroup } from '@/utils/mintMap'
import type { Coin, MintLocation } from '@/types'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { t } from '@/i18n'

const MAP_PAGE_LIMIT = 100

const selectedMintId = ref<number | null>(null)
const errorMessage = ref('')
const mintLocations = ref<MintLocation[]>([])
const collectionCoins = ref<Coin[]>([])
const mintLocationsLoading = ref(false)
const coinsLoading = ref(false)

const loading = computed(() => mintLocationsLoading.value || coinsLoading.value)
const aggregation = computed(() => groupCoinsByMint(collectionCoins.value, mintLocations.value))

const unknownGroup = computed((): MintGroup | null => {
  const allCoins = [
    ...aggregation.value.unknown,
    ...aggregation.value.unmatched.flatMap((g) => g.coins),
  ]
  if (!allCoins.length) return null
  return {
    mint: { id: UNKNOWN_MINT_ID, displayName: t('map.unknownMint'), lat: 0, lng: 0, aliases: [], region: '' },
    coins: allCoins,
    count: allCoins.length,
  }
})

const panelGroups = computed(() => {
  const groups = [...aggregation.value.matched]
  if (unknownGroup.value) groups.push(unknownGroup.value)
  return groups
})

const selectedGroup = computed((): MintGroup | null => {
  if (selectedMintId.value === UNKNOWN_MINT_ID) return unknownGroup.value ?? null
  return aggregation.value.matched.find((group) => group.mint.id === selectedMintId.value) ?? null
})

const mappedCoinCount = computed(() =>
  aggregation.value.matched.reduce((total, group) => total + group.count, 0),
)

function selectMint(group: MintGroup) {
  selectedMintId.value = group.mint.id
}

function unwrapMintLocations(data: MintLocationsResponse): MintLocation[] {
  return Array.isArray(data) ? data : data.mintLocations ?? []
}

async function fetchAllActiveCollectionCoins(): Promise<Coin[]> {
  const allCoins: Coin[] = []
  let page = 1

  while (true) {
    const res = await getCoins({
      wishlist: 'false',
      sold: 'false',
      page,
      limit: MAP_PAGE_LIMIT,
    })
    const pageCoins = res.data.coins ?? []
    allCoins.push(...pageCoins)
    const total = res.data.total ?? allCoins.length

    if (!pageCoins.length || allCoins.length >= total) break
    page += 1
  }

  return allCoins
}

async function loadMapData() {
  errorMessage.value = ''
  selectedMintId.value = null
  mintLocationsLoading.value = true
  try {
    const locationsRes = await getMintLocations()
    mintLocations.value = unwrapMintLocations(locationsRes.data)
  } catch {
    mintLocations.value = []
    errorMessage.value = t('map.locationsFailed')
  } finally {
    mintLocationsLoading.value = false
  }

  if (errorMessage.value) return

  try {
    coinsLoading.value = true
    collectionCoins.value = await fetchAllActiveCollectionCoins()
  } catch {
    collectionCoins.value = []
    errorMessage.value = t('map.collectionFailed')
  } finally {
    coinsLoading.value = false
  }
}

onMounted(() => {
  loadMapData()
})
</script>

<style scoped>
.mint-map-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.mint-map-header {
  flex-direction: row;
  align-items: center;
}

.back-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.4rem;
  color: var(--text-secondary);
  background: transparent;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  text-decoration: none;
  transition: var(--transition-fast);
  flex-shrink: 0;
}

.back-button:hover {
  color: var(--accent-gold);
  border-color: var(--border-accent);
  background: var(--accent-gold-glow);
}

.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.6rem 1rem;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
}

.summary-label {
  color: var(--text-secondary);
  font-size: 0.9rem;
  font-weight: 500;
}

.mapped-count {
  color: var(--accent-gold);
  font-family: var(--font-family-display);
  font-size: 1.2rem;
  font-weight: 600;
}

.loading-card,
.state-card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1.5rem;
}

.loading-card {
  align-items: center;
  color: var(--text-secondary);
}

.state-card p {
  margin: 0;
  color: var(--text-secondary);
}

.map-layout {
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: 1rem;
  /* default align-items: stretch — both columns fill the same row height */
  /* enlarged from min(70vh, 640px) now that the below-map bucket moved into the panel */
  height: min(80vh, 800px);
}

@media (max-width: 768px) {
  .map-layout {
    grid-template-columns: 1fr;
    height: auto;
  }
}
</style>
