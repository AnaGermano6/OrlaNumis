<template>
  <li class="flex min-w-0 flex-col overflow-hidden rounded-md border border-border-subtle bg-card" data-testid="release">
    <!-- Without a picture, a lower space keeps the list easy to scan. -->
    <div class="relative">
      <a
        :href="release.pageUrl"
        target="_blank"
        rel="noopener noreferrer"
        class="relative block bg-input"
        :class="release.imageUrl ? 'aspect-square' : 'aspect-[2/1]'"
        :title="t('collectionOverview.releases.open')"
      >
        <img v-if="release.imageUrl" :src="release.imageUrl" :alt="release.title" loading="lazy" class="h-full w-full object-contain p-2" />
        <span v-else class="flex h-full w-full items-center justify-center text-text-primary/30" data-testid="release-no-image">
          <Coins :size="40" :stroke-width="1" />
        </span>
        <span class="absolute left-2 top-2 flex flex-col items-start gap-1">
          <span v-if="release.upcoming" class="rounded-full bg-gold px-2 py-[0.1rem] text-chip font-semibold text-surface" data-testid="release-upcoming">{{ t('collectionOverview.releases.upcoming') }}</span>
          <span v-if="release.owned" class="inline-flex items-center gap-1 rounded-full bg-gain px-2 py-[0.1rem] text-chip font-semibold text-surface" data-testid="release-owned">
            <Check :size="12" /> {{ t('collectionOverview.releases.owned') }}
          </span>
          <span v-else-if="wishlisted" class="inline-flex items-center gap-1 rounded-full bg-gold px-2 py-[0.1rem] text-chip font-semibold text-surface" data-testid="release-on-wishlist">
            <Heart :size="12" class="fill-current" /> {{ t('collectionOverview.releases.onWishlist') }}
          </span>
        </span>
      </a>
      <!-- In the catalogue: hide the coin, or say it is the same as another. -->
      <button
        v-if="manageable"
        type="button"
        class="absolute right-1.5 top-1.5 inline-flex h-8 w-8 items-center justify-center rounded-full border border-border-subtle bg-card text-text-secondary transition-colors hover:border-gold hover:text-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        :aria-label="t('sets.catalogue.manage.open', { title: release.title })"
        :title="t('sets.catalogue.manage.open', { title: release.title })"
        data-testid="release-manage"
        @click="emit('manage')"
      >
        <MoreHorizontal :size="16" />
      </button>
    </div>
    <div class="flex min-w-0 flex-1 flex-col gap-1 p-3">
      <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">
        <template v-if="showCountry">{{ releaseCountryName(release.country) }} · </template>{{ releaseFaceValues(release) }}
      </span>
      <span class="line-clamp-3 text-body font-medium leading-snug text-text-primary" :title="release.title">{{ release.title }}</span>
      <span class="text-chip text-gold">{{ releasePeriod(release) }}</span>
      <span v-if="release.variants?.length" class="text-chip text-text-muted">{{ release.variants.join(' · ') }}</span>
      <span v-if="release.volume" class="text-chip text-text-muted">{{ t('collectionOverview.releases.volume', { volume: formatCount(release.volume) }) }}</span>
      <!-- Add Coin filled in with this coin (not one still to come, nor one the
           collection has); and the wishlist at once, a heart beside it or a
           button of its own. -->
      <span v-if="canAdd || canWish" class="mt-auto flex gap-2 pt-2">
        <RouterLink
          v-if="canAdd"
          :to="releaseAddRoute(release)"
          class="btn btn-secondary btn-sm min-w-0 flex-1 justify-center whitespace-nowrap"
          :aria-label="t('collectionOverview.releases.add')"
          :title="t('collectionOverview.releases.add')"
          data-testid="release-add"
        >
          <!-- Beside the heart, a phone's card has room for a word. -->
          <Plus :size="14" /><span class="sm:hidden">{{ t('collectionOverview.releases.addShort') }}</span><span class="hidden sm:inline">{{ t('collectionOverview.releases.add') }}</span>
        </RouterLink>
        <button
          v-if="canWish"
          type="button"
          class="btn btn-secondary btn-sm justify-center"
          :class="canAdd ? 'shrink-0 px-2' : 'w-full'"
          :disabled="saving || (wishlisted && !wishlistCoinId)"
          :aria-pressed="wishlisted"
          :aria-label="wishlistLabel"
          :title="wishlistLabel"
          data-testid="release-wishlist"
          @click="toggleWishlist"
        >
          <Heart :size="14" :class="wishlisted ? 'fill-current text-gold' : ''" />
          <span v-if="!canAdd" class="whitespace-nowrap">{{ wishlisted ? t('collectionOverview.releases.onWishlist') : t('collectionOverview.releases.wishlistShort') }}</span>
        </button>
      </span>
    </div>
  </li>
