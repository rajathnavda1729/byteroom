import type { Message } from '@/types'
import { MarkdownRenderer } from './MarkdownRenderer'

interface Props {
  message: Partial<Message>
  isOwnMessage: boolean
  onRetry?: (message: Partial<Message>) => void
}

export function MessageBubble({ message, isOwnMessage, onRetry }: Props) {
  const isPending = message.status === 'pending'
  const isError = message.status === 'error' || message.status === 'failed'

  return (
    <li
      data-testid="message-bubble"
      className={`flex ${isOwnMessage ? 'justify-end' : 'justify-start'} mb-2`}
      aria-label="message"
    >
      <div className={`max-w-[70%] ${isOwnMessage ? 'items-end' : 'items-start'} flex flex-col`}>
        {!isOwnMessage && message.sender && (
          <span className="text-xs text-gray-400 mb-1 px-1">{message.sender.display_name}</span>
        )}

        <div
          className={`px-4 py-2 rounded-2xl text-sm leading-relaxed ${
            isOwnMessage
              ? 'bg-blue-500 text-white rounded-br-sm'
              : 'bg-gray-800 text-gray-100 rounded-bl-sm'
          } ${isPending ? 'opacity-60' : ''} ${isError ? 'border border-red-500' : ''}`}
        >
          {message.content_type === 'markdown' || !message.content_type ? (
            <MarkdownRenderer content={message.content ?? ''} />
          ) : (
            <span>{message.content}</span>
          )}
        </div>

        {(isPending || isError) && (
          <div className="flex items-center gap-2 mt-1 px-1">
            {isPending && <span className="text-xs text-gray-400">Sending…</span>}
            {isError && (
              <>
                <span className="text-xs text-red-400">Failed to send</span>
                {onRetry && (
                  <button
                    onClick={() => onRetry(message)}
                    className="text-xs text-blue-400 hover:text-blue-300"
                    aria-label="Retry"
                  >
                    Retry
                  </button>
                )}
              </>
            )}
          </div>
        )}
      </div>
    </li>
  )
}
