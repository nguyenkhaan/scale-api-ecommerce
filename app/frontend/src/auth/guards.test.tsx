import { describe, expect, it } from 'bun:test'
import type { ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import type { Role } from '../lib/auth-api'
import { AuthNavView } from './AuthNav'
import { AuthContext, type AuthState } from './auth-context'
import { RequireAuth } from './RequireAuth'
import { RequireRole } from './RequireRole'

function withState(state: AuthState, children: ReactNode) {
  const value = {
    state,
    login: async () => {
      throw new Error('unused')
    },
    register: async () => {
      throw new Error('unused')
    },
    logout: async () => undefined,
    refresh: async () => undefined,
  }
  return renderToStaticMarkup(
    <MemoryRouter>
      <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
    </MemoryRouter>,
  )
}

const signedIn = (role: Role): AuthState => ({
  status: 'authenticated',
  user: { id: 'u1', email: 'a@b.co', role },
})

describe('RequireAuth', () => {
  it('shows an accessible loading state', () => {
    const markup = withState({ status: 'loading' }, <RequireAuth>secret</RequireAuth>)
    expect(markup).toContain('aria-busy="true"')
    expect(markup).toContain('role="status"')
    expect(markup).not.toContain('secret')
  })

  it('renders children when authenticated', () => {
    expect(withState(signedIn('customer'), <RequireAuth>secret</RequireAuth>)).toContain('secret')
  })

  it('does not render children when anonymous', () => {
    expect(withState({ status: 'anonymous' }, <RequireAuth>secret</RequireAuth>)).not.toContain('secret')
  })
})

describe('RequireRole', () => {
  it('shows the 403 view for a non-admin', () => {
    const markup = withState(signedIn('customer'), <RequireRole role="admin">secret</RequireRole>)
    expect(markup).toContain('Bạn không có quyền truy cập trang này')
    expect(markup).not.toContain('secret')
  })

  it('renders children for the required role', () => {
    expect(withState(signedIn('admin'), <RequireRole role="admin">secret</RequireRole>)).toContain('secret')
  })
})

describe('AuthNavView', () => {
  it('offers login and register links when anonymous', () => {
    const markup = renderToStaticMarkup(
      <MemoryRouter>
        <AuthNavView state={{ status: 'anonymous' }} onLogout={() => undefined} />
      </MemoryRouter>,
    )
    expect(markup).toContain('href="/login"')
    expect(markup).toContain('href="/register"')
    expect(markup).not.toContain('Đăng xuất')
  })

  it('shows email, role label and logout button when authenticated', () => {
    const markup = renderToStaticMarkup(
      <MemoryRouter>
        <AuthNavView state={signedIn('admin')} onLogout={() => undefined} />
      </MemoryRouter>,
    )
    expect(markup).toContain('a@b.co')
    expect(markup).toContain('Quản trị viên')
    expect(markup).toContain('<button type="button"')
    expect(markup).toContain('Đăng xuất')
    expect(markup).not.toContain('role="alert"')
  })

  it('announces a logout the server did not confirm', () => {
    const markup = renderToStaticMarkup(
      <MemoryRouter>
        <AuthNavView state={signedIn('customer')} onLogout={() => undefined} logoutFailed />
      </MemoryRouter>,
    )
    expect(markup).toContain('role="alert"')
    expect(markup).toContain('Đăng xuất chưa thành công')
  })
})
