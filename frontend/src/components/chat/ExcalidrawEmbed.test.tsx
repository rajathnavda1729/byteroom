import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { ExcalidrawEmbed } from './ExcalidrawEmbed'

vi.mock('@excalidraw/excalidraw', () => ({
  Excalidraw: ({ viewModeEnabled }: { viewModeEnabled?: boolean }) => (
    <div data-testid="excalidraw-mock" data-readonly={viewModeEnabled ? 'true' : 'false'}>
      Excalidraw
    </div>
  ),
}))

const mockState = JSON.stringify({
  type: 'excalidraw',
  version: 2,
  elements: [{ type: 'rectangle', x: 100, y: 100, width: 200, height: 100 }],
  appState: { viewBackgroundColor: '#ffffff' },
})

describe('ExcalidrawEmbed', () => {
  it('renders excalidraw container', async () => {
    render(<ExcalidrawEmbed state={mockState} />)

    await waitFor(() => {
      expect(screen.getByTestId('excalidraw-container')).toBeInTheDocument()
    })
  })

  it('is read-only by default', async () => {
    render(<ExcalidrawEmbed state={mockState} readOnly />)

    await waitFor(() => {
      const container = screen.getByTestId('excalidraw-container')
      expect(container).toHaveAttribute('data-readonly', 'true')
    })
  })

  it('allows editing when readOnly is false', async () => {
    render(<ExcalidrawEmbed state={mockState} readOnly={false} onUpdate={vi.fn()} />)

    await waitFor(() => {
      expect(screen.getByTestId('excalidraw-container')).toBeInTheDocument()
    })
  })

  it('shows expand button', async () => {
    render(<ExcalidrawEmbed state={mockState} />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /expand/i })).toBeInTheDocument()
    })
  })

  it('handles invalid state gracefully', async () => {
    render(<ExcalidrawEmbed state="invalid json" />)

    await waitFor(() => {
      expect(screen.getByText(/failed to load diagram/i)).toBeInTheDocument()
    })
  })
})
