<template>
  <section class="card grid min-w-0 gap-4 p-4" aria-live="polite">
    <button type="button" class="btn btn-ghost btn-sm w-fit" @click="$emit('back')">← {{ t('catalog.backToResults') }}</button>

    <p v-if="loading" class="m-0 text-body text-text-secondary" role="status">{{ text('loadingType') }}</p>
    <p v-else-if="loadFailed" class="m-0 text-body text-text-secondary" role="status">{{ text('typeFailed') }}</p>
    <p v-else-if="result && result.status !== 'ok'" class="m-0 text-body text-text-secondary" role="status">{{ numistaStatusMessage(result.status) }}</p>

    <template v-else-if="type">
      <div class="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-start">
        <div v-if="type.photos.length" class="flex shrink-0 gap-3">
          <CatalogPhoto
            v-for="photo in type.photos"
            :key="photo.side"
            :url="photo.thumbnail"
            :alt="t(photo.side === 'obverse' ? 'catalog.obversePhoto' : 'catalog.reversePhoto')"
            :credit="catalogPhotoCredit(photo)"
            size="7rem"
          />
        </div>
        <div class="min-w-0">
          <h2 class="m-0 text-xl font-medium text-heading [overflow-wrap:anywhere]">{{ type.title }}</h2>
          <p class="mb-1 mt-1 text-body text-text-secondary">{{ [type.issuer.name, numistaYearRange(type.minYear, type.maxYear)].filter(Boolean).join(' · ') }}</p>
          <a :href="type.url" target="_blank" rel="noopener noreferrer" class="text-body text-gold underline">{{ t('coinDetail.marketValue.viewOnNumista') }}</a>
        </div>
      </div>

      <dl class="m-0 grid grid-cols-[minmax(7rem,auto)_minmax(0,1fr)] gap-x-4 gap-y-1 text-body">
        <template v-for="row in detailRows" :key="row.label">
          <dt class="text-text-secondary">{{ row.label }}</dt>
          <dd class="m-0 whitespace-pre-line text-text-primary [overflow-wrap:anywhere]">{{ row.value }}</dd>
        </template>
      </dl>

      <label v-if="issues.length > 1" class="flex flex-col gap-1">
        <span class="form-label">{{ t('coinDetail.marketValue.issue') }}</span>
        <select class="form-select" :value="issueId" @change="selectIssue(Number(($event.target as HTMLSelectElement).value))">
          <option v-for="issue in issues" :key="issue.id" :value="issue.id">{{ numistaIssueLabel(issue) }}</option>
        </select>
      </label>

      <div class="grid gap-2">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h3 class="m-0 text-base font-medium text-text-primary">{{ text('chooseGrade') }}</h3>
          <label class="inline-flex items-center gap-2 text-body text-text-secondary">
            <span>{{ t('coinDetail.marketValue.currency') }}</span>
            <select class="form-select w-auto py-1 text-body" :value="currency" @change="selectCurrency(($event.target as HTMLSelectElement).value)">
              <option v-for="code in numistaCurrencies()" :key="code" :value="code">{{ code }}</option>
            </select>
          </label>
        </div>
        <p v-if="valueLoading && !marketValue" class="m-0 text-body text-text-secondary" role="status">{{ t('coinDetail.marketValue.loading') }}</p>
        <template v-else-if="marketValue">
          <template v-if="marketValue.status === 'ok'">
            <p class="m-0 text-chip text-text-muted">{{ text('chooseGradeHint') }}</p>
            <NumistaPriceGrid
              :class="{ 'opacity-60': valueLoading }"
              :prices="marketValue.prices"
              :currency="marketValue.currency"
              :highlight-grade="grade"
              selectable
              @select="grade = $event as NumistaGrade | null"
            />
          </template>
          <p v-else class="m-0 text-body text-text-secondary" role="status">{{ numistaStatusMessage(marketValue.status) }}</p>
        </template>
        <label v-if="selectedPrice" class="mt-1 inline-flex items-center gap-2 text-body text-text-primary">
          <input v-model="fillCurrentValue" type="checkbox" class="h-4 w-4 accent-[var(--accent-gold)]" />
          {{ text('fillCurrentValue', { currency: currencyName(collectionCurrency) }) }}
        </label>
      </div>

      <div class="grid gap-2">
        <h3 class="m-0 text-base font-medium text-text-primary" data-testid="real-photos-title">{{ category === 'banknote' ? t('photos.banknote.title') : t('photos.title') }}</h3>
        <p class="m-0 text-chip text-text-muted">{{ category === 'banknote' ? t('photos.banknote.chooseHint') : t('photos.chooseHint') }}</p>
        <RealPhotoPicker v-model="photos" :initial-query="photoQuery" :numista-type-id="type.id" :details="photoDetails" :kind="category" />
      </div>

      <div class="flex flex-wrap justify-end gap-2 border-t border-border-subtle pt-3">
        <button type="button" class="btn btn-secondary" :disabled="using" @click="use(true)">{{ t('catalog.addToWishlist') }}</button>
        <button type="button" class="btn btn-primary" :disabled="using" @click="use(false)">{{ t('catalog.addToCollection') }}</button>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
