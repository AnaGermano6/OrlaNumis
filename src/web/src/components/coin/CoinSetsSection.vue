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

    <!-- Sets this coin could start; each fills by itself. -->
    <div v-if="proposals.length" class="mt-4 border-t border-border-subtle pt-4" data-testid="set-proposals">
      <h4 class="mb-2 text-sm font-medium text-text-secondary">{{ isBanknote ? t('coinDetail.sets.proposals.titleBanknote') : t('coinDetail.sets.proposals.title') }}</h4>
      <ul class="m-0 flex list-none flex-col gap-2 p-0">
        <li
          v-for="proposal in proposals"
          :key="proposal.key"
          class="flex items-center gap-3 rounded-sm border border-border-subtle bg-card px-3 py-2"
          :data-testid="`set-proposal-${proposal.key}`"
        >
          <span class="min-w-0 flex-1">
            <span class="block truncate text-body font-medium text-text-primary">{{ proposal.name }}</span>
            <span class="block text-chip text-text-muted">{{ proposal.kind === 'album' ? t('coinDetail.sets.proposals.album') : isBanknote ? t('coinDetail.sets.proposals.smartBanknote') : t('coinDetail.sets.proposals.smart') }}</span>
          </span>
          <RouterLink
            v-if="existing(proposal)"
            :to="{ name: 'set-detail', params: { id: existing(proposal)!.id } }"
            class="btn btn-ghost btn-xs shrink-0"
            data-testid="set-proposal-open"
          >
            {{ t('coinDetail.sets.proposals.open') }}
          </RouterLink>
          <button
            v-else
            type="button"
            class="btn btn-secondary btn-xs shrink-0"
            :disabled="creating !== null"
            data-testid="set-proposal-create"
            @click="createProposal(proposal)"
          >
            {{ t('coinDetail.sets.proposals.create') }}
          </button>
        </li>
      </ul>
      <p v-if="createFailed" class="mb-0 mt-2 text-sm text-loss" role="alert">{{ t('coinDetail.sets.proposals.failed') }}</p>
    </div>
    <!-- The user's own sets this coin would fit, only when there are some. -->
    <div v-if="recommendations.length" class="mt-4 border-t border-border-subtle pt-4" data-testid="set-suggestions">
      <h4 class="mb-2 text-sm font-medium text-text-secondary">{{ t('coinDetail.sets.suggested') }}</h4>
      <div class="flex flex-col gap-2">
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
// The coin's sets: added to one, taken out of one, the sets it would fit,
// and the sets it could start (an album of the catalogue, sets that fill
// themselves), each made at once or opened when it is already there.
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { getSets, addCoinToSet, removeCoinFromSet, getCoinRecommendations, acceptCoinRecommendation, rejectCoinRecommendation, createSet, type CoinStatus } from '@/api/client'
import { normalizeCoinSetType } from '@/types'
import type { Coin, CoinRecommendation, CoinSetSummary } from '@/types'
import { t } from '@/i18n'
import { coinSetProposals, type CoinSetProposal } from '@/utils/coinSetProposals'
import { randomSetColor } from '@/utils/setColors'

type SetChip = Pick<CoinSetSummary, 'id' | 'name' | 'color' | 'setType'>

const props = defineProps<{
  sets: SetChip[]
  coinId: number
  /** The coin, and what is known of it, for the sets it could start. */
  coin?: Coin | null
  status?: CoinStatus | null
}>()

const emit = defineEmits<{
  changed: []
}>()

const userSets = ref<CoinSetSummary[]>([])
const recommendations = ref<CoinRecommendation[]>([])
const showSetPicker = ref(false)
const setToAdd = ref('')

const availableSets = computed(() => {
  const coinSetIds = new Set(props.sets.map(s => s.id))
  return userSets.value.filter(s => canManageSetMembership(s.setType) && !coinSetIds.has(s.id))
})

const proposals = computed(() => (props.coin ? coinSetProposals(props.coin, props.status) : []))
const isBanknote = computed(() => props.coin?.kind === 'banknote')
const creating = ref<string | null>(null)
const createFailed = ref(false)

// A set's name is the user's own: one of that name is the proposal, made before.
function existing(proposal: CoinSetProposal): CoinSetSummary | undefined {
  return userSets.value.find((set) => set.name === proposal.name)
}

async function createProposal(proposal: CoinSetProposal) {
  if (creating.value) return
  creating.value = proposal.key
  createFailed.value = false
  try {
    await createSet({ ...proposal.request, color: randomSetColor() })
    await loadSets()
  } catch {
    createFailed.value = true
  } finally {
    creating.value = null
  }
}

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

// Without them, nothing is suggested: the section is left out.
async function loadRecommendations() {
  try {
    const response = await getCoinRecommendations(props.coinId)
    recommendations.value = response.data?.recommendations ?? []
  } catch {
    recommendations.value = []
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
