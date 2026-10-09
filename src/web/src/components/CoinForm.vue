<template>
  <form class="mx-auto max-w-[900px]" @submit.prevent="handleSubmit">
    <div class="grid gap-6 md:grid-cols-2">
      <!-- Basic Info -->
      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('coinForm.sections.basic') }}</h2>
        <div class="form-group min-w-0">
          <label class="form-label">{{ t('coinForm.name') }}</label>
          <AutocompleteInput v-model="form.name!" field="name" required :placeholder="t('coinForm.namePlaceholder')" />
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('collection.menu.category') }}</label>
            <select v-model="form.category" class="form-select">
              <option v-for="c in categoryOptions" :key="c" :value="c">{{ categoryLabel(c) }}</option>
            </select>
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinFields.material') }}</label>
            <select v-model="form.material" class="form-select">
              <option v-for="m in materialOptions" :key="m" :value="m">{{ materialLabel(m) }}</option>
            </select>
          </div>
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinFields.denomination') }}</label>
            <AutocompleteInput v-model="form.denomination!" field="denomination" :placeholder="t('coinForm.denominationPlaceholder')" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinFields.mint') }}</label>
            <select v-model="mintLocationIdModel" class="form-select" :disabled="mintLocationsLoading">
              <option value="">{{ t('coinForm.mintUnknown') }}</option>
              <optgroup v-if="myMintLocations.length" :label="t('coinForm.myMints')">
                <option v-for="location in myMintLocations" :key="location.id" :value="String(location.id)">
                  {{ location.displayName }}
                </option>
              </optgroup>
              <optgroup v-if="globalMintLocations.length" :label="t('coinForm.mints')">
                <option v-for="location in globalMintLocations" :key="location.id" :value="String(location.id)">
                  {{ location.displayName }}
                </option>
              </optgroup>
              <option value="__create__">{{ t('coinForm.createMint') }}</option>
            </select>
            <p v-if="mintLocationError" class="mt-2 text-body text-text-secondary">{{ mintLocationError }}</p>
            <p v-else-if="form.mint && !form.mintLocationId" class="mt-2 text-body text-text-secondary">
              {{ t('coinForm.unlinkedMint', { mint: form.mint }) }}
            </p>
          </div>
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinFields.ruler') }}</label>
            <AutocompleteInput v-model="form.ruler!" field="ruler" :placeholder="t('coinForm.rulerPlaceholder')" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinFields.era') }}</label>
            <select v-model="form.era" class="form-select">
              <option value="">{{ t('coinForm.eraUnspecified') }}</option>
              <option v-for="era in displayedEraOptions" :key="era" :value="era">{{ eraLabel(era) }}</option>
            </select>
          </div>
        </div>
        <div class="form-group min-w-0">
          <label class="form-label">{{ t('coinForm.dateRange') }}</label>
          <input
            v-model="form.dateRange"
            class="form-input"
            type="text"
            maxlength="200"
            :placeholder="t('coinForm.dateRangePlaceholder')"
          />
        </div>
        <div v-if="form.category === 'Roman'" class="form-group min-w-0">
          <label class="form-label">{{ t('coinForm.imperialFigure') }}</label>
          <ImperialFigurePicker v-model="form.romanImperialFigureId!" />
          <p class="mt-1 text-sm text-text-muted">
            <I18nT keypath="coinForm.imperialFigureHint">
              <template #link><router-link to="/sets/emperors" class="underline">{{ t('coinForm.emperorTracker') }}</router-link></template>
            </I18nT>
          </p>
        </div>
        <div class="form-group min-w-0">
          <label class="form-label">{{ t('coinFields.storageLocation') }}</label>
          <select v-model="storageLocationIdModel" class="form-select" :disabled="storageLocationsLoading">
            <option value="">{{ t('coinForm.storageNone') }}</option>
            <option
              v-for="location in storageLocations"
              :key="location.id"
              :value="String(location.id)"
            >
              {{ location.name }}
            </option>
          </select>
          <p v-if="storageLocationError" class="mt-2 text-body text-text-secondary">{{ storageLocationError }}</p>
          <div v-if="selectedStorageLocation?.type === 'tray'" class="mt-3">
            <p class="text-body text-text-secondary">
              {{ t('coinForm.chooseSlot', { occupied: selectedStorageLocation.occupied, capacity: selectedStorageLocation.capacity }) }}
            </p>
            <p v-if="occupancyLoading" class="text-body text-text-muted">{{ t('coinForm.loadingSlots') }}</p>
            <div
              v-else-if="trayOccupancy"
              class="slot-picker"
              role="grid"
              :aria-label="t('coinForm.slotsLabel', { name: selectedStorageLocation.name })"
              :style="{ gridTemplateColumns: `repeat(${trayOccupancy.columns}, minmax(44px, 1fr))` }"
            >
              <button
                v-for="slot in trayOccupancy.capacity"
                :key="slot"
                type="button"
                role="gridcell"
                class="slot-choice"
                :class="{ selected: form.storageSlot === slot, occupied: isOccupiedSlot(slot) }"
                :disabled="isOccupiedSlot(slot)"
                :aria-label="slotLabel(slot)"
                :aria-selected="form.storageSlot === slot"
                @click="form.storageSlot = slot"
              >
                <span>{{ slotCoordinates(slot) }}</span>
                <small>{{ isOccupiedSlot(slot) ? t('coinForm.slotState.occupied') : form.storageSlot === slot ? t('coinForm.slotState.selected') : t('coinForm.slotState.available') }}</small>
              </button>
            </div>
            <p v-if="occupancyError" class="mt-2 text-body text-text-secondary">
              {{ occupancyError }}
              <button type="button" class="btn btn-xs btn-secondary" @click="loadOccupancy">{{ t('common.retry') }}</button>
            </p>
            <p v-if="slotValidationError" class="mt-2 text-body text-text-secondary" role="alert">{{ slotValidationError }}</p>
          </div>
        </div>
      </fieldset>

      <!-- Physical Details -->
      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('coinForm.sections.physical') }}</h2>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.weightGrams') }}</label>
            <input v-model.number="form.weightGrams" class="form-input" type="number" step="0.01" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.diameterMm') }}</label>
            <input v-model.number="form.diameterMm" class="form-input" type="number" step="0.1" />
          </div>
        </div>
        <div class="form-group min-w-0">
          <label class="form-label">{{ t('coinFields.grade') }}</label>
          <input v-model="form.grade" class="form-input" :placeholder="t('coinForm.gradePlaceholder')" />
        </div>
      </fieldset>

      <!-- Inscriptions, Images & Descriptions -->
      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5 md:col-span-2">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('coinForm.sections.inscriptions') }}</h2>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.obverseInscription') }}</label>
            <input v-model="form.obverseInscription" class="form-input" :placeholder="t('coinForm.obverseLegend')" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.reverseInscription') }}</label>
            <input v-model="form.reverseInscription" class="form-input" :placeholder="t('coinForm.reverseLegend')" />
          </div>
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.obverseImage') }}</label>
            <div v-if="obversePreview || existingObverse" class="relative mb-2 inline-block">
              <img v-if="obversePreview" :src="obversePreview" :alt="t('common.obverse')" class="h-[140px] w-[140px] rounded-sm border border-border-subtle object-cover" />
              <AuthenticatedImage v-else :media-path="existingObverse" :alt="t('common.obverse')" class="h-[140px] w-[140px] rounded-sm border border-border-subtle object-cover" />
              <button type="button" class="absolute -top-1.5 -right-1.5 flex h-[22px] w-[22px] items-center justify-center rounded-full border-0 bg-red-700 p-0 text-white" @click="clearObverse" :title="t('common.remove')" :aria-label="t('coinForm.removeObverseImage')"><X :size="12" /></button>
            </div>
            <div class="flex items-center gap-2">
              <input type="file" accept=".jpg,.jpeg,.png" class="form-input min-w-0 flex-1 text-base" :aria-label="t('coinForm.uploadObverseImage')" @change="onObverseFile" ref="obverseInput" />
              <label v-if="isPwa" class="btn btn-secondary btn-sm inline-flex cursor-pointer items-center gap-1.5 whitespace-nowrap">
                <Camera :size="14" /> {{ t('coinDetail.actions.photo') }}
                <input type="file" accept="image/*" capture="environment" hidden @change="onObverseFile" />
              </label>
            </div>
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.reverseImage') }}</label>
            <div v-if="reversePreview || existingReverse" class="relative mb-2 inline-block">
              <img v-if="reversePreview" :src="reversePreview" :alt="t('common.reverse')" class="h-[140px] w-[140px] rounded-sm border border-border-subtle object-cover" />
              <AuthenticatedImage v-else :media-path="existingReverse" :alt="t('common.reverse')" class="h-[140px] w-[140px] rounded-sm border border-border-subtle object-cover" />
              <button type="button" class="absolute -top-1.5 -right-1.5 flex h-[22px] w-[22px] items-center justify-center rounded-full border-0 bg-red-700 p-0 text-white" @click="clearReverse" :title="t('common.remove')" :aria-label="t('coinForm.removeReverseImage')"><X :size="12" /></button>
            </div>
            <div class="flex items-center gap-2">
              <input type="file" accept=".jpg,.jpeg,.png" class="form-input min-w-0 flex-1 text-base" :aria-label="t('coinForm.uploadReverseImage')" @change="onReverseFile" ref="reverseInput" />
              <label v-if="isPwa" class="btn btn-secondary btn-sm inline-flex cursor-pointer items-center gap-1.5 whitespace-nowrap">
                <Camera :size="14" /> {{ t('coinDetail.actions.photo') }}
                <input type="file" accept="image/*" capture="environment" hidden @change="onReverseFile" />
              </label>
            </div>
          </div>
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.obverseDescription') }}</label>
            <textarea v-model="form.obverseDescription" class="form-textarea" :placeholder="t('coinForm.describeObverse')" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.reverseDescription') }}</label>
            <textarea v-model="form.reverseDescription" class="form-textarea" :placeholder="t('coinForm.describeReverse')" />
          </div>
        </div>
      </fieldset>

      <!-- Purchase Info -->
      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('coinForm.sections.purchase') }}</h2>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.purchasePriceUsd') }}</label>
            <input v-model.number="form.purchasePrice" class="form-input" type="number" step="0.01" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.currentValueUsd') }}</label>
            <input v-model.number="form.currentValue" class="form-input" type="number" step="0.01" />
          </div>
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinFields.purchaseDate') }}</label>
            <input v-model="form.purchaseDate" class="form-input" type="date" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinFields.store') }}</label>
            <AutocompleteInput v-model="form.purchaseLocation!" field="purchaseLocation" :placeholder="t('coinForm.storePlaceholder')" />
          </div>
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinFields.vendorSku') }}</label>
            <input v-model="form.vendorSku" class="form-input" :placeholder="t('coinForm.skuPlaceholder')" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinFields.vendorInvoice') }}</label>
            <input v-model="form.vendorInvoice" class="form-input" :placeholder="t('coinForm.invoicePlaceholder')" />
          </div>
        </div>
      </fieldset>

      <!-- Links & Notes -->
      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('coinForm.sections.reference') }}</h2>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.referenceUrl') }}</label>
            <input v-model="form.referenceUrl" class="form-input" type="url" placeholder="https://..." />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.referenceText') }}</label>
            <input v-model="form.referenceText" class="form-input" :placeholder="t('coinForm.referenceTextPlaceholder')" />
          </div>
        </div>
        <div class="form-group min-w-0">
          <label class="form-label">{{ t('coinForm.storeCardImage') }}</label>
          <p class="mb-2 text-body text-text-muted">{{ t('coinForm.storeCardHint') }}</p>
          <div v-if="cardPreview" class="relative mb-2 inline-block">
            <img :src="cardPreview" :alt="t('coinForm.storeCardAlt')" class="h-[140px] w-[140px] rounded-sm border border-border-subtle object-cover" />
            <button type="button" class="absolute -top-1.5 -right-1.5 flex h-[22px] w-[22px] items-center justify-center rounded-full border-0 bg-red-700 p-0 text-white" @click="clearCard" :title="t('common.remove')" :aria-label="t('coinForm.removeStoreCard')"><X :size="12" /></button>
          </div>
          <input type="file" accept=".jpg,.jpeg,.png" class="form-input text-base" :aria-label="t('coinForm.uploadStoreCard')" @change="onCardFile" ref="cardInput" />
        </div>
        <div class="form-group min-w-0">
          <label class="form-label">{{ t('coinFields.notes') }}</label>
          <textarea v-model="form.notes" class="form-textarea" rows="3" :placeholder="t('coinForm.notesPlaceholder')" />
        </div>
        <div class="form-group flex items-center gap-3">
          <label class="form-label mb-0">{{ t('coinForm.privateCoin') }}</label>
          <BaseToggle
            v-model="form.isPrivate"
            size="sm"
            :label="t('coinForm.privateCoin')"
          />
          <span class="text-chip text-text-secondary">{{ t('coinForm.hiddenFromFollowers') }}</span>
        </div>
      </fieldset>
    </div>

    <div class="mt-8 flex justify-end gap-3 border-t border-border-subtle pt-6">
      <button type="button" class="btn btn-secondary" @click="$router.back()">{{ t('common.cancel') }}</button>
      <button type="submit" class="btn btn-primary" :disabled="loading">
        {{ loading ? t('common.saving') : submitLabel }}
      </button>
    </div>

    <CreateMintModal
      :open="showCreateMintModal"
      :initial-name="pendingMintName"
      @close="onCreateMintClosed"
      @created="onMintCreated"
    />
  </form>
