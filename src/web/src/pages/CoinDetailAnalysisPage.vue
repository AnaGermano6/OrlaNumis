<template>
  <CoinDetailSectionPageShell :section-title="t('coinDetail.sections.analysis.title')">
    <template #default="{ coin, refresh }">
      <CoinAIAnalysis
        :coin-id="coin.id"
        :obverse-analysis="coin.obverseAnalysis ?? ''"
        :reverse-analysis="coin.reverseAnalysis ?? ''"
        :ai-analysis="coin.aiAnalysis ?? ''"
        :has-obverse="coin.images?.some(i => i.imageType === 'obverse') ?? false"
        :has-reverse="coin.images?.some(i => i.imageType === 'reverse') ?? false"
        @analysis-updated="refresh"
      />
    </template>
  </CoinDetailSectionPageShell>
</template>

<script setup lang="ts">
import CoinDetailSectionPageShell from '@/components/coin/CoinDetailSectionPageShell.vue'
import CoinAIAnalysis from '@/components/coin/CoinAIAnalysis.vue'
import { t } from '@/i18n'
</script>
