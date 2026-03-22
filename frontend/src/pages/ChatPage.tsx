import { useEffect, useCallback } from 'react'
import { v4 as uuidv4 } from 'uuid'
import { useChatStore } from '@/stores/chatStore'
import { useAuthStore } from '@/stores/authStore'
import { useWebSocket } from '@/hooks/useWebSocket'
import { api } from '@/services/api'
import { resolveWebSocketURL } from '@/utils/wsUrl'
import { ChatLayout } from '@/components/chat/ChatLayout'
import { MessageList } from '@/components/chat/MessageList'
import { MessageInput } from '@/components/chat/MessageInput'
import { TypingIndicator } from '@/components/chat/TypingIndicator'
import type { WSEvent, Chat, Message } from '@/types'

function getChatDisplayName(chat: Chat | undefined, currentUserId: string | undefined): string {
  if (!chat) return 'Chat'
  if (chat.name) return chat.name
  if (chat.type === 'direct') {
    const other = chat.members.find((m) => m.user_id !== currentUserId)
    if (other) return other.display_name || other.username
  }
  return 'Chat'
}

interface ChatsResponse {
  chats: (Omit<Chat, 'members'> & { members?: Chat['members'] })[]
}
interface MessagesResponse {
  messages: Message[]
}

export function ChatPage() {
  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)
  const activeChat = useChatStore((s) => s.activeChat)
  const activeChats = useChatStore((s) => s.chats)
  const { setChats, loadMessages, addOptimisticMessage, failMessage, typingUsers, confirmMessage } = useChatStore()

  // Load chats on mount
  useEffect(() => {
    api.get<ChatsResponse>('/api/chats').then((res) => {
      setChats((res.chats ?? []).map((c) => ({ ...c, members: c.members ?? [] })))
    })
  }, [setChats])

  // Load messages when active chat changes
  useEffect(() => {
    if (!activeChat) return
    api
      .get<MessagesResponse>(`/api/chats/${activeChat}/messages`)
      .then((res) => loadMessages(activeChat, res.messages ?? []))
      .catch((err) => console.error('Failed to load messages:', err))
  }, [activeChat, loadMessages])

  // Fallback when a WS frame was missed: refresh history when the tab becomes visible.
  useEffect(() => {
    const onVisible = () => {
      if (document.visibilityState !== 'visible') return
      const id = useChatStore.getState().activeChat
      if (!id) return
      void api
        .get<MessagesResponse>(`/api/chats/${id}/messages`)
        .then((res) => useChatStore.getState().loadMessages(id, res.messages ?? []))
        .catch(() => {})
    }
    document.addEventListener('visibilitychange', onVisible)
    return () => document.removeEventListener('visibilitychange', onVisible)
  }, [])

  const handleWSMessage = useCallback((event: WSEvent) => {
    if (event.event === 'message.new') {
      const d = event.data as Message & { created_at?: string }
      const ts = d.timestamp ?? d.created_at ?? new Date().toISOString()
      // getState avoids any stale closure if the WS callback ever lags a render
      useChatStore.getState().addMessage(d.chat_id, {
        ...d,
        timestamp: ts,
        content_type: d.content_type ?? 'markdown',
      })
      return
    }
    if (event.event === 'message.ack') {
      return
    }
    if (event.event === 'user.typing') {
      useChatStore.getState().setTyping(event.data.chat_id, event.data.user_id, event.data.is_typing)
      return
    }
    if (event.event === 'chat.new') {
      useChatStore.getState().addChat({ ...event.data, members: event.data.members ?? [] })
    }
  }, [])

  const { send, connectionState } = useWebSocket({
    url: resolveWebSocketURL('/ws'),
    token,
    onMessage: handleWSMessage,
    enabled: !!token,
  })

  const handleSend = async (content: string) => {
    if (!activeChat || !user?.id) return

    const tempId = `temp-${uuidv4()}`
    const messageId = uuidv4()

    addOptimisticMessage(activeChat, {
      message_id: tempId,
      chat_id: activeChat,
      sender_id: user.id,
      content_type: 'markdown',
      content,
      timestamp: new Date().toISOString(),
      status: 'pending',
    })

    // Persist via HTTP — WebSocket send was silently dropped when the socket
    // was not OPEN yet, which left the DB empty while the UI showed messages.
    try {
      const saved = await api.post<Message>(`/api/chats/${activeChat}/messages`, {
        message_id: messageId,
        content_type: 'markdown',
        content,
      })
      confirmMessage(activeChat, tempId, {
        message_id: saved.message_id,
        chat_id: saved.chat_id,
        sender_id: saved.sender_id,
        content_type: saved.content_type,
        content: saved.content,
        timestamp: saved.timestamp,
        status: 'sent',
      })
    } catch {
      failMessage(activeChat, tempId)
    }
  }

  const handleTyping = (isTyping: boolean) => {
    if (!activeChat) return
    send(isTyping ? 'typing.start' : 'typing.stop', { chat_id: activeChat })
  }

  const typingInChat = activeChat
    ? (typingUsers[activeChat] ?? [])
        .filter((uid) => uid !== user?.id)
        .map((uid) => ({ user_id: uid, display_name: uid }))
    : []

  return (
    <ChatLayout>
      {connectionState === 'reconnecting' && (
        <div className="text-center text-xs text-yellow-400 bg-yellow-900/20 py-1">
          Reconnecting…
        </div>
      )}

      {activeChat ? (
        <div className="flex flex-col flex-1 min-h-0">
          <div data-testid="chat-header" className="px-4 py-3 border-b border-gray-800 font-medium text-sm text-gray-200">
            {getChatDisplayName(activeChats[activeChat], user?.id)}
          </div>
          <MessageList chatId={activeChat} />
          <TypingIndicator typingUsers={typingInChat} />
          <MessageInput onSend={handleSend} onTyping={handleTyping} />
        </div>
      ) : (
        <div className="flex-1 flex items-center justify-center text-gray-500 text-sm">
          Select a chat to start messaging
        </div>
      )}
    </ChatLayout>
  )
}
