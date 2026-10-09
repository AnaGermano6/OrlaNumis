import { ref } from 'vue'
import { CATEGORIES, MATERIALS, COIN_ERAS } from '@/types'
import { coinVocabulary, loadCoinVocabulary } from '@/i18n/coinVocabulary'

// Option lists hold the stored values; render option text with
// categoryLabel()/eraLabel() so the choices read in the current locale.
export function useCoinOptions() {
  const settingsLoaded = ref(false)
  const categoryOptions = ref<string[]>([...CATEGORIES])
  const eraOptions = ref<string[]>([...COIN_ERAS])
  const materialOptions = ref<string[]>([...MATERIALS])

  async function loadOptions() {
    // Never rejects: on error the vocabulary keeps the defaults.
    await loadCoinVocabulary()
    categoryOptions.value = [...coinVocabulary.categories]
    eraOptions.value = [...coinVocabulary.eras]
    materialOptions.value = [...MATERIALS]
    settingsLoaded.value = true
  }

  return {
    settingsLoaded,
    categoryOptions,
    eraOptions,
    materialOptions,
    loadOptions,
  }
}
