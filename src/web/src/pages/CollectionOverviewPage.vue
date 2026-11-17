<template>
  <div class="container">
    <DesktopPageContext :title="t('collectionOverview.releases.title')">
      <header class="page-header">
        <h1>{{ t('collectionOverview.releases.title') }}</h1>
      </header>
    </DesktopPageContext>

    <UpdateNotice />
    <NumistaValuesNotice />

    <CoinReleases />
  </div>
</template>

<script setup lang="ts">
// New releases: notices (a new version, the Numista quota), then new and
// upcoming coins. The collection's numbers are at the top of the statistics.
import CoinReleases from '@/components/collection/CoinReleases.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import NumistaValuesNotice from '@/components/collection/NumistaValuesNotice.vue'
import UpdateNotice from '@/components/collection/UpdateNotice.vue'
import { t } from '@/i18n'
</script>
