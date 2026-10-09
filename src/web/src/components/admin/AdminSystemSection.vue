<template>
  <section class="card">
    <h2 class="mb-5 border-b border-border-subtle pb-3 text-xl font-medium">{{ t('admin.system.title') }}</h2>
    <form @submit.prevent="save">
      <div class="form-group">
        <label class="form-label">{{ t('admin.system.pushoverToken') }}</label>
        <input v-model="localPushoverAppToken" class="form-input" type="password" :placeholder="t('admin.system.pushoverTokenPlaceholder')" />
        <span class="mt-1 block text-sm text-text-muted"><I18nT keypath="admin.system.pushoverHint"><template #link><a href="https://pushover.net/apps" target="_blank" rel="noopener">pushover.net/apps</a></template></I18nT></span>
      </div>
      <div class="form-group">
        <label class="form-label">{{ t('admin.security.checks.publicUrl') }}</label>
        <input v-model="localPublicAppUrl" class="form-input" type="url" placeholder="https://coins.example.com" />
        <span class="mt-1 block text-sm text-text-muted">{{ t('admin.system.publicUrlHint') }}</span>
      </div>
      <div class="form-group">
        <label class="form-label">{{ t('admin.system.logLevel') }}</label>
        <select v-model="localLogLevel" class="form-select">
          <option v-for="level in logLevels" :key="level" :value="level">{{ level }}</option>
        </select>
      </div>

      <section
        class="mt-6 min-w-0 rounded-md border border-border-subtle bg-input p-4 md:p-6"
        aria-labelledby="search-sources-heading"
        data-testid="search-sources-section"
      >
        <div class="mb-5">
          <p class="section-label">{{ t('admin.system.externalSearch') }}</p>
          <h3 id="search-sources-heading" class="m-0 text-lg font-medium text-heading">{{ t('admin.system.searchSources') }}</h3>
        </div>
        <div class="grid min-w-0 gap-4 md:grid-cols-2">
          <div class="form-group min-w-0">
            <label class="form-label" for="dealer-search-sources">{{ t('admin.system.dealerSources') }}</label>
            <textarea
              id="dealer-search-sources"
              v-model="localDealerSearchSources"
              class="form-textarea"
              rows="7"
              spellcheck="false"
              required
            />
            <span class="mt-1 block text-sm text-text-muted">{{ t('admin.system.dealerSourcesHint') }}</span>
          </div>
          <div class="form-group min-w-0">
            <label class="form-label" for="auction-search-sources">{{ t('admin.system.auctionSources') }}</label>
            <textarea
              id="auction-search-sources"
              v-model="localAuctionSearchSources"
              class="form-textarea"
              rows="7"
              spellcheck="false"
              required
            />
            <span class="mt-1 block text-sm text-text-muted">{{ t('admin.system.auctionSourcesHint') }}</span>
          </div>
        </div>
      </section>

      <section
        class="mt-6 min-w-0 rounded-md border border-border-subtle bg-input p-4 md:p-6"
        aria-labelledby="numista-section-heading"
        data-testid="numista-section"
      >
        <div class="mb-5">
          <p class="section-label">{{ t('admin.system.catalogIntegration') }}</p>
          <h3 id="numista-section-heading" class="m-0 text-lg font-medium text-heading">Numista</h3>
        </div>

        <div class="form-group">
          <label class="form-label" for="numista-api-key">{{ t('admin.system.numistaKey') }}</label>
          <input
            id="numista-api-key"
            v-model="localNumistaApiKey"
            class="form-input"
            type="password"
            autocomplete="off"
            aria-describedby="numista-api-key-help"
            :placeholder="t('admin.system.numistaKeyPlaceholder')"
          />
          <span id="numista-api-key-help" class="mt-1 block text-sm text-text-muted"><I18nT keypath="admin.system.numistaKeyHint"><template #link><a href="https://en.numista.com/api/" target="_blank" rel="noopener">numista.com/api</a></template></I18nT></span>
        </div>

        <fieldset class="mb-6 min-w-0">
          <legend class="section-label mb-3">{{ t('admin.system.numistaLimits') }}</legend>
          <div class="grid min-w-0 gap-4 md:grid-cols-2">
            <div v-for="setting in numistaSettingFields" :key="setting.name" class="form-group min-w-0">
              <label class="form-label" :for="setting.name">{{ t(setting.labelKey) }}</label>
              <input
                :id="setting.name"
                v-model="setting.model.value"
                class="form-input"
                type="number"
                :name="setting.name"
                :min="setting.min"
                :max="setting.max"
                step="1"
                required
              />
              <span class="mt-1 block text-sm text-text-muted">{{ t(setting.hintKey, { min: setting.min, max: setting.max, default: setting.fallback }) }}</span>
            </div>
          </div>
        </fieldset>

        <section class="border-t border-border-subtle pt-6" aria-labelledby="numista-health-heading">
          <div class="mb-4 flex flex-wrap items-center justify-between gap-2">
            <h3 id="numista-health-heading" class="m-0 text-lg font-medium text-heading">{{ t('admin.system.numistaHealth') }}</h3>
            <button type="button" class="btn btn-secondary btn-xs" :disabled="healthLoading" @click="loadHealth">
              {{ healthLoading ? t('admin.system.refreshing') : t('admin.logs.refresh') }}
            </button>
          </div>

          <div v-if="healthLoading" class="flex items-center gap-3 py-6 text-sm text-text-secondary" role="status">
            <span class="spinner"></span>
            {{ t('admin.system.loadingNumista') }}
          </div>
          <p v-else-if="healthError" class="rounded-sm border border-[var(--color-negative)] p-3 text-sm text-loss" role="alert">
            {{ t('admin.system.numistaUnavailable') }}
          </p>
          <div v-else-if="health && !hasHealthEvents" class="rounded-sm border border-border-subtle bg-card p-4" role="status">
            <h4 class="section-label mb-2">{{ t('admin.system.noRecentActivity') }}</h4>
            <p class="m-0 text-sm text-text-primary">{{ t('admin.system.noNumistaEvents') }}</p>
            <p class="mb-0 mt-1 text-sm text-text-secondary">
              {{ health.configured ? t('admin.system.numistaConfigured') : t('admin.system.numistaNotConfigured') }}
              {{ t('admin.system.refreshAfterLookup') }}
            </p>
          </div>
          <div v-else-if="health" class="grid gap-4">
            <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              <div class="rounded-sm border border-border-subtle bg-card p-3">
                <div class="section-label">{{ t('admin.groups.configuration') }}</div>
                <div class="mt-1 font-semibold text-text-primary">{{ health.configured ? t('admin.oidc.configured') : t('admin.oidc.notConfigured') }}</div>
                <div class="mt-1 text-sm text-text-secondary">{{ health.configurationValid ? t('admin.system.configValid') : t('admin.system.configInvalid') }}</div>
              </div>
              <div class="rounded-sm border border-border-subtle bg-card p-3">
                <div class="section-label">{{ t('admin.system.requests') }}</div>
                <div class="mt-1 text-sm text-text-primary">{{ t('admin.system.broad', { count: health.broadRequestCount }) }}</div>
                <div class="mt-1 text-sm text-text-secondary">{{ t('admin.system.detail', { count: health.detailRequestCount }) }}</div>
              </div>
              <div class="rounded-sm border border-border-subtle bg-card p-3">
                <div class="section-label">{{ t('admin.system.latency') }}</div>
                <div class="mt-1 text-sm text-text-primary">p50 {{ health.p50ElapsedMs }} ms</div>
                <div class="mt-1 text-sm text-text-secondary">p95 {{ health.p95ElapsedMs }} ms</div>
              </div>
              <div class="rounded-sm border border-border-subtle bg-card p-3">
                <div class="section-label">{{ t('admin.system.latestOutcome') }}</div>
                <div class="mt-1 font-semibold text-text-primary">{{ formatStatus(health.lastOutcome) }}</div>
                <div class="mt-1 text-sm text-text-secondary">{{ formatTimestamp(health.lastCheckedAt) }}</div>
              </div>
            </div>

            <div class="rounded-sm border border-border-subtle bg-card p-4">
              <h4 class="section-label mb-3">{{ t('admin.system.statusCounts') }}</h4>
              <div class="grid grid-cols-2 gap-2 sm:grid-cols-3">
                <div v-for="status in statusRows" :key="status.key" class="flex items-center justify-between gap-2 rounded-sm bg-input px-3 py-2">
                  <span class="text-sm text-text-secondary">{{ t(status.labelKey) }}</span>
                  <strong class="text-gold">{{ health.statusCounts[status.key] ?? 0 }}</strong>
                </div>
              </div>
            </div>

            <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              <div class="rounded-sm border border-border-subtle bg-card p-4">
                <h4 class="section-label mb-2">{{ t('admin.system.cache') }}</h4>
                <p class="m-0 text-sm text-text-primary">{{ t('admin.system.freshHits', { count: health.freshCacheHitCount }) }}</p>
                <p class="mb-0 mt-1 text-sm text-text-secondary">{{ t('admin.system.coalesced', { count: health.coalescedRequestCount, rate: formatPercent(health.freshCacheHitRate) }) }}</p>
              </div>
              <div class="rounded-sm border border-border-subtle bg-card p-4">
                <h4 class="section-label mb-2">{{ t('admin.system.providerLoads') }}</h4>
                <p class="m-0 text-sm text-text-primary">{{ t('admin.system.loads', { count: health.providerLoadCount }) }}</p>
                <p class="mb-0 mt-1 text-sm text-text-secondary">{{ t('admin.system.failedCancelled', { failed: health.providerFailureCount, cancelled: health.cancelledRequestCount }) }}</p>
              </div>
              <div class="rounded-sm border border-border-subtle bg-card p-4">
                <h4 class="section-label mb-2">{{ t('admin.system.enrichment') }}</h4>
                <p class="m-0 text-sm text-text-primary">{{ t('admin.system.attempted', { count: health.enrichmentAttempted }) }}</p>
                <p class="mb-0 mt-1 text-sm text-text-secondary">{{ t('admin.system.succeededFailed', { succeeded: health.enrichmentSucceeded, failed: health.enrichmentFailed }) }}</p>
              </div>
              <div class="rounded-sm border border-border-subtle bg-card p-4">
                <h4 class="section-label mb-2">{{ t('admin.system.quotaSignals') }}</h4>
                <p class="m-0 text-sm text-text-primary">{{ t('admin.system.lastLimited', { value: formatTimestamp(health.lastQuotaLimitedAt) }) }}</p>
                <p class="mb-0 mt-1 text-sm text-text-secondary">{{ t('admin.system.retryAfter', { value: formatRetryAfter(health.lastRetryAfterSeconds) }) }}</p>
              </div>
            </div>
          </div>
          <p v-else class="py-6 text-sm text-text-muted">{{ t('admin.system.noNumistaData') }}</p>
        </section>
      </section>

      <section
        class="mt-6 min-w-0 rounded-md border border-border-subtle bg-input p-4 md:p-6"
        aria-labelledby="coin-copilot-section-heading"
        data-testid="coin-copilot-section"
      >
        <div class="mb-5">
          <p class="section-label">{{ t('admin.system.betaHarness') }}</p>
          <h3 id="coin-copilot-section-heading" class="m-0 text-lg font-medium text-heading">Coin Copilot</h3>
          <p class="mt-1 text-sm text-text-muted">
            {{ t('admin.system.copilotIntro') }}
          </p>
        </div>

        <div class="form-group min-w-0">
          <label class="flex items-center gap-3" for="coin-copilot-enabled">
            <input
              id="coin-copilot-enabled"
              v-model="localCoinCopilotEnabled"
              type="checkbox"
              name="CoinCopilotEnabled"
              aria-describedby="coin-copilot-enabled-help"
            />
            <span class="form-label m-0">{{ t('admin.system.enableCopilot') }}</span>
          </label>
          <span id="coin-copilot-enabled-help" class="mt-1 block text-sm text-text-muted">
            {{ t('admin.system.enableCopilotHint') }}
          </span>
        </div>

        <div class="form-group min-w-0">
          <label class="flex items-center gap-3" for="coin-copilot-attribution-enabled">
            <input
              id="coin-copilot-attribution-enabled"
              v-model="localCoinCopilotAttributionEnabled"
              type="checkbox"
              name="CoinCopilotAttributionEnabled"
              aria-describedby="coin-copilot-attribution-enabled-help"
            />
            <span class="form-label m-0">{{ t('admin.system.allowDeep') }}</span>
          </label>
          <span id="coin-copilot-attribution-enabled-help" class="mt-1 block text-sm text-text-muted">
            {{ t('admin.system.allowDeepHint') }}
          </span>
          <span
            v-if="localCoinCopilotAttributionEnabled && !localDeepIdentificationEnabled"
            class="mt-1 block text-sm text-gold"
            data-testid="coin-copilot-attribution-dependency-hint"
          >
            {{ t('admin.system.allowDeepDependency') }}
          </span>
        </div>

        <fieldset class="mb-2 min-w-0">
          <legend class="section-label mb-3">{{ t('admin.system.copilotLimits') }}</legend>
          <div class="grid min-w-0 gap-4 md:grid-cols-2">
            <div v-for="setting in coinCopilotSettingFields" :key="setting.name" class="form-group min-w-0">
              <label class="form-label" :for="setting.name">{{ t(setting.labelKey) }}</label>
              <input
                :id="setting.name"
                v-model="setting.model.value"
                class="form-input"
                type="number"
                :name="setting.name"
                :min="setting.min"
                :max="setting.max"
                step="1"
                required
              />
              <span class="mt-1 block text-sm text-text-muted">{{ t(setting.hintKey, { min: setting.min, max: setting.max, default: setting.fallback }) }}</span>
            </div>
          </div>
        </fieldset>
      </section>

      <section
        class="mt-6 min-w-0 rounded-md border border-border-subtle bg-input p-4 md:p-6"
        aria-labelledby="ocre-section-heading"
        data-testid="ocre-section"
      >
        <div class="mb-5">
          <p class="section-label">{{ t('admin.system.deepProvider') }}</p>
          <h3 id="ocre-section-heading" class="m-0 text-lg font-medium text-heading">{{ t('admin.system.ocreTitle') }}</h3>
          <p class="mt-1 text-sm text-text-muted">
            {{ t('admin.system.ocreIntro') }}
          </p>
        </div>

        <div class="form-group min-w-0">
          <label class="flex items-center gap-3">
            <input
              id="deep-identification-enabled"
              v-model="localDeepIdentificationEnabled"
              type="checkbox"
              name="DeepIdentificationEnabled"
              data-testid="deep-identification-enabled-toggle"
              aria-describedby="deep-identification-enabled-help"
            />
            <span class="form-label m-0">{{ t('admin.system.enableDeep') }}</span>
          </label>
          <span id="deep-identification-enabled-help" class="mt-1 block text-sm text-text-muted">
            {{ t('admin.system.enableDeepHint') }}
          </span>
        </div>

        <fieldset class="mb-6 min-w-0">
          <legend class="section-label mb-3">{{ t('admin.system.deepLimits') }}</legend>
          <div class="grid min-w-0 gap-4 md:grid-cols-2">
            <div v-for="setting in deepIdentificationSettingFields" :key="setting.name" class="form-group min-w-0">
              <label class="form-label" :for="setting.name">{{ t(setting.labelKey) }}</label>
              <input
                :id="setting.name"
                v-model="setting.model.value"
                class="form-input"
                type="number"
                :name="setting.name"
                :min="setting.min"
                :max="setting.max"
                step="1"
                required
              />
              <span class="mt-1 block text-sm text-text-muted">{{ t(setting.hintKey, { min: setting.min, max: setting.max, default: setting.fallback }) }}</span>
            </div>
          </div>
        </fieldset>

        <div class="form-group min-w-0">
          <label class="flex items-center gap-3">
            <input
              id="ocre-enabled"
              v-model="localOCREEnabled"
              type="checkbox"
              name="DeepIdentificationOCREEnabled"
              data-testid="ocre-enabled-toggle"
              aria-describedby="ocre-enabled-help"
            />
            <span class="form-label m-0">{{ t('admin.system.enableOcre') }}</span>
          </label>
          <span id="ocre-enabled-help" class="mt-1 block text-sm text-text-muted">
            {{ t('admin.system.enableOcreHint') }}
          </span>
        </div>

        <fieldset class="mb-2 min-w-0">
          <legend class="section-label mb-3">{{ t('admin.system.ocreLimits') }}</legend>
          <div class="grid min-w-0 gap-4 md:grid-cols-2">
            <div class="form-group min-w-0">
              <label class="form-label" for="ocre-call-budget">{{ t('admin.system.perJobBudget') }}</label>
              <input
                id="ocre-call-budget"
                v-model="localOCRECallBudget"
                class="form-input"
                type="number"
                name="DeepIdentificationOCRECallBudget"
                :min="1"
                :max="20"
                step="1"
                required
              />
              <span class="mt-1 block text-sm text-text-muted">{{ t('admin.system.hints.requestsPerJob', { min: 1, max: 20, default: 3 }) }}</span>
            </div>
          </div>
        </fieldset>

        <section class="border-t border-border-subtle pt-6" aria-labelledby="ocre-health-heading">
          <div class="mb-4 flex items-center justify-between gap-3">
            <h3 id="ocre-health-heading" class="m-0 text-lg font-medium text-heading">{{ t('admin.system.ocreHealth') }}</h3>
            <button type="button" class="btn btn-secondary btn-xs" :disabled="ocreHealthLoading" @click="loadOCREHealth">
              {{ ocreHealthLoading ? t('admin.system.refreshing') : t('admin.logs.refresh') }}
            </button>
          </div>

          <div v-if="ocreHealthLoading" class="flex items-center gap-3 py-6 text-sm text-text-secondary" role="status">
            <span class="sr-only">{{ t('admin.system.loadingOcre') }}</span>
            {{ t('admin.system.loadingOcre') }}
          </div>
          <p v-else-if="ocreHealthError" class="rounded-sm border border-[var(--color-negative)] p-3 text-sm text-loss" role="alert">
            {{ t('admin.system.ocreUnavailable') }}
          </p>
          <div v-else-if="ocreHealth" class="grid gap-4 md:grid-cols-3" data-testid="ocre-health">
            <div class="rounded-sm border border-border-subtle bg-card p-4">
              <p class="section-label m-0">{{ t('admin.system.provider') }}</p>
              <div class="mt-1 font-semibold text-text-primary">{{ ocreHealth.enabled ? t('admin.oidc.enabled') : t('admin.oidc.disabled') }}</div>
              <div class="mt-1 text-sm text-text-secondary">{{ ocreHealth.gateValidated ? t('admin.system.configValid') : t('admin.system.configInvalid') }}</div>
            </div>
            <div class="rounded-sm border border-border-subtle bg-card p-4">
              <p class="section-label m-0">{{ t('admin.system.callBudget') }}</p>
              <div class="mt-1 text-sm text-text-primary">{{ t('admin.system.perJob', { count: ocreHealth.callBudget }) }}</div>
            </div>
            <div class="rounded-sm border border-border-subtle bg-card p-4">
              <p class="section-label m-0">{{ t('admin.system.lastOutcome') }}</p>
              <div class="mt-1 font-semibold text-text-primary">{{ formatOCREOutcome(ocreHealth.lastOutcome) }}</div>
              <div class="mt-1 text-sm text-text-secondary">{{ formatTimestamp(ocreHealth.lastCheckedAt) }}</div>
            </div>
          </div>
          <p v-else class="py-6 text-sm text-text-muted">{{ t('admin.system.noOcreData') }}</p>
        </section>

        <section class="mt-6 border-t border-border-subtle pt-6" aria-labelledby="deep-observability-heading">
          <div class="mb-4 flex items-center justify-between gap-3">
            <h3 id="deep-observability-heading" class="m-0 text-lg font-medium text-heading">{{ t('admin.system.deepOps') }}</h3>
            <button type="button" class="btn btn-secondary btn-xs min-h-[44px]" :disabled="deepMetricsLoading" @click="loadDeepMetrics">
              {{ deepMetricsLoading ? t('admin.system.refreshing') : t('admin.logs.refresh') }}
            </button>
          </div>
          <div v-if="deepMetricsLoading" class="py-6 text-sm text-text-secondary" role="status">{{ t('admin.system.loadingDeepOps') }}</div>
          <p v-else-if="deepMetricsError" class="rounded-sm border border-[var(--color-negative)] p-3 text-sm text-loss" role="alert">
            {{ t('admin.system.deepOpsUnavailable') }}
          </p>
          <div v-else-if="deepMetrics" class="grid gap-4" data-testid="deep-observability">
            <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              <div class="rounded-sm border border-border-subtle bg-card p-4">
                <p class="section-label m-0">{{ t('admin.system.terminalJobs') }}</p>
                <div class="mt-1 text-sm text-text-primary">{{ terminalJobCount }}</div>
              </div>
              <div class="rounded-sm border border-border-subtle bg-card p-4">
                <p class="section-label m-0">{{ t('admin.system.partialSuccess') }}</p>
                <div class="mt-1 text-sm text-text-primary">{{ formatRate(deepMetrics.partialSuccessRate) }}</div>
              </div>
              <div class="rounded-sm border border-border-subtle bg-card p-4">
                <p class="section-label m-0">{{ t('admin.system.jobDuration') }}</p>
                <div class="mt-1 text-sm text-text-primary">p50 {{ formatDuration(deepMetrics.duration.p50Ms) }} · p95 {{ formatDuration(deepMetrics.duration.p95Ms) }}</div>
              </div>
              <div class="rounded-sm border border-border-subtle bg-card p-4">
                <p class="section-label m-0">{{ t('admin.system.queueStreams') }}</p>
                <div class="mt-1 text-sm text-text-primary">{{ t('admin.system.queuedLive', { queued: deepMetrics.queueDepth, live: deepMetrics.activeSseStreams }) }}</div>
              </div>
            </div>
            <div class="overflow-x-auto">
              <table class="data-table min-w-[540px] text-sm [&_th]:p-2">
                <thead>
                  <tr><th>{{ t('admin.system.provider') }}</th><th>{{ t('admin.system.outcomes') }}</th><th>{{ t('admin.system.latency') }}</th></tr>
                </thead>
                <tbody>
                  <tr v-for="(metrics, provider) in deepMetrics.providers" :key="provider">
                    <td class="p-2 font-semibold uppercase text-text-primary">{{ provider }}</td>
                    <td class="p-2 text-text-secondary">{{ formatStatusCounts(metrics.statusCounts) }}</td>
                    <td class="p-2 text-text-secondary">p50 {{ formatDuration(metrics.latency.p50Ms) }} · p95 {{ formatDuration(metrics.latency.p95Ms) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p class="m-0 text-sm text-text-secondary">
              {{ t('admin.system.deepOpsSummary', {
                reconnects: deepMetrics.reconnectCount,
                truncations: deepMetrics.truncationCount,
                hintSuccess: deepMetrics.hintDeletion.success,
                hintFailure: deepMetrics.hintDeletion.failure,
                recovery: deepMetrics.janitor.recoverySweeps,
                retention: deepMetrics.janitor.retentionSweeps,
                failures: deepMetrics.janitor.failures,
              }) }}
            </p>
          </div>
        </section>
      </section>

      <p v-if="msg" class="my-2 text-body" :class="error ? 'text-loss' : 'text-gold'">{{ msg }}</p>
      <button type="submit" class="btn btn-primary btn-sm" :disabled="saving">
        {{ saving ? t('common.saving') : t('admin.system.save') }}
      </button>
    </form>

    <div class="mt-6 flex items-center gap-2 border-t border-border-subtle pt-4 text-chip text-text-muted">
      <span class="font-semibold uppercase tracking-[0.08em]">{{ t('admin.system.version') }}</span>
      <span class="font-mono text-text-secondary">{{ appVersion }}</span>
      <span v-if="buildDate" class="ml-1">{{ t('admin.system.built', { date: buildDate }) }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { getAdminDeepIdentificationObservability, getAdminNumistaHealth, getAdminOCREHealth } from '@/api/client'
import type { DeepIdentificationObservabilitySummary, NumistaHealthSummary, NumistaLookupStatus, OCREHealthSummary } from '@/types'
import I18nT from '@/components/I18nT.vue'
import { t, intlLocale, type MessageKey } from '@/i18n'
import { deepProviderStatusLabel } from '@/utils/format'

const props = withDefaults(defineProps<{
  numistaApiKey: string
  numistaSearchTTLHours: string
  numistaDetailTTLHours: string
  numistaEnrichmentLimit: string
  numistaSearchResultLimit: string
  numistaSearchTimeoutSeconds: string
  numistaDetailTimeoutSeconds: string
  deepIdentificationEnabled?: string
  deepIdentificationWorkerCount?: string
  deepIdentificationMaxActivePerUser?: string
  deepIdentificationQueueDepth?: string
  deepIdentificationHardTimeoutSeconds?: string
  deepIdentificationEventRetentionHours?: string
  deepIdentificationResultRetentionDays?: string
  deepIdentificationMaxProviders?: string
  deepIdentificationNumistaCallBudget?: string
  deepIdentificationOCREEnabled: string
  deepIdentificationOCRECallBudget: string
  coinCopilotEnabled?: string
  coinCopilotAttributionEnabled?: string
  coinCopilotWorkerCount?: string
  coinCopilotMaxActivePerUser?: string
  coinCopilotQueueDepth?: string
  coinCopilotMaxReasoningIterations?: string
  coinCopilotMaxToolCalls?: string
  coinCopilotHardTimeoutSeconds?: string
  coinCopilotMaxPersistedToolResultBytes?: string
  coinCopilotEventRetentionHours?: string
  coinCopilotCheckpointRetentionDays?: string
  coinCopilotResumeWindowHours?: string
  pushoverAppToken: string
  publicAppUrl: string
  dealerSearchSources?: string
  auctionSearchSources?: string
  uspsApiBaseUrl: string
  uspsApiKey: string
  uspsApiKeyHeader: string
  upsApiBaseUrl: string
  upsTokenUrl: string
  upsClientId: string
  upsClientSecret: string
  upsScope: string
  fedexApiBaseUrl: string
  fedexTokenUrl: string
  fedexClientId: string
  fedexClientSecret: string
  fedexScope: string
  logLevel: string
  logLevels: readonly string[]
  saving: boolean
  msg: string
  error: boolean
  appVersion: string
  buildDate: string
}>(), {
  deepIdentificationEnabled: 'false',
  deepIdentificationWorkerCount: '2',
  deepIdentificationMaxActivePerUser: '1',
  deepIdentificationQueueDepth: '32',
  deepIdentificationHardTimeoutSeconds: '300',
  deepIdentificationEventRetentionHours: '24',
  deepIdentificationResultRetentionDays: '90',
  deepIdentificationMaxProviders: '4',
  deepIdentificationNumistaCallBudget: '4',
  coinCopilotEnabled: 'false',
  coinCopilotAttributionEnabled: 'false',
  coinCopilotWorkerCount: '1',
  coinCopilotMaxActivePerUser: '1',
  coinCopilotQueueDepth: '16',
  coinCopilotMaxReasoningIterations: '8',
  coinCopilotMaxToolCalls: '12',
  coinCopilotHardTimeoutSeconds: '120',
  coinCopilotMaxPersistedToolResultBytes: '32768',
  coinCopilotEventRetentionHours: '168',
  coinCopilotCheckpointRetentionDays: '30',
  coinCopilotResumeWindowHours: '168',
  dealerSearchSources: '',
  auctionSearchSources: '',
})

const emit = defineEmits<{
  save: [settings: {
    numistaApiKey: string
    numistaSearchTTLHours: string
    numistaDetailTTLHours: string
    numistaEnrichmentLimit: string
    numistaSearchResultLimit: string
    numistaSearchTimeoutSeconds: string
    numistaDetailTimeoutSeconds: string
    deepIdentificationEnabled: string
    deepIdentificationWorkerCount: string
    deepIdentificationMaxActivePerUser: string
    deepIdentificationQueueDepth: string
    deepIdentificationHardTimeoutSeconds: string
    deepIdentificationEventRetentionHours: string
    deepIdentificationResultRetentionDays: string
    deepIdentificationMaxProviders: string
    deepIdentificationNumistaCallBudget: string
    deepIdentificationOCREEnabled: string
    deepIdentificationOCRECallBudget: string
    coinCopilotEnabled: string
    coinCopilotAttributionEnabled: string
    coinCopilotWorkerCount: string
    coinCopilotMaxActivePerUser: string
    coinCopilotQueueDepth: string
    coinCopilotMaxReasoningIterations: string
    coinCopilotMaxToolCalls: string
    coinCopilotHardTimeoutSeconds: string
    coinCopilotMaxPersistedToolResultBytes: string
    coinCopilotEventRetentionHours: string
    coinCopilotCheckpointRetentionDays: string
    coinCopilotResumeWindowHours: string
    logLevel: string
    pushoverAppToken: string
    publicAppUrl: string
    dealerSearchSources: string
    auctionSearchSources: string
    uspsApiBaseUrl: string
    uspsApiKey: string
    uspsApiKeyHeader: string
    upsApiBaseUrl: string
    upsTokenUrl: string
    upsClientId: string
    upsClientSecret: string
    upsScope: string
    fedexApiBaseUrl: string
    fedexTokenUrl: string
    fedexClientId: string
    fedexClientSecret: string
    fedexScope: string
  }]
}>()

const localNumistaApiKey = ref(props.numistaApiKey)
const localNumistaSearchTTLHours = ref(props.numistaSearchTTLHours || '24')
const localNumistaDetailTTLHours = ref(props.numistaDetailTTLHours || '168')
const localNumistaEnrichmentLimit = ref(props.numistaEnrichmentLimit || '5')
const localNumistaSearchResultLimit = ref(props.numistaSearchResultLimit || '20')
const localNumistaSearchTimeoutSeconds = ref(props.numistaSearchTimeoutSeconds || '4')
const localNumistaDetailTimeoutSeconds = ref(props.numistaDetailTimeoutSeconds || '3')
const localDeepIdentificationEnabled = ref((props.deepIdentificationEnabled || 'false') === 'true')
const localDeepIdentificationWorkerCount = ref(props.deepIdentificationWorkerCount)
const localDeepIdentificationMaxActivePerUser = ref(props.deepIdentificationMaxActivePerUser)
const localDeepIdentificationQueueDepth = ref(props.deepIdentificationQueueDepth)
const localDeepIdentificationHardTimeoutSeconds = ref(props.deepIdentificationHardTimeoutSeconds)
const localDeepIdentificationEventRetentionHours = ref(props.deepIdentificationEventRetentionHours)
const localDeepIdentificationResultRetentionDays = ref(props.deepIdentificationResultRetentionDays)
const localDeepIdentificationMaxProviders = ref(props.deepIdentificationMaxProviders)
const localDeepIdentificationNumistaCallBudget = ref(props.deepIdentificationNumistaCallBudget)
const localOCREEnabled = ref((props.deepIdentificationOCREEnabled || 'false') === 'true')
const localOCRECallBudget = ref(props.deepIdentificationOCRECallBudget || '3')
const localCoinCopilotEnabled = ref((props.coinCopilotEnabled || 'false') === 'true')
const localCoinCopilotAttributionEnabled = ref((props.coinCopilotAttributionEnabled || 'false') === 'true')
const localCoinCopilotWorkerCount = ref(props.coinCopilotWorkerCount)
const localCoinCopilotMaxActivePerUser = ref(props.coinCopilotMaxActivePerUser)
const localCoinCopilotQueueDepth = ref(props.coinCopilotQueueDepth)
const localCoinCopilotMaxReasoningIterations = ref(props.coinCopilotMaxReasoningIterations)
const localCoinCopilotMaxToolCalls = ref(props.coinCopilotMaxToolCalls)
const localCoinCopilotHardTimeoutSeconds = ref(props.coinCopilotHardTimeoutSeconds)
const localCoinCopilotMaxPersistedToolResultBytes = ref(props.coinCopilotMaxPersistedToolResultBytes)
const localCoinCopilotEventRetentionHours = ref(props.coinCopilotEventRetentionHours)
const localCoinCopilotCheckpointRetentionDays = ref(props.coinCopilotCheckpointRetentionDays)
const localCoinCopilotResumeWindowHours = ref(props.coinCopilotResumeWindowHours)
const localPushoverAppToken = ref(props.pushoverAppToken)
const localPublicAppUrl = ref(props.publicAppUrl)
const localDealerSearchSources = ref(props.dealerSearchSources)
const localAuctionSearchSources = ref(props.auctionSearchSources)
const localUSPSAPIBaseURL = ref(props.uspsApiBaseUrl)
const localUSPSAPIKey = ref(props.uspsApiKey)
const localUSPSAPIKeyHeader = ref(props.uspsApiKeyHeader)
const localUPSAPIBaseURL = ref(props.upsApiBaseUrl)
const localUPSTokenURL = ref(props.upsTokenUrl)
const localUPSClientID = ref(props.upsClientId)
const localUPSClientSecret = ref(props.upsClientSecret)
const localUPSScope = ref(props.upsScope)
const localFedExAPIBaseURL = ref(props.fedexApiBaseUrl)
const localFedExTokenURL = ref(props.fedexTokenUrl)
const localFedExClientID = ref(props.fedexClientId)
const localFedExClientSecret = ref(props.fedexClientSecret)
const localFedExScope = ref(props.fedexScope)
const localLogLevel = ref(props.logLevel)
const health = ref<NumistaHealthSummary | null>(null)
const healthLoading = ref(true)
const healthError = ref(false)
const ocreHealth = ref<OCREHealthSummary | null>(null)
const ocreHealthLoading = ref(true)
const ocreHealthError = ref(false)
const deepMetrics = ref<DeepIdentificationObservabilitySummary | null>(null)
const deepMetricsLoading = ref(true)
const deepMetricsError = ref(false)
const terminalJobCount = computed(() =>
  Object.values(deepMetrics.value?.jobsByTerminalStatus ?? {}).reduce((total, count) => total + count, 0),
)
const hasHealthEvents = computed(() => {
  if (!health.value) return false
  const counters = [
    health.value.broadRequestCount,
    health.value.detailRequestCount,
    health.value.freshCacheHitCount,
    health.value.coalescedRequestCount,
    health.value.providerLoadCount,
    health.value.providerFailureCount,
    health.value.cancelledRequestCount,
    health.value.enrichmentAttempted,
    health.value.enrichmentSucceeded,
    health.value.enrichmentFailed,
  ]
  return counters.some((count) => count > 0) ||
    Object.values(health.value.statusCounts).some((count) => count > 0) ||
    Boolean(health.value.lastOutcome || health.value.lastCheckedAt ||
      health.value.lastQuotaLimitedAt || (health.value.lastRetryAfterSeconds ?? 0) > 0)
})

const numistaSettingFields = [
  { name: 'NumistaSearchTTLHours', labelKey: 'admin.system.fields.searchCacheTtl', min: 1, max: 720, fallback: 24, model: localNumistaSearchTTLHours, hintKey: 'admin.system.hints.hours' },
  { name: 'NumistaDetailTTLHours', labelKey: 'admin.system.fields.detailCacheTtl', min: 1, max: 2160, fallback: 168, model: localNumistaDetailTTLHours, hintKey: 'admin.system.hints.hours' },
  { name: 'NumistaEnrichmentLimit', labelKey: 'admin.system.fields.enrichmentLimit', min: 1, max: 10, fallback: 5, model: localNumistaEnrichmentLimit, hintKey: 'admin.system.hints.candidatesPerLookup' },
  { name: 'NumistaSearchResultLimit', labelKey: 'admin.system.fields.searchResultLimit', min: 1, max: 50, fallback: 20, model: localNumistaSearchResultLimit, hintKey: 'admin.system.hints.broadCandidates' },
  { name: 'NumistaSearchTimeoutSeconds', labelKey: 'admin.system.fields.searchTimeout', min: 1, max: 10, fallback: 4, model: localNumistaSearchTimeoutSeconds, hintKey: 'admin.system.hints.seconds' },
  { name: 'NumistaDetailTimeoutSeconds', labelKey: 'admin.system.fields.detailTimeout', min: 1, max: 10, fallback: 3, model: localNumistaDetailTimeoutSeconds, hintKey: 'admin.system.hints.seconds' },
] as const

const deepIdentificationSettingFields = [
  { name: 'DeepIdentificationWorkerCount', labelKey: 'admin.system.fields.workerCount', min: 1, max: 32, fallback: 2, model: localDeepIdentificationWorkerCount, hintKey: 'admin.system.hints.jobWorkers' },
  { name: 'DeepIdentificationMaxActivePerUser', labelKey: 'admin.system.fields.activeJobsPerUser', min: 1, max: 10, fallback: 1, model: localDeepIdentificationMaxActivePerUser, hintKey: 'admin.system.hints.perUserJobs' },
  { name: 'DeepIdentificationQueueDepth', labelKey: 'admin.system.fields.queueDepth', min: 1, max: 1000, fallback: 32, model: localDeepIdentificationQueueDepth, hintKey: 'admin.system.hints.queuedJobs' },
  { name: 'DeepIdentificationHardTimeoutSeconds', labelKey: 'admin.system.fields.hardTimeout', min: 1, max: 900, fallback: 300, model: localDeepIdentificationHardTimeoutSeconds, hintKey: 'admin.system.hints.secondsPerJob' },
  { name: 'DeepIdentificationEventRetentionHours', labelKey: 'admin.system.fields.eventRetention', min: 1, max: 720, fallback: 24, model: localDeepIdentificationEventRetentionHours, hintKey: 'admin.system.hints.hours' },
  { name: 'DeepIdentificationResultRetentionDays', labelKey: 'admin.system.fields.resultRetention', min: 1, max: 3650, fallback: 90, model: localDeepIdentificationResultRetentionDays, hintKey: 'admin.system.hints.days' },
  { name: 'DeepIdentificationMaxProviders', labelKey: 'admin.system.fields.providerLimit', min: 1, max: 10, fallback: 4, model: localDeepIdentificationMaxProviders, hintKey: 'admin.system.hints.providersPerJob' },
  { name: 'DeepIdentificationNumistaCallBudget', labelKey: 'admin.system.fields.numistaCallBudget', min: 1, max: 20, fallback: 4, model: localDeepIdentificationNumistaCallBudget, hintKey: 'admin.system.hints.requestsPerJob' },
] as const

const coinCopilotSettingFields = [
  { name: 'CoinCopilotWorkerCount', labelKey: 'admin.system.fields.workerCount', min: 1, max: 4, fallback: 1, model: localCoinCopilotWorkerCount, hintKey: 'admin.system.hints.concurrentWorkers' },
  { name: 'CoinCopilotMaxActivePerUser', labelKey: 'admin.system.fields.activeRunsPerUser', min: 1, max: 3, fallback: 1, model: localCoinCopilotMaxActivePerUser, hintKey: 'admin.system.hints.runsPerOwner' },
  { name: 'CoinCopilotQueueDepth', labelKey: 'admin.system.fields.queueDepth', min: 1, max: 100, fallback: 16, model: localCoinCopilotQueueDepth, hintKey: 'admin.system.hints.queuedRuns' },
  { name: 'CoinCopilotMaxReasoningIterations', labelKey: 'admin.system.fields.reasoningIterations', min: 1, max: 20, fallback: 8, model: localCoinCopilotMaxReasoningIterations, hintKey: 'admin.system.hints.planningSteps' },
  { name: 'CoinCopilotMaxToolCalls', labelKey: 'admin.system.fields.toolCallLimit', min: 1, max: 40, fallback: 12, model: localCoinCopilotMaxToolCalls, hintKey: 'admin.system.hints.toolCalls' },
  { name: 'CoinCopilotHardTimeoutSeconds', labelKey: 'admin.system.fields.executionTimeout', min: 15, max: 600, fallback: 120, model: localCoinCopilotHardTimeoutSeconds, hintKey: 'admin.system.hints.secondsPerExecution' },
  { name: 'CoinCopilotMaxPersistedToolResultBytes', labelKey: 'admin.system.fields.persistedResultBytes', min: 4096, max: 131072, fallback: 32768, model: localCoinCopilotMaxPersistedToolResultBytes, hintKey: 'admin.system.hints.persistedBytes' },
  { name: 'CoinCopilotEventRetentionHours', labelKey: 'admin.system.fields.eventRetention', min: 1, max: 720, fallback: 168, model: localCoinCopilotEventRetentionHours, hintKey: 'admin.system.hints.hoursAfterTerminal' },
  { name: 'CoinCopilotCheckpointRetentionDays', labelKey: 'admin.system.fields.checkpointRetention', min: 1, max: 365, fallback: 30, model: localCoinCopilotCheckpointRetentionDays, hintKey: 'admin.system.hints.daysAfterTerminal' },
  { name: 'CoinCopilotResumeWindowHours', labelKey: 'admin.system.fields.resumeWindow', min: 1, max: 720, fallback: 168, model: localCoinCopilotResumeWindowHours, hintKey: 'admin.system.hints.resumeWindow' },
] as const

const statusRows: { key: NumistaLookupStatus; labelKey: MessageKey }[] = [
  { key: 'success', labelKey: 'admin.system.statuses.success' },
  { key: 'empty', labelKey: 'admin.system.statuses.empty' },
  { key: 'unconfigured', labelKey: 'admin.system.statuses.unconfigured' },
  { key: 'quota-limited', labelKey: 'admin.system.statuses.quotaLimited' },
  { key: 'timeout', labelKey: 'admin.system.statuses.timeout' },
  { key: 'unavailable', labelKey: 'admin.system.statuses.unavailable' },
]

function boundedValue(value: string, min: number, max: number, fallback: number) {
  const normalized = String(value ?? '').trim()
  if (!normalized) return String(fallback)
  const parsed = Number(normalized)
  if (!Number.isInteger(parsed)) return String(fallback)
  return String(Math.min(max, Math.max(min, parsed)))
}

function save() {
  for (const setting of numistaSettingFields) {
    setting.model.value = boundedValue(setting.model.value, setting.min, setting.max, setting.fallback)
  }
  for (const setting of deepIdentificationSettingFields) {
    setting.model.value = boundedValue(setting.model.value, setting.min, setting.max, setting.fallback)
  }
  for (const setting of coinCopilotSettingFields) {
    setting.model.value = boundedValue(setting.model.value, setting.min, setting.max, setting.fallback)
  }
  localOCRECallBudget.value = boundedValue(localOCRECallBudget.value, 1, 20, 3)
  emit('save', {
    numistaApiKey: localNumistaApiKey.value,
    numistaSearchTTLHours: localNumistaSearchTTLHours.value,
    numistaDetailTTLHours: localNumistaDetailTTLHours.value,
    numistaEnrichmentLimit: localNumistaEnrichmentLimit.value,
    numistaSearchResultLimit: localNumistaSearchResultLimit.value,
    numistaSearchTimeoutSeconds: localNumistaSearchTimeoutSeconds.value,
    numistaDetailTimeoutSeconds: localNumistaDetailTimeoutSeconds.value,
    deepIdentificationEnabled: localDeepIdentificationEnabled.value ? 'true' : 'false',
    deepIdentificationWorkerCount: localDeepIdentificationWorkerCount.value,
    deepIdentificationMaxActivePerUser: localDeepIdentificationMaxActivePerUser.value,
    deepIdentificationQueueDepth: localDeepIdentificationQueueDepth.value,
    deepIdentificationHardTimeoutSeconds: localDeepIdentificationHardTimeoutSeconds.value,
    deepIdentificationEventRetentionHours: localDeepIdentificationEventRetentionHours.value,
    deepIdentificationResultRetentionDays: localDeepIdentificationResultRetentionDays.value,
    deepIdentificationMaxProviders: localDeepIdentificationMaxProviders.value,
    deepIdentificationNumistaCallBudget: localDeepIdentificationNumistaCallBudget.value,
    deepIdentificationOCREEnabled: localOCREEnabled.value ? 'true' : 'false',
    deepIdentificationOCRECallBudget: localOCRECallBudget.value,
    coinCopilotEnabled: localCoinCopilotEnabled.value ? 'true' : 'false',
    coinCopilotAttributionEnabled: localCoinCopilotAttributionEnabled.value ? 'true' : 'false',
    coinCopilotWorkerCount: localCoinCopilotWorkerCount.value,
    coinCopilotMaxActivePerUser: localCoinCopilotMaxActivePerUser.value,
    coinCopilotQueueDepth: localCoinCopilotQueueDepth.value,
    coinCopilotMaxReasoningIterations: localCoinCopilotMaxReasoningIterations.value,
    coinCopilotMaxToolCalls: localCoinCopilotMaxToolCalls.value,
    coinCopilotHardTimeoutSeconds: localCoinCopilotHardTimeoutSeconds.value,
    coinCopilotMaxPersistedToolResultBytes: localCoinCopilotMaxPersistedToolResultBytes.value,
    coinCopilotEventRetentionHours: localCoinCopilotEventRetentionHours.value,
    coinCopilotCheckpointRetentionDays: localCoinCopilotCheckpointRetentionDays.value,
    coinCopilotResumeWindowHours: localCoinCopilotResumeWindowHours.value,
    logLevel: localLogLevel.value,
    pushoverAppToken: localPushoverAppToken.value,
    publicAppUrl: localPublicAppUrl.value,
    dealerSearchSources: localDealerSearchSources.value,
    auctionSearchSources: localAuctionSearchSources.value,
    uspsApiBaseUrl: localUSPSAPIBaseURL.value,
    uspsApiKey: localUSPSAPIKey.value,
    uspsApiKeyHeader: localUSPSAPIKeyHeader.value,
    upsApiBaseUrl: localUPSAPIBaseURL.value,
    upsTokenUrl: localUPSTokenURL.value,
    upsClientId: localUPSClientID.value,
    upsClientSecret: localUPSClientSecret.value,
    upsScope: localUPSScope.value,
    fedexApiBaseUrl: localFedExAPIBaseURL.value,
    fedexTokenUrl: localFedExTokenURL.value,
    fedexClientId: localFedExClientID.value,
    fedexClientSecret: localFedExClientSecret.value,
    fedexScope: localFedExScope.value,
  })
}

async function loadHealth() {
  healthLoading.value = true
  healthError.value = false
  try {
    const response = await getAdminNumistaHealth()
    health.value = response.data
  } catch {
    health.value = null
    healthError.value = true
  } finally {
    healthLoading.value = false
  }
}

async function loadOCREHealth() {
  ocreHealthLoading.value = true
  ocreHealthError.value = false
  try {
    const response = await getAdminOCREHealth()
    ocreHealth.value = response.data
  } catch {
    ocreHealth.value = null
    ocreHealthError.value = true
  } finally {
    ocreHealthLoading.value = false
  }
}

async function loadDeepMetrics() {
  deepMetricsLoading.value = true
  deepMetricsError.value = false
  try {
    const response = await getAdminDeepIdentificationObservability()
    deepMetrics.value = response.data
  } catch {
    deepMetrics.value = null
    deepMetricsError.value = true
  } finally {
    deepMetricsLoading.value = false
  }
}

function formatRate(value: number) {
  return `${(Math.min(1, Math.max(0, value)) * 100).toFixed(1)}%`
}

function formatDuration(milliseconds: number) {
  if (milliseconds <= 0) return t('admin.system.notObservedLower')
  if (milliseconds < 1000) return `${milliseconds} ms`
  return `${(milliseconds / 1000).toFixed(1)} s`
}

function formatStatusCounts(counts: Record<string, number>) {
  const entries = Object.entries(counts).filter(([, count]) => count > 0)
  return entries.length
    ? entries.map(([status, count]) => `${formatOCREOutcome(status)} ${count}`).join(', ')
    : t('admin.system.noAttempts')
}

function formatOCREOutcome(status?: string | null) {
  if (!status) return t('admin.system.noRecentOutcome')
  return deepProviderStatusLabel(status)
}

function formatStatus(status?: NumistaLookupStatus | null) {
  if (!status) return t('admin.system.noRecentOutcome')
  const row = statusRows.find((candidate) => candidate.key === status)
  return row ? t(row.labelKey) : status.charAt(0).toUpperCase() + status.slice(1)
}

function formatTimestamp(value?: string | null) {
  if (!value) return t('admin.system.notObserved')
  const timestamp = new Date(value)
  return Number.isNaN(timestamp.getTime()) ? t('admin.system.notObserved') : timestamp.toLocaleString(intlLocale())
}

function formatPercent(value: number) {
  const bounded = Math.min(1, Math.max(0, value))
  return t('admin.system.freshHitRate', { value: (bounded * 100).toFixed(1) })
}

function formatRetryAfter(value?: number | null) {
  return value && value > 0 ? t('admin.system.seconds', { count: value }) : t('admin.system.notObserved')
}

watch(() => props.numistaApiKey, (value) => { localNumistaApiKey.value = value })
watch(() => props.numistaSearchTTLHours, (value) => { localNumistaSearchTTLHours.value = value || '24' })
watch(() => props.numistaDetailTTLHours, (value) => { localNumistaDetailTTLHours.value = value || '168' })
watch(() => props.numistaEnrichmentLimit, (value) => { localNumistaEnrichmentLimit.value = value || '5' })
watch(() => props.numistaSearchResultLimit, (value) => { localNumistaSearchResultLimit.value = value || '20' })
watch(() => props.numistaSearchTimeoutSeconds, (value) => { localNumistaSearchTimeoutSeconds.value = value || '4' })
watch(() => props.numistaDetailTimeoutSeconds, (value) => { localNumistaDetailTimeoutSeconds.value = value || '3' })
watch(() => props.deepIdentificationEnabled, (value) => { localDeepIdentificationEnabled.value = (value || 'false') === 'true' })
watch(() => props.deepIdentificationWorkerCount, (value) => { localDeepIdentificationWorkerCount.value = value })
watch(() => props.deepIdentificationMaxActivePerUser, (value) => { localDeepIdentificationMaxActivePerUser.value = value })
watch(() => props.deepIdentificationQueueDepth, (value) => { localDeepIdentificationQueueDepth.value = value })
watch(() => props.deepIdentificationHardTimeoutSeconds, (value) => { localDeepIdentificationHardTimeoutSeconds.value = value })
watch(() => props.deepIdentificationEventRetentionHours, (value) => { localDeepIdentificationEventRetentionHours.value = value })
watch(() => props.deepIdentificationResultRetentionDays, (value) => { localDeepIdentificationResultRetentionDays.value = value })
watch(() => props.deepIdentificationMaxProviders, (value) => { localDeepIdentificationMaxProviders.value = value })
watch(() => props.deepIdentificationNumistaCallBudget, (value) => { localDeepIdentificationNumistaCallBudget.value = value })
watch(() => props.deepIdentificationOCREEnabled, (value) => { localOCREEnabled.value = (value || 'false') === 'true' })
watch(() => props.deepIdentificationOCRECallBudget, (value) => { localOCRECallBudget.value = value || '3' })
watch(() => props.coinCopilotEnabled, (value) => { localCoinCopilotEnabled.value = (value || 'false') === 'true' })
watch(() => props.coinCopilotAttributionEnabled, (value) => { localCoinCopilotAttributionEnabled.value = (value || 'false') === 'true' })
watch(() => props.coinCopilotWorkerCount, (value) => { localCoinCopilotWorkerCount.value = value })
watch(() => props.coinCopilotMaxActivePerUser, (value) => { localCoinCopilotMaxActivePerUser.value = value })
watch(() => props.coinCopilotQueueDepth, (value) => { localCoinCopilotQueueDepth.value = value })
watch(() => props.coinCopilotMaxReasoningIterations, (value) => { localCoinCopilotMaxReasoningIterations.value = value })
watch(() => props.coinCopilotMaxToolCalls, (value) => { localCoinCopilotMaxToolCalls.value = value })
watch(() => props.coinCopilotHardTimeoutSeconds, (value) => { localCoinCopilotHardTimeoutSeconds.value = value })
watch(() => props.coinCopilotMaxPersistedToolResultBytes, (value) => { localCoinCopilotMaxPersistedToolResultBytes.value = value })
watch(() => props.coinCopilotEventRetentionHours, (value) => { localCoinCopilotEventRetentionHours.value = value })
watch(() => props.coinCopilotCheckpointRetentionDays, (value) => { localCoinCopilotCheckpointRetentionDays.value = value })
watch(() => props.coinCopilotResumeWindowHours, (value) => { localCoinCopilotResumeWindowHours.value = value })
watch(() => props.pushoverAppToken, (value) => { localPushoverAppToken.value = value })
watch(() => props.publicAppUrl, (value) => { localPublicAppUrl.value = value })
watch(() => props.dealerSearchSources, (value) => { localDealerSearchSources.value = value })
watch(() => props.auctionSearchSources, (value) => { localAuctionSearchSources.value = value })
watch(() => props.uspsApiBaseUrl, (value) => { localUSPSAPIBaseURL.value = value })
watch(() => props.uspsApiKey, (value) => { localUSPSAPIKey.value = value })
watch(() => props.uspsApiKeyHeader, (value) => { localUSPSAPIKeyHeader.value = value })
watch(() => props.upsApiBaseUrl, (value) => { localUPSAPIBaseURL.value = value })
watch(() => props.upsTokenUrl, (value) => { localUPSTokenURL.value = value })
watch(() => props.upsClientId, (value) => { localUPSClientID.value = value })
watch(() => props.upsClientSecret, (value) => { localUPSClientSecret.value = value })
watch(() => props.upsScope, (value) => { localUPSScope.value = value })
watch(() => props.fedexApiBaseUrl, (value) => { localFedExAPIBaseURL.value = value })
watch(() => props.fedexTokenUrl, (value) => { localFedExTokenURL.value = value })
watch(() => props.fedexClientId, (value) => { localFedExClientID.value = value })
watch(() => props.fedexClientSecret, (value) => { localFedExClientSecret.value = value })
watch(() => props.fedexScope, (value) => { localFedExScope.value = value })
watch(() => props.logLevel, (value) => { localLogLevel.value = value })

onMounted(() => {
  loadHealth()
  loadOCREHealth()
  loadDeepMetrics()
})
</script>
