<template>
  <button
    class="card grid w-full cursor-pointer grid-cols-[auto_minmax(0,1fr)] items-center gap-5 !p-5 text-left max-sm:gap-3 max-sm:!p-3"
    type="button"
    :title="set.name"
    data-testid="set-card"
    @click="$emit('click')"
  >
    <!-- A set of different coins: three of them piled up (blank coins for those without a picture). -->
    <span v-if="stack" class="relative h-28 w-28 shrink-0 max-sm:h-20 max-sm:w-20" data-testid="set-card-cover" data-cover="stack">
      <span
        v-for="(picture, index) in stackPictures"
        :key="index"
        class="absolute flex h-[62%] w-[62%] items-center justify-center overflow-hidden rounded-full border-2 border-card shadow-card"
        :class="[STACK_PLACES[index], picture ? 'bg-input' : 'bg-gold-dim']"
        data-testid="set-card-stack-coin"
      >
        <AuthenticatedImage v-if="picture?.path" :media-path="picture.path" alt="" loading="lazy" class="h-full w-full object-cover">
          <template #fallback><span class="absolute inset-[14%] rounded-full border-2 border-gold/50"></span></template>
        </AuthenticatedImage>
        <img v-else-if="picture?.url" :src="picture.url" alt="" loading="lazy" class="h-full w-full object-cover" />
        <span v-else class="absolute inset-[14%] rounded-full border-2 border-gold/50" data-testid="set-card-blank-coin"></span>
      </span>
    </span>

    <!-- One coin: its picture, round as a coin; without one, a ring of the set's colour. -->
    <span
      v-else
      class="flex h-28 w-28 shrink-0 items-center justify-center overflow-hidden rounded-full max-sm:h-20 max-sm:w-20"
      :style="pictures.length ? undefined : { boxShadow: `inset 0 0 0 3px ${set.color}` }"
      data-testid="set-card-cover"
      data-cover="coin"
    >
      <AuthenticatedImage v-if="pictures[0]?.path" :media-path="pictures[0].path" alt="" loading="lazy" class="h-full w-full object-cover">
        <template #fallback><Coins :size="36" :stroke-width="1.25" class="text-gold-dim" /></template>
      </AuthenticatedImage>
      <img v-else-if="pictures[0]?.url" :src="pictures[0].url" alt="" loading="lazy" class="h-full w-full object-cover" />
      <Coins v-else :size="36" :stroke-width="1.25" class="text-gold-dim" />
    </span>

    <span class="flex min-w-0 flex-col gap-3 self-stretch max-sm:gap-2">
      <!-- Its year and name; at the right, how many of its coins the collection has. -->
      <span class="flex items-start justify-between gap-3">
        <span class="flex min-w-0 flex-col items-start gap-2">
          <span v-if="set.year" class="rounded-full bg-gold px-3 py-0.5 text-body font-semibold text-surface" data-testid="set-card-year">{{ set.year }}</span>
          <span class="overflow-hidden font-sans text-xl font-bold leading-tight text-text-primary max-sm:text-md [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:2]" data-testid="set-card-title">{{ title }}</span>
        </span>
        <span class="shrink-0 whitespace-nowrap text-right" data-testid="set-card-count">
          <span class="font-sans text-4xl font-bold leading-none text-text-primary max-sm:text-2xl">{{ formatNumber(set.coinCount) }}</span>
          <span v-if="set.total != null" class="text-lg text-text-secondary max-sm:text-body"> / {{ formatNumber(set.total) }}</span>
          <span v-else class="block text-chip font-semibold text-text-secondary">{{ set.coinCount === 1 ? t('sets.card.coin') : t('sets.card.coins') }}</span>
        </span>
      </span>

      <!-- How far it is, when it can be completed; else what kind of set it is. -->
      <span v-if="percent != null" class="mt-auto flex items-center gap-3" data-testid="set-card-progress">
        <span class="h-2 min-w-0 flex-1 overflow-hidden rounded-full bg-input" aria-hidden="true">
          <span class="block h-full rounded-full bg-gold" :style="{ width: `${Math.min(100, percent)}%` }" />
        </span>
        <span class="shrink-0 text-body font-semibold text-gold">{{ formatShare(percent / 100) }}</span>
      </span>
      <span v-else class="mt-auto text-body text-text-muted" data-testid="set-card-kind">{{ kind }}</span>
    </span>
  </button>
</template>

<script setup lang="ts">
// A set on the sets page: a picture of its coins, its year (a joint issue's,
// an annual series'), its name and how many of its coins the collection has
// (of how many, for a list to complete), then how far it is.
import { computed } from 'vue'
import { Coins } from 'lucide-vue-next'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { normalizeCoinSetType } from '@/types'
import type { CoinSetSummary } from '@/types'
import { t } from '@/i18n'
import { formatNumber, formatShare } from '@/utils/numberFormat'
import { setDisplayTitle } from '@/utils/setTitle'

const props = defineProps<{
  set: CoinSetSummary
}>()

defineEmits<{
  (e: 'click'): void
}>()

// Pictures of its coins: the catalogue's (a link) or the collection's (a media path).
const pictures = computed(() => [
  ...(props.set.coverImageUrls ?? []).map((url) => ({ url, path: '' })),
  ...(props.set.coverImages ?? []).map((path) => ({ url: '', path })),
])

// More than one coin (a joint issue is one coin, struck by every country): a pile of three.
const stack = computed(() => !props.set.oneCoin && (props.set.total ?? props.set.coinCount) > 1)
// Back left, back right, then the first picture in front.
const STACK_PLACES = ['left-0 top-0 z-[1]', 'right-0 top-0 z-[2]', 'bottom-0 left-[19%] z-[3]']
const stackPictures = computed(() => [pictures.value[1], pictures.value[2], pictures.value[0]])

// Its year is in the pill: the name is what comes after it.
const title = computed(() => setDisplayTitle(props.set.name, props.set.oneCoin ? props.set.year : undefined))

// An agentic set's list is not known while it is being made.
const generating = computed(() => normalizeCoinSetType(props.set.setType) === 'agentic' && props.set.agenticStatus === 'generating')

const percent = computed(() => {
  const type = normalizeCoinSetType(props.set.setType)
  if ((type !== 'checklist' && type !== 'goal' && type !== 'agentic') || generating.value) return null
  return props.set.completionPercentage ?? 0
})

// What kind of set it is, for those without a completion.
const kind = computed(() => {
  if (generating.value) return t('sets.card.agenticGenerating')
  if (normalizeCoinSetType(props.set.setType) === 'smart') return t('sets.card.smart')
  return props.set.coinCount > 0 ? t('sets.card.standard') : t('sets.card.ready')
})
</script>
