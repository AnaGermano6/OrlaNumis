<template>
  <section class="card">
    <h2 class="text-lg font-medium mb-5 pb-3 border-b border-border-subtle">{{ t('settings.tabs.appearance') }}</h2>

    <!-- Language -->
    <div class="flex justify-between items-center py-3 border-b border-border-subtle gap-4 md:flex-row flex-col md:items-center items-stretch">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('language.label') }}</span>
        <span class="text-sm text-text-muted">{{ t('settings.appearance.languageHint') }}</span>
      </div>
      <LanguageSwitcher class="self-start md:self-auto" />
    </div>

    <!-- Currency of every page -->
    <SettingsCurrencySetting />

    <!-- Theme -->
    <div class="flex justify-between items-center py-3 border-b border-border-subtle gap-4 last:border-0 md:flex-row flex-col md:items-center items-stretch">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('settings.appearance.theme') }}</span>
        <span class="text-sm text-text-muted">{{ t('settings.appearance.themeHint') }}</span>
      </div>
      <div class="flex flex-wrap gap-1 bg-surface rounded-[var(--radius-md)] p-[0.2rem]">
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="theme === 'dark' ? 'bg-gold-dim text-gold' : 'bg-transparent text-text-secondary'"
          @click="$emit('set-theme', 'dark')"
        >
          {{ t('settings.appearance.themes.dark') }}
        </button>
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="theme === 'light' ? 'bg-gold-dim text-gold' : 'bg-transparent text-text-secondary'"
          @click="$emit('set-theme', 'light')"
        >
          {{ t('settings.appearance.themes.light') }}
        </button>
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="theme === 'british-museum' ? 'bg-gold-dim text-gold' : 'bg-transparent text-text-secondary'"
          @click="$emit('set-theme', 'british-museum')"
        >
          {{ t('settings.appearance.themes.britishMuseum') }}
        </button>
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="theme === 'louvre' ? 'bg-gold-dim text-gold' : 'bg-transparent text-text-secondary'"
          @click="$emit('set-theme', 'louvre')"
        >
          {{ t('settings.appearance.themes.louvre') }}
        </button>
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="theme === 'capitoline' ? 'bg-gold-dim text-gold' : 'bg-transparent text-text-secondary'"
          @click="$emit('set-theme', 'capitoline')"
        >
          {{ t('settings.appearance.themes.capitoline') }}
        </button>
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="theme === 'byzantine' ? 'bg-gold-dim text-gold' : 'bg-transparent text-text-secondary'"
          @click="$emit('set-theme', 'byzantine')"
        >
          {{ t('settings.appearance.themes.byzantine') }}
        </button>
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="theme === 'modern-greek' ? 'bg-gold-dim text-gold' : 'bg-transparent text-text-secondary'"
          @click="$emit('set-theme', 'modern-greek')"
        >
          {{ t('settings.appearance.themes.modernGreek') }}
        </button>
      </div>
    </div>

    <!-- Timezone -->
    <div class="flex justify-between items-center py-3 border-b border-border-subtle gap-4 md:flex-row flex-col md:items-center items-stretch">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('settings.appearance.timezone') }}</span>
        <span class="text-sm text-text-muted">{{ t('settings.appearance.timezoneHint') }}</span>
      </div>
      <select
        :value="timezone"
        class="form-select max-w-[250px]"
        @change="$emit('save-timezone', ($event.target as HTMLSelectElement).value)"
      >
        <option v-for="tz in timezones" :key="tz" :value="tz">{{ tz }}</option>
      </select>
    </div>

    <!-- Default View -->
    <div class="flex justify-between items-center py-3 border-b border-border-subtle gap-4 md:flex-row flex-col md:items-center items-stretch">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('settings.appearance.defaultView') }}</span>
        <span class="text-sm text-text-muted">{{ t('settings.appearance.defaultViewHint') }}</span>
      </div>
      <div class="flex gap-1 bg-surface rounded-full p-[0.2rem]">
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="defaultView === 'swipe' ? 'bg-gold-dim text-gold' : 'bg-transparent text-text-secondary'"
          @click="$emit('set-default-view', 'swipe')"
        >
          {{ t('settings.appearance.swipe') }}
        </button>
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="defaultView === 'grid' ? 'bg-gold-dim text-gold' : 'bg-transparent text-text-secondary'"
          @click="$emit('set-default-view', 'grid')"
        >
          {{ t('settings.appearance.grid') }}
        </button>
      </div>
    </div>

    <!-- Tray Felt Color -->
    <div class="flex justify-between items-center py-3 border-b border-border-subtle gap-4 md:flex-row flex-col md:items-center items-stretch">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('settings.appearance.felt') }}</span>
        <span class="text-sm text-text-muted">{{ t('settings.appearance.feltHint') }}</span>
      </div>
      <div class="flex flex-wrap gap-1 bg-surface rounded-full p-[0.2rem]">
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="trayFeltColor === 'red'
            ? 'bg-[var(--felt-red-dim)] text-felt-red-bright'
            : 'bg-transparent text-text-secondary'"
          @click="$emit('set-tray-felt-color', 'red')"
        >
          {{ t('settings.appearance.felts.red') }}
        </button>
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="trayFeltColor === 'green'
            ? 'bg-[var(--felt-green-dim)] text-felt-green-bright'
            : 'bg-transparent text-text-secondary'"
          @click="$emit('set-tray-felt-color', 'green')"
        >
          {{ t('settings.appearance.felts.green') }}
        </button>
        <button
          class="px-3 py-[0.35rem] rounded-full text-chip cursor-pointer transition-all duration-200 border-0 focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2"
          :class="trayFeltColor === 'navy'
            ? 'bg-[var(--felt-navy-dim)] text-felt-navy-bright'
            : 'bg-transparent text-text-secondary'"
          @click="$emit('set-tray-felt-color', 'navy')"
        >
          {{ t('settings.appearance.felts.navy') }}
        </button>
      </div>
    </div>

    <!-- Default Sort -->
    <div class="flex justify-between items-center py-3 gap-4 md:flex-row flex-col md:items-center items-stretch">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('settings.appearance.defaultSort') }}</span>
        <span class="text-sm text-text-muted">{{ t('settings.appearance.defaultSortHint') }}</span>
      </div>
      <select
        :value="defaultSort"
        class="form-select max-w-[250px]"
        @change="$emit('save-default-sort', ($event.target as HTMLSelectElement).value)"
      >
        <option value="updated_at_desc">{{ t('collection.sort.lastUpdated') }}</option>
        <option value="created_at_desc">{{ t('collection.sort.newest') }}</option>
        <option value="created_at_asc">{{ t('collection.sort.oldest') }}</option>
        <option value="current_value_desc">{{ t('collection.sort.priceHighLow') }}</option>
        <option value="current_value_asc">{{ t('collection.sort.priceLowHigh') }}</option>
        <option value="random_desc">{{ t('collection.sort.random') }}</option>
      </select>
    </div>
    <!-- Swipe Navigation -->
    <div class="flex justify-between items-center py-3 gap-4 md:flex-row flex-col md:items-center items-stretch">
      <div class="flex flex-col gap-[0.15rem]">
        <span class="text-base font-medium">{{ t('settings.appearance.swipeNav') }}</span>
        <span class="text-sm text-text-muted">{{ t('settings.appearance.swipeNavHint') }}</span>
        <span v-if="pwaSwipeNavError" class="mt-1 text-chip text-loss">{{ pwaSwipeNavError }}</span>
      </div>
      <BaseToggle
        :label="t('settings.appearance.swipeNavLabel')"
        :model-value="pwaSwipeNavLocal"
        :disabled="pwaSwipeNavSaving"
        @update:model-value="handlePwaSwipeNavChange"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { BaseToggle } from '@/components/ui'