// One coin type from the Numista catalogue: what will be filled in, the issue
// and grade of the user's coin, and Numista's price for that grade.
import { computed, ref, watch } from 'vue'
import {
  getNumistaCatalogType,
  getNumistaTypeMarketValue,
  type CoinMarketValue,
  type NumistaCatalogTypeResult,
} from '@/api/client'
import { intlLocale, t } from '@/i18n'
import CatalogPhoto from '@/components/catalog/CatalogPhoto.vue'
import NumistaPriceGrid from '@/components/coin/NumistaPriceGrid.vue'
import RealPhotoPicker from '@/components/photos/RealPhotoPicker.vue'
import type { RealPhotoChoice } from '@/utils/coinPhotos'
import { catalogPhotoCredit, numistaIssueLabel, numistaLang, numistaYearRange } from '@/utils/numistaCatalog'
import type { CatalogSelection } from '@/utils/numistaCatalogPrefill'
import type { NumistaGrade } from '@/utils/numistaGrades'
import { numistaCurrencies, numistaStatusMessage, readPreferredCurrency, savePreferredCurrency } from '@/utils/numistaMarketValue'
import { collectionCurrency, currencyName } from '@/utils/collectionCurrency'

const props = withDefaults(defineProps<{
  typeId: number
  /** Year typed in the search, used to preselect the matching issue. */
  yearHint?: number
  /** A banknote type: its details are a banknote's. */
  category?: 'coin' | 'banknote'
}>(), { yearHint: undefined, category: 'coin' })

// A banknote search speaks of banknotes, a coin search of coins.
type CatalogText = 'intro' | 'query' | 'queryPlaceholder' | 'yearHint' | 'noResults' | 'resultCount' | 'obversePhoto' | 'reversePhoto'
  | 'loadingType' | 'typeFailed' | 'chooseGrade' | 'chooseGradeHint' | 'fillCurrentValue'
function text(key: CatalogText, params?: Parameters<typeof t>[1]): string {
  return props.category === 'banknote' ? t(`catalog.banknote.${key}`, params) : t(`catalog.${key}`, params)
}

const emit = defineEmits<{
  use: [selection: CatalogSelection]
  back: []
}>()

const result = ref<NumistaCatalogTypeResult | null>(null)
const loading = ref(false)
const loadFailed = ref(false)
const issueId = ref<number | undefined>(undefined)
const currency = ref(readPreferredCurrency())
const marketValue = ref<CoinMarketValue | null>(null)
const valueLoading = ref(false)
const grade = ref<NumistaGrade | null>(null)
const fillCurrentValue = ref(true)
const using = ref(false)
const photos = ref<RealPhotoChoice>({})
let valueSeq = 0

const type = computed(() => (result.value?.status === 'ok' ? result.value.type ?? null : null))
const issues = computed(() => result.value?.issues ?? [])
const selectedIssue = computed(() => issues.value.find((issue) => issue.id === issueId.value))
const selectedPrice = computed(() => marketValue.value?.prices.find((price) => price.grade === grade.value))
// The coin's year: the one asked for, else the type's when struck in a single year.
const photoYear = computed(() => {
  const detail = type.value
  return props.yearHint ?? (detail?.minYear !== undefined && detail.minYear === detail.maxYear ? detail.minYear : undefined)
})
// Title plus the year when it is known, e.g. "50 Escudos (Pedro Alvares Cabral) 1968".
const photoQuery = computed(() => {
  const detail = type.value
  if (!detail) return ''
  const year = photoYear.value
  return [detail.title, year !== undefined && year > 0 ? String(year) : ''].filter(Boolean).join(' ')
})
// Lets the search find a €2 commemorative's ECB image.
const photoDetails = computed(() => ({
  year: photoYear.value,
  country: type.value?.issuer.name,
  denomination: type.value?.denomination ?? type.value?.title,
}))

function measure(value: number | undefined, unit: string): string {
  return value ? `${new Intl.NumberFormat(intlLocale(), { maximumFractionDigits: 2 }).format(value)} ${unit}` : ''
}

function side(description?: string, lettering?: string): string {
  return [description, lettering ? `“${lettering}”` : ''].filter(Boolean).join('\n')
}

