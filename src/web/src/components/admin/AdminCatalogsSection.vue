<template>
  <section class="card p-6">
    <h2 class="text-xl font-medium text-heading">{{ t('admin.catalogs.title') }}</h2>
    <p class="mb-6 text-body leading-6 text-text-secondary">
      {{ t('admin.catalogs.intro') }}
    </p>

    <div v-if="loading" class="flex items-center justify-center p-12">
      <div class="spinner"></div>
    </div>

    <div
      v-else-if="error"
      class="flex items-center gap-2 rounded-sm border border-[color-mix(in_srgb,var(--color-negative)_30%,transparent)] bg-[color-mix(in_srgb,var(--color-negative)_10%,transparent)] p-4 text-body text-loss"
      role="alert"
    >
      <AlertCircle :size="20" />
      <span>{{ error }}</span>
    </div>

    <template v-else>
      <div class="mb-4 flex justify-end">
        <button class="btn btn-primary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" @click="openCreateForm">
          {{ t('admin.catalogs.add') }}
        </button>
      </div>

      <div v-if="catalogs.length === 0" class="py-8 text-center text-body text-text-muted">
        {{ t('admin.catalogs.empty') }}
      </div>

      <div v-else class="overflow-x-auto">
        <table class="data-table text-body [&_th]:px-2 [&_th]:py-3 [&_td]:px-2 [&_td]:py-3 [&_td]:align-top">
          <thead>
            <tr>
              <th>{{ t('admin.catalogs.code') }}</th>
              <th>{{ t('admin.catalogs.displayName') }}</th>
              <th class="hidden md:table-cell">{{ t('admin.catalogs.volumeRequired') }}</th>
              <th>{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="cat in catalogs" :key="cat.id" class="last:[&>td]:border-b-0">
              <td>
                <div class="flex flex-col items-start gap-[0.35rem]">
                  <div class="font-semibold text-gold">{{ cat.catalog }}</div>
                  <span
                    class="chip-sm capitalize"
                    :style="{
                      borderColor: colorForLabel(cat.era),
                      backgroundColor: colorForLabelBackground(cat.era),
                      color: colorForLabel(cat.era),
                    }"
                  >
                    {{ eraLabel(cat.era) }}
                  </span>
                </div>
              </td>
              <td class="text-text-primary">{{ cat.displayName }}</td>
              <td class="hidden md:table-cell">
                <BaseToggle
                  size="sm"
                  :model-value="cat.volumeRequired"
                  disabled
                  :label="t('admin.catalogs.volumeRequiredFor', { name: cat.displayName })"
                />
              </td>
              <td>
                <div class="flex flex-wrap justify-end gap-[0.35rem]">
                  <button class="btn btn-ghost btn-xs focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" @click="openEditForm(cat)">
                    {{ t('common.edit') }}
                  </button>
                  <button class="btn btn-danger btn-xs focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" @click="handleDelete(cat.id)">
                    {{ t('common.delete') }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="showForm" class="fixed inset-0 z-[200] flex items-center justify-center bg-overlay-60 p-4" @click.self="closeForm">
        <div class="max-h-[90vh] w-full max-w-[500px] overflow-auto rounded-md border border-border-subtle bg-card shadow-[var(--shadow-card)]">
          <div class="flex items-center justify-between gap-4 border-b border-border-subtle px-6 py-4">
            <h3 class="m-0 text-lg font-medium text-heading">{{ editingCatalog ? t('admin.catalogs.editTitle') : t('admin.catalogs.addTitle') }}</h3>
            <button
              class="inline-flex h-[30px] w-[30px] items-center justify-center p-0 text-text-muted transition-colors hover:text-text-primary focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
              @click="closeForm"
            >
              ×
            </button>
          </div>
          <form class="p-6" @submit.prevent="saveForm">
            <div class="form-group">
              <label class="form-label">{{ t('admin.catalogs.catalogCode') }}<span class="ml-[0.15rem] text-loss">*</span></label>
              <input
                v-model.trim="formData.catalog"
                class="form-input"
                :placeholder="t('admin.catalogs.codePlaceholder')"
                required
                :disabled="!!editingCatalog"
              />
            </div>
            <div class="form-group">
              <label class="form-label">{{ t('admin.catalogs.displayName') }}<span class="ml-[0.15rem] text-loss">*</span></label>
              <input
                v-model.trim="formData.displayName"
                class="form-input"
                :placeholder="t('admin.catalogs.namePlaceholder')"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label">{{ t('admin.catalogs.era') }}<span class="ml-[0.15rem] text-loss">*</span></label>
              <select v-model="formData.era" class="form-input" required>
                <option value="" disabled>{{ t('admin.catalogs.selectEra') }}</option>
                <option v-for="era in eraOptions" :key="era" :value="era">{{ eraLabel(era) }}</option>
              </select>
            </div>
            <div class="form-group flex items-center justify-between gap-4">
              <label class="form-label mb-0">{{ t('admin.catalogs.volumeRequired') }}</label>
              <BaseToggle
                v-model="formData.volumeRequired"
                :label="t('admin.catalogs.volumeRequiredLabel')"
              />
            </div>
            <div class="mt-6 flex flex-col gap-2 md:flex-row md:justify-end">
              <button type="button" class="btn btn-secondary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" @click="closeForm">
                {{ t('common.cancel') }}
              </button>
              <button type="submit" class="btn btn-primary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" :disabled="saving">
                {{ saving ? t('common.saving') : t('common.save') }}
              </button>
            </div>
            <div
              v-if="formError"
              class="mt-4 flex items-center gap-2 rounded-sm border border-[color-mix(in_srgb,var(--color-negative)_30%,transparent)] bg-[color-mix(in_srgb,var(--color-negative)_10%,transparent)] p-3 text-sm text-loss"
              role="alert"
            >
              <AlertCircle :size="16" />
              <span>{{ formError }}</span>
            </div>
          </form>
        </div>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { BaseToggle } from '@/components/ui'
import { ref, onMounted } from 'vue'
import { AlertCircle } from 'lucide-vue-next'
import { listCatalogs, adminCreateCatalog, adminUpdateCatalog, adminDeleteCatalog } from '@/api/client'
import type { CatalogRegistry } from '@/types'
import { useDialog } from '@/composables/useDialog'
import { useCoinOptions } from '@/composables/useCoinOptions'
import { colorForLabel, colorForLabelBackground } from '@/utils/categoryColor'
import { t } from '@/i18n'
import { eraLabel } from '@/i18n/coinVocabulary'

const { showAlert, showConfirm } = useDialog()
const { eraOptions, loadOptions: loadCoinOptions } = useCoinOptions()

const catalogs = ref<CatalogRegistry[]>([])
const loading = ref(true)
const error = ref('')
const saving = ref(false)
const showForm = ref(false)
const editingCatalog = ref<CatalogRegistry | null>(null)
const formError = ref('')
const formData = ref({
  catalog: '',
  displayName: '',
  era: '',
  volumeRequired: false,
})

async function loadCatalogs() {
  loading.value = true
  error.value = ''
  try {
    catalogs.value = await listCatalogs()
  } catch (_e) {
    error.value = t('admin.catalogs.loadFailed')
  } finally {
    loading.value = false
  }
}

function openCreateForm() {
  editingCatalog.value = null
  formData.value = {
    catalog: '',
    displayName: '',
    era: '',
    volumeRequired: false,
  }
  formError.value = ''
  showForm.value = true
}

function openEditForm(catalog: CatalogRegistry) {
  editingCatalog.value = catalog
  formData.value = {
    catalog: catalog.catalog,
    displayName: catalog.displayName,
    era: catalog.era,
    volumeRequired: catalog.volumeRequired,
  }
  formError.value = ''
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editingCatalog.value = null
  formData.value = {
    catalog: '',
    displayName: '',
    era: '',
    volumeRequired: false,
  }
  formError.value = ''
}

async function saveForm() {
  if (!formData.value.catalog || !formData.value.displayName || !formData.value.era) {
    formError.value = t('admin.catalogs.requiredFields')
    return
  }

  saving.value = true
  formError.value = ''
  try {
    const payload = {
      catalog: formData.value.catalog,
      displayName: formData.value.displayName,
      era: formData.value.era,
      volumeRequired: formData.value.volumeRequired,
    }

    if (editingCatalog.value) {
      await adminUpdateCatalog(editingCatalog.value.id, payload)
    } else {
      await adminCreateCatalog(payload)
    }

    await loadCatalogs()
    closeForm()
  } catch (e: unknown) {
    const err = e as { response?: { data?: { error?: string } } }
    formError.value = err.response?.data?.error || t('admin.catalogs.saveFailed')
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: number) {
  const confirmed = await showConfirm(
    t('admin.catalogs.confirmDelete'),
    { title: t('admin.catalogs.deleteTitle'), variant: 'danger' },
  )
  if (!confirmed) return

  try {
    await adminDeleteCatalog(id)
    await loadCatalogs()
  } catch (e: unknown) {
    const err = e as { response?: { status?: number; data?: { error?: string } } }
    if (err.response?.status === 409) {
      await showAlert(
        t('admin.catalogs.inUse'),
        { title: t('admin.catalogs.cannotDelete') },
      )
    } else {
      await showAlert(
        err.response?.data?.error || t('admin.catalogs.deleteFailed'),
        { title: t('admin.catalogs.deleteFailedTitle') },
      )
    }
  }
}

onMounted(() => {
  loadCatalogs()
  loadCoinOptions()
})
</script>