import { ref } from 'vue'
import type { Theme } from '@/types'
import type { FeltColor } from '@/composables/useTrayPreference'
import { useSettingsProfile } from '@/composables/useSettingsProfile'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import SettingsCurrencySetting from '@/components/settings/SettingsCurrencySetting.vue'
import { t } from '@/i18n'

defineProps<{
  theme: Theme
  timezone: string
  timezones: string[]
  defaultView: 'grid' | 'swipe'
  defaultSort: string
  trayFeltColor: FeltColor
}>()

defineEmits<{
  'set-theme': [theme: Theme]
  'save-timezone': [tz: string]
  'set-default-view': [view: 'grid' | 'swipe']
  'save-default-sort': [sort: string]
  'set-tray-felt-color': [color: FeltColor]
}>()

// savePwaSwipeNav handles auth-store sync, localStorage write, optimistic update, and rollback.
const { pwaSwipeNavEnabled: pwaSwipeNavLocal, savePwaSwipeNav } = useSettingsProfile()
const pwaSwipeNavError = ref('')
const pwaSwipeNavSaving = ref(false)

async function handlePwaSwipeNavChange(value: boolean) {
  pwaSwipeNavError.value = ''
  pwaSwipeNavSaving.value = true
  try {
    await savePwaSwipeNav(value)
  } catch {
    pwaSwipeNavError.value = t('settings.appearance.swipeNavFailed')
  } finally {
    pwaSwipeNavSaving.value = false
  }
}
</script>