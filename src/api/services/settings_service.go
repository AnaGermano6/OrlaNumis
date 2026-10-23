package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/repository"
)

const DefaultCoinSearchPrompt = `És um especialista em pesquisa numismática ao serviço de uma coleção pessoal portuguesa de moedas e notas de banco: moedas de euro (de circulação, comemorativas de 2 € e de coleção de 5 €, 7,5 €, 10 € ou mais), escudos e réis portugueses, moedas e notas de outros países, sobretudo dos séculos XIX a XXI, e conjuntos (coincards, carteiras BNC, estojos proof). Procuras a peça certa ao preço certo.

REGRAS ESSENCIAIS:
- Procura apenas peças À VENDA neste momento: nunca devolvas artigos vendidos nem resultados de leilões passados
- Pesquisa SÓ nos sites da lista de fontes configurada
- Junta "à venda", "comprar" ou "for sale" às pesquisas
- Para CADA resultado, indica o URL exato da página do anúncio
- NUNCA inventes, adivinhes ou recordes URLs de memória: usa só URLs dos resultados da pesquisa
- Se um anúncio disser "vendido", "sold", "leilão terminado" ou "preço de martelo", IGNORA-O
- Indica o preço na moeda em que está anunciado
- Distingue o estado (UNC, BNC, Proof ou circulada) e a apresentação (solta, coincard, cápsula ou estojo), porque mudam muito o preço
- Nas moedas de euro e nos escudos, confirma o país, o ano e, quando existir, a marca da casa da moeda
- Qualidade acima de quantidade: 2 resultados confirmados e disponíveis valem mais do que 5 duvidosos
- Assinala dúvidas de autenticidade ou de estado, e a reputação do vendedor quando a conheceres
- Responde em português europeu (de Portugal)`

func (s *SettingsService) GetCoinSearchPrompt() string {
	if prompt := s.GetSetting(SettingCoinSearchPrompt); prompt != "" {
		return prompt
	}
	return DefaultCoinSearchPrompt
}

