<template>
  <div class="container">
    <div class="form-wrapper grid gap-4">
      <DesktopPageContext :title="t('quickCapture.title')">
        <template #actions>
          <AppIconButton to="/quick-capture/drafts" :title="t('quickCapture.allCaptures')">
            <List :size="20" />
          </AppIconButton>
        </template>
        <div class="page-header">
          <h1>{{ t('quickCapture.title') }}</h1>
          <div v-if="isPwa" class="pwa-actions">
            <RouterLink class="pwa-icon-btn" to="/quick-capture/drafts" :title="t('quickCapture.allCaptures')" :aria-label="t('quickCapture.allCaptures')">
              <List :size="22" />
            </RouterLink>
          </div>
          <div v-else class="header-actions">
            <RouterLink class="btn btn-secondary" to="/quick-capture/drafts">
              <List :size="16" /> {{ t('quickCapture.all') }}
            </RouterLink>
          </div>
        </div>
      </DesktopPageContext>
      <p class="m-0 text-base text-text-secondary">{{ t('quickCapture.intro') }}</p>
      <QuickCaptureForm @saved="onSaved" />
      <div v-if="lastDraft" class="card grid gap-1">
        <strong>{{ t('quickCapture.draftSaved') }}</strong>
        <span class="text-base text-text-secondary">{{ t('quickCapture.draftSavedDetail', { id: lastDraft.id }) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { List } from 'lucide-vue-next'
import QuickCaptureForm from '@/components/quick-capture/QuickCaptureForm.vue'
import type { QuickCaptureDraft } from '@/types'
import { usePwa } from '@/composables/usePwa'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { t } from '@/i18n'

const lastDraft = ref<QuickCaptureDraft | null>(null)
const { isPwa } = usePwa()

function onSaved(draft: QuickCaptureDraft) {
  lastDraft.value = draft
}
</script>
