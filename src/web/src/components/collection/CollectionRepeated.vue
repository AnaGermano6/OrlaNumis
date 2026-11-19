<template>
  <section :aria-labelledby="titleId" data-testid="trade-list">
    <!-- Its title; copying the list as text, or sharing it as a file (a page of its own has them at its top). -->
    <div v-if="header" class="flex flex-wrap items-center justify-between gap-3">
      <h2 :id="titleId" class="m-0 font-display text-xl font-semibold text-heading md:text-2xl">{{ title }}</h2>
      <span class="flex flex-wrap items-center gap-2">
        <button type="button" class="btn btn-secondary" :disabled="!shown.length" data-testid="trade-copy" @click="copy">
          <Copy :size="16" aria-hidden="true" /> {{ t('collectionOverview.repeated.copy') }}
        </button>
        <button type="button" class="btn btn-primary" data-testid="trade-share" @click="emit('share')">
          <Share2 :size="16" aria-hidden="true" /> {{ t('collectionOverview.repeated.share') }}
        </button>
      </span>
    </div>
    <hr v-if="header" class="my-4 border-0 border-t border-border-accent" data-testid="wishlist-trades-divider" />

    <!-- The coins or the banknotes; at the right, one searched for, or a country's. -->
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-wrap gap-2" role="group" :aria-label="title">
        <button
          v-for="option in KINDS"
          :key="option"
          type="button"
          class="inline-flex items-center gap-1.5 rounded-full border px-4 py-1.5 text-body transition-colors"
          :class="option === kind ? 'border-gold bg-gold font-semibold text-surface' : 'border-border-subtle text-text-primary hover:border-gold'"
          :aria-pressed="option === kind"
          :data-testid="`trade-kind-${option}`"
          @click="choose(option)"
        >
          <Check v-if="option === kind" :size="16" aria-hidden="true" />
          {{ t(option === 'coin' ? 'collectionOverview.repeated.tabCoins' : 'collectionOverview.repeated.tabBanknotes', { count: lists[option].length }) }}
        </button>
      </div>
      <div class="flex w-full flex-wrap items-center gap-2 sm:w-auto">
        <SearchBar v-model="search" box class="!max-w-none w-full sm:w-64" :placeholder="t(kind === 'coin' ? 'collectionOverview.repeated.searchCoins' : 'collectionOverview.repeated.searchBanknotes')" data-testid="trade-search" />
        <select v-model="country" class="form-select !w-auto min-w-40" :aria-label="t('wishlist.filters.country')" data-testid="trade-country">
          <option value="">{{ t('wishlist.filters.allCountries') }}</option>
          <option v-for="option in countries" :key="option" :value="option">{{ withCountryFlag(option) }}</option>
        </select>
      </div>
    </div>

    <p v-if="failed" class="m-0 text-body text-text-secondary" role="alert">{{ t('collectionOverview.repeated.failed') }}</p>
    <p v-else-if="loaded && !lists[kind].length" class="m-0 text-body text-text-secondary" data-testid="repeated-empty">
      {{ kind === 'banknote' ? t('collectionOverview.repeated.emptyBanknotes') : t('collectionOverview.repeated.emptyCoins') }}
    </p>
    <p v-else-if="loaded && !shown.length" class="m-0 text-body text-text-secondary" data-testid="trade-no-match">{{ t('collectionOverview.repeated.noMatch') }}</p>
    <!-- Each coin: its photo, name, country and year; how many there are, and how many to trade. -->
    <ul v-else class="m-0 grid list-none gap-4 p-0 lg:grid-cols-2">
      <li v-for="coin in shown" :key="coin.id" class="card flex min-w-0 items-center gap-5 !p-5 max-sm:gap-3 max-sm:!p-3" data-testid="repeated-item">
        <RouterLink
          :to="`/coin/${coin.id}`"
          class="flex shrink-0 items-center justify-center overflow-hidden bg-input"
          :class="kind === 'banknote' ? 'h-24 w-36 rounded-md max-sm:h-16 max-sm:w-24' : 'h-28 w-28 rounded-full max-sm:h-20 max-sm:w-20'"
          :title="coin.name"
        >
          <AuthenticatedImage v-if="thumbnail(coin)" :media-path="thumbnail(coin)" :alt="coin.name" class="h-full w-full" :class="kind === 'banknote' ? 'object-contain' : 'object-cover'" loading="lazy">
            <template #fallback><component :is="kind === 'banknote' ? Banknote : Coins" :size="28" :stroke-width="1" class="text-text-primary/30" /></template>
          </AuthenticatedImage>
          <component :is="kind === 'banknote' ? Banknote : Coins" v-else :size="28" :stroke-width="1" class="text-text-primary/30" />
        </RouterLink>
        <div class="flex min-w-0 flex-1 flex-col gap-2 self-stretch">
          <RouterLink :to="`/coin/${coin.id}`" class="line-clamp-2 font-sans text-md font-bold leading-snug text-text-primary no-underline hover:text-gold" :title="coin.name">{{ coin.name }}</RouterLink>
          <span v-if="coin.country || yearText(coin)" class="truncate text-body text-text-secondary" data-testid="repeated-origin">
            <span v-if="coin.country && hasCountryFlag(coin.country)" class="mr-1.5" aria-hidden="true"><CountryFlag :country="coin.country" /></span>{{ [coin.country, yearText(coin)].filter(Boolean).join(' · ') }}
          </span>
          <span class="mt-auto flex flex-col items-end gap-1.5">
            <span class="text-chip text-text-muted" data-testid="repeated-total">{{ t('collectionOverview.repeated.total', { count: coin.quantity }) }}</span>
            <!-- One fewer, or one more, to trade (how many of it the collection has). -->
            <span class="inline-flex items-center gap-1 rounded-full border border-border-accent px-1 py-0.5" data-testid="repeated-stepper">
              <button
                type="button"
                class="inline-flex h-7 w-7 items-center justify-center rounded-full text-text-primary transition-colors hover:bg-gold-glow disabled:opacity-30"
                :disabled="coin.quantity <= 1 || saving === coin.id"
                :aria-label="t('collectionOverview.repeated.fewer', { name: coin.name })"
                data-testid="repeated-fewer"
                @click="change(coin, -1)"
              >
                <Minus :size="14" aria-hidden="true" />
              </button>
              <span class="px-1 text-body font-semibold text-text-primary tabular-nums" data-testid="repeated-spare">{{ t('collectionOverview.repeated.spare', { count: coin.quantity - 1 }) }}</span>
              <button
                type="button"
                class="inline-flex h-7 w-7 items-center justify-center rounded-full text-text-primary transition-colors hover:bg-gold-glow disabled:opacity-30"
                :disabled="coin.quantity >= 9999 || saving === coin.id"
                :aria-label="t('collectionOverview.repeated.more', { name: coin.name })"
                data-testid="repeated-more"
                @click="change(coin, 1)"
              >
                <Plus :size="14" aria-hidden="true" />
              </button>
            </span>
          </span>
        </div>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
