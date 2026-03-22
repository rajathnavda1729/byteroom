import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { ChatListItem } from './ChatListItem'

describe('ChatListItem', () => {
  const base = {
    chat_id: 'chat-1',
    name: 'Test Chat',
    type: 'group' as const,
    members: [],
    isActive: false,
    onClick: vi.fn(),
  }

  it('renders chat name', () => {
    render(<ChatListItem chat={base} isActive={false} onClick={vi.fn()} />)
    expect(screen.getByText('Test Chat')).toBeInTheDocument()
  })

  it('displays unread count badge', () => {
    render(<ChatListItem chat={{ ...base, unreadCount: 5 }} isActive={false} onClick={vi.fn()} />)
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('displays relative time for last message', () => {
    const recentTime = new Date(Date.now() - 5 * 60 * 1000).toISOString()

    render(
      <ChatListItem
        chat={{
          ...base,
          last_message: { timestamp: recentTime, content: 'Hi', sender_name: 'Alice' },
        }}
        isActive={false}
        onClick={vi.fn()}
      />,
    )

    expect(screen.getByText(/5m/)).toBeInTheDocument()
  })

  it('shows last message preview', () => {
    const recentTime = new Date().toISOString()
    render(
      <ChatListItem
        chat={{
          ...base,
          last_message: { timestamp: recentTime, content: 'Last msg', sender_name: 'Alice' },
        }}
        isActive={false}
        onClick={vi.fn()}
      />,
    )
    expect(screen.getByText('Last msg')).toBeInTheDocument()
  })

  it('uses member name for direct chat without name', () => {
    render(
      <ChatListItem
        chat={{
          chat_id: 'chat-2',
          name: null,
          type: 'direct',
          members: [{ user_id: 'u2', username: 'bob', display_name: 'Bob' }],
        }}
        isActive={false}
        onClick={vi.fn()}
      />,
    )
    expect(screen.getByText('Bob')).toBeInTheDocument()
  })

  it('calls onClick when clicked', async () => {
    const onClick = vi.fn()
    const user = userEvent.setup()

    render(<ChatListItem chat={base} isActive={false} onClick={onClick} />)
    await user.click(screen.getByText('Test Chat'))

    expect(onClick).toHaveBeenCalledOnce()
  })

  it('applies active style when isActive is true', () => {
    render(<ChatListItem chat={base} isActive={true} onClick={vi.fn()} />)
    const li = screen.getByRole('listitem')
    expect(li.className).toContain('bg-blue-100')
  })
})
