<template>
  <PullToRefresh :on-refresh="handleRefresh">
    <div class="container">
      <DesktopPageContext :title="t('nav.auctions')">
        <template #actions>
          <AppIconButton :disabled="syncing" :title="t('auctions.page.syncWatchlist')" @click="syncWatchlist">
            <RefreshCw :size="20" :class="syncing ? 'animate-spin' : ''" />
          </AppIconButton>
          <AppIconButton :active="selectMode" :title="selectMode ? t('common.cancelSelect') : t('common.select')" @click="toggleSelectMode">
            <CheckSquare :size="20" />
          </AppIconButton>
          <AppIconButton :title="t('auctions.importLot.add')" @click="showImport = true">
            <CirclePlus :size="20" />
          </AppIconButton>
        </template>
        <div class="page-header">
          <h1>{{ t('nav.auctions') }}</h1>
          <div v-if="isPwa" class="pwa-actions">
            <AppIconButton :disabled="syncing" :title="t('auctions.page.syncWatchlist')" :aria-label="t('auctions.page.syncWatchlist')" @click="syncWatchlist">
              <RefreshCw :size="22" :class="syncing ? 'animate-spin' : ''" />
            </AppIconButton>
            <AppIconButton :active="selectMode" :title="selectMode ? t('common.cancelSelect') : t('common.select')" :aria-label="t('common.select')" @click="toggleSelectMode">
              <CheckSquare :size="22" />
            </AppIconButton>
            <AppIconButton :title="t('auctions.importLot.add')" :aria-label="t('auctions.importLot.add')" @click="showImport = true">
              <CirclePlus :size="22" />
            </AppIconButton>
          </div>
          <div v-else class="header-actions gap-3">
            <button class="btn btn-secondary" :disabled="syncing" @click="syncWatchlist">
              <RefreshCw :size="16" :class="syncing ? 'animate-spin' : ''" />
              {{ syncing ? t('auctions.page.syncing') : t('auctions.page.syncWatchlists') }}
            </button>
            <button class="btn" :class="selectMode ? 'btn-primary' : 'btn-secondary'" @click="toggleSelectMode">
              <CheckSquare :size="16" /> {{ selectMode ? t('common.cancel') : t('common.select') }}
            </button>
            <button class="btn btn-primary" @click="showImport = true"><Plus :size="16" /> {{ t('auctions.importLot.add') }}</button>
          </div>
        </div>
      </DesktopPageContext>

      <div
        v-if="syncMessage"
        class="mb-4 animate-fade-in rounded-sm border border-border-subtle bg-card px-4 py-[0.6rem] text-center text-body text-text-primary"
      >
        {{ syncMessage }}
      </div>

      <div data-testid="auction-filter-row" class="mb-4 flex flex-wrap items-center gap-3">
        <div class="flex min-w-0 flex-1 flex-wrap gap-[0.35rem]" :aria-label="t('auctions.page.sourceFilter')">
          <button
            v-for="source in sourceOptions"
            :key="source.value"
            class="chip"
            :class="{ active: activeSource === source.value }"
            @click="activeSource = source.value"
          >
            {{ source.label }}
          </button>
          <button
            class="chip flex items-center gap-1"
            :class="{ active: attentionOnly }"
            :title="t('auctions.page.attentionTitle')"
            @click="attentionOnly = !attentionOnly"
          >
            <AlertTriangle :size="13" /> {{ t('auctions.page.attention') }}
          </button>
          <template v-if="groupingAvailable">
            <span class="mx-1 h-6 w-px self-center bg-border-subtle" aria-hidden="true"></span>
            <div class="flex" :aria-label="t('auctions.page.grouping')">
              <button
                class="chip"
                :class="{ active: groupingEnabled }"
                :aria-pressed="groupingEnabled"
                @click="groupingEnabled = !groupingEnabled"
              >
                {{ t('auctions.page.groupByAuction') }}
              </button>
            </div>
          </template>
        </div>
        <AuctionStatusFilter v-model="activeStatus" :counts="statusCounts" />
      </div>

      <div v-if="selectMode" class="mb-4 flex flex-wrap items-center gap-[0.6rem]">
        <button class="btn btn-sm btn-secondary" @click="selectAllLots">{{ t('collection.selectAll') }}</button>
        <button class="btn btn-sm btn-secondary" @click="deselectAllLots">{{ t('collection.deselectAll') }}</button>
        <span class="text-body font-medium text-text-secondary">{{ t('auctions.page.selectedCount', { count: selectedLotIds.size }) }}</span>
      </div>

      <div v-if="loading" class="loading-overlay">
        <div class="spinner"></div>
      </div>

      <div v-else-if="visibleLots.length && groupingAvailable && groupingEnabled">
        <div v-for="house in groupedLots" :key="house.auctionHouse" class="mb-5">
          <h3 class="mb-1 mt-4 text-heading">{{ house.auctionHouse }}</h3>
          <div v-for="sale in house.sales" :key="sale.saleName" class="mb-4">
            <p class="section-label mb-2">{{ sale.saleName }}</p>
            <div class="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-5">
              <AuctionLotCard
                v-for="lot in sale.lots"
                :key="lot.id"
                :lot="lot"
                :selectable="selectMode"
                :selected="selectedLotIds.has(lot.id)"
                :price-alerts="alertsByLot[lot.id] ?? []"
                :bid-reminders="remindersByLot[lot.id] ?? []"
                @select="openLot"
                @toggle-select="toggleLotSelect"
              />
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="visibleLots.length" class="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-5">
        <AuctionLotCard
          v-for="lot in visibleLots"
          :key="lot.id"
          :lot="lot"
          :selectable="selectMode"
          :selected="selectedLotIds.has(lot.id)"
          :price-alerts="alertsByLot[lot.id] ?? []"
          :bid-reminders="remindersByLot[lot.id] ?? []"
          @select="openLot"
          @toggle-select="toggleLotSelect"
        />
      </div>

      <div v-else class="empty-state">
        <h3>{{ t('auctions.empty.title', { suffix: emptyStateSuffix }) }}</h3>
        <p>{{ t('auctions.empty.body') }}</p>
        <button class="btn btn-primary mt-3" @click="showImport = true">
          <Plus :size="16" /> {{ t('auctions.empty.importFirst') }}
        </button>
        <SafeExternalLink href="https://www.numisbids.com/" class="btn btn-secondary mt-3 inline-flex items-center gap-1.5 no-underline">
          <ExternalLink :size="16" /> {{ t('auctions.empty.visitNumisBids') }}
        </SafeExternalLink>
        <SafeExternalLink href="https://auctions.cngcoins.com/" class="btn btn-secondary mt-3 inline-flex items-center gap-1.5 no-underline">
          <ExternalLink :size="16" /> {{ t('auctions.empty.visitCng') }}
        </SafeExternalLink>
      </div>

      <ImportLotModal v-if="showImport" @close="showImport = false" @imported="handleImported" />

      <Teleport to="body">
        <AuctionLotDetailModal
          v-if="selectedLot"
          :lot="selectedLot"
          :price-alerts="alertsByLot[selectedLot.id] ?? []"
          :bid-reminders="remindersByLot[selectedLot.id] ?? []"
          @close="closeLot"
          @updated="handleLotUpdated"
          @alerts-updated="fetchAlertState"
        />
      </Teleport>

      <div v-if="bulkLinkMessage" class="card mb-4" role="status">
        <p>{{ bulkLinkMessage }}</p>
        <p v-for="failure in bulkLinkFailures" :key="failure.lotId">{{ t('auctions.bulk.lotFailure', { id: failure.lotId, error: failure.error }) }}</p>
      </div>
      <AuctionBulkActionBar
        v-if="selectMode"
        :selected-count="selectedLotIds.size"
        :calendar-events="calendarEvents"
        @link-event="handleBulkLinkEvent"
      />
    </div>
  </PullToRefresh>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getAuctionLot, getAuctionLots, getAuctionLotCounts, syncNumisBidsWatchlist, listCalendarEvents, bulkLinkAuctionLotEvent, listAlerts, listReminders } from '@/api/client'