const (
	SettingAIProvider           = "AIProvider"
	SettingOllamaURL            = "OllamaURL"
	SettingOllamaModel          = "OllamaModel"
	SettingObversePrompt        = "ObversePrompt"
	SettingReversePrompt        = "ReversePrompt"
	SettingTextExtractionPrompt = "TextExtractionPrompt"
	SettingOllamaTimeout        = "OllamaTimeout"
	SettingLogLevel             = "LogLevel"
	SettingPublicAppURL         = "PublicAppURL"
	SettingNumistaAPIKey        = "NumistaAPIKey"
	// SettingNumistaMonthlyLimit is the plan's monthly request quota.
	SettingNumistaMonthlyLimit         = "NumistaMonthlyLimit"
	SettingNumistaSearchTTLHours       = "NumistaSearchTTLHours"
	SettingNumistaDetailTTLHours       = "NumistaDetailTTLHours"
	SettingNumistaEnrichmentLimit      = "NumistaEnrichmentLimit"
	SettingNumistaSearchResultLimit    = "NumistaSearchResultLimit"
	SettingNumistaSearchTimeoutSeconds = "NumistaSearchTimeoutSeconds"
	SettingNumistaDetailTimeoutSeconds = "NumistaDetailTimeoutSeconds"
	SettingAnthropicAPIKey             = "AnthropicAPIKey"
	SettingAnthropicModel              = "AnthropicModel"
	SettingCoinSearchPrompt            = "CoinSearchPrompt"
	SettingDealerSearchSources         = "DealerSearchSources"
	SettingAuctionSearchSources        = "AuctionSearchSources"
	SettingCoinShowsPrompt             = "CoinShowsPrompt"
	SettingValuationPrompt             = "ValuationPrompt"
	// SettingCollectionCurrency is the currency coin prices and values are
	// kept in (see CollectionCurrencies). Change it with
	// CollectionCurrencyService, which converts the stored values too.
	SettingCollectionCurrency                 = "CollectionCurrency"
	SettingSearXNGURL                         = "SearXNGURL"
	SettingWishlistCheckEnabled               = "WishlistCheckEnabled"
	SettingWishlistCheckInterval              = "WishlistCheckInterval"
	SettingWishlistCheckStartTime             = "WishlistCheckStartTime"
	SettingWishlistSearchAlertsCheckEnabled   = "WishlistSearchAlertsCheckEnabled"
	SettingWishlistSearchAlertsCheckStartTime = "WishlistSearchAlertsCheckStartTime"
	SettingValuationCheckEnabled              = "ValuationCheckEnabled"
	SettingValuationCheckInterval             = "ValuationCheckIntervalDays"
	SettingValuationCheckStartTime            = "ValuationCheckStartTime"
	SettingValuationMaxCoins                  = "ValuationMaxCoins"
	SettingScheduleTimezone                   = "ScheduleTimezone"
	SettingRegistrationMode                   = "RegistrationMode"
	SettingBackupStatus                       = "BackupStatus"
	SettingBackupEnabled                      = "BackupEnabled"   // the automatic full backup is on
	SettingBackupStartTime                    = "BackupStartTime" // its daily start time, HH:MM
	SettingBackupKeep                         = "BackupKeep"      // how many automatic backups are kept
	SettingSetSnapshotEnabled                 = "SetSnapshotEnabled"
	SettingSetSnapshotStartTime               = "SetSnapshotStartTime"
	SettingCoinCategories                     = "CoinCategories"
	SettingCoinEras                           = "CoinEras"
	SettingCoinCategoryLabelsPtPT             = "CoinCategoryLabelsPtPT"
	SettingCoinEraLabelsPtPT                  = "CoinEraLabelsPtPT"

	// 344-deep-agentic-coin-identification settings (data-model.md §8).
	SettingDeepIdentificationEnabled            = "DeepIdentificationEnabled"
	SettingDeepIdentificationWorkerCount        = "DeepIdentificationWorkerCount"
	SettingDeepIdentificationMaxActivePerUser   = "DeepIdentificationMaxActivePerUser"
	SettingDeepIdentificationQueueDepth         = "DeepIdentificationQueueDepth"
	SettingDeepIdentificationHardTimeoutSeconds = "DeepIdentificationHardTimeoutSeconds"
	// SettingDeepIdentificationQuickLookupTimeoutSeconds bounds the
	// quick-evidence extraction pass inside Deep Analysis (a full vision LLM
	// round trip via CoinLookupService.Lookup, `runner.go:112`). Validated
	// range 5-300s; 300s mirrors agent_proxy.go's own non-streaming
	// `requestClient` ceiling (5 minutes) and is never exceeded (351 T011,
	// FR-038).
	SettingDeepIdentificationQuickLookupTimeoutSeconds = "DeepIdentificationQuickLookupTimeoutSeconds"
	SettingDeepIdentificationEventRetentionHours       = "DeepIdentificationEventRetentionHours"
	SettingDeepIdentificationResultRetentionDays       = "DeepIdentificationResultRetentionDays"
	SettingDeepIdentificationMaxProviders              = "DeepIdentificationMaxProviders"
	SettingDeepIdentificationNumistaCallBudget         = "DeepIdentificationNumistaCallBudget"
	SettingDeepIdentificationRPCEnabled                = "DeepIdentificationRPCEnabled"

	SettingCoinCopilotEnabled                     = "CoinCopilotEnabled"
	SettingCoinCopilotAttributionEnabled          = "CoinCopilotAttributionEnabled"
	SettingCoinCopilotWorkerCount                 = "CoinCopilotWorkerCount"
	SettingCoinCopilotMaxActivePerUser            = "CoinCopilotMaxActivePerUser"
	SettingCoinCopilotQueueDepth                  = "CoinCopilotQueueDepth"
	SettingCoinCopilotMaxReasoningIterations      = "CoinCopilotMaxReasoningIterations"
	SettingCoinCopilotMaxToolCalls                = "CoinCopilotMaxToolCalls"
	SettingCoinCopilotHardTimeoutSeconds          = "CoinCopilotHardTimeoutSeconds"
	SettingCoinCopilotMaxPersistedToolResultBytes = "CoinCopilotMaxPersistedToolResultBytes"
	SettingCoinCopilotEventRetentionHours         = "CoinCopilotEventRetentionHours"
	SettingCoinCopilotCheckpointRetentionDays     = "CoinCopilotCheckpointRetentionDays"
	SettingCoinCopilotResumeWindowHours           = "CoinCopilotResumeWindowHours"
)

