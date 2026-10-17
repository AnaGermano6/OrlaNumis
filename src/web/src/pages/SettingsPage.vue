<template>
  <PullToRefresh :on-refresh="handleRefresh">
    <div class="container">
      <DesktopPageContext :title="t('nav.settings')">
        <div class="page-header !flex-nowrap !items-center">
          <h1>{{ t('nav.settings') }}</h1>
          <div v-if="isPwa" class="relative">
            <button class="btn btn-secondary btn-sm gap-1.5 text-body" @click="settingsMenuOpen = !settingsMenuOpen">
              <component :is="tabIcons[activeTab]" :size="16" />
              {{ tabs.find(tab => tab.id === activeTab)?.label }}
              <Menu :size="16" />
            </button>
            <Transition
              enter-active-class="transition-opacity duration-150 ease-out"
              enter-from-class="opacity-0"
              leave-active-class="transition-opacity duration-150 ease-in"
              leave-to-class="opacity-0"
            >
              <div v-if="settingsMenuOpen" class="absolute right-0 top-full z-50 mt-2 flex min-w-[180px] flex-col gap-0.5 rounded-md border border-border-subtle bg-card p-[0.3rem] shadow-[0_4px_20px_var(--overlay-40)]">
                <button
                  v-for="tab in tabs"
                  :key="tab.id"
                  class="flex w-full items-center gap-2 rounded-sm px-3 py-2 text-left text-body font-medium text-text-secondary transition-colors hover:bg-surface hover:text-text-primary"
                  :class="{ 'bg-[var(--accent-gold-dim)] text-gold hover:bg-[var(--accent-gold-dim)] hover:text-gold': activeTab === tab.id }"
                  @click="selectTab(tab.id); settingsMenuOpen = false"
                >
                  <component :is="tabIcons[tab.id]" :size="16" />
                  {{ tab.label }}
                </button>
              </div>
            </Transition>
          </div>
        </div>
      </DesktopPageContext>

      <div class="mx-auto flex max-w-[800px] flex-col gap-6">
        <!-- Rows of equal columns (2, then 4 on wide screens), so every tab has the same width and spacing. -->
        <div v-if="!isPwa" class="grid grid-cols-2 gap-1 rounded-md border border-border-subtle bg-card p-[0.3rem] lg:grid-cols-4" data-testid="settings-tabs">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            class="flex min-w-0 items-center justify-center gap-1.5 rounded-sm px-1.5 py-2 text-center text-chip font-medium text-text-secondary transition-colors hover:text-text-primary md:px-3 md:py-2.5 md:text-body"
            :class="{ 'bg-[var(--accent-gold-dim)] text-gold hover:text-gold': activeTab === tab.id }"
            @click="selectTab(tab.id)"
          >
            <component :is="tabIcons[tab.id]" :size="16" /> {{ tab.label }}
          </button>
        </div>

        <SettingsAccountSection v-if="activeTab === 'account'" />

        <SettingsAppearanceSection
          v-if="activeTab === 'appearance'"
          :theme="theme"
          :timezone="timezone"
          :timezones="timezones"
          :default-view="defaultView"
          :default-sort="defaultSort"
          :tray-felt-color="feltColor"
          @set-theme="setTheme"
          @save-timezone="(tz: string) => { timezone = tz; saveTimezone() }"
          @set-default-view="setDefaultView"
          @save-default-sort="(sort: string) => { defaultSort = sort; saveDefaultSort() }"
          @set-tray-felt-color="(color) => { feltColor = color }"
        />

        <SettingsDataSection v-if="activeTab === 'data'" ref="dataSection" />

        <SettingsFullBackup v-if="activeTab === 'backups' && auth.isAdmin" />
        <SettingsBackupsSection v-if="activeTab === 'backups'" ref="backupsSection" />

        <SettingsToolsSection
          v-if="activeTab === 'tools'"
          @saved="handleProcessSaved"
        />

        <SavedConversationsSection
          v-if="activeTab === 'conversations'"
          :conversations="conversations"
          :loading="conversationsLoading"
          @open="openConversation"
          @delete="handleDeleteConversation"
        />

        <HelpSection v-if="activeTab === 'help'" />

        <CoinSearchChat
          v-if="showChat"
          :load-conversation="chatConversation"
          @close="showChat = false; chatConversation = null"
          @added="() => {}"
        />
      </div>
    </div>
  </PullToRefresh>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, type Component } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import PullToRefresh from '@/components/PullToRefresh.vue'
import { listConversations, getConversation, deleteConversation } from '@/api/client'
import type { ConversationSummary } from '@/api/client'
import { useDialog } from '@/composables/useDialog'
import { usePwa } from '@/composables/usePwa'
import { useTrayPreference } from '@/composables/useTrayPreference'
import type { Theme } from '@/types'
import CoinSearchChat from '@/components/CoinSearchChat.vue'
import HelpSection from '@/components/HelpSection.vue'
import SettingsAccountSection from '@/components/settings/SettingsAccountSection.vue'
import SettingsAppearanceSection from '@/components/settings/SettingsAppearanceSection.vue'
import SettingsDataSection from '@/components/settings/SettingsDataSection.vue'
import SettingsBackupsSection from '@/components/settings/SettingsBackupsSection.vue'
import SettingsFullBackup from '@/components/settings/SettingsFullBackup.vue'
import SavedConversationsSection from '@/components/settings/SavedConversationsSection.vue'
import SettingsToolsSection from '@/components/settings/SettingsToolsSection.vue'
import { User, Palette, Database, MessageSquare, HelpCircle, Wrench, Menu, ShieldCheck, Archive } from 'lucide-vue-next'
import DesktopPageContext from '@/components/DesktopPageContext.vue'
import { t, type MessageKey } from '@/i18n'

