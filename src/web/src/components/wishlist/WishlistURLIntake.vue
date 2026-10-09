<template>
  <section class="url-intake">
    <button
      v-if="!embedded"
      type="button"
      class="btn btn-secondary btn-sm"
      :aria-expanded="expanded"
      @click="expanded = !expanded"
    >
      <Link :size="16" /> {{ t('wishlist.add.fromUrl') }}
    </button>

    <div v-if="expanded" class="url-intake-panel" :class="{ embedded }">
      <form class="url-row" @submit.prevent="analyze">
        <label for="wishlist-url" class="sr-only">{{ t('wishlistUrl.listingUrl') }}</label>
        <input
          id="wishlist-url"
          v-model.trim="sourceURL"
          type="url"
          class="form-input min-w-0 flex-1"
          :placeholder="t('wishlistUrl.placeholder')"
          required
          :disabled="status === 'pending'"
        />
        <button type="submit" class="btn btn-primary btn-sm" :disabled="status === 'pending' || !sourceURL">
          {{ status === 'pending' ? t('wishlistUrl.analyzing') : t('wishlistUrl.analyze') }}
        </button>
        <button v-if="status === 'pending'" type="button" class="btn btn-ghost btn-sm" @click="cancel">
          {{ t('common.cancel') }}
        </button>
      </form>

      <p v-if="message" class="status-message" role="status">{{ message }}</p>

      <div v-if="analysis?.outcome === 'duplicate' && analysis.existingCoinId" class="review-actions">
        <router-link :to="`/coin/${analysis.existingCoinId}`" class="btn btn-secondary btn-sm">
          {{ t('wishlistUrl.viewExisting') }}
        </router-link>
      </div>

      <form v-else-if="analysis?.hypothesis" class="proposal" @submit.prevent="confirmCreate">
        <div class="proposal-heading">
          <div>
            <span class="section-label">{{ t('wishlistUrl.reviewProposal') }}</span>
            <p>{{ t('wishlistUrl.reviewHint') }}</p>
          </div>
          <span v-if="analysis.hypothesis.listingStatus" class="chip-sm">
            {{ analysis.hypothesis.listingStatus.value }}
          </span>
        </div>

        <div class="proposal-grid">
          <label class="form-group proposal-wide">
            <span>{{ t('coinForm.name') }}</span>
            <input v-model.trim="draft.name" class="form-input" maxlength="200" required />
          </label>
          <label class="form-group">
            <span>{{ t('collection.menu.category') }}</span>
            <select v-model="draft.category" class="form-input">
              <option v-for="category in categoryOptions" :key="category" :value="category">{{ categoryLabel(category) }}</option>
            </select>
          </label>
          <label class="form-group">
            <span>{{ t('coinFields.era') }}</span>
            <select v-model="draft.era" class="form-input">
              <option value="">{{ t('addCoin.eraUnknown') }}</option>
              <option v-for="era in eraOptions" :key="era" :value="era">{{ eraLabel(era) }}</option>
            </select>
          </label>
          <label class="form-group">
            <span>{{ t('coinFields.material') }}</span>
            <select v-model="draft.material" class="form-input">
              <option v-for="material in materialOptions" :key="material" :value="material">{{ materialLabel(material) }}</option>
            </select>
          </label>
          <label v-for="field in textFields" :key="field.key" class="form-group">
            <span>{{ t(field.labelKey) }}</span>
            <input v-model.trim="draft[field.key]" class="form-input" :maxlength="field.maxlength" />
          </label>
          <label class="form-group">
            <span>{{ t('coinFields.weightG') }}</span>
            <input v-model="draft.weightGrams" class="form-input" type="number" min="0" step="0.01" />
          </label>
          <label class="form-group">
            <span>{{ t('coinForm.diameterMm') }}</span>
            <input v-model="draft.diameterMm" class="form-input" type="number" min="0" step="0.01" />
          </label>
          <label class="form-group">
            <span>{{ t('wishlistUrl.listedValue') }}</span>
            <input v-model="draft.currentValue" class="form-input" type="number" min="0" step="0.01" />
          </label>
          <label class="form-group">
            <span>{{ t('wishlistUrl.listingStatus') }}</span>
            <select v-model="draft.listingStatus" class="form-input">
              <option value="">{{ t('common.unknown') }}</option>
              <option value="available">{{ t('wishlistUrl.statuses.available') }}</option>
              <option value="sold">{{ t('wishlistUrl.statuses.sold') }}</option>
              <option value="reserved">{{ t('wishlistUrl.statuses.reserved') }}</option>
              <option value="withdrawn">{{ t('wishlistUrl.statuses.withdrawn') }}</option>
            </select>
          </label>
          <label class="form-group proposal-wide">
            <span>{{ t('wishlistUrl.obverseLegend') }}</span>
            <input v-model.trim="draft.obverseInscription" class="form-input" maxlength="1000" />
          </label>
          <label class="form-group proposal-wide">
            <span>{{ t('wishlistUrl.obverseDescription') }}</span>
            <textarea v-model.trim="draft.obverseDescription" class="form-input" rows="2" maxlength="2000"></textarea>
          </label>
          <label class="form-group proposal-wide">
            <span>{{ t('wishlistUrl.reverseLegend') }}</span>
            <input v-model.trim="draft.reverseInscription" class="form-input" maxlength="1000" />
          </label>
          <label class="form-group proposal-wide">
            <span>{{ t('wishlistUrl.reverseDescription') }}</span>
            <textarea v-model.trim="draft.reverseDescription" class="form-input" rows="2" maxlength="2000"></textarea>
          </label>
          <label class="form-group proposal-wide">
            <span>{{ t('coinFields.notes') }}</span>
            <textarea v-model.trim="draft.notes" class="form-input" rows="3" maxlength="10000"></textarea>
          </label>
        </div>

        <div v-if="previewURL" class="image-review">
          <img :src="previewURL" :alt="t('wishlistUrl.imagePreviewAlt')" />
          <label class="image-choice">
            <input v-model="attachImage" type="checkbox" />
            {{ t('wishlistUrl.attachImage') }}
          </label>
        </div>

        <ul v-if="analysis.warnings?.length" class="warnings">
          <li v-for="warning in analysis.warnings" :key="warning">{{ warning }}</li>
        </ul>

        <div class="review-actions">
          <button type="submit" class="btn btn-primary btn-sm" :disabled="status === 'creating' || !draft.name">
            {{ status === 'creating' ? t('common.adding') : t('wishlistUrl.confirmAdd') }}
          </button>
          <button type="button" class="btn btn-ghost btn-sm" :disabled="status === 'creating'" @click="resetProposal">
            {{ t('common.cancel') }}
          </button>
        </div>
      </form>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { Link } from 'lucide-vue-next'
