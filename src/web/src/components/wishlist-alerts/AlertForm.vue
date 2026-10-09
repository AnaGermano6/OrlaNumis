<template>
  <form class="grid gap-3" @submit.prevent="submit">
    <p class="m-0 text-body text-text-muted">{{ t('wishlistAlerts.form.intro') }}</p>
    <label class="grid gap-1 text-base text-text-secondary">
      {{ t('coinFields.name') }}
      <input
        v-model.trim="draft.name"
        class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
        required
        maxlength="200"
      />
    </label>
    <div class="grid gap-3 md:grid-cols-2">
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('wishlistAlerts.form.rulerOrIssuer') }}
        <input v-model.trim="draft.criteria.rulerOrIssuer" class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" maxlength="200" />
      </label>
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('wishlistAlerts.form.coinType') }}
        <input v-model.trim="draft.criteria.coinType" class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" maxlength="200" />
      </label>
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('coinFields.mint') }}
        <select v-model="mintLocationIdModel" class="form-select focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" :disabled="mintLocationsLoading">
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
        <p v-if="mintLocationError" class="mt-1 text-body text-text-secondary">{{ mintLocationError }}</p>
        <p v-else-if="draft.criteria.mint && draft.criteria.mintLocationId == null" class="mt-1 text-body text-text-secondary">
          {{ t('wishlistAlerts.form.legacyMint', { mint: draft.criteria.mint }) }}
        </p>
      </label>
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('coinFields.material') }}
        <input v-model.trim="draft.criteria.material" class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" maxlength="100" />
      </label>
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('wishlistAlerts.form.gradeOrCondition') }}
        <input v-model.trim="draft.criteria.gradeOrCondition" class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" maxlength="200" />
      </label>
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('wishlistAlerts.form.keywords') }}
        <input v-model.trim="draft.criteria.keywords" class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" maxlength="500" />
      </label>
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('wishlistAlerts.form.dateFrom') }}
        <input v-model.number="draft.criteria.dateFrom" class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" type="number" />
      </label>
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('wishlistAlerts.form.dateTo') }}
        <input v-model.number="draft.criteria.dateTo" class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" type="number" />
      </label>
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('wishlistAlerts.form.priceMin') }}
        <input v-model.number="draft.criteria.priceMin" class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" type="number" min="0" step="0.01" />
      </label>
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('wishlistAlerts.form.priceMax') }}
        <input v-model.number="draft.criteria.priceMax" class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" type="number" min="0" step="0.01" />
      </label>
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('wishlistAlerts.form.currency') }}
        <input v-model.trim="draft.criteria.currency" class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" maxlength="3" />
      </label>
      <label class="grid gap-1 text-base text-text-secondary">
        {{ t('wishlistAlerts.form.cadence') }}
        <select v-model="draft.cadence" class="form-select focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]">
          <option value="manual">{{ alertCadenceLabel('manual') }}</option>
          <option value="daily">{{ alertCadenceLabel('daily') }}</option>
          <option value="weekly">{{ alertCadenceLabel('weekly') }}</option>
          <option value="monthly">{{ alertCadenceLabel('monthly') }}</option>
        </select>
      </label>
    </div>
    <p class="m-0 text-body text-text-muted">{{ t('wishlistAlerts.form.cadenceHint') }}</p>
    <label class="grid gap-1 text-base text-text-secondary">
      {{ t('wishlistAlerts.form.sourceDomains') }}
      <input
        v-model.trim="sourceFiltersText"
        class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
        placeholder="vcoins.com, ma-shops.com"
      />
    </label>
    <label class="grid gap-1 text-base text-text-secondary">
      {{ t('wishlistAlerts.form.dealerPreference') }}
      <input v-model.trim="draft.criteria.dealerPreference" class="form-input focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" maxlength="500" />
    </label>
    <label class="grid gap-1 text-base text-text-secondary">
      {{ t('coinFields.notes') }}
      <textarea
        v-model.trim="draft.criteria.notes"
        class="form-input min-h-20 resize-y focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
        maxlength="5000"
      />
    </label>
    <label class="flex items-center gap-2 text-base text-text-secondary">
      <input
        v-model="draft.isActive"
        class="h-4 w-4 rounded-sm border border-border-subtle bg-input accent-[var(--accent-gold)] focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]"
        type="checkbox"
      />
      {{ t('wishlistAlerts.summary.active') }}
    </label>
    <p v-if="error" class="m-0 text-body text-bronze">{{ error }}</p>
    <div class="flex flex-wrap gap-2">
      <button class="btn btn-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" type="submit" :disabled="!!error || saving">{{ saving ? t('common.saving') : t('wishlistAlerts.form.save') }}</button>
      <button class="btn btn-secondary focus:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent-gold)]" type="button" @click="$emit('cancel')">{{ t('common.cancel') }}</button>
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
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { getMintLocations, type MintLocationsResponse } from '@/api/client'
import type { MintLocation, WishlistSearchAlert, WishlistSearchAlertInput } from '@/types'
import CreateMintModal from '@/components/CreateMintModal.vue'
import { alertCadenceLabel } from '@/utils/format'
import { t } from '@/i18n'