</template>

<script setup lang="ts">
// A coin the ECB lists or a mint sells: its picture (opening the source's
// page), country, face values, title, date, variants and mintage, marked when
// it is to come, the collection has it or the wishlist; where asked, buttons
// to add it to the collection, and to put it on the wishlist or take it off.
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { Check, Coins, Heart, MoreHorizontal, Plus } from 'lucide-vue-next'
import type { CoinRelease } from '@/api/client'
import { t } from '@/i18n'
import { formatCount, releaseCountryName, releaseFaceValues, releasePeriod } from '@/utils/coinReleases'
import { releaseAddRoute } from '@/utils/releasePrefill'
import { useReleaseWishlist } from '@/composables/useReleaseWishlist'
import { useToast } from '@/composables/useToast'

const props = withDefaults(defineProps<{
  release: CoinRelease
  /** Off where the country is already said, as in a country's catalogue. */
  showCountry?: boolean
  /** Offers to add the coin to the collection (not one still to come, nor one it has). */
  addable?: boolean
  /** Offers to save the coin to the wishlist at once (not one the collection has), or to take it off. */
  wishlistable?: boolean
  /** Offers the catalogue's options for the coin: hide it, join it to another. */
  manageable?: boolean
}>(), { showCountry: true, addable: false, wishlistable: false, manageable: false })

const emit = defineEmits<{ manage: [] }>()

const { saveToWishlist, removeFromWishlist } = useReleaseWishlist()
const { showToast } = useToast()
const saving = ref(false)
// On the wishlist, and its coin (to take it off with a second press).
const wishlisted = ref(props.release.wishlist === true)
const wishlistCoinId = ref<number | null>(props.release.wishlistCoinId ?? null)
// A list loaded again says what the server knows.
watch(() => [props.release.wishlist, props.release.wishlistCoinId] as const, ([onWishlist, coinId]) => {
  wishlisted.value = onWishlist === true
  wishlistCoinId.value = coinId ?? null
})

const canAdd = computed(() => props.addable && !props.release.upcoming && !props.release.owned)
// A coin the collection has may still be on the wishlist: the heart takes it off.
const canWish = computed(() => props.wishlistable && (!props.release.owned || wishlisted.value))
const wishlistLabel = computed(() => !wishlisted.value
  ? t('collectionOverview.releases.saveWishlist')
  : wishlistCoinId.value ? t('collectionOverview.releases.removeWishlist') : t('collectionOverview.releases.onWishlist'))

async function toggleWishlist() {
  if (saving.value) return
  if (wishlisted.value) await remove()
  else await save()
}

async function save() {
  saving.value = true
  try {
    const { result, coinId } = await saveToWishlist(props.release)
    wishlisted.value = true
    wishlistCoinId.value = coinId ?? null
    showToast(result === 'saved'
      ? t('collectionOverview.releases.savedWishlist', { title: props.release.title })
      : t('collectionOverview.releases.alreadyWishlist'), result === 'saved' ? 'success' : 'info')
  } catch {
    showToast(t('collectionOverview.releases.wishlistFailed'), 'error')
  } finally {
    saving.value = false
  }
}

async function remove() {
  if (!wishlistCoinId.value) return
  saving.value = true
  try {
    await removeFromWishlist(wishlistCoinId.value)
    wishlisted.value = false
    wishlistCoinId.value = null
    showToast(t('collectionOverview.releases.removedWishlist', { title: props.release.title }), 'success')
  } catch {
    showToast(t('collectionOverview.releases.removeWishlistFailed'), 'error')
  } finally {
    saving.value = false
  }
}
</script>
