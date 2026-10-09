<template>
  <section class="card">
    <h2 class="mb-5 border-b border-border-subtle pb-3 text-lg">{{ t('settings.data.title') }}</h2>
    <div class="mt-8 grid grid-cols-1 gap-6 md:grid-cols-2">
      <section class="min-w-0" aria-labelledby="tags-heading">
        <h3 id="tags-heading" class="mb-3 text-base font-medium text-text-secondary">{{ t('settings.data.tagsTitle') }}</h3>
        <p class="text-sm text-text-muted">{{ t('settings.data.tagsIntro') }}</p>
        <router-link
          to="/sets"
          class="btn btn-secondary btn-sm mt-3 inline-flex focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        >
          {{ t('settings.data.openSets') }}
        </router-link>

        <div class="my-4 flex flex-wrap items-center gap-2">
          <input
            v-model="newTagName"
            type="text"
            class="form-input min-w-[150px] flex-1 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
            :placeholder="t('settings.data.newTagPlaceholder')"
            maxlength="50"
            @keydown.enter="handleCreateTag"
          />
          <div class="flex items-center gap-[0.3rem]">
            <button
              v-for="c in TAG_COLORS"
              :key="c"
              class="h-[22px] w-[22px] cursor-pointer rounded-full border-2 border-transparent p-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
              :class="newTagColor === c ? 'border-text-primary ring-2 ring-[var(--bg-card)]' : ''"
              :style="{ backgroundColor: c }"
              @click="newTagColor = c"
            ></button>
          </div>
          <button
            class="btn btn-primary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
            @click="handleCreateTag"
            :disabled="!newTagName.trim()"
          >
            {{ t('settings.data.createTag') }}
          </button>
        </div>
        <p v-if="tagError" class="mt-1 text-body text-byzantine">{{ tagError }}</p>

        <div v-if="tagList.length" class="mt-4 flex flex-col gap-2">
          <div v-for="tag in tagList" :key="tag.id" class="flex flex-wrap items-center gap-2 rounded-sm border border-border-subtle p-2">
            <template v-if="editingTag?.id === tag.id">
              <input
                v-model="editTagName"
                class="form-input min-w-[120px] flex-1 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                maxlength="50"
                @keydown.enter="handleSaveTag"
              />
              <div class="flex items-center gap-[0.3rem]">
                <button
                  v-for="c in TAG_COLORS"
                  :key="c"
                  class="h-[18px] w-[18px] cursor-pointer rounded-full border-2 border-transparent p-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                  :class="editTagColor === c ? 'border-text-primary ring-2 ring-[var(--bg-card)]' : ''"
                  :style="{ backgroundColor: c }"
                  @click="editTagColor = c"
                ></button>
              </div>
              <button
                class="btn btn-primary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                @click="handleSaveTag"
              >
                {{ t('common.save') }}
              </button>
              <button
                class="btn btn-secondary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                @click="editingTag = null"
              >
                {{ t('common.cancel') }}
              </button>
            </template>
            <template v-else>
              <span
                class="shrink-0 rounded-full border px-[0.6rem] py-[0.2rem] text-chip"
                :style="{ backgroundColor: tag.color + '22', color: tag.color, borderColor: tag.color + '44' }"
              >
                {{ tag.name }}
              </span>
              <div class="ml-auto flex gap-1">
                <button
                  class="btn btn-secondary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                  @click="startEditTag(tag)"
                >
                  {{ t('common.edit') }}
                </button>
                <button
                  class="btn btn-danger btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                  @click="handleDeleteTag(tag)"
                >
                  {{ t('common.delete') }}
                </button>
              </div>
            </template>
          </div>
        </div>
        <p v-else class="mt-4 text-body text-text-secondary">{{ t('settings.data.noTags') }}</p>
      </section>

      <section class="min-w-0" aria-labelledby="storage-locations-heading">
        <h3 id="storage-locations-heading" class="mb-3 text-base font-medium text-text-secondary">{{ t('settings.data.storageTitle') }}</h3>
        <p class="text-sm text-text-muted">{{ t('settings.data.storageIntro') }}</p>

        <div class="my-4 flex flex-wrap items-center gap-2">
          <select
            v-model="newStorageLocationType"
            :aria-label="t('settings.data.storageType')"
            class="form-select"
            :disabled="storageLocationSaving"
          >
            <option value="standard">{{ t('settings.data.standardLocation') }}</option>
            <option value="tray">{{ t('settings.data.coinTray') }}</option>
          </select>
          <input
            v-model="newStorageLocationName"
            type="text"
            class="form-input min-w-[150px] flex-1 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
            :placeholder="t('settings.data.newStoragePlaceholder')"
            maxlength="100"
            :disabled="storageLocationSaving"
            @keydown.enter="handleCreateStorageLocation"
          />
          <template v-if="newStorageLocationType === 'tray'">
            <input v-model.number="newStorageLocationRows" :aria-label="t('settings.data.trayRows')" class="form-input w-24" type="number" min="1" max="20" />
            <span aria-hidden="true">×</span>
            <input v-model.number="newStorageLocationColumns" :aria-label="t('settings.data.trayColumns')" class="form-input w-24" type="number" min="1" max="20" />
            <span class="text-sm text-text-muted">{{ t('settings.data.slots', { count: newStorageLocationRows * newStorageLocationColumns }) }}</span>
          </template>
          <button
            class="btn btn-primary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
            @click="handleCreateStorageLocation"
            :disabled="!newStorageLocationName.trim() || storageLocationSaving"
          >
            {{ storageLocationSaving ? t('common.saving') : t('settings.data.createLocation') }}
          </button>
        </div>
        <p v-if="storageLocationError" class="mt-1 text-body text-byzantine">{{ storageLocationError }}</p>
        <p v-if="storageLocationsLoading" class="mt-4 text-body text-text-secondary">{{ t('settings.data.loadingStorage') }}</p>

        <div v-else-if="storageLocationList.length" class="mt-4 flex flex-col gap-2">
          <div v-for="location in storageLocationList" :key="location.id" class="flex flex-wrap items-center gap-2 rounded-sm border border-border-subtle p-2">
            <template v-if="editingStorageLocation?.id === location.id">
              <input
                v-model="editStorageLocationName"
                class="form-input min-w-[120px] flex-1 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                maxlength="100"
                @keydown.enter="handleSaveStorageLocation"
              />
              <template v-if="location.type === 'tray'">
                <input v-model.number="editStorageLocationRows" :aria-label="t('settings.data.trayRows')" class="form-input w-24" type="number" min="1" max="20" :disabled="location.occupied > 0" />
                <span aria-hidden="true">×</span>
                <input v-model.number="editStorageLocationColumns" :aria-label="t('settings.data.trayColumns')" class="form-input w-24" type="number" min="1" max="20" :disabled="location.occupied > 0" />
                <span v-if="location.occupied > 0" class="text-sm text-text-muted">{{ t('settings.data.emptyTray') }}</span>
              </template>
              <button
                class="btn btn-primary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                @click="handleSaveStorageLocation"
                :disabled="storageLocationSaving"
              >
                {{ t('common.save') }}
              </button>
              <button
                class="btn btn-secondary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                @click="editingStorageLocation = null"
                :disabled="storageLocationSaving"
              >
                {{ t('common.cancel') }}
              </button>
            </template>
            <template v-else>
              <span class="chip-sm shrink-0 bg-input text-text-primary">{{ location.name }}</span>
              <span class="text-sm text-text-muted">
                {{ location.type === 'tray' ? t('settings.data.trayInfo', { rows: location.rows ?? 0, columns: location.columns ?? 0, occupied: location.occupied, capacity: location.capacity }) : t('settings.data.standardLocation') }}
              </span>
              <div class="ml-auto flex gap-1">
                <button
                  class="btn btn-secondary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                  @click="startEditStorageLocation(location)"
                >
                  {{ t('common.edit') }}
                </button>
                <button
                  class="btn btn-danger btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                  :disabled="deletingStorageLocationId === location.id"
                  @click="handleDeleteStorageLocation(location)"
                >
                  {{ deletingStorageLocationId === location.id ? t('common.deleting') : t('common.delete') }}
                </button>
              </div>
            </template>
          </div>
        </div>
        <p v-else class="mt-4 text-body text-text-secondary">{{ t('settings.data.noStorage') }}</p>
      </section>
    </div>

    <section class="mt-8 border-t border-border-subtle pt-8" aria-labelledby="custom-mint-locations-heading">
      <div class="mb-3 flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div>
          <h3 id="custom-mint-locations-heading" class="m-0 text-base font-medium text-text-secondary">{{ t('settings.data.mintTitle') }}</h3>
          <p class="mt-1 text-sm text-text-muted">{{ t('settings.data.mintIntro') }}</p>
        </div>
        <span class="chip-sm shrink-0">{{ t('settings.data.locationCount', { count: userMintLocations.length }) }}</span>
      </div>

      <p v-if="mintLocationError" class="mb-2 text-body text-byzantine">{{ mintLocationError }}</p>
      <p v-if="mintLocationsLoading" class="text-body text-text-secondary">{{ t('common.loading') }}</p>

      <div v-else class="grid gap-4 md:grid-cols-[minmax(0,1.1fr)_minmax(260px,0.9fr)] md:items-start">
        <div class="flex max-h-[22rem] min-w-0 flex-col gap-2 overflow-y-auto pr-1 [scrollbar-gutter:stable]" :aria-label="t('settings.data.mintList')">
          <div v-if="!userMintLocations.length" class="rounded-sm border border-border-subtle bg-input p-3 text-body text-text-secondary">
            {{ t('settings.data.noMints') }}
          </div>
          <div
            v-for="location in sortedUserMintLocations"
            :key="location.id"
            class="rounded-sm border border-border-subtle bg-input p-3"
          >
            <div class="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-3">
              <div class="flex min-w-0 flex-col gap-1">
                <strong class="min-w-0 text-text-primary [overflow-wrap:anywhere]">{{ location.displayName }}</strong>
                <span class="block text-body text-text-secondary [overflow-wrap:anywhere]">{{ location.region || t('settings.data.noRegion') }} · {{ location.lat }}, {{ location.lng }}</span>
                <span v-if="location.aliases.length" class="block text-body text-text-secondary [overflow-wrap:anywhere]">{{ location.aliases.join(', ') }}</span>
              </div>
              <div class="flex shrink-0 flex-wrap justify-end gap-[0.35rem]">
                <button
                  type="button"
                  class="btn btn-secondary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                  :disabled="mintLocationSaving"
                  @click="startEditMintLocation(location)"
                >
                  {{ t('common.edit') }}
                </button>
                <button
                  type="button"
                  class="btn btn-danger btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
                  :disabled="deletingMintLocationId === location.id"
                  @click="handleDeleteMintLocation(location)"
                >
                  {{ deletingMintLocationId === location.id ? t('common.deleting') : t('common.delete') }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <form class="flex min-w-0 flex-col gap-3 self-start rounded-sm border border-border-subtle bg-input p-3" @submit.prevent="handleSaveMintLocation">
          <h4 class="m-0 text-base font-medium text-heading">{{ editingMintLocation ? t('settings.data.editLocation') : t('settings.data.addLocation') }}</h4>
          <div class="grid gap-3 md:grid-cols-2">
            <label class="flex flex-col gap-[0.35rem]">
              <span class="form-label">{{ t('settings.data.displayName') }}</span>
              <input
                v-model="mintLocationForm.displayName"
                class="form-input"
                type="text"
                maxlength="120"
                required
                :placeholder="t('settings.data.displayNamePlaceholder')"
              />
            </label>
            <label class="flex flex-col gap-[0.35rem]">
              <span class="form-label">{{ t('settings.data.region') }}</span>
              <input
                v-model="mintLocationForm.region"
                class="form-input"
                type="text"
                maxlength="120"
                :placeholder="t('settings.data.regionPlaceholder')"
              />
            </label>
            <label class="flex flex-col gap-[0.35rem]">
              <span class="form-label">{{ t('settings.data.latitude') }}</span>
              <input
                v-model="mintLocationForm.lat"
                class="form-input"
                type="number"
                min="-90"
                max="90"
                step="0.000001"
                required
                placeholder="41.9"
              />
            </label>
            <label class="flex flex-col gap-[0.35rem]">
              <span class="form-label">{{ t('settings.data.longitude') }}</span>
              <input
                v-model="mintLocationForm.lng"
                class="form-input"
                type="number"
                min="-180"
                max="180"
                step="0.000001"
                required
                placeholder="12.5"
              />
            </label>
          </div>
          <label class="flex flex-col gap-[0.35rem]">
            <span class="form-label">{{ t('settings.data.aliases') }}</span>
            <textarea
              v-model="mintLocationForm.aliases"
              class="form-textarea min-h-16 resize-y bg-card text-body"
              rows="3"
              :placeholder="t('settings.data.aliasesPlaceholder')"
            />
            <span class="form-hint">{{ t('settings.data.aliasesHint') }}</span>
          </label>
          <div class="flex flex-wrap justify-start gap-[0.35rem]">
            <button
              type="submit"
              class="btn btn-primary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
              :disabled="mintLocationSaving"
            >
              {{ mintLocationSaving ? t('common.saving') : editingMintLocation ? t('settings.data.saveLocation') : t('settings.data.addLocation') }}
            </button>
            <button
              v-if="editingMintLocation"
              type="button"
              class="btn btn-secondary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
              :disabled="mintLocationSaving"
              @click="resetMintLocationForm"
            >
              {{ t('common.cancel') }}
            </button>
          </div>
        </form>
      </div>
    </section>

    <section class="mt-8 border-t border-border-subtle pt-8" aria-labelledby="migration-heading">
      <div class="mb-2 flex items-center gap-2">
        <Database :size="20" />
        <h3 id="migration-heading" class="m-0 text-base font-medium text-text-secondary">{{ t('settings.data.migrationTitle') }}</h3>
      </div>
      <p class="text-sm text-text-muted">
        {{ t('settings.data.migrationIntro') }}
      </p>
      
      <button
        class="btn btn-primary mt-4 inline-flex items-center gap-2 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
        :disabled="migrationRunning"
        @click="handleMigrate"
      >
        <RefreshCw :size="16" :class="migrationRunning ? 'animate-spin' : ''" />
        {{ migrationRunning ? t('settings.data.migrating') : t('settings.data.runMigration') }}
      </button>

      <div v-if="migrationResult" class="mt-6 rounded-sm border border-border-subtle bg-input p-4">
        <div class="grid grid-cols-1 gap-3 md:grid-cols-3 md:gap-4">
          <div class="flex flex-col gap-[0.35rem] text-center">
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('settings.data.succeeded') }}</span>
            <span class="text-xl font-semibold text-gold">{{ migrationResult.succeeded }}</span>
          </div>
          <div class="flex flex-col gap-[0.35rem] text-center">
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('settings.data.skipped') }}</span>
            <span class="text-xl font-semibold text-text-secondary">{{ migrationResult.skipped }}</span>
          </div>
          <div class="flex flex-col gap-[0.35rem] text-center">
            <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-muted">{{ t('settings.data.failed') }}</span>
            <span class="text-xl font-semibold text-warning">{{ migrationResult.failed }}</span>
          </div>
        </div>
        <p v-if="migrationResult.message" class="mt-3 border-t border-border-subtle pt-3 text-center text-body text-text-secondary">
          {{ migrationResult.message }}
        </p>
      </div>

      <p v-if="migrationError" class="mt-1 text-body text-byzantine">{{ migrationError }}</p>
    </section>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { Database, RefreshCw } from 'lucide-vue-next'
