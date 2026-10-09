<template>
  <CoinDetailSectionPageShell :section-title="t('coinDetail.sections.notes.title')">
    <template #default="{ coin }">
      <div v-if="coin.notes" class="mt-0">
        <div class="section-content-card">
          <p class="text-base text-text-secondary whitespace-pre-wrap mb-0">{{ coin.notes }}</p>
        </div>
      </div>
      <div v-else class="empty-state text-center py-12 px-8">
        <p class="text-text-muted mb-6">{{ t('coinDetail.notes.empty') }}</p>
        <router-link :to="`/edit/${coin.id}`" class="btn btn-secondary btn-sm">
          {{ t('coinDetail.notes.add') }}
        </router-link>
      </div>
    </template>
  </CoinDetailSectionPageShell>
</template>

<script setup lang="ts">
import CoinDetailSectionPageShell from '@/components/coin/CoinDetailSectionPageShell.vue'
import { t } from '@/i18n'
</script>