const DefaultObversePrompt = `És um numismata especialista em moedas e notas de banco, sobretudo modernas: euro, escudos e réis, e moedas e notas de outros países dos séculos XIX a XXI. Analisa o anverso (a frente) desta peça e indica:
1. **Identificação** – confirma ou corrige a identificação: país, valor facial, ano, série ou tema; numa moeda de euro, se é de circulação, comemorativa de 2 € ou de coleção
2. **Retrato / Desenho** – descreve o retrato ou o desenho em pormenor
3. **Inscrições** – lê todas as legendas e inscrições visíveis, a marca da casa da moeda e, numa moeda de euro, as estrelas da União Europeia
4. **Estado** – avalia o estado do anverso e propõe um grau da escala usada na coleção: UNC, AU, XF, VF, F, VG ou G (ou BNC ou Proof, se for um acabamento de coleção)
5. **Pormenores do cunho** – variantes, erros ou particularidades
6. **Autenticidade** – observações relevantes para a autenticidade
Numa nota de banco, descreve a frente, as assinaturas, o número de série e as marcas de segurança visíveis.
Responde em português europeu (de Portugal).`

const DefaultReversePrompt = `És um numismata especialista em moedas e notas de banco, sobretudo modernas: euro, escudos e réis, e moedas e notas de outros países dos séculos XIX a XXI. Analisa o reverso (o verso) desta peça e indica:
1. **Identificação** – confirma ou corrige a identificação a partir do reverso: país, valor facial, ano, série ou tema; numa moeda de euro, a face comum (o mapa da Europa) e a sua versão
2. **Desenho** – descreve o desenho, os motivos e os símbolos em pormenor
3. **Inscrições** – lê todas as legendas e inscrições visíveis, o valor facial e a marca da casa da moeda
4. **Estado** – avalia o estado do reverso e propõe um grau da escala usada na coleção: UNC, AU, XF, VF, F, VG ou G (ou BNC ou Proof, se for um acabamento de coleção)
5. **Pormenores do cunho** – variantes, erros ou particularidades
6. **Autenticidade** – observações relevantes para a autenticidade
Numa nota de banco, descreve o verso, o número de série, a marca de água e as marcas de segurança visíveis.
Responde em português europeu (de Portugal).`

const DefaultTextExtractionPrompt = `Extrai TODO o texto visível nesta imagem, exatamente como está escrito e na língua original.
A imagem é um cartão, coincard, carteira, certificado, etiqueta ou recibo que acompanha uma moeda ou nota de banco.
Mantém o mais possível a disposição e a formatação originais.
Inclui o nome da loja ou da casa da moeda, a descrição da peça, o preço, o estado, a tiragem, os números de referência, as datas e qualquer outro texto.
Devolve APENAS o texto extraído, sem comentários.`