import type { AuctionLot, BidReminder, PriceAlert } from '@/types'

import AuctionLotCard from '@/components/AuctionLotCard.vue'
import ImportLotModal from '@/components/ImportLotModal.vue'
import PullToRefresh from '@/components/PullToRefresh.vue'
import AuctionStatusFilter from '@/components/auction/AuctionStatusFilter.vue'
import AuctionLotDetailModal from '@/components/auction/AuctionLotDetailModal.vue'
import AuctionBulkActionBar from '@/components/auction/AuctionBulkActionBar.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import { Plus, CirclePlus, RefreshCw, CheckSquare, ExternalLink, AlertTriangle } from 'lucide-vue-next'
import SafeExternalLink from '@/components/SafeExternalLink.vue'
import { usePwa } from '@/composables/usePwa'
import { useAuthStore } from '@/stores/auth'
import { useQuickAccess } from '@/composables/useQuickAccess'
import { auctionLotNeedsAttention } from '@/utils/auctionLot'
import { auctionStatusLabel } from '@/utils/format'
import { t } from '@/i18n'

const { isPwa } = usePwa()
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const { refresh: refreshQuickAccess } = useQuickAccess()

const lots = ref<AuctionLot[]>([])
const statusCounts = ref<Record<string, number>>({})
const loading = ref(true)
const showImport = ref(false)
const selectedLot = ref<AuctionLot | null>(null)
const activeStatus = ref('bidding')
const activeSource = ref('')
const syncing = ref(false)
const syncMessage = ref('')
const calendarEvents = ref<Array<{ id: number; title: string; auctionHouse: string; startDate: string | null }>>([])
const priceAlerts = ref<PriceAlert[]>([])
const bidReminders = ref<BidReminder[]>([])
let lotRequestId = 0

