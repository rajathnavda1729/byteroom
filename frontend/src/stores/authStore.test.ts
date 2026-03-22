import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, beforeEach } from 'vitest'
import { useAuthStore } from './authStore'

describe('authStore', () => {
  beforeEach(() => {
    useAuthStore.getState().logout()
    localStorage.clear()
  })

  it('initializes with null user', () => {
    const { result } = renderHook(() => useAuthStore())
    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
  })

  it('sets user on login', () => {
    const { result } = renderHook(() => useAuthStore())

    act(() => {
      result.current.setAuth({
        user: { id: 'user-1', username: 'alice', displayName: 'Alice' },
        token: 'jwt-token',
        refreshToken: 'refresh-token',
      })
    })

    expect(result.current.user?.username).toBe('alice')
    expect(result.current.isAuthenticated).toBe(true)
  })

  it('clears user on logout', () => {
    const { result } = renderHook(() => useAuthStore())

    act(() => {
      result.current.setAuth({
        user: { id: 'user-1', username: 'alice', displayName: 'Alice' },
        token: 'jwt-token',
      })
      result.current.logout()
    })

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
  })

  it('persists token to localStorage', () => {
    const { result } = renderHook(() => useAuthStore())

    act(() => {
      result.current.setAuth({
        user: { id: 'user-1', username: 'alice', displayName: 'Alice' },
        token: 'jwt-token',
      })
    })

    expect(localStorage.getItem('token')).toBe('jwt-token')
  })

  it('removes token from localStorage on logout', () => {
    const { result } = renderHook(() => useAuthStore())

    act(() => {
      result.current.setAuth({
        user: { id: 'user-1', username: 'alice', displayName: 'Alice' },
        token: 'jwt-token',
      })
      result.current.logout()
    })

    expect(localStorage.getItem('token')).toBeNull()
  })
})
