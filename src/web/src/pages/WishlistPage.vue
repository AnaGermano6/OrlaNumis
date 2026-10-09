<template>
  <div class="container">
    <DesktopPageContext :title="t('nav.wishlist')">
      <template #actions>
        <AppIconButton
          :title="t('wishlist.checkAvailability')"
          :disabled="checking"
          @click="handleCheckAvailability"
        >
          <span
            v-if="checking"
            class="inline-block h-[14px] w-[14px] animate-spin rounded-full border-2 border-border-subtle border-t-gold"
          ></span>
          <ShieldCheck v-else :size="20" />
        </AppIconButton>
        <AppIconButton
          v-if="store.coins.length"
          to="/wishlist/search-alerts"
          :title="t('wishlist.searchAlerts')"
        >
          <CalendarClock :size="20" />
        </AppIconButton>
        <AppIconButton to="/wishlist/availability-runs" :title="t('wishlist.runHistory')">
          <History :size="20" />
        </AppIconButton>
        <AppIconButton :title="t('wishlist.addTo')" @click="showAddDialog = true">
          <CirclePlus :size="20" />
        </AppIconButton>
      </template>
      <div class="page-header">
        <h1>{{ t('nav.wishlist') }}</h1>
        <div v-if="isPwa" class="pwa-actions">
          <button
            class="pwa-icon-btn focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
            :disabled="checking"
            @click="handleCheckAvailability"
            :title="t('wishlist.checkAvailability')"
          >
            <span
              v-if="checking"
              class="inline-block h-[14px] w-[14px] animate-spin rounded-full border-2 border-border-subtle border-t-gold"
            ></span>
            <ShieldCheck v-else :size="22" />
          </button>
          <router-link
            v-if="store.coins.length"
            to="/wishlist/search-alerts"
            class="pwa-icon-btn focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
            :title="t('wishlist.finderAgent')"
            :aria-label="t('wishlist.finderAgent')"
          >
            <CalendarClock :size="22" />
          </router-link>
          <router-link
            to="/wishlist/availability-runs"
            class="pwa-icon-btn focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
            :title="t('wishlist.runHistory')"
            :aria-label="t('wishlist.runHistory')"
          >
            <History :size="22" />
          </router-link>
          <button
            type="button"
            class="pwa-icon-btn focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
            :title="t('wishlist.addTo')"
            :aria-label="t('wishlist.addTo')"
            @click="showAddDialog = true"
          >
            <CirclePlus :size="22" />
          </button>
        </div>
        <div v-else class="header-actions flex-wrap gap-3">
          <button class="btn btn-secondary" :disabled="checking" @click="handleCheckAvailability">
            <ShieldCheck :size="16" /> {{ checking ? t('common.checking') : t('wishlist.checkAvailability') }}
          </button>
          <router-link v-if="store.coins.length" to="/wishlist/search-alerts" class="btn btn-secondary">
            <CalendarClock :size="16" /> {{ t('wishlist.searchAlerts') }}
          </router-link>
          <router-link to="/wishlist/availability-runs" class="btn btn-secondary">
            <History :size="16" /> {{ t('wishlist.runHistory') }}
          </router-link>
          <button type="button" class="btn btn-secondary" @click="showAddDialog = true">
            <CirclePlus :size="16" /> {{ t('nav.addCoin') }}
          </button>
        </div>
      </div>
    </DesktopPageContext>

    <div
      v-if="checkResult"
      class="mb-4 flex flex-wrap items-center gap-4 rounded-md border border-border-subtle bg-card px-4 py-3 text-body"
    >
      <span class="inline-flex rounded-full bg-status-success-bg px-2.5 py-1 text-chip font-semibold text-status-success-fg">
        {{ t('wishlist.check.available', { count: checkResult.available }) }}
      </span>
      <span class="inline-flex rounded-full bg-status-error-bg px-2.5 py-1 text-chip font-semibold text-status-error-fg">
        {{ t('wishlist.check.unavailable', { count: checkResult.unavailable }) }}
      </span>
      <span class="inline-flex rounded-full bg-status-warning-bg px-2.5 py-1 text-chip font-semibold text-status-warning-fg">
        {{ t('wishlist.check.unknown', { count: checkResult.unknown }) }}
      </span>
      <span class="ml-auto text-text-muted">{{ t('wishlist.check.checked', { count: checkResult.coinsChecked }) }}</span>
      <button
        class="px-1 text-xl leading-none text-text-muted transition hover:text-text-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
        @click="checkResult = null"
      >
        &times;
      </button>
    </div>

    <div v-if="store.loading" class="loading-overlay">
      <div class="spinner"></div>
    </div>

    <div v-else-if="store.coins.length" class="coins-grid">
      <CoinCard
        v-for="coin in store.coins"
        :key="coin.id"
        :coin="coin"
        wishlist
        :active-reminder="remindersByCoinId.get(coin.id) ?? null"
        @purchase="openPurchaseModal"
        @dismiss-status="handleDismissStatus"
      />
    </div>

    <div v-else class="empty-state">
      <h3>{{ t('wishlist.empty.title') }}</h3>
      <p>{{ t('wishlist.empty.body') }}</p>
      <div class="mt-3 flex flex-wrap justify-center gap-3">
        <button class="btn btn-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" @click="showChat = true">
          <Bot :size="16" /> {{ t('wishlist.empty.searchAi') }}
        </button>
        <router-link
          to="/wishlist/search-alerts"
          class="btn btn-secondary focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
          :title="t('wishlist.finderAgent')"
        >
          <CalendarClock :size="16" /> {{ t('wishlist.finderAgent') }}
        </router-link>
      </div>
    </div>

    <div v-if="store.coins.length && store.total > pageSize" class="mt-6 flex flex-wrap items-center justify-center gap-4 py-4">
      <button class="btn btn-secondary btn-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" :disabled="page <= 1" @click="page--">← {{ t('common.previous') }}</button>
      <span class="text-body text-text-secondary">{{ t('common.pageOf', { page, total: Math.ceil(store.total / pageSize) }) }}</span>
      <button class="btn btn-secondary btn-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" :disabled="page * pageSize >= store.total" @click="page++">{{ t('common.next') }} →</button>
    </div>

    <PurchaseModal
      v-if="purchaseTarget"
      :coin="purchaseTarget"
      @close="purchaseTarget = null"
      @confirm="handlePurchaseConfirm"
    />

    <div
      v-if="showAddDialog"
      class="fixed inset-0 z-[1000] flex items-center justify-center bg-overlay p-4"
      @click.self="closeAddDialog"
      @keydown.esc="closeAddDialog"
    >
      <div
        ref="addDialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="wishlist-add-title"
        class="card max-h-[90vh] w-full max-w-2xl overflow-y-auto !p-6"
        tabindex="-1"
      >
        <div class="mb-4 flex items-center justify-between gap-3">
          <h2 id="wishlist-add-title" class="m-0 text-heading">{{ t('wishlist.addTo') }}</h2>
          <button
            type="button"
            class="pwa-icon-btn"
            :title="t('common.close')"
            :aria-label="t('common.close')"
            @click="closeAddDialog"
          >
            <X :size="18" />
          </button>
        </div>

        <div v-if="addMethod === null" class="grid gap-3 sm:grid-cols-2">
          <button type="button" class="add-method" @click="addMethod = 'url'">
            <Link :size="24" />
            <span>
              <strong>{{ t('wishlist.add.fromUrl') }}</strong>
              <small>{{ t('wishlist.add.fromUrlHint') }}</small>
            </span>
          </button>
          <router-link to="/lookup" class="add-method" @click="closeAddDialog">
            <ScanLine :size="24" />
            <span>
              <strong>{{ t('wishlist.add.fromImage') }}</strong>
              <small>{{ t('wishlist.add.fromImageHint') }}</small>
            </span>
          </router-link>
        </div>

        <div v-else>
          <button type="button" class="btn btn-ghost btn-xs mb-4" @click="addMethod = null">
            <ChevronLeft :size="15" /> {{ t('wishlist.add.another') }}
          </button>
          <WishlistURLIntake embedded @created="handleURLCreated" />
        </div>
      </div>
    </div>

    <CoinSearchChat v-if="showChat" @close="showChat = false" @added="loadCoins" />
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref, watch, onBeforeUnmount } from 'vue'
import { useCoinsStore } from '@/stores/coins'
import CoinCard from '@/components/CoinCard.vue'
import CoinSearchChat from '@/components/CoinSearchChat.vue'
import PurchaseModal from '@/components/PurchaseModal.vue'
import WishlistURLIntake from '@/components/wishlist/WishlistURLIntake.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { purchaseCoin, checkWishlistAvailability, updateListingStatus, listPurchaseReminders } from '@/api/client'
import type { Coin, AvailabilityRunSummary, PurchaseReminder } from '@/types'
import { CirclePlus, Bot, ShieldCheck, CalendarClock, History, ChevronLeft, Link, ScanLine, X } from 'lucide-vue-next'
import { usePwa } from '@/composables/usePwa'
import { useQuickAccess } from '@/composables/useQuickAccess'
import { t } from '@/i18n'

