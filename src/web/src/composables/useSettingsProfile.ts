import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { changePassword, uploadAvatar, deleteAvatar, updateProfile } from '@/api/client'
import { t } from '@/i18n'

export function useSettingsProfile() {
  const auth = useAuthStore()

  // Avatar
  const avatarUrl = ref('/coin-logo.jpg')

  function updateAvatarUrl() {
    avatarUrl.value = auth.user?.avatarPath ? `/uploads/${auth.user.avatarPath}` : '/coin-logo.jpg'
  }
  updateAvatarUrl()

  async function handleAvatarUpload(e: Event) {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (!file) return
    try {
      const res = await uploadAvatar(file)
      if (auth.user) {
        auth.user.avatarPath = res.data.avatarPath
        localStorage.setItem('user', JSON.stringify(auth.user))
      }
      updateAvatarUrl()
    } catch { /* ignore */ }
  }

  async function handleAvatarDelete() {
    try {
      await deleteAvatar()
      if (auth.user) {
        auth.user.avatarPath = ''
        localStorage.setItem('user', JSON.stringify(auth.user))
      }
      updateAvatarUrl()
    } catch { /* ignore */ }
  }

  // Profile
  const profileEmail = ref(auth.user?.email || '')
  const profileBio = ref(auth.user?.bio || '')
  const profileZipCode = ref(auth.user?.zipCode || '')
  const profilePublic = ref(auth.user?.isPublic || false)
  const coinOfDayEnabled = ref(auth.user?.coinOfDayEnabled ?? true)
  const coinOfDayIncludeWishlist = ref(auth.user?.coinOfDayIncludeWishlist ?? true)
  const emperorTrackerEnabled = ref(auth.user?.emperorTrackerEnabled ?? false)
  const emperorTrackerShowUsurpers = ref(auth.user?.emperorTrackerShowUsurpers ?? false)
  const emperorTrackerShowEmpresses = ref(auth.user?.emperorTrackerShowEmpresses ?? false)
  const emperorTrackerShowOtherFigures = ref(auth.user?.emperorTrackerShowOtherFigures ?? false)
  const pwaSwipeNavEnabled = ref(auth.user?.pwaSwipeNavEnabled ?? false)
  const profileMsg = ref('')
  const profileError = ref(false)
  const profileSaving = ref(false)
  const showPrivacyWarning = ref(false)

  function onPublicToggle(e: Event) {
    const checked = (e.target as HTMLInputElement).checked
    if (!checked && profilePublic.value) {
      ;(e.target as HTMLInputElement).checked = true
      showPrivacyWarning.value = true
    } else {
      profilePublic.value = checked
    }
  }

  function confirmGoPrivate() {
    profilePublic.value = false
    showPrivacyWarning.value = false
  }

  function cancelGoPrivate() {
    showPrivacyWarning.value = false
  }

  async function handleSaveProfile() {
    profileMsg.value = ''
    profileError.value = false
    profileSaving.value = true
    try {
      const data: Record<string, unknown> = {
        email: profileEmail.value,
        bio: profileBio.value,
        zipCode: profileZipCode.value,
        isPublic: profilePublic.value,
        coinOfDayEnabled: coinOfDayEnabled.value,
        coinOfDayIncludeWishlist: coinOfDayIncludeWishlist.value,
        emperorTrackerEnabled: emperorTrackerEnabled.value,
        emperorTrackerShowUsurpers: emperorTrackerShowUsurpers.value,
        emperorTrackerShowEmpresses: emperorTrackerShowEmpresses.value,
        emperorTrackerShowOtherFigures: emperorTrackerShowOtherFigures.value,
      }
      const res = await updateProfile(data as Parameters<typeof updateProfile>[0])
      if (auth.user) {
        auth.user.email = res.data.email
        auth.user.bio = res.data.bio
        auth.user.zipCode = res.data.zipCode
        auth.user.isPublic = res.data.isPublic
        auth.user.coinOfDayEnabled = res.data.coinOfDayEnabled
        auth.user.coinOfDayIncludeWishlist = res.data.coinOfDayIncludeWishlist
        auth.user.emperorTrackerEnabled = res.data.emperorTrackerEnabled
        auth.user.emperorTrackerShowUsurpers = res.data.emperorTrackerShowUsurpers
        auth.user.emperorTrackerShowEmpresses = res.data.emperorTrackerShowEmpresses
        auth.user.emperorTrackerShowOtherFigures = res.data.emperorTrackerShowOtherFigures
        localStorage.setItem('user', JSON.stringify(auth.user))
      }
      profileMsg.value = t('settings.account.profileSaved')
    } catch {
      profileMsg.value = t('settings.account.profileSaveFailed')
      profileError.value = true
    } finally {
      profileSaving.value = false
    }
  }


  // Swipe Navigation -- immediate-apply save for View Settings / Appearance tab.
  // Uses a narrow payload so Account saves cannot clobber a newer value set here.
  async function savePwaSwipeNav(value: boolean) {
    const previous = pwaSwipeNavEnabled.value
    pwaSwipeNavEnabled.value = value
    try {
      const res = await updateProfile({ pwaSwipeNavEnabled: value })
      if (auth.user) {
        auth.user.pwaSwipeNavEnabled = res.data.pwaSwipeNavEnabled
        localStorage.setItem('user', JSON.stringify(auth.user))
      }
      pwaSwipeNavEnabled.value = res.data.pwaSwipeNavEnabled
    } catch (err) {
      pwaSwipeNavEnabled.value = previous
      throw err
    }
  }
  // Password
  const currentPassword = ref('')
  const newPassword = ref('')
  const confirmPassword = ref('')
  const passwordMsg = ref('')
  const passwordError = ref(false)
  const passwordLoading = ref(false)

  async function handleChangePassword() {
    passwordMsg.value = ''
    passwordError.value = false

    if (newPassword.value !== confirmPassword.value) {
      passwordMsg.value = t('settings.account.passwordMismatch')
      passwordError.value = true
      return
    }

    passwordLoading.value = true
    try {
      await changePassword(currentPassword.value, newPassword.value)
      passwordMsg.value = t('settings.account.passwordChanged')
      currentPassword.value = ''
      newPassword.value = ''
      confirmPassword.value = ''
    } catch {
      passwordMsg.value = t('settings.account.passwordChangeFailed')
      passwordError.value = true
    } finally {
      passwordLoading.value = false
    }
  }

  return {
    // Avatar
    avatarUrl,
    handleAvatarUpload,
    handleAvatarDelete,
    // Profile
    profileEmail,
    profileBio,
    profileZipCode,
    profilePublic,
    profileMsg,
    profileError,
    profileSaving,
    showPrivacyWarning,
    onPublicToggle,
    confirmGoPrivate,
    cancelGoPrivate,
    handleSaveProfile,
    coinOfDayEnabled,
    coinOfDayIncludeWishlist,
    emperorTrackerEnabled,
    emperorTrackerShowUsurpers,
    emperorTrackerShowEmpresses,
    emperorTrackerShowOtherFigures,
    // Password
    pwaSwipeNavEnabled,
    savePwaSwipeNav,
    currentPassword,
    newPassword,
    confirmPassword,
    passwordMsg,
    passwordError,
    passwordLoading,
    handleChangePassword,
  }
}
