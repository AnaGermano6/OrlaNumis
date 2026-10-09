import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import {
  changePassword, uploadAvatar, deleteAvatar, updateProfile,
  validateNumisBidsCredentials, testPushover,
} from '@/api/client'
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
  const nbUsername = ref(auth.user?.numisBidsUsername || '')
  const nbPassword = ref('')
  const cngUsername = ref(auth.user?.cngUsername || '')
  const cngPassword = ref('')
  const parcelAppAPIKey = ref('')
  const pushoverKey = ref('')
  const pushoverTesting = ref(false)
  const pushoverTestMsg = ref('')
  const pushoverTestError = ref(false)
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

  // NumisBids validation
  const nbValidating = ref(false)
  const nbValidationError = ref('')
  const cngValidating = ref(false)
  const cngValidationError = ref('')

  async function handleSaveProfile() {
    profileMsg.value = ''
    profileError.value = false
    profileSaving.value = true
    nbValidationError.value = ''
    cngValidationError.value = ''
    try {
      if (nbPassword.value && nbUsername.value) {
        nbValidating.value = true
        try {
          const valRes = await validateNumisBidsCredentials(nbUsername.value, nbPassword.value)
          if (!valRes.data.valid) {
            nbValidationError.value = valRes.data.error || t('settings.connections.numisbids.invalid')
            profileSaving.value = false
            nbValidating.value = false
            return
          }
        } catch {
          nbValidationError.value = t('settings.connections.numisbids.validateFailed')
          profileSaving.value = false
          nbValidating.value = false
          return
        } finally {
          nbValidating.value = false
        }
      }
      if (cngPassword.value && cngUsername.value) {
        cngValidating.value = true
        try {
          const valRes = await validateNumisBidsCredentials(cngUsername.value, cngPassword.value, 'cng')
          if (!valRes.data.valid) {
            cngValidationError.value = valRes.data.error || t('settings.connections.cng.invalid')
            profileSaving.value = false
            cngValidating.value = false
            return
          }
        } catch {
          cngValidationError.value = t('settings.connections.cng.validateFailed')
          profileSaving.value = false
          cngValidating.value = false
          return
        } finally {
          cngValidating.value = false
        }
      }

      const data: Record<string, unknown> = {
        email: profileEmail.value,
        bio: profileBio.value,
        zipCode: profileZipCode.value,
        isPublic: profilePublic.value,
        numisBidsUsername: nbUsername.value,
        cngUsername: cngUsername.value,
        coinOfDayEnabled: coinOfDayEnabled.value,
        coinOfDayIncludeWishlist: coinOfDayIncludeWishlist.value,
        emperorTrackerEnabled: emperorTrackerEnabled.value,
        emperorTrackerShowUsurpers: emperorTrackerShowUsurpers.value,
        emperorTrackerShowEmpresses: emperorTrackerShowEmpresses.value,
        emperorTrackerShowOtherFigures: emperorTrackerShowOtherFigures.value,
      }
      if (nbPassword.value) {
        data.numisBidsPassword = nbPassword.value
      }
      if (cngPassword.value) {
        data.cngPassword = cngPassword.value
      }
      if (parcelAppAPIKey.value !== '') {
        data.parcelAppAPIKey = parcelAppAPIKey.value
      }
      if (pushoverKey.value !== '') {
        data.pushoverUserKey = pushoverKey.value
      }
      const res = await updateProfile(data as Parameters<typeof updateProfile>[0])
      if (auth.user) {
        auth.user.email = res.data.email
        auth.user.bio = res.data.bio
        auth.user.zipCode = res.data.zipCode
        auth.user.isPublic = res.data.isPublic
        auth.user.numisBidsUsername = res.data.numisBidsUsername
        auth.user.numisBidsConfigured = res.data.numisBidsConfigured
        auth.user.cngUsername = res.data.cngUsername
        auth.user.cngConfigured = res.data.cngConfigured
        auth.user.parcelAppConfigured = res.data.parcelAppConfigured
        auth.user.pushoverEnabled = res.data.pushoverEnabled
        auth.user.coinOfDayEnabled = res.data.coinOfDayEnabled
        auth.user.coinOfDayIncludeWishlist = res.data.coinOfDayIncludeWishlist
        auth.user.emperorTrackerEnabled = res.data.emperorTrackerEnabled
        auth.user.emperorTrackerShowUsurpers = res.data.emperorTrackerShowUsurpers
        auth.user.emperorTrackerShowEmpresses = res.data.emperorTrackerShowEmpresses
        auth.user.emperorTrackerShowOtherFigures = res.data.emperorTrackerShowOtherFigures
        localStorage.setItem('user', JSON.stringify(auth.user))
      }
      nbPassword.value = ''
      cngPassword.value = ''
      parcelAppAPIKey.value = ''
      pushoverKey.value = ''
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

  // Pushover test
  async function handleTestPushover() {
    pushoverTesting.value = true
    pushoverTestMsg.value = ''
    pushoverTestError.value = false
    try {
      await testPushover()
      pushoverTestMsg.value = t('settings.connections.pushover.sent')
    } catch {
      pushoverTestMsg.value = t('settings.connections.pushover.failed')
      pushoverTestError.value = true
    } finally {
      pushoverTesting.value = false
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
    nbUsername,
    nbPassword,
    cngUsername,
    cngPassword,
    parcelAppAPIKey,
    pushoverKey,
    pushoverTesting,
    pushoverTestMsg,
    pushoverTestError,
    handleTestPushover,
    profilePublic,
    profileMsg,
    profileError,
    profileSaving,
    showPrivacyWarning,
    onPublicToggle,
    confirmGoPrivate,
    cancelGoPrivate,
    nbValidating,
    nbValidationError,
    cngValidating,
    cngValidationError,
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