const detailRows = computed(() => {
  const detail = type.value
  if (!detail) return []
  if (props.category === 'banknote') {
    const size = detail.diameterMm && detail.heightMm ? `${measure(detail.diameterMm, '').trim()} × ${measure(detail.heightMm, 'mm')}` : measure(detail.diameterMm, 'mm')
    return [
      { label: t('banknoteForm.faceValue'), value: detail.denomination ?? '' },
      { label: t('banknoteForm.issuingBank'), value: detail.issuingEntity || detail.rulers.join(', ') },
      { label: t('banknoteForm.series'), value: detail.series ?? '' },
      { label: t('banknoteForm.material'), value: detail.composition ?? '' },
      { label: t('banknoteForm.dimensions'), value: size },
      { label: t('banknoteForm.printer'), value: (detail.printers ?? []).join(', ') },
      { label: t('catalog.banknote.obverse'), value: side(detail.obverse.description, detail.obverse.lettering) },
      { label: t('catalog.banknote.reverse'), value: side(detail.reverse.description, detail.reverse.lettering) },
      { label: t('banknoteForm.watermark'), value: detail.watermark ?? '' },
      { label: t('catalog.references'), value: detail.references.map((ref) => `${ref.catalogue} ${ref.number}`).join(', ') },
    ].filter((row) => row.value)
  }
  return [
    { label: t('coinFields.denomination'), value: detail.denomination ?? '' },
    { label: t('coinFields.ruler'), value: detail.rulers.join(', ') },
    { label: t('catalog.composition'), value: detail.composition ?? '' },
    { label: t('coinFields.weight'), value: measure(detail.weightGrams, 'g') },
    { label: t('coinFields.diameter'), value: measure(detail.diameterMm, 'mm') },
    { label: t('catalog.thickness'), value: measure(detail.thicknessMm, 'mm') },
    { label: t('catalog.shape'), value: detail.shape ?? '' },
    { label: t('coinFields.mint'), value: detail.mints.join(', ') },
    { label: t('catalog.obverse'), value: side(detail.obverse.description, detail.obverse.lettering) },
    { label: t('catalog.reverse'), value: side(detail.reverse.description, detail.reverse.lettering) },
    { label: t('catalog.edge'), value: detail.edge.description ?? '' },
    { label: t('catalog.references'), value: detail.references.map((ref) => `${ref.catalogue} ${ref.number}`).join(', ') },
  ].filter((row) => row.value)
})

async function loadMarketValue() {
  const seq = ++valueSeq
  valueLoading.value = true
  try {
    const res = await getNumistaTypeMarketValue(props.typeId, { currency: currency.value, issueId: issueId.value, year: issueId.value ? undefined : props.yearHint })
    if (seq !== valueSeq) return
    marketValue.value = res.data
    if (!issueId.value && res.data.selectedIssueId) issueId.value = res.data.selectedIssueId
  } catch {
    if (seq === valueSeq) marketValue.value = null
  } finally {
    if (seq === valueSeq) valueLoading.value = false
  }
}

watch(() => props.typeId, async (typeId) => {
  result.value = null
  marketValue.value = null
  issueId.value = undefined
  grade.value = null
  photos.value = {}
  loading.value = true
  loadFailed.value = false
  try {
    const res = await getNumistaCatalogType(typeId, numistaLang())
    if (props.typeId !== typeId) return
    result.value = res.data
  } catch {
    loadFailed.value = true
  } finally {
    loading.value = false
  }
  if (type.value) void loadMarketValue()
}, { immediate: true })

function selectIssue(id: number) {
  issueId.value = id
  void loadMarketValue()
}

function selectCurrency(code: string) {
  currency.value = code
  savePreferredCurrency(code)
  void loadMarketValue()
}

// The coin form stores current value in the collection currency, so the price
// used for it is Numista's estimate in that currency whatever currency is being viewed.
async function collectionPrice(forGrade: NumistaGrade): Promise<number | undefined> {
  if (marketValue.value?.currency === collectionCurrency.value) return selectedPrice.value?.price
  try {
    const res = await getNumistaTypeMarketValue(props.typeId, { currency: collectionCurrency.value, issueId: issueId.value })
    return res.data.prices.find((price) => price.grade === forGrade)?.price
  } catch {
    return undefined
  }
}

async function use(wishlist: boolean) {
  if (!type.value || using.value) return
  using.value = true
  try {
    const currentValue = fillCurrentValue.value && grade.value && selectedPrice.value ? await collectionPrice(grade.value) : undefined
    emit('use', {
      type: type.value,
      issue: selectedIssue.value,
      grade: grade.value ?? undefined,
      currentValue,
      wishlist,
      photos: photos.value,
    })
  } finally {
    using.value = false
  }
}
</script>
