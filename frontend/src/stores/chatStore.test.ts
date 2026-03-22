import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, beforeEach } from 'vitest'
import { useChatStore } from './chatStore'

describe('chatStore', () => {
  beforeEach(() => {
    useChatStore.getState().reset()
  })

  it('adds chat to store', () => {
    const { result } = renderHook(() => useChatStore())

    act(() => {
      result.current.addChat({
        chat_id: 'chat-1',
        name: 'Tech Discussion',
        type: 'group',
        members: [],
        last_message: null,
      })
    })

    expect(result.current.chats['chat-1']).toBeDefined()
    expect(result.current.chats['chat-1'].name).toBe('Tech Discussion')
  })

  it('sets active chat', () => {
    const { result } = renderHook(() => useChatStore())

    act(() => {
      result.current.setActiveChat('chat-1')
    })

    expect(result.current.activeChat).toBe('chat-1')
  })

  it('adds message to chat', () => {
    const { result } = renderHook(() => useChatStore())

    act(() => {
      result.current.addMessage('chat-1', {
        message_id: 'msg-1',
        chat_id: 'chat-1',
        sender_id: 'user-1',
        content_type: 'markdown',
        content: 'Hello!',
        timestamp: new Date().toISOString(),
        status: 'sent',
      })
    })

    expect(result.current.messages['chat-1']).toHaveLength(1)
    expect(result.current.messages['chat-1'][0].content).toBe('Hello!')
  })

  it('does not add duplicate messages', () => {
    const { result } = renderHook(() => useChatStore())
    const msg = {
      message_id: 'msg-1',
      chat_id: 'chat-1',
      sender_id: 'user-1',
      content_type: 'markdown' as const,
      content: 'Hello!',
      timestamp: new Date().toISOString(),
    }

    act(() => {
      result.current.addMessage('chat-1', msg)
      result.current.addMessage('chat-1', msg)
    })

    expect(result.current.messages['chat-1']).toHaveLength(1)
  })

  it('adds optimistic message with pending status', () => {
    const { result } = renderHook(() => useChatStore())

    act(() => {
      result.current.addOptimisticMessage('chat-1', {
        message_id: 'temp-1',
        chat_id: 'chat-1',
        sender_id: 'user-1',
        content_type: 'markdown',
        content: 'Sending...',
        timestamp: new Date().toISOString(),
        status: 'pending',
      })
    })

    expect(result.current.messages['chat-1'][0].status).toBe('pending')
  })

  it('confirms optimistic message', () => {
    const { result } = renderHook(() => useChatStore())

    act(() => {
      result.current.addOptimisticMessage('chat-1', {
        message_id: 'temp-1',
        chat_id: 'chat-1',
        sender_id: 'user-1',
        content_type: 'markdown',
        content: 'Hello!',
        timestamp: new Date().toISOString(),
        status: 'pending',
      })

      result.current.confirmMessage('chat-1', 'temp-1', {
        message_id: 'msg-1',
        chat_id: 'chat-1',
        sender_id: 'user-1',
        content_type: 'markdown',
        content: 'Hello!',
        timestamp: new Date().toISOString(),
        status: 'sent',
      })
    })

    expect(result.current.messages['chat-1'][0].status).toBe('sent')
    expect(result.current.messages['chat-1'][0].message_id).toBe('msg-1')
  })

  it('tracks typing users', () => {
    const { result } = renderHook(() => useChatStore())

    act(() => {
      result.current.setTyping('chat-1', 'user-2', true)
    })

    expect(result.current.typingUsers['chat-1']).toContain('user-2')

    act(() => {
      result.current.setTyping('chat-1', 'user-2', false)
    })

    expect(result.current.typingUsers['chat-1'] ?? []).not.toContain('user-2')
  })

  it('loads messages for a chat', () => {
    const { result } = renderHook(() => useChatStore())

    act(() => {
      result.current.loadMessages('chat-1', [
        {
          message_id: 'msg-1',
          chat_id: 'chat-1',
          sender_id: 'user-1',
          content_type: 'markdown',
          content: 'Hello!',
          timestamp: new Date().toISOString(),
        },
      ])
    })

    expect(result.current.messages['chat-1']).toHaveLength(1)
  })

  it('resets all state', () => {
    const { result } = renderHook(() => useChatStore())

    act(() => {
      result.current.addChat({ chat_id: 'c1', name: 'Test', type: 'group', members: [] })
      result.current.setActiveChat('c1')
      result.current.reset()
    })

    expect(result.current.chats).toEqual({})
    expect(result.current.activeChat).toBeNull()
  })
})
