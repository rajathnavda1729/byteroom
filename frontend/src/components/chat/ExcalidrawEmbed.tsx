import { lazy, Suspense, useState } from 'react'

const ExcalidrawLazy = lazy(() =>
  import('@excalidraw/excalidraw').then((m) => ({ default: m.Excalidraw })),
)

interface Props {
  state: string
  readOnly?: boolean
  onUpdate?: (state: string) => void
}

interface ParsedState {
  type?: string
  version?: number
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  elements?: any[]
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  appState?: Record<string, any>
}

export function ExcalidrawEmbed({ state, readOnly = true, onUpdate }: Props) {
  const [error, setError] = useState(false)

  let parsed: ParsedState | null = null
  try {
    parsed = JSON.parse(state) as ParsedState
  } catch {
    if (!error) setError(true)
  }

  if (error || !parsed) {
    return (
      <div className="bg-red-900/20 border border-red-700 rounded-lg p-3 text-sm text-red-400">
        Failed to load diagram
      </div>
    )
  }

  return (
    <div
      data-testid="excalidraw-container"
      data-readonly={readOnly ? 'true' : 'false'}
      className="relative rounded-xl overflow-hidden border border-gray-700 my-2"
      style={{ height: 400 }}
    >
      <div className="absolute top-2 right-2 z-10 flex gap-2">
        <button
          aria-label="Expand"
          className="bg-gray-800 hover:bg-gray-700 text-white text-xs px-2 py-1 rounded"
        >
          ⛶ Expand
        </button>
      </div>

      <Suspense
        fallback={
          <div className="flex items-center justify-center h-full text-gray-400 text-sm">
            Loading diagram…
          </div>
        }
      >
        <ExcalidrawLazy
          initialData={{
            elements: parsed.elements ?? [],
            appState: parsed.appState,
          }}
          viewModeEnabled={readOnly}
          onChange={
            !readOnly && onUpdate
              ? (elements, appState) => {
                  onUpdate(
                    JSON.stringify({
                      type: 'excalidraw',
                      version: 2,
                      elements,
                      appState,
                    }),
                  )
                }
              : undefined
          }
        />
      </Suspense>
    </div>
  )
}
