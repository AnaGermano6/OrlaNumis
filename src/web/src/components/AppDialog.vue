<template>
  <BaseModal
    :open="dialog.state.value.visible"
    :aria-label="dialog.state.value.title || t('common.dialog')"
    max-width="max-w-[400px]"
    :sheet="false"
    alert
    layer="dialog"
    @close="onCancel"
  >
    <h3 v-if="dialog.state.value.title" class="mb-2 text-base text-heading">{{ dialog.state.value.title }}</h3>
    <p class="m-0 text-base leading-6 text-text-secondary">{{ dialog.state.value.message }}</p>
    <div class="mt-6 flex justify-end gap-3">
      <button
        v-if="dialog.state.value.type === 'confirm'"
        class="btn btn-secondary"
        @click="onCancel"
      >
        {{ dialog.state.value.cancelLabel }}
      </button>
      <button
        class="btn"
        :class="dialog.state.value.variant === 'danger' ? 'btn-danger' : 'btn-primary'"
        @click="dialog.handleConfirm()"
      >
        {{ dialog.state.value.confirmLabel }}
      </button>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
import BaseModal from '@/components/ui/BaseModal.vue'
import { useDialog } from '@/composables/useDialog'
import { t } from '@/i18n'

const dialog = useDialog()

function onCancel() {
  dialog.handleCancel()
}
</script>