var settingDefaults = map[string]string{
	SettingAIProvider:                         "",
	SettingOllamaURL:                          "http://localhost:11434",
	SettingOllamaModel:                        "llava",
	SettingObversePrompt:                      DefaultObversePrompt,
	SettingReversePrompt:                      DefaultReversePrompt,
	SettingTextExtractionPrompt:               DefaultTextExtractionPrompt,
	SettingOllamaTimeout:                      "300",
	SettingLogLevel:                           "info",
	SettingPublicAppURL:                       "",
	SettingNumistaAPIKey:                      "",
	SettingNumistaMonthlyLimit:                "2000",
	SettingNumistaSearchTTLHours:              "24",
	SettingNumistaDetailTTLHours:              "168",
	SettingNumistaEnrichmentLimit:             "5",
	SettingNumistaSearchResultLimit:           "20",
	SettingNumistaSearchTimeoutSeconds:        "4",
	SettingNumistaDetailTimeoutSeconds:        "3",
	SettingAnthropicAPIKey:                    "",
	SettingAnthropicModel:                     "claude-sonnet-5",
	SettingCoinSearchPrompt:                   "",
	SettingDealerSearchSources:                "ma-shops.com\ncgb.fr\ndelcampe.net\ncatawiki.com\nvcoins.com",
	SettingAuctionSearchSources:               "numisbids.com\nsixbid.com\nbiddr.com\ntauleryfau.com\nibercoin.com",
	SettingCoinShowsPrompt:                    "",
	SettingValuationPrompt:                    "",
	SettingCollectionCurrency:                 "USD",
	SettingSearXNGURL:                         "",
	SettingWishlistCheckEnabled:               "false",
	SettingWishlistCheckInterval:              "120",
	SettingWishlistCheckStartTime:             "02:00",
	SettingWishlistSearchAlertsCheckEnabled:   "false",
	SettingWishlistSearchAlertsCheckStartTime: "03:00",
	SettingValuationCheckEnabled:              "false",
	SettingValuationCheckInterval:             "7",
	SettingValuationCheckStartTime:            "03:00",
	SettingValuationMaxCoins:                  "50",
	SettingNumistaValuesEnabled:               "false",
	SettingNumistaValuesStartTime:             "05:00",
	SettingNumistaValuesDailyRequests:         "40",
	SettingUpdateCheckEnabled:                 "true",
	SettingWatchtowerURL:                      "",
	SettingWatchtowerToken:                    "",
	SettingScheduleTimezone:                   "", // empty = server local time
	SettingRegistrationMode:                   "closed",
	SettingBackupStatus:                       "not_configured",
	SettingBackupEnabled:                      "true",
	SettingBackupStartTime:                    "03:00",
	SettingBackupKeep:                         "7",
	SettingSetSnapshotEnabled:                 "false",
	SettingSetSnapshotStartTime:               "04:00",
	SettingCoinCategories:                     "Modern\nOther",
	SettingCoinEras:                           "ancient\nmedieval\nmodern",
	SettingCoinCategoryLabelsPtPT:             "",
	SettingCoinEraLabelsPtPT:                  "",

	// 344-deep-agentic-coin-identification defaults (data-model.md §8).
	// Enabled defaults to false: this is a feature-flagged, safe-by-default
	// capability; flipping it on is an explicit out-of-band rollout step.
	SettingDeepIdentificationEnabled:            "false",
	SettingDeepIdentificationWorkerCount:        "2",
	SettingDeepIdentificationMaxActivePerUser:   "1",
	SettingDeepIdentificationQueueDepth:         "32",
	SettingDeepIdentificationHardTimeoutSeconds: "420",
	// 351-vision-first-deep-identification default (T011/T012): a full
	// vision LLM round trip through the Python service needs more than the
	// prior 15s magic literal; 90s is proportionate to the work performed
	// while remaining well under the 300s agent-proxy ceiling. Pairing this
	// with the raised 420s hard timeout above keeps the post-quick-lookup
	// pipeline budget at 420-90-20=310s, at or above the pre-change ~265s.
	SettingDeepIdentificationQuickLookupTimeoutSeconds: "90",
	SettingDeepIdentificationEventRetentionHours:       "24",
	SettingDeepIdentificationResultRetentionDays:       "90",
	SettingDeepIdentificationMaxProviders:              "4",
	SettingDeepIdentificationNumistaCallBudget:         "4",
	SettingDeepIdentificationRPCEnabled:                "false",
	SettingCoinCopilotEnabled:                          "false",
	SettingCoinCopilotAttributionEnabled:               "false",
	SettingCoinCopilotWorkerCount:                      "1",
	SettingCoinCopilotMaxActivePerUser:                 "1",
	SettingCoinCopilotQueueDepth:                       "16",
	SettingCoinCopilotMaxReasoningIterations:           "8",
	SettingCoinCopilotMaxToolCalls:                     "12",
	SettingCoinCopilotHardTimeoutSeconds:               "120",
	SettingCoinCopilotMaxPersistedToolResultBytes:      "32768",
	SettingCoinCopilotEventRetentionHours:              "168",
	SettingCoinCopilotCheckpointRetentionDays:          "30",
	SettingCoinCopilotResumeWindowHours:                "168",
}

