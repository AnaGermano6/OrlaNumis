import { ref } from 'vue'
import {
  getAppSettings, getAppSettingDefaults, updateAppSettings, getOllamaStatus,
  getAnthropicModels, getCoinSearchPrompt, getCoinShowsPrompt, getValuationPrompt,
  testAnthropicConnection, testSearXNGConnection,
} from '@/api/client'
import type { AnthropicModel } from '@/api/client'
import type { AppSettings } from '@/types'
import { CATEGORIES, COIN_ERAS } from '@/types'
import { formatOptionList } from '@/utils/options'
import { t } from '@/i18n'
import { loadCoinVocabulary } from '@/i18n/coinVocabulary'

// Settings of retired features (auction tracking, shipments) that an older
// database may still hold: dropped so the page does not save them back.
const retiredSettingKeys = [
  'PriceAlertCheckEnabled', 'PriceAlertCheckStartTime', 'PriceAlertCheckInterval',
  'AuctionAlertsCheckEnabled', 'AuctionAlertsCheckStartTime', 'AuctionAlertsCheckInterval',
  'AuctionEndingCheckEnabled', 'AuctionEndingCheckStartTime', 'AuctionEndingCheckInterval',
  'ShipmentSyncEnabled', 'ShipmentSyncStartTime', 'ShipmentSyncInterval', 'ShipmentSyncBatchSize',
  'USPSAPIBaseURL', 'USPSAPIKey', 'USPSAPIKeyHeader',
  'UPSAPIBaseURL', 'UPSTokenURL', 'UPSClientID', 'UPSClientSecret', 'UPSScope',
  'FedExAPIBaseURL', 'FedExTokenURL', 'FedExClientID', 'FedExClientSecret', 'FedExScope',
] as const

