<template>
  <div class="flex items-stretch gap-1">
    <button
      type="button"
      class="btn btn-secondary btn-sm shrink-0 justify-center px-3"
      :aria-label="t('coinForm.quantityLess')"
      :disabled="current <= QUANTITY_MIN"
      data-testid="quantity-less"
      @click="emit('update:modelValue', current - 1)"
    >
      <Minus :size="16" />
    </button>
    <input
      :id="id"
      class="form-input !w-20 text-center [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
      type="number"
      inputmode="numeric"
      :min="QUANTITY_MIN"
      :max="QUANTITY_MAX"
      step="1"
      :value="current"
      data-testid="quantity-input"
      @input="onInput"
      @change="onChange"
    />
    <button
      type="button"
      class="btn btn-secondary btn-sm shrink-0 justify-center px-3"
      :aria-label="t('coinForm.quantityMore')"
      :disabled="current >= QUANTITY_MAX"
      data-testid="quantity-more"
      @click="emit('update:modelValue', current + 1)"
    >
      <Plus :size="16" />
    </button>
  </div>
</template>

<script setup lang="ts">
// How many of a coin or note the collector has: a whole number from 1, typed
// or stepped with − and +.
import { computed } from 'vue'
import { Minus, Plus } from 'lucide-vue-next'
import { t } from '@/i18n'
import { QUANTITY_MAX, QUANTITY_MIN } from '@/utils/coinAttributes'

const props = defineProps<{
  modelValue: number | null | undefined
  id?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number]
}>()

const current = computed(() => props.modelValue && props.modelValue >= QUANTITY_MIN ? props.modelValue : QUANTITY_MIN)

function valid(raw: string): number | null {
  const value = Number(raw)
  return Number.isInteger(value) && value >= QUANTITY_MIN && value <= QUANTITY_MAX ? value : null
}

// While typing, only whole numbers in range are taken; an empty box can be
// typed over without snapping back.
function onInput(event: Event) {
  const value = valid((event.target as HTMLInputElement).value)
  if (value !== null) emit('update:modelValue', value)
}

// On leaving the box, anything else goes back to the last good number.
function onChange(event: Event) {
  const input = event.target as HTMLInputElement
  if (valid(input.value) === null) input.value = String(current.value)
}
</script>