const props = defineProps<{ alert?: WishlistSearchAlert | null; saving?: boolean }>()
const emit = defineEmits<{ save: [value: WishlistSearchAlertInput]; cancel: [] }>()

function unwrapMintLocations(data: MintLocationsResponse): MintLocation[] {
  return Array.isArray(data) ? data : data.mintLocations ?? []
}

const blank = (): WishlistSearchAlertInput => ({
  name: '',
  criteria: {
    rulerOrIssuer: '',
    coinType: '',
    dateFrom: null,
    dateTo: null,
    mint: '',
    mintLocationId: null,
    material: '',
    gradeOrCondition: '',
    priceMin: null,
    priceMax: null,
    currency: 'USD',
    dealerPreference: '',
    sourceFilters: [],
    keywords: '',
    notes: '',
  },
  cadence: 'manual',
  isActive: true,
})

const draft = reactive<WishlistSearchAlertInput>(blank())
const sourceFiltersText = ref('')
const mintLocations = ref<MintLocation[]>([])
const mintLocationsLoading = ref(false)
const mintLocationError = ref('')
const showCreateMintModal = ref(false)
const pendingMintName = ref('')

const myMintLocations = computed(() => mintLocations.value.filter((m) => m.userId != null))
const globalMintLocations = computed(() => mintLocations.value.filter((m) => m.userId == null))

const mintLocationIdModel = computed({
  get: () => draft.criteria.mintLocationId == null ? '' : String(draft.criteria.mintLocationId),
  set: (value: string) => {
    if (value === '__create__') {
      pendingMintName.value = draft.criteria.mint.trim()
      showCreateMintModal.value = true
      return
    }
    draft.criteria.mintLocationId = value === '' ? null : Number(value)
    const selected = mintLocations.value.find((m) => String(m.id) === value)
    draft.criteria.mint = selected ? selected.displayName : draft.criteria.mint
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
  draft.criteria.mintLocationId = mintLocation.id
  draft.criteria.mint = mintLocation.displayName
}

watch(() => props.alert, (alert) => {
  Object.assign(draft, blank())
  if (!alert) {
    sourceFiltersText.value = ''
    return
  }
  draft.name = alert.name
  draft.criteria = {
    rulerOrIssuer: alert.rulerOrIssuer,
    coinType: alert.coinType,
    dateFrom: alert.dateFrom,
    dateTo: alert.dateTo,
    mint: alert.mint,
    mintLocationId: alert.mintLocationId,
    material: alert.material,
    gradeOrCondition: alert.gradeOrCondition,
    priceMin: alert.priceMin,
    priceMax: alert.priceMax,
    currency: alert.currency || 'USD',
    dealerPreference: alert.dealerPreference,
    sourceFilters: [...alert.sourceFilters],
    keywords: alert.keywords,
    notes: alert.notes,
  }
  draft.cadence = alert.cadence
  draft.isActive = alert.isActive
  sourceFiltersText.value = alert.sourceFilters.join(', ')
}, { immediate: true })

onMounted(() => {
  void loadMintLocations()
})

const error = computed(() => {
  const c = draft.criteria
  const hasCriteria = [c.rulerOrIssuer, c.coinType, c.mint, c.material, c.gradeOrCondition, c.dealerPreference, c.keywords, sourceFiltersText.value].some(Boolean) || c.dateFrom != null || c.dateTo != null || c.priceMin != null || c.priceMax != null
  if (!hasCriteria) return t('wishlistAlerts.form.needCriterion')
  if (c.priceMin != null && c.priceMax != null && c.priceMin > c.priceMax) return t('wishlistAlerts.form.priceOrder')
  if (c.dateFrom != null && c.dateTo != null && c.dateFrom > c.dateTo) return t('wishlistAlerts.form.dateOrder')
  if (!/^[A-Za-z]{3}$/.test(c.currency)) return t('wishlistAlerts.form.currencyCode')
  return ''
})

function submit() {
  draft.criteria.sourceFilters = sourceFiltersText.value.split(',').map((s) => s.trim()).filter(Boolean)
  draft.criteria.currency = draft.criteria.currency.toUpperCase()
  emit('save', JSON.parse(JSON.stringify(draft)))
}
</script>
