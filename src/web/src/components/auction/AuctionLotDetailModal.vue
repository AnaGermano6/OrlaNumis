<template>
  <div class="fixed inset-0 z-[1000] flex overflow-x-hidden overflow-y-auto bg-overlay-full p-4 [-webkit-overflow-scrolling:touch]" @click.self="$emit('close')">
    <div class="card mx-auto my-auto w-full min-w-0 max-w-[560px] p-0">
      <div class="flex items-start justify-between gap-4 border-b border-border-subtle px-6 py-5">
        <h2 class="min-w-0 text-md leading-[1.35] [overflow-wrap:anywhere]">{{ lot.title }}</h2>
        <div class="header-actions shrink-0 gap-1">
          <button
            v-if="canPin"
            class="flex h-11 w-11 items-center justify-center rounded-sm text-text-secondary transition hover:bg-gold-glow hover:text-gold disabled:opacity-55"
            :class="{ 'text-gold': lotPinned }"
            :disabled="pinBusy"
            :aria-busy="pinBusy"
            :aria-label="pinBusy ? t('auctions.detail.pinUpdating') : pinLabel"
            :title="pinLabel"
            :aria-pressed="lotPinned"
            @click="togglePin"
          >
            <PinOff v-if="lotPinned" :size="16" />
            <Pin v-else :size="16" />
          </button>
          <button v-if="!isEditing" class="flex items-center rounded-sm p-[0.35rem] text-text-secondary transition hover:bg-gold-glow hover:text-gold" :title="t('auctions.detail.editDetails')" @click="startEdit">
            <Pencil :size="16" />
          </button>
          <button class="flex items-center rounded-sm p-1 text-text-secondary transition hover:text-text-primary" @click="$emit('close')"><X :size="18" /></button>
        </div>
      </div>

      <div v-if="proxiedImageUrl" class="flex max-h-[300px] w-full items-center justify-center overflow-hidden bg-surface">
        <img :src="proxiedImageUrl" :alt="lot.title" class="h-[300px] w-full object-contain" />
      </div>

      <div v-if="!isEditing" class="px-6 py-5">
        <div v-if="lot.auctionHouse" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ t('auctions.detail.auctionHouse') }}</span>
          <span class="min-w-0 text-right [overflow-wrap:anywhere]">{{ lot.auctionHouse }}</span>
        </div>
        <div v-if="lot.saleName" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ t('auctions.detail.sale') }}</span>
          <span class="min-w-0 text-right [overflow-wrap:anywhere]">{{ lot.saleName }}</span>
        </div>
        <div v-if="lot.lotNumber" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ t('auctions.detail.lotNumber') }}</span>
          <span class="min-w-0 text-right [overflow-wrap:anywhere]">{{ lot.lotNumber }}</span>
        </div>
        <div v-if="lot.saleDate" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ t('auctions.detail.saleDate') }}</span>
          <span class="min-w-0 text-right [overflow-wrap:anywhere]">{{ formatDate(lot.saleDate) }}</span>
        </div>
        <div v-if="lot.auctionEndTime" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ t('auctions.detail.ends') }}</span>
          <span class="min-w-0 text-right [overflow-wrap:anywhere]">{{ formatDateTime(lot.auctionEndTime) }}</span>
        </div>
        <div v-if="lot.estimate" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ t('auctions.detail.estimate') }}</span>
          <span class="min-w-0 text-right [overflow-wrap:anywhere]">{{ formatCurrency(lot.estimate, lot.currency) }}</span>
        </div>
        <div v-if="lot.initialBid" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ t('auctions.detail.initialBid') }}</span>
          <span class="min-w-0 text-right [overflow-wrap:anywhere]">{{ formatCurrency(lot.initialBid, lot.currency) }}</span>
        </div>
        <div v-if="lot.currentBid" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ t('auctions.detail.currentBid') }}</span>
          <span class="min-w-0 text-right font-semibold text-gold [overflow-wrap:anywhere]">{{ formatCurrency(lot.currentBid, lot.currency) }}</span>
        </div>
        <div v-if="lot.maxBid" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ t('auctions.card.maxBid') }}</span>
          <span class="min-w-0 text-right font-semibold text-gold/80 [overflow-wrap:anywhere]">{{ formatCurrency(lot.maxBid, lot.currency) }}</span>
        </div>
        <div v-if="lot.winningBid" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ lot.status === 'lost' ? t('auctions.card.soldFor') : t('auctions.detail.winningBid') }}</span>
          <span class="min-w-0 text-right font-semibold [overflow-wrap:anywhere]" :class="lot.status === 'lost' ? 'text-text-primary' : 'text-gain'">{{ formatCurrency(lot.winningBid, lot.currency) }}</span>
        </div>
        <div v-if="biddingIndicator" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ t('auctions.detail.bidStatus') }}</span>
          <span class="min-w-0 text-right font-semibold [overflow-wrap:anywhere]" :class="biddingIndicator.cls">{{ biddingIndicator.label }}</span>
        </div>
        <div class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-base">
          <span class="text-chip text-text-secondary">{{ t('common.status') }}</span>
          <span
            class="rounded-full px-[0.55rem] py-[0.15rem] text-label font-semibold tracking-[0.08em] uppercase"
            :class="{
              'bg-status-info-bg text-status-info-fg': currentStatus === 'watching',
              'bg-gold-glow text-gold': currentStatus === 'bidding',
              'bg-status-success-bg text-status-success-fg': currentStatus === 'won',
              'bg-status-error-bg text-status-error-fg': currentStatus === 'lost',
              'bg-status-neutral-bg text-status-neutral-fg': currentStatus === 'passed',
            }"
          >
            {{ auctionStatusLabel(currentStatus).toLowerCase() }}
          </span>
        </div>
        <div v-if="statusSourceLabel" class="flex min-w-0 items-center justify-between gap-3 border-b border-border-subtle py-2 text-chip text-text-muted" :title="statusSourceLabel.title">
          <span class="text-chip text-text-secondary">{{ t('auctions.detail.confirmedBy') }}</span>
          <span>{{ statusSourceLabel.text }}</span>
        </div>
        <div v-if="needsAttention" class="mt-2 flex items-center gap-1.5 rounded-sm bg-status-warning-bg px-2.5 py-1.5 text-chip font-semibold text-status-warning-fg">
          <AlertTriangle :size="14" /> {{ t('auctions.card.closedUnconfirmed') }}
        </div>
        <div v-if="lot.description" class="mt-3">
          <span class="text-chip text-text-secondary">{{ t('common.description') }}</span>
          <p class="mt-1.5 text-body leading-6 text-text-secondary [overflow-wrap:anywhere]">{{ lot.description }}</p>
        </div>
        <div v-if="lot.notes" class="mt-3">
          <span class="text-chip text-text-secondary">{{ t('coinFields.notes') }}</span>
          <p class="mt-1.5 text-body leading-6 text-text-secondary [overflow-wrap:anywhere]">{{ lot.notes }}</p>
        </div>

        <section v-if="canManageAlerts" class="mt-4 grid gap-3 border-t border-border-subtle pt-4">
          <div class="grid gap-2">
            <div class="flex items-center justify-between gap-2">
              <span class="text-chip text-text-secondary">{{ t('auctions.detail.priceAlerts') }}</span>
              <span v-if="priceAlerts.length" class="chip-sm">{{ priceAlerts.length }}</span>
            </div>
            <div v-if="priceAlerts.length" class="grid gap-[0.35rem]">
              <div v-for="alert in priceAlerts" :key="alert.id" class="flex items-center justify-between gap-2 rounded-sm border border-border-subtle bg-card-hover p-2">
                <div class="min-w-0">
                  <span class="block text-chip font-medium text-text-primary">{{ alert.direction === 'above' ? t('auctions.detail.atOrAbove') : t('auctions.detail.atOrBelow') }} {{ formatCurrency(alert.targetPrice, lot.currency) }}</span>
                  <span class="mt-[0.15rem] block text-chip text-text-muted">{{ alert.isTriggered ? t('auctions.detail.triggered', { date: formatOptionalDate(alert.triggeredAt) }) : t('common.waiting') }}</span>
                </div>
                <button class="btn btn-ghost btn-xs" :disabled="alertBusy" @click="removeAlert(alert.id)">{{ t('common.delete') }}</button>
              </div>
            </div>
            <div class="control-row">
              <div class="control-left-group control-left-group--alert">
                <select v-model="alertForm.direction" class="form-input control-input" :aria-label="t('auctions.detail.alertDirection')">
                  <option value="above">{{ t('auctions.detail.aboveCurrent') }}</option>
                  <option value="below">{{ t('auctions.detail.belowCurrent') }}</option>
                </select>
                <input
                  v-model.number="alertForm.targetPrice"
                  type="number"
                  class="form-input control-input control-number"
                  min="0"
                  step="0.01"
                  :placeholder="lot.currentBid ? String(lot.currentBid) : t('auctions.detail.target')"
                  :aria-label="t('auctions.detail.targetPrice')"
                />
              </div>
              <button class="btn btn-secondary btn-sm control-btn control-btn-fixed" :disabled="alertBusy || !canCreateAlert" @click="saveAlert">{{ t('auctions.detail.addAlert') }}</button>
            </div>
          </div>

          <div class="grid gap-2">
            <div class="flex items-center justify-between gap-2">
              <span class="text-chip text-text-secondary">{{ t('auctions.detail.bidReminders') }}</span>
              <span v-if="bidReminders.length" class="chip-sm">{{ bidReminders.length }}</span>
            </div>
            <div v-if="bidReminders.length" class="grid gap-[0.35rem]">
              <div v-for="reminder in bidReminders" :key="reminder.id" class="flex items-center justify-between gap-2 rounded-sm border border-border-subtle bg-card-hover p-2">
                <div class="min-w-0">
                  <span class="block text-chip font-medium text-text-primary">{{ t('auctions.detail.minutesBefore', { count: reminder.minutesBefore }) }}</span>
                  <span class="mt-[0.15rem] block text-chip text-text-muted">{{ reminder.isNotified ? t('auctions.detail.notified', { date: formatOptionalDate(reminder.notifiedAt) }) : t('common.waiting') }}</span>
                </div>
                <button class="btn btn-ghost btn-xs" :disabled="reminderBusy" @click="removeReminder(reminder.id)">{{ t('common.delete') }}</button>
              </div>
            </div>
            <div class="control-row">
              <input
                v-model.number="reminderForm.minutesBefore"
                type="number"
                class="form-input control-input control-number"
                min="1"
                step="5"
                :aria-label="t('auctions.detail.reminderMinutes')"
              />
              <button class="btn btn-secondary btn-sm control-btn control-btn-fixed" :disabled="reminderBusy || !canCreateReminder" @click="saveReminder">{{ t('auctions.detail.addReminder') }}</button>
            </div>
          </div>
          <p v-if="alertMessage" class="m-0 text-chip" :class="alertError ? 'text-loss' : 'text-gold'">{{ alertMessage }}</p>
        </section>
      </div>

      <div v-else class="flex flex-col gap-[0.85rem] px-6 py-5">
        <div class="form-group">
          <label class="text-chip text-text-secondary">{{ t('common.title') }}</label>
          <input v-model="editForm.title" type="text" class="form-input w-full text-base" />
        </div>
        <div class="form-group">
          <label class="text-chip text-text-secondary">{{ t('auctions.detail.auctionUrl') }}</label>
          <input v-model="editForm.numisBidsUrl" type="url" class="form-input w-full text-base" placeholder="https://..." />
        </div>
        <div class="grid gap-[0.85rem] md:grid-cols-2">
          <div class="form-group">
            <label class="text-chip text-text-secondary">{{ t('auctions.detail.auctionHouse') }}</label>
            <input v-model="editForm.auctionHouse" type="text" class="form-input w-full text-base" />
          </div>
          <div class="form-group">
            <label class="text-chip text-text-secondary">{{ t('auctions.detail.saleName') }}</label>
            <input v-model="editForm.saleName" type="text" class="form-input w-full text-base" />
          </div>
        </div>
        <div class="grid gap-[0.85rem] md:grid-cols-2">
          <div class="form-group">
            <label class="text-chip text-text-secondary">{{ t('auctions.detail.lotNumber') }}</label>
            <input v-model.number="editForm.lotNumber" type="number" class="form-input w-full text-base" min="0" />
          </div>
          <div class="form-group">
            <label class="text-chip text-text-secondary">{{ t('auctions.detail.estimate') }}</label>
            <input v-model.number="editForm.estimate" type="number" class="form-input w-full text-base" min="0" step="0.01" />
          </div>
        </div>
        <div class="grid gap-[0.85rem] md:grid-cols-2">
          <div class="form-group">
            <label class="text-chip text-text-secondary">{{ t('auctions.detail.saleDate') }}</label>
            <input v-model="editForm.saleDate" type="date" class="form-input w-full text-base" />
          </div>
          <div class="form-group">
            <label class="text-chip text-text-secondary">{{ t('auctions.detail.endDateTime') }}</label>
            <input v-model="editForm.auctionEndTime" type="datetime-local" class="form-input w-full text-base" />
          </div>
        </div>
        <div class="form-group">
          <label class="text-chip text-text-secondary">{{ t('common.description') }}</label>
          <textarea v-model="editForm.description" class="form-input w-full resize-y text-base leading-[1.45]" rows="3" />
        </div>
        <div class="form-group">
          <label class="text-chip text-text-secondary">{{ t('coinFields.notes') }}</label>
          <textarea
            v-model="editForm.notes"
            class="form-input w-full resize-y text-base leading-[1.45]"
            rows="4"
            :placeholder="t('auctions.detail.notesPlaceholder')"
          />
        </div>
        <p v-if="editError" class="m-0 text-chip text-loss">{{ editError }}</p>
        <div class="mt-2 flex justify-end gap-2">
          <button class="btn btn-secondary" :disabled="editSaving" @click="cancelEdit">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="editSaving" @click="saveEdit">
            {{ editSaving ? t('common.saving') : t('common.saveChanges') }}
          </button>
        </div>
      </div>

      <div v-if="!isEditing" class="flex flex-col gap-3 border-t border-border-subtle px-6 py-4">
        <div class="control-row">
          <select v-model="newStatus" class="form-input control-input">
            <option value="watching">{{ auctionStatusLabel('watching') }}</option>
            <option value="bidding">{{ auctionStatusLabel('bidding') }}</option>
            <option value="won">{{ auctionStatusLabel('won') }}</option>
            <option value="lost">{{ auctionStatusLabel('lost') }}</option>
            <option value="passed">{{ auctionStatusLabel('passed') }}</option>
          </select>
          <button class="btn btn-secondary control-btn control-btn-fixed" @click="changeStatus" :disabled="statusBusy || !hasPendingStatusUpdate">
            {{ statusBusy ? t('common.updating') : t('auctions.detail.updateStatus') }}
          </button>
        </div>
        <p
          v-if="statusMessage"
          class="m-0 text-chip"
          :class="statusError ? 'text-loss' : 'text-gold'"
          :role="statusError ? 'alert' : 'status'"
          :aria-live="statusError ? 'assertive' : 'polite'"
          aria-atomic="true"
        >
          {{ statusMessage }}
        </p>
        <div v-if="newStatus === 'bidding'" class="grid gap-1.5">
          <label class="text-chip text-text-secondary">{{ t('auctions.card.maxBid') }}</label>
          <input
            v-model.number="maxBidInput"
            type="number"
            class="form-input bid-input control-input control-number"
            :aria-label="t('auctions.detail.maxBidAria')"
            :placeholder="lot.currency || 'USD'"
            min="0"
            step="1"
          />
        </div>
        <div v-if="newStatus === 'bidding'" class="text-chip text-text-secondary">
          <template v-if="bidRecommendation?.suggestedMaxBid">
            <button
              type="button"
              class="text-gold underline decoration-dotted underline-offset-2"
              :title="bidRecommendation.rationale"
              @click="maxBidInput = bidRecommendation.suggestedMaxBid"
            >
              {{ t('auctions.detail.suggested', { amount: formatCurrency(bidRecommendation.suggestedMaxBid, lot.currency) }) }}
            </button>
            <span class="text-text-muted"> ({{ t('auctions.detail.suggestedBasis', { level: confidenceLabel(bidRecommendation.confidence), count: bidRecommendation.sampleSize }) }})</span>
          </template>
          <span v-else-if="bidRecommendation" class="text-text-muted">{{ bidRecommendation.rationale }}</span>
          <span v-else-if="bidRecommendationError" class="text-text-muted">{{ t('auctions.detail.suggestionFailed') }}</span>
        </div>
        <div v-if="newStatus === 'bidding'" class="text-chip text-text-secondary">
          <span v-if="!providerConfigured" class="text-warning">
            <I18nT keypath="auctions.detail.providerNotConfigured">
              <template #link><a href="/admin" class="font-semibold text-gold underline" @click="$emit('close')">{{ t('auctions.detail.goToAdmin') }}</a></template>
            </I18nT>
          </span>
          <template v-else-if="marketSignalLoading">
            <span class="text-text-muted">{{ t('auctions.detail.checkingMarket') }}</span>
          </template>
          <template v-else-if="marketSignal">
            <template v-if="marketSignal.status === 'ok'">
              <span class="font-medium text-text-primary">{{ t('auctions.detail.marketTrend', { trend: trendLabel(marketSignal.trendDirection) }) }}</span>
              <span v-if="marketSignal.priceLow != null && marketSignal.priceHigh != null" class="text-text-muted">
                ({{ formatCurrency(marketSignal.priceLow, marketSignal.currency || lot.currency) }}–{{ formatCurrency(marketSignal.priceHigh, marketSignal.currency || lot.currency) }},
                {{ t('auctions.detail.basedOnRecent', { count: marketSignal.sampleSize ?? 0 }) }}
              </span>
              <span :title="marketSignal.rationale" class="ml-1 cursor-help text-text-muted underline decoration-dotted underline-offset-2">{{ t('auctions.detail.why') }}</span>
            </template>
            <span v-else class="text-text-muted">{{ marketSignal.rationale }}</span>
          </template>
          <template v-else>
            <button
              type="button"
              class="text-gold underline decoration-dotted underline-offset-2"
              @click="fetchMarketSignal"
            >
              {{ marketSignalError ? t('common.tryAgain') : t('auctions.detail.checkMarket') }}
            </button>
            <span v-if="marketSignalError" class="ml-1 text-text-muted">{{ t('auctions.detail.marketFailed') }}</span>
          </template>
        </div>
        <div v-if="newStatus === 'won' || newStatus === 'lost'" class="grid gap-1.5">
          <label class="text-chip text-text-secondary">{{ winningBidLabel }}</label>
          <input
            v-model.number="winningBidInput"
            type="number"
            class="form-input winning-bid-input control-input control-number"
            :aria-label="winningBidLabel"
            :placeholder="lot.currency || 'USD'"
            min="0"
            step="0.01"
          />
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="inline-flex items-center gap-[0.35rem] text-chip text-text-secondary"><CalendarDays :size="14" /> {{ t('auctions.detail.calendarEvent') }}</label>
          <div class="control-row">
            <select v-model="selectedEventId" class="form-input control-input">
              <option value="">{{ t('common.none') }}</option>
              <option v-for="evt in calendarEvents" :key="evt.id" :value="evt.id">
                {{ evt.title }}
              </option>
            </select>
            <button
              class="btn btn-secondary btn-sm control-btn control-btn-fixed"
              @click="linkEvent"
              :disabled="(selectedEventId === '' ? null : Number(selectedEventId)) === (lot.eventId ?? null)"
            >
              {{ t('auctions.bulk.link') }}
            </button>
          </div>
        </div>
        <div v-if="lot.status === 'won'" class="rounded-sm border border-border-subtle bg-card-hover p-3">
          <label class="mb-2 inline-flex items-center gap-2 text-chip text-text-secondary">
            <input v-model="attachShipment" type="checkbox" />
            {{ t('auctions.detail.attachShipment') }}
          </label>
          <div v-if="attachShipment" class="grid gap-2 md:grid-cols-2">
            <p class="m-0 text-chip text-text-muted md:col-span-2">{{ t('purchase.shipmentHint') }}</p>
            <input v-model="shipmentTrackingNumber" class="form-input" :placeholder="t('shipment.trackingNumber')" />
            <input v-model="shipmentNotes" class="form-input md:col-span-2" :placeholder="t('purchase.shipmentNotes')" />
          </div>
        </div>
        <div class="flex flex-wrap gap-[0.6rem]">
          <SafeExternalLink v-if="externalUrl" :href="externalUrl" class="btn btn-primary" target="_blank" rel="noopener noreferrer">
            <ExternalLink :size="14" /> {{ t('auctions.card.viewOn', { provider: providerLabel }) }}
          </SafeExternalLink>
          <button v-if="lot.status === 'won'" class="btn btn-primary" :disabled="statusBusy" @click="convertToCoin">
            <ArrowRightCircle :size="14" /> {{ t('addCoin.addToCollection') }}
          </button>
          <button class="btn btn-danger !border-status-error-border !bg-transparent px-[0.9rem] py-2 text-chip !text-status-error-fg hover:!border-status-error-fg hover:!bg-status-error-tint" :disabled="statusBusy" @click="removeLot">
            <Trash2 :size="14" /> {{ t('common.remove') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { updateAuctionLotStatus, updateAuctionLot, convertAuctionLotToCoin, deleteAuctionLot, listCalendarEvents, linkAuctionLotEvent, createAlert, deleteAlert, createReminder, deleteReminder, getAuctionLotBidRecommendation, getAuctionLotMarketSignal, getAgentStatus } from '@/api/client'
import { useProxiedImage } from '@/composables/useProxiedImage'
import { useQuickAccess } from '@/composables/useQuickAccess'
import type { AuctionLot, AuctionLotStatus, BidReminder, BidRecommendation, MarketSignal, PriceAlert, PriceAlertDirection, ShipmentUpsertInput } from '@/types'
import { X, ExternalLink, ArrowRightCircle, Trash2, CalendarDays, Pencil, AlertTriangle, Pin, PinOff } from 'lucide-vue-next'
import { auctionStatusLabel, confidenceLabel, formatCurrency } from '@/utils/format'
import { t, intlLocale } from '@/i18n'
import I18nT from '@/components/I18nT.vue'
import { auctionLotNeedsAttention, auctionLotStatusSourceLabel } from '@/utils/auctionLot'
import SafeExternalLink from '@/components/SafeExternalLink.vue'

const props = defineProps<{
  lot: AuctionLot
  priceAlerts?: PriceAlert[]
  bidReminders?: BidReminder[]
}>()

const emit = defineEmits<{
  close: []
  updated: []
  alertsUpdated: []
}>()

const router = useRouter()
const {
  error: quickAccessError,
  refresh: refreshQuickAccess,
  pin: pinQuickAccessItem,
  unpin: unpinQuickAccessItem,
  forget: forgetQuickAccess,
  isPinned,
  isBusy,
} = useQuickAccess()

const newStatus = ref<AuctionLotStatus>(props.lot.status)
const currentStatus = ref<AuctionLotStatus>(props.lot.status)
const maxBidInput = ref<number | null>(props.lot.maxBid ?? null)
const winningBidInput = ref<number | null>(props.lot.winningBid ?? null)
const calendarEvents = ref<Array<{ id: number; title: string; auctionHouse: string; startDate: string | null }>>([])
const selectedEventId = ref<number | string>(props.lot.eventId ?? '')
const bidRecommendation = ref<BidRecommendation | null>(null)
const bidRecommendationError = ref(false)
const marketSignal = ref<MarketSignal | null>(null)
const marketSignalLoading = ref(false)
const marketSignalError = ref(false)
const providerConfigured = ref(true)

const lotImageSource = computed(() => props.lot.imageUrl ?? '')
const { proxiedImageUrl } = useProxiedImage(lotImageSource)
const providerLabel = computed(() => props.lot.source === 'cng' ? 'CNG' : 'NumisBids')
const externalUrl = computed(() => props.lot.sourceUrl || props.lot.numisBidsUrl)
const normalizedMaxBidInput = computed(() => typeof maxBidInput.value === 'number' && !Number.isNaN(maxBidInput.value) ? maxBidInput.value : null)
const normalizedWinningBidInput = computed(() => typeof winningBidInput.value === 'number' && !Number.isNaN(winningBidInput.value) ? winningBidInput.value : null)
const maxBidChanged = computed(() => newStatus.value === 'bidding' && normalizedMaxBidInput.value !== null && normalizedMaxBidInput.value !== (props.lot.maxBid ?? null))
const winningBidChanged = computed(() => (newStatus.value === 'won' || newStatus.value === 'lost') && normalizedWinningBidInput.value !== null && normalizedWinningBidInput.value !== (props.lot.winningBid ?? null))
const winningBidLabel = computed(() => newStatus.value === 'lost' ? t('auctions.detail.winningBidLost') : t('auctions.detail.winningBid'))
const hasPendingStatusUpdate = computed(() => newStatus.value !== currentStatus.value || maxBidChanged.value || winningBidChanged.value)
const priceAlerts = computed(() => props.priceAlerts ?? [])
const bidReminders = computed(() => props.bidReminders ?? [])
const canManageAlerts = computed(() => currentStatus.value === 'watching' || currentStatus.value === 'bidding')
const canPin = computed(() => currentStatus.value === 'watching' || currentStatus.value === 'bidding')
const pinnedState = isPinned('auction_lot', props.lot.id)
const busyState = isBusy('auction_lot', props.lot.id)
const lotPinned = computed(() => pinnedState.value)
const pinBusy = computed(() => busyState.value)
const pinLabel = computed(() => lotPinned.value ? t('auctions.detail.unpin') : t('auctions.detail.pin'))
const biddingIndicator = computed(() => {
  if (props.lot.status !== 'bidding' || !props.lot.currentBid || !props.lot.maxBid) return null
  if (props.lot.maxBid >= props.lot.currentBid) {
    return { label: t('auctions.card.winning'), cls: 'text-gain' }
  }
  return { label: t('auctions.card.outbid'), cls: 'text-loss' }
})
const needsAttention = computed(() => auctionLotNeedsAttention(props.lot))
const statusSourceLabel = computed(() => auctionLotStatusSourceLabel(props.lot))
const alertBusy = ref(false)
const reminderBusy = ref(false)
const alertMessage = ref('')
const alertError = ref(false)
const statusBusy = ref(false)
const statusMessage = ref('')
const statusError = ref(false)
const alertForm = reactive<{ targetPrice: number | null; direction: PriceAlertDirection }>({
  targetPrice: props.lot.currentBid ?? props.lot.maxBid ?? props.lot.estimate ?? null,
  direction: 'above',
})
const reminderForm = reactive<{ minutesBefore: number | null }>({
  minutesBefore: 30,
})
const canCreateAlert = computed(() => typeof alertForm.targetPrice === 'number' && alertForm.targetPrice > 0)
const canCreateReminder = computed(() => typeof reminderForm.minutesBefore === 'number' && reminderForm.minutesBefore > 0)
const attachShipment = ref(false)
const shipmentTrackingNumber = ref('')
const shipmentNotes = ref('')

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString(intlLocale(), { year: 'numeric', month: 'short', day: 'numeric' })
}

// Market trend values come from the API as English words; unknown ones pass through.
function trendLabel(trend: string | undefined) {
  switch (trend) {
    case 'rising': return t('auctions.detail.trends.rising')
    case 'falling': return t('auctions.detail.trends.falling')
    case 'stable': return t('auctions.detail.trends.stable')
    default: return trend ?? ''
  }
}

function formatDateTime(dateStr: string) {
  return new Date(dateStr).toLocaleString(intlLocale(), {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

function formatOptionalDate(dateStr: string | null) {
  if (!dateStr) return ''
  return formatDate(dateStr)
}

function setAlertMessage(message: string, isError = false) {
  alertMessage.value = message
  alertError.value = isError
}

function setStatusMessage(message: string, isError = false) {
  statusMessage.value = message
  statusError.value = isError
}

function errorMessageFrom(err: unknown, fallback: string): string {
  const maybeAxiosError = err as { response?: { data?: { error?: string } } }
  return maybeAxiosError?.response?.data?.error || fallback
}

async function togglePin() {
  if (pinBusy.value) return
  setStatusMessage('')
  try {
    if (lotPinned.value) await unpinQuickAccessItem('auction_lot', props.lot.id)
    else await pinQuickAccessItem('auction_lot', props.lot.id)
  } catch {
    setStatusMessage(quickAccessError.value || t('auctions.detail.pinFailed'), true)
  }
}

// Edit mode
const isEditing = ref(false)
const editSaving = ref(false)
const editError = ref('')

interface EditForm {
  title: string
  numisBidsUrl: string
  auctionHouse: string
  saleName: string
  lotNumber: number | null
  saleDate: string
  auctionEndTime: string
  description: string
  notes: string
  estimate: number | null
}

const editForm = reactive<EditForm>({
  title: '',
  numisBidsUrl: '',
  auctionHouse: '',
  saleName: '',
  lotNumber: null,
  saleDate: '',
  auctionEndTime: '',
  description: '',
  notes: '',
  estimate: null,
})

function isoToDateInput(iso: string | null): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  const yyyy = d.getFullYear()
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}`
}

function isoToDateTimeLocalInput(iso: string | null): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  const yyyy = d.getFullYear()
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  const hh = String(d.getHours()).padStart(2, '0')
  const mi = String(d.getMinutes()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}T${hh}:${mi}`
}

function startEdit() {
  editError.value = ''
  editForm.title = props.lot.title || ''
  editForm.numisBidsUrl = externalUrl.value || ''
  editForm.auctionHouse = props.lot.auctionHouse || ''
  editForm.saleName = props.lot.saleName || ''
  editForm.lotNumber = props.lot.lotNumber || null
  editForm.saleDate = isoToDateInput(props.lot.saleDate)
  editForm.auctionEndTime = isoToDateTimeLocalInput(props.lot.auctionEndTime)
  editForm.description = props.lot.description || ''
  editForm.notes = props.lot.notes || ''
  editForm.estimate = props.lot.estimate
  isEditing.value = true
}

function cancelEdit() {
  isEditing.value = false
  editError.value = ''
}

async function saveEdit() {
  editError.value = ''
  const title = editForm.title.trim()
  const url = editForm.numisBidsUrl.trim()
  if (!title) {
    editError.value = t('auctions.detail.titleRequired')
    return
  }
  if (!url) {
    editError.value = t('auctions.detail.urlRequired')
    return
  }
  if (!/^https?:\/\//i.test(url)) {
    editError.value = t('auctions.detail.urlScheme')
    return
  }

  editSaving.value = true
  try {
    await updateAuctionLot(props.lot.id, {
      title,
      numisBidsUrl: url,
      auctionHouse: editForm.auctionHouse.trim(),
      saleName: editForm.saleName.trim(),
      lotNumber: editForm.lotNumber ?? 0,
      saleDate: editForm.saleDate ? new Date(editForm.saleDate).toISOString() : null,
      auctionEndTime: editForm.auctionEndTime ? new Date(editForm.auctionEndTime).toISOString() : null,
      description: editForm.description,
      notes: editForm.notes,
      estimate: editForm.estimate,
    })
    await refreshQuickAccess()
    isEditing.value = false
    emit('updated')
  } catch {
    editError.value = t('common.saveChangesFailed')
  } finally {
    editSaving.value = false
  }
}

async function fetchCalendarEvents() {
  try {
    const res = await listCalendarEvents()
    calendarEvents.value = res.data?.events ?? []
  } catch { /* ignore */ }
}

let bidRecommendationFetched = false
async function fetchBidRecommendation() {
  if (bidRecommendationFetched) return
  bidRecommendationFetched = true
  bidRecommendationError.value = false
  try {
    const res = await getAuctionLotBidRecommendation(props.lot.id)
    bidRecommendation.value = res.data
  } catch {
    bidRecommendationError.value = true
  }
}

async function fetchMarketSignal() {
  marketSignalLoading.value = true
  marketSignalError.value = false
  try {
    const res = await getAuctionLotMarketSignal(props.lot.id)
    marketSignal.value = res.data
  } catch {
    marketSignalError.value = true
  } finally {
    marketSignalLoading.value = false
  }
}

watch(newStatus, status => {
  if (status === 'bidding') {
    fetchBidRecommendation()
    getAgentStatus().then(res => { providerConfigured.value = res.data.configured }).catch(() => { providerConfigured.value = true })
  }
}, { immediate: true })

watch(() => props.lot.status, (status) => {
  currentStatus.value = status
  newStatus.value = status
})

async function linkEvent() {
  const eventId = selectedEventId.value === '' ? null : Number(selectedEventId.value)
  setStatusMessage('')
  try {
    await linkAuctionLotEvent(props.lot.id, eventId)
    emit('updated')
  } catch (err) {
    setStatusMessage(errorMessageFrom(err, t('auctions.detail.calendarFailed')), true)
  }
}

async function saveAlert() {
  if (!canCreateAlert.value) return
  alertBusy.value = true
  setAlertMessage('')
  try {
    await createAlert({
      auctionLotId: props.lot.id,
      targetPrice: alertForm.targetPrice ?? 0,
      direction: alertForm.direction,
    })
    emit('alertsUpdated')
    setAlertMessage(t('auctions.detail.alertSaved'))
  } catch {
    setAlertMessage(t('auctions.detail.alertSaveFailed'), true)
  } finally {
    alertBusy.value = false
  }
}

async function removeAlert(id: number) {
  alertBusy.value = true
  setAlertMessage('')
  try {
    await deleteAlert(id)
    emit('alertsUpdated')
    setAlertMessage(t('auctions.detail.alertDeleted'))
  } catch {
    setAlertMessage(t('auctions.detail.alertDeleteFailed'), true)
  } finally {
    alertBusy.value = false
  }
}

async function saveReminder() {
  if (!canCreateReminder.value) return
  reminderBusy.value = true
  setAlertMessage('')
  try {
    await createReminder({
      auctionLotId: props.lot.id,
      minutesBefore: reminderForm.minutesBefore ?? 30,
    })
    emit('alertsUpdated')
    setAlertMessage(t('auctions.detail.reminderSaved'))
  } catch {
    setAlertMessage(t('auctions.detail.reminderSaveFailed'), true)
  } finally {
    reminderBusy.value = false
  }
}

async function removeReminder(id: number) {
  reminderBusy.value = true
  setAlertMessage('')
  try {
    await deleteReminder(id)
    emit('alertsUpdated')
    setAlertMessage(t('auctions.detail.reminderDeleted'))
  } catch {
    setAlertMessage(t('auctions.detail.reminderDeleteFailed'), true)
  } finally {
    reminderBusy.value = false
  }
}

async function changeStatus() {
  statusBusy.value = true
  setStatusMessage('')
  try {
    const bid = maxBidChanged.value ? normalizedMaxBidInput.value : undefined
    const winBid = winningBidChanged.value ? normalizedWinningBidInput.value : undefined
    await updateAuctionLotStatus(props.lot.id, newStatus.value, bid, winBid)
    currentStatus.value = newStatus.value
    if (newStatus.value === 'watching' || newStatus.value === 'bidding') {
      await refreshQuickAccess()
    } else {
      forgetQuickAccess('auction_lot', props.lot.id)
    }

    if (newStatus.value === 'won') {
      const shipment = buildShipmentInput()
      if (shipment === null) {
        return
      }
      try {
        const coinRes = await convertAuctionLotToCoin(props.lot.id, shipment ? { shipment } : undefined)
        emit('close')
        router.push(`/edit/${coinRes.data.id}`)
        return
      } catch (err) {
        setStatusMessage(errorMessageFrom(err, t('auctions.detail.wonConvertFailed')), true)
        emit('updated')
        return
      }
    }

    emit('updated')
  } catch (err) {
    setStatusMessage(errorMessageFrom(err, t('auctions.detail.statusFailed')), true)
  } finally {
    statusBusy.value = false
  }
}

async function convertToCoin() {
  statusBusy.value = true
  setStatusMessage('')
  const shipment = buildShipmentInput()
  if (shipment === null) {
    statusBusy.value = false
    return
  }
  try {
    const coinRes = await convertAuctionLotToCoin(props.lot.id, shipment ? { shipment } : undefined)
    emit('close')
    router.push(`/edit/${coinRes.data.id}`)
  } catch (err) {
    setStatusMessage(errorMessageFrom(err, t('auctions.detail.convertFailed')), true)
  } finally {
    statusBusy.value = false
  }
}

async function removeLot() {
  statusBusy.value = true
  setStatusMessage('')
  try {
    await deleteAuctionLot(props.lot.id)
    forgetQuickAccess('auction_lot', props.lot.id)
    emit('close')
    emit('updated')
  } catch (err) {
    setStatusMessage(errorMessageFrom(err, t('auctions.detail.removeFailed')), true)
  } finally {
    statusBusy.value = false
  }
}

onMounted(fetchCalendarEvents)

function buildShipmentInput(): ShipmentUpsertInput | null | undefined {
  if (!attachShipment.value) return undefined
  if (!shipmentTrackingNumber.value.trim()) {
    setStatusMessage(t('purchase.trackingRequired'), true)
    return null
  }
  return {
    carrier: 'parcel',
    trackingNumber: shipmentTrackingNumber.value.trim(),
    notes: shipmentNotes.value.trim() || undefined,
  }
}
</script>

<style scoped>
.control-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: stretch;
  gap: 0.6rem;
}

.control-left-group {
  min-width: 0;
}

.control-left-group--alert {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 9rem;
  gap: 0.6rem;
}

.control-input {
  min-height: 2.5rem;
  width: 100%;
}

.control-number {
  min-width: 0;
}

.control-btn {
  min-height: 2.5rem;
  justify-content: center;
}

.control-btn-fixed {
  width: 9rem;
  min-width: 9rem;
  flex: 0 0 9rem;
}

.bid-input,
.winning-bid-input {
  width: 10rem;
  max-width: 10rem;
  flex: 0 0 10rem;
}

@media (max-width: 575px) {
  .control-row {
    grid-template-columns: 1fr;
  }

  .control-left-group--alert {
    grid-template-columns: 1fr;
  }

  .control-btn-fixed,
  .control-number,
  .bid-input,
  .winning-bid-input {
    min-width: 0;
    width: 100%;
    max-width: 100%;
    flex: 1 1 auto;
  }
}
</style>
