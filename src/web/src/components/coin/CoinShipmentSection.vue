<template>
  <section class="mb-6">
    <h3 class="mb-3 font-display text-base font-medium text-heading">{{ t('shipment.title') }}</h3>
    <div class="rounded-sm border border-border-subtle bg-card p-4">
      <div v-if="loading" class="text-body text-text-secondary">{{ t('shipment.loading') }}</div>

      <template v-else-if="shipment">
        <div class="mb-3 flex flex-wrap items-center gap-2">
          <span class="chip-sm">{{ carrierLabel(shipment.carrier, shipment.manualCarrierName) }}</span>
          <span class="chip-sm">{{ statusLabel(shipment.currentStatus) }}</span>
          <a v-if="trackingUrl" :href="trackingUrl" target="_blank" rel="noopener" class="text-chip text-gold hover:underline">{{ t('shipment.openCarrier') }}</a>
        </div>
        <div class="mb-4 grid gap-2 text-body text-text-secondary">
          <p class="m-0"><strong class="text-text-primary">{{ t('shipment.tracking') }}</strong> {{ shipment.trackingNumber }}</p>
          <p v-if="shipment.notes" class="m-0"><strong class="text-text-primary">{{ t('shipment.notes') }}</strong> {{ shipment.notes }}</p>
          <p class="m-0"><strong class="text-text-primary">{{ t('shipment.status') }}</strong> {{ statusLabel(shipment.currentStatus) }}</p>
          <p v-if="shipment.lastSyncError" class="m-0 text-loss"><strong>ParcelApp:</strong> {{ shipment.lastSyncError }}</p>
        </div>

        <div class="mb-4 grid gap-2 border-t border-border-subtle pt-4">
          <label class="form-label">{{ t('shipment.updateStatus') }}</label>
          <p class="m-0 text-body text-text-secondary">{{ t('shipment.manualHint') }}</p>
          <div class="grid gap-2 md:grid-cols-2">
            <select v-model="manualStatus" class="form-input">
              <option v-for="value in statusOptions" :key="value" :value="value">{{ statusLabel(value) }}</option>
            </select>
            <input v-model="manualNote" class="form-input" :placeholder="t('shipment.statusNotePlaceholder')" />
          </div>
          <button class="btn btn-secondary btn-sm w-fit" :disabled="saving" @click="saveStatus">
            {{ saving ? t('common.saving') : t('shipment.saveStatus') }}
          </button>
        </div>

        <div class="mb-4 grid gap-2 border-t border-border-subtle pt-4">
          <h4 class="m-0 text-sm font-semibold text-text-primary">{{ t('nav.timeline') }}</h4>
          <p v-if="!shipment.events?.length" class="m-0 text-body text-text-muted">{{ t('shipment.noEvents') }}</p>
          <ul v-else class="m-0 grid list-none gap-2 p-0">
            <li v-for="event in shipment.events" :key="event.id" class="rounded-sm border border-border-subtle bg-card-hover px-3 py-2 text-body text-text-secondary">
              <p class="m-0 font-medium text-text-primary">{{ statusLabel(event.status) }}</p>
              <p class="m-0">{{ formatDate(event.occurredAt) }}<span v-if="event.location"> • {{ event.location }}</span></p>
              <p v-if="event.description" class="m-0">{{ event.description }}</p>
            </li>
          </ul>
        </div>

        <div class="flex flex-wrap gap-2 border-t border-border-subtle pt-4">
          <button
            class="btn btn-secondary btn-sm"
            :disabled="syncing || shipment.carrier !== 'parcel' || isDelivered"
            :title="isDelivered ? t('shipment.deliveredTitle') : undefined"
            @click="syncShipment"
          >
            {{ isDelivered ? t('shipment.trackingComplete') : (syncing ? t('common.checking') : t('shipment.checkNow')) }}
          </button>
          <button class="btn btn-danger btn-sm" :disabled="deleting" @click="removeShipment">{{ deleting ? t('common.removing') : t('shipment.remove') }}</button>
        </div>
      </template>

      <template v-else>
        <p class="m-0 mb-4 text-body text-text-secondary">{{ t('shipment.none') }}</p>
        <p class="m-0 mb-4 text-body text-text-secondary">{{ t('shipment.enterTracking') }}</p>
        <div class="grid gap-3 md:grid-cols-2">
          <input v-model="form.trackingNumber" class="form-input" :placeholder="t('shipment.trackingNumber')" />
          <input v-model="form.notes" class="form-input md:col-span-2" :placeholder="t('shipment.optionalNotes')" />
        </div>
        <div class="mt-4">
          <button class="btn btn-primary btn-sm" :disabled="saving" @click="saveShipment">{{ saving ? t('common.saving') : t('shipment.save') }}</button>
        </div>
      </template>

      <p v-if="message" class="mt-3 mb-0 text-body" :class="messageError ? 'text-loss' : 'text-gold'">{{ message }}</p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { deleteCoinShipment, getCoinShipment, setCoinShipmentManualOverride, syncCoinShipment, upsertCoinShipment } from '@/api/client'
import type { Shipment, ShipmentCarrier, ShipmentStatus, ShipmentUpsertInput } from '@/types'
import { shipmentStatusLabel } from '@/utils/format'
import { intlLocale, t } from '@/i18n'

