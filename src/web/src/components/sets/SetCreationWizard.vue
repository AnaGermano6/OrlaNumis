<template>
  <div class="w-full">
    <form @submit.prevent="submit">
      <div class="form-group mb-4">
        <label for="setType" class="form-label mb-2 block">{{ t('sets.wizard.setType') }}</label>
        <select id="setType" v-model="form.setType" class="form-input w-full">
          <option value="standard">{{ t('sets.types.standard') }}</option>
          <option value="goal">{{ t('sets.types.goal') }}</option>
          <option value="smart">{{ t('sets.types.smart') }}</option>
          <option value="agentic">{{ t('sets.types.agentic') }}</option>
        </select>
      </div>
      <div v-if="form.setType === 'goal'" class="form-group mb-4 rounded-sm border border-border-subtle bg-card p-4">
        <span class="section-label">{{ t('sets.wizard.goalTitle') }}</span>
        <p class="mt-2 mb-0 text-body text-text-secondary">
          {{ t('sets.wizard.goalP1') }}
        </p>
        <p class="mt-2 mb-0 text-body text-text-secondary">
          {{ t('sets.wizard.goalP2') }}
        </p>
        <p class="mt-2 mb-0 text-chip text-text-muted">
          {{ t('sets.wizard.goalP3') }}
        </p>
      </div>
      <SetSmartRuleBuilder
        v-if="form.setType === 'smart'"
        @update="form.smartCriteria = $event"
      />
      <div v-if="form.setType === 'agentic'" class="form-group mb-4 rounded-sm border border-border-subtle bg-card p-4">
        <span class="section-label">{{ t('sets.wizard.agenticTitle') }}</span>
        <p class="mt-2 mb-0 text-body text-text-secondary">
          {{ t('sets.wizard.agenticP1') }}
        </p>
        <p class="mt-2 mb-0 text-body text-text-secondary">
          {{ t('sets.wizard.agenticP2') }}
        </p>
        <p class="mt-2 mb-0 text-chip text-text-muted">
          {{ t('sets.wizard.agenticP3') }}
        </p>
      </div>
      <div v-if="form.setType === 'agentic'" class="form-group mb-4">
        <label for="agenticPrompt" class="form-label mb-2 block">{{ t('sets.wizard.agenticPrompt') }}</label>
        <textarea
          id="agenticPrompt"
          v-model="form.agenticPrompt"
          rows="3"
          maxlength="500"
          class="form-input w-full"
          :placeholder="t('sets.wizard.agenticPlaceholder')"
          :required="form.setType === 'agentic'"
        />
      </div>
      <div class="form-group mb-4">
        <label for="setName" class="form-label mb-2 block">{{ t('common.name') }}</label>
        <input
          id="setName"
          v-model="form.name"
          type="text"
          required
          maxlength="80"
          class="form-input w-full"
        />
      </div>
      <div class="form-group mb-4">
        <label for="setDescription" class="form-label mb-2 block">{{ t('common.description') }}</label>
        <textarea
          id="setDescription"
          v-model="form.description"
          rows="3"
          maxlength="2000"
          class="form-input w-full"
        />
      </div>
      <div class="form-group mb-4">
        <label for="setColor" class="form-label mb-2 block">{{ t('common.color') }}</label>
        <input
          id="setColor"
          v-model="form.color"
          type="color"
          class="h-10 w-full rounded-sm border border-border-subtle bg-input p-1 text-text-primary"
        />
      </div>
      <div class="mt-6 flex justify-end gap-2">
        <button type="button" class="btn btn-secondary" @click="$emit('cancel')">{{ t('common.cancel') }}</button>
        <button type="submit" class="btn btn-primary" :disabled="!form.name.trim()">
          {{ submitLabel || t('common.create') }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'
import { normalizeCoinSetType } from '@/types'
import type { CreateCoinSetRequest, SmartCriteriaGroup } from '@/types'
import SetSmartRuleBuilder from '@/components/sets/SetSmartRuleBuilder.vue'
import { randomSetColor } from '@/utils/setColors'
import { t } from '@/i18n'

const props = withDefaults(defineProps<{
  initialValue?: Partial<CreateCoinSetRequest>
  submitLabel?: string
}>(), {
  submitLabel: '',
})

const emit = defineEmits<{
  submit: [value: CreateCoinSetRequest, csv?: string]
  cancel: []
}>()

const form = reactive({
  name: props.initialValue?.name ?? '',
  description: props.initialValue?.description ?? '',
  color: props.initialValue?.color ?? randomSetColor(),
  setType: props.initialValue?.setType ? normalizeCoinSetType(props.initialValue.setType) : 'standard',
  templateId: props.initialValue?.templateId ?? '',
  targetCompletionDate: props.initialValue?.targetCompletionDate ?? '',
  smartCriteria: props.initialValue?.smartCriteria as SmartCriteriaGroup | undefined,
  agenticPrompt: props.initialValue?.agenticPrompt ?? '',
})

watch(() => props.initialValue, (value) => {
  form.name = value?.name ?? ''
  form.description = value?.description ?? ''
  form.color = value?.color ?? randomSetColor()
  form.setType = value?.setType ? normalizeCoinSetType(value.setType) : 'standard'
  form.templateId = value?.templateId ?? ''
  form.targetCompletionDate = value?.targetCompletionDate ?? ''
  form.smartCriteria = value?.smartCriteria as SmartCriteriaGroup | undefined
  form.agenticPrompt = value?.agenticPrompt ?? ''
})

function submit() {
  const name = form.name.trim()
  if (!name) return
  emit('submit', {
    name,
    description: form.description.trim(),
    color: form.color,
    setType: form.setType,
    templateId: form.setType !== 'goal' ? (form.templateId || undefined) : undefined,
    targetCompletionDate: form.setType !== 'goal' ? (form.targetCompletionDate || undefined) : undefined,
    smartCriteria: form.smartCriteria ?? undefined,
    agenticPrompt: form.setType === 'agentic' ? form.agenticPrompt.trim() || undefined : undefined,
  }, undefined)
}
</script>
