<template>
  <!-- Two by two on phones (the coins' wishlist on a row of its own); in a row on a wide screen. -->
  <section
    class="grid grid-cols-2 gap-3"
    :class="wishlist == null ? 'lg:grid-cols-4' : 'lg:grid-cols-5'"
    :aria-label="t('collectionOverview.kpis.label')"
  >
    <div :class="CARD" :title="t('collectionOverview.kpis.itemsHint')" :data-testid="`${testId}-items`">
      <span :class="VALUE">{{ count(summary?.coins) }}</span>
      <span :class="LABEL">{{ itemsLabel }}</span>
      <span v-if="summary" :class="NOTE">{{ t('collectionOverview.kpis.unique', { count: summary.uniqueCoins ?? summary.coins }) }}</span>
    </div>

    <div :class="CARD" :title="valueHint" :data-testid="`${testId}-value`">
      <span :class="VALUE">{{ summary ? formatCurrency(summary.totalValue) : '—' }}</span>
      <span :class="LABEL">{{ t('collectionOverview.kpis.value') }}</span>
      <span v-if="faceLine" :class="NOTE" :title="faceHint" :data-testid="`${testId}-face`">{{ faceLine }}</span>
    </div>

    <!-- What the pieces with a purchase price cost. -->
    <div :class="CARD" :title="investedHint" :data-testid="`${testId}-invested`">
      <!-- None with a purchase price: nothing is known of what they cost. -->
      <span :class="VALUE">{{ summary?.invested != null && summary.investedPieces ? formatCurrency(summary.invested) : '—' }}</span>
      <span :class="LABEL">{{ t('collectionOverview.kpis.invested') }}</span>
      <span v-if="summary?.investedPieces != null" :class="NOTE">
        {{ summary.investedPieces ? t('collectionOverview.kpis.investedPieces', { count: summary.investedPieces }) : t('collectionOverview.kpis.investedNone') }}
      </span>
    </div>

    <div :class="CARD" :title="countriesHint" :data-testid="`${testId}-countries`">
      <span :class="VALUE">{{ count(summary?.countries) }}</span>
      <span :class="LABEL">{{ t('collectionOverview.kpis.countries') }}</span>
      <template v-if="summary?.extinctCountries">
        <button
          v-if="extinctLink"
          type="button"
          :class="NOTE" class="rounded-sm font-medium text-gold underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :title="t('collectionOverview.kpis.extinctHint')"
          :data-testid="`${testId}-extinct`"
          @click="emit('show-extinct')"
        >
          {{ t('collectionOverview.kpis.extinct', { count: summary.extinctCountries }) }}
        </button>
        <span v-else :class="NOTE" :data-testid="`${testId}-extinct`">{{ t('collectionOverview.kpis.extinct', { count: summary.extinctCountries }) }}</span>
      </template>
    </div>

    <!-- For the coins: those on the wishlist, which opens it. -->
    <router-link
      v-if="wishlist != null"
      to="/wishlist"
      :class="[CARD, 'col-span-2 no-underline transition-colors hover:border-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2 sm:col-span-1']"
      :title="t('collectionOverview.kpis.wishlistHint')"
      :data-testid="`${testId}-wishlist`"
    >
      <span :class="VALUE">{{ count(wishlist) }}</span>
      <span :class="LABEL">{{ t('collectionOverview.kpis.wishlist') }}</span>
    </router-link>
  </section>
</template>

<script setup lang="ts">
// The headline numbers of the whole collection, its coins or its banknotes:
// how many pieces (counting quantities) and how many different ones; what
// they are worth altogether in the currency every page uses, with their face
// value and how much more than it they are worth; what those with a purchase
// price cost; and from how many countries, of which how many no longer exist
// (for the coins, a link to their coins). What the cards leave out (items
// without a value, a face value or a country) is in their tooltips. The
// coins' have one more card: how many are on the wishlist.
import { computed } from 'vue'
import type { CollectionSummary } from '@/api/client'
import { t } from '@/i18n'
import { numberFormat } from '@/utils/numberFormat'
import { collectionCurrency, currencyName } from '@/utils/collectionCurrency'
import { formatCurrency } from '@/utils/format'

const props = withDefaults(defineProps<{
  summary: CollectionSummary | null
  kind?: 'all' | 'coin' | 'banknote'
  /** The extinct countries open their coins. */
  extinctLink?: boolean
  /** How many are on the wishlist: a fourth card, when given. */
  wishlist?: number | null
}>(), { kind: 'all', extinctLink: false, wishlist: null })

const emit = defineEmits<{ 'show-extinct': [] }>()

const CARD = 'flex min-w-0 flex-col items-center gap-1 rounded-md border border-border-subtle bg-card p-4 text-center md:px-5'
const VALUE = 'whitespace-nowrap font-display text-xl font-semibold text-text-primary md:text-2xl'
const LABEL = 'text-label font-semibold uppercase tracking-[0.08em] text-text-muted'
const NOTE = 'text-chip text-text-muted'

const testId = computed(() => `overview-${props.kind}`)

function count(value: number | undefined): string {
  return value === undefined ? '—' : numberFormat().format(value)
}

const itemsLabel = computed(() => {
  if (props.kind === 'coin') return t('collectionOverview.kpis.coins')
  if (props.kind === 'banknote') return t('collectionOverview.kpis.banknotes')
  return t('collectionOverview.kpis.items')
})

// "Facial: 538,27 € (+89,1%)": how much more than their face value they are worth.
const faceLine = computed(() => {
  const face = props.summary?.faceValue
  if (face == null) return ''
  const over = props.summary?.valueOverFace
  const value = formatCurrency(face)
  if (over == null) return t('collectionOverview.kpis.face', { value })
  const percent = numberFormat({ style: 'percent', signDisplay: 'exceptZero', maximumFractionDigits: 1 }).format(over / 100)
  return t('collectionOverview.kpis.faceOver', { value, over: percent })
})

function lines(...parts: string[]): string {
  return parts.filter(Boolean).join('\n')
}

const valueHint = computed(() => lines(
  t('collectionOverview.kpis.valueHint', { currency: currencyName(collectionCurrency.value) }),
  props.summary?.unvaluedCoins ? t('collectionOverview.kpis.unvalued', { count: props.summary.unvaluedCoins }) : '',
))

const faceHint = computed(() => lines(
  t('collectionOverview.kpis.faceHint', { currency: currencyName(collectionCurrency.value) }),
  props.summary?.valueOverFace == null ? '' : t('collectionOverview.kpis.overFaceHint'),
  props.summary?.withoutFaceValue ? t('collectionOverview.kpis.withoutFaceValue', { count: props.summary.withoutFaceValue }) : '',
))

const investedHint = computed(() => t('collectionOverview.kpis.investedHint', { currency: currencyName(collectionCurrency.value) }))

const countriesHint = computed(() => lines(
  t('collectionOverview.kpis.countriesHint'),
  props.summary?.withoutCountry ? t('collectionOverview.kpis.withoutCountry', { count: props.summary.withoutCountry }) : '',
))
</script>

