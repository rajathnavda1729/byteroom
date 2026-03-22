import { create } from 'zustand'
import type { ConnectionStatus } from '@/types'

interface ConnectionState {
  status: ConnectionStatus
  reconnectAttempts: number
}

interface ConnectionActions {
  setConnected: () => void
  setDisconnected: () => void
  setReconnecting: () => void
  incrementReconnectAttempts: () => void
  resetReconnectAttempts: () => void
}

export const useConnectionStore = create<ConnectionState & ConnectionActions>()((set) => ({
  status: 'disconnected',
  reconnectAttempts: 0,

  setConnected: () => set({ status: 'connected', reconnectAttempts: 0 }),
  setDisconnected: () => set({ status: 'disconnected' }),
  setReconnecting: () => set({ status: 'reconnecting' }),
  incrementReconnectAttempts: () =>
    set((state) => ({ reconnectAttempts: state.reconnectAttempts + 1 })),
  resetReconnectAttempts: () => set({ reconnectAttempts: 0 }),
}))