</template>

<script setup lang="ts">
import { BaseToggle } from '@/components/ui'
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { getStorageLocations, getStorageLocationOccupancy, getMintLocations, type MintLocationsResponse } from '@/api/client'
import type { Coin, StorageLocation, MintLocation, TrayOccupancy } from '@/types'
import AutocompleteInput from '@/components/AutocompleteInput.vue'
import ImperialFigurePicker from '@/components/ImperialFigurePicker.vue'
import CreateMintModal from '@/components/CreateMintModal.vue'
import { X, Camera } from 'lucide-vue-next'
import { usePwa } from '@/composables/usePwa'
import { useCoinOptions } from '@/composables/useCoinOptions'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import I18nT from '@/components/I18nT.vue'
import { materialLabel } from '@/utils/format'
import { t } from '@/i18n'
import { categoryLabel, eraLabel } from '@/i18n/coinVocabulary'

function unwrapMintLocations(data: MintLocationsResponse): MintLocation[] {
  return Array.isArray(data) ? data : data.mintLocations ?? []
}

const { isPwa } = usePwa()
const { categoryOptions, eraOptions, materialOptions, loadOptions } = useCoinOptions()

const props = defineProps<{
  form: Partial<Coin>
  submitLabel: string
  loading?: boolean
  coinId?: number
}>()

