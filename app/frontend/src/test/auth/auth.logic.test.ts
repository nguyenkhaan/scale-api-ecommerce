import { describe, expect, it } from 'bun:test'
import { loginRedirectPath, safeNextPath } from '@/features/auth/auth.hook'

describe('auth redirect helpers', () => {
  it('preserves the admin destination through re-authentication', () => {
    expect(loginRedirectPath('/admin/flash-sales', '')).toBe('/login?next=%2Fadmin%2Fflash-sales')
  })

  it('rejects unsafe next paths', () => {
    expect(safeNextPath('//external.example')).toBe('/')
    expect(safeNextPath('/admin/flash-sales')).toBe('/admin/flash-sales')
  })
})
