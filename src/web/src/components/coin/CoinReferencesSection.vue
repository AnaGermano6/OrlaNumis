<template>
  <section id="catalog-references" class="mb-6 scroll-mt-24">
    <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
      <h3 class="m-0 text-base font-medium text-text-primary">{{ t('coinDetail.references.title') }}</h3>
      <div class="inline-flex flex-wrap items-center gap-[0.35rem]">
        <button type="button" class="btn btn-secondary btn-sm" :disabled="saving" @click="startCreate">
          {{ t('coinDetail.references.add') }}
        </button>
        <button
          ref="numistaDisclosure"
          type="button"
          class="btn btn-secondary btn-sm"
          :aria-expanded="numistaExpanded"
          :aria-controls="numistaPanelId"
          @click="toggleNumista"
          @keydown.enter.prevent="toggleNumista"
          @keydown.space.prevent="toggleNumista"
        >
          {{ t('coinDetail.references.searchNumista') }}
        </button>
      </div>
    </div>

    <div
      v-if="numistaExpanded"
      :id="numistaPanelId"
      class="mb-3 min-w-0 overflow-hidden rounded-sm border border-border-subtle bg-card p-3"
    >
      <CoinNumistaPanel
        :coin-id="coinId"
        :coin-name="coinName"
        :coin-ruler="coinRuler"
        :coin-denomination="coinDenomination"
        :coin-mint="coinMint"
        :coin-date-range="coinDateRange"
        :coin-material="coinMaterial"
        :coin-obverse-inscription="coinObverseInscription"
        :coin-reverse-inscription="coinReverseInscription"
        @reference-added="handleNumistaReferenceAdded"
      />
    </div>

    <div v-if="!rows.length && !editing" class="section-content-card text-body text-text-secondary">
      {{ t('coinDetail.references.empty') }}
    </div>

    <div v-else class="flex flex-col gap-2">
      <article
        v-for="reference in rows"
        :key="reference.id"
        class="flex flex-wrap items-center justify-between gap-3 rounded-sm border border-border-subtle bg-card p-3"
      >
        <template v-if="editing?.mode === 'edit' && editing.id === reference.id">
          <form class="flex w-full flex-col gap-2" @submit.prevent="saveEdit">
            <div class="grid gap-2 md:grid-cols-3">
              <select v-model="draft.catalog" class="form-input">
                <option value="" disabled>{{ t('coinDetail.references.selectCatalog') }}</option>
                <option v-for="opt in editingCatalogOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
              <input v-model.trim="draft.volume" class="form-input" :placeholder="t('coinDetail.references.volume')" />
              <input v-model.trim="draft.number" class="form-input" :placeholder="t('coinDetail.references.number')" />
            </div>
            <input v-model.trim="draft.uri" class="form-input" :placeholder="t('coinDetail.references.url')" />
            <div class="inline-flex flex-wrap items-center gap-[0.35rem]">
              <button type="submit" class="btn btn-primary btn-sm" :disabled="saving">{{ t('common.save') }}</button>
              <button type="button" class="btn btn-secondary btn-sm" :disabled="saving" @click="cancelEdit">{{ t('common.cancel') }}</button>
            </div>
          </form>
        </template>

        <template v-else>
          <div class="inline-flex flex-wrap items-center gap-2">
            <span class="chip-sm border-border-accent text-gold">{{ reference.catalog }}</span>
            <span class="text-base text-text-primary">
              <template v-if="reference.volume">{{ reference.volume }} </template>{{ reference.number }}
            </span>
          </div>
          <div class="inline-flex flex-wrap items-center gap-[0.35rem]">
            <SafeExternalLink
              v-if="reference.uri"
              :href="reference.uri"
              class="btn btn-ghost btn-xs"
            >
              {{ t('common.open') }}
            </SafeExternalLink>
            <button type="button" class="btn btn-ghost btn-xs" :disabled="saving" @click="startEdit(reference)">
              {{ t('common.edit') }}
            </button>
            <button type="button" class="btn btn-danger btn-xs" :disabled="saving" @click="removeReference(reference)">
              {{ t('common.delete') }}
            </button>
          </div>
        </template>
      </article>

      <article v-if="editing?.mode === 'create'" class="rounded-sm border border-border-subtle bg-card p-3">
        <form class="flex w-full flex-col gap-2" @submit.prevent="saveCreate">
          <div class="grid gap-2 md:grid-cols-3">
            <select v-model="draft.catalog" class="form-input">
              <option value="" disabled>{{ t('coinDetail.references.selectCatalog') }}</option>
              <option v-for="c in catalogs" :key="c.id" :value="c.catalog">{{ c.catalog }} — {{ c.displayName }}</option>
            </select>
            <input v-model.trim="draft.volume" class="form-input" :placeholder="t('coinDetail.references.volume')" />
            <input v-model.trim="draft.number" class="form-input" :placeholder="t('coinDetail.references.number')" />
          </div>
          <input v-model.trim="draft.uri" class="form-input" :placeholder="t('coinDetail.references.url')" />
          <div class="inline-flex flex-wrap items-center gap-[0.35rem]">
            <button type="submit" class="btn btn-primary btn-sm" :disabled="saving">{{ t('common.save') }}</button>
            <button type="button" class="btn btn-secondary btn-sm" :disabled="saving" @click="cancelEdit">{{ t('common.cancel') }}</button>
          </div>
        </form>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch, onMounted, onUpdated } from 'vue'
