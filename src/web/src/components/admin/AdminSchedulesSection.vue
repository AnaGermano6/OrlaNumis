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
import AdminAuctionAlertReminderSchedule from '@/components/admin/schedules/AdminAuctionAlertReminderSchedule.vue'
import AdminAuctionEndingSchedule from '@/components/admin/schedules/AdminAuctionEndingSchedule.vue'
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
  'update:availSettingsMsg': [val: string]
  'update:availSettingsError': [val: boolean]
  'update:healthSettingsMsg': [val: string]
  'update:healthSettingsError': [val: boolean]
}>()
</script>
