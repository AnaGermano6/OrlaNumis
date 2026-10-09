<template>
  <CoinDetailSectionPageShell :section-title="t('coinDetail.sections.journal.title')">
    <template #default="{ coin }">
      <CoinActivityJournal
        :entries="journalEntries"
        :coin-id="coin.id"
        @add="handleAddJournalEntry"
        @delete="handleDeleteJournalEntry"
      />
    </template>
  </CoinDetailSectionPageShell>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import CoinDetailSectionPageShell from '@/components/coin/CoinDetailSectionPageShell.vue'
import CoinActivityJournal from '@/components/coin/CoinActivityJournal.vue'
import { getJournalEntries, addJournalEntry, deleteJournalEntry } from '@/api/client'
import { useDialog } from '@/composables/useDialog'
import type { CoinJournal } from '@/types'
import { t } from '@/i18n'

const { showAlert } = useDialog()
const route = useRoute()

const journalEntries = ref<CoinJournal[]>([])

onMounted(() => {
  loadJournal()
})

async function loadJournal() {
  const coinId = Number(route.params.id)
  try {
    const res = await getJournalEntries(coinId)
    journalEntries.value = res.data || []
  } catch {
    journalEntries.value = []
  }
}

async function handleAddJournalEntry(entry: string) {
  const coinId = Number(route.params.id)
  if (!entry) return
  try {
    await addJournalEntry(coinId, entry)
    loadJournal()
  } catch {
    await showAlert(t('coinDetail.journal.addFailed'), { title: t('common.error') })
  }
}

async function handleDeleteJournalEntry(entryId: number) {
  const coinId = Number(route.params.id)
  try {
    await deleteJournalEntry(coinId, entryId)
    loadJournal()
  } catch {
    await showAlert(t('coinDetail.journal.deleteFailed'), { title: t('common.error') })
  }
}
</script>
