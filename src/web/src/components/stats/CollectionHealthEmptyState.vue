<template>
  <div class="card !px-8 !py-12 flex flex-col items-center gap-4 text-center max-md:!px-6 max-md:!py-8">
    <div class="mb-2 text-text-muted opacity-60">
      <FileQuestion :size="48" />
    </div>
    <h3 class="m-0 text-lg text-heading">{{ t('stats.health.noActiveTitle') }}</h3>
    <p class="m-0 max-w-[500px] text-base leading-[1.6] text-text-secondary">
      {{ t('stats.health.noActiveBody') }}
    </p>
    <p class="m-0 max-w-[500px] text-body leading-[1.6] text-text-muted">
      {{ t('stats.health.noActiveHint') }}
    </p>
    <router-link to="/add" class="btn btn-primary mt-2">
      <CirclePlus :size="16" /> {{ t('collection.empty.addFirst') }}
    </router-link>
  </div>
</template>

<script setup lang="ts">
import { FileQuestion, CirclePlus } from 'lucide-vue-next'
import { t } from '@/i18n'
</script>
