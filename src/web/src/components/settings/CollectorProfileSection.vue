<template>
  <section class="card bg-card text-text-primary">
    <div class="mb-4 border-b border-border-subtle pb-3">
      <h2 class="text-lg text-heading">{{ t('settings.collector.title') }}</h2>
      <p class="mt-1 text-sm text-text-muted">
        {{ t('settings.collector.intro') }}
      </p>
    </div>

    <p v-if="loading" class="text-sm text-text-muted">{{ t('settings.collector.loading') }}</p>

    <form v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2" @submit.prevent="save">
      <div>
        <label for="collector-budget-min" class="form-label">{{ t('settings.collector.budgetMin') }}</label>
        <input id="collector-budget-min" v-model="form.budgetMin" data-test="budget-min" class="form-input min-h-11" type="number" min="0" max="100000000" step="any" inputmode="decimal" />
      </div>
      <div>
        <label for="collector-budget-max" class="form-label">{{ t('settings.collector.budgetMax') }}</label>
        <input id="collector-budget-max" v-model="form.budgetMax" data-test="budget-max" class="form-input min-h-11" type="number" min="0" max="100000000" step="any" inputmode="decimal" />
      </div>
      <div class="sm:col-span-2">
        <label for="collector-currency" class="form-label">{{ t('settings.collector.currency') }}</label>
        <input id="collector-currency" v-model="form.currency" data-test="currency" class="form-input min-h-11 uppercase" maxlength="3" placeholder="USD" autocomplete="off" />
      </div>
      <div>
        <label for="collector-periods" class="form-label">{{ t('settings.collector.periods') }}</label>
        <textarea id="collector-periods" v-model="form.preferredPeriods" data-test="periods" class="form-input min-h-11 resize-y" rows="2" :placeholder="t('settings.collector.periodsPlaceholder')"></textarea>
      </div>
      <div>
        <label for="collector-categories" class="form-label">{{ t('settings.collector.categories') }}</label>
        <textarea id="collector-categories" v-model="form.preferredCategories" class="form-input min-h-11 resize-y" rows="2" :placeholder="t('settings.collector.categoriesPlaceholder')"></textarea>
      </div>
      <div>
        <label for="collector-exclusions" class="form-label">{{ t('settings.collector.exclusions') }}</label>
        <textarea id="collector-exclusions" v-model="form.excludedCategories" class="form-input min-h-11 resize-y" rows="2" :placeholder="t('settings.collector.exclusionsPlaceholder')"></textarea>
      </div>
      <div>
        <label for="collector-dealers" class="form-label">{{ t('settings.collector.dealers') }}</label>
        <textarea id="collector-dealers" v-model="form.preferredDealers" class="form-input min-h-11 resize-y" rows="2" :placeholder="t('settings.collector.dealersPlaceholder')"></textarea>
      </div>
      <div class="sm:col-span-2">
        <label for="collector-goals" class="form-label">{{ t('settings.collector.goals') }}</label>
        <textarea id="collector-goals" v-model="form.collectingGoals" class="form-input min-h-11 resize-y" rows="3" :placeholder="t('settings.collector.goalsPlaceholder')"></textarea>
      </div>

      <p v-if="message" :role="hasError ? 'alert' : 'status'" class="sm:col-span-2 text-sm" :class="hasError ? 'text-loss' : 'text-gain'">
        {{ message }}
      </p>

      <div class="flex flex-wrap gap-2 sm:col-span-2">
        <button type="submit" class="btn btn-primary min-h-11 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" :disabled="saving">
          {{ saving ? t('common.saving') : t('settings.account.saveProfile') }}
        </button>
        <button type="button" data-test="clear" class="btn btn-secondary min-h-11 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" :disabled="saving" @click="clearProfile">
          {{ t('settings.collector.clear') }}
        </button>
      </div>
    </form>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { getCollectorProfile, replaceCollectorProfile } from '@/api/endpoints/collectorProfile'
import type { CollectorProfile, CollectorProfileInput } from '@/types/collectorProfile'
import { t } from '@/i18n'

interface ProfileForm {
  budgetMin: string
  budgetMax: string
  currency: string
  preferredPeriods: string
  preferredCategories: string
  excludedCategories: string
  preferredDealers: string
  collectingGoals: string
}

const form = reactive<ProfileForm>(emptyForm())
const loading = ref(true)
const saving = ref(false)
const message = ref('')
const hasError = ref(false)

function emptyForm(): ProfileForm {
  return {
    budgetMin: '', budgetMax: '', currency: '', preferredPeriods: '',
    preferredCategories: '', excludedCategories: '', preferredDealers: '',
    collectingGoals: '',
  }
}

function applyProfile(profile: CollectorProfile) {
  form.budgetMin = profile.budgetMin == null ? '' : String(profile.budgetMin)
  form.budgetMax = profile.budgetMax == null ? '' : String(profile.budgetMax)
  form.currency = profile.currency ?? ''
  form.preferredPeriods = profile.preferredPeriods.join(', ')
  form.preferredCategories = profile.preferredCategories.join(', ')
  form.excludedCategories = profile.excludedCategories.join(', ')
  form.preferredDealers = profile.preferredDealers.join('\n')
  form.collectingGoals = profile.collectingGoals.join('\n')
}

function parseList(value: string): string[] {
  return value.split(/[,\n]/).map(item => item.trim()).filter(Boolean)
}

function nullableNumber(value: string): number | null {
  return value.trim() === '' ? null : Number(value)
}

function payload(): CollectorProfileInput {
  const currency = form.currency.trim()
  return {
    budgetMin: nullableNumber(form.budgetMin),
    budgetMax: nullableNumber(form.budgetMax),
    currency: currency === '' ? null : currency,
    preferredPeriods: parseList(form.preferredPeriods),
    preferredCategories: parseList(form.preferredCategories),
    excludedCategories: parseList(form.excludedCategories),
    preferredDealers: parseList(form.preferredDealers),
    collectingGoals: parseList(form.collectingGoals),
  }
}

function errorField(error: unknown): string | null {
  if (typeof error !== 'object' || error === null || !('response' in error)) return null
  const response = (error as { response?: { data?: unknown } }).response
  if (typeof response?.data !== 'object' || response.data === null || !('field' in response.data)) return null
  const field = (response.data as { field?: unknown }).field
  return typeof field === 'string' ? field : null
}

async function load() {
  try {
    const response = await getCollectorProfile()
    applyProfile(response.data)
  } catch {
    hasError.value = true
    message.value = t('settings.collector.loadFailed')
  } finally {
    loading.value = false
  }
}

async function replace(input: CollectorProfileInput) {
  saving.value = true
  message.value = ''
  try {
    const response = await replaceCollectorProfile(input)
    applyProfile(response.data)
    hasError.value = false
    message.value = t('settings.collector.saved')
  } catch (error: unknown) {
    const field = errorField(error)
    hasError.value = true
    message.value = field ? t('settings.collector.correctField', { field }) : t('settings.collector.saveFailed')
  } finally {
    saving.value = false
  }
}

async function save() {
  await replace(payload())
}

async function clearProfile() {
  await replace({
    budgetMin: null, budgetMax: null, currency: null,
    preferredPeriods: [], preferredCategories: [], excludedCategories: [],
    preferredDealers: [], collectingGoals: [],
  })
}

onMounted(load)
</script>
