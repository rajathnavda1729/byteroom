import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MermaidDiagram } from './MermaidDiagram'

vi.mock('mermaid', () => ({
  default: {
    initialize: vi.fn(),
    render: vi.fn(async (_id: string, content: string) => {
      if (!content.startsWith('graph') && !content.startsWith('sequenceDiagram') && !content.startsWith('flowchart')) {
        throw new Error('Parse error')
      }
      return { svg: `<svg role="img" aria-label="Mermaid diagram"><text>${content}</text></svg>` }
    }),
  },
}))

describe('MermaidDiagram', () => {

  it('shows loading state initially', () => {
    render(<MermaidDiagram content="graph TD\nA-->B" />)
    expect(screen.getByText(/loading diagram/i)).toBeInTheDocument()
  })

  it('renders mermaid diagram as SVG', async () => {
    render(<MermaidDiagram content="graph TD\n  A[Start] --> B[End]" />)

    await waitFor(() => {
      const imgs = screen.getAllByRole('img', { name: /diagram/i })
      expect(imgs.length).toBeGreaterThan(0)
    })
  })

  it('shows error for invalid mermaid syntax', async () => {
    render(<MermaidDiagram content="not valid mermaid syntax!!!" />)

    await waitFor(() => {
      expect(screen.getByText(/failed to render/i)).toBeInTheDocument()
    })
  })

  it('shows zoom buttons for diagrams with many lines', async () => {
    const manyLines = Array.from({ length: 15 }, (_, i) => `N${i} --> N${i + 1}`).join('\n')

    render(<MermaidDiagram content={`graph TD\n${manyLines}`} />)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /zoom in/i })).toBeInTheDocument()
    })
  })
})
