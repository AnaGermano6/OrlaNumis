<template>
  <div class="container flex flex-col gap-6">
    <DesktopPageContext :title="t('emperors.statsTitle')">
      <template #actions>
        <AppIconButton to="/sets/emperors" :title="t('emperors.backToEmperors')"><ArrowLeft :size="20" /></AppIconButton>
      </template>
      <header class="page-header flex flex-nowrap items-center justify-between gap-4">
        <div>
          <p class="section-label">{{ t('nav.emperors') }}</p>
          <h1>{{ t('nav.stats') }}</h1>
          <p class="mt-[0.35rem] text-base text-text-secondary">
            {{ t('emperors.statsIntro') }}
          </p>
        </div>
        <router-link
          class="inline-flex shrink-0 items-center justify-center rounded-sm border border-border-subtle bg-transparent p-[0.4rem] text-text-secondary transition hover:border-border-accent hover:bg-gold-glow hover:text-gold"
          to="/sets/emperors"
          :aria-label="t('emperors.backToEmperors')"
        >
          <ArrowLeft :size="20" />
        </router-link>
      </header>
      <template #desktop>
        <p class="text-base text-text-secondary">{{ t('emperors.statsIntro') }}</p>
      </template>
    </DesktopPageContext>

    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="errorMessage" class="card px-8 py-12 text-center">
      <p class="text-text-secondary">{{ errorMessage }}</p>
    </div>

    <template v-else-if="result">
      <section class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        <article class="card p-4">
          <p class="section-label mb-1">{{ t('emperors.overall') }}</p>
          <p class="m-0 text-lg font-semibold text-heading">
            {{ t('emperors.ofTotal', { owned: totalOwned, total: totalFigures }) }}
            <span class="text-base font-normal text-text-secondary">({{ formatPct(totalPct) }}%)</span>
          </p>
        </article>
        <article class="card p-4">
          <p class="section-label mb-1">{{ t('emperors.remainingAugustuses') }}</p>
          <p class="m-0 text-lg font-semibold text-heading">{{ remainingAugustuses }}</p>
        </article>
        <article class="card p-4">
          <p class="section-label mb-1">{{ t('emperors.completedDynasties') }}</p>
          <p class="m-0 text-lg font-semibold text-heading">{{ t('emperors.ofTotal', { owned: completedDynasties, total: totalDynasties }) }}</p>
        </article>
        <article class="card p-4">
          <p class="section-label mb-1">{{ t('emperors.topDynasty') }}</p>
          <p class="m-0 text-lg font-semibold text-heading">
            {{ bestDynasty?.dynasty ?? t('common.na') }}
            <span v-if="bestDynasty" class="text-base font-normal text-text-secondary">
              ({{ t('emperors.ofTotal', { owned: bestDynasty.owned, total: bestDynasty.total }) }})
            </span>
          </p>
        </article>
        <article class="card p-4">
          <p class="section-label mb-1">{{ t('emperors.pursuitSuggestions') }}</p>
          <p class="m-0 text-lg font-semibold text-heading">{{ result.suggestions.length }}</p>
        </article>
      </section>

      <section v-if="categoryRows.length" class="card p-5">
        <h2 class="m-0 mb-3 text-lg text-heading">{{ t('emperors.categoryCoverage') }}</h2>
        <div class="flex flex-col gap-2">
          <div
            v-for="category in categoryRows"
            :key="category.label"
            class="flex items-center justify-between gap-3 border-b border-border-subtle pb-2 text-body last:border-0 last:pb-0"
          >
            <span class="text-text-secondary">{{ category.label }}</span>
            <span class="font-semibold text-heading">
              {{ t('emperors.ofTotal', { owned: category.owned, total: category.total }) }}
              <span class="font-normal text-text-secondary">({{ formatPct(category.percentage) }}%)</span>
            </span>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowLeft } from 'lucide-vue-next'
import { getApiErrorMessage, getEmperorTrackerProgress } from '@/api/client'
import type { CategoryProgress, DynastyProgress, EmperorTrackerResult } from '@/types'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { t } from '@/i18n'

const loading = ref(true)
const errorMessage = ref('')
const result = ref<EmperorTrackerResult | null>(null)

const categoryRows = computed(() => {
  if (!result.value) return []
  const rows: Array<{ label: string } & CategoryProgress> = []
  if (result.value.usurpers) rows.push({ label: t('emperors.usurpers'), ...result.value.usurpers })
  if (result.value.empresses) rows.push({ label: t('emperors.empresses'), ...result.value.empresses })
  if (result.value.other) rows.push({ label: t('emperors.otherFigures'), ...result.value.other })
  return rows
})

const totalOwned = computed(() => result.value?.emperor.owned ?? 0)
const totalFigures = computed(() => result.value?.emperor.total ?? 0)
const totalPct = computed(() => totalFigures.value > 0 ? (totalOwned.value / totalFigures.value) * 100 : 0)
const remainingAugustuses = computed(() => Math.max(0, (result.value?.emperor.total ?? 0) - (result.value?.emperor.owned ?? 0)))
const emperorDynasties = computed(() => result.value?.emperor.dynasties ?? [])
const totalDynasties = computed(() => emperorDynasties.value.length)
const completedDynasties = computed(() => emperorDynasties.value.filter((dynasty) => dynasty.total > 0 && dynasty.owned >= dynasty.total).length)
const bestDynasty = computed<DynastyProgress | null>(() => {
  if (emperorDynasties.value.length === 0) return null
  return [...emperorDynasties.value].sort((a, b) => {
    const aPct = a.total > 0 ? a.owned / a.total : 0
    const bPct = b.total > 0 ? b.owned / b.total : 0
    if (bPct === aPct) return b.owned - a.owned
    return bPct - aPct
  })[0] ?? null
})

function formatPct(value: number): string {
  return Math.round(value).toString()
}

async function load() {
  loading.value = true
  errorMessage.value = ''
  try {
    const res = await getEmperorTrackerProgress()
    result.value = res.data
  } catch (error) {
    errorMessage.value = getApiErrorMessage(error) || t('emperors.statsFailed')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
