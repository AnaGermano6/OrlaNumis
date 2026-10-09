<template>
  <div class="container pb-6">
    <DesktopPageContext :title="t('notes.titleCount', { count: notes.length })">
      <template #actions>
        <AppIconButton v-if="pwaMode !== 'list'" :title="t('notes.back')" @click="showPwaList">
          <ArrowLeft :size="20" />
        </AppIconButton>
        <AppIconButton :title="t('notes.new')" @click="startNewNote">
          <CirclePlus :size="20" />
        </AppIconButton>
      </template>
      <div class="page-header">
        <h1 class="inline-flex items-center gap-2">
          {{ t('nav.notes') }}
          <span class="text-base font-medium text-text-secondary">({{ notes.length }})</span>
        </h1>
        <div v-if="isPwa" class="pwa-actions">
          <button v-if="pwaMode !== 'list'" class="pwa-icon-btn" :title="t('notes.back')" :aria-label="t('notes.back')" @click="showPwaList">
            <ArrowLeft :size="22" />
          </button>
          <button class="pwa-icon-btn" :title="t('notes.new')" @click="startNewNote">
            <CirclePlus :size="22" />
          </button>
        </div>
        <div v-else class="header-actions">
          <button class="btn btn-primary" @click="startNewNote">
            <Plus :size="16" /> {{ t('notes.new') }}
          </button>
        </div>
      </div>
    </DesktopPageContext>

    <div v-if="loading" class="loading-overlay">
      <div class="spinner" />
      <p>{{ t('notes.loading') }}</p>
    </div>

    <div v-else-if="loadError" class="empty-state flex flex-col items-center gap-3 rounded-md border border-border-subtle bg-card text-text-secondary shadow-[var(--shadow-card)]">
      <AlertCircle :size="42" />
      <h3>{{ t('notes.loadErrorTitle') }}</h3>
      <p>{{ loadError }}</p>
      <button class="btn btn-secondary btn-sm" @click="loadNotes">{{ t('common.tryAgain') }}</button>
    </div>

    <div v-else class="grid items-start gap-4" :class="isPwa ? 'grid-cols-1' : 'md:grid-cols-[minmax(240px,0.85fr)_minmax(0,2fr)]'">
      <aside
        v-if="!isPwa || pwaMode === 'list'"
        class="flex flex-col gap-2"
        :class="isPwa ? 'min-h-[calc(100vh-15rem)]' : 'max-h-[280px] overflow-y-auto md:max-h-none md:min-h-[460px]'"
        :aria-label="t('notes.list')"
      >
        <div class="card px-4 py-3">
          <span class="section-label m-0">{{ t('notes.all') }}</span>
        </div>

        <div v-if="notes.length === 0 && !editMode" class="flex min-h-[320px] flex-col items-center justify-center gap-3 text-center text-text-muted">
          <StickyNote :size="36" />
          <h3 class="m-0 text-lg text-text-primary">{{ t('notes.empty') }}</h3>
          <p class="max-w-[34rem] text-text-secondary">{{ t('notes.emptyHint') }}</p>
          <button class="btn btn-primary btn-sm" @click="startNewNote">{{ t('chat.note.create') }}</button>
        </div>

        <button
          v-for="note in notes"
          :key="note.id"
          class="card !p-0 flex w-full items-start gap-3 border-l-[3px] border-l-transparent !px-4 !py-[0.85rem] text-left transition-colors hover:bg-card-hover"
          :class="selectedId === note.id ? 'border-l-gold bg-gold/4' : ''"
          @click="selectNote(note.id)"
        >
          <div class="mt-[2px] shrink-0 text-text-muted">
            <StickyNote :size="18" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="mb-[0.2rem] flex items-start justify-between gap-2">
              <span class="truncate text-base font-semibold text-text-primary">{{ note.title || t('notes.untitled') }}</span>
              <span v-if="note.body" class="chip-sm shrink-0">{{ t('chat.note.markdown') }}</span>
            </div>
            <div class="truncate text-body leading-[1.4] text-text-secondary">{{ previewText(note.body) }}</div>
            <div class="mt-[0.35rem] text-sm text-text-muted">{{ t('notes.updated', { date: formatDate(note.updatedAt) }) }}</div>
          </div>
        </button>
      </aside>

      <section v-if="!isPwa || pwaMode !== 'list'" class="rounded-md border border-border-subtle bg-card p-6 shadow-[var(--shadow-card)] md:min-h-[460px]">
        <div v-if="!editMode && !selectedNote" class="flex min-h-[320px] flex-col items-center justify-center gap-3 text-center text-text-muted">
          <BookOpen :size="44" />
          <h2 class="m-0 text-xl text-text-primary">{{ t('notes.selectOrCreate') }}</h2>
          <p class="max-w-[34rem] text-text-secondary">{{ t('notes.selectHint') }}</p>
          <button class="btn btn-primary" @click="startNewNote">{{ t('notes.new') }}</button>
        </div>

        <form v-else-if="editMode" @submit.prevent="saveNote">
          <div class="mb-4 flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
            <div>
              <p class="section-label">{{ selectedId === null ? t('notes.newNote') : t('notes.editing') }}</p>
              <h2 class="m-0 text-xl text-text-primary">{{ selectedId === null ? t('notes.draft') : draftTitle || t('notes.untitled') }}</h2>
              <p class="text-chip text-text-muted">{{ t('notes.writeHint') }}</p>
            </div>
            <div class="flex flex-wrap gap-2 md:justify-end">
              <button type="button" class="btn btn-ghost btn-sm" @click="cancelEdit">{{ t('common.cancel') }}</button>
              <button type="submit" class="btn btn-primary btn-sm" :disabled="saving">
                <Save :size="16" />
                {{ saving ? t('common.saving') : t('common.save') }}
              </button>
            </div>
          </div>

          <p v-if="formError" class="mb-4 rounded-sm border border-error-bg bg-overlay p-3 text-base text-text-primary">{{ formError }}</p>

          <label class="form-group">
            <span class="form-label">{{ t('common.title') }}</span>
            <input v-model="draftTitle" class="form-input" type="text" maxlength="200" :placeholder="t('notes.titlePlaceholder')" />
          </label>

          <label class="form-group mb-0 block">
            <span class="form-label">{{ t('chat.note.markdown') }}</span>
            <textarea
              v-model="draftBody"
              class="form-textarea min-h-[300px] leading-[1.6] md:min-h-[360px]"
              maxlength="20000"
              rows="16"
              :placeholder="t('notes.bodyPlaceholder')"
            ></textarea>
          </label>
        </form>

        <article v-else-if="selectedNote">
          <div class="mb-4">
            <div>
              <h2 class="m-0 text-xl text-text-primary">{{ selectedNote.title || t('notes.untitled') }}</h2>
              <div class="mt-3 flex flex-wrap gap-2">
                <button class="btn btn-secondary btn-sm" @click="editSelected">
                  <Pencil :size="16" />
                  {{ t('common.edit') }}
                </button>
                <button class="btn btn-danger btn-sm" :disabled="deleting" @click="deleteSelected">
                  <Trash2 :size="16" />
                  {{ deleting ? t('common.deleting') : t('common.delete') }}
                </button>
              </div>
              <p class="text-chip text-text-muted">{{ t('notes.updated', { date: formatDate(selectedNote.updatedAt) }) }}</p>
            </div>
          </div>

          <div
            v-if="selectedNote.body"
            class="markdown-rendered rounded-sm border border-border-subtle bg-input p-4 text-body leading-[1.7] text-text-secondary [&_a]:break-words [&_a]:text-gold [&_blockquote]:mb-3 [&_h1]:mb-2 [&_h1]:mt-4 [&_h1]:text-gold [&_h2]:mb-2 [&_h2]:mt-4 [&_h2]:text-gold [&_h3]:mb-2 [&_h3]:mt-4 [&_h3]:text-gold [&_ol]:mb-3 [&_ol]:pl-5 [&_p]:mb-3 [&_strong]:text-text-primary [&_ul]:mb-3 [&_ul]:pl-5"
            v-html="renderedSelectedBody"
          ></div>
          <div v-else class="empty-state flex min-h-[320px] flex-col items-center justify-center gap-3 text-center text-text-muted">
            <p>{{ t('notes.noBody') }}</p>
            <button class="btn btn-secondary btn-sm" @click="editSelected">{{ t('notes.addMarkdown') }}</button>
          </div>
        </article>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { AlertCircle, ArrowLeft, BookOpen, CirclePlus, Pencil, Plus, Save, StickyNote, Trash2 } from 'lucide-vue-next'
