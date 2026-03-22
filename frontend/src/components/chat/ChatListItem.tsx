import type { Chat } from '@/types'
import { useAuthStore } from '@/stores/authStore'
import { relativeTime } from '@/utils/time'

interface Props {
  chat: Chat & { unreadCount?: number }
  isActive: boolean
  onClick: () => void
}

function chatDisplayName(chat: Chat, currentUserId: string | undefined): string {
  if (chat.name) return chat.name
  if (chat.type === 'direct') {
    const other = chat.members.find((m) => m.user_id !== currentUserId)
    if (other) return other.display_name || other.username
    // fallback: show first member if current user not found (e.g. self-chat)
    if (chat.members.length > 0) return chat.members[0].display_name || chat.members[0].username
  }
  if (chat.members.length > 0) return chat.members[0].display_name || chat.members[0].username
  return 'Chat'
}

export function ChatListItem({ chat, isActive, onClick }: Props) {
  const currentUserId = useAuthStore((s) => s.user?.id)
  const name = chatDisplayName(chat, currentUserId)

  return (
    <li
      data-testid="chat-item"
      onClick={onClick}
      className={`flex items-center gap-3 px-3 py-3 rounded-xl cursor-pointer transition select-none ${
        isActive ? 'bg-blue-100 dark:bg-blue-900/40' : 'hover:bg-gray-100 dark:hover:bg-gray-800'
      }`}
    >
      <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-semibold text-sm flex-shrink-0">
        {name[0]?.toUpperCase()}
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between">
          <span className="font-medium text-sm text-gray-900 dark:text-white truncate">{name}</span>
          {chat.last_message && (
            <span className="text-xs text-gray-400 flex-shrink-0 ml-2">
              {relativeTime(chat.last_message.timestamp)}
            </span>
          )}
        </div>

        {chat.last_message && (
          <p className="text-xs text-gray-500 dark:text-gray-400 truncate mt-0.5">
            {chat.last_message.content}
          </p>
        )}
      </div>

      {(chat.unreadCount ?? 0) > 0 && (
        <span className="bg-blue-500 text-white text-xs font-bold rounded-full min-w-[1.25rem] h-5 flex items-center justify-center px-1 flex-shrink-0">
          {chat.unreadCount}
        </span>
      )}
    </li>
  )
}
