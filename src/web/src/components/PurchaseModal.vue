<template>
  <div class="fixed inset-0 z-[1000] flex items-center justify-center bg-overlay p-4" @click.self="$emit('close')">
    <div class="card w-full max-w-[420px] !p-8">
      <h3 class="mb-1 text-lg font-medium text-heading">{{ t('featured.moveToCollection') }}</h3>
      <p class="mb-5 truncate text-base text-gold">{{ coin.name }}</p>

      <div class="form-group">
        <label class="form-label">{{ t('coinFields.purchasePrice') }}</label>
        <div class="relative flex items-center">
          <span class="pointer-events-none absolute left-3 text-base text-text-secondary">{{ collectionCurrencySymbol() }}</span>
          <input
            ref="priceInput"
            v-model="priceStr"
            type="number"
            step="0.01"
            min="0"
            class="form-input pl-6 [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
            placeholder="0.00"
          />
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">{{ t('coinFields.purchaseDate') }}</label>
        <input
          v-model="purchaseDate"
          type="date"
          class="form-input"
        />
      </div>

      <div class="form-group">
        <label class="form-label">{{ t('purchase.purchasedFrom') }}</label>
        <input
          v-model="purchaseLocation"
          type="text"
          class="form-input"
          :placeholder="t('purchase.purchasedFromPlaceholder')"
        />
      </div>

      <p class="mb-3 text-chip text-text-secondary">{{ t('purchase.allOptional') }}</p>

      <div v-if="error" class="mb-3 text-chip text-loss">{{ error }}</div>

      <div class="mt-6 flex justify-end gap-3">
        <button class="btn btn-secondary" @click="$emit('close')">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="submitting" @click="handleSubmit">
          {{ submitting ? t('common.saving') : t('addCoin.addToCollection') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { collectionCurrencySymbol } from '@/utils/collectionCurrency'
import { ref, onMounted } from 'vue'
import type { Coin } from '@/types'
import { t } from '@/i18n'

const props = defineProps<{
  coin: Coin
}>()

const emit = defineEmits<{
  close: []
  confirm: [data: {
    purchasePrice?: number
    purchaseDate?: string
    purchaseLocation?: string
  }]
}>()

const priceStr = ref('')
const purchaseDate = ref(new Date().toISOString().slice(0, 10))
const purchaseLocation = ref('')
const error = ref('')
const submitting = ref(false)
const priceInput = ref<HTMLInputElement>()

onMounted(() => {
  if (props.coin.purchasePrice) {
    priceStr.value = String(props.coin.purchasePrice)
  }
  if (props.coin.purchaseLocation) {
    purchaseLocation.value = props.coin.purchaseLocation
  }
  priceInput.value?.focus()
})

function handleSubmit() {
  error.value = ''

  const data: {
    purchasePrice?: number
    purchaseDate?: string
    purchaseLocation?: string
  } = {}

  if (priceStr.value) {
    const price = parseFloat(priceStr.value)
    if (isNaN(price) || price < 0) {
      error.value = t('common.validPrice')
      return
    }
    data.purchasePrice = price
  }

  if (purchaseDate.value) {
    data.purchaseDate = purchaseDate.value
  }

  if (purchaseLocation.value.trim()) {
    data.purchaseLocation = purchaseLocation.value.trim()
  }

  submitting.value = true
  emit('confirm', data)
}
</script>
