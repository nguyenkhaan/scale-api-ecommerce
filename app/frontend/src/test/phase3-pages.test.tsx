import { describe, expect, it } from 'bun:test'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import { CartPage } from '@/features/cart/pages/CartPage'
import { AdminFlashSalePage } from '@/features/admin-flashsale/pages/AdminFlashSalePage'
import { FlashSalesPage } from '@/features/flashsale/pages/FlashSalesPage'

function renderPage(page: ReactNode) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return renderToStaticMarkup(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>{page}</MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('Phase 3 Page Rendering', () => {
  it('renders FlashSalesPage heading and cart link', () => {
    const markup = renderPage(<FlashSalesPage />)
    expect(markup).toContain('Flash Sale')
    expect(markup).toContain('Xem giỏ hàng')
  })

  it('renders CartPage heading and flash sale navigation link', () => {
    const markup = renderPage(<CartPage />)
    expect(markup).toContain('Giỏ hàng của bạn')
    expect(markup).toContain('Tiếp tục săn Flash Sale')
  })

  it('renders AdminFlashSalePage heading and create button', () => {
    const markup = renderPage(<AdminFlashSalePage />)
    expect(markup).toContain('Quản lý Flash Sale')
    expect(markup).toContain('Tạo chiến dịch mới')
  })
})