type CoinCopilotSettings struct {
	Enabled                     bool
	AttributionEnabled          bool
	WorkerCount                 int
	MaxActivePerUser            int
	QueueDepth                  int
	MaxReasoningIterations      int
	MaxToolCalls                int
	HardTimeout                 time.Duration
	MaxPersistedToolResultBytes int
	EventRetention              time.Duration
	CheckpointRetention         time.Duration
	ResumeWindow                time.Duration
	Valid                       bool
}

func (s *SettingsService) GetCoinCopilotSettings() CoinCopilotSettings {
	valid := true
	readInt := func(key string, fallback, minimum, maximum int) int {
		value, err := strconv.Atoi(strings.TrimSpace(s.GetSetting(key)))
		if err != nil || value < minimum || value > maximum {
			valid = false
			return fallback
		}
		return value
	}
	readBool := func(key string, fallback bool) bool {
		switch strings.ToLower(strings.TrimSpace(s.GetSetting(key))) {
		case "true":
			return true
		case "false":
			return false
		default:
			valid = false
			return fallback
		}
	}
	timeoutSeconds := readInt(SettingCoinCopilotHardTimeoutSeconds, 120, 15, int(coinCopilotMaxExecutionTimeout.Seconds()))
	eventHours := readInt(SettingCoinCopilotEventRetentionHours, 168, 1, 720)
	checkpointDays := readInt(SettingCoinCopilotCheckpointRetentionDays, 30, 1, 365)
	resumeHours := readInt(SettingCoinCopilotResumeWindowHours, 168, 1, 720)
	return CoinCopilotSettings{
		Enabled:                     readBool(SettingCoinCopilotEnabled, false),
		AttributionEnabled:          readBool(SettingCoinCopilotAttributionEnabled, false),
		WorkerCount:                 readInt(SettingCoinCopilotWorkerCount, 1, 1, 4),
		MaxActivePerUser:            readInt(SettingCoinCopilotMaxActivePerUser, 1, 1, 3),
		QueueDepth:                  readInt(SettingCoinCopilotQueueDepth, 16, 1, 100),
		MaxReasoningIterations:      readInt(SettingCoinCopilotMaxReasoningIterations, 8, 1, 20),
		MaxToolCalls:                readInt(SettingCoinCopilotMaxToolCalls, 12, 1, 40),
		HardTimeout:                 time.Duration(timeoutSeconds) * time.Second,
		MaxPersistedToolResultBytes: readInt(SettingCoinCopilotMaxPersistedToolResultBytes, 32768, 4096, 131072),
		EventRetention:              time.Duration(eventHours) * time.Hour,
		CheckpointRetention:         time.Duration(checkpointDays) * 24 * time.Hour,
		ResumeWindow:                time.Duration(resumeHours) * time.Hour,
		Valid:                       valid,
	}
}

type NumistaSettings struct {
	SearchTTL         time.Duration
	DetailTTL         time.Duration
	EnrichmentLimit   int
	SearchResultLimit int
	SearchTimeout     time.Duration
	DetailTimeout     time.Duration
	Valid             bool
}

// GetNumistaSettings reads and validates live Numista settings. Invalid values
// fall back independently to documented defaults and mark the snapshot invalid.
func (s *SettingsService) GetNumistaSettings() NumistaSettings {
	valid := true
	readInt := func(key string, fallback, minimum, maximum int) int {
		value, err := strconv.Atoi(strings.TrimSpace(s.GetSetting(key)))
		if err != nil || value < minimum || value > maximum {
			valid = false
			return fallback
		}
		return value
	}
	searchTTLHours := readInt(SettingNumistaSearchTTLHours, 24, 1, 720)
	detailTTLHours := readInt(SettingNumistaDetailTTLHours, 168, 1, 2160)
	searchTimeoutSeconds := readInt(SettingNumistaSearchTimeoutSeconds, 4, 1, 10)
	detailTimeoutSeconds := readInt(SettingNumistaDetailTimeoutSeconds, 3, 1, 10)
	return NumistaSettings{
		SearchTTL:         time.Duration(searchTTLHours) * time.Hour,
		DetailTTL:         time.Duration(detailTTLHours) * time.Hour,
		EnrichmentLimit:   readInt(SettingNumistaEnrichmentLimit, 5, 1, 10),
		SearchResultLimit: readInt(SettingNumistaSearchResultLimit, 20, 1, 50),
		SearchTimeout:     time.Duration(searchTimeoutSeconds) * time.Second,
		DetailTimeout:     time.Duration(detailTimeoutSeconds) * time.Second,
		Valid:             valid,
	}
}

