<template>
  <div class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 max-md:mb-4">
    <button class="btn btn-ghost btn-xs justify-self-start whitespace-nowrap" @click="router.push(backTarget)">
      <ArrowLeft :size="14" />
      {{ backLabel }}
    </button>
    <div class="flex min-w-0 items-center justify-end gap-[0.45rem]">
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
