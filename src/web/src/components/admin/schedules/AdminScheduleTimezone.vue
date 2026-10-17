<template>
  <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.schedules.tzTitle') }}</h3>
  <p class="mb-4 text-base text-text-secondary">{{ t('admin.schedules.tzIntro') }}</p>
  <div class="mb-4">
    <div class="form-group">
      <label class="form-label" for="schedule-timezone">{{ t('admin.schedules.tz') }}</label>
      <select
        id="schedule-timezone"
        class="form-input w-full max-w-[320px]"
        :value="settings.ScheduleTimezone ?? ''"
        @change="settings.ScheduleTimezone = ($event.target as HTMLSelectElement).value"
      >
        <option value="">{{ t('admin.schedules.tzServer') }}</option>
        <option v-for="zone in timezoneOptions" :key="zone" :value="zone">{{ zone }}</option>
      </select>
      <span class="form-hint">{{ t('admin.schedules.tzApply') }}</span>
      <button
        v-if="browserTimezone && settings.ScheduleTimezone !== browserTimezone"
        type="button"
        class="mt-1 self-start bg-transparent p-0 text-sm text-gold underline"
        @click="settings.ScheduleTimezone = browserTimezone"
      >
        {{ t('admin.schedules.tzUseMine', { zone: browserTimezone }) }}
      </button>
    </div>
    <div class="mt-4 flex w-full flex-col gap-3 md:flex-row md:items-center">
      <button class="btn btn-primary btn-sm" :disabled="settingsSaving" @click="emit('save')">
        {{ settingsSaving ? t('common.saving') : t('admin.schedules.tzSave') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { getBrowserTimezone, timezoneOptionsWith } from '@/composables/useScheduleTimezone'
import type { AppSettings } from '@/types'
import { t } from '@/i18n'

const props = defineProps<{
  settings: AppSettings
  settingsSaving: boolean
}>()

const emit = defineEmits<{
  save: []
}>()

const browserTimezone = getBrowserTimezone()
const timezoneOptions = computed(() => timezoneOptionsWith(props.settings.ScheduleTimezone))
</script>
