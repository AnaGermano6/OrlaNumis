<template>
  <section class="card">
    <h2 class="text-xl font-medium mb-5 pb-3 border-b border-border-subtle">{{ t('admin.users.title') }}</h2>
    <div v-if="loading" class="loading-overlay"><div class="spinner"></div></div>
    <table v-else class="data-table [&_th]:px-2 [&_th]:py-3 [&_td]:px-2 [&_td]:py-3">
      <thead>
        <tr>
          <th>{{ t('auth.username') }}</th>
          <th>{{ t('settings.account.role') }}</th>
          <th>{{ t('admin.users.created') }}</th>
          <th>{{ t('common.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="user in users" :key="user.id">
          <td>
            <span class="font-medium">{{ user.username }}</span>
            <span v-if="user.id === currentUserId" class="text-label text-text-muted ml-[0.3rem]">{{ t('admin.users.you') }}</span>
          </td>
          <td>
            <span class="badge" :class="`badge-${user.role === 'admin' ? 'roman' : 'modern'}`">
              {{ user.role === 'admin' ? t('settings.account.roles.admin') : user.role === 'user' ? t('settings.account.roles.user') : user.role }}
            </span>
          </td>
          <td class="text-body text-text-secondary">{{ formatDate(user.createdAt) }}</td>
          <td>
            <div v-if="user.id !== currentUserId" class="flex gap-[0.4rem]">
              <button class="btn btn-secondary btn-sm" @click="$emit('edit', user)">{{ t('common.edit') }}</button>
            </div>
            <span v-else class="text-text-muted">—</span>
          </td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
import type { UserInfo } from '@/types'
import { t, intlLocale } from '@/i18n'

defineProps<{
  users: UserInfo[]
  loading: boolean
  currentUserId: number
}>()

defineEmits<{
  edit: [user: UserInfo]
}>()

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString(intlLocale())
}
</script>
