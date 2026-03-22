import { useState, useEffect, useRef } from 'react'
import { api } from '@/services/api'
import { useChatStore } from '@/stores/chatStore'
import type { Chat } from '@/types'

interface UserResult {
  user_id: string
  username: string
  display_name: string
  avatar_url?: string | null
}

interface SearchResponse {
  users: UserResult[]
}

interface CreateChatResponse {
  chat_id: string
  name: string
  type: 'direct' | 'group'
  members: string[]
  created_by: string
  created_at: string
}

type Tab = 'direct' | 'group'

interface Props {
  onClose: () => void
}

export function NewChatModal({ onClose }: Props) {
  const [tab, setTab] = useState<Tab>('direct')

  // Direct chat state
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<UserResult[]>([])
  const [searching, setSearching] = useState(false)

  // Group chat state
  const [groupName, setGroupName] = useState('')
  const [memberQuery, setMemberQuery] = useState('')
  const [memberResults, setMemberResults] = useState<UserResult[]>([])
  const [selectedMembers, setSelectedMembers] = useState<UserResult[]>([])

  const [creating, setCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { addChat, setActiveChat } = useChatStore()
  const searchTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const memberTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Debounced user search for direct chat
  useEffect(() => {
    if (searchTimer.current) clearTimeout(searchTimer.current)
    if (query.length < 2) { setResults([]); return }
    searchTimer.current = setTimeout(async () => {
      setSearching(true)
      try {
        const res = await api.get<SearchResponse>(`/api/users/search?q=${encodeURIComponent(query)}`)
        setResults(res.users ?? [])
      } finally {
        setSearching(false)
      }
    }, 300)
    return () => { if (searchTimer.current) clearTimeout(searchTimer.current) }
  }, [query])

  // Debounced user search for group member picker
  useEffect(() => {
    if (memberTimer.current) clearTimeout(memberTimer.current)
    if (memberQuery.length < 2) { setMemberResults([]); return }
    memberTimer.current = setTimeout(async () => {
      const res = await api.get<SearchResponse>(`/api/users/search?q=${encodeURIComponent(memberQuery)}`)
      setMemberResults((res.users ?? []).filter(
        (u) => !selectedMembers.some((m) => m.user_id === u.user_id),
      ))
    }, 300)
    return () => { if (memberTimer.current) clearTimeout(memberTimer.current) }
  }, [memberQuery, selectedMembers])

  const openChat = (raw: CreateChatResponse) => {
    const chat: Chat = {
      chat_id: raw.chat_id,
      name: raw.name,
      type: raw.type,
      members: [],
    }
    addChat(chat)
    setActiveChat(raw.chat_id)
    onClose()
  }

  const startDirectChat = async (user: UserResult) => {
    setCreating(true)
    setError(null)
    try {
      const res = await api.post<CreateChatResponse>('/api/chats', {
        type: 'direct',
        member_ids: [user.user_id],
      })
      openChat(res)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create chat')
    } finally {
      setCreating(false)
    }
  }

  const createGroupChat = async () => {
    if (!groupName.trim()) { setError('Group name is required'); return }
    setCreating(true)
    setError(null)
    try {
      const res = await api.post<CreateChatResponse>('/api/chats', {
        type: 'group',
        name: groupName.trim(),
        member_ids: selectedMembers.map((m) => m.user_id),
      })
      openChat(res)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create group')
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="bg-gray-900 border border-gray-700 rounded-2xl w-full max-w-md shadow-2xl overflow-hidden">

        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-gray-800">
          <h2 className="font-semibold text-white text-base">New Conversation</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition text-xl leading-none"
            aria-label="Close"
          >
            ✕
          </button>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-gray-800">
          {(['direct', 'group'] as Tab[]).map((t) => (
            <button
              key={t}
              onClick={() => { setTab(t); setError(null) }}
              className={`flex-1 py-2.5 text-sm font-medium transition ${
                tab === t
                  ? 'text-blue-400 border-b-2 border-blue-500'
                  : 'text-gray-400 hover:text-gray-200'
              }`}
            >
              {t === 'direct' ? '💬 Direct Message' : '👥 Group Chat'}
            </button>
          ))}
        </div>

        <div className="p-5 space-y-4">
          {error && (
            <p className="text-xs text-red-400 bg-red-900/20 border border-red-800 rounded-lg px-3 py-2">
              {error}
            </p>
          )}

          {tab === 'direct' && (
            <>
              <input
                autoFocus
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search by username…"
                className="w-full bg-gray-800 border border-gray-700 rounded-xl px-4 py-2.5 text-sm text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />

              {searching && (
                <p className="text-xs text-gray-400 text-center">Searching…</p>
              )}

              {!searching && results.length === 0 && query.length >= 2 && (
                <p className="text-xs text-gray-500 text-center">No users found</p>
              )}

              <ul className="space-y-1 max-h-60 overflow-y-auto">
                {results.map((u) => (
                  <li key={u.user_id}>
                    <button
                      disabled={creating}
                      onClick={() => startDirectChat(u)}
                      className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-gray-800 transition text-left"
                    >
                      <div className="w-9 h-9 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white text-sm font-semibold flex-shrink-0">
                        {(u.display_name || u.username)[0].toUpperCase()}
                      </div>
                      <div>
                        <p className="text-sm font-medium text-white">{u.display_name || u.username}</p>
                        <p className="text-xs text-gray-400">@{u.username}</p>
                      </div>
                    </button>
                  </li>
                ))}
              </ul>
            </>
          )}

          {tab === 'group' && (
            <>
              <input
                autoFocus
                type="text"
                value={groupName}
                onChange={(e) => setGroupName(e.target.value)}
                placeholder="Group name…"
                className="w-full bg-gray-800 border border-gray-700 rounded-xl px-4 py-2.5 text-sm text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />

              {/* Selected members */}
              {selectedMembers.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {selectedMembers.map((m) => (
                    <span
                      key={m.user_id}
                      className="flex items-center gap-1.5 bg-blue-600/30 border border-blue-600/50 text-blue-300 text-xs px-2.5 py-1 rounded-full"
                    >
                      {m.display_name || m.username}
                      <button
                        onClick={() => setSelectedMembers((prev) => prev.filter((x) => x.user_id !== m.user_id))}
                        className="text-blue-400 hover:text-white leading-none"
                        aria-label={`Remove ${m.username}`}
                      >
                        ✕
                      </button>
                    </span>
                  ))}
                </div>
              )}

              <input
                type="text"
                value={memberQuery}
                onChange={(e) => setMemberQuery(e.target.value)}
                placeholder="Add members by username…"
                className="w-full bg-gray-800 border border-gray-700 rounded-xl px-4 py-2.5 text-sm text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />

              <ul className="space-y-1 max-h-40 overflow-y-auto">
                {memberResults.map((u) => (
                  <li key={u.user_id}>
                    <button
                      onClick={() => {
                        setSelectedMembers((prev) => [...prev, u])
                        setMemberQuery('')
                        setMemberResults([])
                      }}
                      className="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-gray-800 transition text-left"
                    >
                      <div className="w-8 h-8 rounded-full bg-gradient-to-br from-green-500 to-teal-600 flex items-center justify-center text-white text-xs font-semibold flex-shrink-0">
                        {(u.display_name || u.username)[0].toUpperCase()}
                      </div>
                      <div>
                        <p className="text-sm font-medium text-white">{u.display_name || u.username}</p>
                        <p className="text-xs text-gray-400">@{u.username}</p>
                      </div>
                    </button>
                  </li>
                ))}
              </ul>

              <button
                onClick={createGroupChat}
                disabled={creating || !groupName.trim()}
                className="w-full bg-blue-600 hover:bg-blue-500 disabled:bg-gray-700 disabled:cursor-not-allowed text-white text-sm font-medium py-2.5 rounded-xl transition"
              >
                {creating ? 'Creating…' : 'Create Group'}
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