const emit = defineEmits<{ submit: [] }>()

const obverseFile = ref<File | null>(null)
const reverseFile = ref<File | null>(null)
const cardFile = ref<File | null>(null)
const obversePreview = ref<string | null>(null)
const reversePreview = ref<string | null>(null)
const cardPreview = ref<string | null>(null)
const obverseInput = ref<HTMLInputElement | null>(null)
const reverseInput = ref<HTMLInputElement | null>(null)
const cardInput = ref<HTMLInputElement | null>(null)
const removedObverseId = ref<number | null>(null)
const removedReverseId = ref<number | null>(null)
const storageLocations = ref<StorageLocation[]>([])
const storageLocationsLoading = ref(false)
const storageLocationError = ref('')
const trayOccupancy = ref<TrayOccupancy | null>(null)
const occupancyLoading = ref(false)
const occupancyError = ref('')
const slotValidationError = ref('')
const selectedStorageLocation = computed(() =>
  storageLocations.value.find((location) => location.id === props.form.storageLocationId) ?? null
)

const storageLocationIdModel = computed({
  get: () => props.form.storageLocationId == null ? '' : String(props.form.storageLocationId),
  set: (value: string) => {
    const next = value === '' ? null : Number(value)
    if (props.form.storageLocationId !== next) {
      props.form.storageSlot = null
      trayOccupancy.value = null
      slotValidationError.value = ''
    }
    props.form.storageLocationId = next
  },
})

