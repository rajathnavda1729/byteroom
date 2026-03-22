import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { MessageInput } from './MessageInput'

describe('MessageInput', () => {
  it('renders textarea and send button', () => {
    render(<MessageInput onSend={vi.fn()} />)

    expect(screen.getByRole('textbox')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /send/i })).toBeInTheDocument()
  })

  it('calls onSend with content when button clicked', async () => {
    const onSend = vi.fn()
    const user = userEvent.setup()

    render(<MessageInput onSend={onSend} />)

    await user.type(screen.getByRole('textbox'), 'Hello world')
    await user.click(screen.getByRole('button', { name: /send/i }))

    expect(onSend).toHaveBeenCalledWith('Hello world')
  })

  it('clears input after sending', async () => {
    const user = userEvent.setup()

    render(<MessageInput onSend={vi.fn()} />)

    const input = screen.getByRole('textbox')
    await user.type(input, 'Hello')
    await user.click(screen.getByRole('button', { name: /send/i }))

    expect(input).toHaveValue('')
  })

  it('sends on Enter key (without shift)', async () => {
    const onSend = vi.fn()
    const user = userEvent.setup()

    render(<MessageInput onSend={onSend} />)

    await user.type(screen.getByRole('textbox'), 'Hello{enter}')

    expect(onSend).toHaveBeenCalledWith('Hello')
  })

  it('allows newline with Shift+Enter', async () => {
    const onSend = vi.fn()
    const user = userEvent.setup()

    render(<MessageInput onSend={onSend} />)

    await user.type(screen.getByRole('textbox'), 'Line 1{shift>}{enter}{/shift}Line 2')

    expect(onSend).not.toHaveBeenCalled()
    expect(screen.getByRole('textbox')).toHaveValue('Line 1\nLine 2')
  })

  it('disables send when input is empty', () => {
    render(<MessageInput onSend={vi.fn()} />)
    expect(screen.getByRole('button', { name: /send/i })).toBeDisabled()
  })

  it('emits typing events on input change', async () => {
    const onTyping = vi.fn()
    const user = userEvent.setup()

    render(<MessageInput onSend={vi.fn()} onTyping={onTyping} />)

    await user.type(screen.getByRole('textbox'), 'H')

    expect(onTyping).toHaveBeenCalledWith(true)
  })
})
