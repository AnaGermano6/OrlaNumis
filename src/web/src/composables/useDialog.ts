import { ref, readonly } from 'vue'
import { t } from '@/i18n'

export type DialogType = 'confirm' | 'alert'
export type DialogVariant = 'default' | 'danger'

interface DialogState {
  visible: boolean
  type: DialogType
  variant: DialogVariant
  title: string
  message: string
  confirmLabel: string
  cancelLabel: string
  resolve: ((value: boolean) => void) | null
}

const state = ref<DialogState>({
  visible: false,
  type: 'confirm',
  variant: 'default',
  title: '',
  message: '',
  confirmLabel: t('common.ok'),
  cancelLabel: t('common.cancel'),
  resolve: null,
})

function showConfirm(
  message: string,
  options?: { title?: string; confirmLabel?: string; cancelLabel?: string; variant?: DialogVariant },
): Promise<boolean> {
  return new Promise((resolve) => {
    state.value = {
      visible: true,
      type: 'confirm',
      variant: options?.variant ?? 'default',
      title: options?.title ?? t('common.confirm'),
      message,
      confirmLabel: options?.confirmLabel ?? t('common.confirm'),
      cancelLabel: options?.cancelLabel ?? t('common.cancel'),
      resolve,
    }
  })
}

function showAlert(
  message: string,
  options?: { title?: string; confirmLabel?: string; variant?: DialogVariant },
): Promise<boolean> {
  return new Promise((resolve) => {
    state.value = {
      visible: true,
      type: 'alert',
      variant: options?.variant ?? 'default',
      title: options?.title ?? '',
      message,
      confirmLabel: options?.confirmLabel ?? t('common.ok'),
      cancelLabel: t('common.cancel'),
      resolve,
    }
  })
}

function handleConfirm() {
  state.value.resolve?.(true)
  state.value.visible = false
  state.value.resolve = null
}

function handleCancel() {
  state.value.resolve?.(false)
  state.value.visible = false
  state.value.resolve = null
}

export function useDialog() {
  return {
    state: readonly(state),
    showConfirm,
    showAlert,
    handleConfirm,
    handleCancel,
  }
}