// The trade list: the coins and the banknotes held more than once, most
// copies first, each with how many there are and how many are left to
// trade, which − and + change (the quantity the collection has; one left
// with none to trade stays here until the page is opened again). A coin
// searched for or a country's; the list copied as text, or shared as a file.
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { Banknote, Check, Coins, Copy, Minus, Plus, Share2 } from 'lucide-vue-next'
import { getCoins, updateCoinQuantity } from '@/api/client'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import CountryFlag from '@/components/CountryFlag.vue'
import SearchBar from '@/components/SearchBar.vue'
import { useToast } from '@/composables/useToast'
import { intlLocale, t } from '@/i18n'
import type { Coin } from '@/types'
import { hasCountryFlag, withCountryFlag } from '@/utils/countryFlags'
import { formatNumistaYear } from '@/utils/numistaCatalog'

type Kind = 'coin' | 'banknote'
const KINDS: Kind[] = ['coin', 'banknote']

withDefaults(defineProps<{
  title?: string
  /** Its own title and buttons; off on a page that has them. */
  header?: boolean
}>(), { title: () => t('wishlist.trades.title'), header: true })

const emit = defineEmits<{ share: [] }>()

const { showToast } = useToast()
const titleId = 'trade-list-title'
const lists = ref<Record<Kind, Coin[]>>({ coin: [], banknote: [] })
const loaded = ref(false)
const failed = ref(false)
const kind = ref<Kind>('coin')
const search = ref('')
const country = ref('')
/** The coin whose quantity is being saved. */
const saving = ref<number | null>(null)

