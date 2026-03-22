import { render, screen } from '@testing-library/react'
import { describe, it, expect, beforeEach } from 'vitest'
import { MessageList } from './MessageList'
import { useChatStore } from '@/stores/chatStore'
import { useAuthStore } from '@/stores/authStore'

describe('MessageList', () => {
  beforeEach(() => {
    useAuthStore.getState().setAuth({
      user: { id: 'user-1', username: 'alice', displayName: 'Alice' },
      token: 'token',
    })
    useChatStore.getState().reset()
  })

  it('renders messages in chronological order', () => {
    useChatStore.getState().loadMessages('chat-1', [
      {
        message_id: 'msg-1',
        chat_id: 'chat-1',
        sender_id: 'user-1',
        content_type: 'markdown',
        content: 'First',
        timestamp: '2026-03-21T10:00:00Z',
      },
      {
        message_id: 'msg-2',
        chat_id: 'chat-1',
        sender_id: 'user-2',
        content_type: 'markdown',
        content: 'Second',
        timestamp: '2026-03-21T10:01:00Z',
      },
    ])

    render(<MessageList chatId="chat-1" />)

    const messages = screen.getAllByRole('listitem')
    expect(messages[0]).toHaveTextContent('First')
    expect(messages[1]).toHaveTextContent('Second')
  })

  it('groups messages by date', () => {
    const today = new Date()
    const yesterday = new Date(today)
    yesterday.setDate(today.getDate() - 2) // 2 days ago to avoid timezone edge cases

    useChatStore.getState().loadMessages('chat-1', [
      {
        message_id: 'msg-1',
        chat_id: 'chat-1',
        sender_id: 'user-1',
        content_type: 'markdown',
        content: 'Older message',
        timestamp: yesterday.toISOString(),
      },
      {
        message_id: 'msg-2',
        chat_id: 'chat-1',
        sender_id: 'user-2',
        content_type: 'markdown',
        content: 'Newer message',
        timestamp: today.toISOString(),
      },
    ])

    render(<MessageList chatId="chat-1" />)

    // Should show at least one date separator
    expect(screen.getByText('Today')).toBeInTheDocument()
  })

  it('shows loading state when fetching history', () => {
    render(<MessageList chatId="chat-1" isLoading />)
    expect(screen.getByRole('status')).toBeInTheDocument()
  })

  it('shows empty state for new chats', () => {
    render(<MessageList chatId="chat-1" />)
    expect(screen.getByText(/no messages yet/i)).toBeInTheDocument()
  })
})