async function loadOccupancy() {
  const location = selectedStorageLocation.value
  if (!location || location.type !== 'tray') {
    trayOccupancy.value = null
    return
  }
  occupancyLoading.value = true
  occupancyError.value = ''
  try {
    const response = await getStorageLocationOccupancy(location.id, props.coinId)
    trayOccupancy.value = response.data
  } catch {
    trayOccupancy.value = null
    occupancyError.value = t('coinForm.occupancyFailed')
  } finally {
    occupancyLoading.value = false
  }
}

function isOccupiedSlot(slot: number): boolean {
  return trayOccupancy.value?.occupiedSlots.includes(slot) === true &&
    trayOccupancy.value.currentCoinSlot !== slot
}

function slotCoordinates(slot: number): string {
  const columns = trayOccupancy.value?.columns ?? 1
  return `${Math.floor((slot - 1) / columns) + 1},${((slot - 1) % columns) + 1}`
}

function slotLabel(slot: number): string {
  const state = isOccupiedSlot(slot)
    ? t('coinForm.slotStateLower.occupied')
    : props.form.storageSlot === slot ? t('coinForm.slotStateLower.selected') : t('coinForm.slotStateLower.available')
  const [row = '', column = ''] = slotCoordinates(slot).split(',')
  return t('coinForm.slotLabel', { row, column, state })
}

