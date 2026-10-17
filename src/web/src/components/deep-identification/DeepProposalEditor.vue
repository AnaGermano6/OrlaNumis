<template>
  <section class="grid min-w-0 gap-4 overflow-hidden" :aria-label="t('deepAnalysis.proposal.label')">
    <h3 class="m-0 text-lg font-semibold text-text-primary">{{ t('deepAnalysis.proposal.title') }}</h3>
    <p class="m-0 text-sm text-text-secondary">
      {{ t('deepAnalysis.proposal.intro') }}
    </p>
    <ul class="m-0 grid gap-3 p-0" style="list-style: none;">
      <li
        v-for="name in fieldNames"
        :key="name"
        class="grid min-w-0 gap-2 overflow-hidden rounded-sm border border-border-subtle bg-card p-3"
      >
        <div class="flex flex-wrap items-baseline justify-between gap-2">
          <span class="text-label font-semibold uppercase tracking-[0.08em] text-text-primary">{{ fieldLabel(name) }}</span>
          <div class="flex flex-wrap items-center gap-2">
            <span
              v-if="isImageDerived(name)"
              class="chip-sm"
              :title="t('deepAnalysis.proposal.imageOnlyTitle')"
            >
              {{ t('deepAnalysis.proposal.imageOnly') }}
            </span>
            <span v-if="entryOf(name).ownerEdited" class="text-label font-semibold uppercase tracking-[0.08em] text-byzantine">
              {{ t('deepAnalysis.proposal.editedByYou') }}
            </span>
            <span v-else class="text-label font-semibold uppercase tracking-[0.08em] text-gold">{{ t('deepAnalysis.proposal.aiProposed') }}</span>
          </div>
        </div>

        <p class="m-0 break-words text-sm text-text-secondary [overflow-wrap:anywhere]">
          {{ t('deepAnalysis.proposal.aiValue') }} <span class="font-medium text-text-primary">{{ displayValue(entryOf(name).proposed) }}</span>
        </p>

        <label class="grid gap-1 text-sm text-text-secondary" :for="`deep-proposal-field-${name}`">
          {{ t('deepAnalysis.proposal.yourValue') }}
          <textarea
            v-if="name === 'notes'"
            :id="`deep-proposal-field-${name}`"
            class="min-h-[132px] min-w-0 resize-y rounded-sm border border-border-subtle bg-input px-2 py-1 text-text-primary"
            :value="ownerValue(name)"
            @input="onOwnerValueInput(name, $event)"
          ></textarea>
          <input
            v-else
            :id="`deep-proposal-field-${name}`"
            type="text"
            class="min-h-[44px] rounded-sm border border-border-subtle bg-input px-2 py-1 text-text-primary"
            :value="ownerValue(name)"
            @input="onOwnerValueInput(name, $event)"
          >
        </label>

        <div class="flex items-center gap-2" role="group" :aria-label="t('deepAnalysis.proposal.decision', { field: fieldLabel(name) })">
          <button
            type="button"
            class="chip min-h-[44px]"
            :class="isAccepted(name) ? 'border-gold bg-gold text-white' : 'border-border-subtle text-text-secondary'"
            :aria-pressed="isAccepted(name)"
            @click="setAccepted(name, true)"
          >
            {{ t('common.accept') }}
          </button>
          <button
            type="button"
            class="chip min-h-[44px]"
            :class="isRejected(name) ? 'border-byzantine bg-byzantine text-white' : 'border-border-subtle text-text-secondary'"
            :aria-pressed="isRejected(name)"
            @click="setAccepted(name, false)"
          >
            {{ t('common.reject') }}
          </button>
        </div>
      </li>
    </ul>

    <button
      type="button"
      class="btn btn-primary min-h-[44px] justify-self-start"
      :disabled="confirmDisabled"
      @click="$emit('confirm')"
    >
      {{ applying ? applyingLabel : actionLabel }}
    </button>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { DeepProposal, DeepProposalFieldEntry } from '@/types'
import { formatFieldName } from '@/utils/format'
import { effectiveDeepProposalAcceptance } from '@/utils/deepProposalAcceptance'
import { t } from '@/i18n'

const props = defineProps<{
  proposal: DeepProposal
  applying?: boolean
  actionLabel?: string
  applyingLabel?: string
}>()

const emit = defineEmits<{
  (e: 'update-field', name: string, edit: { ownerValue?: unknown; accepted?: boolean | null }): void
  (e: 'confirm'): void
}>()

const emptyEntry: DeepProposalFieldEntry = { proposed: null, ownerEdited: false, ownerValue: null, accepted: null }

const fields = computed(() => props.proposal.fields)
const fieldNames = computed(() => Object.keys(props.proposal.fields).sort())
const actionLabel = computed(() => props.actionLabel ?? t('deepAnalysis.proposal.applyToCoin'))
const applyingLabel = computed(() => props.applyingLabel ?? t('deepAnalysis.proposal.applying'))

function entryOf(name: string): DeepProposalFieldEntry {
  return fields.value[name] ?? emptyEntry
}

const confirmDisabled = computed(() => {
  if (props.applying) return true
  return !fieldNames.value.some((name) => isAccepted(name))
})

function fieldLabel(name: string): string {
  return formatFieldName(name)
}

/**
 * A field whose evidence array is present but empty is supported only by
 * the vision hypothesis — no provider has corroborated it (data-model.md
 * §6 / FR-026). Marked visibly distinct from provider-cited fields.
 */
function isImageDerived(name: string): boolean {
  const evidence = entryOf(name).evidence
  return Array.isArray(evidence) && evidence.length === 0
}

/**
 * RD-3: confidence-driven, not source-driven. An explicit owner decision
 * always wins; otherwise the field defaults to accepted once confidence
 * crosses DEEP_PROPOSAL_ACCEPTANCE_THRESHOLD, regardless of source.
 */
function isAccepted(name: string): boolean {
  return effectiveDeepProposalAcceptance(entryOf(name)) === true
}

function isRejected(name: string): boolean {
  return effectiveDeepProposalAcceptance(entryOf(name)) === false
}

function displayValue(value: unknown): string {
  if (value === null || value === undefined) return '—'
  return String(value)
}

function ownerValue(name: string): string {
  const entry = entryOf(name)
  const value = entry.ownerEdited ? entry.ownerValue : entry.proposed
  return value === null || value === undefined ? '' : String(value)
}

function onOwnerValueInput(name: string, event: Event) {
  const target = event.target as HTMLInputElement
  emit('update-field', name, { ownerValue: target.value })
}

function setAccepted(name: string, accepted: boolean) {
  emit('update-field', name, { accepted })
}

</script>
