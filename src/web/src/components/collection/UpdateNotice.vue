<template>
  <p
    v-if="latest"
    class="mb-6 flex flex-wrap items-baseline gap-x-3 gap-y-1 rounded-md border border-border-accent bg-card px-4 py-3 text-body text-text-primary"
    role="status"
    data-testid="update-notice"
  >
    <span class="min-w-0 flex-1">{{ t('collectionOverview.update.available', { version: latest }) }}</span>
    <RouterLink :to="{ path: '/admin', query: { tab: 'updates' } }" class="text-gold underline-offset-2 hover:underline">
      {{ t('collectionOverview.update.open') }}
    </RouterLink>
  </p>
</template>

<script setup lang="ts">
// Tells an admin, on the overview, that a new version of Orla Numis is out.
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getUpdateStatus } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { t } from '@/i18n'

const auth = useAuthStore()
const latest = ref('')

onMounted(async () => {
  if (!auth.isAdmin) return
  try {
    const status = (await getUpdateStatus()).data
    latest.value = status.updateAvailable ? status.latest ?? '' : ''
  } catch {
    latest.value = ''
  }
})
</script>
