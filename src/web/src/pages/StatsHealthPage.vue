<template>
  <PullToRefresh :on-refresh="handleRefresh">
    <div class="container flex flex-col gap-6">
      <DesktopPageContext :title="t('nav.health')">
        <template #actions>
          <AppIconButton to="/stats" :title="t('stats.backToStats')"><ArrowLeft :size="20" /></AppIconButton>
        </template>
        <header class="page-header flex flex-nowrap items-center justify-between gap-4">
          <div>
            <p class="section-label">{{ t('stats.insights') }}</p>
            <h1>{{ t('nav.health') }}</h1>
            <p class="mt-[0.35rem] text-base text-text-secondary">{{ t('stats.healthIntro') }}</p>
          </div>
          <router-link class="inline-flex shrink-0 items-center justify-center rounded-sm border border-border-subtle bg-transparent p-[0.4rem] text-text-secondary transition hover:border-border-accent hover:bg-gold-glow hover:text-gold" to="/stats" :aria-label="t('stats.backToStats')">
            <ArrowLeft :size="20" />
          </router-link>
        </header>
        <template #desktop>
          <p class="text-base text-text-secondary">{{ t('stats.healthIntro') }}</p>
        </template>
      </DesktopPageContext>

      <div v-if="healthLoading" class="loading-overlay">
        <div class="spinner"></div>
      </div>

      <!-- The diagnosis, then the coins to fix first (of the part it points to, when one). -->
      <section v-else-if="collectionHealth" class="flex flex-col gap-6">
        <HealthDiagnosis :summary="collectionHealth" :focused="focused" @focus="focused = $event" />
        <HealthPriorities ref="prioritiesRef" :dimension="focused" @clear-dimension="focused = null" />
      </section>

      <CollectionHealthEmptyState v-else />
    </div>
  </PullToRefresh>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowLeft } from 'lucide-vue-next'
import { useCoinsStore } from '@/stores/coins'
import PullToRefresh from '@/components/PullToRefresh.vue'
import HealthDiagnosis from '@/components/stats/HealthDiagnosis.vue'
import HealthPriorities from '@/components/stats/HealthPriorities.vue'
import CollectionHealthEmptyState from '@/components/stats/CollectionHealthEmptyState.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { t } from '@/i18n'
import type { HealthChecklistDimension } from '@/types'

const store = useCoinsStore()
const collectionHealth = computed(() => store.collectionHealth)
const healthLoading = computed(() => store.healthLoading)
/** The part of the score the priority list keeps to, if one. */
const focused = ref<HealthChecklistDimension | null>(null)
const prioritiesRef = ref<InstanceType<typeof HealthPriorities> | null>(null)

async function handleRefresh() {
  await Promise.all([store.fetchCollectionHealth().catch(() => {}), prioritiesRef.value?.load()])
}

onMounted(() => {
  store.fetchCollectionHealth().catch(() => {})
})
</script>
