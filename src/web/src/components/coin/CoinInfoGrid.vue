<template>
  <div class="mb-6 grid grid-cols-1 gap-3 sm:grid-cols-2">
    <div class="rounded-sm border border-border-subtle bg-card p-3" v-if="purchasePrice != null">
      <span class="section-label !mb-[0.2rem] block">{{ t('coinFields.purchasePrice') }}</span>
      <span class="text-base font-medium text-text-primary">{{ formatCurrency(purchasePrice) }}</span>
    </div>
    <div class="rounded-sm border border-border-subtle bg-card p-3" v-if="currentValue != null">
      <span class="section-label !mb-[0.2rem] block">{{ t('coinFields.currentValue') }}</span>
      <span class="text-base font-medium text-gold">{{ formatCurrency(currentValue) }}</span>
    </div>
    <div class="rounded-sm border border-border-subtle bg-card p-3" v-if="denomination">
      <span class="section-label !mb-[0.2rem] block">{{ t('coinFields.denomination') }}</span>
      <span class="text-base font-medium text-text-primary">{{ denomination }}</span>
    </div>
    <div class="rounded-sm border border-border-subtle bg-card p-3" v-if="era">
      <span class="section-label !mb-[0.2rem] block">{{ t('coinFields.era') }}</span>
      <span class="text-base font-medium text-text-primary">{{ eraLabel(era) }}</span>
    </div>
    <div class="rounded-sm border border-border-subtle bg-card p-3" v-if="mint">
      <span class="section-label !mb-[0.2rem] block">{{ t('coinFields.mint') }}</span>
      <span class="text-base font-medium text-text-primary">{{ mint }}</span>
    </div>
    <div class="rounded-sm border border-border-subtle bg-card p-3" v-if="material">
      <span class="section-label !mb-[0.2rem] block">{{ t('coinFields.material') }}</span>
      <span class="text-base font-medium" :class="`material-${material.toLowerCase()}`">{{ materialLabel(material) }}</span>
    </div>
    <div class="rounded-sm border border-border-subtle bg-card p-3" v-if="weightGrams != null">
      <span class="section-label !mb-[0.2rem] block">{{ t('coinFields.weight') }}</span>
      <span class="text-base font-medium text-text-primary">{{ weightGrams }}g</span>
    </div>
    <div class="rounded-sm border border-border-subtle bg-card p-3" v-if="diameterMm != null">
      <span class="section-label !mb-[0.2rem] block">{{ t('coinFields.diameter') }}</span>
      <span class="text-base font-medium text-text-primary">{{ diameterMm }}mm</span>
    </div>
    <div class="rounded-sm border border-border-subtle bg-card p-3" v-if="grade">
      <span class="section-label !mb-[0.2rem] block">{{ t('coinFields.grade') }}</span>
      <span class="text-base font-medium text-gold">{{ grade }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { formatCurrency, materialLabel } from '@/utils/format'
import { t } from '@/i18n'
import { eraLabel } from '@/i18n/coinVocabulary'

defineProps<{
  purchasePrice?: number | null
  currentValue?: number | null
  denomination?: string | null
  era?: string | null
  mint?: string | null
  material?: string | null
  weightGrams?: number | null
  diameterMm?: number | null
  grade?: string | null
}>()
</script>
