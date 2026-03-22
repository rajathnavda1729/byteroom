import { useEffect, useRef } from 'react'
import type { Message } from '@/types'
import { useChatStore } from '@/stores/chatStore'
import { useAuthStore } from '@/stores/authStore'
import { MessageBubble } from './MessageBubble'
import { dayLabel } from '@/utils/time'

const EMPTY: Message[] = []

interface Props {
  chatId: string
  isLoading?: boolean
  onRetry?: (message: Partial<Message>) => void
}

function groupByDay(messages: Message[]): { date: string; messages: Message[] }[] {
  const groups: Map<string, Message[]> = new Map()
  for (const msg of messages) {
    const key = new Date(msg.timestamp).toDateString()
    if (!groups.has(key)) groups.set(key, [])
    groups.get(key)!.push(msg)
  }
  return Array.from(groups.entries()).map(([, msgs]) => ({
    date: msgs[0].timestamp,
    messages: msgs,
  }))
}

export function MessageList({ chatId, isLoading, onRetry }: Props) {
  const messages = useChatStore((s) => s.messages[chatId] ?? EMPTY)
  const currentUserId = useAuthStore((s) => s.user?.id)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages.length])

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center" role="status" aria-label="Loading">
        <div className="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
      </div>
    )
  }

  if (messages.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center text-gray-500 text-sm">
        No messages yet
      </div>
    )
  }

  const groups = groupByDay(messages)

  return (
    <div data-testid="message-list" className="flex-1 overflow-y-auto px-4 py-4">
      {groups.map((group) => (
        <div key={group.date}>
          <div className="flex items-center gap-3 my-4">
            <div className="flex-1 h-px bg-gray-800" />
            <span className="text-xs text-gray-500 flex-shrink-0">{dayLabel(group.date)}</span>
            <div className="flex-1 h-px bg-gray-800" />
          </div>

          <ul className="space-y-1">
            {group.messages.map((msg) => (
              <MessageBubble
                key={msg.message_id}
                message={msg}
                isOwnMessage={msg.sender_id === currentUserId}
                onRetry={onRetry}
              />
            ))}
          </ul>
        </div>
      ))}

      <div ref={bottomRef} />
    </div>
  )
}
