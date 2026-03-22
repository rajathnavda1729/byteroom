import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BrowserRouter } from 'react-router-dom'
import { describe, it, expect, beforeEach } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/test/mocks/server'
import { LoginPage } from './LoginPage'
import { useAuthStore } from '@/stores/authStore'

const renderWithRouter = (component: React.ReactNode) =>
  render(<BrowserRouter>{component}</BrowserRouter>)

describe('LoginPage', () => {
  beforeEach(() => {
    useAuthStore.getState().logout()
    localStorage.clear()
  })

  it('renders login form', () => {
    renderWithRouter(<LoginPage />)

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('shows validation errors for empty fields', async () => {
    renderWithRouter(<LoginPage />)
    const user = userEvent.setup()

    await user.click(screen.getByRole('button', { name: /sign in/i }))

    expect(await screen.findByText(/email is required/i)).toBeInTheDocument()
    expect(await screen.findByText(/password is required/i)).toBeInTheDocument()
  })

  it('submits form with credentials and sets auth state', async () => {
    server.use(
      http.post('/api/auth/login', () => {
        return HttpResponse.json({
          user_id: 'user-1',
          username: 'alice',
          display_name: 'Alice',
          token: 'jwt-token',
        })
      }),
    )

    renderWithRouter(<LoginPage />)
    const user = userEvent.setup()

    await user.type(screen.getByLabelText(/email/i), 'alice')
    await user.type(screen.getByLabelText(/password/i), 'password123')
    await user.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => {
      expect(useAuthStore.getState().isAuthenticated).toBe(true)
    })
  })

  it('shows error message on failed login', async () => {
    server.use(
      http.post('/api/auth/login', () => {
        return HttpResponse.json({ message: 'Invalid credentials' }, { status: 401 })
      }),
    )

    renderWithRouter(<LoginPage />)
    const user = userEvent.setup()

    await user.type(screen.getByLabelText(/email/i), 'alice')
    await user.type(screen.getByLabelText(/password/i), 'wrongpassword')
    await user.click(screen.getByRole('button', { name: /sign in/i }))

    expect(await screen.findByText(/invalid credentials/i)).toBeInTheDocument()
  })
})
