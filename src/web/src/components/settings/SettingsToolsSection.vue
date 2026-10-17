<template>
  <section class="card">
    <h2 class="text-md font-medium mb-5 pb-3 border-b border-border-subtle">{{ t('settings.tools.title') }}</h2>
    <p class="text-sm text-text-muted mb-4">
      {{ t('settings.tools.intro') }}
    </p>
    <ImageProcessor @saved="(coinId: number) => $emit('saved', coinId)" />
  </section>
</template>

<script setup lang="ts">
import ImageProcessor from '@/components/ImageProcessor.vue'
import { t } from '@/i18n'

defineEmits<{
  saved: [coinId: number]
}>()
</script>
