<template>
  <div class="mt-4 border-t border-border-subtle pt-3">
    <h3 :id="headingId" class="m-0 text-base font-medium text-heading">{{ t('admin.properties.portugueseNames') }}</h3>
    <p class="mt-1 text-body text-text-secondary">{{ t('admin.properties.portugueseNamesHint') }}</p>
    <div class="mt-3 grid gap-2" role="group" :aria-labelledby="headingId">
      <label
        v-for="option in options"
        :key="option"
        class="grid grid-cols-[minmax(0,2fr)_minmax(0,3fr)] items-center gap-3"
      >
        <span class="text-body text-text-primary [overflow-wrap:anywhere]">{{ option }}</span>
        <input
          class="form-input"
          type="text"
          maxlength="80"
          :value="names[option] ?? ''"
          :placeholder="builtInCoinOptionLabel(option) ?? option"
          :aria-label="t('admin.properties.portugueseNameFor', { value: option })"
          @input="setName(option, ($event.target as HTMLInputElement).value)"
        />
      </label>
    </div>
  </div>
</template>

<script setup lang="ts">
// Per-option Portuguese display names for a category or era list, stored as
// "Value=Name" lines. Names typed for an option stay in memory while the admin
// edits the list above, but only options still in the list are saved.
import { ref, watch } from 'vue'
import { t } from '@/i18n'
import { builtInCoinOptionLabel } from '@/i18n/coinVocabulary'
import { formatOptionLabelMap, parseOptionLabelMap } from '@/utils/options'

const props = defineProps<{
  options: string[]
  modelValue: string
  headingId: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const names = ref<Record<string, string>>(parseOptionLabelMap(props.modelValue))

function serialize(): string {
  return formatOptionLabelMap(props.options, names.value)
}

function publish() {
  const next = serialize()
  if (next !== props.modelValue) emit('update:modelValue', next)
}

watch(() => props.modelValue, (value) => {
  if (value !== serialize()) names.value = parseOptionLabelMap(value)
})
watch(() => props.options, publish)

function setName(option: string, value: string) {
  names.value = { ...names.value, [option]: value }
  publish()
}
</script>
