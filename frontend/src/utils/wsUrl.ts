/**
 * Resolves the WebSocket URL for the backend.
 *
 * In development we connect directly to the Go server (port 8080) instead of
 * routing through Vite's `/ws` proxy. Some Vite/http-proxy setups mishandle
 * server→client frames, which looks like "real-time never works until reload"
 * (history still loads over HTTP).
 *
 * Override with `VITE_WS_URL` (e.g. `wss://api.example.com`) in production
 * if the API is on another host.
 */
export function resolveWebSocketURL(path: string): string {
  const p = path.startsWith('/') ? path : `/${path}`
  const explicit = import.meta.env.VITE_WS_URL as string | undefined
  if (explicit?.trim()) {
    const base = explicit.replace(/\/$/, '')
    return `${base}${p}`
  }
  if (import.meta.env.DEV) {
    return `ws://127.0.0.1:8080${p}`
  }
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}${p}`
}