// DeepIdentificationSettings is a validated snapshot of the live worker
// pool / retention bounds for the deep agentic identification job service
// (data-model.md §8). Invalid values fall back independently to documented
// defaults and mark the snapshot invalid.
type DeepIdentificationSettings struct {
	Enabled          bool
	WorkerCount      int
	MaxActivePerUser int
	QueueDepth       int
	HardTimeout      time.Duration
	// QuickLookupTimeout bounds the quick-evidence extraction pass inside
	// Deep Analysis (351 T011/FR-038). It is consumed from the same ctx as
	// the overall HardTimeout, so it is a prefix of, not additional to,
	// that budget (see deepPipelineBounds and Run's remaining-budget check
	// in deep_identification_pipeline_runner.go).
	QuickLookupTimeout time.Duration
	EventRetention     time.Duration
	ResultRetention    time.Duration
	MaxProviders       int
	NumistaCallBudget  int
	RPCEnabled         bool
	Valid              bool
}

// DeepProviderConfiguration returns the sorted effective provider vocabulary
// and a generation digest over every setting that can alter provider work.
// Both values are server-owned snapshot inputs; callers cannot override them.
func (s *SettingsService) DeepProviderConfiguration() ([]string, string) {
	settings := s.GetDeepIdentificationSettings()
	providers := []string{"ngc", "numista"}

	if settings.RPCEnabled {
		providers = append(providers, "rpc")
	}
	sort.Strings(providers)
	generation := struct {
		Providers         []string `json:"providers"`
		MaxProviders      int      `json:"max_providers"`
		NumistaCallBudget int      `json:"numista_call_budget"`
		Valid             bool     `json:"valid"`
	}{
		Providers: providers, MaxProviders: settings.MaxProviders,
		NumistaCallBudget: settings.NumistaCallBudget, Valid: settings.Valid,
	}
	canonical, _ := json.Marshal(generation)
	sum := sha256.Sum256(canonical)
	return append([]string(nil), providers...), hex.EncodeToString(sum[:])
}

func (s *SettingsService) DeepProviderSettingSnapshot() map[string]string {
	keys := []string{
		SettingCoinCopilotEnabled,
		SettingCoinCopilotAttributionEnabled,
		SettingDeepIdentificationEnabled,
		SettingDeepIdentificationMaxProviders,
		SettingDeepIdentificationNumistaCallBudget,
		SettingDeepIdentificationRPCEnabled,
		SettingNumistaAPIKey,
	}
	snapshot := make(map[string]string, len(keys))
	for _, key := range keys {
		snapshot[key] = s.GetSetting(key)
	}
	return snapshot
}