import { analyzeWishlistURL, createCoin, getApiErrorMessage, proxyImage, uploadImage } from '@/api/client'
import type { Category, CoinMutationPayload, Material, WishlistURLAnalysis, WishlistURLHypothesis } from '@/types'
import { useCoinOptions } from '@/composables/useCoinOptions'
import { materialLabel } from '@/utils/format'
import { t } from '@/i18n'
import { categoryLabel, eraLabel } from '@/i18n/coinVocabulary'

const props = withDefaults(defineProps<{
  embedded?: boolean
}>(), {
  embedded: false,
})

const emit = defineEmits<{ created: [] }>()

const textFields = [
  { key: 'ruler', labelKey: 'wishlistUrl.fields.ruler', maxlength: 200 },
  { key: 'denomination', labelKey: 'coinFields.denomination', maxlength: 200 },
  { key: 'mint', labelKey: 'coinFields.mint', maxlength: 200 },
  { key: 'dateRange', labelKey: 'wishlistUrl.fields.dateRange', maxlength: 100 },
  { key: 'grade', labelKey: 'coinFields.grade', maxlength: 100 },
  { key: 'rarityRating', labelKey: 'wishlistUrl.fields.rarity', maxlength: 100 },
  { key: 'references', labelKey: 'wishlistUrl.fields.references', maxlength: 1000 },
] as const

