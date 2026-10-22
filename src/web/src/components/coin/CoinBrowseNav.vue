<template>
  <nav
    v-if="neighbours"
    class="flex items-center gap-1"
    :aria-label="t('coinDetail.browse.label')"
    data-testid="coin-browse-nav"
  >
    <AppIconButton
      :disabled="!neighbours.previous"
      :title="previousLabel"
      :aria-label="previousLabel"
      data-testid="coin-browse-previous"
      @click="go(neighbours.previous)"
    >
      <ChevronLeft :size="22" />
    </AppIconButton>
    <span class="whitespace-nowrap text-chip tabular-nums text-text-secondary" data-testid="coin-browse-position">
      {{ t('coinDetail.browse.position', { position: neighbours.position, total: neighbours.total }) }}
    </span>
    <AppIconButton
      :disabled="!neighbours.next"
      :title="nextLabel"
      :aria-label="nextLabel"
      data-testid="coin-browse-next"
      @click="go(neighbours.next)"
    >
      <ChevronRight :size="22" />
    </AppIconButton>
  </nav>
</template>

<script setup lang="ts">
// The previous and the next coin (or banknote) of the list the page was
// opened from, in that list's order, and the coin's place in it; the left and
// right arrow keys do the same. Each step replaces the page, so the browser's
// back goes to the list.
import { computed, onBeforeUnmount, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { browseNeighbours } from '@/utils/browseSequence'
import { t } from '@/i18n'

const props = withDefaults(defineProps<{
  coinId: number
  isBanknote?: boolean
}>(), { isBanknote: false })

const router = useRouter()
const neighbours = computed(() => browseNeighbours(props.coinId))

const previousLabel = computed(() => t(props.isBanknote ? 'coinDetail.browse.previousBanknote' : 'coinDetail.browse.previousCoin'))
const nextLabel = computed(() => t(props.isBanknote ? 'coinDetail.browse.nextBanknote' : 'coinDetail.browse.nextCoin'))

function go(id: number | null | undefined) {
  if (id) void router.replace(`/coin/${id}`)
}

// The arrow keys, except while typing, with a modifier (Alt+← is the
// browser's back) or with a dialog open (the photo viewer uses them).
function onKeydown(event: KeyboardEvent) {
  if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return
  if (event.defaultPrevented || event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return
  const target = event.target as HTMLElement | null
  if (typeof target?.closest === 'function' && target.closest('input, textarea, select, [contenteditable="true"], [role="dialog"], [role="listbox"], [role="combobox"]')) return
  if (document.querySelector('[aria-modal="true"], [role="dialog"], [role="alertdialog"]')) return
  const id = event.key === 'ArrowLeft' ? neighbours.value?.previous : neighbours.value?.next
  if (!id) return
  event.preventDefault()
  go(id)
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>