const store = useCoinsStore()
const { isPwa } = usePwa()
const { refresh: refreshQuickAccess } = useQuickAccess()
const showChat = ref(false)
const showAddDialog = ref(false)
const addMethod = ref<'url' | null>(null)
const addDialog = ref<HTMLElement | null>(null)
const purchaseTarget = ref<Coin | null>(null)
const checking = ref(false)
const checkResult = ref<AvailabilityRunSummary | null>(null)
let dismissTimer: ReturnType<typeof setTimeout> | null = null
const page = ref(1)
const pageSize = 50

const remindersByCoinId = ref(new Map<number, PurchaseReminder>())

async function fetchReminderMap() {
  try {
    const res = await listPurchaseReminders()
    const map = new Map<number, PurchaseReminder>()
    for (const r of res.data.reminders ?? []) {
      map.set(r.coinId, r)
    }
    remindersByCoinId.value = map
  } catch {
    // non-critical: badge is best-effort
  }
}

function loadCoins() {
  store.fetchCoins({ wishlist: 'true', sort: 'updated_at', order: 'desc', page: page.value })
  fetchReminderMap()
}

function closeAddDialog() {
  showAddDialog.value = false
  addMethod.value = null
}

function handleURLCreated() {
  loadCoins()
  closeAddDialog()
}

