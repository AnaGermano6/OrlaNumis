<template>
  <fieldset class="m-0 space-y-3 border-0 p-0" data-testid="set-filters">
    <legend class="sr-only">{{ t('sets.filters.title') }}</legend>
    <div class="grid gap-3 sm:grid-cols-2">
      <div class="form-group m-0">
        <label for="filterCountry" class="form-label mb-1 block">{{ t('sets.filters.country') }}</label>
        <CountrySelect
          id="filterCountry"
          :model-value="modelValue.country"
          :options="countryChoices"
          :label="t('sets.filters.country')"
          @update:model-value="update({ country: $event })"
        />
      </div>
      <div class="form-group m-0">
        <label for="filterDenomination" class="form-label mb-1 block">{{ t('sets.filters.denomination') }}</label>
        <input
          id="filterDenomination"
          class="form-input w-full"
          list="filterDenominationOptions"
          autocomplete="off"
          :placeholder="t('sets.filters.denominationPlaceholder')"
          :value="modelValue.denomination"
          @input="typeDenomination(($event.target as HTMLInputElement).value)"
        />
        <datalist id="filterDenominationOptions">
          <option v-for="value in denominations" :key="value" :value="value" />
        </datalist>
      </div>
      <div class="form-group m-0">
        <span class="form-label mb-1 block" id="filterYearsLabel">{{ t('sets.filters.years') }}</span>
        <div class="flex items-center gap-2" role="group" aria-labelledby="filterYearsLabel">
          <input
            id="filterYearFrom"
            type="number"
            inputmode="numeric"
            min="1"
            max="9999"
            class="form-input w-full"
            :aria-label="t('sets.filters.yearFrom')"
            :placeholder="t('sets.filters.yearFrom')"
            :value="modelValue.yearFrom ?? ''"
            @input="update({ yearFrom: year(($event.target as HTMLInputElement).value) })"
          />
          <span class="text-text-muted" aria-hidden="true">–</span>
          <input
            id="filterYearTo"
            type="number"
            inputmode="numeric"
            min="1"
            max="9999"
            class="form-input w-full"
            :aria-label="t('sets.filters.yearTo')"
            :placeholder="t('sets.filters.yearTo')"
            :value="modelValue.yearTo ?? ''"
            @input="update({ yearTo: year(($event.target as HTMLInputElement).value) })"
          />
        </div>
      </div>
      <div class="form-group m-0">
        <label for="filterPresentation" class="form-label mb-1 block">{{ t('sets.filters.presentation') }}</label>
        <select id="filterPresentation" class="form-input w-full" :value="modelValue.presentation" @change="update({ presentation: ($event.target as HTMLSelectElement).value })">
          <option value="">{{ t('sets.filters.any') }}</option>
          <option v-for="value in COIN_PRESENTATIONS" :key="value" :value="value">{{ presentationLabel(value) }}</option>
        </select>
      </div>
    </div>
    <div>
      <span class="form-label mb-1 block">{{ t('sets.filters.classification') }}</span>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="option in CLASSIFICATIONS"
          :key="option || 'all'"
          type="button"
          class="chip border"
          :class="option === modelValue.classification ? 'border-gold text-gold' : 'border-border-subtle'"
          :aria-pressed="option === modelValue.classification"
          :data-testid="`filter-classification-${option || 'all'}`"
          @click="update({ classification: option })"
        >
          {{ option ? classificationLabel(option) : t('sets.filters.allKinds') }}
        </button>
      </div>
    </div>
    <p class="m-0 rounded-sm border border-border-subtle bg-card px-3 py-2 text-body text-text-secondary" aria-live="polite" data-testid="filter-preview">
      <template v-if="preview === null">{{ t('sets.filters.counting') }}</template>
      <template v-else>{{ t('sets.filters.preview', { count: preview.coinCount }) }}<template v-if="preview.coinCount"> · {{ formatCurrency(preview.totalValue) }}</template></template>
    </p>
  </fieldset>
</template>

<script setup lang="ts">
// Simple filters for a set, with how many coins of the collection they
// keep: the countries are the collection's, and the face values it has are
// offered as one types.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { getCoinCountries, getSuggestions, previewSmartSet } from '@/api/client'
import CountrySelect, { type CountrySelectOption } from '@/components/CountrySelect.vue'
import type { SmartSetPreview } from '@/types'
import { intlLocale, t } from '@/i18n'
import { COIN_PRESENTATIONS, presentationLabel } from '@/utils/coinAttributes'
import { formatCurrency } from '@/utils/format'
import { COIN_CLASSIFICATIONS, classificationLabel, filtersToCriteria, type CoinClassification, type SimpleSetFilters } from '@/utils/setFilters'

const CLASSIFICATIONS: Array<CoinClassification | ''> = ['', ...COIN_CLASSIFICATIONS]
/** Wait this long after a change before counting or suggesting. */
const DELAY = 300

const props = defineProps<{ modelValue: SimpleSetFilters }>()
const emit = defineEmits<{ 'update:modelValue': [value: SimpleSetFilters] }>()

const countries = ref<string[]>([])
const denominations = ref<string[]>([])
const preview = ref<SmartSetPreview | null>(null)
let previewTimer: ReturnType<typeof setTimeout> | undefined
let suggestTimer: ReturnType<typeof setTimeout> | undefined
let previewRequest = 0

// The collection's countries, each with its flag (drawn ones too, as in the
// galleries); the one chosen stays even when not among them.
const countryChoices = computed((): CountrySelectOption[] => {
  const chosen = props.modelValue.country
  const names = chosen && !countries.value.includes(chosen) ? [chosen, ...countries.value] : countries.value
  return [
    { value: '', label: t('sets.filters.any') },
    ...names.map((name) => ({ value: name, label: name, country: name })),
  ]
})

function update(change: Partial<SimpleSetFilters>) {
  emit('update:modelValue', { ...props.modelValue, ...change })
}

function year(value: string): number | null {
  const parsed = Number.parseInt(value, 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null
}

function typeDenomination(value: string) {
  update({ denomination: value })
  clearTimeout(suggestTimer)
  suggestTimer = setTimeout(() => { void suggest(value.trim()) }, DELAY)
}

async function suggest(query: string) {
  try {
    denominations.value = (await getSuggestions('denomination', query)).data ?? []
  } catch {
    denominations.value = []
  }
}

async function count() {
  const request = ++previewRequest
  try {
    const { data } = await previewSmartSet(filtersToCriteria(props.modelValue))
    if (request === previewRequest) preview.value = data
  } catch {
    if (request === previewRequest) preview.value = { coinIds: [], coinCount: 0, totalValue: 0 }
  }
}

watch(() => props.modelValue, () => {
  preview.value = null
  clearTimeout(previewTimer)
  previewTimer = setTimeout(() => { void count() }, DELAY)
}, { deep: true })

onMounted(async () => {
  void count()
  void suggest('')
  try {
    countries.value = (await getCoinCountries('coin')).data.map((entry) => entry.country).filter(Boolean)
      .sort((a, b) => a.localeCompare(b, intlLocale()))
  } catch {
    countries.value = []
  }
})

onBeforeUnmount(() => {
  clearTimeout(previewTimer)
  clearTimeout(suggestTimer)
  previewRequest++
})
</script>
