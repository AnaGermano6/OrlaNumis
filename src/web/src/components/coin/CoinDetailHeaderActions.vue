<template>
  <div class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 max-md:mb-4">
    <button class="btn btn-ghost btn-xs justify-self-start whitespace-nowrap" @click="router.push(backTarget)">
      <ArrowLeft :size="14" />
      {{ t('coinDetail.backToGallery') }}
    </button>
    <div class="flex min-w-0 items-center justify-end gap-[0.45rem]">
      <AppIconButton
        v-if="showPinButton"
        :disabled="pinBusy"
        :title="pinBusy ? t('coinDetail.pin.updating') : pinLabel"
        :aria-label="pinBusy ? t('coinDetail.pin.updating') : pinLabel"
        :aria-busy="pinBusy"
        :active="coinPinned"
        :aria-pressed="coinPinned"
        @click="togglePin"
      >
        <PinOff v-if="coinPinned" :size="24" />
        <Pin v-else :size="24" />
      </AppIconButton>
      <AppIconButton
        v-if="showReminderButton"
        :title="reminderActive ? t('coinDetail.reminder.edit') : t('coinDetail.reminder.set')"
        :aria-label="reminderActive ? t('coinDetail.reminder.edit') : t('coinDetail.reminder.set')"
        :active="reminderActive"
        @click="emit('reminder')"
      >
        <BellRing :size="24" />
      </AppIconButton>
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
        :coin-id="coinId"
        :duplicating="duplicating"
        :show-share="collapseSecondaryActions"
        :sharing="sharing"
        :show-reminder="collapseSecondaryActions && showReminderAction"
        :reminder-active="reminderActive"
        :show-pin="collapseSecondaryActions && !isSold"
        :pinned="coinPinned"
        :pin-busy="pinBusy"
        @duplicate="emit('duplicate')"
        @share="emit('share')"
        @reminder="emit('reminder')"
        @pin="togglePin"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { computed } from 'vue'
import { ArrowLeft, BellRing, Pencil, Pin, PinOff, Share2, Trash2 } from 'lucide-vue-next'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import CoinDetailOverflowMenu from '@/components/coin/CoinDetailOverflowMenu.vue'
import { usePwa } from '@/composables/usePwa'
import { useQuickAccess } from '@/composables/useQuickAccess'
import { useToast } from '@/composables/useToast'
import { t } from '@/i18n'

const props = withDefaults(defineProps<{
  isWishlist: boolean
  isSold: boolean
  coinId: number
  sharing?: boolean
  duplicating?: boolean
  showReminderAction?: boolean
  reminderActive?: boolean
}>(), {
  sharing: false,
  duplicating: false,
  showReminderAction: false,
  reminderActive: false,
})

const emit = defineEmits<{
  share: []
  duplicate: []
  edit: []
  delete: []
  reminder: []
}>()

const router = useRouter()
const backTarget = computed(() => props.isWishlist ? '/wishlist' : '/')
const { error, pin, unpin, isPinned, isBusy } = useQuickAccess()
const { showToast } = useToast()
const pinned = isPinned('coin', props.coinId)
const busy = isBusy('coin', props.coinId)
const coinPinned = computed(() => pinned.value)
const pinBusy = computed(() => busy.value)
const pinLabel = computed(() => coinPinned.value ? t('coinDetail.pin.unpin') : t('coinDetail.pin.pin'))
const { isPwa } = usePwa()
// In the installed PWA the action row keeps only Edit/Delete inline; share, reminder and pin
// move into the overflow menu. In the browser every action stays on the row.
const collapseSecondaryActions = computed(() => isPwa)
const showPinButton = computed(() => !collapseSecondaryActions.value && !props.isSold)
const showReminderButton = computed(() => !collapseSecondaryActions.value && props.showReminderAction)

async function togglePin() {
  if (pinBusy.value) return
  try {
    if (coinPinned.value) await unpin('coin', props.coinId)
    else await pin('coin', props.coinId)
  } catch {
    showToast(error.value || t('coinDetail.pin.failed'), 'error')
  }
}
</script>
