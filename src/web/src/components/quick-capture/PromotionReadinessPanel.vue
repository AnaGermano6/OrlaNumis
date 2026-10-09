<template>
  <section class="card mt-6 min-w-0 overflow-hidden [overflow-wrap:anywhere]">
    <div class="mb-4 flex items-start justify-between gap-4">
      <div>
        <span class="section-label">{{ t('quickCapture.promote.eyebrow') }}</span>
        <h2 class="mt-1">{{ t('quickCapture.promote.title') }}</h2>
      </div>
    </div>

    <!-- Already promoted -->
    <template v-if="alreadyPromoted">
      <p class="text-base text-text-secondary">{{ t('quickCapture.promote.alreadyPromoted') }}</p>
      <RouterLink class="btn btn-primary" :to="`/coin/${promotedCoinId}`">{{ t('common.viewCoin') }}</RouterLink>
    </template>

    <!-- Success -->
    <template v-else-if="successCoinId">
      <p class="text-base text-text-secondary">{{ t('quickCapture.promote.success') }}</p>
      <RouterLink class="btn btn-primary" :to="`/coin/${successCoinId}`">{{ t('common.viewCoin') }}</RouterLink>
    </template>

    <!-- Promotion form -->
    <template v-else>
      <p class="mb-4 text-body text-text-secondary">{{ t('quickCapture.promote.intro') }}</p>

      <fieldset class="m-0 grid min-w-0 gap-3 border-0 p-0 [grid-template-columns:repeat(auto-fit,minmax(min(220px,100%),1fr))]">
        <legend class="section-label col-span-full mb-1">{{ t('quickCapture.promote.promoteTo') }}</legend>
        <label
          class="flex cursor-pointer items-start gap-3 rounded-sm border border-border-subtle bg-input p-3 text-text-primary transition-[border-color,background,box-shadow] duration-200"
          :class="target === 'collection' ? 'border-gold bg-gold-glow shadow-glow' : ''"
        >
          <input v-model="target" type="radio" value="collection" class="mt-[0.2rem] shrink-0 accent-gold">
          <Coins :size="20" class="shrink-0 text-gold" />
          <span class="grid gap-1">
            <strong class="text-base font-semibold text-text-primary">{{ t('quickCapture.promote.collection') }}</strong>
            <small class="text-sm text-text-secondary">{{ t('quickCapture.promote.collectionHint') }}</small>
          </span>
        </label>
        <label
          class="flex cursor-pointer items-start gap-3 rounded-sm border border-border-subtle bg-input p-3 text-text-primary transition-[border-color,background,box-shadow] duration-200"
          :class="target === 'wishlist' ? 'border-gold bg-gold-glow shadow-glow' : ''"
        >
          <input v-model="target" type="radio" value="wishlist" class="mt-[0.2rem] shrink-0 accent-gold">
          <Bookmark :size="20" class="shrink-0 text-gold" />
          <span class="grid gap-1">
            <strong class="text-base font-semibold text-text-primary">{{ t('quickCapture.promote.wishlist') }}</strong>
            <small class="text-sm text-text-secondary">{{ t('quickCapture.promote.wishlistHint') }}</small>
          </span>
        </label>
        <span v-if="fieldErrors.target" class="col-span-full text-body text-warning">{{ fieldErrors.target }}</span>
      </fieldset>

      <div class="mt-4 grid gap-3 rounded-sm border border-border-subtle bg-input p-3" aria-live="polite">
        <div class="flex items-start gap-3 text-text-secondary" :class="{ 'text-text-primary': hasRequiredName }">
          <CheckCircle v-if="hasRequiredName" :size="18" class="mt-[0.15rem] shrink-0 text-gold" />
          <AlertCircle v-else :size="18" class="mt-[0.15rem] shrink-0 text-warning" />
          <span class="grid gap-1">
            <strong class="text-base font-semibold text-text-primary">{{ hasRequiredName ? t('quickCapture.promote.titleReady') : t('quickCapture.promote.titleRequired') }}</strong>
            <small class="text-sm text-text-secondary">{{ hasRequiredName ? t('quickCapture.promote.titleReadyHint') : t('quickCapture.promote.titleRequiredHint') }}</small>
          </span>
        </div>
        <div class="flex items-start gap-3 text-text-primary">
          <ImageIcon :size="18" class="mt-[0.15rem] shrink-0 text-gold" />
          <span class="grid gap-1">
            <strong class="text-base font-semibold text-text-primary">{{ imageCountLabel }}</strong>
            <small class="text-sm text-text-secondary">{{ t('quickCapture.promote.imagesHint') }}</small>
          </span>
        </div>
        <div class="flex items-start gap-3 text-text-primary">
          <LinkIcon :size="18" class="mt-[0.15rem] shrink-0 text-gold" />
          <span class="grid gap-1">
            <strong class="text-base font-semibold text-text-primary">{{ referenceReadinessLabel }}</strong>
            <small class="text-sm text-text-secondary">{{ referenceReadinessMessage }}</small>
          </span>
        </div>
        <p class="m-0 text-sm text-text-secondary">{{ t('quickCapture.promote.fieldsHint') }}</p>
        <span v-if="fieldErrors.name" class="text-body text-warning">{{ fieldErrors.name }}</span>
        <div v-if="readinessFieldErrors.length" class="grid text-body text-warning">
          <span v-for="[field, message] in readinessFieldErrors" :key="field">{{ message }}</span>
        </div>
      </div>

      <label class="my-4 flex cursor-pointer items-start gap-[0.6rem] text-base text-text-secondary">
        <input v-model="confirmed" type="checkbox" class="mt-[0.15rem] shrink-0 accent-gold">
        <span>{{ t('quickCapture.promote.confirm', { destination: destinationLabel }) }}</span>
      </label>

      <p v-if="promoteError" class="mb-4 text-base text-warning">{{ promoteError }}</p>

      <div class="flex justify-end">
        <button
          type="button"
          class="btn btn-primary w-full justify-center sm:w-auto"
          :disabled="!confirmed || promoting || !hasRequiredName || selectionChangesPending"
          @click="doPromote"
        >
          {{ promoting ? t('quickCapture.promote.promoting') : t('quickCapture.promote.promoteToDestination', { destination: destinationLabel }) }}
        </button>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getApiErrorMessage, promoteQuickCaptureDraft } from '@/api/client'