const props = defineProps<{ coinId: number }>()
const emit = defineEmits<{ changed: [] }>()

const loading = ref(true)
const saving = ref(false)
const syncing = ref(false)
const deleting = ref(false)
const shipment = ref<Shipment | null>(null)
const trackingUrl = ref('')
const message = ref('')
const messageError = ref(false)
const manualStatus = ref<ShipmentStatus>('in_transit')
const manualNote = ref('')
const form = reactive<ShipmentUpsertInput>({
  carrier: 'parcel',
  trackingNumber: '',
  notes: '',
  manualCarrierName: '',
})

const statusOptions: ShipmentStatus[] = ['pending', 'label_created', 'in_transit', 'out_for_delivery', 'delivered', 'exception', 'returned', 'unknown']

// Manual ParcelApp checks are pointless once a shipment is delivered — the terminal
// status won't change on the carrier side (spec 340 delivered-terminal-state). This is
// reactive off shipment.currentStatus, so switching status away from delivered (e.g. via
// the manual override above) immediately re-enables the action.
const isDelivered = computed(() => shipment.value?.currentStatus === 'delivered')

function setMessage(next: string, isError = false) {
  message.value = next
  messageError.value = isError
}

function apiErrorMessage(err: unknown): string {
  const maybe = err as { response?: { data?: { error?: string; message?: string } }; message?: string }
  return maybe?.response?.data?.error || maybe?.response?.data?.message || maybe?.message || ''
}

function statusLabel(status: string) {
  return shipmentStatusLabel(status)
}

function carrierLabel(carrier: ShipmentCarrier, manualName: string) {
  if (carrier === 'parcel') return 'ParcelApp'
  if (carrier === 'other' && manualName) return manualName
  return carrier.toUpperCase()
}

function formatDate(value: string) {
  return new Date(value).toLocaleString(intlLocale())
}

async function loadShipment() {
  loading.value = true
  try {
    const res = await getCoinShipment(props.coinId)
    shipment.value = res.data.shipment
    trackingUrl.value = res.data.trackingUrl ?? ''
    manualStatus.value = (shipment.value.currentStatus || shipment.value.manualOverrideStatus || 'in_transit') as ShipmentStatus
    manualNote.value = shipment.value.manualOverrideNote || ''
  } catch (err: unknown) {
    const msg = apiErrorMessage(err).toLowerCase()
    if (msg.includes('shipment not found')) {
      shipment.value = null
      trackingUrl.value = ''
    } else {
      setMessage(apiErrorMessage(err) || t('shipment.loadFailed'), true)
    }
  } finally {
    loading.value = false
  }
}

function buildUpsertInput(): ShipmentUpsertInput | null {
  const trackingNumber = form.trackingNumber.trim()
  if (!trackingNumber) {
    setMessage(t('shipment.trackingRequired'), true)
    return null
  }
  return {
    carrier: 'parcel',
    trackingNumber,
    notes: form.notes?.trim() || '',
    manualCarrierName: '',
  }
}

async function saveShipment() {
  const input = buildUpsertInput()
  if (!input) return
  saving.value = true
  setMessage('')
  try {
    const res = await upsertCoinShipment(props.coinId, input)
    shipment.value = res.data.shipment
    trackingUrl.value = res.data.trackingUrl ?? ''
    if (shipment.value.lastSyncError) {
      setMessage(t('shipment.savedSyncFailed', { error: shipment.value.lastSyncError }), true)
    } else {
      setMessage(t('shipment.saved'))
    }
    emit('changed')
  } catch (err: unknown) {
    setMessage(apiErrorMessage(err) || t('shipment.saveFailed'), true)
  } finally {
    saving.value = false
  }
}

async function saveStatus() {
  if (!shipment.value) return
  saving.value = true
  setMessage('')
  try {
    const res = await setCoinShipmentManualOverride(props.coinId, {
      enabled: true,
      status: manualStatus.value,
      note: manualNote.value.trim(),
    })
    shipment.value = res.data.shipment
    trackingUrl.value = res.data.trackingUrl ?? ''
    setMessage(t('shipment.statusSaved'))
  } catch (err: unknown) {
    setMessage(apiErrorMessage(err) || t('shipment.statusSaveFailed'), true)
  } finally {
    saving.value = false
  }
}

async function syncShipment() {
  if (!shipment.value) return
  syncing.value = true
  setMessage('')
  try {
    const res = await syncCoinShipment(props.coinId)
    shipment.value = res.data.shipment
    trackingUrl.value = res.data.trackingUrl ?? ''
    if (shipment.value.lastSyncError) {
      setMessage(t('shipment.syncFailed', { error: shipment.value.lastSyncError }), true)
    } else {
      setMessage(t('shipment.synced'))
    }
  } catch (err: unknown) {
    setMessage(apiErrorMessage(err) || t('shipment.syncRequestFailed'), true)
  } finally {
    syncing.value = false
  }
}

async function removeShipment() {
  deleting.value = true
  setMessage('')
  try {
    await deleteCoinShipment(props.coinId)
    shipment.value = null
    trackingUrl.value = ''
    setMessage(t('shipment.removed'))
    emit('changed')
  } catch (err: unknown) {
    setMessage(apiErrorMessage(err) || t('shipment.removeFailed'), true)
  } finally {
    deleting.value = false
  }
}

onMounted(loadShipment)
</script>
