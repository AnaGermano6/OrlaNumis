<template>
  <section class="card">
    <h2 class="mb-5 border-b border-border-subtle pb-3 text-lg">{{ t('settings.data.title') }}</h2>
    <section aria-labelledby="custom-mint-locations-heading">
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
  getMintLocations, createMintLocation, updateMintLocation, deleteMintLocation,
  migrateLegacyReferences,
  type MintLocationInput,
  type MintLocationsResponse,
} from '@/api/client'
import { useDialog } from '@/composables/useDialog'
import type { MintLocation, LegacyMigrationResult } from '@/types'
import { t } from '@/i18n'

const { showConfirm } = useDialog()
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
  loadMintLocations()
})

defineExpose({ loadMintLocations })
</script>