import type { QuickCaptureDraft, QuickCapturePromoteOverrides } from '@/types'
import { AlertCircle, Bookmark, CheckCircle, Coins, Image as ImageIcon, Link as LinkIcon } from 'lucide-vue-next'
import { t } from '@/i18n'

const props = withDefaults(
  defineProps<{
    draft: QuickCaptureDraft
    promotionOverrides?: QuickCapturePromoteOverrides
    selectionChangesPending?: boolean
  }>(),
  {
    promotionOverrides: () => ({}),
    selectionChangesPending: false,
  }
)
const emit = defineEmits<{ promoted: [coinId: number] }>()

type PromotionTarget = 'collection' | 'wishlist'

const alreadyPromoted = computed(
  () => props.draft.status === 'promoted' && props.draft.promotedCoinId != null
)
const promotedCoinId = computed(() => props.draft.promotedCoinId)

const target = ref<PromotionTarget>('collection')
const confirmed = ref(false)
const promoting = ref(false)
const promoteError = ref('')
const fieldErrors = ref<Record<string, string>>({})
const successCoinId = ref<number | null>(null)
const destinationLabel = computed(() => target.value === 'wishlist' ? t('quickCapture.promote.wishlist') : t('quickCapture.promote.collection'))
const hasRequiredName = computed(() => (props.promotionOverrides.name ?? props.draft.workingTitle ?? '').trim().length > 0)
const imageCount = computed(() => props.draft.images?.length ?? 0)
const imageCountLabel = computed(() => t('quickCapture.promote.savedImages', { count: imageCount.value }))
const referenceReadinessLabel = computed(() => {
  if (props.selectionChangesPending) return t('quickCapture.promote.saveReferenceFirst')
  return props.draft.selectedNumistaReference
    ? t('quickCapture.promote.numistaSelected', { number: props.draft.selectedNumistaReference.number })
    : t('quickCapture.promote.noReference')
})
const referenceReadinessMessage = computed(() => props.selectionChangesPending
  ? t('quickCapture.promote.pendingHint')
  : t('quickCapture.promote.optionalHint'))
const readinessFieldErrors = computed(() => Object.entries(fieldErrors.value)
  .filter(([field]) => !['target', 'name', 'confirm'].includes(field)))

async function doPromote() {
  promoteError.value = ''
  fieldErrors.value = {}
  if (props.selectionChangesPending) {
    promoteError.value = t('quickCapture.promote.saveNumistaFirst')
    return
  }
  if (!hasRequiredName.value) {
    fieldErrors.value = { name: t('quickCapture.promote.nameRequired') }
    promoteError.value = t('quickCapture.promote.completeFields')
    return
  }
  promoting.value = true
  try {
    const res = await promoteQuickCaptureDraft(props.draft.id, {
      confirm: true,
      target: target.value,
      overrides: props.promotionOverrides,
    })
    if (res.data.alreadyPromoted) {
      // trigger idempotent path in parent
      emit('promoted', res.data.coinId)
    } else {
      successCoinId.value = res.data.coinId
      emit('promoted', res.data.coinId)
    }
  } catch (err: unknown) {
    const errData = (err as { response?: { data?: { error?: string; fields?: Record<string, string> } } })
      ?.response?.data
    if (errData?.fields) {
      fieldErrors.value = errData.fields
      promoteError.value = errData.error ?? t('quickCapture.promote.completeFields')
    } else {
      promoteError.value = getApiErrorMessage(err) || t('quickCapture.promote.failed')
    }
  } finally {
    promoting.value = false
  }
}
</script>
