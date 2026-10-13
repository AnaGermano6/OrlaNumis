<template>
  <div class="flex flex-col gap-2 border-b border-border-subtle py-3" data-testid="settings-currency">
    <div class="flex flex-col items-stretch justify-between gap-4 md:flex-row md:items-center">
      <div class="flex flex-col gap-[0.15rem]">
        <label class="text-base font-medium" for="settings-currency-select">{{ t('settings.currency.label') }}</label>
        <span class="text-sm text-text-muted">{{ t('settings.currency.hint') }}</span>
      </div>
      <form v-if="auth.isAdmin" class="flex shrink-0 flex-wrap items-center gap-2 self-start md:self-auto" @submit.prevent="apply">
        <select id="settings-currency-select" v-model="selected" class="form-select w-auto" data-testid="collection-currency">
          <option v-for="option in options" :key="option.code" :value="option.code">{{ option.name }}</option>
        </select>
        <button
          type="submit"
          class="btn btn-primary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :disabled="saving || selected === collectionCurrency"
        >
          {{ saving ? t('common.saving') : t('settings.currency.apply') }}
        </button>
      </form>
      <span v-else id="settings-currency-select" class="self-start text-body text-text-primary md:self-auto">{{ currencyName(collectionCurrency) }}</span>
    </div>
    <p v-if="!auth.isAdmin" class="m-0 text-chip text-text-muted">{{ t('settings.currency.adminOnly') }}</p>
    <p v-else-if="selected !== collectionCurrency" class="m-0 text-chip text-text-muted">{{ t('settings.currency.conversionNote') }}</p>
    <p v-if="message" class="m-0 text-body" :class="failed ? 'text-byzantine' : 'text-gold'" role="status">{{ message }}</p>
  </div>
</template>

<script setup lang="ts">
// The currency every page shows amounts in, which is also the one they are
// kept in. An administrator can switch it; the server converts every stored
// price and value with the ECB rate in the same step.
import { computed, ref, watch } from 'vue'
import { changeCollectionCurrency } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { intlLocale, t } from '@/i18n'
import { COLLECTION_CURRENCIES, collectionCurrency, currencyName, setCollectionCurrency, type CollectionCurrency } from '@/utils/collectionCurrency'

const auth = useAuthStore()

// The usual ones first, then the others by name in the app's language.
const FIRST = ['EUR', 'USD', 'GBP', 'CHF']
const options = computed(() => COLLECTION_CURRENCIES
  .map((code) => ({ code, name: currencyName(code) }))
  .sort((a, b) => {
    const rank = (code: string) => (FIRST.includes(code) ? FIRST.indexOf(code) : FIRST.length)
    return rank(a.code) - rank(b.code) || a.name.localeCompare(b.name, intlLocale())
  }))
const selected = ref<CollectionCurrency>(collectionCurrency.value)
const saving = ref(false)
const message = ref('')
const failed = ref(false)

// The currency arrives with the coin options, possibly after this mounts.
watch(collectionCurrency, (code) => {
  if (!saving.value) selected.value = code
})

async function apply() {
  if (selected.value === collectionCurrency.value) return
  if (!window.confirm(t('settings.currency.confirm', { currency: currencyName(selected.value) }))) return
  saving.value = true
  message.value = ''
  failed.value = false
  try {
    const { data } = await changeCollectionCurrency(selected.value)
    setCollectionCurrency(data.currency)
    const factor = data.factor ? new Intl.NumberFormat(intlLocale(), { maximumFractionDigits: 4 }).format(data.factor) : ''
    const date = data.rateDate ? new Date(`${data.rateDate}T00:00:00Z`).toLocaleDateString(intlLocale(), { timeZone: 'UTC' }) : ''
    message.value = factor
      ? t('settings.currency.done', { count: data.coinsConverted, from: data.previous, to: data.currency, factor, date })
      : t('settings.currency.unchanged')
  } catch {
    failed.value = true
    message.value = t('settings.currency.failed')
    selected.value = collectionCurrency.value
  } finally {
    saving.value = false
  }
}
</script>
