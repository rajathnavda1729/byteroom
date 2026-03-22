import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BrowserRouter } from 'react-router-dom'
import { describe, it, expect, beforeEach } from 'vitest'
import { ChatLayout } from './ChatLayout'
import { useChatStore } from '@/stores/chatStore'
import { useAuthStore } from '@/stores/authStore'

const wrap = (ui: React.ReactNode) => render(<BrowserRouter>{ui}</BrowserRouter>)

describe('ChatLayout', () => {
  beforeEach(() => {
    useChatStore.getState().reset()
    useAuthStore.getState().setAuth({
      user: { id: 'u1', username: 'alice', displayName: 'Alice' },
      token: 'tok',
    })
    useChatStore.getState().addChat({
      chat_id: 'chat-1',
      name: 'Tech Discussion',
      type: 'group',
      members: [],
      last_message: { content: 'Last msg', sender_name: 'Alice', timestamp: new Date().toISOString() },
    })
    useChatStore.getState().addChat({
      chat_id: 'chat-2',
      name: null,
      type: 'direct',
      members: [{ user_id: 'u2', username: 'bob', display_name: 'Bob' }],
      last_message: null,
    })
  })

  it('renders sidebar with chat list', () => {
    wrap(<ChatLayout />)
    expect(screen.getByText('Tech Discussion')).toBeInTheDocument()
    expect(screen.getByText('Bob')).toBeInTheDocument()
  })

  it('selects chat on click', async () => {
    wrap(<ChatLayout />)
    const user = userEvent.setup()

    await user.click(screen.getByText('Tech Discussion'))

    expect(useChatStore.getState().activeChat).toBe('chat-1')
  })

  it('shows last message preview', () => {
    wrap(<ChatLayout />)
    expect(screen.getByText('Last msg')).toBeInTheDocument()
  })

  it('highlights active chat', async () => {
    wrap(<ChatLayout />)
    const user = userEvent.setup()

    await user.click(screen.getByText('Tech Discussion'))

    const chatItem = screen.getByText('Tech Discussion').closest('li')
    expect(chatItem?.className).toContain('bg-blue-100')
  })
})
