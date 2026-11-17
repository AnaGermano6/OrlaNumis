<template>
  <section class="card" data-testid="settings-about">
    <h2 class="mb-5 border-b border-border-subtle pb-3 text-lg font-medium">{{ t('settings.tabs.about') }}</h2>

    <div class="mb-5 flex items-center gap-3">
      <img src="/coin-logo.jpg" alt="" class="h-10 w-10 rounded-full border border-gold-dim object-cover" />
      <div class="flex flex-col">
        <span class="font-display text-md font-semibold text-gold">Orla Numis</span>
        <span class="text-sm text-text-muted" data-testid="about-version">{{ t('settings.about.version', { version }) }}</span>
      </div>
    </div>

    <ul class="m-0 flex list-none flex-col gap-3 p-0 text-body text-text-secondary">
      <li>
        <I18nT keypath="settings.about.catalogue">
          <template #source><SafeExternalLink :href="NUMISTA" class="text-gold hover:underline">Numista</SafeExternalLink></template>
        </I18nT>
      </li>
      <li>{{ t('settings.about.photos') }}</li>
      <li data-testid="about-credits">
        <I18nT keypath="settings.about.basedOn">
          <template #project><SafeExternalLink :href="ORIGINAL_PROJECT" class="text-gold hover:underline">Aurearia</SafeExternalLink></template>
        </I18nT>
      </li>
      <li class="text-text-muted">© {{ year }} Orla Numis</li>
    </ul>
  </section>
</template>

<script setup lang="ts">
// Settings › About: the app, its version, where its catalogue data and
// official photos come from, and the project it is based on.
import I18nT from '@/components/I18nT.vue'
import SafeExternalLink from '@/components/SafeExternalLink.vue'
import { t } from '@/i18n'

/** Orla Numis is based on Aurearia, by Brian DeNicola, under the MIT license. */
const ORIGINAL_PROJECT = 'https://github.com/briandenicola/Aurearia'
const NUMISTA = 'https://en.numista.com/'

const version = import.meta.env.VITE_APP_VERSION || 'dev'
const year = new Date().getFullYear()
</script>
