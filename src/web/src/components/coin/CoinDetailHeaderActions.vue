<template>
  <!-- On a phone the previous and next coin get a row of their own, below. -->
  <div class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-3 gap-y-2 max-md:mb-4 md:grid-cols-[auto_minmax(0,1fr)_auto]">
    <button class="btn btn-ghost btn-xs col-start-1 row-start-1 justify-self-start whitespace-nowrap" @click="router.push(backTarget)">
      <ArrowLeft :size="14" />
      {{ backLabel }}
    </button>
    <CoinBrowseNav
      :coin-id="coinId"
      :is-banknote="isBanknote"
      class="col-span-2 col-start-1 row-start-2 justify-self-center md:col-span-1 md:col-start-2 md:row-start-1 md:justify-self-start"
    />
    <div class="col-start-2 row-start-1 flex min-w-0 items-center justify-end gap-[0.45rem] md:col-start-3">
      <AppIconButton
        v-if="!collapseSecondaryActions"
        :disabled="sharing"
        :title="sharing ? t('common.sharing') : t('common.share')"
        :aria-label="sharing ? t('common.sharing') : t('common.share')"
        @click="$emit('share')"
      >
        <Share2 :size="24" />
      </AppIconButton>
      <AppIconButton :title="t('common.edit')" :aria-label="t('common.edit')" @click="$emit('edit')">
        <Pencil :size="24" />
      </AppIconButton>
      <AppIconButton :title="t('common.delete')" :aria-label="t('common.delete')" @click="$emit('delete')">
        <Trash2 :size="24" />
      </AppIconButton>
      <CoinDetailOverflowMenu
        :is-banknote="isBanknote"
        :coin-id="coinId"
        :duplicating="duplicating"
        :show-share="collapseSecondaryActions"
        :sharing="sharing"
        @duplicate="emit('duplicate')"
        @share="emit('share')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { computed } from 'vue'
import { ArrowLeft, Pencil, Share2, Trash2 } from 'lucide-vue-next'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import CoinDetailOverflowMenu from '@/components/coin/CoinDetailOverflowMenu.vue'
import CoinBrowseNav from '@/components/coin/CoinBrowseNav.vue'
import { usePwa } from '@/composables/usePwa'
import { t } from '@/i18n'

const props = withDefaults(defineProps<{
  isWishlist: boolean
  /** A banknote goes back to the banknotes, not to the coin gallery. */
  isBanknote?: boolean
  coinId: number
  sharing?: boolean
  duplicating?: boolean
}>(), {
  isBanknote: false,
  sharing: false,
  duplicating: false,
})

const emit = defineEmits<{
  share: []
  duplicate: []
  edit: []
  delete: []
}>()

const router = useRouter()
const backTarget = computed(() => props.isWishlist ? '/wishlist' : props.isBanknote ? '/banknotes' : '/')
// Named as the page it goes back to.
const backLabel = computed(() => props.isWishlist ? t('coinDetail.backToWishlist') : props.isBanknote ? t('coinDetail.backToBanknotes') : t('coinDetail.backToGallery'))
const { isPwa } = usePwa()
// In the installed PWA the action row keeps only Edit/Delete inline; share moves into the
// overflow menu. In the browser every action stays on the row.
const collapseSecondaryActions = computed(() => isPwa)
</script>
