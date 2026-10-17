<template>
  <hr class="my-6 border-0 border-t border-border-subtle" />

  <!-- Automatic full backup -->
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.backup.title') }}</h3>
  <p class="mb-4 text-base text-text-secondary">{{ t('admin.schedules.backup.intro') }}</p>
  <div class="mb-4" data-testid="backup-schedule">
    <div class="form-group flex items-center justify-between gap-3">
      <label class="form-label" for="backup-enabled">{{ t('admin.schedules.backup.enable') }}</label>
      <BaseToggle
        size="sm"
        id="backup-enabled"
        :label="t('admin.schedules.backup.enable')"
        :model-value="settings.BackupEnabled !== 'false'"
        @update:model-value="settings.BackupEnabled = $event ? 'true' : 'false'"
      />
    </div>
    <div class="form-group">
      <label class="form-label" for="backup-start-time">{{ t('admin.schedules.startDaily') }}</label>
      <input
        id="backup-start-time"
        v-model="settings.BackupStartTime"
        class="form-input w-full max-w-[120px]"
        type="time"
        aria-describedby="backup-start-time-hint"
      />
      <span id="backup-start-time-hint" class="form-hint">{{ t('admin.schedules.backup.startHint') }}</span>
      <AdminScheduleSummary :start-time="settings.BackupStartTime" default-start-time="03:00" :zone="settings.ScheduleTimezone" />
    </div>
    <div class="form-group">
      <label class="form-label" for="backup-keep">{{ t('admin.schedules.backup.keep') }}</label>
      <input
        id="backup-keep"
        v-model="settings.BackupKeep"
        class="form-input w-full max-w-[120px]"
        type="number"
        min="1"
        max="365"
        aria-describedby="backup-keep-hint"
      />
      <span id="backup-keep-hint" class="form-hint">{{ t('admin.schedules.backup.keepHint') }}</span>
    </div>
    <div class="mt-4 flex w-full flex-col gap-3 md:flex-row md:items-center">
      <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="emit('save')">
        {{ settingsSaving ? t('common.saving') : t('admin.schedules.backup.save') }}
      </button>
      <RouterLink :to="{ path: '/settings', query: { tab: 'backups' } }" class="text-body text-gold underline-offset-2 hover:underline">{{ t('admin.schedules.backup.open') }}</RouterLink>
    </div>
  </div>
</template>

<script setup lang="ts">
// The automatic full backup's schedule: on, its daily time, how many are kept.
import { RouterLink } from 'vue-router'
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
