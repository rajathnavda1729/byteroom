import '@testing-library/jest-dom'
import { server } from './mocks/server'
import { beforeAll, afterAll, afterEach, vi } from 'vitest'

beforeAll(() => server.listen({ onUnhandledRequest: 'warn' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

// jsdom doesn't implement URL.createObjectURL
Object.defineProperty(URL, 'createObjectURL', {
  writable: true,
  configurable: true,
  value: vi.fn((file: Blob) => `blob:mock/${(file as File).name ?? 'file'}`),
})
Object.defineProperty(URL, 'revokeObjectURL', {
  writable: true,
  configurable: true,
  value: vi.fn(),
})

// jsdom doesn't implement scrollIntoView
Element.prototype.scrollIntoView = vi.fn()

// Allow jest-websocket-mock / mock-socket to replace global.WebSocket
if (typeof globalThis.WebSocket !== 'undefined') {
  Object.defineProperty(globalThis, 'WebSocket', {
    writable: true,
    configurable: true,
    value: globalThis.WebSocket,
  })
}