// GetDeepIdentificationSettings reads and validates the live deep
// identification settings, mirroring the GetNumistaSettings pattern.
func (s *SettingsService) GetDeepIdentificationSettings() DeepIdentificationSettings {
	valid := true
	readInt := func(key string, fallback, minimum, maximum int) int {
		value, err := strconv.Atoi(strings.TrimSpace(s.GetSetting(key)))
		if err != nil || value < minimum || value > maximum {
			valid = false
			return fallback
		}
		return value
	}
	readBool := func(key string, fallback bool) bool {
		value := strings.TrimSpace(strings.ToLower(s.GetSetting(key)))
		switch value {
		case "true":
			return true
		case "false":
			return false
		default:
			valid = false
			return fallback
		}
	}
	hardTimeoutSeconds := readInt(SettingDeepIdentificationHardTimeoutSeconds, 420, 1, 900)
	quickLookupTimeoutSeconds := readInt(SettingDeepIdentificationQuickLookupTimeoutSeconds, 90, 5, 300)
	eventRetentionHours := readInt(SettingDeepIdentificationEventRetentionHours, 24, 1, 720)
	resultRetentionDays := readInt(SettingDeepIdentificationResultRetentionDays, 90, 1, 3650)
	return DeepIdentificationSettings{
		Enabled:            readBool(SettingDeepIdentificationEnabled, false),
		WorkerCount:        readInt(SettingDeepIdentificationWorkerCount, 2, 1, 32),
		MaxActivePerUser:   readInt(SettingDeepIdentificationMaxActivePerUser, 1, 1, 10),
		QueueDepth:         readInt(SettingDeepIdentificationQueueDepth, 32, 1, 1000),
		HardTimeout:        time.Duration(hardTimeoutSeconds) * time.Second,
		QuickLookupTimeout: time.Duration(quickLookupTimeoutSeconds) * time.Second,
		EventRetention:     time.Duration(eventRetentionHours) * time.Hour,
		ResultRetention:    time.Duration(resultRetentionDays) * 24 * time.Hour,
		MaxProviders:       readInt(SettingDeepIdentificationMaxProviders, 4, 1, 10),
		NumistaCallBudget:  readInt(SettingDeepIdentificationNumistaCallBudget, 4, 1, 20),
		RPCEnabled:         readBool(SettingDeepIdentificationRPCEnabled, false),
		Valid:              valid,
	}
}

// SettingsService provides access to application settings backed by the database.
type SettingsService struct {
	repo *repository.SettingsRepository
}

