import { onScopeDispose, ref, watch, type Ref } from 'vue'
import { getCoinsInSet, getSet, getSetChecklist, getSetCompletion } from '@/api/client'
import { normalizeCoinSetType, type ChecklistResponse, type Coin, type CoinSetCompletion, type CoinSetDetail } from '@/types'
import { t } from '@/i18n'

export function useSetDetails(setId: Ref<number>) {
  const set = ref<CoinSetDetail | null>(null)
  const coins = ref<Coin[]>([])
  const completion = ref<CoinSetCompletion | null>(null)
  /** A checklist set's slots. */
  const checklist = ref<ChecklistResponse | null>(null)
  const loading = ref(true)
  const error = ref('')
  let generation = 0
  let request = 0

  function capture() {
    const id = setId.value
    const currentGeneration = generation
    return { id, isCurrent: () => currentGeneration === generation && id === setId.value }
  }

  async function load() {
    const resource = capture()
    const currentRequest = ++request
    const isCurrent = () => resource.isCurrent() && currentRequest === request
    loading.value = true
    error.value = ''
    try {
      const [setResponse, coinsResponse] = await Promise.all([getSet(resource.id), getCoinsInSet(resource.id)])
      if (!isCurrent()) return
      const type = normalizeCoinSetType(setResponse.data.setType)
      const [details, slots] = await Promise.all([
        type === 'goal' || type === 'agentic' ? getSetCompletion(resource.id).then((response) => response.data) : null,
        type === 'checklist' ? getSetChecklist(resource.id).then((response) => response.data) : null,
      ])
      if (!isCurrent()) return
      set.value = setResponse.data
      coins.value = coinsResponse.data.coins
      completion.value = details
      checklist.value = slots
    } catch {
      if (isCurrent()) error.value = t('sets.loadFailed')
    } finally {
      if (isCurrent()) loading.value = false
    }
  }

  watch(setId, () => {
    generation += 1
    set.value = null
    coins.value = []
    completion.value = null
    checklist.value = null
    void load()
  }, { immediate: true, flush: 'sync' })
  onScopeDispose(() => { generation += 1 })

  return { set, coins, completion, checklist, loading, error, load, capture }
}
