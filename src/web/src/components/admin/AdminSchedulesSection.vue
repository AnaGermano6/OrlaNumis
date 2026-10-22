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

    <AdminValuationSchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      :settings-msg="valSettingsMsg"
      :settings-error="valSettingsError"
      @save="emit('save')"
      @update:settings-msg="emit('update:valSettingsMsg', $event)"
      @update:settings-error="emit('update:valSettingsError', $event)"
    />

    <AdminNumistaValuesSchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      @save="emit('save')"
    />

    <AdminBackupSchedule
      :settings="settings"
      :settings-saving="settingsSaving"
      @save="emit('save')"
    />
  </section>
</template>

<script setup lang="ts">
import AdminAvailabilitySchedule from '@/components/admin/schedules/AdminAvailabilitySchedule.vue'
import AdminWishlistSearchAlertSchedule from '@/components/admin/schedules/AdminWishlistSearchAlertSchedule.vue'
import AdminBackupSchedule from '@/components/admin/schedules/AdminBackupSchedule.vue'
import AdminScheduleTimezone from '@/components/admin/schedules/AdminScheduleTimezone.vue'
import AdminValuationSchedule from '@/components/admin/schedules/AdminValuationSchedule.vue'
import AdminNumistaValuesSchedule from '@/components/admin/schedules/AdminNumistaValuesSchedule.vue'
import type { AppSettings } from '@/types'
import { t } from '@/i18n'

defineProps<{
  settings: AppSettings
  settingsSaving: boolean
  availSettingsMsg: string
  availSettingsError: boolean
  valSettingsMsg: string
  valSettingsError: boolean
}>()

const emit = defineEmits<{
  save: []
  'update:valSettingsMsg': [val: string]
  'update:valSettingsError': [val: boolean]
  'update:availSettingsMsg': [val: string]
  'update:availSettingsError': [val: boolean]
}>()
</script>
