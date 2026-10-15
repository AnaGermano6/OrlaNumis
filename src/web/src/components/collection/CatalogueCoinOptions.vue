<template>
  <Teleport to="body">
    <div v-if="coin" class="fixed inset-0 z-[300] flex items-center justify-center bg-overlay-60 px-4" @click="emit('close')">
      <div
        class="flex max-h-[85vh] w-full max-w-[440px] flex-col rounded-md border border-border-subtle bg-card p-6 shadow-[0_12px_40px_var(--overlay-50)]"
        role="dialog"
        aria-modal="true"
        aria-labelledby="catalogue-options-title"
        data-testid="catalogue-options"
        @click.stop
        @keydown.esc="emit('close')"
      >
        <h3 id="catalogue-options-title" class="m-0 text-base text-heading">{{ coin.title }}</h3>
        <p class="mb-4 mt-1 text-chip text-text-muted">{{ releaseFaceValues(coin) }} · {{ releasePeriod(coin) }}</p>

        <button type="button" class="btn btn-secondary btn-sm self-start" :disabled="saving" data-testid="catalogue-hide" @click="emit('hide')">
          <EyeOff :size="16" /> {{ t('sets.catalogue.manage.hide') }}
        </button>
        <p class="mb-4 mt-1 text-chip text-text-muted">{{ t('sets.catalogue.manage.hideHint') }}</p>

        <h4 class="m-0 text-body font-medium text-text-primary">{{ t('sets.catalogue.manage.sameAs') }}</h4>
        <p class="mb-2 mt-1 text-chip text-text-muted">{{ t('sets.catalogue.manage.sameAsHint') }}</p>
        <input
          v-model="query"
          type="search"
          class="form-input mb-2"
          :placeholder="t('sets.catalogue.manage.search')"
          :aria-label="t('sets.catalogue.manage.search')"
          data-testid="catalogue-same-search"
        />
        <ul v-if="candidates.length" class="m-0 min-h-0 flex-1 list-none overflow-y-auto p-0">
          <li v-for="candidate in candidates" :key="candidate.id">
            <button
              type="button"
              class="flex w-full items-center gap-3 rounded-sm px-2 py-2 text-left transition-colors hover:bg-card-hover disabled:opacity-55"
              :disabled="saving"
              data-testid="catalogue-same-option"
              @click="emit('join', candidate.id!)"
            >
              <img v-if="candidate.imageUrl" :src="candidate.imageUrl" alt="" loading="lazy" class="h-10 w-10 shrink-0 object-contain" />
              <span v-else class="flex h-10 w-10 shrink-0 items-center justify-center text-text-primary/30"><Coins :size="24" :stroke-width="1" /></span>
              <span class="min-w-0 flex-1">
                <span class="block truncate text-body text-text-primary">{{ candidate.title }}</span>
                <span class="block text-chip text-text-muted">{{ releaseFaceValues(candidate) }} · {{ candidate.from.slice(0, 4) }}</span>
              </span>
            </button>
          </li>
        </ul>
        <p v-else class="m-0 text-body text-text-muted">{{ t('sets.catalogue.manage.noMatch') }}</p>

        <p v-if="failed" class="mb-0 mt-3 text-body text-loss" role="alert">{{ t('sets.catalogue.manage.failed') }}</p>
        <button type="button" class="btn btn-secondary btn-sm mt-3 self-end" @click="emit('close')">{{ t('common.cancel') }}</button>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
// A coin of the catalogue's options: hide it, or say it is the same coin as
// another of its country (the catalogue listing one coin twice), chosen from
// a list with the coins of its year first.
import { computed, ref, watch } from 'vue'
import { Coins, EyeOff } from 'lucide-vue-next'
import type { CoinRelease } from '@/api/client'
import { t } from '@/i18n'
import { releaseFaceValues, releasePeriod } from '@/utils/coinReleases'

const props = defineProps<{
  /** The coin, or null when closed. */
  coin: CoinRelease | null
  /** The country's other coins it may be the same as. */
  coins: CoinRelease[]
  saving?: boolean
  failed?: boolean
}>()

const emit = defineEmits<{
  hide: []
  join: [sameAs: string]
  close: []
}>()

const query = ref('')
watch(() => props.coin, () => { query.value = '' })

function plain(text: string): string {
  return text.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase()
}

const candidates = computed(() => {
  const coin = props.coin
  if (!coin) return []
  const year = Number(coin.from.slice(0, 4))
  const words = plain(query.value).split(/\s+/).filter(Boolean)
  return props.coins
    .filter((other) => other.id && other.id !== coin.id && words.every((word) => plain(other.title).includes(word)))
    .map((other, order) => ({ other, order, distance: Math.abs(Number(other.from.slice(0, 4)) - year) }))
    .sort((a, b) => a.distance - b.distance || a.order - b.order)
    .map(({ other }) => other)
})
</script>