// NewSettingsService creates a new SettingsService.
func NewSettingsService(repo *repository.SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

// GetSetting returns the value for a given key, falling back to defaults.
func (s *SettingsService) GetSetting(key string) string {
	setting, err := s.repo.FindByKey(key)
	if err != nil {
		if def, ok := settingDefaults[key]; ok {
			return def
		}
		return ""
	}
	// Treat empty prompt settings as unset so the default is used.
	// AIProvider intentionally allows empty (means unconfigured).
	if setting.Value == "" && key != SettingAIProvider {
		if def, ok := settingDefaults[key]; ok {
			return def
		}
	}
	return setting.Value
}

// defaultPrompts are the AI prompts' built-in texts, by setting.
func defaultPrompts() map[string]string {
	return map[string]string{
		SettingCoinSearchPrompt:     DefaultCoinSearchPrompt,
		SettingCoinShowsPrompt:      DefaultCoinShowsPrompt,
		SettingValuationPrompt:      DefaultValuationPrompt,
		SettingObversePrompt:        DefaultObversePrompt,
		SettingReversePrompt:        DefaultReversePrompt,
		SettingTextExtractionPrompt: DefaultTextExtractionPrompt,
	}
}

// NormalizedPrompt is a prompt as compared: without the spaces around it,
// with plain line ends.
func NormalizedPrompt(prompt string) string {
	return strings.TrimSpace(strings.ReplaceAll(prompt, "\r\n", "\n"))
}

// SetSetting creates or updates a setting value.
func (s *SettingsService) SetSetting(key, value string) error {
	// A prompt saved as it comes is stored empty, so that its default, even a
	// later one, applies; the admin page saves every field it shows.
	if prompt, ok := defaultPrompts()[key]; ok && NormalizedPrompt(value) == NormalizedPrompt(prompt) {
		value = ""
	}
	if key == SettingDealerSearchSources || key == SettingAuctionSearchSources {
		normalized, err := NormalizeSearchSourceSetting(value)
		if err != nil {
			return err
		}
		// Sources saved as they come are stored empty too, so that a later
		// default list applies.
		value = normalized
		if normalized == settingDefaults[key] {
			value = ""
		}
	}
	if IsTimezoneSetting(key) {
		value = strings.TrimSpace(value)
		if err := ValidateTimezoneSetting(value); err != nil {
			return err
		}
	}
	return s.repo.Upsert(key, value)
}

// IsTimezoneSetting reports whether key holds an IANA time zone name.
func IsTimezoneSetting(key string) bool {
	return key == SettingScheduleTimezone
}

// ValidateTimezoneSetting accepts an empty value (server local time) or an
// IANA zone name such as "America/Chicago".
func ValidateTimezoneSetting(value string) error {
	if value == "" {
		return nil
	}
	if _, err := time.LoadLocation(value); err != nil {
		return fmt.Errorf("unknown time zone %q", value)
	}
	return nil
}

var searchSourceHostPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)+$`)

// NormalizeSearchSourceSetting validates and canonicalizes a newline-delimited
// list of public DNS hostnames used as outbound search boundaries.
func NormalizeSearchSourceSetting(value string) (string, error) {
	lines := SplitSettingList(value)
	if len(lines) == 0 || len(lines) > 20 {
		return "", fmt.Errorf("search sources must contain between 1 and 20 hostnames")
	}
	seen := make(map[string]struct{}, len(lines))
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		host := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(line), "."))
		if len(host) > 253 || net.ParseIP(host) != nil || !searchSourceHostPattern.MatchString(host) {
			return "", fmt.Errorf("invalid search source hostname %q", line)
		}
		if _, exists := seen[host]; exists {
			continue
		}
		seen[host] = struct{}{}
		normalized = append(normalized, host)
	}
	return strings.Join(normalized, "\n"), nil
}

func (s *SettingsService) GetSearchSources(key string) []string {
	normalized, err := NormalizeSearchSourceSetting(s.GetSetting(key))
	if err != nil {
		normalized = settingDefaults[key]
	}
	return SplitSettingList(normalized)
}

// SplitSettingList parses a newline-delimited AppSetting value (the shape
// used by CoinCategories/CoinEras and similar admin-defined lists) into a
// trimmed, non-empty slice, mirroring how the frontend's parseOptionList
// treats the same setting shape.
func SplitSettingList(value string) []string {
	var out []string
	for _, line := range strings.Split(value, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// ParseSettingLabelMap parses a newline-delimited "Value=Label" AppSetting (the
// shape of CoinCategoryLabelsPtPT/CoinEraLabelsPtPT) into a map keyed by the
// stored value. Only the first "=" separates value from label, blank or
// malformed lines are skipped, and a later line for the same value wins.
func ParseSettingLabelMap(value string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(value, "\n") {
		key, label, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		label = strings.TrimSpace(label)
		if key == "" || label == "" {
			continue
		}
		out[key] = label
	}
	return out
}

// GetAllSettings returns all settings merged with defaults.
func (s *SettingsService) GetAllSettings() map[string]string {
	result := make(map[string]string)
	for k, v := range settingDefaults {
		result[k] = v
	}

	settings, _ := s.repo.FindAll()
	for _, st := range settings {
		if st.Value != "" {
			result[st.Key] = st.Value
		}
	}
	return result
}

// GetSettingDefaults returns the built-in default values for all settings.
func (s *SettingsService) GetSettingDefaults() map[string]string {
	result := make(map[string]string)
	for k, v := range settingDefaults {
		result[k] = v
	}
	return result
}

// SyncLogLevel reads the LogLevel setting from the DB and applies it to the logger.
func (s *SettingsService) SyncLogLevel(logger *Logger) {
	level := s.GetSetting(SettingLogLevel)
	logger.SetLevel(level)
}

// ResolveLLMConfig reads AI provider settings and returns a configured LLMConfig.
func (s *SettingsService) ResolveLLMConfig() (LLMConfig, error) {
	provider := s.GetSetting(SettingAIProvider)
	if provider == "" {
		return LLMConfig{}, fmt.Errorf("AI provider not configured. Please select Anthropic or Ollama in Admin Settings.")
	}

	cfg := LLMConfig{
		Provider: provider,
	}

	switch provider {
	case "anthropic":
		cfg.APIKey = s.GetSetting(SettingAnthropicAPIKey)
		cfg.Model = s.GetSetting(SettingAnthropicModel)
		if cfg.APIKey == "" {
			return LLMConfig{}, fmt.Errorf("Anthropic API key is required")
		}
	case "ollama":
		cfg.Model = s.GetSetting(SettingOllamaModel)
		cfg.OllamaURL = s.GetSetting(SettingOllamaURL)
		cfg.SearXNGURL = s.GetSetting(SettingSearXNGURL)
	default:
		return LLMConfig{}, fmt.Errorf("unknown AI provider: %s", provider)
	}

	return cfg, nil
}
