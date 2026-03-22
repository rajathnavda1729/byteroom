import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { BrowserRouter } from 'react-router-dom'
import { describe, it, expect, beforeEach } from 'vitest'
import { RegisterPage } from './RegisterPage'
import { useAuthStore } from '@/stores/authStore'

const renderWithRouter = (component: React.ReactNode) =>
  render(<BrowserRouter>{component}</BrowserRouter>)

describe('RegisterPage', () => {
  beforeEach(() => {
    useAuthStore.getState().logout()
    localStorage.clear()
  })

  it('renders register form', () => {
    renderWithRouter(<RegisterPage />)

    expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/^password/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign up/i })).toBeInTheDocument()
  })

  it('validates password minimum length', async () => {
    renderWithRouter(<RegisterPage />)
    const user = userEvent.setup()

    await user.type(screen.getByLabelText(/^password/i), 'weak')
    await user.click(screen.getByRole('button', { name: /sign up/i }))

    expect(await screen.findByText(/at least 8 characters/i)).toBeInTheDocument()
  })

  it('validates matching passwords', async () => {
    renderWithRouter(<RegisterPage />)
    const user = userEvent.setup()

    await user.type(screen.getByLabelText(/username/i), 'alice')
    await user.type(screen.getByLabelText(/^password/i), 'Password123!')
    await user.type(screen.getByLabelText(/confirm password/i), 'Different123!')
    await user.click(screen.getByRole('button', { name: /sign up/i }))

    expect(await screen.findByText(/passwords must match/i)).toBeInTheDocument()
  })

  it('shows validation errors for empty form', async () => {
    renderWithRouter(<RegisterPage />)
    const user = userEvent.setup()

    await user.click(screen.getByRole('button', { name: /sign up/i }))

    expect(await screen.findByText(/username is required/i)).toBeInTheDocument()
    expect(await screen.findByText(/password is required/i)).toBeInTheDocument()
  })
})
