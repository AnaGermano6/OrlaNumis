<template>
  <article
    class="card group grid cursor-pointer grid-cols-[minmax(0,2fr)_minmax(0,3fr)] gap-x-5 gap-y-3 !p-5 max-sm:gap-x-3 max-sm:!p-3 sm:grid-rows-[1fr_auto]"
    data-testid="wishlist-card"
    @click="router.push(`/coin/${coin.id}`)"
  >
    <!-- Its photo in a light box (a banknote's wider), with whether it can still be bought;
         beside it, from top to bottom, all the rest. -->
    <div
      class="relative flex w-full max-w-64 items-center justify-center self-start overflow-hidden rounded-md bg-input sm:row-span-2"
      :class="coin.kind === 'banknote' ? 'aspect-[3/2] p-2' : 'aspect-square p-[6%]'"
    >
      <AuthenticatedImage v-if="primaryImage" :media-path="primaryImage" :alt="coin.name" class="h-full w-full object-contain transition duration-300 group-hover:scale-[1.02]" loading="lazy">
        <template #fallback>
          <span class="flex aspect-square w-3/5 items-center justify-center rounded-full border-2 border-gold-dim text-gold-dim">
            <component :is="coin.kind === 'banknote' ? Banknote : Coins" :size="32" :stroke-width="1.25" />
          </span>
        </template>
      </AuthenticatedImage>
      <span v-else class="flex aspect-square w-3/5 items-center justify-center rounded-full border-2 border-gold-dim text-gold-dim" data-testid="wishlist-card-placeholder">
        <component :is="coin.kind === 'banknote' ? Banknote : Coins" :size="32" :stroke-width="1.25" />
      </span>
      <template v-if="coin.listingStatus === 'unavailable'">
        <div class="pointer-events-none absolute inset-0 z-[2] bg-black/50"></div>
        <span class="absolute top-2 right-2 z-[3] rounded-full bg-red-600/85 px-2 py-[0.2rem] text-label font-semibold uppercase tracking-[0.08em] text-white">{{ t('coinCard.unavailable') }}</span>
        <button
          type="button"
          class="absolute right-2 bottom-2 z-[3] rounded-sm border border-border-subtle bg-black/70 px-2 py-[0.15rem] text-micro text-text-secondary transition-colors hover:bg-black/85 hover:text-text-primary"
          @click.stop="emit('dismiss-status', coin.id)"
        >
          {{ t('common.dismiss') }}
        </button>
      </template>
    </div>

    <!-- Its name, its country and where it is sold, and its price. -->
    <div class="flex min-w-0 flex-col gap-3 max-sm:gap-2">
      <h3 class="m-0 overflow-hidden font-sans text-xl font-medium leading-snug text-text-primary max-sm:text-md [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:3]">
        <span
          v-if="coin.listingStatus === 'available' || coin.listingStatus === 'unknown'"
          class="mr-2 inline-block h-2 w-2 align-middle rounded-full"
          :class="coin.listingStatus === 'available' ? 'bg-gain' : 'bg-warning'"
          :title="coin.listingStatus === 'available' ? t('coinCard.available') : t('coinCard.unknown')"
          data-testid="wishlist-card-status"
        ></span>
        {{ coin.name }}
      </h3>
      <p v-if="coin.country || reference" class="m-0 flex min-w-0 items-center gap-2 text-base text-text-secondary max-sm:text-body" data-testid="wishlist-card-origin">
        <CountryFlag v-if="coin.country" :country="coin.country" />
        <span class="min-w-0 truncate">
          <template v-if="coin.country">{{ coin.country }}</template>
          <template v-if="coin.country && reference"> · </template>
          <SafeExternalLink
            v-if="reference && coin.referenceUrl"
            :href="coin.referenceUrl"
            class="text-text-secondary underline-offset-4 hover:text-gold hover:underline"
            target="_blank"
            rel="noopener noreferrer"
            @click.stop
          >{{ reference }}</SafeExternalLink>
          <template v-else-if="reference">{{ reference }}</template>
        </span>
      </p>
      <span
        v-if="price"
        class="self-start rounded-full bg-gold px-3.5 py-1 text-base font-semibold text-surface max-sm:text-body"
        :title="price.title"
        data-testid="wishlist-card-price"
      >{{ price.label }}</span>
    </div>

    <!-- What to do with it: under the rest; on a phone, across the card. -->
    <div class="flex items-center gap-2 self-end max-sm:col-span-2">
      <button type="button" class="btn btn-primary max-sm:flex-1 max-sm:justify-center" data-testid="wishlist-card-purchase" @click.stop="emit('purchase', coin)">
        <Check :size="16" aria-hidden="true" /> {{ t('wishlist.card.markPurchased') }}
      </button>
      <button
        type="button"
        class="btn btn-secondary !px-3"
        :title="t('wishlist.card.remove')"
        :aria-label="t('wishlist.card.remove')"
        data-testid="wishlist-card-remove"
        @click.stop="emit('remove', coin)"
      >
        <Trash2 :size="16" aria-hidden="true" />
      </button>
    </div>
  </article>
</template>

<script setup lang="ts">
// A coin (or banknote) of the wishlist, side by side: its photo, then its
// name, its country and where it is sold, its price, and the buttons to mark
// it bought or take it off the wishlist. Clicking it opens its page.
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Banknote, Check, Coins, Trash2 } from 'lucide-vue-next'
import type { Coin } from '@/types'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import CountryFlag from '@/components/CountryFlag.vue'
import SafeExternalLink from '@/components/SafeExternalLink.vue'
import { t } from '@/i18n'
import { formatCurrency } from '@/utils/format'
import { parseDenomination } from '@/utils/ucoinNumista'

const props = defineProps<{ coin: Coin }>()

const emit = defineEmits<{
  purchase: [coin: Coin]
  remove: [coin: Coin]
  'dismiss-status': [coinId: number]
}>()

const router = useRouter()

const primaryImage = computed(() => {
  const images = props.coin.images ?? []
  return (images.find((image) => image.isPrimary) ?? images[0])?.filePath ?? null
})

// Where it is sold: the reference's name, else its site's.
const reference = computed(() => {
  const text = props.coin.referenceText?.trim()
  if (text) return text
  try {
    return props.coin.referenceUrl ? new URL(props.coin.referenceUrl).hostname.replace(/^www\./, '') : ''
  } catch {
    return ''
  }
})

// A euro coin's face value ("7½ euro" → 7.5, "10 cêntimos de euro" → 0.1),
// what the mints sell new ones for; null for another currency, or none.
function euroFaceValue(denomination: string | undefined): number | null {
  if (!/\beuros?\b/i.test(denomination ?? '')) return null
  const { value, subunit } = parseDenomination(denomination ?? '')
  if (!value) return null
  return subunit ? value / 100 : value
}

// Its price: the one saved (what it is worth, else what is asked); without
// one, about its face value, for a euro coin.
const price = computed(() => {
  const saved = props.coin.currentValue ?? props.coin.purchasePrice
  if (saved != null) return { label: formatCurrency(saved), title: undefined }
  const face = props.coin.kind === 'banknote' ? null : euroFaceValue(props.coin.denomination)
  if (face == null) return null
  return {
    label: t('wishlist.card.about', { price: formatCurrency(face, 'EUR') }),
    title: t('wishlist.card.faceValue', { denomination: props.coin.denomination.trim() }),
  }
})
</script>