const selectMode = ref(false)
const selectedLotIds = ref(new Set<number>())
const bulkLinkMessage = ref('')
const bulkLinkFailures = ref<{ lotId: number; error: string }[]>([])
const linkingEvent = ref(false)
const sourceOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'numisbids', label: 'Numis' },
  { value: 'cng', label: 'CNG' },
])
const alertsByLot = computed(() => groupByLot(priceAlerts.value))
const remindersByLot = computed(() => groupByLot(bidReminders.value))

function groupByLot<T extends { auctionLotId: number }>(items: T[]): Record<number, T[]> {
  return items.reduce<Record<number, T[]>>((acc, item) => {
    if (!acc[item.auctionLotId]) acc[item.auctionLotId] = []
    acc[item.auctionLotId]?.push(item)
    return acc
  }, {})
}

function toggleSelectMode() {
  selectMode.value = !selectMode.value
  if (!selectMode.value) {
    selectedLotIds.value = new Set()
  } else {
    fetchCalendarEvents()
  }
}

function toggleLotSelect(lotId: number) {
  const next = new Set(selectedLotIds.value)
  if (next.has(lotId)) next.delete(lotId)
  else next.add(lotId)
  selectedLotIds.value = next
}

function selectAllLots() {
  selectedLotIds.value = new Set(visibleLots.value.map(l => l.id))
}

function deselectAllLots() {
  selectedLotIds.value = new Set()
}

async function fetchCalendarEvents() {
  try {
    const res = await listCalendarEvents()
    calendarEvents.value = res.data?.events ?? []
  } catch { /* ignore */ }
}

