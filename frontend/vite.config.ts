import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, './src'),
    },
  },
  server: {
    // Give Vite's own HMR websocket a dedicated port so it never
    // collides with the /ws proxy that forwards to the Go backend.
    hmr: { clientPort: 5173, port: 5174 },
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
        changeOrigin: true,
        // EPIPE / connection-reset errors are expected in development:
        // React StrictMode double-mounts components, the WS hook closes the
        // first connection intentionally, and Vite's proxy logs the resulting
        // broken-pipe as an error.  We suppress it here to reduce noise.
        configure: (proxy) => {
          proxy.on('error', (err: NodeJS.ErrnoException) => {
            if (err.code === 'EPIPE' || err.code === 'ECONNRESET') return
            console.error('[ws proxy error]', err.message)
          })
        },
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          if (id.includes('node_modules')) {
            if (id.includes('react-dom') || id.includes('/react/')) return 'react'
            if (id.includes('react-router')) return 'router'
            if (id.includes('zustand') || id.includes('immer')) return 'zustand'
            if (
              id.includes('react-markdown') ||
              id.includes('remark-gfm') ||
              id.includes('react-syntax-highlighter')
            )
              return 'markdown'
          }
        },
      },
    },
    // react-syntax-highlighter ships all language grammars; it's loaded lazily
    chunkSizeWarningLimit: 900,
  },
})
