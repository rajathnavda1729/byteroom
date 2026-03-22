import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { MessageBubble } from './MessageBubble'

describe('MessageBubble', () => {
  it('renders own messages with blue style', () => {
    render(
      <MessageBubble
        message={{ sender_id: 'user-1', content: 'My message', content_type: 'markdown' }}
        isOwnMessage={true}
      />,
    )

    const li = screen.getByRole('listitem')
    expect(li.innerHTML).toContain('bg-blue-500')
  })

  it('renders other messages with sender info', () => {
    render(
      <MessageBubble
        message={{
          sender_id: 'user-2',
          sender: { display_name: 'Bob' },
          content: 'Their message',
          content_type: 'markdown',
        }}
        isOwnMessage={false}
      />,
    )

    expect(screen.getByText('Bob')).toBeInTheDocument()
    expect(screen.getByText('Their message')).toBeInTheDocument()
  })

  it('shows pending status for optimistic messages', () => {
    render(
      <MessageBubble
        message={{ content: 'Sending...', status: 'pending', content_type: 'markdown' }}
        isOwnMessage={true}
      />,
    )

    expect(screen.getAllByText(/sending/i).length).toBeGreaterThan(0)
  })

  it('shows error status and retry button for failed messages', () => {
    const onRetry = vi.fn()

    render(
      <MessageBubble
        message={{ content: 'Oops', status: 'error', content_type: 'markdown' }}
        isOwnMessage={true}
        onRetry={onRetry}
      />,
    )

    expect(screen.getAllByText(/failed/i).length).toBeGreaterThan(0)
    expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument()
  })

  it('calls onRetry when retry button clicked', async () => {
    const onRetry = vi.fn()
    const user = userEvent.setup()
    const msg = { content: 'Failed', status: 'error' as const, content_type: 'markdown' as const }

    render(<MessageBubble message={msg} isOwnMessage={true} onRetry={onRetry} />)

    await user.click(screen.getByRole('button', { name: /retry/i }))
    expect(onRetry).toHaveBeenCalledWith(msg)
  })
})