async function handleBulkLinkEvent(eventIdRaw: number | string) {
  if (linkingEvent.value) return
  const eventId = eventIdRaw === '' ? null : Number(eventIdRaw)
  linkingEvent.value = true
  bulkLinkMessage.value = t('auctions.bulk.linking')
  bulkLinkFailures.value = []
  try {
    const { data } = await bulkLinkAuctionLotEvent([...selectedLotIds.value], eventId)
    bulkLinkFailures.value = data.failures
    bulkLinkMessage.value = t('auctions.bulk.result', { updated: data.updated, failed: data.failures.length })
    selectedLotIds.value = new Set(data.failures.map(failure => failure.lotId))
    selectMode.value = data.failures.length > 0
    await fetchLots()
  } catch {
    bulkLinkMessage.value = t('auctions.bulk.linkFailed')
  } finally {
    linkingEvent.value = false
  }
}

watch([activeStatus, activeSource], () => {
  selectedLotIds.value = new Set()
  fetchLots()
  fetchAllCounts()
})

async function fetchLots() {
  loading.value = true
  try {
    const params: Record<string, string> = { sort: 'updated_at', order: 'desc' }
    if (activeStatus.value) params.status = activeStatus.value
    if (activeSource.value) params.source = activeSource.value
    const res = await getAuctionLots(params)
    lots.value = res.data?.lots ?? []
  } catch {
    lots.value = []
  } finally {
    loading.value = false
  }
}

async function fetchAllCounts() {
  try {
    const params = activeSource.value ? { source: activeSource.value } : undefined
    const res = await getAuctionLotCounts(params)
    statusCounts.value = res.data?.counts ?? {}
  } catch { /* ignore */ }
}

async function handleRefresh() {
  await Promise.all([fetchLots(), fetchAllCounts(), fetchAlertState()])
}

function openLot(lot: AuctionLot) {
  lotRequestId += 1
  selectedLot.value = lot
}

function closeLot() {
  lotRequestId += 1
  selectedLot.value = null
  // Drop the deep-link parameter so the lot does not reopen on the next navigation back here.
  if (route.query.lot !== undefined) {
    const { lot: _removed, ...rest } = route.query
    router.replace({ query: rest })
  }
}

// Deep-link support for auction notifications: both the in-app notification and the Pushover
// push for newly tracked lots link to /auctions?lot=<id>, so the lot must open on arrival. The
// lot is always fetched by id rather than looked up in the loaded list, because the list is
// filtered (Bidding by default) and usually will not contain the linked lot.
async function openLotFromRoute(rawLotId: unknown) {
  const lotId = Number(rawLotId)
  const requestId = ++lotRequestId
  selectedLot.value = null
  if (!Number.isInteger(lotId) || lotId <= 0) return
  try {
    const res = await getAuctionLot(lotId)
    if (requestId === lotRequestId && res.data) selectedLot.value = res.data
  } catch {
    // Lot was deleted, or belongs to someone else — leave the page on the plain list.
  }
}

watch(() => route.query.lot, lotId => { openLotFromRoute(lotId) })

function handleImported() {
  showImport.value = false
  fetchLots()
  fetchAllCounts()
}

function handleLotUpdated() {
  fetchLots()
  fetchAllCounts()
  void refreshQuickAccess()
}

async function fetchAlertState() {
  try {
    const [alertsRes, remindersRes] = await Promise.all([listAlerts(), listReminders()])
    priceAlerts.value = alertsRes.data?.alerts ?? []
    bidReminders.value = remindersRes.data?.reminders ?? []
  } catch {
    priceAlerts.value = []
    bidReminders.value = []
  }
}

