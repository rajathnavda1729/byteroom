import { useRef, useState } from 'react'
import { api } from '@/services/api'

const ALLOWED_TYPES = ['image/png', 'image/jpeg', 'image/gif', 'image/webp']
const MAX_BYTES = 10 * 1024 * 1024 // 10 MB

interface UploadURLResponse {
  upload_url: string
  file_key: string
}

interface Props {
  onUpload: (fileKey: string) => void
}

export function ImageUpload({ onUpload }: Props) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [uploading, setUploading] = useState(false)

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    setError(null)
    setPreview(null)
    setSelectedFile(null)

    if (!file) return

    if (!ALLOWED_TYPES.includes(file.type)) {
      setError('Unsupported file type. Only PNG, JPEG, GIF, and WebP are allowed.')
      return
    }
    if (file.size > MAX_BYTES) {
      setError('File too large. Maximum size is 10 MB.')
      return
    }

    setSelectedFile(file)
    const url = URL.createObjectURL(file)
    setPreview(url)
  }

  const handleUpload = async () => {
    if (!selectedFile) return
    setUploading(true)
    setError(null)

    try {
      const { upload_url, file_key } = await api.post<UploadURLResponse>('/api/upload/request', {
        file_name: selectedFile.name,
        mime_type: selectedFile.type,
        file_size: selectedFile.size,
      })

      await fetch(upload_url, {
        method: 'PUT',
        body: selectedFile,
        headers: { 'Content-Type': selectedFile.type },
      })

      onUpload(file_key)
      setPreview(null)
      setSelectedFile(null)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Upload failed')
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="flex flex-col gap-2">
      <input
        ref={inputRef}
        data-testid="file-input"
        type="file"
        className="hidden"
        onChange={handleFileChange}
      />

      <button
        type="button"
        aria-label="Attach file"
        onClick={() => inputRef.current?.click()}
        className="p-2 text-gray-400 hover:text-white transition"
      >
        📎
      </button>

      {error && <p className="text-xs text-red-400">{error}</p>}

      {preview && (
        <div className="flex items-center gap-3">
          <img src={preview} alt="Preview" aria-label="preview" className="w-20 h-20 object-cover rounded-lg" />
          <button
            type="button"
            aria-label="Upload"
            onClick={handleUpload}
            disabled={uploading}
            className="text-sm bg-blue-600 hover:bg-blue-500 disabled:bg-blue-800 text-white px-3 py-1.5 rounded-lg"
          >
            {uploading ? 'Uploading…' : 'Upload'}
          </button>
        </div>
      )}
    </div>
  )
}
