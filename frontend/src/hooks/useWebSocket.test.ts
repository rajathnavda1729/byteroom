import { renderHook, act, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import WS from 'jest-websocket-mock'
import { useWebSocket } from './useWebSocket'

describe('useWebSocket', () => {
  let server: WS

  beforeEach(async () => {
    server = new WS('ws://localhost:8080/ws')
  })

  afterEach(() => {
    WS.clean()
  })

  it('connects to WebSocket server with token', async () => {
    renderHook(() =>
      useWebSocket({
        url: 'ws://localhost:8080/ws',
        token: 'test-token',
        onMessage: vi.fn(),
      }),
    )

    await server.connected
    expect(server).toBeDefined()
  })

  it('sends messages when connected', async () => {
    const { result } = renderHook(() =>
      useWebSocket({
        url: 'ws://localhost:8080/ws',
        token: 'test-token',
        onMessage: vi.fn(),
      }),
    )

    await server.connected

    act(() => {
      result.current.send('message.send', { content: 'Hello' })
    })

    await expect(server).toReceiveMessage(
      JSON.stringify({ event: 'message.send', data: { content: 'Hello' } }),
    )
  })

  it('calls onMessage when receiving data', async () => {
    const onMessage = vi.fn()

    renderHook(() =>
      useWebSocket({
        url: 'ws://localhost:8080/ws',
        token: 'test-token',
        onMessage,
      }),
    )

    await server.connected

    act(() => {
      server.send(JSON.stringify({ event: 'message.new', data: { content: 'Hi' } }))
    })

    await waitFor(() => {
      expect(onMessage).toHaveBeenCalledWith(
        expect.objectContaining({ event: 'message.new' }),
      )
    })
  })

  it('reports connected state after connection', async () => {
    const { result } = renderHook(() =>
      useWebSocket({
        url: 'ws://localhost:8080/ws',
        token: 'test-token',
        onMessage: vi.fn(),
      }),
    )

    await server.connected

    await waitFor(() => {
      expect(result.current.connectionState).toBe('connected')
    })
  })

  it('transitions to reconnecting on connection close', async () => {
    const { result } = renderHook(() =>
      useWebSocket({
        url: 'ws://localhost:8080/ws',
        token: 'test-token',
        onMessage: vi.fn(),
      }),
    )

    await server.connected

    act(() => {
      server.close()
    })

    await waitFor(() => {
      expect(result.current.connectionState).toBe('reconnecting')
    })
  })

  it('does not connect when token is null', () => {
    renderHook(() =>
      useWebSocket({
        url: 'ws://localhost:8080/ws',
        token: null,
        onMessage: vi.fn(),
      }),
    )

    expect(server.server.clients().length).toBe(0)
  })

  it('does not connect when disabled', () => {
    renderHook(() =>
      useWebSocket({
        url: 'ws://localhost:8080/ws',
        token: 'test-token',
        onMessage: vi.fn(),
        enabled: false,
      }),
    )

    expect(server.server.clients().length).toBe(0)
  })
})
