<template>
  <Teleport v-if="!isMobile" defer to="#desktop-page-title">
    <span class="block truncate" :title="title">{{ title }}</span>
  </Teleport>
  <Teleport v-if="!isMobile && $slots.actions" defer to="#desktop-page-actions">
    <slot name="actions" />
  </Teleport>
  <div
    v-if="$slots.default"
    class="desktop-page-context-local"
    :class="{ 'is-mobile': isMobile }"
  >
    <slot />
  </div>
  <div
    v-if="$slots.desktop"
    class="desktop-page-context-desktop"
    :class="{ 'is-mobile': isMobile }"
  >
    <slot name="desktop" />
  </div>
</template>

<script setup lang="ts">
// A page's title and actions: in the app's top bar on a wider screen, in the
// page itself on a phone (the default slot), also a phone turned sideways.
import { useLayout } from '@/composables/useLayout'

defineProps<{
  title: string
}>()

const { isMobile } = useLayout()
</script>

<style scoped>
.desktop-page-context-desktop {
  display: none;
}

@media (min-width: 768px) {
  .desktop-page-context-local:not(.is-mobile) {
    display: none;
  }

  .desktop-page-context-desktop:not(.is-mobile) {
    display: block;
  }
}
</style>
