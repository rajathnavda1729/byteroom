import { describe, it, expect, beforeEach } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/test/mocks/server'
import { api } from './api'

describe('API Client', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('sends request without auth header when no token', async () => {
    let captured: string | null = null

    server.use(
      http.get('/api/test', ({ request }) => {
        captured = request.headers.get('Authorization')
        return HttpResponse.json({ ok: true })
      }),
    )

    await api.get('/api/test')
    expect(captured).toBeNull()
  })

  it('adds Authorization header when token exists', async () => {
    let captured: string | null = null

    server.use(
      http.get('/api/test', ({ request }) => {
        captured = request.headers.get('Authorization')
        return HttpResponse.json({ ok: true })
      }),
    )

    localStorage.setItem('token', 'test-token')
    await api.get('/api/test')

    expect(captured).toBe('Bearer test-token')
  })

  it('handles 401 by removing token from localStorage', async () => {
    server.use(
      http.get('/api/test', () => {
        return HttpResponse.json({ error: 'unauthorized' }, { status: 401 })
      }),
    )

    localStorage.setItem('token', 'expired-token')
    await expect(api.get('/api/test')).rejects.toThrow()
    expect(localStorage.getItem('token')).toBeNull()
  })

  it('throws APIError on non-2xx response', async () => {
    server.use(
      http.post('/api/test', () => {
        return HttpResponse.json({ message: 'Bad request' }, { status: 400 })
      }),
    )

    await expect(api.post('/api/test', {})).rejects.toThrow('Bad request')
  })

  it('posts JSON body', async () => {
    let capturedBody: unknown = null

    server.use(
      http.post('/api/test', async ({ request }) => {
        capturedBody = await request.json()
        return HttpResponse.json({ ok: true })
      }),
    )

    await api.post('/api/test', { name: 'alice' })
    expect(capturedBody).toEqual({ name: 'alice' })
  })
})
