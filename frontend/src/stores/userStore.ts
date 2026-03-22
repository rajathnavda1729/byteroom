import { create } from 'zustand'
import type { User } from '@/types'

interface UserState {
  currentUser: User | null
  token: string | null
}

interface UserActions {
  setCurrentUser: (user: User, token: string) => void
  clearCurrentUser: () => void
}

export const useUserStore = create<UserState & UserActions>()((set) => ({
  currentUser: null,
  token: null,

  setCurrentUser: (user, token) => set({ currentUser: user, token }),
  clearCurrentUser: () => set({ currentUser: null, token: null }),
}))