import { createNote, deleteNote, getNotes, updateNote } from '@/api/client'
import type { UserNote } from '@/types'
import { useDialog } from '@/composables/useDialog'
import { renderSafeMarkdown } from '@/composables/useMarkdown'
import { usePwa } from '@/composables/usePwa'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import AppIconButton from '@/components/ui/AppIconButton.vue'
import { t, intlLocale } from '@/i18n'

const { showConfirm, showAlert } = useDialog()
const { isPwa } = usePwa()

const notes = ref<UserNote[]>([])
const selectedId = ref<number | null>(null)
const loading = ref(true)
const saving = ref(false)
const deleting = ref(false)
const editMode = ref(false)
const loadError = ref('')
const formError = ref('')
const draftTitle = ref('')
const draftBody = ref('')
const pwaMode = ref<'list' | 'detail' | 'edit'>('list')

const selectedNote = computed(() => notes.value.find(note => note.id === selectedId.value) ?? null)
const renderedSelectedBody = computed(() => renderSafeMarkdown(selectedNote.value?.body))

onMounted(loadNotes)

async function loadNotes() {
  loading.value = true
  loadError.value = ''
  try {
    const res = await getNotes()
    notes.value = [...(res.data.notes ?? [])].sort(sortByUpdatedDesc)
    selectedId.value = notes.value[0]?.id ?? null
  } catch {
    loadError.value = t('notes.loadFailed')
  } finally {
    loading.value = false
  }
}