import {
  getTags, createTag, updateTag as updateTagApi, deleteTag,
  getStorageLocations, createStorageLocation, updateStorageLocation, deleteStorageLocation,
  getMintLocations, createMintLocation, updateMintLocation, deleteMintLocation,
  migrateLegacyReferences,
  type MintLocationInput,
  type MintLocationsResponse,
} from '@/api/client'
import { useDialog } from '@/composables/useDialog'
import type { Tag, StorageLocation, MintLocation, LegacyMigrationResult } from '@/types'
import { t } from '@/i18n'

const { showConfirm } = useDialog()
// Tag management
const tagList = ref<Tag[]>([])
const newTagName = ref('')
const newTagColor = ref('#6b7280')
const editingTag = ref<Tag | null>(null)
const editTagName = ref('')
const editTagColor = ref('')
const tagError = ref('')

const TAG_COLORS = ['#6b7280', '#ef4444', '#f59e0b', '#10b981', '#3b82f6', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316', '#6366f1']

async function loadTags() {
  try {
    const res = await getTags()
    tagList.value = res.data?.tags ?? []
  } catch { tagList.value = [] }
}

async function handleCreateTag() {
  tagError.value = ''
  const name = newTagName.value.trim()
  if (!name) return
  try {
    await createTag({ name, color: newTagColor.value })
    newTagName.value = ''
    newTagColor.value = '#6b7280'
    await loadTags()
  } catch (e: unknown) {
    if (typeof e === 'object' && e !== null && 'response' in e) {
      const axiosErr = e as { response?: { data?: { error?: string } } }
      tagError.value = axiosErr.response?.data?.error ?? t('settings.data.createTagFailed')
    } else {
      tagError.value = t('settings.data.createTagFailed')
    }
  }
}

function startEditTag(tag: Tag) {
  editingTag.value = tag
  editTagName.value = tag.name
  editTagColor.value = tag.color
}

async function handleSaveTag() {
  tagError.value = ''
  if (!editingTag.value) return
  try {
    await updateTagApi(editingTag.value.id, { name: editTagName.value.trim(), color: editTagColor.value })
    editingTag.value = null
    await loadTags()
  } catch (e: unknown) {
    if (typeof e === 'object' && e !== null && 'response' in e) {
      const axiosErr = e as { response?: { data?: { error?: string } } }
      tagError.value = axiosErr.response?.data?.error ?? t('settings.data.updateTagFailed')
    } else {
      tagError.value = t('settings.data.updateTagFailed')
    }
  }
}

async function handleDeleteTag(tag: Tag) {
  const confirmed = await showConfirm(t('settings.data.confirmDeleteTag', { name: tag.name }), { title: t('settings.data.deleteTagTitle'), variant: 'danger' })
  if (!confirmed) return
  try {
    await deleteTag(tag.id)
    await loadTags()
  } catch { /* ignore */ }
}

// Storage location management
const storageLocationList = ref<StorageLocation[]>([])
const newStorageLocationName = ref('')
const newStorageLocationType = ref<'standard' | 'tray'>('standard')
const newStorageLocationRows = ref(3)
const newStorageLocationColumns = ref(3)
const editingStorageLocation = ref<StorageLocation | null>(null)
const editStorageLocationName = ref('')
const editStorageLocationRows = ref(1)
const editStorageLocationColumns = ref(1)
const storageLocationError = ref('')
const storageLocationsLoading = ref(false)
const storageLocationSaving = ref(false)
const deletingStorageLocationId = ref<number | null>(null)

function apiErrorText(error: unknown, fallback: string): string {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const axiosErr = error as { response?: { status?: number; data?: { error?: string; message?: string; count?: number } } }
    const message = axiosErr.response?.data?.message ?? axiosErr.response?.data?.error
    if (axiosErr.response?.status === 409) {
      return message ?? t('settings.data.locationInUse')
    }
    return message ?? fallback
  }
  return fallback
}

