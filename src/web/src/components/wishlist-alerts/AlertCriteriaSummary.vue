<template>
  <div class="flex flex-wrap items-center gap-1.5">
    <span
      :class="[
        'inline-flex rounded-full border px-2 py-0.5 text-chip',
        alert.isActive
          ? 'border-status-success-border bg-status-success-tint text-status-success-fg'
          : 'border-border-subtle text-text-muted',
      ]"
    >
      {{ alert.isActive ? t('wishlistAlerts.summary.active') : t('wishlistAlerts.summary.disabled') }}
    </span>
    <span class="inline-flex rounded-full border border-border-subtle px-2 py-0.5 text-chip text-gold">{{ cadenceLabel }}</span>
    <span
      v-for="item in criteriaItems"
      :key="item"
      class="inline-flex rounded-full border border-border-subtle bg-card px-2 py-0.5 text-chip text-text-secondary"
    >
      {{ item }}
    </span>
    <span v-if="!criteriaItems.length" class="text-chip text-text-muted">{{ t('wishlistAlerts.summary.noCriteria') }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { WishlistSearchAlert } from '@/types'
import { alertCadenceLabel } from '@/utils/format'
import { t } from '@/i18n'

const props = defineProps<{ alert: WishlistSearchAlert }>()

const cadenceLabel = computed(() => t('wishlistAlerts.summary.cadence', { cadence: alertCadenceLabel(props.alert.cadence).toLowerCase() }))
const criteriaItems = computed(() => {
  const a = props.alert
  const items = [
    a.rulerOrIssuer,
    a.coinType,
    a.mint,
    a.material,
    a.gradeOrCondition,
    a.keywords,
    ...a.sourceFilters,
  ].filter(Boolean)
  const any = t('wishlistAlerts.summary.any')
  if (a.priceMin != null || a.priceMax != null) items.push(t('wishlistAlerts.summary.price', { min: a.priceMin ?? 0, max: a.priceMax ?? any, currency: a.currency }))
  if (a.dateFrom != null || a.dateTo != null) items.push(t('wishlistAlerts.summary.date', { from: a.dateFrom ?? any, to: a.dateTo ?? any }))
  return items
})
</script>