import { createCoinReference, deleteCoinReference, updateCoinReference, listCatalogs } from '@/api/client'
import type { Coin, CoinReference, CatalogRegistry, NumistaCandidate } from '@/types'
import { useDialog } from '@/composables/useDialog'
import SafeExternalLink from '@/components/SafeExternalLink.vue'
import CoinNumistaPanel from '@/components/coin/CoinNumistaPanel.vue'
import { t } from '@/i18n'

type ReferenceDraft = {
  catalog: string
  volume: string
  number: string
  uri: string
}

const props = withDefaults(defineProps<{
  coinId: number
  references: CoinReference[]
  coinName?: string
  coinRuler?: string
  coinDenomination?: string
  coinMint?: string
  coinDateRange?: string
  coinMaterial?: Coin['material']
  coinObverseInscription?: string
  coinReverseInscription?: string
}>(), {
  coinName: '',
  coinRuler: '',
  coinDenomination: '',
  coinMint: '',
  coinDateRange: '',
  coinMaterial: 'Other',
  coinObverseInscription: '',
  coinReverseInscription: '',
})

const emit = defineEmits<{
  changed: []
}>()

const { showAlert, showConfirm } = useDialog()

const localReferences = ref<CoinReference[]>([])
const catalogs = ref<CatalogRegistry[]>([])
const saving = ref(false)
const editing = ref<{ mode: 'create' } | { mode: 'edit'; id: number } | null>(null)
const numistaExpanded = ref(false)
const numistaDisclosure = ref<HTMLElement | null>(null)
const pendingNumistaNumber = ref<string | null>(null)
const numistaReferenceBaseline = ref(new Set<number>())
const numistaPanelId = `coin-${props.coinId}-numista-lookup`
let restoreNumistaDisclosureFocus = false
const draft = ref<ReferenceDraft>({
  catalog: '',
  volume: '',
  number: '',
  uri: '',
})

const rows = computed(() =>
  [...localReferences.value].sort((a, b) => {
    if (a.catalog !== b.catalog) return a.catalog.localeCompare(b.catalog)
    return a.number.localeCompare(b.number)
  }),
)

const editingCatalogOptions = computed(() => {
  const opts = catalogs.value.map(c => ({ value: c.catalog, label: `${c.catalog} — ${c.displayName}` }))
  if (draft.value.catalog && !catalogs.value.some(c => c.catalog === draft.value.catalog)) {
    opts.push({ value: draft.value.catalog, label: t('coinDetail.references.legacy', { catalog: draft.value.catalog }) })
  }
  return opts
})

onMounted(async () => {
  try {
    catalogs.value = await listCatalogs()
  } catch {
    // ignore catalog load failure
  }
})

