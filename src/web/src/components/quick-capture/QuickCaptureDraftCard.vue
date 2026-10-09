<template>
  <RouterLink
    class="card grid max-w-full grid-cols-[64px_minmax(0,1fr)] gap-3 overflow-hidden no-underline text-text-primary min-[601px]:grid-cols-[76px_minmax(0,1fr)] min-[601px]:gap-4"
    :to="`/quick-capture/drafts/${draft.id}`"
  >
    <AuthenticatedImage
      v-if="previewImage"
      :media-path="previewImage.filePath"
      :alt="draft.workingTitle || t('quickCapture.card.previewAlt')"
      class="h-16 w-16 rounded-sm bg-input object-cover min-[601px]:h-[76px] min-[601px]:w-[76px]"
    />
    <div v-else class="grid h-16 w-16 place-items-center rounded-sm bg-input text-sm text-text-muted min-[601px]:h-[76px] min-[601px]:w-[76px]">{{ t('common.noImage') }}</div>
    <div class="min-w-0 overflow-hidden">
      <h3 class="mb-[0.35rem] break-words">{{ draft.workingTitle || t('quickCapture.card.untitled') }}</h3>
      <div
        v-if="draft.notes"
        class="markdown-rendered mb-[0.35rem] max-h-[8.5rem] overflow-hidden break-words text-body leading-[1.4] text-text-secondary [&_ol]:mb-[0.4rem] [&_p]:mb-[0.4rem] [&_strong]:font-semibold [&_strong]:text-text-primary [&_ul]:mb-[0.4rem]"
        v-html="renderedNotes"
      ></div>
      <p v-else class="mb-[0.35rem] break-words text-body leading-[1.4] text-text-secondary">{{ draft.acquisitionSource || t('quickCapture.card.incomplete') }}</p>
      <div class="flex min-w-0 flex-wrap items-center gap-2 overflow-hidden">
        <span class="chip-sm inline-block max-w-full truncate align-middle">{{ statusLabel }}</span>
        <span v-if="draft.source === 'find_coin_ai'" class="chip-sm inline-block max-w-full truncate align-middle">{{ t('quickCapture.card.aiDraft') }}</span>
        <span v-if="draft.ngcCertNumber" class="chip-sm inline-block max-w-full truncate align-middle">NGC {{ draft.ngcCertNumber }}</span>
        <span
          v-if="draft.selectedNumistaReference"
          class="chip-sm inline-block max-w-full whitespace-normal break-all align-middle"
        >
          {{ t('quickCapture.numistaNumber', { number: draft.selectedNumistaReference.number }) }}
        </span>
        <span class="text-sm text-text-muted">{{ relativeTime }}</span>
      </div>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import type { QuickCaptureDraft } from '@/types'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { renderSafeMarkdown } from '@/composables/useMarkdown'
import { t, intlLocale } from '@/i18n'

const props = defineProps<{ draft: QuickCaptureDraft }>()
const previewImage = computed(() => props.draft.images.find(image => image.isPrimary) ?? props.draft.images[0])
const renderedNotes = computed(() => renderSafeMarkdown(props.draft.notes))

const statusLabel = computed(() => {
  switch (props.draft.status) {
    case 'active': return t('quickCapture.statuses.active')
    case 'promoting': return t('quickCapture.statuses.promoting')
    case 'promoted': return t('quickCapture.statuses.promoted')
    case 'discarded': return t('quickCapture.statuses.discarded')
    default: return props.draft.status
  }
})

const relativeTime = computed(() => {
  const date = new Date(props.draft.updatedAt)
  const diffMs = Date.now() - date.getTime()
  const diffMins = Math.floor(diffMs / 60_000)
  if (diffMins < 1) return t('common.time.justNow')
  if (diffMins < 60) return t('common.time.minutesAgo', { count: diffMins })
  const diffHours = Math.floor(diffMins / 60)
  if (diffHours < 24) return t('common.time.hoursAgo', { count: diffHours })
  const diffDays = Math.floor(diffHours / 24)
  if (diffDays < 30) return t('common.time.daysAgo', { count: diffDays })
  return date.toLocaleDateString(intlLocale())
})
</script>

<style scoped>
/*
 * :deep() audit — markdown-rendered content
 * Target: HTML elements emitted by markdown-it inside .markdown-rendered.
 * Draft body text is rendered by the markdown parser at runtime; the
 * resulting nodes do not carry Vue scope attributes and cannot be styled
 * or Tailwind utilities.
 */
.markdown-rendered :deep(p),
.markdown-rendered :deep(ul),
.markdown-rendered :deep(ol) {
  margin: 0 0 0.4rem;
}

.markdown-rendered :deep(strong) {
  color: var(--text-primary);
  font-weight: 600;
}
</style>
