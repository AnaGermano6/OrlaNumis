<template>
  <div v-if="user" class="fixed inset-0 z-[200] flex items-center justify-center bg-overlay-60 p-4" @click.self="$emit('close')">
    <div class="card w-full max-w-[520px] p-6">
      <h3 class="mb-4 text-lg font-medium text-heading">{{ t('admin.users.editTitle', { name: user.username }) }}</h3>

      <div class="mb-4">
        <label class="form-label">{{ t('settings.account.role') }}</label>
        <div class="grid gap-2 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
          <select v-model="selectedRole" class="form-input" :disabled="savingRole || isCurrentUser">
            <option value="user">{{ t('settings.account.roles.user') }}</option>
            <option value="admin">{{ t('settings.account.roles.admin') }}</option>
          </select>
          <button class="btn btn-secondary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" :disabled="savingRole || isCurrentUser" @click="updateRole">
            {{ savingRole ? t('common.saving') : t('admin.users.updateRole') }}
          </button>
        </div>
        <p v-if="isCurrentUser" class="mt-2 text-body text-text-muted">{{ t('admin.users.ownRole') }}</p>
        <p v-if="roleMsg" class="mt-2 text-body text-gold" :class="{ 'text-byzantine': roleError }">{{ roleMsg }}</p>
      </div>

      <div class="mb-4">
        <label class="form-label">{{ t('admin.users.resetPassword') }}</label>
        <div class="grid gap-2 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
          <input
            v-model="password"
            type="password"
            class="form-input"
            :placeholder="t('admin.users.newPasswordPlaceholder')"
            minlength="6"
          />
          <button class="btn btn-secondary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" :disabled="resettingPassword" @click="resetPassword">
            {{ resettingPassword ? t('admin.users.resetting') : t('admin.users.resetPassword') }}
          </button>
        </div>
        <p v-if="passwordMsg" class="mt-2 text-body text-gold" :class="{ 'text-byzantine': passwordError }">{{ passwordMsg }}</p>
      </div>

      <div class="mb-4">
        <label class="form-label">{{ t('admin.users.deleteUser') }}</label>
        <button class="btn btn-danger btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" :disabled="deletingUser || isCurrentUser" @click="deleteTargetUser">
          {{ deletingUser ? t('common.deleting') : t('admin.users.deleteUser') }}
        </button>
        <p v-if="isCurrentUser" class="mt-2 text-body text-text-muted">{{ t('admin.users.ownDelete') }}</p>
        <p v-if="deleteMsg" class="mt-2 text-body text-gold" :class="{ 'text-byzantine': deleteError }">{{ deleteMsg }}</p>
      </div>

      <div class="flex justify-end">
        <button type="button" class="btn btn-primary btn-sm focus-visible:outline-2 focus-visible:outline-gold focus-visible:outline-offset-2" @click="$emit('close')">{{ t('common.done') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { UserInfo } from '@/types'
import { deleteUser, resetUserPassword, updateUserRole } from '@/api/client'
import { useDialog } from '@/composables/useDialog'
import { t } from '@/i18n'

const props = defineProps<{
  user: UserInfo | null
  currentUserId: number
}>()

const emit = defineEmits<{
  close: []
  'role-updated': [payload: { userId: number; role: UserInfo['role'] }]
  deleted: [userId: number]
}>()

const { showConfirm } = useDialog()

const selectedRole = ref<UserInfo['role']>('user')
const password = ref('')

const savingRole = ref(false)
const resettingPassword = ref(false)
const deletingUser = ref(false)

const roleMsg = ref('')
const roleError = ref(false)
const passwordMsg = ref('')
const passwordError = ref(false)
const deleteMsg = ref('')
const deleteError = ref(false)

const isCurrentUser = computed(() => (props.user?.id ?? 0) === props.currentUserId)

watch(() => props.user, (user) => {
  selectedRole.value = user?.role ?? 'user'
  password.value = ''
  roleMsg.value = ''
  roleError.value = false
  passwordMsg.value = ''
  passwordError.value = false
  deleteMsg.value = ''
  deleteError.value = false
}, { immediate: true })

async function updateRole() {
  if (!props.user || isCurrentUser.value) return
  savingRole.value = true
  roleMsg.value = ''
  roleError.value = false
  try {
    await updateUserRole(props.user.id, selectedRole.value)
    roleMsg.value = t('admin.users.roleUpdated')
    emit('role-updated', { userId: props.user.id, role: selectedRole.value })
  } catch {
    roleError.value = true
    roleMsg.value = t('admin.users.roleUpdateFailed')
  } finally {
    savingRole.value = false
  }
}

async function resetPassword() {
  if (!props.user) return
  if (!password.value || password.value.length < 6) {
    passwordError.value = true
    passwordMsg.value = t('admin.users.passwordTooShort')
    return
  }
  resettingPassword.value = true
  passwordMsg.value = ''
  passwordError.value = false
  try {
    await resetUserPassword(props.user.id, password.value)
    passwordMsg.value = t('admin.users.passwordReset')
    password.value = ''
  } catch {
    passwordError.value = true
    passwordMsg.value = t('admin.users.passwordResetFailed')
  } finally {
    resettingPassword.value = false
  }
}

async function deleteTargetUser() {
  if (!props.user || isCurrentUser.value) return
  const confirmed = await showConfirm(
    t('admin.users.confirmDelete', { name: props.user.username }),
    { title: t('admin.users.deleteUser'), variant: 'danger' }
  )
  if (!confirmed) return

  deletingUser.value = true
  deleteMsg.value = ''
  deleteError.value = false
  try {
    await deleteUser(props.user.id)
    emit('deleted', props.user.id)
  } catch {
    deleteError.value = true
    deleteMsg.value = t('admin.users.deleteFailed')
  } finally {
    deletingUser.value = false
  }
}
</script>
