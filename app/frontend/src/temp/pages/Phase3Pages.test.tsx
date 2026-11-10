import { describe, expect, it } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'
import { FlashSalesPage } from './FlashSalesPage'
import { CartPage } from './CartPage'
import { AdminFlashSalePage } from './AdminFlashSalePage'

describe('Phase 3 Page Rendering', () => {
  it('renders FlashSalesPage heading and cart link', () => {
    const markup = renderToStaticMarkup(
      <MemoryRouter>
        <FlashSalesPage />
      </MemoryRouter>,
    )
    expect(markup).toContain('Flash Sale')
    expect(markup).toContain('Xem giỏ hàng')
  })

  it('renders CartPage heading and flash sale navigation link', () => {
    const markup = renderToStaticMarkup(
      <MemoryRouter>
        <CartPage />
      </MemoryRouter>,
    )
    expect(markup).toContain('Giỏ hàng của bạn')
    expect(markup).toContain('Tiếp tục săn Flash Sale')
  })

  it('renders AdminFlashSalePage heading and create button', () => {
    const markup = renderToStaticMarkup(
      <MemoryRouter>
        <AdminFlashSalePage />
      </MemoryRouter>,
    )
    expect(markup).toContain('Quản lý Flash Sale')
    expect(markup).toContain('Tạo chiến dịch mới')
  })
})
