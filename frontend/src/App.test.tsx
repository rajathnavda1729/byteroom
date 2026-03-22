import { render, screen } from '@testing-library/react'
import { describe, it, expect, beforeEach } from 'vitest'
import App from './App'
import { useAuthStore } from './stores/authStore'

describe('App', () => {
  beforeEach(() => {
    useAuthStore.getState().logout()
    localStorage.clear()
  })

  it('redirects to login when not authenticated', () => {
    render(<App />)
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('renders chat page when authenticated', () => {
    useAuthStore.getState().setAuth({
      user: { id: 'u1', username: 'alice', displayName: 'Alice' },
      token: 'tok',
    })
    render(<App />)
    expect(screen.getByText('ByteRoom')).toBeInTheDocument()
  })
})