async function loadStorageLocations() {
  storageLocationsLoading.value = true
  storageLocationError.value = ''
  try {
    const res = await getStorageLocations()
    storageLocationList.value = res.data?.storageLocations ?? []
  } catch {
    storageLocationList.value = []
    storageLocationError.value = t('settings.data.loadStorageFailed')
  } finally {
    storageLocationsLoading.value = false
  }
}

async function handleCreateStorageLocation() {
  storageLocationError.value = ''
  const name = newStorageLocationName.value.trim()
  if (!name) return
  storageLocationSaving.value = true
  try {
    await createStorageLocation({
      name,
      type: newStorageLocationType.value,
      rows: newStorageLocationType.value === 'tray' ? newStorageLocationRows.value : null,
      columns: newStorageLocationType.value === 'tray' ? newStorageLocationColumns.value : null,
    })
    newStorageLocationName.value = ''
    await loadStorageLocations()
  } catch (error: unknown) {
    storageLocationError.value = apiErrorText(error, t('settings.data.createStorageFailed'))
  } finally {
    storageLocationSaving.value = false
  }
}

function startEditStorageLocation(location: StorageLocation) {
  editingStorageLocation.value = location
  editStorageLocationName.value = location.name
  editStorageLocationRows.value = location.rows ?? 1
  editStorageLocationColumns.value = location.columns ?? 1
  storageLocationError.value = ''
}

