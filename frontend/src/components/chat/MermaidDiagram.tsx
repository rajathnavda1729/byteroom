import { useEffect, useRef, useState } from 'react'

interface Props {
  content: string
}

let mermaidInitialized = false

async function getMermaid() {
  const m = await import('mermaid')
  if (!mermaidInitialized) {
    m.default.initialize({
      startOnLoad: false,
      theme: 'dark',
      securityLevel: 'loose',
    })
    mermaidInitialized = true
  }
  return m.default
}

export function MermaidDiagram({ content }: Props) {
  const [svg, setSvg] = useState<string | null>(null)
  const [error, setError] = useState(false)
  const [zoom, setZoom] = useState(1)
  const idRef = useRef(`mermaid-${Math.random().toString(36).slice(2)}`)
  const hasMany = content.split('\n').length > 10

  useEffect(() => {
    let cancelled = false
    setSvg(null)
    setError(false)

    getMermaid()
      .then((mermaid) => mermaid.render(idRef.current, content.trim()))
      .then(({ svg: result }) => {
        if (!cancelled) setSvg(result)
      })
      .catch(() => {
        if (!cancelled) setError(true)
      })

    return () => {
      cancelled = true
    }
  }, [content])

  if (error) {
    return (
      <div className="bg-red-900/20 border border-red-700 rounded-lg p-3 text-sm text-red-400">
        Failed to render diagram
      </div>
    )
  }

  if (!svg) {
    return <div className="text-gray-400 text-sm py-2">Loading diagram…</div>
  }

  return (
    <div className="relative my-2">
      {hasMany && (
        <div className="flex gap-2 mb-2">
          <button
            onClick={() => setZoom((z) => Math.min(z + 0.25, 3))}
            aria-label="Zoom in"
            className="text-xs bg-gray-700 hover:bg-gray-600 text-white px-2 py-1 rounded"
          >
            +
          </button>
          <button
            onClick={() => setZoom((z) => Math.max(z - 0.25, 0.25))}
            aria-label="Zoom out"
            className="text-xs bg-gray-700 hover:bg-gray-600 text-white px-2 py-1 rounded"
          >
            −
          </button>
          <button
            onClick={() => setZoom(1)}
            aria-label="Reset zoom"
            className="text-xs bg-gray-700 hover:bg-gray-600 text-white px-2 py-1 rounded"
          >
            Reset
          </button>
        </div>
      )}
      <div
        role="img"
        aria-label="Mermaid diagram"
        style={{ transform: `scale(${zoom})`, transformOrigin: 'top left', overflow: 'auto' }}
        // eslint-disable-next-line react/no-danger
        dangerouslySetInnerHTML={{ __html: svg }}
      />
    </div>
  )
}
