<template>
  <p
    v-if="message"
    class="mb-6 flex flex-wrap items-baseline gap-x-3 gap-y-1 rounded-md border border-status-warning-border bg-status-warning-bg px-4 py-3 text-body text-text-primary"
    role="status"
    data-testid="numista-values-notice"
  >
    <span class="min-w-0 flex-1">{{ message }}</span>
    <RouterLink :to="{ path: '/admin', query: { tab: 'schedules' } }" class="text-gold underline-offset-2 hover:underline">
      {{ t('collectionOverview.numistaValues.open') }}
    </RouterLink>
  </p>
</template>

<script setup lang="ts">
// Tells an admin, on the overview, that the month's Numista quota is running
// low or spent, while the daily update of coin values from Numista is on.
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getNumistaValuesStatus } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { NumistaValueStatus } from '@/types'
import { intlLocale, t } from '@/i18n'

const auth = useAuthStore()
const status = ref<NumistaValueStatus | null>(null)

const message = computed(() => {
  const value = status.value
  if (!value?.enabled || value.quotaLevel === 'ok') return ''
  const count = (n: number) => n.toLocaleString(intlLocale())
  if (value.quotaLevel === 'paused') {
    const date = new Date(`${value.resumesOn}T00:00:00`).toLocaleDateString(intlLocale(), { day: 'numeric', month: 'long' })
    return t('collectionOverview.numistaValues.paused', { date })
  }
  return t('collectionOverview.numistaValues.low', {
    requests: count(value.usage.requests),
    limit: count(value.usage.limit),
    pauseAt: count(value.usage.pauseAt),
  })
})

onMounted(async () => {
  if (!auth.isAdmin) return
  try {
    status.value = (await getNumistaValuesStatus()).data
  } catch {
    status.value = null
  }
})
</script>