async function handleSaveStorageLocation() {
  storageLocationError.value = ''
  if (!editingStorageLocation.value) return
  const name = editStorageLocationName.value.trim()
  if (!name) return
  storageLocationSaving.value = true
  try {
    await updateStorageLocation(editingStorageLocation.value.id, {
      name,
      ...(editingStorageLocation.value.type === 'tray' && editingStorageLocation.value.occupied === 0
        ? { rows: editStorageLocationRows.value, columns: editStorageLocationColumns.value }
        : {}),
    })
    editingStorageLocation.value = null
    await loadStorageLocations()
  } catch (error: unknown) {
    storageLocationError.value = apiErrorText(error, t('settings.data.updateStorageFailed'))
  } finally {
    storageLocationSaving.value = false
  }
}

async function handleDeleteStorageLocation(location: StorageLocation) {
  storageLocationError.value = ''
  const confirmed = await showConfirm(t('settings.data.confirmDeleteStorage', { name: location.name }), { title: t('settings.data.deleteStorageTitle'), variant: 'danger' })
  if (!confirmed) return
  deletingStorageLocationId.value = location.id
  try {
    await deleteStorageLocation(location.id)
    await loadStorageLocations()
  } catch (error: unknown) {
    storageLocationError.value = apiErrorText(error, t('settings.data.deleteStorageFailed'))
  } finally {
    deletingStorageLocationId.value = null
  }
}

