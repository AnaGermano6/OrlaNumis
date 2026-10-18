<template>
  <form class="mx-auto max-w-[900px]" @submit.prevent="handleSubmit">
    <div class="grid gap-6 md:grid-cols-2">
      <!-- Basic Info -->
      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('coinForm.sections.basic') }}</h2>
        <div v-if="coinId" class="form-group min-w-0">
          <label class="form-label" for="coin-kind">{{ t('coinForm.kind') }}</label>
          <select id="coin-kind" v-model="kindModel" class="form-select" data-testid="coin-kind">
            <option value="coin">{{ t('coinForm.kindCoin') }}</option>
            <option value="banknote">{{ t('coinForm.kindBanknote') }}</option>
          </select>
        </div>
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
        <div class="grid gap-3 md:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
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
          <div class="form-group min-w-0">
            <label class="form-label" for="coin-year">{{ t('coinForm.year') }}</label>
            <input
              id="coin-year"
              class="form-input"
              type="number"
              inputmode="numeric"
              step="1"
              min="-5000"
              max="3000"
              :value="form.year ?? ''"
              :placeholder="yearPlaceholder"
              aria-describedby="coin-year-hint"
              @input="setYear(($event.target as HTMLInputElement).value)"
            />
          </div>
        </div>
        <p id="coin-year-hint" class="-mt-2 mb-3 text-chip text-text-muted">{{ t('coinForm.yearHint') }}</p>
        <div class="form-group min-w-0">
          <label class="form-label">{{ t('coinFields.country') }}</label>
          <AutocompleteInput v-model="countryModel" field="country" :placeholder="t('coinForm.countryPlaceholder')" />
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label" for="coin-presentation">{{ t('coinFields.presentation') }}</label>
            <select id="coin-presentation" v-model="presentationModel" class="form-select">
              <option value="">{{ t('coinForm.notSpecified') }}</option>
              <option v-for="option in COIN_PRESENTATIONS" :key="option" :value="option">{{ presentationLabel(option) }}</option>
            </select>
          </div>
          <div class="form-group min-w-0">
            <label class="form-label" for="coin-rarity">{{ t('coinFields.rarity') }}</label>
            <select id="coin-rarity" v-model="rarityModel" class="form-select">
              <option value="">{{ t('coinForm.notSpecified') }}</option>
              <option v-for="option in COIN_RARITIES" :key="option" :value="option">{{ rarityLabel(option) }}</option>
            </select>
          </div>
        </div>
        <div class="form-group min-w-0">
          <label class="form-label" for="coin-circulation">{{ t('coinFields.circulation') }}</label>
          <select id="coin-circulation" v-model="circulationModel" class="form-select" aria-describedby="coin-circulation-hint">
            <option value="">{{ t('coinForm.circulationAuto') }}</option>
            <option v-for="option in COIN_CIRCULATIONS" :key="option" :value="option">{{ circulationLabel(option) }}</option>
          </select>
          <p id="coin-circulation-hint" class="mt-1 text-chip text-text-muted">{{ t('coinForm.circulationHint') }}</p>
        </div>
      </fieldset>

      <!-- Physical Details -->
      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('coinForm.sections.physical') }}</h2>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.weightGrams') }}</label>
            <input v-model.number="form.weightGrams" class="form-input" type="number" step="any" min="0" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.diameterMm') }}</label>
            <input v-model.number="form.diameterMm" class="form-input" type="number" step="any" min="0" />
          </div>
        </div>
        <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto]">
          <div class="form-group min-w-0">
            <label class="form-label" for="coin-grade">{{ t('coinFields.grade') }}</label>
            <!-- As the banknotes': each grade with what it means. -->
            <select id="coin-grade" v-model="gradeModel" class="form-select" data-testid="coin-grade">
              <option value="">{{ t('coinForm.notSpecified') }}</option>
              <option v-if="customGrade" :value="customGrade">{{ customGrade }}</option>
              <optgroup :label="t('coinForm.gradeGroups.conservation')">
                <option v-for="grade in COIN_GRADES" :key="grade" :value="grade">{{ coinGradeLabel(grade) }}</option>
              </optgroup>
              <optgroup :label="t('coinForm.gradeGroups.finishes')">
                <option v-for="grade in COIN_FINISHES" :key="grade" :value="grade">{{ coinGradeLabel(grade) }}</option>
              </optgroup>
            </select>
          </div>
          <div class="form-group min-w-0">
            <label class="form-label" for="coin-quantity">{{ t('coinFields.quantity') }}</label>
            <QuantityInput id="coin-quantity" v-model="form.quantity" />
          </div>
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
            <label class="form-label">{{ t('coinForm.purchasePriceUsd', { currency: collectionCurrencySymbol() }) }}</label>
            <input v-model.number="form.purchasePrice" class="form-input" type="number" step="0.01" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinForm.currentValueUsd', { currency: collectionCurrencySymbol() }) }}</label>
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
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { getMintLocations, type MintLocationsResponse } from '@/api/client'
import type { Coin, CoinKind, MintLocation } from '@/types'
import AutocompleteInput from '@/components/AutocompleteInput.vue'
import QuantityInput from '@/components/ui/QuantityInput.vue'
import CreateMintModal from '@/components/CreateMintModal.vue'
import { X, Camera } from 'lucide-vue-next'
import { usePwa } from '@/composables/usePwa'
import { useCoinOptions } from '@/composables/useCoinOptions'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import { materialLabel } from '@/utils/format'
import { t } from '@/i18n'
import { categoryLabel, eraLabel } from '@/i18n/coinVocabulary'
import { COIN_CIRCULATIONS, COIN_FINISHES, COIN_GRADES, COIN_PRESENTATIONS, COIN_RARITIES, circulationLabel, coinGradeLabel, deriveCoinYear, presentationLabel, rarityLabel } from '@/utils/coinAttributes'
import { formatNumistaYear } from '@/utils/numistaCatalog'
import { collectionCurrencySymbol } from '@/utils/collectionCurrency'

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

