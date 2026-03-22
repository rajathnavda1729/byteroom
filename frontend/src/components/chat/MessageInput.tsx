import { useRef, useState, useCallback } from 'react'

interface Props {
  onSend: (content: string) => void
  onTyping?: (isTyping: boolean) => void
  disabled?: boolean
}

export function MessageInput({ onSend, onTyping, disabled }: Props) {
  const [value, setValue] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const typingTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const autoResize = useCallback(() => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = `${el.scrollHeight}px`
  }, [])

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setValue(e.target.value)
    autoResize()

    if (onTyping) {
      onTyping(true)
      if (typingTimerRef.current) clearTimeout(typingTimerRef.current)
      typingTimerRef.current = setTimeout(() => onTyping(false), 3000)
    }
  }

  const handleSend = () => {
    const trimmed = value.trim()
    if (!trimmed || disabled) return

    onSend(trimmed)
    setValue('')

    if (onTyping) {
      if (typingTimerRef.current) clearTimeout(typingTimerRef.current)
      onTyping(false)
    }

    // Reset height
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const isEmpty = value.trim().length === 0

  return (
    <div className="flex items-end gap-3 px-4 py-3 bg-gray-900 border-t border-gray-800">
      <textarea
        ref={textareaRef}
        data-testid="message-input"
        value={value}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        disabled={disabled}
        rows={1}
        className="flex-1 bg-gray-800 border border-gray-700 rounded-xl px-4 py-2.5 text-white placeholder-gray-500 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition max-h-36 overflow-y-auto"
        placeholder="Message…"
        aria-label="Message"
      />

      <button
        onClick={handleSend}
        disabled={isEmpty || disabled}
        data-testid="send-button"
        aria-label="Send"
        className="flex-shrink-0 w-10 h-10 bg-blue-600 hover:bg-blue-500 disabled:bg-gray-700 disabled:cursor-not-allowed rounded-xl flex items-center justify-center transition"
      >
        <svg className="w-5 h-5 text-white rotate-90" fill="currentColor" viewBox="0 0 24 24">
          <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" />
        </svg>
      </button>
    </div>
  )
}
