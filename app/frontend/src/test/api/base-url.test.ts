import { describe, expect, it } from 'bun:test'
import { normalizeApiBaseUrl } from '@/api/base-url'

describe('normalizeApiBaseUrl', () => {
  it('uses the Vite proxy API root when no origin is configured', () => {
    expect(normalizeApiBaseUrl(undefined)).toBe('/api')
    expect(normalizeApiBaseUrl('')).toBe('/api')
  })

  it('appends one api prefix to an origin', () => {
    expect(normalizeApiBaseUrl('http://localhost:4000')).toBe('http://localhost:4000/api')
  })

  it('preserves an existing api prefix', () => {
    expect(normalizeApiBaseUrl('http://localhost:4000/api')).toBe('http://localhost:4000/api')
    expect(normalizeApiBaseUrl('http://localhost:4000/api/')).toBe('http://localhost:4000/api')
  })

  it('removes trailing slashes before adding the prefix', () => {
    expect(normalizeApiBaseUrl('http://localhost:4000///')).toBe('http://localhost:4000/api')
  })
})
