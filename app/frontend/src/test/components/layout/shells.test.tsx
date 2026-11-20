import { describe, expect, it } from 'bun:test'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthContext } from '@/context/AuthContext'
import { AdminShell, CustomerHeader, CustomerShell } from '@/components/layout'
import { AppRoute } from '@/routes/AppRoute'
import type { AuthContextValue, User } from '@/features/auth/auth.type'

const customer: User = { id: 'u-customer', email: 'customer@example.com', role: 'customer' }
const admin: User = { id: 'u-admin', email: 'admin@example.com', role: 'admin' }

function authValue(user: User | null): AuthContextValue {
  return {
    state: user === null ? { status: 'anonymous' } : { status: 'authenticated', user },
    login: async () => user ?? customer,
    register: async () => user ?? customer,
    logout: async () => {},
    refresh: async () => {},
  }
}

function renderWithAuth(children: ReactNode, user: User | null) {
  return renderToStaticMarkup(
    <QueryClientProvider client={new QueryClient()}>
      <AuthContext.Provider value={authValue(user)}>{children}</AuthContext.Provider>
    </QueryClientProvider>,
  )
}

describe('application shells', () => {
  it('renders the customer header, search, cart and auth controls', () => {
    const markup = renderWithAuth(
      <MemoryRouter initialEntries={['/']}>
        <CustomerHeader />
      </MemoryRouter>,
      null,
    )

    expect(markup).toContain('SaleScale')
    expect(markup).toContain('role="search"')
    expect(markup).toContain('Giỏ hàng')
    expect(markup).toContain('Đăng nhập')
    expect(markup).toContain('Đăng ký')
  })

  it('renders customer content through an outlet under the public shell', () => {
    const markup = renderWithAuth(
      <MemoryRouter initialEntries={['/products']}>
        <Routes>
          <Route element={<CustomerShell />}>
            <Route path="products" element={<p>Catalog outlet</p>} />
          </Route>
        </Routes>
      </MemoryRouter>,
      null,
    )

    expect(markup).toContain('Catalog outlet')
    expect(markup).toContain('<main')
  })

  it('renders only implemented admin navigation and marks flash sales active', () => {
    const markup = renderWithAuth(
      <MemoryRouter initialEntries={['/admin/flash-sales']}>
        <Routes>
          <Route path="admin" element={<AdminShell />}>
            <Route path="flash-sales" element={<p>Admin outlet</p>} />
          </Route>
        </Routes>
      </MemoryRouter>,
      admin,
    )

    expect(markup).toContain('Quản lý Flash Sale')
    expect(markup).toContain('aria-current="page"')
    expect(markup).toContain('admin@example.com')
    expect(markup).not.toContain('Tồn kho')
    expect(markup).not.toContain('Sản phẩm quản trị')
    expect(markup).toContain('Admin outlet')
  })

  it('keeps the catalog public while protecting the cart route', () => {
    const publicCatalog = renderWithAuth(
      <MemoryRouter initialEntries={['/products']}>
        <AppRoute />
      </MemoryRouter>,
      null,
    )
    const protectedCart = renderWithAuth(
      <MemoryRouter initialEntries={['/cart']}>
        <AppRoute />
      </MemoryRouter>,
      null,
    )

    expect(publicCatalog).toContain('Sản phẩm')
    expect(protectedCart).not.toContain('Giỏ hàng của bạn')
    expect(protectedCart).toContain('Đăng nhập')
  })
})
