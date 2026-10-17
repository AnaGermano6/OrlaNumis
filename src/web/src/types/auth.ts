// auth types. Split out of the former single-file src/types/index.ts;
// re-exported from '@/types' so existing imports keep working.

export interface User {
  id: number
  username: string
  role: 'admin' | 'user'
  email: string
  avatarPath: string
  bio: string
  zipCode: string
  pwaSwipeNavEnabled?: boolean
}

export interface AuthResponse {
  token: string
  refreshToken: string
  user: User
}

export interface UserInfo {
  id: number
  username: string
  role: 'admin' | 'user'
  email: string
  avatarPath: string
  bio: string
  zipCode: string
  emailMissing: boolean
  pwaSwipeNavEnabled?: boolean
  lockedUntil?: string | null
  failedLoginAttempts?: number
  createdAt: string
}