// Mint location management (user-scoped)
const allMintLocations = ref<MintLocation[]>([])
const userMintLocations = computed(() => allMintLocations.value.filter((m) => m.userId != null))
const sortedUserMintLocations = computed(() =>
  [...userMintLocations.value].sort((a, b) => a.displayName.localeCompare(b.displayName)),
)
const mintLocationsLoading = ref(false)
const mintLocationSaving = ref(false)
const deletingMintLocationId = ref<number | null>(null)
const mintLocationError = ref('')
const editingMintLocation = ref<MintLocation | null>(null)
const mintLocationForm = reactive({
  displayName: '',
  region: '',
  lat: '',
  lng: '',
  aliases: '',
})

function unwrapMintLocations(data: MintLocationsResponse): MintLocation[] {
  return Array.isArray(data) ? data : data.mintLocations ?? []
}

function parseMintAliases(value: string): string[] {
  return value
    .split(/[\n,]+/)
    .map((item) => item.trim())
    .filter((item, index, items) => item.length > 0 && items.indexOf(item) === index)
}

function buildMintLocationPayload(): MintLocationInput | null {
  const displayName = mintLocationForm.displayName.trim()
  const lat = Number(mintLocationForm.lat)
  const lng = Number(mintLocationForm.lng)
  if (!displayName) {
    mintLocationError.value = t('settings.data.displayNameRequired')
    return null
  }
  if (!Number.isFinite(lat) || lat < -90 || lat > 90) {
    mintLocationError.value = t('settings.data.latitudeRange')
    return null
  }
  if (!Number.isFinite(lng) || lng < -180 || lng > 180) {
    mintLocationError.value = t('settings.data.longitudeRange')
    return null
  }
  return {
    displayName,
    lat,
    lng,
    region: mintLocationForm.region.trim(),
    aliases: parseMintAliases(mintLocationForm.aliases),
  }
}

