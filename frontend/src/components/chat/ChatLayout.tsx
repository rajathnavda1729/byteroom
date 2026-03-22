import { useState } from 'react'
import { useShallow } from 'zustand/react/shallow'
import { useChatStore } from '@/stores/chatStore'
import { useAuthStore } from '@/stores/authStore'
import { ChatListItem } from './ChatListItem'
import { NewChatModal } from './NewChatModal'

interface Props {
  children?: React.ReactNode
}

export function ChatLayout({ children }: Props) {
  const chats = useChatStore(useShallow((s) => Object.values(s.chats)))
  const activeChat = useChatStore((s) => s.activeChat)
  const setActiveChat = useChatStore((s) => s.setActiveChat)
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [showNewChat, setShowNewChat] = useState(false)

  return (
    <div className="flex h-screen bg-gray-950 text-white overflow-hidden">
      {/* Sidebar */}
      <aside
        className={`${sidebarOpen ? 'w-72' : 'w-0 overflow-hidden'} flex flex-col bg-gray-900 border-r border-gray-800 transition-all duration-200 flex-shrink-0`}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-4 border-b border-gray-800">
          <span className="font-bold text-lg">ByteRoom</span>
          <div className="flex items-center gap-1">
            <button
              aria-label="New conversation"
              onClick={() => setShowNewChat(true)}
              className="p-1.5 rounded-lg hover:bg-gray-800 text-gray-400 hover:text-white transition text-base"
              title="New conversation"
            >
              ✏️
            </button>
            <button
              aria-label="Toggle sidebar"
              onClick={() => setSidebarOpen(false)}
              className="p-1 rounded hover:bg-gray-800 text-gray-400"
            >
              ✕
            </button>
          </div>
        </div>

        {/* Chat list */}
        <ul data-testid="chat-list" className="flex-1 overflow-y-auto p-2 space-y-1">
          {chats.map((chat) => (
            <ChatListItem
              key={chat.chat_id}
              chat={chat}
              isActive={activeChat === chat.chat_id}
              onClick={() => setActiveChat(chat.chat_id)}
            />
          ))}
          {chats.length === 0 && (
            <li className="text-center text-gray-500 text-sm py-8">
              <p>No chats yet</p>
              <button
                onClick={() => setShowNewChat(true)}
                className="mt-2 text-blue-400 hover:text-blue-300 text-xs underline"
              >
                Start a conversation
              </button>
            </li>
          )}
        </ul>

        {/* User profile */}
        <div data-testid="user-menu" className="px-4 py-3 border-t border-gray-800 flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white text-sm font-semibold flex-shrink-0">
            {user?.displayName?.[0]?.toUpperCase() ?? '?'}
          </div>
          <span className="flex-1 text-sm font-medium truncate">{user?.displayName ?? user?.username}</span>
          <button
            onClick={logout}
            className="text-xs text-gray-400 hover:text-white transition"
            aria-label="Logout"
          >
            Logout
          </button>
        </div>
      </aside>

      {showNewChat && <NewChatModal onClose={() => setShowNewChat(false)} />}

      {/* Main content */}
      <main data-testid="chat-main" className="flex-1 flex flex-col min-w-0">
        {!sidebarOpen && (
          <button
            onClick={() => setSidebarOpen(true)}
            className="absolute top-4 left-4 z-10 p-2 bg-gray-800 rounded-lg hover:bg-gray-700"
            aria-label="Open sidebar"
          >
            ☰
          </button>
        )}
        {children}
      </main>
    </div>
  )
}
