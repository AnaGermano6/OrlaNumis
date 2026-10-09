<template>
  <section id="market-value" class="mb-6 scroll-mt-24" aria-labelledby="market-value-heading">
    <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
      <h3 id="market-value-heading" class="m-0 text-base font-medium text-text-primary">{{ t('coinDetail.marketValue.title') }}</h3>
      <label v-if="numistaTypeId" class="inline-flex items-center gap-2 text-body text-text-secondary">
        <span>{{ t('coinDetail.marketValue.currency') }}</span>
        <select class="form-select w-auto py-1 text-body" :value="currency" @change="selectCurrency(($event.target as HTMLSelectElement).value)">
          <option v-for="code in NUMISTA_CURRENCIES" :key="code" :value="code">{{ code }}</option>
        </select>
      </label>
    </div>

    <p v-if="!numistaTypeId" class="m-0 text-body text-text-secondary">{{ t('coinDetail.marketValue.notLinked') }}</p>

    <template v-else>
      <div v-if="value?.catalogPhotos.length" class="mb-3 flex flex-wrap gap-3">
        <CatalogPhoto
          v-for="photo in value.catalogPhotos"
          :key="photo.side"
          :url="photo.thumbnail"
          :alt="t(photo.side === 'obverse' ? 'catalog.obversePhoto' : 'catalog.reversePhoto')"
          :credit="catalogPhotoCredit(photo)"
          size="5.5rem"
          @unavailable="unavailablePhotos.add(photo.side)"
        />
        <p v-if="unavailablePhotos.size < value.catalogPhotos.length" class="m-0 basis-full text-chip text-text-muted">{{ t('catalog.referencePhotos') }}</p>
      </div>

      <label v-if="value && value.issues.length > 1" class="mb-3 flex flex-col gap-1">
        <span class="form-label">{{ t('coinDetail.marketValue.issue') }}</span>
        <select class="form-select" :value="value.selectedIssueId" @change="selectIssue(Number(($event.target as HTMLSelectElement).value))">
          <option v-for="issue in value.issues" :key="issue.id" :value="issue.id">{{ numistaIssueLabel(issue) }}</option>
        </select>
      </label>

      <p v-if="loading && !value" class="m-0 text-body text-text-secondary" role="status">{{ t('coinDetail.marketValue.loading') }}</p>
      <p v-else-if="loadFailed" class="m-0 text-body text-text-secondary" role="status">{{ t('coinDetail.marketValue.loadFailed') }}</p>

      <template v-else-if="value">
        <NumistaPriceGrid
          v-if="value.status === 'ok'"
          :class="{ 'opacity-60': loading }"
          :prices="value.prices"
          :currency="value.currency"
          :highlight-grade="coinNumistaGrade"
        />
        <p v-else class="m-0 text-body text-text-secondary" role="status">{{ numistaStatusMessage(value.status) }}</p>

        <p v-if="value.fetchedAt || value.numistaUrl" class="mb-0 mt-2 text-chip text-text-muted">
          <span v-if="value.fetchedAt">{{ t('coinDetail.marketValue.source', { date: formatDate(value.fetchedAt) }) }}</span>
          <span v-if="value.stale"> {{ t('coinDetail.marketValue.stale') }}</span>
          <a
            v-if="value.numistaUrl"
            :href="value.numistaUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="ml-1 text-gold underline"
          >{{ t('coinDetail.marketValue.viewOnNumista') }}</a>
        </p>
      </template>
    </template>
  </section>
</template>

<script setup lang="ts">
// Numista's community price estimates per grade for the coin's linked Numista
// type. The API caches them for days, so opening a coin rarely costs a
// request against Numista's free monthly quota.
import { computed, reactive, ref, watch } from 'vue'
import { getCoinMarketValue, type CoinMarketValue } from '@/api/client'
import { intlLocale, t } from '@/i18n'
import CatalogPhoto from '@/components/catalog/CatalogPhoto.vue'
import NumistaPriceGrid from '@/components/coin/NumistaPriceGrid.vue'
import { catalogPhotoCredit, numistaIssueLabel, numistaLang } from '@/utils/numistaCatalog'
import { numistaGradeForCoinGrade } from '@/utils/numistaGrades'
import { NUMISTA_CURRENCIES, numistaStatusMessage, readPreferredCurrency, savePreferredCurrency } from '@/utils/numistaMarketValue'

const props = defineProps<{
  coinId: number
  /** Numista type from the coin's "Numista" catalogue reference; null when not linked. */
  numistaTypeId: number | null
  coinGrade?: string | null
}>()

const issueKey = (coinId: number) => `numista:marketValueIssue:${coinId}`

function savedIssue(coinId: number): number | undefined {
  try {
    const saved = Number(localStorage.getItem(issueKey(coinId)))
    return Number.isInteger(saved) && saved > 0 ? saved : undefined
  } catch {
    return undefined
  }
}

const currency = ref(readPreferredCurrency())
const issueId = ref<number | undefined>(undefined)
const value = ref<CoinMarketValue | null>(null)
const loading = ref(false)
const loadFailed = ref(false)
// Sides whose catalogue photo Numista refused, so the caption is hidden when none show.
const unavailablePhotos = reactive(new Set<string>())
let requestSeq = 0

const coinNumistaGrade = computed(() => numistaGradeForCoinGrade(props.coinGrade))

async function load() {
  if (!props.numistaTypeId) {
    value.value = null
    return
  }
  const seq = ++requestSeq
  loading.value = true
  loadFailed.value = false
  try {
    const res = await getCoinMarketValue(props.coinId, { currency: currency.value, issueId: issueId.value, lang: numistaLang() })
    if (seq === requestSeq) value.value = res.data
  } catch {
    if (seq === requestSeq) {
      value.value = null
      loadFailed.value = true
    }
  } finally {
    if (seq === requestSeq) loading.value = false
  }
}

watch(() => [props.coinId, props.numistaTypeId] as const, ([coinId]) => {
  value.value = null
  unavailablePhotos.clear()
  issueId.value = savedIssue(coinId)
  void load()
}, { immediate: true })

function selectCurrency(code: string) {
  currency.value = code
  savePreferredCurrency(code)
  void load()
}

function selectIssue(id: number) {
  issueId.value = id
  try {
    localStorage.setItem(issueKey(props.coinId), String(id))
  } catch {
    // The choice still applies for this visit.
  }
  void load()
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(intlLocale(), { day: 'numeric', month: 'short', year: 'numeric' })
}
</script>