function resetMintLocationForm() {
  editingMintLocation.value = null
  mintLocationForm.displayName = ''
  mintLocationForm.region = ''
  mintLocationForm.lat = ''
  mintLocationForm.lng = ''
  mintLocationForm.aliases = ''
  mintLocationError.value = ''
}

function startEditMintLocation(location: MintLocation) {
  editingMintLocation.value = location
  mintLocationForm.displayName = location.displayName
  mintLocationForm.region = location.region ?? ''
  mintLocationForm.lat = String(location.lat)
  mintLocationForm.lng = String(location.lng)
  mintLocationForm.aliases = location.aliases.join('\n')
  mintLocationError.value = ''
}

async function loadMintLocations() {
  mintLocationsLoading.value = true
  mintLocationError.value = ''
  try {
    const res = await getMintLocations()
    allMintLocations.value = unwrapMintLocations(res.data)
  } catch (error: unknown) {
    allMintLocations.value = []
    mintLocationError.value = apiErrorText(error, t('settings.data.loadMintsFailed'))
  } finally {
    mintLocationsLoading.value = false
  }
}

async function handleSaveMintLocation() {
  mintLocationError.value = ''
  const payload = buildMintLocationPayload()
  if (!payload) return
  mintLocationSaving.value = true
  try {
    if (editingMintLocation.value) {
      await updateMintLocation(editingMintLocation.value.id, payload)
    } else {
      await createMintLocation(payload)
    }
    resetMintLocationForm()
    await loadMintLocations()
  } catch (error: unknown) {
    mintLocationError.value = apiErrorText(error, t('settings.data.saveMintFailed'))
  } finally {
    mintLocationSaving.value = false
  }
}

