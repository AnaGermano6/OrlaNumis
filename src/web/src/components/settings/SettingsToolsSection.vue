<template>
  <section class="card">
    <h2 class="text-md font-medium mb-5 pb-3 border-b border-border-subtle">{{ t('settings.tools.title') }}</h2>
    <p class="text-sm text-text-muted mb-4">
      {{ t('settings.tools.intro') }}
    </p>
    <ImageProcessor @saved="(coinId: number) => $emit('saved', coinId)" />

    <h3 class="text-base mt-5 mb-3 text-text-secondary">{{ t('settings.tools.blocked') }}</h3>
    <p class="text-sm text-text-muted mb-3">
      {{ t('settings.tools.blockedHint') }}
    </p>
    <div v-if="blockedUsers.length" class="flex flex-col gap-2 mt-4">
      <div
        v-for="user in blockedUsers"
        :key="user.id"
        class="flex justify-between items-center py-[0.6rem] border-b border-border-subtle last:border-0 gap-3"
      >
        <div class="flex items-center gap-2">
          <AuthenticatedImage
            :media-path="user.avatarPath ? user.avatarPath : '/coin-logo.jpg'"
            :alt="user.username"
            class="w-7 h-7 rounded-full object-cover border border-border-subtle shrink-0"
          />
          <span class="text-base font-medium">{{ user.username }}</span>
        </div>
        <button
          class="btn btn-secondary btn-sm"
          :disabled="blockedLoading"
          @click="$emit('unblock', user)"
        >
          {{ t('settings.tools.unblock') }}
        </button>
      </div>
    </div>
    <p v-else class="text-sm text-text-muted mt-2">{{ t('settings.tools.noBlocked') }}</p>
  </section>
</template>

<script setup lang="ts">
import ImageProcessor from '@/components/ImageProcessor.vue'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { t } from '@/i18n'

defineProps<{
  blockedUsers: { id: number; username: string; avatarPath: string }[]
  blockedLoading: boolean
}>()

defineEmits<{
  saved: [coinId: number]
  unblock: [user: { id: number; username: string; avatarPath: string }]
}>()
</script>
