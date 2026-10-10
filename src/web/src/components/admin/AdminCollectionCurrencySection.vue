<template>
  <section class="card p-6" aria-labelledby="collection-currency-title">
    <p class="section-label">{{ t('admin.currency.eyebrow') }}</p>
    <h2 id="collection-currency-title" class="m-0 text-xl font-medium text-heading">{{ t('admin.currency.title') }}</h2>
    <p class="mb-4 mt-2 text-base text-text-secondary">{{ t('admin.currency.intro') }}</p>

    <form class="flex flex-wrap items-end gap-3" @submit.prevent="apply">
      <label class="flex flex-col gap-1">
        <span class="form-label">{{ t('admin.currency.label') }}</span>
        <select v-model="selected" class="form-select w-auto" data-testid="collection-currency">
          <option v-for="code in COLLECTION_CURRENCIES" :key="code" :value="code">{{ t(`admin.currency.options.${code}`) }}</option>
        </select>
      </label>
      <button
        type="submit"
        class="btn btn-primary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        :disabled="saving || selected === collectionCurrency"
      >
        {{ saving ? t('common.saving') : t('admin.currency.apply') }}
      </button>
    </form>
    <p v-if="selected !== collectionCurrency" class="mb-0 mt-3 text-chip text-text-muted">{{ t('admin.currency.conversionNote') }}</p>
    <p v-if="message" class="mb-0 mt-3 text-body" :class="failed ? 'text-byzantine' : 'text-gold'" role="status">{{ message }}</p>
  </section>
</template>

<script setup lang="ts">
// Switches the currency coin prices and values are kept in. The server
// converts every stored amount with the ECB rate in the same step.
import { ref, watch } from 'vue'
import { changeCollectionCurrency } from '@/api/client'
import { intlLocale, t } from '@/i18n'
import { COLLECTION_CURRENCIES, collectionCurrency, setCollectionCurrency, type CollectionCurrency } from '@/utils/collectionCurrency'

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
  if (!window.confirm(t('admin.currency.confirm', { currency: t(`admin.currency.options.${selected.value}`) }))) return
  saving.value = true
  message.value = ''
  failed.value = false
  try {
    const { data } = await changeCollectionCurrency(selected.value)
    setCollectionCurrency(data.currency)
    const rate = data.exchangeRate
      ? new Intl.NumberFormat(intlLocale(), { maximumFractionDigits: 4 }).format(data.exchangeRate.usdPerEur)
      : ''
    message.value = t('admin.currency.done', { count: data.coinsConverted, rate })
  } catch {
    failed.value = true
    message.value = t('admin.currency.failed')
    selected.value = collectionCurrency.value
  } finally {
    saving.value = false
  }
}
</script>
