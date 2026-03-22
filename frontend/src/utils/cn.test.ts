import { describe, it, expect } from 'vitest'
import { cn } from './cn'

describe('cn', () => {
  it('joins string class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('filters out falsy values', () => {
    expect(cn('foo', undefined, null, false, 'bar')).toBe('foo bar')
  })

  it('handles object syntax', () => {
    expect(cn({ active: true, disabled: false })).toBe('active')
  })

  it('combines strings and objects', () => {
    expect(cn('base', { extra: true })).toBe('base extra')
  })

  it('returns empty string for all falsy', () => {
    expect(cn(undefined, null, false)).toBe('')
  })
})
