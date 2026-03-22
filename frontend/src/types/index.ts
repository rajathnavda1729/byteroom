export type ContentType = 'markdown' | 'diagram_state' | 'image'
export type MessageStatus = 'pending' | 'sent' | 'delivered' | 'failed' | 'error'
export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'reconnecting'

export interface Message {
  message_id: string
  chat_id: string
  sender_id: string
  content_type: ContentType
  content: string
  timestamp: string
  status?: MessageStatus
  sender?: {
    display_name: string
    avatar_url?: string | null
  }
}

export interface ChatMember {
  user_id: string
  username: string
  display_name: string
  avatar_url?: string | null
}

export interface LastMessage {
  content: string
  sender_name: string
  timestamp: string
}

export interface Chat {
  chat_id: string
  name: string | null
  type: 'direct' | 'group'
  members: ChatMember[]
  created_at?: string
  last_message?: LastMessage | null
  unreadCount?: number
}

export interface User {
  id: string
  username: string
  displayName: string
  avatarUrl?: string | null
}

export type WSEvent =
  | { event: 'message.send'; data: Omit<Message, 'timestamp' | 'status' | 'sender'> }
  | { event: 'message.ack'; data: { message_id: string; status: boolean; request_id?: string } }
  | { event: 'message.new'; data: Message }
  | { event: 'message.error'; data: { error: string; request_id?: string } }
  | { event: 'user.typing'; data: { chat_id: string; user_id: string; username?: string; is_typing: boolean } }
  | { event: 'chat.new'; data: Chat }
  | { event: 'ping'; data: Record<string, never> }
  | { event: 'pong'; data: Record<string, never> }
