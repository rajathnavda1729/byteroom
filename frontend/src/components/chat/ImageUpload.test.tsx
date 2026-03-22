import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/test/mocks/server'
import { ImageUpload } from './ImageUpload'

describe('ImageUpload', () => {
  it('renders attach button and hidden file input', () => {
    render(<ImageUpload onUpload={vi.fn()} />)

    expect(screen.getByRole('button', { name: /attach/i })).toBeInTheDocument()
    expect(screen.getByTestId('file-input')).toBeInTheDocument()
  })

  it('shows preview after valid file selection', async () => {
    const user = userEvent.setup()
    const file = new File(['img'], 'test.png', { type: 'image/png' })

    render(<ImageUpload onUpload={vi.fn()} />)

    const input = screen.getByTestId('file-input')
    await user.upload(input, file)

    expect(await screen.findByRole('img', { name: /preview/i })).toBeInTheDocument()
  })

  it('validates file type', async () => {
    const user = userEvent.setup()
    const file = new File(['test'], 'test.exe', { type: 'application/x-executable' })

    render(<ImageUpload onUpload={vi.fn()} />)

    await user.upload(screen.getByTestId('file-input'), file)

    expect(await screen.findByText(/unsupported file type/i)).toBeInTheDocument()
  })

  it('validates file size', async () => {
    const user = userEvent.setup()
    const largeBuffer = new ArrayBuffer(11 * 1024 * 1024)
    const largeFile = new File([largeBuffer], 'large.png', { type: 'image/png' })

    render(<ImageUpload onUpload={vi.fn()} />)

    await user.upload(screen.getByTestId('file-input'), largeFile)

    expect(await screen.findByText(/file too large/i)).toBeInTheDocument()
  })

  it('uploads file and calls onUpload with key', async () => {
    server.use(
      http.post('/api/upload/request', () => {
        return HttpResponse.json({
          upload_url: 'https://s3.example.com/upload',
          file_key: 'uploads/abc123.png',
        })
      }),
      http.put('https://s3.example.com/upload', () => {
        return new HttpResponse(null, { status: 200 })
      }),
    )

    const onUpload = vi.fn()
    const user = userEvent.setup()
    const file = new File(['img'], 'test.png', { type: 'image/png' })

    render(<ImageUpload onUpload={onUpload} />)

    await user.upload(screen.getByTestId('file-input'), file)
    await user.click(await screen.findByRole('button', { name: /upload/i }))

    await waitFor(() => {
      expect(onUpload).toHaveBeenCalledWith('uploads/abc123.png')
    })
  })
})