function handleSubmit() {
  if (selectedStorageLocation.value?.type === 'tray' && props.form.storageSlot == null) {
    slotValidationError.value = t('coinForm.chooseSlotBeforeSave')
    return
  }
  slotValidationError.value = ''
  emit('submit')
}

watch(() => props.form.storageLocationId, () => {
  void loadOccupancy()
})

const mintLocations = ref<MintLocation[]>([])
const mintLocationsLoading = ref(false)
const mintLocationError = ref('')
const showCreateMintModal = ref(false)
const pendingMintName = ref('')

const myMintLocations = computed(() => mintLocations.value.filter((m) => m.userId != null))
const globalMintLocations = computed(() => mintLocations.value.filter((m) => m.userId == null))

const mintLocationIdModel = computed({
  get: () => props.form.mintLocationId == null ? '' : String(props.form.mintLocationId),
  set: (value: string) => {
    if (value === '__create__') {
      pendingMintName.value = ''
      showCreateMintModal.value = true
      return
    }
    props.form.mintLocationId = value === '' ? null : Number(value)
    const selected = mintLocations.value.find((m) => String(m.id) === value)
    props.form.mint = selected ? selected.displayName : ''
  },
})

async function loadMintLocations() {
  mintLocationsLoading.value = true
  try {
    const res = await getMintLocations()
    mintLocations.value = unwrapMintLocations(res.data)
    mintLocationError.value = ''
  } catch {
    mintLocations.value = []
    mintLocationError.value = t('coinForm.mintsUnavailable')
  } finally {
    mintLocationsLoading.value = false
  }
}

function onCreateMintClosed() {
  showCreateMintModal.value = false
}

