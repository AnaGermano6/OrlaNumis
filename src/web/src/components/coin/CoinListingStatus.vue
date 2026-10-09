<template>
  <div v-if="listingStatus" class="mt-4 rounded-md border border-border-subtle bg-surface px-4 py-3">
    <div class="mb-[0.35rem] flex items-center justify-between gap-3">
      <span
        class="inline-block rounded-full px-[0.6rem] py-[0.2rem] text-sm font-semibold"
        :class="{
          'bg-status-success-bg text-gain': listingStatus === 'available',
          'bg-status-error-bg text-loss': listingStatus === 'unavailable',
          'bg-status-warning-bg text-warning': listingStatus === 'unknown',
        }"
      >{{ listingStatus === 'available' ? t('coinCard.available') : listingStatus === 'unavailable' ? t('coinCard.unavailable') : t('coinCard.unknown') }}</span>
      <button class="btn btn-ghost btn-xs" @click="handleDismiss">{{ t('common.dismiss') }}</button>
    </div>
    <p v-if="listingCheckReason" class="my-1 text-chip text-text-secondary">{{ listingCheckReason }}</p>
    <p v-if="listingCheckedAt" class="m-0 text-sm text-text-muted">{{ t('coinDetail.listing.lastChecked', { date: formatDate(listingCheckedAt) }) }}</p>
  </div>
</template>

<script setup lang="ts">
import { updateListingStatus } from '@/api/client'
import { intlLocale, t } from '@/i18n'

const props = defineProps<{
  coinId: number
  listingStatus?: string | null
  listingCheckReason?: string | null
  listingCheckedAt?: string | null
}>()

const emit = defineEmits<{
  dismissed: []
}>()

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString(intlLocale(), {
    month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit',
  })
}

async function handleDismiss() {
  try {
    await updateListingStatus(props.coinId, '')
    emit('dismissed')
  } catch {
    // silently fail
  }
}
</script>
