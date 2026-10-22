<template>
  <div v-if="typeId" class="form-group min-w-0" data-testid="banknote-issue">
    <label class="form-label" :for="issues.length ? 'banknote-issue' : undefined">{{ t('banknoteForm.issue') }}</label>
    <select
      v-if="issues.length"
      id="banknote-issue"
      class="form-select"
      :value="selectedId ?? ''"
      aria-describedby="banknote-issue-hint"
      data-testid="banknote-issue-select"
      @change="choose(Number(($event.target as HTMLSelectElement).value))"
    >
      <option value="" disabled>{{ t('banknoteForm.issuePlaceholder') }}</option>
      <option v-for="issue in issues" :key="issue.id" :value="issue.id">{{ numistaIssueLabel(issue) }}</option>
    </select>
    <button v-else type="button" class="btn btn-secondary btn-sm" :disabled="loading" aria-describedby="banknote-issue-hint" data-testid="banknote-issue-load" @click="load">
      <ListChecks :size="14" aria-hidden="true" />
      {{ loading ? t('banknoteForm.issueLoading') : form.numistaIssueId ? t('banknoteForm.issueChange') : t('banknoteForm.issueLoad') }}
    </button>
    <p id="banknote-issue-hint" class="mt-1 text-chip" :class="message ? 'text-text-secondary' : 'text-text-muted'" role="status" data-testid="banknote-issue-hint">
      {{ message || t('banknoteForm.issueHint') }}
    </p>
  </div>
</template>

<script setup lang="ts">
// Lets a banknote linked to a Numista type say which of its issues it is,
// told apart by date and signatories: choosing one fills in the note's
// signatures, Pick number, printed date and kind of note. The issues come
// from the server's cache, so looking at them seldom costs a Numista request;
// they are only loaded when asked for.
import { computed, ref, watch } from 'vue'
import { ListChecks } from 'lucide-vue-next'
import { getNumistaTypeIssues, type NumistaIssueSummary } from '@/api/client'
import { t } from '@/i18n'
import type { Coin } from '@/types'
import { applyNumistaIssue, numistaTypeIdOf } from '@/utils/banknoteIssue'
import { numistaIssueLabel } from '@/utils/numistaCatalog'

const props = defineProps<{ form: Partial<Coin> }>()

const typeId = computed(() => numistaTypeIdOf(props.form.references))
const issues = ref<NumistaIssueSummary[]>([])
const selectedId = ref<number | undefined>(undefined)
const loading = ref(false)
const message = ref('')

// Another type's issues are not this note's.
watch(typeId, () => {
  issues.value = []
  selectedId.value = undefined
  message.value = ''
})

async function load() {
  const id = typeId.value
  if (!id || loading.value) return
  loading.value = true
  message.value = ''
  try {
    const { data } = await getNumistaTypeIssues(id)
    if (id !== typeId.value) return
    if (data.status === 'ok' && data.issues.length) {
      issues.value = data.issues
      selectedId.value = data.issues.some((issue) => issue.id === props.form.numistaIssueId) ? props.form.numistaIssueId : undefined
    } else if (data.status === 'ok' || data.status === 'no_issues') {
      message.value = t('banknoteForm.issueNone')
    } else if (data.status === 'quota_limited') {
      message.value = t('coinDetail.marketValue.quotaLimited')
    } else {
      message.value = t('banknoteForm.issueFailed')
    }
  } catch {
    message.value = t('banknoteForm.issueFailed')
  } finally {
    loading.value = false
  }
}

function choose(id: number) {
  const issue = issues.value.find((item) => item.id === id)
  if (!issue) return
  selectedId.value = id
  applyNumistaIssue(props.form, issue)
  message.value = t('banknoteForm.issueApplied')
}
</script>
