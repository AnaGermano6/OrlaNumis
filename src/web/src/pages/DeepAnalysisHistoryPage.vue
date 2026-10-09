<template>
  <div class="container">
    <DesktopPageContext :title="t('deepAnalysis.history.title')">
      <template #actions>
        <AppIconButton to="/lookup" :title="t('nav.identifyCoin')">
          <Search :size="20" />
        </AppIconButton>
      </template>
      <div class="page-header">
        <h1>{{ t('deepAnalysis.history.title') }}</h1>
        <RouterLink class="pwa-icon-btn" to="/lookup" :title="t('nav.identifyCoin')" :aria-label="t('nav.identifyCoin')">
          <Search :size="22" />
        </RouterLink>
      </div>
    </DesktopPageContext>

    <div v-if="loading && jobs.length === 0" class="py-12 text-center text-text-secondary">
      {{ t('deepAnalysis.history.loading') }}
    </div>

    <div v-else-if="loadError" role="alert" class="card !px-6 !py-8 text-center text-byzantine">
      {{ loadError }}
    </div>

    <div v-else-if="jobs.length === 0" class="empty-state card !px-6 !py-12 text-text-secondary">
      <ScanSearch :size="48" class="mb-4 text-text-muted" />
      <p>{{ t('deepAnalysis.history.empty') }}</p>
      <p class="mx-auto mt-2 max-w-[360px] text-body text-text-muted">
        {{ t('deepAnalysis.history.emptyHint') }}
      </p>
      <RouterLink class="btn btn-secondary btn-sm mt-4" to="/lookup">{{ t('nav.identifyCoin') }}</RouterLink>
    </div>

    <div v-else class="flex flex-col gap-2">
      <button
        v-for="entry in jobs"
        :key="entry.id"
        type="button"
        class="card !p-0 flex w-full cursor-pointer items-center gap-3 border-0 !px-4 !py-[0.85rem] text-left transition-colors hover:bg-card-hover"
        @click="openJob(entry.id)"
      >
        <div class="min-w-0 flex-1">
          <div class="mb-[0.35rem] flex flex-wrap items-center gap-2">
            <span class="text-base font-semibold text-text-primary">{{ t('deepAnalysis.jobNumber', { id: entry.id }) }}</span>
            <BaseBadge>{{ deepJobStatusLabel(entry.status) }}</BaseBadge>
            <span class="chip-sm">{{ sourceLabel(entry.source) }}</span>
            <span v-if="linkageLabel(entry)" class="chip-sm" :class="linkageClass(entry)">{{ linkageLabel(entry) }}</span>
          </div>
          <div class="text-sm text-text-muted">{{ formatDate(entry.createdAt) }}</div>
        </div>
        <ChevronRight :size="18" class="shrink-0 text-text-muted" />
      </button>

      <div v-if="hasMore" class="py-4 text-center">
        <button class="btn btn-secondary btn-sm" :disabled="loading" @click="loadMore">
          {{ loading ? t('common.loading') : t('common.loadMore') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { ChevronRight, ScanSearch, Search } from 'lucide-vue-next'
import BaseBadge from '@/components/ui/BaseBadge.vue'
import { listDeepIdentificationJobs, getApiErrorMessage } from '@/api/client'
import type { DeepJob, DeepJobSource } from '@/types'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { t, intlLocale } from '@/i18n'
import { deepJobStatusLabel } from '@/utils/format'

const router = useRouter()

const jobs = ref<DeepJob[]>([])
const cursor = ref<string | undefined>(undefined)
const loading = ref(false)
const loadError = ref('')

const hasMore = computed(() => Boolean(cursor.value))

async function fetchPage(reset: boolean) {
  loading.value = true
  loadError.value = ''
  try {
    const res = await listDeepIdentificationJobs({ cursor: reset ? undefined : cursor.value, limit: 20 })
    jobs.value = reset ? (res.data.jobs ?? []) : [...jobs.value, ...(res.data.jobs ?? [])]
    cursor.value = res.data.nextCursor
  } catch (err) {
    loadError.value = getApiErrorMessage(err) || t('deepAnalysis.history.loadFailed')
  } finally {
    loading.value = false
  }
}

function loadMore() {
  fetchPage(false)
}

function openJob(id: number) {
  router.push({ name: 'deep-analysis', params: { jobId: String(id) } })
}

function sourceLabel(source: DeepJobSource): string {
  return source === 'saved_coin' ? t('deepAnalysis.history.savedCoin') : t('nav.identifyCoin')
}

function linkageLabel(entry: DeepJob): string {
  if (!entry.appliedAt) return ''
  if (entry.source === 'saved_coin' && entry.appliedCoinExists === false) return t('deepAnalysis.history.coinRemoved')
  return entry.source === 'saved_coin' ? t('deepAnalysis.history.applied') : t('deepAnalysis.history.savedAsDraft')
}

function linkageClass(entry: DeepJob): string {
  if (entry.source === 'saved_coin' && entry.appliedCoinExists === false) return '!border-byzantine !text-byzantine'
  return ''
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString(intlLocale())
}

onMounted(() => fetchPage(true))
</script>
