<template>
  <form class="mx-auto max-w-[900px]" data-testid="banknote-form" @submit.prevent="emit('submit')">
    <div class="grid gap-6 md:grid-cols-2">
      <!-- What the note is -->
      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('banknoteForm.sections.identification') }}</h2>
        <div v-if="coinId" class="form-group min-w-0">
          <label class="form-label" for="banknote-kind">{{ t('coinForm.kind') }}</label>
          <select id="banknote-kind" v-model="kindModel" class="form-select" data-testid="banknote-kind">
            <option value="coin">{{ t('coinForm.kindCoin') }}</option>
            <option value="banknote">{{ t('coinForm.kindBanknote') }}</option>
          </select>
        </div>
        <div class="form-group min-w-0">
          <label class="form-label" for="banknote-name">{{ t('coinForm.name') }}</label>
          <input id="banknote-name" v-model="form.name" class="form-input" required maxlength="200" :placeholder="t('banknoteForm.namePlaceholder')" />
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('coinFields.country') }}</label>
            <AutocompleteInput v-model="textModel('country').value" field="country" :placeholder="t('coinForm.countryPlaceholder')" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label" for="banknote-bank">{{ t('banknoteForm.issuingBank') }}</label>
            <input id="banknote-bank" v-model="textModel('ruler').value" class="form-input" maxlength="200" :placeholder="t('banknoteForm.issuingBankPlaceholder')" />
          </div>
        </div>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label">{{ t('banknoteForm.faceValue') }}</label>
            <AutocompleteInput v-model="textModel('denomination').value" field="denomination" :placeholder="t('banknoteForm.faceValuePlaceholder')" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label" for="banknote-series">{{ t('banknoteForm.series') }}</label>
            <input id="banknote-series" v-model="textModel('series').value" class="form-input" maxlength="200" :placeholder="t('banknoteForm.seriesPlaceholder')" />
          </div>
        </div>
        <div class="grid gap-3 md:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
          <div class="form-group min-w-0">
            <label class="form-label" for="banknote-date">{{ t('banknoteForm.issueDate') }}</label>
            <input id="banknote-date" v-model="textModel('dateRange').value" class="form-input" maxlength="200" :placeholder="t('banknoteForm.issueDatePlaceholder')" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label" for="banknote-year">{{ t('coinForm.year') }}</label>
            <input
              id="banknote-year"
              class="form-input"
              type="number"
              inputmode="numeric"
              step="1"
              min="1000"
              max="3000"
              :value="form.year ?? ''"
              :placeholder="yearPlaceholder"
              @input="setYear(($event.target as HTMLInputElement).value)"
            />
          </div>
        </div>
        <div class="form-group min-w-0">
          <label class="form-label" for="banknote-rarity">{{ t('coinFields.rarity') }}</label>
          <select id="banknote-rarity" v-model="textModel('rarity').value" class="form-select" data-testid="banknote-rarity">
            <option value="">{{ t('coinForm.notSpecified') }}</option>
            <option v-for="option in COIN_RARITIES" :key="option" :value="option">{{ rarityLabel(option) }}</option>
          </select>
        </div>
        <div class="form-group min-w-0">
          <label class="form-label" for="banknote-pick">{{ t('banknoteForm.pick') }}</label>
          <input id="banknote-pick" v-model.lazy="pickModel" class="form-input" maxlength="40" :placeholder="t('banknoteForm.pickPlaceholder')" aria-describedby="banknote-pick-hint" data-testid="banknote-pick" />
          <p id="banknote-pick-hint" class="mt-1 text-chip text-text-muted">{{ t('banknoteForm.pickHint') }}</p>
        </div>
      </fieldset>

      <!-- This very note -->
      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('banknoteForm.sections.thisNote') }}</h2>
        <div class="grid gap-3 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label" for="banknote-serial">{{ t('banknoteForm.serialNumber') }}</label>
            <input id="banknote-serial" v-model="textModel('serialNumber').value" class="form-input font-mono" maxlength="100" :placeholder="t('banknoteForm.serialNumberPlaceholder')" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label" for="banknote-variant">{{ t('banknoteForm.variant') }}</label>
            <select id="banknote-variant" v-model="textModel('noteVariant').value" class="form-select" data-testid="banknote-variant">
              <option v-for="variant in NOTE_VARIANTS" :key="variant" :value="variant">{{ noteVariantLabel(variant) }}</option>
            </select>
          </div>
        </div>
        <div class="form-group min-w-0">
          <label class="form-label" for="banknote-signatures">{{ t('banknoteForm.signatures') }}</label>
          <input id="banknote-signatures" v-model="textModel('signatures').value" class="form-input" maxlength="300" :placeholder="t('banknoteForm.signaturesPlaceholder')" />
        </div>
        <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto]">
          <div class="form-group min-w-0">
            <label class="form-label" for="banknote-grade">{{ t('banknoteForm.grade') }}</label>
            <select id="banknote-grade" v-model="textModel('grade').value" class="form-select" data-testid="banknote-grade">
              <option value="">{{ t('coinForm.notSpecified') }}</option>
              <option v-if="customGrade" :value="customGrade">{{ customGrade }}</option>
              <option v-for="grade in BANKNOTE_GRADES" :key="grade" :value="grade">{{ banknoteGradeLabel(grade) }}</option>
            </select>
          </div>
          <div class="form-group min-w-0">
            <label class="form-label" for="banknote-quantity">{{ t('coinFields.quantity') }}</label>
            <QuantityInput id="banknote-quantity" v-model="form.quantity" />
          </div>
        </div>
        <fieldset class="m-0 rounded-sm border border-border-subtle p-3">
          <legend class="px-1 text-chip font-medium text-text-secondary">{{ t('banknoteForm.certification') }}</legend>
          <div class="grid gap-3 md:grid-cols-2">
            <div class="form-group mb-0 min-w-0">
              <label class="form-label" for="banknote-company">{{ t('banknoteForm.gradingCompany') }}</label>
              <input id="banknote-company" v-model="certified.company" class="form-input" list="banknote-companies" maxlength="30" :placeholder="t('banknoteForm.gradingCompanyPlaceholder')" data-testid="banknote-company" />
              <datalist id="banknote-companies">
                <option v-for="company in GRADING_COMPANIES" :key="company" :value="company" />
              </datalist>
            </div>
            <div class="form-group mb-0 min-w-0">
              <label class="form-label" for="banknote-score">{{ t('banknoteForm.gradingScore') }}</label>
              <input id="banknote-score" class="form-input" type="number" inputmode="numeric" min="1" max="70" step="1" :value="certified.score ?? ''" data-testid="banknote-score" @input="setScore(($event.target as HTMLInputElement).value)" />
            </div>
          </div>
          <div class="mt-3 flex flex-wrap gap-4">
            <label class="flex items-center gap-2 text-body text-text-primary" :title="t('banknoteForm.epqHint')">
              <input v-model="certified.epq" type="checkbox" class="h-4 w-4 accent-[var(--accent-gold)]" data-testid="banknote-epq" />
              {{ t('banknoteForm.epq') }}
            </label>
            <label class="flex items-center gap-2 text-body text-text-primary" :title="t('banknoteForm.starHint')">
              <input v-model="certified.star" type="checkbox" class="h-4 w-4 accent-[var(--accent-gold)]" />
              {{ t('banknoteForm.star') }}
            </label>
          </div>
          <div class="form-group mt-3 mb-0 min-w-0">
            <label class="form-label" for="banknote-certificate">{{ t('banknoteForm.certificationNumber') }}</label>
            <input id="banknote-certificate" v-model="textModel('certificationNumber').value" class="form-input" maxlength="60" />
          </div>
        </fieldset>
      </fieldset>

      <!-- The paper -->
      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('banknoteForm.sections.physical') }}</h2>
        <div class="form-group min-w-0">
          <label class="form-label" for="banknote-material">{{ t('banknoteForm.material') }}</label>
          <select id="banknote-material" v-model="materialModel" class="form-select" data-testid="banknote-material">
            <option v-if="customMaterial" :value="customMaterial">{{ customMaterial }}</option>
            <option v-for="material in BANKNOTE_MATERIALS" :key="material" :value="material">{{ banknoteMaterialLabel(material) }}</option>
          </select>
        </div>
        <div class="grid gap-3 grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label" for="banknote-width">{{ t('banknoteForm.widthMm') }}</label>
            <input id="banknote-width" v-model.number="form.widthMm" class="form-input" type="number" step="any" min="0" />
          </div>
          <div class="form-group min-w-0">
            <label class="form-label" for="banknote-height">{{ t('banknoteForm.heightMm') }}</label>
            <input id="banknote-height" v-model.number="form.heightMm" class="form-input" type="number" step="any" min="0" />
          </div>
        </div>
        <div class="form-group min-w-0">
          <label class="form-label" for="banknote-printer">{{ t('banknoteForm.printer') }}</label>
          <input id="banknote-printer" v-model="textModel('printer').value" class="form-input" maxlength="200" :placeholder="t('banknoteForm.printerPlaceholder')" />
        </div>
        <div class="form-group mb-0 min-w-0">
          <label class="form-label" for="banknote-watermark">{{ t('banknoteForm.watermark') }}</label>
          <input id="banknote-watermark" v-model="textModel('watermark').value" class="form-input" maxlength="500" :placeholder="t('banknoteForm.watermarkPlaceholder')" />
        </div>
      </fieldset>

      <!-- Purchase -->
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
            <AutocompleteInput v-model="textModel('purchaseLocation').value" field="purchaseLocation" :placeholder="t('coinForm.storePlaceholder')" />
          </div>
        </div>
      </fieldset>

      <!-- Front and back -->
      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5 md:col-span-2">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('banknoteForm.sections.sides') }}</h2>
        <div class="grid gap-6 md:grid-cols-2">
          <div v-for="side in SIDES" :key="side" class="flex min-w-0 flex-col gap-3">
            <div class="form-group mb-0 min-w-0">
              <label class="form-label">{{ side === 'obverse' ? t('banknoteForm.frontImage') : t('banknoteForm.backImage') }}</label>
              <div v-if="preview(side) || existing(side)" class="relative mb-2 inline-block">
                <img v-if="preview(side)" :src="preview(side)!" :alt="sideName(side)" class="h-[100px] w-[200px] rounded-sm border border-border-subtle object-contain" />
                <AuthenticatedImage v-else :media-path="existing(side)!" :alt="sideName(side)" class="h-[100px] w-[200px] rounded-sm border border-border-subtle object-contain" />
                <button type="button" class="absolute -top-1.5 -right-1.5 flex h-[22px] w-[22px] items-center justify-center rounded-full border-0 bg-red-700 p-0 text-white" :title="t('common.remove')" :aria-label="t('banknoteForm.removeImage', { side: sideName(side) })" @click="photos.clear(side)"><X :size="12" /></button>
              </div>
              <div class="flex items-center gap-2">
                <input :ref="(element) => setInput(side, element)" type="file" accept=".jpg,.jpeg,.png,.webp" class="form-input min-w-0 flex-1 text-base" :aria-label="t('banknoteForm.uploadImage', { side: sideName(side) })" @change="photos.pick(side, $event)" />
                <label v-if="isPwa" class="btn btn-secondary btn-sm inline-flex cursor-pointer items-center gap-1.5 whitespace-nowrap">
                  <Camera :size="14" /> {{ t('coinDetail.actions.photo') }}
                  <input type="file" accept="image/*" capture="environment" hidden @change="photos.pick(side, $event)" />
                </label>
              </div>
            </div>
            <div class="form-group mb-0 min-w-0">
              <label class="form-label">{{ side === 'obverse' ? t('banknoteForm.frontText') : t('banknoteForm.backText') }}</label>
              <input v-model="textModel(side === 'obverse' ? 'obverseInscription' : 'reverseInscription').value" class="form-input" maxlength="1000" />
            </div>
            <div class="form-group mb-0 min-w-0">
              <label class="form-label">{{ side === 'obverse' ? t('banknoteForm.frontDescription') : t('banknoteForm.backDescription') }}</label>
              <textarea v-model="textModel(side === 'obverse' ? 'obverseDescription' : 'reverseDescription').value" class="form-textarea" maxlength="2000" />
            </div>
          </div>
        </div>
      </fieldset>

      <fieldset class="m-0 rounded-md border border-border-subtle bg-card p-5 md:col-span-2">
        <h2 class="mb-4 font-display text-lg font-medium text-gold">{{ t('coinFields.notes') }}</h2>
        <textarea v-model="textModel('notes').value" class="form-textarea" rows="3" maxlength="5000" :aria-label="t('coinFields.notes')" :placeholder="t('coinForm.notesPlaceholder')" />
      </fieldset>
    </div>

    <div class="mt-8 flex justify-end gap-3 border-t border-border-subtle pt-6">
      <button type="button" class="btn btn-secondary" @click="$router.back()">{{ t('common.cancel') }}</button>
      <button type="submit" class="btn btn-primary" :disabled="loading">
        {{ loading ? t('common.saving') : submitLabel }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
// The form of a banknote: what it is (country, issuing bank, value, series,
// printed date, Pick number), this very note (serial number, signatures,
// replacement or specimen, grade and a grading company's certificate), its
// paper (material, size, printer, watermark), purchase, and its front and
// back. The page saves it and uploads the photos chosen here.
import { computed, reactive, ref, watch, type WritableComputedRef } from 'vue'
import { Camera, X } from 'lucide-vue-next'
import type { Coin, CoinKind } from '@/types'
import AutocompleteInput from '@/components/AutocompleteInput.vue'
import AuthenticatedImage from '@/components/AuthenticatedImage.vue'
import QuantityInput from '@/components/ui/QuantityInput.vue'
import { usePwa } from '@/composables/usePwa'
import { useSidePhotoInputs, type PhotoSide } from '@/composables/useSidePhotoInputs'
import { t } from '@/i18n'
import {
  BANKNOTE_GRADES,
  BANKNOTE_MATERIALS,
  GRADING_COMPANIES,
  NOTE_VARIANTS,
  banknoteGradeLabel,
  banknoteMaterialLabel,
  composeCertifiedGrade,
  noteVariantLabel,
  parseCertifiedGrade,
  pickNumber,
  withPickNumber,
} from '@/utils/banknoteAttributes'
import { COIN_RARITIES, deriveCoinYear, rarityLabel } from '@/utils/coinAttributes'
import { formatNumistaYear } from '@/utils/numistaCatalog'
import { collectionCurrencySymbol } from '@/utils/collectionCurrency'

const SIDES = ['obverse', 'reverse'] as const

const props = defineProps<{
  form: Partial<Coin>
  submitLabel: string
  loading?: boolean
  coinId?: number
}>()

const emit = defineEmits<{ submit: [] }>()

const { isPwa } = usePwa()
const photos = useSidePhotoInputs(() => props.form)

type TextField = 'country' | 'ruler' | 'denomination' | 'series' | 'rarity' | 'dateRange' | 'serialNumber' | 'noteVariant' | 'signatures' | 'grade'
  | 'certificationNumber' | 'printer' | 'watermark' | 'purchaseLocation' | 'obverseInscription' | 'reverseInscription'
  | 'obverseDescription' | 'reverseDescription' | 'notes'

// Items saved before a field existed may not carry it; edit it as ''.
const textModels = new Map<TextField, WritableComputedRef<string>>()
function textModel(field: TextField) {
  let model = textModels.get(field)
  if (!model) {
    model = computed({
      get: () => (props.form[field] as string | undefined) ?? '',
      set: (value: string) => { (props.form as Record<string, unknown>)[field] = value },
    })
    textModels.set(field, model)
  }
  return model
}

const kindModel = computed({
  get: () => props.form.kind ?? 'banknote',
  set: (value: CoinKind) => { props.form.kind = value },
})

const materialModel = computed({
  get: () => props.form.material || 'Paper',
  set: (value: string) => { props.form.material = value as Coin['material'] },
})
// A material from before (an item that was a coin) stays selectable.
const customMaterial = computed(() => {
  const material = props.form.material ?? ''
  return material && !(BANKNOTE_MATERIALS as readonly string[]).includes(material) ? material : ''
})

// A grade typed elsewhere (an import, an older edit) stays selectable.
const customGrade = computed(() => {
  const grade = props.form.grade?.trim() ?? ''
  return grade && !(BANKNOTE_GRADES as readonly string[]).includes(grade) ? grade : ''
})

const pickModel = computed({
  get: () => pickNumber(props.form.references),
  set: (value: string) => { props.form.references = withPickNumber(props.form.references, value) },
})

// The grading company's grade is kept as one text ("PMG 66 EPQ ★").
const certified = reactive(parseCertifiedGrade(props.form.certifiedGrade))
watch(() => props.form.certifiedGrade, (value) => {
  if (value !== composeCertifiedGrade(certified)) Object.assign(certified, parseCertifiedGrade(value))
})
watch(certified, () => { props.form.certifiedGrade = composeCertifiedGrade(certified) }, { deep: true })

function setScore(value: string) {
  const score = Math.trunc(Number(value))
  certified.score = value.trim() === '' || !Number.isFinite(score) ? null : Math.min(70, Math.max(1, score))
}

const yearPlaceholder = computed(() => {
  const derived = deriveCoinYear(props.form.dateRange)
  return derived === undefined ? t('coinForm.yearPlaceholder') : t('coinForm.yearAuto', { year: formatNumistaYear(derived) })
})

function setYear(value: string) {
  const year = Math.trunc(Number(value))
  props.form.year = value.trim() === '' || !Number.isFinite(year) || year === 0 ? null : year
}

function sideName(side: PhotoSide): string {
  return side === 'obverse' ? t('banknoteForm.front') : t('banknoteForm.back')
}

function preview(side: PhotoSide): string | null {
  return side === 'obverse' ? photos.obversePreview.value : photos.reversePreview.value
}

function existing(side: PhotoSide): string | null {
  return side === 'obverse' ? photos.existingObverse.value : photos.existingReverse.value
}

function setInput(side: PhotoSide, element: unknown) {
  const input = element instanceof HTMLInputElement ? element : null
  if (side === 'obverse') photos.obverseInput.value = input
  else photos.reverseInput.value = input
}

// Banknotes have no store card; the pages read it all the same.
const cardFile = ref<File | null>(null)

defineExpose({
  obverseFile: photos.obverseFile,
  reverseFile: photos.reverseFile,
  cardFile,
  removedObverseId: photos.removedObverseId,
  removedReverseId: photos.removedReverseId,
})
</script>