type DraftKey = typeof textFields[number]['key']
type Draft = Record<DraftKey, string> & {
  name: string
  category: Category
  era: string
  material: string
  weightGrams: string
  diameterMm: string
  currentValue: string
  listingStatus: string
  obverseInscription: string
  obverseDescription: string
  reverseInscription: string
  reverseDescription: string
  notes: string
}

const expanded = ref(props.embedded)
const { categoryOptions, eraOptions, materialOptions, loadOptions } = useCoinOptions()
const sourceURL = ref('')
const status = ref<'idle' | 'pending' | 'cancelled' | 'creating' | 'created' | 'image-warning' | 'failed'>('idle')
const message = ref('')
const analysis = ref<WishlistURLAnalysis | null>(null)
const attachImage = ref(true)
const previewURL = ref('')
const previewFile = ref<File | null>(null)
let controller: AbortController | null = null

const draft = reactive<Draft>(emptyDraft())

function emptyDraft(): Draft {
  return {
    name: '', category: 'Other', era: '', ruler: '', denomination: '', material: 'Other',
    mint: '', dateRange: '', grade: '', rarityRating: '', weightGrams: '', diameterMm: '',
    currentValue: '', listingStatus: '', references: '',
    obverseInscription: '', obverseDescription: '', reverseInscription: '',
    reverseDescription: '', notes: '',
  }
}

function value(hypothesis: WishlistURLHypothesis, key: keyof WishlistURLHypothesis): string {
  const field = hypothesis[key]
  return typeof field === 'object' && field !== null && 'value' in field ? field.value : ''
}

function fillDraft(hypothesis: WishlistURLHypothesis) {
  Object.assign(draft, {
    name: value(hypothesis, 'name'),
    category: categoryOptions.value.includes(value(hypothesis, 'category'))
      ? value(hypothesis, 'category') as Category
      : 'Other',
    era: value(hypothesis, 'era'),
    ruler: value(hypothesis, 'ruler'),
    denomination: value(hypothesis, 'denomination'),
    material: value(hypothesis, 'material') || 'Other',
    mint: value(hypothesis, 'mint'),
    dateRange: value(hypothesis, 'dateRange'),
    grade: value(hypothesis, 'grade'),
    rarityRating: value(hypothesis, 'rarityRating'),
    references: value(hypothesis, 'references') || value(hypothesis, 'coin_type'),
    weightGrams: value(hypothesis, 'weightGrams'),
    diameterMm: value(hypothesis, 'diameterMm'),
    currentValue: value(hypothesis, 'listedPrice'),
    listingStatus: value(hypothesis, 'listingStatus'),
    obverseInscription: value(hypothesis, 'obverseInscription'),
    obverseDescription: value(hypothesis, 'obverseDescription'),
    reverseInscription: value(hypothesis, 'reverseInscription'),
    reverseDescription: value(hypothesis, 'reverseDescription'),
    notes: value(hypothesis, 'notes'),
  })
}

function clearPreview() {
  if (previewURL.value) URL.revokeObjectURL(previewURL.value)
  previewURL.value = ''
  previewFile.value = null
}

async function loadPreview(imageURL: string) {
  clearPreview()
  if (!imageURL) return
  try {
    const response = await proxyImage(imageURL)
    const blob = response.data as Blob
    if (blob.size === 0) return
    const extension = blob.type.includes('png') ? '.png' : '.jpg'
    previewFile.value = new File([blob], `listing${extension}`, { type: blob.type || 'image/jpeg' })
    previewURL.value = URL.createObjectURL(blob)
  } catch {
    message.value = t('wishlistUrl.messages.previewFailed')
  }
}

async function analyze() {
  controller?.abort()
  controller = new AbortController()
  status.value = 'pending'
  message.value = ''
  analysis.value = null
  clearPreview()
  try {
    const response = await analyzeWishlistURL(sourceURL.value, controller.signal)
    analysis.value = response.data
    if (response.data.outcome === 'duplicate') {
      message.value = t('wishlistUrl.messages.duplicate')
      status.value = 'idle'
      return
    }
    if (!response.data.hypothesis) throw new Error(t('wishlistUrl.messages.noProposal'))
    fillDraft(response.data.hypothesis)
    status.value = 'idle'
    message.value = response.data.outcome === 'ready'
      ? t('wishlistUrl.messages.ready')
      : t('wishlistUrl.messages.needsReview')
    await loadPreview(response.data.imageUrl ?? '')
  } catch (error) {
    if (controller?.signal.aborted) {
      status.value = 'cancelled'
      message.value = t('wishlistUrl.messages.cancelled')
    } else {
      status.value = 'failed'
      message.value = getApiErrorMessage(error) || t('wishlistUrl.messages.analyzeFailed')
    }
  } finally {
    controller = null
  }
}

