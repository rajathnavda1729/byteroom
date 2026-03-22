import { useEffect, useRef, useCallback, useState } from 'react'
import type { ConnectionStatus, WSEvent } from '@/types'

interface UseWebSocketOptions {
  url: string
  token: string | null
  onMessage: (event: WSEvent) => void
  enabled?: boolean
}

interface UseWebSocketReturn {
  send: (event: string, data: unknown, requestId?: string) => void
  connectionState: ConnectionStatus
}

const INITIAL_RETRY_DELAY = 1000
const MAX_RETRY_DELAY = 30_000
const PING_INTERVAL = 30_000

export function useWebSocket({
  url,
  token,
  onMessage,
  enabled = true,
}: UseWebSocketOptions): UseWebSocketReturn {
  const [connectionState, setConnectionState] = useState<ConnectionStatus>('connecting')
  const wsRef = useRef<WebSocket | null>(null)
  const pendingSendsRef = useRef<{ event: string; data: unknown; requestId?: string }[]>([])
  const retryDelayRef = useRef(INITIAL_RETRY_DELAY)
  const retryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const pingTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const isMountedRef = useRef(true)
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage

  const clearTimers = useCallback(() => {
    if (retryTimerRef.current) {
      clearTimeout(retryTimerRef.current)
      retryTimerRef.current = null
    }
    if (pingTimerRef.current) {
      clearInterval(pingTimerRef.current)
      pingTimerRef.current = null
    }
  }, [])

  const connect = useCallback(() => {
    if (!isMountedRef.current || !token || !enabled) return

    // Query strings treat '+' as space; JWTs can contain '+' — always encode.
    const wsUrl = token ? `${url}?token=${encodeURIComponent(token)}` : url
    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => {
      if (!isMountedRef.current) return
      retryDelayRef.current = INITIAL_RETRY_DELAY
      setConnectionState('connected')

      // Flush frames queued while connecting (typing, etc.)
      while (pendingSendsRef.current.length > 0 && ws.readyState === WebSocket.OPEN) {
        const p = pendingSendsRef.current.shift()!
        ws.send(
          JSON.stringify({
            event: p.event,
            data: p.data,
            ...(p.requestId ? { request_id: p.requestId } : {}),
          }),
        )
      }

      // Application-level heartbeat
      pingTimerRef.current = setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ event: 'ping', data: {} }))
        }
      }, PING_INTERVAL)
    }

    ws.onmessage = (ev) => {
      try {
        const raw = typeof ev.data === 'string' ? ev.data : ''
        const frame = JSON.parse(raw) as WSEvent
        if (import.meta.env.DEV && frame && typeof frame === 'object' && 'event' in frame) {
          const ev = (frame as { event?: string }).event
          if (ev === 'message.new' || ev === 'message.error' || ev === 'chat.new') {
            console.debug('[ws ←]', ev)
          }
        }
        onMessageRef.current(frame)
      } catch (e) {
        if (import.meta.env.DEV) console.warn('[ws] parse error', e)
      }
    }

    ws.onclose = () => {
      if (!isMountedRef.current) return
      clearTimers()
      setConnectionState('reconnecting')

      const delay = retryDelayRef.current
      retryDelayRef.current = Math.min(delay * 2, MAX_RETRY_DELAY)
      retryTimerRef.current = setTimeout(() => {
        if (isMountedRef.current) connect()
      }, delay)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [url, token, enabled, clearTimers])

  useEffect(() => {
    isMountedRef.current = true
    if (enabled && token) {
      connect()
    }
    return () => {
      isMountedRef.current = false
      clearTimers()
      if (wsRef.current) {
        wsRef.current.onclose = null // prevent reconnect on intentional close
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [connect, enabled, token, clearTimers])

  const send = useCallback((event: string, data: unknown, requestId?: string) => {
    const payload = JSON.stringify({ event, data, ...(requestId ? { request_id: requestId } : {}) })
    const ws = wsRef.current
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(payload)
    } else {
      pendingSendsRef.current.push({ event, data, requestId })
    }
  }, [])

  return { send, connectionState }
}
