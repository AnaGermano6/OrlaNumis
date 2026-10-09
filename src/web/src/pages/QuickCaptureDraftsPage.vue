<template>
  <div class="container">
    <div class="form-wrapper grid gap-4">
      <DesktopPageContext :title="t('quickCapture.draftsTitle')">
        <template #actions>
          <AppIconButton to="/quick-capture" :title="t('quickCapture.newCapture')">
            <CirclePlus :size="20" />
          </AppIconButton>
        </template>
        <div class="page-header">
          <h1>{{ t('quickCapture.title') }}</h1>
          <div v-if="isPwa" class="pwa-actions">
            <RouterLink class="pwa-icon-btn" to="/quick-capture" :title="t('quickCapture.newCapture')" :aria-label="t('quickCapture.newCapture')">
              <CirclePlus :size="22" />
            </RouterLink>
          </div>
          <div v-else class="header-actions">
            <RouterLink class="btn btn-primary" to="/quick-capture">
              <Plus :size="16" /> {{ t('quickCapture.new') }}
            </RouterLink>
          </div>
        </div>
      </DesktopPageContext>
      <p v-if="loading" class="m-0 text-base text-text-secondary">{{ t('quickCapture.loadingDrafts') }}</p>
      <p v-else-if="error" class="m-0 text-chip text-warning">{{ error }}</p>
      <p v-else-if="drafts.length === 0" class="m-0 text-base text-text-secondary">{{ t('quickCapture.noDrafts') }}</p>
      <div v-else class="grid gap-4">
        <QuickCaptureDraftCard v-for="draft in drafts" :key="draft.id" :draft="draft" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { CirclePlus, Plus } from 'lucide-vue-next'
import { getApiErrorMessage, listQuickCaptureDrafts } from '@/api/client'
import type { QuickCaptureDraft } from '@/types'
import QuickCaptureDraftCard from '@/components/quick-capture/QuickCaptureDraftCard.vue'
import { usePwa } from '@/composables/usePwa'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { t } from '@/i18n'

const drafts = ref<QuickCaptureDraft[]>([])
const loading = ref(true)
const error = ref('')
const { isPwa } = usePwa()

onMounted(async () => {
  try {
    const response = await listQuickCaptureDrafts({ status: 'active', limit: 50 })
    drafts.value = response.data.drafts
  } catch (err) {
    error.value = getApiErrorMessage(err) || t('quickCapture.loadDraftsFailed')
  } finally {
    loading.value = false
  }
})
</script>
