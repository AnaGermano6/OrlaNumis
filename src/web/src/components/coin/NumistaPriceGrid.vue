<template>
  <ul class="m-0 grid list-none grid-cols-2 gap-2 p-0 sm:grid-cols-3 lg:grid-cols-4" :aria-label="t('coinDetail.marketValue.pricesLabel')">
    <li v-for="price in prices" :key="price.grade">
      <component
        :is="selectable ? 'button' : 'div'"
        :type="selectable ? 'button' : undefined"
        :aria-pressed="selectable ? price.grade === highlightGrade : undefined"
        class="flex h-full w-full flex-col gap-[0.15rem] rounded-sm border p-3 text-left"
        :class="[
          price.grade === highlightGrade ? 'border-gold bg-gold-glow' : 'border-border-subtle bg-card',
          selectable ? 'cursor-pointer transition-colors hover:border-gold focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2' : '',
        ]"
        @click="selectable && $emit('select', price.grade === highlightGrade ? null : price.grade)"
      >
        <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ numistaGradeShort(price.grade) }}</span>
        <span class="text-chip text-text-secondary">{{ numistaGradeName(price.grade) }}</span>
        <span class="mt-1 text-lg font-medium text-text-primary">{{ formatCurrency(price.price, currency) }}</span>
        <span v-if="price.grade === highlightGrade" class="chip-sm mt-1 w-fit">{{ t('coinDetail.marketValue.yourGrade') }}</span>
      </component>
    </li>
  </ul>
</template>

<script setup lang="ts">
// Numista's estimated price per grade. With `selectable`, each card is a
// toggle button used to pick the grade of the user's coin.
import type { NumistaGradePrice } from '@/api/client'
import { t } from '@/i18n'
import { formatCurrency } from '@/utils/format'
import { numistaGradeName, numistaGradeShort } from '@/utils/numistaGrades'

withDefaults(defineProps<{
  prices: NumistaGradePrice[]
  currency: string
  highlightGrade?: string | null
  selectable?: boolean
}>(), { highlightGrade: null, selectable: false })

defineEmits<{
  select: [grade: string | null]
}>()
</script>
