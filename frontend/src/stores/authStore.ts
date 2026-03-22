import { create } from 'zustand'
import type { User } from '@/types'

interface AuthPayload {
  user: User
  token: string
  refreshToken?: string
}

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
}

interface AuthActions {
  setAuth: (payload: AuthPayload) => void
  logout: () => void
}

function loadPersistedAuth(): Pick<AuthState, 'user' | 'token' | 'isAuthenticated'> {
  try {
    const token = localStorage.getItem('token')
    const userRaw = localStorage.getItem('auth_user')
    if (token && userRaw) {
      return { token, user: JSON.parse(userRaw) as User, isAuthenticated: true }
    }
  } catch {
    // ignore malformed stored data
  }
  return { token: null, user: null, isAuthenticated: false }
}

export const useAuthStore = create<AuthState & AuthActions>()((set) => ({
  ...loadPersistedAuth(),

  setAuth: ({ user, token }) => {
    localStorage.setItem('token', token)
    localStorage.setItem('auth_user', JSON.stringify(user))
    set({ user, token, isAuthenticated: true })
  },

  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('auth_user')
    set({ user: null, token: null, isAuthenticated: false })
  },
}))
