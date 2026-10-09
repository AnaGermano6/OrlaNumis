<template>
  <div class="container flex flex-col gap-4">
    <DesktopPageContext :title="t('nav.emperors')">
      <template #actions>
        <AppIconButton to="/sets" :title="t('emperors.backToSets')"><ArrowLeft :size="20" /></AppIconButton>
        <div class="relative">
          <AppIconButton :active="menuOpen" :title="t('emperors.actions')" @click="menuOpen = !menuOpen"><Menu :size="20" /></AppIconButton>
          <div v-if="menuOpen && result" class="absolute right-0 top-[calc(100%+0.5rem)] z-20 w-[220px] rounded-md border border-border-subtle bg-card p-2 shadow-[0_10px_26px_var(--overlay-45)]">
            <router-link to="/sets/emperors/stats" class="inline-flex w-full items-center gap-2 whitespace-nowrap rounded-sm px-3 py-2 text-body text-text-secondary no-underline transition-all hover:bg-card-hover hover:text-text-primary" @click="menuOpen = false">
              <ChartNoAxesColumn :size="15" /> {{ t('nav.stats') }}
            </router-link>
            <router-link v-if="result.suggestions.length" to="/sets/emperors/pursuits" class="inline-flex w-full items-center gap-2 whitespace-nowrap rounded-sm px-3 py-2 text-body text-text-secondary no-underline transition-all hover:bg-card-hover hover:text-text-primary" @click="menuOpen = false">
              <Crown :size="15" /> {{ t('emperors.pursueNext') }}
            </router-link>
          </div>
        </div>
      </template>
      <header class="page-header flex flex-nowrap items-center justify-between gap-4">
        <div>
          <p class="section-label">{{ t('nav.sets') }}</p>
          <h1>{{ t('nav.emperors') }}</h1>
        </div>
        <div class="relative flex shrink-0 items-center gap-2">
          <router-link
            class="inline-flex items-center justify-center rounded-sm border border-border-subtle bg-transparent p-[0.4rem] text-text-secondary transition hover:border-border-accent hover:bg-gold-glow hover:text-gold"
            to="/sets"
            :aria-label="t('emperors.backToSets')"
          >
            <ArrowLeft :size="20" />
          </router-link>
          <button
            class="inline-flex items-center justify-center rounded-sm border border-border-subtle bg-transparent p-[0.4rem] text-text-secondary transition hover:border-border-accent hover:bg-gold-glow hover:text-gold"
            :class="menuOpen ? 'border-border-accent bg-gold-glow text-gold' : ''"
            :aria-label="t('emperors.actions')"
            @click="menuOpen = !menuOpen"
          >
            <Menu :size="20" />
          </button>
          <div v-if="menuOpen && result" class="absolute right-0 top-[calc(100%+0.5rem)] z-20 w-[220px] rounded-md border border-border-subtle bg-card p-2 shadow-[0_10px_26px_var(--overlay-45)]">
            <router-link
              to="/sets/emperors/stats"
              class="inline-flex w-full items-center gap-2 whitespace-nowrap rounded-sm px-3 py-2 text-body text-text-secondary no-underline transition-all hover:bg-card-hover hover:text-text-primary"
              @click="menuOpen = false"
            >
              <ChartNoAxesColumn :size="15" />
              {{ t('nav.stats') }}
            </router-link>
            <router-link
              v-if="result.suggestions.length"
              to="/sets/emperors/pursuits"
              class="inline-flex w-full items-center gap-2 whitespace-nowrap rounded-sm px-3 py-2 text-body text-text-secondary no-underline transition-all hover:bg-card-hover hover:text-text-primary"
              @click="menuOpen = false"
            >
              <Crown :size="15" />
              {{ t('emperors.pursueNext') }}
            </router-link>
          </div>
        </div>
      </header>
      <p class="m-0 -mt-1 text-base italic text-text-secondary">
        {{ t('emperors.intro') }}
      </p>
      <template #desktop>
        <p class="m-0 text-base italic text-text-secondary">{{ t('emperors.intro') }}</p>
      </template>
    </DesktopPageContext>
    <button v-if="menuOpen" class="fixed inset-0 z-10 bg-transparent" :aria-label="t('emperors.closeActions')" @click="menuOpen = false"></button>

    <div v-if="loading" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="!enabled" class="card px-8 py-12 text-center">
      <Crown :size="48" :stroke-width="1" class="mx-auto mb-4 text-text-muted" />
      <h2 class="mb-2 text-xl text-heading">{{ t('emperors.notEnabled') }}</h2>
      <p class="mb-6 text-text-secondary">
        {{ t('emperors.notEnabledBody') }}
      </p>
      <router-link to="/settings" class="btn btn-primary">{{ t('emperors.goToSettings') }}</router-link>
    </div>

    <div v-else-if="errorMessage" class="card px-8 py-12 text-center">
      <p class="text-text-secondary">{{ errorMessage }}</p>
    </div>

    <template v-else-if="result">
      <div
        v-for="dynasty in result.emperor.dynasties"
        :key="dynasty.dynasty"
        class="space-y-2"
      >
        <h2 class="m-0 font-display text-xl font-medium text-heading">{{ dynasty.dynasty }}</h2>
        <section class="card flex flex-col gap-3 p-5">
          <p class="m-0 text-body text-text-secondary">
            {{ t('emperors.ofTotal', { owned: dynasty.owned, total: dynasty.total }) }}
            ({{ formatPct(dynasty.total ? (dynasty.owned / dynasty.total) * 100 : 0) }}%)
          </p>
          <ImperialFigureWellGrid :slots="dynasty.figures" @highlight-updated="load" />
        </section>
      </div>

      <section v-if="result.usurpers" class="flex flex-col gap-3">
        <h2 class="m-0 font-display text-lg font-medium text-gold">
          {{ t('emperors.usurpersHeading', { owned: result.usurpers.owned, total: result.usurpers.total, pct: formatPct(result.usurpers.percentage) }) }}
        </h2>
        <div
          v-for="dynasty in result.usurpers.dynasties"
          :key="dynasty.dynasty"
          class="space-y-2"
        >
          <h3 class="m-0 font-display text-lg text-heading">{{ dynasty.dynasty }}</h3>
          <section class="card flex flex-col gap-3 p-5">
            <p class="m-0 text-body text-text-secondary">{{ t('emperors.ofTotal', { owned: dynasty.owned, total: dynasty.total }) }}</p>
            <ImperialFigureWellGrid :slots="dynasty.figures" @highlight-updated="load" />
          </section>
        </div>
      </section>

      <section v-if="result.empresses" class="flex flex-col gap-3">
        <h2 class="m-0 font-display text-lg font-medium text-gold">
          {{ t('emperors.empressesHeading', { owned: result.empresses.owned, total: result.empresses.total, pct: formatPct(result.empresses.percentage) }) }}
        </h2>
        <div
          v-for="dynasty in result.empresses.dynasties"
          :key="dynasty.dynasty"
          class="space-y-2"
        >
          <h3 class="m-0 font-display text-lg text-heading">{{ dynasty.dynasty }}</h3>
          <section class="card flex flex-col gap-3 p-5">
            <p class="m-0 text-body text-text-secondary">{{ t('emperors.ofTotal', { owned: dynasty.owned, total: dynasty.total }) }}</p>
            <ImperialFigureWellGrid :slots="dynasty.figures" @highlight-updated="load" />
          </section>
        </div>
      </section>

      <section v-if="result.other" class="flex flex-col gap-3">
        <h2 class="m-0 font-display text-lg font-medium text-gold">
          {{ t('emperors.otherHeading', { owned: result.other.owned, total: result.other.total, pct: formatPct(result.other.percentage) }) }}
        </h2>
        <div
          v-for="dynasty in result.other.dynasties"
          :key="dynasty.dynasty"
          class="space-y-2"
        >
          <h3 class="m-0 font-display text-lg text-heading">{{ dynasty.dynasty }}</h3>
          <section class="card flex flex-col gap-3 p-5">
            <p class="m-0 text-body text-text-secondary">{{ t('emperors.ofTotal', { owned: dynasty.owned, total: dynasty.total }) }}</p>
            <ImperialFigureWellGrid :slots="dynasty.figures" @highlight-updated="load" />
          </section>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ArrowLeft, ChartNoAxesColumn, Crown, Menu } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import { getEmperorTrackerProgress, getApiErrorMessage } from '@/api/client'
import ImperialFigureWellGrid from '@/components/emperor-tracker/ImperialFigureWellGrid.vue'
import type { EmperorTrackerResult } from '@/types'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { t } from '@/i18n'

const auth = useAuthStore()

const loading = ref(true)
const enabled = ref(true)
const errorMessage = ref('')
const result = ref<EmperorTrackerResult | null>(null)
const menuOpen = ref(false)

function formatPct(value: number): string {
  return Math.round(value).toString()
}

async function load() {
  loading.value = true
  menuOpen.value = false
  errorMessage.value = ''
  try {
    const res = await getEmperorTrackerProgress()
    result.value = res.data
    enabled.value = true
  } catch (err) {
    const status = (err as { response?: { status?: number } })?.response?.status
    if (status === 403) {
      enabled.value = auth.user?.emperorTrackerEnabled ?? false
    } else {
      errorMessage.value = getApiErrorMessage(err) || t('emperors.loadFailed')
    }
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
