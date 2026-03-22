import { useEffect, useState } from 'react'

interface Props {
  content: string
}

let mermaidInitialized = false

/** Backend sanitizer (bluemonday) may escape `>` as `&gt;` in older stored messages. */
function decodeHtmlEntities(text: string): string {
  if (typeof document === 'undefined') {
    return text
      .replaceAll('&amp;', '&')
      .replaceAll('&lt;', '<')
      .replaceAll('&gt;', '>')
      .replaceAll('&quot;', '"')
      .replaceAll('&#39;', "'")
  }
  const t = document.createElement('textarea')
  t.innerHTML = text
  return t.value
}

function normalizeMermaidSource(source: string): string {
  const decoded = decodeHtmlEntities(source.replace(/^\uFEFF/, '').replace(/\r\n/g, '\n'))
  return decoded.trim()
}

async function getMermaid() {
  const m = await import('mermaid')
  const api = m.default ?? m
  if (!mermaidInitialized) {
    api.initialize({
      startOnLoad: false,
      theme: 'dark',
      securityLevel: 'loose',
    })
    mermaidInitialized = true
  }
  return api
}

export function MermaidDiagram({ content }: Props) {
  const [svg, setSvg] = useState<string | null>(null)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [zoom, setZoom] = useState(1)
  const normalized = normalizeMermaidSource(content)
  const hasMany = normalized.split('\n').length > 10

  useEffect(() => {
    let cancelled = false
    setSvg(null)
    setErrorMessage(null)

    if (!normalized) {
      setErrorMessage('Empty diagram')
      return
    }

    const renderId = `br-mermaid-${crypto.randomUUID()}`

    getMermaid()
      .then((mermaid) => mermaid.render(renderId, normalized))
      .then(({ svg: result }) => {
        if (!cancelled) setSvg(result)
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          const msg = err instanceof Error ? err.message : 'Unknown error'
          setErrorMessage(msg)
        }
      })

    return () => {
      cancelled = true
    }
  }, [normalized])

  if (errorMessage != null) {
    return (
      <div className="bg-red-900/20 border border-red-700 rounded-lg p-3 text-sm text-red-400">
        <div>Failed to render diagram</div>
        <div className="text-xs mt-1 text-red-300/90 font-mono break-words">{errorMessage}</div>
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
