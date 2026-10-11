<template>
  <footer
    class="mt-12 border-t border-border-subtle bg-surface"
    :class="isPwa ? 'pb-[calc(5.5rem+env(safe-area-inset-bottom))]' : 'pb-6'"
    :aria-label="t('app.footer.label')"
  >
    <div class="mx-auto flex max-w-[1200px] flex-col gap-2 px-4 pt-6 text-chip text-text-muted md:flex-row md:items-center md:justify-between md:gap-6">
      <div class="flex items-center gap-2">
        <img src="/coin-logo.jpg" alt="" class="h-6 w-6 rounded-full border border-gold-dim object-cover" />
        <span class="font-display font-semibold text-gold">Aurearia</span>
        <span aria-hidden="true">·</span>
        <span data-testid="app-version">{{ t('app.footer.version', { version }) }}</span>
      </div>
      <p class="m-0">{{ t('app.footer.catalogue') }} · {{ t('app.footer.photos') }}</p>
      <p class="m-0">© {{ year }} Aurearia</p>
    </div>
  </footer>
</template>

<script setup lang="ts">
// The footer at the end of every page: the app's name and version, where its
// catalogue data and official photos come from, and the year. In the
// installed app it leaves room for the floating assistant button.
import { usePwa } from '@/composables/usePwa'
import { t } from '@/i18n'

const { isPwa } = usePwa()
const version = import.meta.env.VITE_APP_VERSION || 'dev'
const year = new Date().getFullYear()
</script>
