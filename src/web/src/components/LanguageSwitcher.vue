<template>
  <div
    class="inline-flex shrink-0 items-center rounded-full border border-border-subtle bg-surface p-[0.15rem]"
    role="group"
    :aria-label="t('language.label')"
  >
    <button
      v-for="option in SUPPORTED_LOCALES"
      :key="option.code"
      type="button"
      class="cursor-pointer rounded-full border-0 px-2 py-[0.2rem] text-label font-semibold tracking-[0.04em] transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-gold"
      :class="locale === option.code ? 'bg-gold-glow text-gold' : 'bg-transparent text-text-secondary hover:text-gold'"
      :aria-pressed="locale === option.code"
      :lang="option.code"
      :title="option.label"
      @click="setLocale(option.code)"
    >
      {{ option.short }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { SUPPORTED_LOCALES, locale, setLocale, t } from '@/i18n'
</script>
