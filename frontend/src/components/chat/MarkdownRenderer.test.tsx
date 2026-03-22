import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { MarkdownRenderer } from './MarkdownRenderer'

vi.mock('./MermaidDiagram', () => ({
  MermaidDiagram: ({ content }: { content: string }) => (
    <div data-testid="mermaid-diagram">{content}</div>
  ),
}))

describe('MarkdownRenderer', () => {
  it('renders plain text', () => {
    render(<MarkdownRenderer content="Hello world" />)
    expect(screen.getByText('Hello world')).toBeInTheDocument()
  })

  it('renders bold text', () => {
    render(<MarkdownRenderer content="**bold text**" />)
    const el = screen.getByText('bold text')
    expect(el.tagName).toMatch(/^(STRONG|B)$/)
  })

  it('renders italic text', () => {
    render(<MarkdownRenderer content="*italic text*" />)
    const el = screen.getByText('italic text')
    expect(el.tagName).toMatch(/^(EM|I)$/)
  })

  it('renders inline code', () => {
    render(<MarkdownRenderer content="Use `const` for constants" />)
    const code = screen.getByText('const')
    expect(code.tagName).toBe('CODE')
  })

  it('renders code blocks', () => {
    const code = '```javascript\nconst x = 42;\nconsole.log(x);\n```'
    render(<MarkdownRenderer content={code} />)
    expect(screen.getByText(/const/)).toBeInTheDocument()
  })

  it('renders mermaid fenced blocks with MermaidDiagram', () => {
    const md = '```mermaid\ngraph TD\n  A-->B\n```'
    render(<MarkdownRenderer content={md} />)
    const el = screen.getByTestId('mermaid-diagram')
    expect(el).toHaveTextContent('graph TD')
    expect(el).toHaveTextContent('A-->B')
  })

  it('renders links with target="_blank"', () => {
    render(<MarkdownRenderer content="[Google](https://google.com)" />)

    const link = screen.getByRole('link', { name: 'Google' })
    expect(link).toHaveAttribute('href', 'https://google.com')
    expect(link).toHaveAttribute('target', '_blank')
    expect(link).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('sanitizes script tags', () => {
    render(<MarkdownRenderer content="<script>alert('xss')</script>Safe text" />)
    expect(document.querySelector('script')).toBeNull()
  })

  it('renders lists', () => {
    render(<MarkdownRenderer content={'- Item 1\n- Item 2'} />)
    expect(screen.getByText('Item 1')).toBeInTheDocument()
    expect(screen.getByText('Item 2')).toBeInTheDocument()
  })

  it('renders tables', () => {
    const table = `\n| Header 1 | Header 2 |\n|----------|----------|\n| Cell 1   | Cell 2   |\n`
    render(<MarkdownRenderer content={table} />)
    expect(screen.getByText('Header 1')).toBeInTheDocument()
    expect(screen.getByText('Cell 1')).toBeInTheDocument()
  })
})
