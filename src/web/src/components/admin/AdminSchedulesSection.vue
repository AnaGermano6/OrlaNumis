<template>
  <section class="admin-section card flex flex-col">
    <h2 class="mb-5 border-b border-border-subtle pb-3 text-xl font-medium">{{ t('admin.tabs.schedules') }}</h2>

    <AdminScheduleTimezone
      :settings="settings"
      :settings-saving="settingsSaving"
      @save="emit('save')"
    />

    <hr class="my-6 border-0 border-t border-border-subtle" />

    <AdminAvailabilitySchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      :settings-msg="availSettingsMsg"
      :settings-error="availSettingsError"
      @save="emit('save')"
      @update:settings-msg="emit('update:availSettingsMsg', $event)"
      @update:settings-error="emit('update:availSettingsError', $event)"
    />

    <hr class="my-6 border-0 border-t border-border-subtle" />

    <!-- ParcelApp Shipment Tracking -->
    <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.parcelTitle') }}</h3>
    <p class="mb-4 text-base text-text-secondary">{{ t('admin.schedules.parcelIntro') }}</p>
    <div class="mb-4">
      <div class="form-group flex items-center justify-between gap-3">
        <label class="form-label">{{ t('admin.schedules.parcelEnable') }}</label>
        <BaseToggle
          size="sm"
          :label="t('admin.schedules.parcelEnable')"
          :model-value="settings.ParcelAppEnabled === 'true'"
          @update:model-value="settings.ParcelAppEnabled = $event ? 'true' : 'false'"
        />
      </div>
      <div class="form-group">
        <label class="form-label">{{ t('admin.schedules.repeatMinutes') }}</label>
        <input
          v-model="settings.ShipmentSyncInterval"
          class="form-input w-full max-w-[120px]"
          type="number"
          min="20"
          step="5"
        />
        <span class="form-hint">{{ t('admin.schedules.parcelMinimum') }}</span>
      </div>
      <div class="mt-4 flex w-full flex-col gap-3 md:flex-row md:items-center">
        <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="emit('save')">
          {{ settingsSaving ? t('common.saving') : t('admin.schedules.parcelSave') }}
        </button>
      </div>
    </div>

    <AdminWishlistSearchAlertSchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      @save="emit('save')"
    />

    <AdminAuctionEndingSchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      :settings-msg="auctionSettingsMsg"
      :settings-error="auctionSettingsError"
      @save="emit('save')"
      @update:settings-msg="emit('update:auctionSettingsMsg', $event)"
      @update:settings-error="emit('update:auctionSettingsError', $event)"
    />

    <AdminAuctionAlertReminderSchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      :settings-msg="alertReminderSettingsMsg"
      :settings-error="alertReminderSettingsError"
      @save="emit('save')"
      @update:settings-msg="emit('update:alertReminderSettingsMsg', $event)"
      @update:settings-error="emit('update:alertReminderSettingsError', $event)"
    />

    <AdminAuctionWatchBidDigestSchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      :settings-msg="watchBidDigestSettingsMsg"
      :settings-error="watchBidDigestSettingsError"
      @save="emit('save')"
      @update:settings-msg="emit('update:watchBidDigestSettingsMsg', $event)"
      @update:settings-error="emit('update:watchBidDigestSettingsError', $event)"
    />

    <AdminValuationSchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      :settings-msg="valSettingsMsg"
      :settings-error="valSettingsError"
      @save="emit('save')"
      @update:settings-msg="emit('update:valSettingsMsg', $event)"
      @update:settings-error="emit('update:valSettingsError', $event)"
    />

    <AdminCollectionHealthSchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      :settings-msg="healthSettingsMsg"
      :settings-error="healthSettingsError"
      @save="emit('save')"
      @update:settings-msg="emit('update:healthSettingsMsg', $event)"
      @update:settings-error="emit('update:healthSettingsError', $event)"
    />

    <AdminCoinOfDaySchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      @save="emit('save')"
    />

    <AdminPurchaseReminderSchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      @save="emit('save')"
    />
  </section>
</template>

<script setup lang="ts">
import { BaseToggle } from '@/components/ui'
import AdminAuctionAlertReminderSchedule from '@/components/admin/schedules/AdminAuctionAlertReminderSchedule.vue'
import AdminAuctionEndingSchedule from '@/components/admin/schedules/AdminAuctionEndingSchedule.vue'
import AdminAuctionWatchBidDigestSchedule from '@/components/admin/schedules/AdminAuctionWatchBidDigestSchedule.vue'
import AdminAvailabilitySchedule from '@/components/admin/schedules/AdminAvailabilitySchedule.vue'
import AdminWishlistSearchAlertSchedule from '@/components/admin/schedules/AdminWishlistSearchAlertSchedule.vue'
import AdminCoinOfDaySchedule from '@/components/admin/schedules/AdminCoinOfDaySchedule.vue'
import AdminPurchaseReminderSchedule from '@/components/admin/schedules/AdminPurchaseReminderSchedule.vue'
import AdminScheduleTimezone from '@/components/admin/schedules/AdminScheduleTimezone.vue'
import AdminCollectionHealthSchedule from '@/components/admin/schedules/AdminCollectionHealthSchedule.vue'
import AdminValuationSchedule from '@/components/admin/schedules/AdminValuationSchedule.vue'
import type { AppSettings } from '@/types'
import { t } from '@/i18n'

defineProps<{
  settings: AppSettings
  settingsSaving: boolean
  availSettingsMsg: string
  availSettingsError: boolean
  auctionSettingsMsg: string
  auctionSettingsError: boolean
  alertReminderSettingsMsg: string
  alertReminderSettingsError: boolean
  watchBidDigestSettingsMsg: string
  watchBidDigestSettingsError: boolean
  healthSettingsMsg: string
  healthSettingsError: boolean
  valSettingsMsg: string
  valSettingsError: boolean
}>()

const emit = defineEmits<{
  save: []
  'update:valSettingsMsg': [val: string]
  'update:valSettingsError': [val: boolean]
  'update:auctionSettingsMsg': [val: string]
  'update:auctionSettingsError': [val: boolean]
  'update:alertReminderSettingsMsg': [val: string]
  'update:alertReminderSettingsError': [val: boolean]
  'update:watchBidDigestSettingsMsg': [val: string]
  'update:watchBidDigestSettingsError': [val: boolean]
  'update:availSettingsMsg': [val: string]
  'update:availSettingsError': [val: boolean]
  'update:healthSettingsMsg': [val: string]
  'update:healthSettingsError': [val: boolean]
}>()
</script>
