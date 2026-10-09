<template>
  <hr class="my-6 border-0 border-t border-border-subtle" />

  <!-- Purchase Reminder Delivery -->
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.reminder.title') }}</h3>
  <p class="mb-4 text-base text-text-secondary">{{ t('admin.schedules.reminder.intro') }}</p>
  <div class="mb-4">
    <div class="form-group flex items-center justify-between gap-3">
      <label class="form-label" for="reminder-check-enabled">{{ t('admin.schedules.reminder.enable') }}</label>
      <BaseToggle
        size="sm"
        id="reminder-check-enabled"
        :label="t('admin.schedules.reminder.enable')"
        :model-value="settings.ReminderCheckEnabled === 'true'"
        @update:model-value="settings.ReminderCheckEnabled = $event ? 'true' : 'false'"
      />
    </div>
    <div class="form-group">
      <label class="form-label" for="reminder-check-start-time">{{ t('admin.schedules.startDaily') }}</label>
      <input
        id="reminder-check-start-time"
        v-model="settings.ReminderCheckStartTime"
        class="form-input w-full max-w-[120px]"
        type="time"
        aria-describedby="reminder-check-start-time-hint"
      />
      <span id="reminder-check-start-time-hint" class="form-hint">{{ t('admin.schedules.reminder.startHint') }}</span>
      <AdminScheduleSummary :start-time="settings.ReminderCheckStartTime" default-start-time="08:00" :zone="settings.ScheduleTimezone" />
    </div>
    <div class="mt-4 flex w-full flex-col gap-3 md:flex-row md:items-center">
      <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="emit('save')">
        {{ settingsSaving ? t('common.saving') : t('admin.schedules.reminder.save') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { BaseToggle } from '@/components/ui'
import AdminScheduleSummary from '@/components/admin/schedules/AdminScheduleSummary.vue'
import type { AppSettings } from '@/types'
import { t } from '@/i18n'

defineProps<{
  settings: AppSettings
  settingsSaving: boolean
}>()

const emit = defineEmits<{
  save: []
}>()
</script>