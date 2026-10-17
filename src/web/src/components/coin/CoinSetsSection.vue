<template>
  <section class="mb-6">
    <h3 class="mb-3 text-base font-medium text-text-primary">{{ t('coinDetail.sets.title') }}</h3>
    <div class="flex flex-wrap items-center gap-2">
      <span
        v-for="set in sets"
        :key="set.id"
        class="chip inline-flex items-center gap-[0.35rem]"
        :style="{ backgroundColor: set.color + '22', color: set.color, borderColor: set.color + '44', borderStyle: 'dashed' }"
      >
        {{ set.name }}
        <button
          v-if="canManageSetMembership(set.setType)"
          class="bg-transparent p-0 text-sm leading-none opacity-60 transition-opacity hover:opacity-100"
          type="button"
          :aria-label="t('coinDetail.sets.remove', { name: set.name })"
          @click="handleRemoveSet(set.id)"
        >
          x
        </button>
      </span>
      <button
        v-if="!showSetPicker"
        class="rounded-full border border-dashed border-border-subtle px-[0.85rem] py-[0.35rem] text-chip text-text-secondary transition-colors hover:border-text-primary hover:text-text-primary"
        type="button"
        @click="showSetPicker = true"
      >
        {{ t('coinDetail.sets.add') }}
      </button>
    </div>
    <div v-if="showSetPicker" class="mt-3 flex items-center gap-2 max-sm:flex-col max-sm:items-stretch">
      <select v-model="setToAdd" class="form-select min-w-0 flex-1" @change="handleAddSet">
        <option value="" disabled>{{ t('coinDetail.sets.selectPlaceholder') }}</option>
        <option v-for="set in availableSets" :key="set.id" :value="String(set.id)">{{ set.name }}</option>
        <option v-if="!availableSets.length" value="" disabled>{{ t('coinDetail.sets.noneAvailable') }}</option>
      </select>
      <button class="btn btn-ghost btn-xs" type="button" @click="showSetPicker = false">{{ t('common.cancel') }}</button>
    </div>

    <div class="mt-4 border-t border-border-subtle pt-4">
      <h4 class="mb-2 text-sm font-medium text-text-secondary">{{ t('coinDetail.sets.suggested') }}</h4>
      <div v-if="recommendationsLoading" class="text-sm text-text-muted">{{ t('coinDetail.sets.loadingSuggestions') }}</div>
      <div v-else-if="recommendationsError" class="flex flex-col items-start gap-2">
        <p class="text-sm text-loss">{{ recommendationsError }}</p>
        <button class="btn btn-ghost btn-xs" type="button" @click="loadRecommendations">{{ t('common.retry') }}</button>
      </div>
      <div v-else-if="!recommendations.length" class="text-sm text-text-muted">{{ t('coinDetail.sets.noSuggestions') }}</div>
      <div v-else class="flex flex-col gap-2">
        <div
          v-for="recommendation in recommendations"
          :key="recommendation.id"
          class="rounded-sm border border-border-subtle bg-card p-2"
        >
          <div class="mb-1 flex items-center gap-2">
            <span class="text-sm font-medium text-text-primary">{{ recommendation.targetName }}</span>
            <span class="chip-sm ml-auto capitalize">{{ recommendation.confidence }}</span>
          </div>
          <p class="m-0 text-xs text-text-secondary">{{ recommendation.reasons.join(' • ') }}</p>
          <div class="mt-2 flex gap-2">
            <button class="btn btn-secondary btn-xs" type="button" @click="handleAcceptRecommendation(recommendation.id)">{{ t('common.accept') }}</button>
            <button class="btn btn-ghost btn-xs" type="button" @click="handleRejectRecommendation(recommendation.id)">{{ t('common.reject') }}</button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
// The coin's sets: added to one, taken out of one, and the sets it would fit.
import { ref, computed, onMounted } from 'vue'
import { getSets, addCoinToSet, removeCoinFromSet, getCoinRecommendations, acceptCoinRecommendation, rejectCoinRecommendation } from '@/api/client'
import { normalizeCoinSetType } from '@/types'
import type { CoinRecommendation, CoinSetSummary } from '@/types'
import { t } from '@/i18n'

type SetChip = Pick<CoinSetSummary, 'id' | 'name' | 'color' | 'setType'>

const props = defineProps<{
  sets: SetChip[]
  coinId: number
}>()

const emit = defineEmits<{
  changed: []
}>()

const userSets = ref<CoinSetSummary[]>([])
const recommendations = ref<CoinRecommendation[]>([])
const recommendationsLoading = ref(false)
const recommendationsError = ref('')
const showSetPicker = ref(false)
const setToAdd = ref('')

const availableSets = computed(() => {
  const coinSetIds = new Set(props.sets.map(s => s.id))
  return userSets.value.filter(s => canManageSetMembership(s.setType) && !coinSetIds.has(s.id))
})

function canManageSetMembership(setType: CoinSetSummary['setType']): boolean {
  const normalizedType = normalizeCoinSetType(setType)
  return normalizedType !== 'smart' && normalizedType !== 'agentic'
}

onMounted(async () => {
  await Promise.all([loadSets(), loadRecommendations()])
})

async function loadSets() {
  try {
    userSets.value = (await getSets()).data?.sets ?? []
  } catch { /* ignore */ }
}

async function loadRecommendations() {
  recommendationsLoading.value = true
  recommendationsError.value = ''
  try {
    const response = await getCoinRecommendations(props.coinId)
    recommendations.value = response.data?.recommendations ?? []
  } catch {
    recommendationsError.value = t('coinDetail.sets.loadFailed')
    recommendations.value = []
  } finally {
    recommendationsLoading.value = false
  }
}

async function handleAddSet() {
  if (!setToAdd.value) return
  try {
    await addCoinToSet(Number(setToAdd.value), { coinId: props.coinId })
    emit('changed')
  } catch { /* ignore */ }
  setToAdd.value = ''
  showSetPicker.value = false
  await Promise.all([loadSets(), loadRecommendations()])
}

async function handleRemoveSet(setId: number) {
  try {
    await removeCoinFromSet(setId, props.coinId)
    emit('changed')
  } catch { /* ignore */ }
  await Promise.all([loadSets(), loadRecommendations()])
}

async function handleAcceptRecommendation(recommendationId: number) {
  try {
    await acceptCoinRecommendation(props.coinId, recommendationId)
    emit('changed')
  } catch { /* ignore */ }
  await Promise.all([loadSets(), loadRecommendations()])
}

async function handleRejectRecommendation(recommendationId: number) {
  try {
    await rejectCoinRecommendation(props.coinId, recommendationId)
  } catch { /* ignore */ }
  await loadRecommendations()
}
</script>
