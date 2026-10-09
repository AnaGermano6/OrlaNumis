<template>
  <div class="smart-builder">
    <div class="builder-header">
      <span class="section-label">{{ t('sets.smart.title') }}</span>
      <div class="operator-toggle">
        <span class="operator-label">{{ t('sets.smart.match') }}</span>
        <button
          type="button"
          :class="['btn btn-xs', operator === 'and' ? 'btn-primary' : 'btn-ghost']"
          @click="operator = 'and'"
        >
          {{ t('sets.smart.all') }}
        </button>
        <button
          type="button"
          :class="['btn btn-xs', operator === 'or' ? 'btn-primary' : 'btn-ghost']"
          @click="operator = 'or'"
        >
          {{ t('sets.smart.any') }}
        </button>
        <span class="operator-label">{{ t('sets.smart.ofTheseRules') }}</span>
      </div>
    </div>

    <div v-for="(rule, idx) in rules" :key="idx" class="rule-row">
      <select v-model="rule.field" class="rule-input" @change="onFieldChange(rule)">
        <optgroup :label="t('sets.smart.groups.coin')">
          <option value="material">{{ t('coinFields.material') }}</option>
          <option value="category">{{ t('collection.menu.category') }}</option>
          <option value="denomination">{{ t('coinFields.denomination') }}</option>
          <option value="ruler">{{ t('sets.smart.fields.rulerIssuer') }}</option>
          <option value="era">{{ t('coinFields.era') }}</option>
          <option value="mint">{{ t('coinFields.mint') }}</option>
          <option value="grade">{{ t('coinFields.grade') }}</option>
        </optgroup>
        <optgroup :label="t('sets.smart.groups.storage')">
          <option value="storageLocationId">{{ t('sets.smart.fields.storageLocation') }}</option>
        </optgroup>
        <optgroup :label="t('sets.smart.groups.valueDates')">
          <option value="currentValue">{{ t('sets.smart.fields.currentValue') }}</option>
          <option value="purchasePrice">{{ t('sets.smart.fields.purchasePrice') }}</option>
          <option value="purchaseDate">{{ t('sets.smart.fields.purchaseDate') }}</option>
          <option value="createdAt">{{ t('sets.smart.fields.dateAdded') }}</option>
        </optgroup>
        <optgroup :label="t('sets.smart.groups.status')">
          <option value="isWishlist">{{ t('sets.smart.fields.wishlist') }}</option>
          <option value="isSold">{{ t('sets.smart.fields.sold') }}</option>
          <option value="isPrivate">{{ t('sets.smart.fields.private') }}</option>
        </optgroup>
      </select>

      <select v-model="rule.op" class="rule-input" @change="onOpChange(rule)">
        <option v-for="op in opsForField(rule.field)" :key="op.value" :value="op.value">
          {{ op.label }}
        </option>
      </select>

      <template v-if="rule.op !== 'isNull' && rule.op !== 'isNotNull'">
        <input
          v-if="isBoolField(rule.field)"
          type="hidden"
          :value="true"
        />
        <select
          v-else-if="rule.field === 'material'"
          v-model="rule.value"
          class="rule-input"
          @change="emitCriteria"
        >
          <option value="Silver">{{ materialLabel('Silver') }}</option>
          <option value="Gold">{{ materialLabel('Gold') }}</option>
          <option value="Bronze">{{ materialLabel('Bronze') }}</option>
          <option value="Copper">{{ materialLabel('Copper') }}</option>
          <option value="Electrum">{{ materialLabel('Electrum') }}</option>
          <option value="Other">{{ materialLabel('Other') }}</option>
        </select>
        <select
          v-else-if="rule.field === 'category'"
          v-model="rule.value"
          class="rule-input"
          @change="emitCriteria"
        >
          <option v-for="cat in categoryOptions" :key="cat" :value="cat">{{ categoryLabel(cat) }}</option>
        </select>
        <select
          v-else-if="rule.field === 'era'"
          v-model="rule.value"
          class="rule-input"
          @change="emitCriteria"
        >
          <option v-for="era in eraOptions" :key="era" :value="era">{{ eraLabel(era) }}</option>
        </select>
        <select
          v-else-if="rule.field === 'mint' && (rule.op === 'eq' || rule.op === 'neq')"
          v-model="rule.value"
          class="rule-input"
          @change="emitCriteria"
        >
          <option value="">{{ t('coinForm.mintUnknown') }}</option>
          <optgroup v-if="myMintLocations.length" :label="t('coinForm.myMints')">
            <option v-for="location in myMintLocations" :key="location.id" :value="location.displayName">
              {{ location.displayName }}
            </option>
          </optgroup>
          <optgroup v-if="globalMintLocations.length" :label="t('coinForm.mints')">
            <option v-for="location in globalMintLocations" :key="location.id" :value="location.displayName">
              {{ location.displayName }}
            </option>
          </optgroup>
        </select>
        <select
          v-else-if="rule.field === 'storageLocationId'"
          v-model="rule.value"
          class="rule-input"
          @change="emitCriteria"
        >
          <option disabled value="">{{ t('sets.smart.selectLocation') }}</option>
          <option v-for="location in storageLocations" :key="location.id" :value="location.id">
            {{ location.name }}
          </option>
        </select>
        <input
          v-else-if="isNumericField(rule.field)"
          v-model="rule.value"
          type="number"
          class="rule-input rule-input--value"
          :placeholder="rule.op === 'between' ? 'min,max' : t('sets.smart.amount')"
          @input="emitCriteria"
        />
        <input
          v-else-if="isDateField(rule.field)"
          v-model="rule.value"
          type="date"
          class="rule-input rule-input--value"
          @change="emitCriteria"
        />
        <input
          v-else
          v-model="rule.value"
          class="rule-input rule-input--value"
          :placeholder="t('sets.smart.value')"
          @input="emitCriteria"
        />
      </template>

      <span v-if="validationErrors[idx]" class="rule-error">{{ validationErrors[idx] }}</span>

      <button
        v-if="rules.length > 1"
        type="button"
        class="btn btn-ghost btn-xs remove-btn"
        :title="t('sets.smart.removeRule')"
        @click="removeRule(idx)"
      >
        ✕
      </button>
    </div>

    <div class="rule-actions">
      <button type="button" class="btn btn-ghost btn-sm" @click="addRule">{{ t('sets.smart.addRule') }}</button>
      <button type="button" class="btn btn-secondary btn-sm" @click="preview" :disabled="previewLoading">
        {{ previewLoading ? t('sets.smart.checking') : t('sets.smart.preview') }}
      </button>
    </div>

    <p v-if="previewResult !== null" class="preview-result">
      <I18nT keypath="sets.smart.previewResult" :params="{ count: previewResult.coinCount }">
        <template #num><span class="preview-count">{{ previewResult.coinCount }}</span></template>
        <template #value><span class="preview-value">${{ previewResult.totalValue.toFixed(2) }}</span></template>
      </I18nT>
    </p>
    <p v-if="previewError" class="preview-error">{{ previewError }}</p>

    <!-- Suggestions -->
    <div v-if="suggestions.length" class="section-block">
      <span class="section-label">{{ t('sets.smart.suggested') }}</span>
      <div class="suggestion-chips">
        <button
          v-for="s in suggestions"
          :key="s.id"
          type="button"
          class="chip chip-sm"
          :title="s.description"
          @click="applySuggestion(s)"
        >
          {{ s.name }}
        </button>
      </div>
    </div>

    <!-- Saved templates -->
    <div v-if="savedTemplates.length || canSave" class="section-block">
      <span class="section-label">{{ t('sets.smart.savedTemplates') }}</span>
      <div class="template-row">
        <select
          v-if="savedTemplates.length"
          class="rule-input template-select"
          @change="onTemplateSelect"
        >
          <option value="">{{ t('sets.smart.applyTemplate') }}</option>
          <option v-for="tmpl in savedTemplates" :key="tmpl.id" :value="tmpl.id">{{ tmpl.name }}</option>
        </select>
        <button
          v-if="canSave"
          type="button"
          class="btn btn-ghost btn-sm"
          @click="showSaveForm = !showSaveForm"
        >
          {{ t('sets.smart.saveAsTemplate') }}
        </button>
      </div>
      <div v-if="showSaveForm" class="save-form">
        <input
          v-model="saveTemplateName"
          class="rule-input rule-input--value"
          :placeholder="t('sets.smart.templateName')"
          maxlength="80"
        />
        <button type="button" class="btn btn-primary btn-sm" :disabled="!saveTemplateName.trim() || saving" @click="saveTemplate">
          {{ saving ? t('sets.smart.saving') : t('common.save') }}
        </button>
        <button type="button" class="btn btn-ghost btn-sm" @click="showSaveForm = false">{{ t('common.cancel') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  previewSmartSet,
  getSuggestedCriteria,
  listCriteriaTemplates,
  saveCriteriaTemplate,
  getMintLocations,
  getStorageLocations,
  type MintLocationsResponse,
} from '@/api/client'
import { useCoinOptions } from '@/composables/useCoinOptions'
import { materialLabel } from '@/utils/format'
import I18nT from '@/components/I18nT.vue'
import type {
  SmartCriteriaGroup,
  SmartCriteriaRule,
  SmartCriteriaRuleOp,
  SmartSetPreview,
  SmartCriteriaTemplate,
  SuggestedSmartCriteria,
  MintLocation,
  StorageLocation,
} from '@/types'
import { t } from '@/i18n'
import { categoryLabel, eraLabel } from '@/i18n/coinVocabulary'

function unwrapMintLocations(data: MintLocationsResponse): MintLocation[] {
  return Array.isArray(data) ? data : data.mintLocations ?? []
}
const { categoryOptions, eraOptions, loadOptions: loadCoinOptions } = useCoinOptions()

const emit = defineEmits<{
  update: [criteria: SmartCriteriaGroup]
}>()

// ---- rule state ----
const operator = ref<'and' | 'or'>('and')

interface EditableRule {
  field: string
  op: SmartCriteriaRuleOp
  value: unknown
}

const rules = reactive<EditableRule[]>([{ field: 'material', op: 'eq', value: 'Silver' }])

// ---- preview state ----
const previewResult = ref<SmartSetPreview | null>(null)
const previewError = ref<string | null>(null)
const previewLoading = ref(false)

// ---- validation ----
const validationErrors = ref<string[]>([])

// ---- template / suggestion state ----
const suggestions = ref<SuggestedSmartCriteria[]>([])
const savedTemplates = ref<SmartCriteriaTemplate[]>([])
const showSaveForm = ref(false)
const saveTemplateName = ref('')
const saving = ref(false)

// ---- mint location options ----
const mintLocations = ref<MintLocation[]>([])
const myMintLocations = computed(() => mintLocations.value.filter((m) => m.userId != null))
const globalMintLocations = computed(() => mintLocations.value.filter((m) => m.userId == null))
const storageLocations = ref<StorageLocation[]>([])

// ---- lifecycle ----
onMounted(async () => {
  try {
    const [suggestRes, tmplRes] = await Promise.all([
      getSuggestedCriteria(),
      listCriteriaTemplates(),
      loadCoinOptions(),
    ])
    suggestions.value = suggestRes.data.suggestions
    savedTemplates.value = tmplRes.data.templates
  } catch {
    // non-fatal
  }
  try {
    const mintRes = await getMintLocations()
    mintLocations.value = unwrapMintLocations(mintRes.data)
  } catch {
    mintLocations.value = []
  }
  try {
    const storageRes = await getStorageLocations()
    storageLocations.value = storageRes.data.storageLocations
  } catch {
    storageLocations.value = []
  }
  emitCriteria()
})

watch([operator, rules], emitCriteria, { deep: true })

// ---- field helpers ----
const BOOL_FIELDS = new Set(['isWishlist', 'isSold', 'isPrivate'])
const NUMERIC_FIELDS = new Set(['currentValue', 'purchasePrice'])
const DATE_FIELDS = new Set(['purchaseDate', 'createdAt'])
const STORAGE_LOCATION_FIELD = 'storageLocationId'

function isBoolField(field: string) { return BOOL_FIELDS.has(field) }
function isNumericField(field: string) { return NUMERIC_FIELDS.has(field) }
function isDateField(field: string) { return DATE_FIELDS.has(field) }

interface OpOption { value: SmartCriteriaRuleOp; label: string }

function opsForField(field: string): OpOption[] {
  if (field === STORAGE_LOCATION_FIELD) {
    return [
      { value: 'eq', label: t('sets.smart.ops.equals') },
      { value: 'neq', label: t('sets.smart.ops.notEquals') },
      { value: 'isNull', label: t('sets.smart.ops.isEmpty') },
      { value: 'isNotNull', label: t('sets.smart.ops.isNotEmpty') },
    ]
  }
  if (BOOL_FIELDS.has(field)) {
    return [
      { value: 'eq', label: t('sets.smart.ops.isTrue') },
      { value: 'neq', label: t('sets.smart.ops.isFalse') },
    ]
  }
  if (NUMERIC_FIELDS.has(field)) {
    return [
      { value: 'eq', label: t('sets.smart.ops.equals') },
      { value: 'neq', label: t('sets.smart.ops.notEquals') },
      { value: 'gte', label: t('sets.smart.ops.atLeast') },
      { value: 'lte', label: t('sets.smart.ops.atMost') },
      { value: 'between', label: t('sets.smart.ops.betweenMinMax') },
      { value: 'isNull', label: t('sets.smart.ops.hasNoValue') },
      { value: 'isNotNull', label: t('sets.smart.ops.hasValue') },
    ]
  }
  if (DATE_FIELDS.has(field)) {
    return [
      { value: 'eq', label: t('sets.smart.ops.on') },
      { value: 'gte', label: t('sets.smart.ops.onOrAfter') },
      { value: 'lte', label: t('sets.smart.ops.onOrBefore') },
      { value: 'between', label: t('sets.smart.ops.betweenStartEnd') },
      { value: 'isNull', label: t('sets.smart.ops.notSet') },
      { value: 'isNotNull', label: t('sets.smart.ops.isSet') },
    ]
  }
  return [
    { value: 'eq', label: t('sets.smart.ops.equals') },
    { value: 'neq', label: t('sets.smart.ops.notEquals') },
    { value: 'contains', label: t('sets.smart.ops.contains') },
    { value: 'startsWith', label: t('sets.smart.ops.startsWith') },
    { value: 'in', label: t('sets.smart.ops.oneOf') },
    { value: 'isNull', label: t('sets.smart.ops.isEmpty') },
    { value: 'isNotNull', label: t('sets.smart.ops.isNotEmpty') },
  ]
}

function onFieldChange(rule: EditableRule) {
  // Reset op to first valid one for the new field
  const ops = opsForField(rule.field)
  if (!ops.find(o => o.value === rule.op)) {
    rule.op = ops[0]?.value ?? rule.op
  }
  // Reset value for bool fields
  if (BOOL_FIELDS.has(rule.field)) {
    rule.value = true
  } else {
    rule.value = ''
  }
  emitCriteria()
}

function onOpChange(rule: EditableRule) {
  if (rule.op === 'isNull' || rule.op === 'isNotNull') {
    rule.value = undefined
  }
  emitCriteria()
}

// ---- rule management ----
function addRule() {
  rules.push({ field: 'material', op: 'eq', value: 'Silver' })
}

function removeRule(idx: number) {
  rules.splice(idx, 1)
}

// ---- criteria building ----
function buildCriteria(): SmartCriteriaGroup {
  return {
    operator: operator.value,
    rules: rules.map(r => {
      const ruleObj: SmartCriteriaRule = { field: r.field, op: r.op }
      if (r.op !== 'isNull' && r.op !== 'isNotNull') {
        ruleObj.value = BOOL_FIELDS.has(r.field) ? (r.op === 'neq' ? false : true) : r.value
      }
      return ruleObj
    }),
  }
}

function validateRules(): boolean {
  const errors: string[] = rules.map((r) => {
    if (r.op === 'isNull' || r.op === 'isNotNull') return ''
    if (BOOL_FIELDS.has(r.field)) return ''
    const val = r.value
    if (val === '' || val === null || val === undefined) return t('sets.smart.valueRequired')
    return ''
  })
  validationErrors.value = errors
  return errors.every(e => e === '')
}

function emitCriteria() {
  validateRules()
  emit('update', buildCriteria())
}

// ---- preview ----
async function preview() {
  if (!validateRules()) return
  previewLoading.value = true
  previewError.value = null
  try {
    const res = await previewSmartSet(buildCriteria())
    previewResult.value = res.data
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    previewError.value = err?.response?.data?.error ?? t('sets.smart.previewFailed')
    previewResult.value = null
  } finally {
    previewLoading.value = false
  }
}

// ---- suggestions ----
function applySuggestion(s: SuggestedSmartCriteria) {
  applyCriteriaGroup(s.criteria)
}

function applyCriteriaGroup(criteria: SmartCriteriaGroup) {
  operator.value = criteria.operator
  rules.splice(0, rules.length)
  for (const r of criteria.rules) {
    if ('field' in r) {
      rules.push({ field: r.field, op: r.op, value: r.value ?? '' })
    }
  }
  previewResult.value = null
}

// ---- templates ----
function onTemplateSelect(e: Event) {
  const id = Number((e.target as HTMLSelectElement).value)
  if (!id) return
  const tmpl = savedTemplates.value.find(template => template.id === id)
  if (tmpl) applyCriteriaGroup(tmpl.criteria)
  ;(e.target as HTMLSelectElement).value = ''
}

const canSave = ref(true)

async function saveTemplate() {
  if (!saveTemplateName.value.trim()) return
  saving.value = true
  try {
    const tmpl = await saveCriteriaTemplate({
      name: saveTemplateName.value.trim(),
      criteria: buildCriteria(),
    })
    savedTemplates.value.push(tmpl.data)
    showSaveForm.value = false
    saveTemplateName.value = ''
  } catch {
    // ignore — user can retry
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.smart-builder {
  margin-top: 1rem;
}

.builder-header {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.operator-toggle {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  margin-left: auto;
}

.operator-label {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.rule-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem;
  margin-bottom: 0.4rem;
}

.rule-input {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  color: var(--text-primary);
  padding: 0.4rem 0.55rem;
  font-size: 0.85rem;
  min-width: 0;
}

.rule-input--value {
  flex: 1 1 120px;
}

.rule-error {
  color: var(--status-error-fg);
  font-size: 0.75rem;
  white-space: nowrap;
}

.remove-btn {
  opacity: 0.6;
  padding: 0.2rem 0.45rem;
}
.remove-btn:hover {
  opacity: 1;
}

.rule-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.25rem;
}

.preview-result {
  color: var(--text-secondary);
  font-size: 0.85rem;
  margin-top: 0.5rem;
}

.preview-count {
  font-weight: 600;
  color: var(--accent-gold);
}

.preview-value {
  font-weight: 600;
  color: var(--accent-gold);
}

.preview-error {
  color: var(--status-error-fg);
  font-size: 0.8rem;
  margin-top: 0.4rem;
}

.section-block {
  margin-top: 1rem;
  border-top: 1px solid var(--border-subtle);
  padding-top: 0.75rem;
}

.suggestion-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-top: 0.4rem;
}

.template-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.4rem;
  align-items: center;
}

.template-select {
  flex: 1 1 180px;
}

.save-form {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
  flex-wrap: wrap;
  align-items: center;
}
</style>