watch(page, loadCoins)
watch(showAddDialog, async (open) => {
  if (!open) return
  await nextTick()
  addDialog.value?.focus()
})

function openPurchaseModal(coin: Coin) {
  purchaseTarget.value = coin
}

async function handleCheckAvailability() {
  checking.value = true
  checkResult.value = null
  if (dismissTimer) { clearTimeout(dismissTimer); dismissTimer = null }
  try {
    const res = await checkWishlistAvailability()
    checkResult.value = res.data
    loadCoins()
    dismissTimer = setTimeout(() => { checkResult.value = null }, 10000)
  } catch {
    // silently fail
  } finally {
    checking.value = false
  }
}

async function handleDismissStatus(coinId: number) {
  try {
    await updateListingStatus(coinId, '')
    loadCoins()
  } catch {
    // silently fail
  }
}

async function handlePurchaseConfirm(data: { purchasePrice?: number; purchaseDate?: string; purchaseLocation?: string }) {
  if (!purchaseTarget.value) return
  try {
    await purchaseCoin(purchaseTarget.value.id, data)
    await refreshQuickAccess()
    purchaseTarget.value = null
    loadCoins()
  } catch {
    purchaseTarget.value = null
  }
}

loadCoins()

onBeforeUnmount(() => {
  if (dismissTimer) clearTimeout(dismissTimer)
})
</script>

<style scoped>
.add-method {
  display: flex;
  min-height: 96px;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1rem;
  text-align: left;
  color: var(--text-primary);
  background: var(--bg-input);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  transition: var(--transition-fast);
}

.add-method:hover {
  background: var(--bg-card-hover);
  border-color: var(--border-accent);
}

.add-method svg {
  flex-shrink: 0;
  color: var(--accent-gold);
}

.add-method span {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.add-method strong {
  font-size: 0.9rem;
}

.add-method small {
  color: var(--text-secondary);
  font-size: 0.8rem;
  line-height: 1.5;
}
</style>
