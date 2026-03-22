import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { useState } from 'react'
import type { Components } from 'react-markdown'

interface Props {
  content: string
}

function CopyButton({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = () => {
    navigator.clipboard.writeText(code).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  return (
    <button
      onClick={handleCopy}
      aria-label="Copy code"
      className="absolute top-2 right-2 text-xs text-gray-400 hover:text-white bg-gray-700 hover:bg-gray-600 px-2 py-1 rounded transition"
    >
      {copied ? 'Copied!' : 'Copy'}
    </button>
  )
}

const components: Components = {
  code({ className, children, ...props }) {
    const match = /language-(\w+)/.exec(className ?? '')
    const lang = match?.[1] ?? ''
    const code = String(children).replace(/\n$/, '')
    const inline = !className

    if (!inline && lang) {
      return (
        <div className="relative my-2 rounded-lg overflow-hidden">
          <CopyButton code={code} />
          <SyntaxHighlighter
            style={oneDark}
            language={lang}
            showLineNumbers
            PreTag="div"
            customStyle={{ margin: 0, borderRadius: '0.5rem', paddingTop: '2.5rem' }}
          >
            {code}
          </SyntaxHighlighter>
        </div>
      )
    }

    return (
      <code className="bg-gray-800 text-orange-300 px-1 py-0.5 rounded text-xs font-mono" {...props}>
        {children}
      </code>
    )
  },

  a({ href, children }) {
    return (
      <a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        className="text-blue-400 hover:text-blue-300 underline"
      >
        {children}
      </a>
    )
  },

  p({ children }) {
    return <p className="mb-1 last:mb-0">{children}</p>
  },

  ul({ children }) {
    return <ul className="list-disc list-inside mb-1 space-y-0.5">{children}</ul>
  },

  ol({ children }) {
    return <ol className="list-decimal list-inside mb-1 space-y-0.5">{children}</ol>
  },

  blockquote({ children }) {
    return (
      <blockquote className="border-l-2 border-gray-600 pl-3 text-gray-400 italic my-1">
        {children}
      </blockquote>
    )
  },

  table({ children }) {
    return (
      <div className="overflow-x-auto my-2">
        <table className="min-w-full text-sm border-collapse border border-gray-700">
          {children}
        </table>
      </div>
    )
  },

  th({ children }) {
    return (
      <th className="border border-gray-700 px-3 py-1.5 bg-gray-800 text-left font-medium">
        {children}
      </th>
    )
  },

  td({ children }) {
    return <td className="border border-gray-700 px-3 py-1.5">{children}</td>
  },
}

export function MarkdownRenderer({ content }: Props) {
  return (
    <div className="prose prose-invert prose-sm max-w-none break-words">
      <ReactMarkdown components={components} remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
    </div>
  )
}
