import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import type { Chat, Message } from '@/types'

interface ChatState {
  chats: Record<string, Chat>
  messages: Record<string, Message[]>
  activeChat: string | null
  typingUsers: Record<string, string[]> // chatId -> [userId, ...]
}

interface ChatActions {
  reset: () => void
  setActiveChat: (chatId: string) => void
  setChats: (chats: Chat[]) => void
  addChat: (chat: Chat) => void
  addMessage: (chatId: string, message: Message) => void
  loadMessages: (chatId: string, messages: Message[]) => void
  addOptimisticMessage: (chatId: string, message: Message) => void
  confirmMessage: (chatId: string, tempId: string, confirmed: Message) => void
  failMessage: (chatId: string, tempId: string) => void
  setTyping: (chatId: string, userId: string, isTyping: boolean) => void
}

const initialState: ChatState = {
  chats: {},
  messages: {},
  activeChat: null,
  typingUsers: {},
}

export const useChatStore = create<ChatState & ChatActions>()(
  immer((set) => ({
    ...initialState,

    reset: () => set(() => ({ ...initialState })),

    setActiveChat: (chatId) =>
      set((state) => {
        state.activeChat = chatId
      }),

    setChats: (chats) =>
      set((state) => {
        state.chats = Object.fromEntries(chats.map((c) => [c.chat_id, c]))
      }),

    addChat: (chat) =>
      set((state) => {
        state.chats[chat.chat_id] = chat
      }),

    addMessage: (chatId, message) =>
      set((state) => {
        if (!state.messages[chatId]) state.messages[chatId] = []
        // Avoid duplicates
        const exists = state.messages[chatId].some((m) => m.message_id === message.message_id)
        if (!exists) state.messages[chatId].push(message)
      }),

    loadMessages: (chatId, messages) =>
      set((state) => {
        state.messages[chatId] = messages
      }),

    addOptimisticMessage: (chatId, message) =>
      set((state) => {
        if (!state.messages[chatId]) state.messages[chatId] = []
        state.messages[chatId].push({ ...message, status: 'pending' })
      }),

    confirmMessage: (chatId, tempId, confirmed) =>
      set((state) => {
        const msgs = state.messages[chatId]
        if (!msgs) return
        const idx = msgs.findIndex((m) => m.message_id === tempId)
        if (idx !== -1) msgs[idx] = confirmed
      }),

    failMessage: (chatId, tempId) =>
      set((state) => {
        const msgs = state.messages[chatId]
        if (!msgs) return
        const idx = msgs.findIndex((m) => m.message_id === tempId)
        if (idx !== -1) msgs[idx].status = 'error'
      }),

    setTyping: (chatId, userId, isTyping) =>
      set((state) => {
        if (!state.typingUsers[chatId]) state.typingUsers[chatId] = []
        if (isTyping) {
          if (!state.typingUsers[chatId].includes(userId)) {
            state.typingUsers[chatId].push(userId)
          }
        } else {
          state.typingUsers[chatId] = state.typingUsers[chatId].filter((id) => id !== userId)
        }
      }),
  })),
)