async function syncWatchlist() {
  syncing.value = true
  syncMessage.value = ''
  try {
    const providers = configuredAuctionProviders()
    if (!providers.length) {
      syncMessage.value = t('auctions.sync.configureFirst')
      setTimeout(() => { syncMessage.value = '' }, 5000)
      return
    }
    const results = await Promise.allSettled(providers.map((source) => syncNumisBidsWatchlist(source)))
    await refreshQuickAccess()
    const synced = results.reduce((total, result) => total + (result.status === 'fulfilled' ? result.value.data?.synced ?? 0 : 0), 0)
    const failed = results.filter((result) => result.status === 'rejected')
    const providerLabel = providers.length > 1 ? t('auctions.sync.watchlists') : providerName(providers[0] ?? 'numisbids')
    const syncedLots = t('auctions.sync.lots', { count: synced })
    syncMessage.value = failed.length
      ? t('auctions.sync.partial', { lots: syncedLots, providers: t('auctions.sync.providersFailed', { count: failed.length }) })
      : t('auctions.sync.done', { lots: syncedLots, source: providerLabel, note: providerCapabilityNote(providers) })
    fetchLots()
    fetchAllCounts()
    setTimeout(() => { syncMessage.value = '' }, 4000)
  } catch (err: unknown) {
    const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? t('auctions.sync.failed')
    syncMessage.value = msg
    setTimeout(() => { syncMessage.value = '' }, 5000)
  } finally {
    syncing.value = false
  }
}

function configuredAuctionProviders(): string[] {
  const providers: string[] = []
  if (auth.user?.numisBidsConfigured) providers.push('numisbids')
  if (auth.user?.cngConfigured) providers.push('cng')
  return providers
}

function providerName(source: string): string {
  return source === 'cng' ? 'CNG Auctions' : 'NumisBids'
}

function providerCapabilityNote(providers: string[]): string {
  if (providers.length !== 1) return ''
  return providers[0] === 'cng'
    ? t('auctions.sync.cngNote')
    : t('auctions.sync.numisNote')
}

const emptyStateSuffix = computed(() => {
  const parts: string[] = []
  if (activeStatus.value) parts.push(t('auctions.empty.status', { status: auctionStatusLabel(activeStatus.value).toLowerCase() }))
  if (activeSource.value) parts.push(providerName(activeSource.value))
  if (attentionOnly.value) parts.push(t('auctions.empty.needingAttention'))
  return parts.length ? t('auctions.empty.matching', { parts: parts.join(t('auctions.empty.and')) }) : ''
})

const attentionOnly = ref(false)
const visibleLots = computed(() => attentionOnly.value ? lots.value.filter(auctionLotNeedsAttention) : lots.value)

// Watching and Bidding are the two statuses most useful to browse by Auction House -> Sale,
// since those lots are still active and users track them per-sale. Defaults on for those two
// views; the toggle below lets a visitor turn it off, and that choice is preserved for the
// rest of the visit even if they switch between statuses/sources.
const groupingAvailable = computed(() => activeStatus.value === 'watching' || activeStatus.value === 'bidding')
const groupingEnabled = ref(true)

interface AuctionLotSaleGroup {
  saleName: string
  lots: AuctionLot[]
}

interface AuctionLotHouseGroup {
  auctionHouse: string
  sales: AuctionLotSaleGroup[]
}

const groupedLots = computed<AuctionLotHouseGroup[]>(() => {
  const byHouse = new Map<string, Map<string, AuctionLot[]>>()
  for (const lot of visibleLots.value) {
    const house = lot.auctionHouse || t('auctions.unknownHouse')
    const sale = lot.saleName || t('auctions.unknownSale')
    if (!byHouse.has(house)) byHouse.set(house, new Map())
    const sales = byHouse.get(house)!
    if (!sales.has(sale)) sales.set(sale, [])
    sales.get(sale)!.push(lot)
  }
  return [...byHouse.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([auctionHouse, sales]) => ({
      auctionHouse,
      sales: [...sales.entries()]
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([saleName, saleLots]) => ({ saleName, lots: saleLots })),
    }))
})

fetchLots()
fetchAllCounts()
fetchAlertState()
openLotFromRoute(route.query.lot)
</script>
