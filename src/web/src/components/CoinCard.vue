<template>
  <div
    class="card group flex cursor-pointer flex-col gap-3 overflow-hidden !p-3 max-md:gap-2 max-md:!p-2"
    :class="selectable && selected ? 'outline-2 outline-gold outline-offset-[-2px]' : ''"
    @click="handleClick"
  >
    <!-- The photo in a light box (a banknote's wider), its grade in the corner. -->
    <div
      class="relative flex w-full items-center justify-center overflow-hidden rounded-md bg-input"
      :class="coin.kind === 'banknote' ? 'aspect-[3/2] p-2' : 'aspect-square p-[6%]'"
    >
      <AuthenticatedImage v-if="primaryImage" :media-path="primaryImage" :alt="coin.name" class="h-full w-full object-contain transition duration-300 group-hover:scale-[1.02]" loading="lazy">
        <template #fallback>
          <span class="flex aspect-square w-3/5 items-center justify-center rounded-full border-2 border-gold-dim text-gold-dim" data-testid="coin-card-placeholder">
            <component :is="coin.kind === 'banknote' ? Banknote : Coins" :size="32" :stroke-width="1.25" />
          </span>
        </template>
      </AuthenticatedImage>
      <span v-else class="flex aspect-square w-3/5 items-center justify-center rounded-full border-2 border-gold-dim text-gold-dim" data-testid="coin-card-placeholder">
        <component :is="coin.kind === 'banknote' ? Banknote : Coins" :size="32" :stroke-width="1.25" />
      </span>
      <span
        v-if="gradeBadge && !wishlist"
        class="absolute top-2 right-2 z-[2] flex h-9 min-w-9 items-center justify-center rounded-full border border-border-subtle bg-card px-1.5 text-chip font-semibold text-text-secondary max-md:h-7 max-md:min-w-7 max-md:text-micro"
        :title="gradeTitle"
        data-testid="coin-card-grade"
      >{{ gradeBadge }}</span>
      <div v-if="wishlist && coin.listingStatus === 'unavailable'" class="pointer-events-none absolute inset-0 z-[2] bg-black/50"></div>
      <span v-if="wishlist && coin.listingStatus === 'unavailable'" class="absolute top-2 right-2 z-[3] rounded-full bg-red-600/85 px-2 py-[0.2rem] text-label font-semibold uppercase tracking-[0.08em] text-white">{{ t('coinCard.unavailable') }}</span>
      <button
        v-if="wishlist && coin.listingStatus === 'unavailable'"
        class="absolute right-2 bottom-2 z-[3] rounded-sm border border-border-subtle bg-black/70 px-2 py-[0.15rem] text-micro text-text-secondary transition-colors hover:bg-black/85 hover:text-text-primary"
        @click.stop="emit('dismiss-status', coin.id)"
      >
        {{ t('common.dismiss') }}
      </button>
      <div
        v-if="selectable"
        class="absolute top-2 left-2 z-[4] flex h-8 w-8 items-center justify-center rounded-full border-2 border-white/70 bg-black/40 transition-all max-md:h-9 max-md:w-9"
        :class="selected ? 'border-gold bg-gold text-black' : ''"
        @click.stop="emit('toggle-select', coin.id)"
      >
        <Check v-if="selected" :size="16" :stroke-width="3" />
      </div>
    </div>

    <!-- Its name, then its year, material and kind. -->
    <div class="flex flex-1 flex-col gap-1 px-1">
      <h3 class="m-0 overflow-hidden font-display text-md uppercase leading-[1.3] tracking-[0.02em] text-text-primary max-md:text-body max-md:[-webkit-line-clamp:3] [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:2]">
        <span
          v-if="wishlist && coin.listingStatus === 'available'"
          class="mr-1.5 inline-block h-2 w-2 shrink-0 align-middle rounded-full bg-gain"
          :title="t('coinCard.available')"
        ></span>
        <span
          v-if="wishlist && coin.listingStatus === 'unknown'"
          class="mr-1.5 inline-block h-2 w-2 shrink-0 align-middle rounded-full bg-warning"
          :title="t('coinCard.unknown')"
        ></span>
        {{ coin.name }}
      </h3>
      <p v-if="details.length || badge" class="m-0 truncate text-chip text-text-secondary max-md:text-micro" data-testid="coin-card-details">
        <template v-for="detail in details" :key="detail">{{ detail }} · </template>
        <span v-if="badge" :title="coin.monetarySystem?.trim() || undefined" data-testid="coin-card-badge">{{ badge }}</span>
      </p>
      <SafeExternalLink
        v-if="wishlist && coin.referenceUrl"
        :href="coin.referenceUrl"
        class="block truncate text-chip text-gold hover:underline"
        target="_blank"
        rel="noopener noreferrer"
        @click.stop
      >
        {{ coin.referenceText || coin.referenceUrl }}
      </SafeExternalLink>
    </div>

    <!-- How many, and what it is worth. -->
    <div class="flex items-center justify-between gap-2 border-t border-border-subtle px-1 pt-2.5 max-md:pt-1.5">
      <span class="text-body text-text-secondary max-md:text-chip" data-testid="coin-card-quantity">{{ wishlist ? '' : t('coinCard.quantity', { count: coin.quantity || 1 }) }}</span>
      <span v-if="value != null" class="text-base font-semibold text-gold max-md:text-body" data-testid="coin-card-value">{{ formatCurrency(value) }}</span>
      <span v-else class="text-body text-text-muted" aria-hidden="true">—</span>
    </div>
    <button
      v-if="wishlist"
      class="btn btn-primary btn-sm flex w-full justify-center gap-[0.35rem] text-chip"
      @click.stop="emit('purchase', coin)"
    >
      <ShoppingCart :size="14" /> {{ t('coinCard.purchased') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import type { Coin, ImageType } from '@/types'
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Banknote, Check, Coins, ShoppingCart } from 'lucide-vue-next'
import { formatCurrency, materialLabel } from '@/utils/format'
import SafeExternalLink from '@/components/SafeExternalLink.vue'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { t } from '@/i18n'
import { categoryLabel } from '@/i18n/coinVocabulary'
import { coinGradeLabel, monetarySystemShortName } from '@/utils/coinAttributes'
import { numistaGradeForCoinGrade } from '@/utils/numistaGrades'
import { formatNumistaYear } from '@/utils/numistaCatalog'

const router = useRouter()

const props = withDefaults(defineProps<{
  coin: Coin
  imageSide?: ImageType | null
  wishlist?: boolean
  selectable?: boolean
  selected?: boolean
}>(), {
  imageSide: null,
  wishlist: false,
  selectable: false,
  selected: false,
})

const emit = defineEmits<{
  purchase: [coin: Coin]
  'dismiss-status': [coinId: number]
  'toggle-select': [coinId: number]
}>()

function handleClick() {
  if (props.selectable) {
    emit('toggle-select', props.coin.id)
  } else {
    router.push(`/coin/${props.coin.id}`)
  }
}

// The coin's monetary system, without its dates ("Escudo"), or else its category.
const badge = computed(() => {
  const system = props.coin.monetarySystem?.trim()
  if (system) return monetarySystemShortName(system)
  return props.coin.category ? categoryLabel(props.coin.category) : ''
})

const primaryImage = computed(() => {
  if (props.imageSide) {
    const byType = props.coin.images?.find((img) => img.imageType === props.imageSide)
    if (byType) return byType.filePath
  }
  const primary = props.coin.images?.find((img) => img.isPrimary)
  const first = props.coin.images?.[0]
  const img = primary || first
  return img ? img.filePath : null
})

// "1924 · Bronze": its year (or dates) and material, before its kind.
const details = computed(() => [
  props.coin.year != null ? formatNumistaYear(props.coin.year) : props.coin.dateRange?.trim() ?? '',
  props.coin.material && props.coin.material !== 'Other' ? materialLabel(props.coin.material) : '',
].filter(Boolean))

// Its grade in short ("UNC", "VF"): as written when short, else on Numista's scale.
const gradeBadge = computed(() => {
  const grade = props.coin.grade?.trim()
  if (!grade) return ''
  if (grade.length <= 4) return grade.toUpperCase()
  return numistaGradeForCoinGrade(grade)?.toUpperCase() ?? ''
})
const gradeTitle = computed(() => (props.coin.grade?.trim() ? coinGradeLabel(props.coin.grade.trim()) : undefined))

// Its current value; on the wishlist, the price asked when it has none.
const value = computed(() => props.coin.currentValue ?? (props.wishlist ? props.coin.purchasePrice ?? null : null))
</script>