export function useAdminConfig() {
  const settings = ref<AppSettings>({
    AIProvider: '',
    OllamaURL: 'http://localhost:11434',
    OllamaModel: 'llava',
    ObversePrompt: '',
    ReversePrompt: '',
    TextExtractionPrompt: '',
    OllamaTimeout: '300',
    SearXNGURL: '',
    LogLevel: 'info',
    NumistaSearchTTLHours: '24',
    NumistaDetailTTLHours: '168',
    NumistaEnrichmentLimit: '5',
    NumistaSearchResultLimit: '20',
    NumistaSearchTimeoutSeconds: '4',
    NumistaDetailTimeoutSeconds: '3',
    CoinCopilotEnabled: 'false',
    CoinCopilotAttributionEnabled: 'false',
    CoinCopilotWorkerCount: '1',
    CoinCopilotMaxActivePerUser: '1',
    CoinCopilotQueueDepth: '16',
    CoinCopilotMaxReasoningIterations: '8',
    CoinCopilotMaxToolCalls: '12',
    CoinCopilotHardTimeoutSeconds: '120',
    CoinCopilotMaxPersistedToolResultBytes: '32768',
    CoinCopilotEventRetentionHours: '168',
    CoinCopilotCheckpointRetentionDays: '30',
    CoinCopilotResumeWindowHours: '168',
    PublicAppURL: '',
    WishlistSearchAlertsCheckEnabled: 'false',
    WishlistSearchAlertsCheckStartTime: '03:00',
    CoinCategories: '',
    CoinEras: '',
  })
  const settingDefaults = ref<AppSettings>({
    AIProvider: '',
    OllamaURL: '',
    OllamaModel: '',
    ObversePrompt: '',
    ReversePrompt: '',
    TextExtractionPrompt: '',
    OllamaTimeout: '',
    SearXNGURL: '',
    LogLevel: '',
    NumistaSearchTTLHours: '24',
    NumistaDetailTTLHours: '168',
    NumistaEnrichmentLimit: '5',
    NumistaSearchResultLimit: '20',
    NumistaSearchTimeoutSeconds: '4',
    NumistaDetailTimeoutSeconds: '3',
    CoinCopilotEnabled: 'false',
    CoinCopilotAttributionEnabled: 'false',
    CoinCopilotWorkerCount: '1',
    CoinCopilotMaxActivePerUser: '1',
    CoinCopilotQueueDepth: '16',
    CoinCopilotMaxReasoningIterations: '8',
    CoinCopilotMaxToolCalls: '12',
    CoinCopilotHardTimeoutSeconds: '120',
    CoinCopilotMaxPersistedToolResultBytes: '32768',
    CoinCopilotEventRetentionHours: '168',
    CoinCopilotCheckpointRetentionDays: '30',
    CoinCopilotResumeWindowHours: '168',
    PublicAppURL: '',
    WishlistSearchAlertsCheckEnabled: 'false',
    WishlistSearchAlertsCheckStartTime: '03:00',
    CoinCategories: '',
    CoinEras: '',
  })
  const settingsMsg = ref('')
  const settingsError = ref(false)
  const settingsSaving = ref(false)

  let saveTimerId: ReturnType<typeof setTimeout> | null = null

  // Ollama
  const ollamaTesting = ref(false)
  const ollamaTestResult = ref('')
  const ollamaTestOk = ref(false)

  // Anthropic
  const anthropicTesting = ref(false)
  const anthropicTestResult = ref('')
  const anthropicTestOk = ref(false)
  let savedAnthropicAPIKey = ''
  const anthropicModels = ref<AnthropicModel[]>([
    { id: 'claude-sonnet-5', name: 'Claude Sonnet 5' },
    { id: 'claude-haiku-4-5-20251001', name: 'Claude Haiku 4.5' },
    { id: 'claude-opus-5', name: 'Claude Opus 5' },
  ])

  // SearXNG
  const searxngTesting = ref(false)
  const searxngTestResult = ref('')
  const searxngTestOk = ref(false)

  // Prompts
  const coinSearchPromptDefault = ref('')
  const coinShowsPromptDefault = ref('')
  const valuationPromptDefault = ref('')

  // Schedule-tab save messages (cleared alongside main settingsMsg)
  const availSettingsMsg = ref('')
  const availSettingsError = ref(false)
  const valSettingsMsg = ref('')
  const valSettingsError = ref(false)

  async function loadSettings() {
    try {
      const [settingsRes, defaultsRes] = await Promise.all([
        getAppSettings(),
        getAppSettingDefaults(),
      ])
      settingDefaults.value = { ...settingDefaults.value, ...defaultsRes.data }
      settings.value = { ...settings.value, ...settingsRes.data }
      savedAnthropicAPIKey = settings.value.AnthropicAPIKey ?? ''
      retiredSettingKeys.forEach(key => {
        delete settings.value[key]
      })

      // Apply defaults for coin property options if not set
      if (!settings.value.CoinCategories) {
        settings.value.CoinCategories = formatOptionList(CATEGORIES)
      }
      if (!settings.value.CoinEras) {
        settings.value.CoinEras = formatOptionList(COIN_ERAS)
      }

      // Apply defaults for schedule settings if not set
      if (!settings.value.WishlistSearchAlertsCheckEnabled) {
        settings.value.WishlistSearchAlertsCheckEnabled = 'false'
      }
      if (!settings.value.WishlistSearchAlertsCheckStartTime) {
        settings.value.WishlistSearchAlertsCheckStartTime = '03:00'
      }

      // Apply defaults for collection health snapshot settings if not set

      const [modelsRes, coinSearchRes, coinShowsRes, valPromptRes] = await Promise.all([
        getAnthropicModels().catch(() => null),
        getCoinSearchPrompt().catch(() => null),
        getCoinShowsPrompt().catch(() => null),
        getValuationPrompt().catch(() => null),
      ])

      if (modelsRes?.data?.length) {
        anthropicModels.value = modelsRes.data
      }

      if (coinSearchRes?.data) {
        coinSearchPromptDefault.value = coinSearchRes.data.default
        if (!settings.value.CoinSearchPrompt) {
          settings.value.CoinSearchPrompt = coinSearchRes.data.prompt
        }
      }

      if (coinShowsRes?.data) {
        coinShowsPromptDefault.value = coinShowsRes.data.default
        if (!settings.value.CoinShowsPrompt) {
          settings.value.CoinShowsPrompt = coinShowsRes.data.prompt
        }
      }

      if (valPromptRes?.data) {
        valuationPromptDefault.value = valPromptRes.data.default
        if (!settings.value.ValuationPrompt) {
          settings.value.ValuationPrompt = valPromptRes.data.prompt
        }
      }
    } catch { /* use defaults */ }
  }

  async function saveSettings() {
    settingsSaving.value = true
    settingsMsg.value = ''
    settingsError.value = false
    availSettingsMsg.value = ''
    availSettingsError.value = false
    valSettingsMsg.value = ''
    valSettingsError.value = false
    try {
      // The collection currency changes only with its stored values, from
      // its own section (POST /admin/collection-currency).
      const entries = Object.entries(settings.value)
        .filter(([key]) => key !== 'CollectionCurrency')
        .map(([key, value]) => ({ key, value: String(value) }))
      await updateAppSettings(entries)
      savedAnthropicAPIKey = settings.value.AnthropicAPIKey ?? ''
      // Pick up renamed or newly translated categories and eras right away.
      void loadCoinVocabulary()
      settingsMsg.value = t('admin.config.saved')
      availSettingsMsg.value = t('admin.config.saved')
      valSettingsMsg.value = t('admin.config.saved')
      if (saveTimerId) clearTimeout(saveTimerId)
      saveTimerId = setTimeout(() => { availSettingsMsg.value = ''; valSettingsMsg.value = '' }, 3000)
    } catch {
      settingsMsg.value = t('admin.config.saveFailed')
      settingsError.value = true
      availSettingsMsg.value = t('admin.config.saveFailed')
      availSettingsError.value = true
      valSettingsMsg.value = t('admin.config.saveFailed')
      valSettingsError.value = true
    } finally {
      settingsSaving.value = false
    }
  }

  async function testOllamaConnection() {
    ollamaTesting.value = true
    ollamaTestResult.value = ''
    try {
      const res = await getOllamaStatus()
      ollamaTestOk.value = res.data.available
      ollamaTestResult.value = res.data.message
    } catch {
      ollamaTestOk.value = false
      ollamaTestResult.value = t('admin.config.ollamaFailed')
    } finally {
      ollamaTesting.value = false
    }
  }

  async function testAnthropicConn() {
    if ((settings.value.AnthropicAPIKey ?? '') !== savedAnthropicAPIKey) {
      anthropicTestOk.value = false
      anthropicTestResult.value = t('admin.config.saveBeforeTest')
      return
    }

    anthropicTesting.value = true
    anthropicTestResult.value = ''
    try {
      const res = await testAnthropicConnection()
      anthropicTestOk.value = res.data.available
      anthropicTestResult.value = res.data.message
    } catch {
      anthropicTestOk.value = false
      anthropicTestResult.value = t('admin.config.anthropicFailed')
    } finally {
      anthropicTesting.value = false
    }
  }

  async function testSearxngConn() {
    searxngTesting.value = true
    searxngTestResult.value = ''
    try {
      const res = await testSearXNGConnection()
      searxngTestOk.value = res.data.available
      searxngTestResult.value = res.data.message
    } catch {
      searxngTestOk.value = false
      searxngTestResult.value = t('admin.config.searxngFailed')
    } finally {
      searxngTesting.value = false
    }
  }

  return {
    settings,
    settingDefaults,
    settingsMsg,
    settingsError,
    settingsSaving,
    // Ollama
    ollamaTesting,
    ollamaTestResult,
    ollamaTestOk,
    // Anthropic
    anthropicTesting,
    anthropicTestResult,
    anthropicTestOk,
    anthropicModels,
    // SearXNG
    searxngTesting,
    searxngTestResult,
    searxngTestOk,
    // Prompts
    coinSearchPromptDefault,
    coinShowsPromptDefault,
    valuationPromptDefault,
    // Schedule messages
    availSettingsMsg,
    availSettingsError,
    valSettingsMsg,
    valSettingsError,
    // Functions
    loadSettings,
    saveSettings,
    testOllamaConnection,
    testAnthropicConn,
    testSearxngConn,
    cleanup() {
      if (saveTimerId) clearTimeout(saveTimerId)
    },
  }
}
