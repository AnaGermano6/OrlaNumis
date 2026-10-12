<template>
  <li class="flex flex-col gap-2" :data-testid="`ucoin-not-found-${coin.id}`">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <RouterLink :to="`/coin/${coin.id}`" class="text-gold underline">{{ coin.name }}</RouterLink>
      <button v-if="!result || result === 'failed'" type="button" class="btn btn-secondary btn-sm" :aria-expanded="open" @click="toggle">
        {{ open ? t('ucoinImport.pick.close') : t('ucoinImport.pick.open') }}
      </button>
    </div>

    <p v-if="result === 'completed' || result === 'pending'" class="m-0 text-chip text-text-secondary" role="status">
      {{ t(result === 'completed' ? 'ucoinImport.pick.completed' : 'ucoinImport.pick.pending') }}
    </p>

    <div v-else-if="open" class="flex flex-col gap-3 rounded-sm border border-border-subtle p-3">
      <p v-if="loading" class="m-0 text-chip text-text-muted" role="status">{{ t('ucoinImport.pick.loading') }}</p>
      <p v-else-if="searchFailed" class="m-0 text-chip text-text-secondary">{{ t('ucoinImport.pick.unavailable') }}</p>
      <template v-else>
        <p class="m-0 text-chip text-text-secondary">{{ candidates.length ? t('ucoinImport.pick.candidates') : t('ucoinImport.pick.none') }}</p>
        <ul v-if="candidates.length" class="m-0 flex list-none flex-col gap-1 p-0">
          <li v-for="item in candidates" :key="item.id" class="flex flex-wrap items-center justify-between gap-2 rounded-sm px-2 py-1 hover:bg-card-hover">
            <span class="min-w-0 text-body text-text-primary">
              {{ item.title }}
              <span class="text-chip text-text-muted">· {{ years(item) }} ·</span>
              <a :href="`https://pt.numista.com/${item.id}`" target="_blank" rel="noopener noreferrer" class="text-chip text-gold underline">{{ t('ucoinImport.pick.view') }}</a>
            </span>
            <button type="button" class="btn btn-secondary btn-sm" :disabled="saving" @click="choose(item.id)">{{ t('ucoinImport.pick.choose') }}</button>
          </li>
        </ul>
      </template>

      <form class="flex flex-wrap items-end gap-2" @submit.prevent="chooseTyped">
        <label class="flex min-w-[16rem] flex-1 flex-col gap-1">
          <span class="text-chip text-text-secondary">{{ t('ucoinImport.pick.manual') }}</span>
          <input v-model="typed" type="text" inputmode="url" class="form-input" placeholder="https://pt.numista.com/380096" />
        </label>
        <button type="submit" class="btn btn-secondary btn-sm" :disabled="saving || !typed.trim()">{{ t('ucoinImport.pick.link') }}</button>
      </form>
      <p v-if="invalid" class="m-0 text-chip text-byzantine" role="alert">{{ t('ucoinImport.pick.invalid') }}</p>
      <p v-if="result === 'failed'" class="m-0 text-chip text-byzantine" role="alert">{{ t('ucoinImport.pick.failed') }}</p>
    </div>
  </li>
</template>

<script setup lang="ts">
// A coin the uCoin import did not find on Numista: its Numista type can be
// chosen among the types of the same country, value and year (the searches
// the import made, at no cost), or given by pasting the type's link.
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import type { NumistaCatalogSearchItem } from '@/api/client'
import type { Coin } from '@/types'
import type { UcoinChoiceResult } from '@/composables/useUcoinImport'
import { t } from '@/i18n'
import { formatNumistaYear } from '@/utils/numistaCatalog'
import { parseNumistaTypeId } from '@/utils/ucoinNumista'

const props = defineProps<{
  coin: Coin
  candidatesFor: (coin: Coin) => Promise<NumistaCatalogSearchItem[]>
  chooseType: (coin: Coin, typeId: number) => Promise<UcoinChoiceResult>
}>()

const open = ref(false)
const loading = ref(false)
const loaded = ref(false)
const searchFailed = ref(false)
const candidates = ref<NumistaCatalogSearchItem[]>([])
const typed = ref('')
const invalid = ref(false)
const saving = ref(false)
const result = ref<UcoinChoiceResult | null>(null)

function years(item: NumistaCatalogSearchItem): string {
  if (item.minYear === undefined) return ''
  return item.maxYear === undefined || item.maxYear === item.minYear
    ? formatNumistaYear(item.minYear)
    : `${formatNumistaYear(item.minYear)}–${formatNumistaYear(item.maxYear)}`
}

async function toggle() {
  open.value = !open.value
  if (!open.value || loaded.value) return
  loading.value = true
  try {
    candidates.value = await props.candidatesFor(props.coin)
    searchFailed.value = false
    loaded.value = true
  } catch {
    searchFailed.value = true
  } finally {
    loading.value = false
  }
}

async function choose(typeId: number) {
  saving.value = true
  invalid.value = false
  try {
    result.value = await props.chooseType(props.coin, typeId)
  } finally {
    saving.value = false
  }
}

async function chooseTyped() {
  const typeId = parseNumistaTypeId(typed.value)
  if (typeId === undefined) {
    invalid.value = true
    return
  }
  await choose(typeId)
}
</script>