function choose(option: Kind) {
  kind.value = option
  country.value = ''
}

function fold(value: string): string {
  return value.toLocaleLowerCase(intlLocale()).normalize('NFD').replace(/[̀-ͯ]/g, '')
}

// The countries of the coins (or banknotes) of the tab, by name.
const countries = computed(() => [...new Set(lists.value[kind.value].map((coin) => coin.country?.trim() ?? '').filter(Boolean))]
  .sort((a, b) => a.localeCompare(b, intlLocale())))

const shown = computed(() => {
  const query = fold(search.value.trim())
  return lists.value[kind.value].filter((coin) => (!country.value || coin.country?.trim() === country.value)
    && (!query || fold(coin.name).includes(query)))
})

function thumbnail(coin: Coin): string | undefined {
  const images = coin.images ?? []
  return (images.find((image) => image.isPrimary) ?? images.find((image) => image.imageType === 'obverse') ?? images[0])?.filePath
}

function yearText(coin: Coin): string {
  return coin.year != null ? formatNumistaYear(coin.year) : coin.dateRange ?? ''
}

// One more, or one fewer, of the coin: its quantity saved, nothing else.
async function change(coin: Coin, by: number) {
  const quantity = coin.quantity + by
  if (quantity < 1 || saving.value === coin.id) return
  saving.value = coin.id
  try {
    await updateCoinQuantity(coin.id, quantity)
    coin.quantity = quantity
  } catch {
    showToast(t('collectionOverview.repeated.saveFailed'), 'error')
  } finally {
    saving.value = null
  }
}

// The list shown, as lines of text to paste in a message.
async function copy() {
  const lines = shown.value.map((coin) => `• ${coin.name}${[coin.country, yearText(coin)].filter(Boolean).length ? ` (${[coin.country, yearText(coin)].filter(Boolean).join(', ')})` : ''}: ${t('collectionOverview.repeated.spare', { count: coin.quantity - 1 })}`)
  const text = [t(kind.value === 'coin' ? 'collectionOverview.repeated.copyCoins' : 'collectionOverview.repeated.copyBanknotes'), ...lines].join('\n')
  try {
    await navigator.clipboard.writeText(text)
    showToast(t('collectionOverview.repeated.copied'), 'success')
  } catch {
    showToast(t('collectionOverview.repeated.copyFailed'), 'error')
  }
}

// Whether there is a list to copy, and copying it: for a page's own buttons.
const canCopy = computed(() => shown.value.length > 0)
defineExpose({ copy, canCopy })

onMounted(async () => {
  try {
    const [coins, banknotes] = await Promise.all(KINDS.map((option) => getCoins({ kind: option, repeated: 'true', wishlist: 'false', sort: 'quantity', order: 'desc', limit: 100 })))
    lists.value = { coin: coins!.data.coins, banknote: banknotes!.data.coins }
    // Banknotes only: their tab first.
    if (!lists.value.coin.length && lists.value.banknote.length) kind.value = 'banknote'
  } catch {
    failed.value = true
  } finally {
    loaded.value = true
  }
})
</script>
