import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, AuthResponse } from '@/types'
import * as api from '@/api/client'
import { onTokenRefreshed } from '@/api/client'
import { clearPrivateMediaBlobCache } from '@/utils/media'

const PRIVATE_MEDIA_CACHE_NAMES = ['coin-images']

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  function setTokens(data: AuthResponse) {
    token.value = data.token
    user.value = data.user
    localStorage.setItem('token', data.token)
    localStorage.setItem('refreshToken', data.refreshToken)
    localStorage.setItem('user', JSON.stringify(data.user))
  }

  function isUserSwitch(nextUser: User) {
    return user.value !== null && user.value.id !== nextUser.id
  }

  async function clearPrivateMediaCaches() {
    clearPrivateMediaBlobCache()
    if (typeof caches === 'undefined') return

    await Promise.allSettled(
      PRIVATE_MEDIA_CACHE_NAMES.map((cacheName) => caches.delete(cacheName)),
    )
  }

  // Keep Pinia store in sync after silent token refresh
  onTokenRefreshed((data) => {
    token.value = data.token
    user.value = data.user
  })

  async function applyAuthResponse(data: AuthResponse) {
    if (isUserSwitch(data.user)) {
      await clearPrivateMediaCaches()
    }
    setTokens(data)
  }

  async function doLogin(username: string, password: string) {
    const res = await api.login(username, password)
    await applyAuthResponse(res.data)
  }

  async function doRegister(username: string, password: string, email?: string) {
    const res = await api.register(username, password, email)
    await applyAuthResponse(res.data)
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('user')
    void clearPrivateMediaCaches()
  }

  return { token, user, isAuthenticated, isAdmin, applyAuthResponse, doLogin, doRegister, logout }
})