function cancel() {
  controller?.abort()
}

function resetProposal() {
  analysis.value = null
  Object.assign(draft, emptyDraft())
  clearPreview()
  status.value = 'idle'
  message.value = ''
}

async function confirmCreate() {
  if (!analysis.value?.hypothesis || status.value === 'creating') return
  status.value = 'creating'
  message.value = ''
  const hypothesis = analysis.value.hypothesis
  const currentValue = Number(draft.currentValue)
  const payload: CoinMutationPayload = {
    name: draft.name,
    category: draft.category,
    era: draft.era,
    ruler: draft.ruler,
    denomination: draft.denomination,
    material: draft.material as Material,
    mint: draft.mint,
    dateRange: draft.dateRange,
    grade: draft.grade,
    rarityRating: draft.rarityRating,
    weightGrams: draft.weightGrams ? Number(draft.weightGrams) : null,
    diameterMm: draft.diameterMm ? Number(draft.diameterMm) : null,
    obverseInscription: draft.obverseInscription,
    obverseDescription: draft.obverseDescription,
    reverseInscription: draft.reverseInscription,
    reverseDescription: draft.reverseDescription,
    notes: draft.notes,
    referenceUrl: analysis.value.sourceUrl,
    referenceText: draft.references || value(hypothesis, 'dealerName'),
    listingStatus: draft.listingStatus,
    currentValue: draft.currentValue && Number.isFinite(currentValue) ? currentValue : null,
    isWishlist: true,
  }
  try {
    const created = await createCoin(payload)
    let imageWarning = false
    if (attachImage.value && previewFile.value) {
      try {
        await uploadImage(created.data.id, previewFile.value, 'obverse', true)
      } catch {
        imageWarning = true
      }
    }
    status.value = imageWarning ? 'image-warning' : 'created'
    message.value = imageWarning
      ? t('wishlistUrl.messages.addedNoImage')
      : t('wishlistUrl.messages.added')
    analysis.value = null
    clearPreview()
    emit('created')
  } catch (error) {
    status.value = 'failed'
    message.value = getApiErrorMessage(error) || t('wishlistUrl.messages.addFailed')
  }
}

onBeforeUnmount(() => {
  controller?.abort()
  clearPreview()
})

onMounted(loadOptions)
</script>

<style scoped>
.url-intake {
  margin-bottom: 1.5rem;
}

.url-intake-panel {
  margin-top: 0.75rem;
  padding: 1rem;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
}

.url-intake-panel.embedded {
  margin-top: 0;
  padding: 0;
  border: 0;
  background: transparent;
}

.url-row,
.review-actions,
.proposal-heading,
.image-choice {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-message,
.proposal-heading p,
.warnings {
  margin-top: 0.75rem;
  color: var(--text-secondary);
  font-size: 0.85rem;
}

.proposal {
  margin-top: 1.5rem;
}

.proposal-heading {
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.proposal-heading p {
  margin-bottom: 0;
}

.proposal-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.proposal-wide {
  grid-column: 1 / -1;
}

.image-review {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-top: 1rem;
}

.image-review img {
  width: 96px;
  height: 96px;
  object-fit: contain;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
}

.warnings {
  padding-left: 1.25rem;
}

.review-actions {
  margin-top: 1rem;
}

.url-intake-panel .btn,
.url-intake-panel .form-input {
  min-height: 44px;
}

@media (max-width: 640px) {
  .url-row,
  .image-review {
    align-items: stretch;
    flex-direction: column;
  }

  .proposal-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .proposal-wide {
    grid-column: auto;
  }
}
</style>