async function handleDeleteMintLocation(location: MintLocation) {
  mintLocationError.value = ''
  const confirmed = await showConfirm(
    t('settings.data.confirmDeleteMint', { name: location.displayName }),
    { title: t('settings.data.deleteMintTitle'), variant: 'danger' },
  )
  if (!confirmed) return
  deletingMintLocationId.value = location.id
  try {
    await deleteMintLocation(location.id)
    if (editingMintLocation.value?.id === location.id) {
      resetMintLocationForm()
    }
    await loadMintLocations()
  } catch (error: unknown) {
    mintLocationError.value = apiErrorText(error, t('settings.data.deleteMintFailed'))
  } finally {
    deletingMintLocationId.value = null
  }
}

// Migration
const migrationRunning = ref(false)
const migrationResult = ref<LegacyMigrationResult | null>(null)
const migrationError = ref('')

async function handleMigrate() {
  migrationRunning.value = true
  migrationError.value = ''
  migrationResult.value = null
  
  try {
    const res = await migrateLegacyReferences()
    migrationResult.value = res.data
  } catch (error: unknown) {
    migrationError.value = apiErrorText(error, t('settings.data.migrationFailed'))
  } finally {
    migrationRunning.value = false
  }
}

onMounted(() => {
  loadTags()
  loadStorageLocations()
  loadMintLocations()
})

defineExpose({ loadTags, loadStorageLocations, loadMintLocations })
</script>
