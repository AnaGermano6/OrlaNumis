<template>
  <div class="relative flex shrink-0 justify-end" @keydown.esc="menuOpen = false">
    <button
      type="button"
      class="btn btn-sm btn-ghost justify-center px-3"
      :class="modelValue ? 'border-gold bg-gold-glow text-gold' : ''"
      :aria-label="t('auctions.statusFilters')"
      aria-haspopup="menu"
      :aria-expanded="menuOpen"
      :title="t('auctions.statusFiltersTitle')"
      @click="menuOpen = !menuOpen"
    >
      <Menu :size="18" />
    </button>

    <div v-if="menuOpen" class="absolute top-[calc(100%+0.35rem)] right-0 z-10 flex min-w-36 flex-col items-stretch gap-[0.35rem] rounded-sm border border-border-subtle bg-card p-2 shadow-card" role="menu" :aria-label="t('auctions.statusFilters')">
      <button
        v-for="s in statuses"
        :key="s.value"
        type="button"
        class="chip w-full justify-between gap-2"
        :class="{ active: modelValue === s.value }"
        role="menuitemradio"
        :aria-checked="modelValue === s.value"
        @click="selectStatus(s.value)"
      >
        {{ s.label }}
        <span v-if="counts[s.value]" class="rounded-full bg-surface px-1.5 py-[0.05rem] text-label font-semibold">{{ counts[s.value] }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Menu } from 'lucide-vue-next'
import { auctionStatusLabel } from '@/utils/format'
import { t } from '@/i18n'

defineProps<{
  modelValue: string
  counts: Record<string, number>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const menuOpen = ref(false)

const statuses = computed(() => [
  { value: '', label: t('common.all') },
  ...['watching', 'bidding', 'won', 'lost', 'passed'].map((value) => ({ value, label: auctionStatusLabel(value) })),
])

function selectStatus(value: string) {
  emit('update:modelValue', value)
  menuOpen.value = false
}
</script>
