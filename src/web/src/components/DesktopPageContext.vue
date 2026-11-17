<template>
  <!-- On a wider screen: the page's title and its actions at its top (the
       bar above holds the menu); then what only a wider screen shows. -->
  <header v-if="!isMobile" class="page-header" data-testid="page-context">
    <h1 class="min-w-0 truncate" :title="title">{{ title }}</h1>
    <div v-if="$slots.actions" class="flex shrink-0 items-center gap-1">
      <slot name="actions" />
    </div>
  </header>
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
// A page's title and actions: at the top of the page on a wider screen (the
// title, the actions slot, then the desktop slot), and the page's own header
// on a phone (the default slot), also a phone turned sideways.
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
