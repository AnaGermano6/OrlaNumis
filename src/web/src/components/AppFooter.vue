<template>
  <footer
    class="mt-12 border-t border-border-subtle bg-surface"
    :class="isPwa ? 'pb-[calc(5.5rem+env(safe-area-inset-bottom))]' : 'pb-6'"
    :aria-label="t('app.footer.label')"
  >
    <div class="mx-auto flex max-w-[1200px] flex-col gap-2 px-4 pt-6 text-chip text-text-muted md:flex-row md:items-center md:justify-between md:gap-6">
      <div class="flex shrink-0 items-center gap-2 whitespace-nowrap">
        <img src="/coin-logo.jpg" alt="" class="h-6 w-6 rounded-full border border-gold-dim object-cover" />
        <span class="font-display font-semibold text-gold">Orla Numis</span>
      </div>
      <p class="m-0">{{ t('app.footer.catalogue') }} · {{ t('app.footer.photos') }}</p>
      <p class="m-0" data-testid="app-credits">
        © {{ year }} Orla Numis ·
        <I18nT keypath="app.footer.basedOn">
          <template #project><SafeExternalLink :href="ORIGINAL_PROJECT" class="text-gold hover:underline">Aurearia</SafeExternalLink></template>
        </I18nT>
      </p>
    </div>
  </footer>
</template>

<script setup lang="ts">
// The footer at the end of every page: the app's name, where its
// catalogue data and official photos come from, the year, and the project it
// is based on. In the installed app it leaves room for the floating
// assistant button.
import I18nT from '@/components/I18nT.vue'
import SafeExternalLink from '@/components/SafeExternalLink.vue'
import { usePwa } from '@/composables/usePwa'
import { t } from '@/i18n'

/** Orla Numis is based on Aurearia, by Brian DeNicola, under the MIT license. */
const ORIGINAL_PROJECT = 'https://github.com/briandenicola/Aurearia'

const { isPwa } = usePwa()
const year = new Date().getFullYear()
</script>