const tabIcons: Record<string, Component> = {
  account: User,
  appearance: Palette,
  data: Database,
  backups: Archive,
  tools: Wrench,
  conversations: MessageSquare,
  help: HelpCircle,
  admin: ShieldCheck,
}

const { showConfirm, showAlert } = useDialog()
const activeTab = ref('account')
const settingsMenuOpen = ref(false)
const { isPwa } = usePwa()
const { feltColor } = useTrayPreference()

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const dataSection = ref<InstanceType<typeof SettingsDataSection> | null>(null)
const backupsSection = ref<InstanceType<typeof SettingsBackupsSection> | null>(null)

const tabLabelKeys: Record<string, MessageKey> = {
  account: 'settings.tabs.account',
  appearance: 'settings.tabs.appearance',
  data: 'settings.tabs.data',
  tools: 'settings.tabs.tools',
  conversations: 'settings.tabs.conversations',
  backups: 'settings.tabs.backups',
  help: 'settings.tabs.help',
  admin: 'nav.admin',
}
const baseTabIds = ['account', 'appearance', 'data', 'tools', 'conversations', 'backups', 'help']
const validTabIds = baseTabIds.concat('admin')
// In the installed app, admins reach Admin through the settings menu, right after Account.
const tabs = computed(() => {
  const ids = isPwa && auth.isAdmin ? ['account', 'admin', ...baseTabIds.slice(1)] : baseTabIds
  return ids.map((id) => ({ id, label: t(tabLabelKeys[id]!) }))
})

function applyTabFromRoute(tabValue: unknown) {
  if (typeof tabValue !== 'string' || !validTabIds.includes(tabValue)) {
    return
  }
  if (tabValue === 'admin') {
    if (auth.isAdmin) {
      router.push('/admin')
    }
    return
  }
  activeTab.value = tabValue
}

function selectTab(tabId: string) {
  if (tabId === 'admin') {
    router.push('/admin')
    return
  }
  activeTab.value = tabId
  router.replace({ query: { ...route.query, tab: tabId } })
}

applyTabFromRoute(route.query.tab)

watch(() => route.query.tab, (tab) => {
  applyTabFromRoute(tab)
})

function handleProcessSaved(savedCoinId: number) {
  router.push(`/edit/${savedCoinId}`)
}

// Theme
const theme = ref<Theme>((localStorage.getItem('theme') as Theme) || 'dark')

function setTheme(t: Theme) {
  theme.value = t
  localStorage.setItem('theme', t)
  document.documentElement.setAttribute('data-theme', t)
}

// Timezone
const timezones = 'supportedValuesOf' in Intl
  ? (Intl as unknown as { supportedValuesOf: (key: string) => string[] }).supportedValuesOf('timeZone')
  : [] as string[]
const timezone = ref(localStorage.getItem('timezone') || Intl.DateTimeFormat().resolvedOptions().timeZone)

function saveTimezone() {
  localStorage.setItem('timezone', timezone.value)
}

// Default view
const defaultView = ref<'swipe' | 'grid'>((localStorage.getItem('defaultView') as 'swipe' | 'grid') || 'swipe')

function setDefaultView(v: 'swipe' | 'grid') {
  defaultView.value = v
  localStorage.setItem('defaultView', v)
}

// Default sort
const defaultSort = ref(localStorage.getItem('defaultSort') || 'updated_at_desc')

function saveDefaultSort() {
  localStorage.setItem('defaultSort', defaultSort.value)
}

// Saved Conversations
const conversations = ref<ConversationSummary[]>([])
const conversationsLoading = ref(false)
const showChat = ref(false)
const chatConversation = ref<{ id: number; title: string; messages: string } | null>(null)

async function loadConversations() {
  conversationsLoading.value = true
  try {
    const res = await listConversations()
    conversations.value = res.data
  } catch {
    // silently fail
  } finally {
    conversationsLoading.value = false
  }
}

async function openConversation(id: number) {
  try {
    const res = await getConversation(id)
    chatConversation.value = {
      id: res.data.id,
      title: res.data.title,
      messages: res.data.messages,
    }
    showChat.value = true
  } catch {
    await showAlert(t('settings.conversations.loadFailed'), { title: t('common.error') })
  }
}

async function handleDeleteConversation(id: number) {
  if (!await showConfirm(t('settings.conversations.confirmDelete'), { title: t('settings.conversations.deleteTitle'), variant: 'danger' })) return
  try {
    await deleteConversation(id)
    conversations.value = conversations.value.filter(c => c.id !== id)
  } catch {
    await showAlert(t('settings.conversations.deleteFailed'), { title: t('common.error') })
  }
}

async function handleRefresh() {
  await loadConversations()
}

onMounted(() => {
  loadConversations()
})
</script>
