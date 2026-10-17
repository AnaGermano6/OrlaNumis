<template>
  <AppOverflowMenu :backdrop-aria-label="t('coinDetail.menu.closeBackdrop')">
    <template #trigger="{ open, toggle }">
      <AppIconButton :title="t('coinDetail.menu.more')" :aria-label="t('coinDetail.menu.open')" :active="open" @click="toggle">
        <Menu :size="24" />
      </AppIconButton>
    </template>
    <template #default="{ close }">
      <button
        v-if="showShare"
        class="inline-flex w-full items-center gap-2 whitespace-nowrap rounded-sm px-3 py-2 text-left text-body text-text-secondary transition-all hover:bg-card-hover hover:text-text-primary disabled:cursor-not-allowed disabled:opacity-55"
        :disabled="sharing"
        :aria-label="sharing ? t('common.sharing') : t('common.share')"
        @click="handleShare(close)"
      >
        <Share2 :size="16" />
        {{ sharing ? t('common.sharing') : t('common.share') }}
      </button>
      <div v-if="showShare" class="my-1 border-t border-border-subtle"></div>
      <button
        class="inline-flex w-full items-center gap-2 whitespace-nowrap rounded-sm px-3 py-2 text-left text-body text-text-secondary transition-all hover:bg-card-hover hover:text-text-primary disabled:cursor-not-allowed disabled:opacity-55"
        :disabled="duplicating"
        :aria-label="duplicating ? (isBanknote ? t('coinDetail.menu.copyingBanknote') : t('coinDetail.menu.copyingCoin')) : copyLabel"
        @click="handleDuplicate(close)"
      >
        <Copy :size="16" />
        {{ duplicating ? t('coinDetail.menu.copying') : copyLabel }}
      </button>
      <router-link
        v-for="section in sections"
        :key="section.id"
        :to="section.route(coinId)"
        class="inline-flex w-full items-center rounded-sm px-3 py-2 text-body text-text-secondary no-underline transition-all hover:bg-card-hover hover:text-text-primary"
        @click="close"
      >
        {{ section.title }}
      </router-link>
    </template>
  </AppOverflowMenu>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Copy, Menu, Share2 } from 'lucide-vue-next'
import { COIN_DETAIL_SECTIONS, SECTION_ORDER, type CoinDetailSection } from '@/constants/coinDetailSections'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import AppOverflowMenu from '@/components/ui/AppOverflowMenu.vue'
import { t } from '@/i18n'

const props = withDefaults(defineProps<{
  coinId: number
  /** A banknote is copied as a banknote: "Copy banknote". */
  isBanknote?: boolean
  duplicating?: boolean
  showShare?: boolean
  sharing?: boolean
}>(), {
  isBanknote: false,
  duplicating: false,
  showShare: false,
  sharing: false,
})

const copyLabel = computed(() => (props.isBanknote ? t('coinDetail.menu.copyBanknote') : t('coinDetail.menu.copy')))

const emit = defineEmits<{
  duplicate: []
  share: []
}>()

const sections: CoinDetailSection[] = SECTION_ORDER.map(id => COIN_DETAIL_SECTIONS[id]) as CoinDetailSection[]

function handleDuplicate(close: () => void) {
  if (props.duplicating) return
  close()
  emit('duplicate')
}

function handleShare(close: () => void) {
  if (props.sharing) return
  close()
  emit('share')
}
</script>
