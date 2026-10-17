// auth endpoints. Split out of the former monolithic client.ts;
// re-exported from '@/api/client' so existing imports keep working.
import { api } from '@/api/http'
import type {
  AuthResponse,
  Coin,
  UserInfo,
} from '@/types'

// Auth
export const login = (username: string, password: string) =>
  api.post<AuthResponse>('/auth/login', { username, password })

export const register = (username: string, password: string, email?: string) =>
  api.post<AuthResponse>('/auth/register', { username, password, email })

export const getMe = () => api.get<UserInfo>('/auth/me')

export const changePassword = (currentPassword: string, newPassword: string) =>
  api.post('/auth/change-password', { currentPassword, newPassword })

export const exportCollection = () => api.get('/user/export', { responseType: 'blob' })

export const exportCatalogPDF = () => api.get('/user/export/catalog', { responseType: 'blob' })

export const importCollection = (coins: Partial<Coin>[]) => api.post('/user/import', coins)

// Profile
export const updateProfile = (data: { email?: string; bio?: string; zipCode?: string; pwaSwipeNavEnabled?: boolean }) =>
  api.put<{ id: number; username: string; role: string; email: string; avatarPath: string; bio: string; zipCode: string; pwaSwipeNavEnabled: boolean }>('/user/profile', data)

export const uploadAvatar = (file: File) => {
  const form = new FormData()
  form.append('avatar', file)
  return api.post<{ avatarPath: string }>('/user/avatar', form)
}

export const deleteAvatar = () => api.delete('/user/avatar')