watch(
  () => props.references,
  (next) => {
    localReferences.value = (next ?? []).map((item) => ({ ...item }))
    if (pendingNumistaNumber.value !== null) {
      const refreshedMatch = localReferences.value.some((item) => {
        if (item.catalog.toLowerCase() !== 'numista') return false
        if (pendingNumistaNumber.value) return item.number === pendingNumistaNumber.value
        return !numistaReferenceBaseline.value.has(item.id)
      })
      if (refreshedMatch) {
        restoreNumistaDisclosureFocus = true
        numistaExpanded.value = false
        pendingNumistaNumber.value = null
      }
    }
    if (editing.value?.mode === 'edit') {
      const editID = editing.value.id
      const stillExists = localReferences.value.some((item) => item.id === editID)
      if (!stillExists) editing.value = null
    }
  },
  { immediate: true, deep: true },
)

onUpdated(() => {
  if (!restoreNumistaDisclosureFocus) return
  restoreNumistaDisclosureFocus = false
  numistaDisclosure.value?.focus()
})

async function toggleNumista() {
  numistaExpanded.value = !numistaExpanded.value
  if (!numistaExpanded.value) return
  await nextTick()
  document.getElementById('numista-query')?.focus()
}

function handleNumistaReferenceAdded(candidate?: NumistaCandidate) {
  numistaReferenceBaseline.value = new Set(localReferences.value.map((item) => item.id))
  pendingNumistaNumber.value = candidate ? String(candidate.id) : ''
  emit('changed')
}

function resetDraft(value?: Partial<CoinReference>) {
  draft.value = {
    catalog: value?.catalog ?? '',
    volume: value?.volume ?? '',
    number: value?.number ?? '',
    uri: value?.uri ?? '',
  }
}

function startCreate() {
  editing.value = { mode: 'create' }
  resetDraft()
}

function startEdit(reference: CoinReference) {
  editing.value = { mode: 'edit', id: reference.id }
  resetDraft(reference)
}

function cancelEdit() {
  editing.value = null
  resetDraft()
}

async function saveCreate() {
  if (!draft.value.catalog || !draft.value.number) {
    await showAlert(t('coinDetail.references.required'), { title: t('coinDetail.references.missingData') })
    return
  }

  saving.value = true
  try {
    const created = await createCoinReference(props.coinId, draft.value)
    localReferences.value = [...localReferences.value, created.data]
    editing.value = null
    emit('changed')
  } catch (error) {
    await showAlert(getErrorMessage(error), { title: t('coinDetail.references.addFailed') })
  } finally {
    saving.value = false
  }
}

async function saveEdit() {
  if (editing.value?.mode !== 'edit') return
  if (!draft.value.catalog || !draft.value.number) {
    await showAlert(t('coinDetail.references.required'), { title: t('coinDetail.references.missingData') })
    return
  }

  saving.value = true
  try {
    const editID = editing.value.id
    const updated = await updateCoinReference(props.coinId, editID, draft.value)
    localReferences.value = localReferences.value.map((item) =>
      item.id === editID ? updated.data : item,
    )
    editing.value = null
    emit('changed')
  } catch (error) {
    await showAlert(getErrorMessage(error), { title: t('coinDetail.references.updateFailed') })
  } finally {
    saving.value = false
  }
}

async function removeReference(reference: CoinReference) {
  const confirmed = await showConfirm(
    t('coinDetail.references.confirmDelete', { reference: `${reference.catalog} ${reference.volume ? `${reference.volume} ` : ''}${reference.number}` }),
    { title: t('coinDetail.references.deleteTitle'), variant: 'danger' },
  )
  if (!confirmed) return

  saving.value = true
  try {
    await deleteCoinReference(props.coinId, reference.id)
    localReferences.value = localReferences.value.filter((item) => item.id !== reference.id)
    emit('changed')
  } catch (error) {
    await showAlert(getErrorMessage(error), { title: t('coinDetail.references.deleteFailed') })
  } finally {
    saving.value = false
  }
}

function getErrorMessage(error: unknown): string {
  const err = error as { response?: { data?: { error?: string } } }
  return err.response?.data?.error || t('common.requestFailed')
}
</script>
