<template>
  <div class="container">
    <DesktopPageContext :title="t('wishlist.trades.title')">
      <!-- Copying the list as text, or sharing it as a file. -->
      <template #actions>
        <span class="flex flex-wrap items-center gap-2">
          <button type="button" class="btn btn-secondary" :disabled="!list?.canCopy" data-testid="trade-copy" @click="list?.copy()">
            <Copy :size="16" aria-hidden="true" /> {{ t('collectionOverview.repeated.copy') }}
          </button>
          <button type="button" class="btn btn-primary" data-testid="trade-share" @click="showShare = true">
            <Share2 :size="16" aria-hidden="true" /> {{ t('collectionOverview.repeated.share') }}
          </button>
        </span>
      </template>
      <header class="page-header">
        <h1>{{ t('wishlist.trades.title') }}</h1>
        <div class="pwa-actions">
          <button type="button" class="pwa-icon-btn" :disabled="!list?.canCopy" :title="t('collectionOverview.repeated.copy')" :aria-label="t('collectionOverview.repeated.copy')" @click="list?.copy()">
            <Copy :size="22" />
          </button>
          <button type="button" class="pwa-icon-btn" :title="t('collectionOverview.repeated.share')" :aria-label="t('collectionOverview.repeated.share')" @click="showShare = true">
            <Share2 :size="22" />
          </button>
        </div>
      </header>
    </DesktopPageContext>

    <CollectionRepeated ref="list" :header="false" />

    <ShareListDialog :open="showShare" initial-list="trades" @close="showShare = false" />
  </div>
</template>

<script setup lang="ts">
// The trade list on a page of its own (in the Collection menu): the coins
// and banknotes held more than once, to copy as text or share as a file.
import { ref } from 'vue'
import { Copy, Share2 } from 'lucide-vue-next'
import CollectionRepeated from '@/components/collection/CollectionRepeated.vue'
import ShareListDialog from '@/components/collection/ShareListDialog.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import { t } from '@/i18n'

const list = ref<InstanceType<typeof CollectionRepeated> | null>(null)
const showShare = ref(false)
</script>
