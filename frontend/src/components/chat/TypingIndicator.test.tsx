import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { TypingIndicator } from './TypingIndicator'

describe('TypingIndicator', () => {
  it('shows nothing when no one is typing', () => {
    const { container } = render(<TypingIndicator typingUsers={[]} />)
    expect(container).toBeEmptyDOMElement()
  })

  it('shows single user typing', () => {
    render(<TypingIndicator typingUsers={[{ user_id: 'u1', display_name: 'Alice' }]} />)
    expect(screen.getByText(/alice is typing/i)).toBeInTheDocument()
  })

  it('shows two users typing', () => {
    render(
      <TypingIndicator
        typingUsers={[
          { user_id: 'u1', display_name: 'Alice' },
          { user_id: 'u2', display_name: 'Bob' },
        ]}
      />,
    )
    expect(screen.getByText(/alice and bob are typing/i)).toBeInTheDocument()
  })

  it('shows multiple users typing', () => {
    render(
      <TypingIndicator
        typingUsers={[
          { user_id: 'u1', display_name: 'Alice' },
          { user_id: 'u2', display_name: 'Bob' },
          { user_id: 'u3', display_name: 'Charlie' },
        ]}
      />,
    )
    expect(screen.getByText(/3 people are typing/i)).toBeInTheDocument()
  })

  it('shows animated dots', () => {
    render(<TypingIndicator typingUsers={[{ user_id: 'u1', display_name: 'Alice' }]} />)
    expect(screen.getByTestId('typing-dots')).toBeInTheDocument()
  })
})