// Coins saved before these fields existed may not carry them; edit them as ''.
const countryModel = computed({
  get: () => props.form.country ?? '',
  set: (value: string) => { props.form.country = value },
})
const presentationModel = computed({
  get: () => props.form.presentation ?? '',
  set: (value: string) => { props.form.presentation = value },
})
const gradeModel = computed({
  get: () => props.form.grade ?? '',
  set: (value: string) => { props.form.grade = value },
})
// A grade written otherwise (an import, the AI, an older edit: "S", "MS-65") stays selectable.
const customGrade = computed(() => {
  const grade = props.form.grade?.trim() ?? ''
  return grade && !([...COIN_GRADES, ...COIN_FINISHES] as readonly string[]).includes(grade) ? grade : ''
})
// A coin unless it says it is a banknote (older items have no kind).
const kindModel = computed({
  get: () => props.form.kind ?? 'coin',
  set: (value: CoinKind) => { props.form.kind = value },
})

const circulationModel = computed({
  get: () => props.form.circulation ?? '',
  set: (value: string) => { props.form.circulation = value },
})
const rarityModel = computed({
  get: () => props.form.rarity ?? '',
  set: (value: string) => { props.form.rarity = value },
})

// An empty year is worked out from "Date or range" by the server; show which.
const yearPlaceholder = computed(() => {
  const derived = deriveCoinYear(props.form.dateRange)
  return derived === undefined ? t('coinForm.yearPlaceholder') : t('coinForm.yearAuto', { year: formatNumistaYear(derived) })
})

function setYear(value: string) {
  const year = Math.trunc(Number(value))
  props.form.year = value.trim() === '' || !Number.isFinite(year) || year === 0 ? null : year
}

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
function handleSubmit() {
  emit('submit')
}

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

onMounted(async () => {
  // Load coin property options from settings
  loadOptions()

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
})
</script>