function sortByUpdatedDesc(a: UserNote, b: UserNote): number {
  return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
}

function startNewNote() {
  selectedId.value = null
  draftTitle.value = ''
  draftBody.value = ''
  formError.value = ''
  editMode.value = true
  if (isPwa) pwaMode.value = 'edit'
}

function selectNote(id: number) {
  selectedId.value = id
  formError.value = ''
  editMode.value = false
  if (isPwa) pwaMode.value = 'detail'
}

function editSelected() {
  if (!selectedNote.value) return
  draftTitle.value = selectedNote.value.title
  draftBody.value = selectedNote.value.body
  formError.value = ''
  editMode.value = true
  if (isPwa) pwaMode.value = 'edit'
}

function cancelEdit() {
  formError.value = ''
  if (selectedId.value === null) {
    editMode.value = false
    selectedId.value = notes.value[0]?.id ?? null
    if (isPwa) pwaMode.value = 'list'
    return
  }
  editMode.value = false
  if (isPwa) pwaMode.value = 'detail'
}

async function saveNote() {
  const body = draftBody.value
  const title = deriveTitle(draftTitle.value, body)
  if (!title && !body.trim()) {
    formError.value = t('notes.titleOrBody')
    return
  }

  saving.value = true
  formError.value = ''
  try {
    if (selectedId.value === null) {
      const res = await createNote({ title: title || t('notes.untitled'), body })
      notes.value = [res.data, ...notes.value].sort(sortByUpdatedDesc)
      selectedId.value = res.data.id
    } else {
      const res = await updateNote(selectedId.value, { title: title || t('notes.untitled'), body })
      notes.value = notes.value.map(note => note.id === res.data.id ? res.data : note).sort(sortByUpdatedDesc)
      selectedId.value = res.data.id
    }
    editMode.value = false
    if (isPwa) pwaMode.value = 'detail'
  } catch {
    formError.value = t('notes.saveFailed')
  } finally {
    saving.value = false
  }
}

function deriveTitle(rawTitle: string, body: string): string {
  const title = rawTitle.trim()
  if (title) return title
  const firstLine = body.split('\n').find(line => line.trim())
  return firstLine?.replace(/^#+\s*/, '').trim().slice(0, 80) ?? ''
}

async function deleteSelected() {
  if (!selectedNote.value) return
  const confirmed = await showConfirm(t('notes.confirmDelete'), {
    title: t('notes.deleteTitle'),
    confirmLabel: t('common.delete'),
    variant: 'danger',
  })
  if (!confirmed) return

  deleting.value = true
  try {
    const id = selectedNote.value.id
    await deleteNote(id)
    notes.value = notes.value.filter(note => note.id !== id)
    selectedId.value = notes.value[0]?.id ?? null
    editMode.value = false
    if (isPwa) pwaMode.value = 'list'
  } catch {
    await showAlert(t('notes.deleteFailed'), { title: t('admin.catalogs.deleteFailedTitle') })
  } finally {
    deleting.value = false
  }

}

function showPwaList() {
  pwaMode.value = 'list'
  editMode.value = false
  formError.value = ''
}

function previewText(body: string): string {
  const compact = body.replace(/[#*_`>\-[\]()]/g, ' ').replace(/\s+/g, ' ').trim()
  return compact || t('notes.noBodyPreview')
}

function formatDate(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return t('notes.unknownDate')
  return date.toLocaleDateString(intlLocale(), { month: 'short', day: 'numeric', year: 'numeric' })
}
</script>

<style scoped>
/*
 * :deep() audit — markdown-rendered content
 * Target: HTML elements emitted by markdown-it inside .markdown-rendered.
 * Notes are stored as Markdown and rendered at runtime; the resulting HTML
 * nodes are not authored in the Vue template and cannot be styled by scope hash
 * or Tailwind utilities.
 */
.markdown-rendered :deep(h1),
.markdown-rendered :deep(h2),
.markdown-rendered :deep(h3) {
  margin: 1rem 0 0.5rem;
  color: var(--accent-gold);
}

.markdown-rendered :deep(p),
.markdown-rendered :deep(ul),
.markdown-rendered :deep(ol),
.markdown-rendered :deep(blockquote) {
  margin-bottom: 0.75rem;
}

.markdown-rendered :deep(ul),
.markdown-rendered :deep(ol) {
  padding-left: 1.25rem;
}

.markdown-rendered :deep(strong) {
  color: var(--text-primary);
}

.markdown-rendered :deep(a) {
  color: var(--accent-gold);
  word-break: break-word;
}
</style>