function onMintCreated(mintLocation: MintLocation) {
  showCreateMintModal.value = false
  mintLocations.value.push(mintLocation)
  props.form.mintLocationId = mintLocation.id
  props.form.mint = mintLocation.displayName
}

const displayedEraOptions = computed(() => {
  const currentEra = typeof props.form.era === 'string' ? props.form.era.trim() : ''
  if (currentEra && !eraOptions.value.includes(currentEra)) {
    return [currentEra, ...eraOptions.value]
  }
  return eraOptions.value
})

watch(() => props.form.category, (category) => {
  if (category !== 'Roman') {
    props.form.romanImperialFigureId = null
  }
})

onMounted(async () => {
  // Load coin property options from settings
  loadOptions()
  
  // Load storage locations
  storageLocationsLoading.value = true
  try {
    const res = await getStorageLocations()
    storageLocations.value = res.data?.storageLocations ?? []
    await loadOccupancy()
  } catch {
    storageLocations.value = []
    storageLocationError.value = t('coinForm.storageUnavailable')
  } finally {
    storageLocationsLoading.value = false
  }

  loadMintLocations()
})

const existingObverse = computed(() => {
  if (removedObverseId.value) return null
  const img = props.form.images?.find((i) => i.imageType === 'obverse')
  return img ? img.filePath : null
})

const existingReverse = computed(() => {
  if (removedReverseId.value) return null
  const img = props.form.images?.find((i) => i.imageType === 'reverse')
  return img ? img.filePath : null
})

function onObverseFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (obversePreview.value) URL.revokeObjectURL(obversePreview.value)
  obverseFile.value = file
  obversePreview.value = URL.createObjectURL(file)
}

function onReverseFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (reversePreview.value) URL.revokeObjectURL(reversePreview.value)
  reverseFile.value = file
  reversePreview.value = URL.createObjectURL(file)
}

function clearObverse() {
  const existing = props.form.images?.find((i) => i.imageType === 'obverse')
  if (existing) removedObverseId.value = existing.id
  if (obversePreview.value) URL.revokeObjectURL(obversePreview.value)
  obverseFile.value = null
  obversePreview.value = null
  if (obverseInput.value) obverseInput.value.value = ''
}

function clearReverse() {
  const existing = props.form.images?.find((i) => i.imageType === 'reverse')
  if (existing) removedReverseId.value = existing.id
  if (reversePreview.value) URL.revokeObjectURL(reversePreview.value)
  reverseFile.value = null
  reversePreview.value = null
  if (reverseInput.value) reverseInput.value.value = ''
}

function onCardFile(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (cardPreview.value) URL.revokeObjectURL(cardPreview.value)
  cardFile.value = file
  cardPreview.value = URL.createObjectURL(file)
}

function clearCard() {
  if (cardPreview.value) URL.revokeObjectURL(cardPreview.value)
  cardFile.value = null
  cardPreview.value = null
  if (cardInput.value) cardInput.value.value = ''
}

onBeforeUnmount(() => {
  if (obversePreview.value) URL.revokeObjectURL(obversePreview.value)
  if (reversePreview.value) URL.revokeObjectURL(reversePreview.value)
  if (cardPreview.value) URL.revokeObjectURL(cardPreview.value)
})

// Expose pending images for parent to upload after save
defineExpose({
  obverseFile,
  reverseFile,
  cardFile,
  removedObverseId,
  removedReverseId,
  refreshStorageOccupancy: loadOccupancy,
})
</script>

<style scoped>
.slot-picker {
  display: grid;
  gap: 0.35rem;
  max-width: 100%;
  overflow-x: auto;
  padding: 0.35rem;
}

.slot-choice {
  min-width: 44px;
  min-height: 44px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-primary);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.slot-choice small {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.slot-choice.selected {
  border-color: var(--accent-gold);
  background: var(--accent-gold-dim);
}

.slot-choice.occupied {
  opacity: 0.55;
}

.slot-choice:focus-visible {
  outline: 2px solid var(--accent-gold);
  outline-offset: 2px;
}
</style>
