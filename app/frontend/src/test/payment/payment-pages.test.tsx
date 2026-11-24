import { describe, expect, it } from 'bun:test'
import type { ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'

import { PaymentPageView, type PaymentPageViewProps } from '@/features/payment/pages/PaymentPage'
import { PaymentSimulatorView } from '@/features/payment/pages/PaymentSimulatorPage'
import type { SimulatorSession } from '@/features/payment/payment.type'

import { makeOrder } from '../bill/order-fixtures'

const render = (node: ReactNode) => renderToStaticMarkup(<MemoryRouter>{node}</MemoryRouter>)

const session = { id: 's1', order_id: 'order-1', status: 'pending' as const, amount: 848000, currency: 'VND' as const, simulator_reference: 'PAY-7F3K9Q' }
const pageProps = (overrides: Partial<PaymentPageViewProps> = {}): PaymentPageViewProps => ({
  order: makeOrder(),
  loading: false,
  loadError: null,
  notFound: false,
  sessionState: { session, failure: null, retry: () => undefined },
  onRetryLoad: () => undefined,
  ...overrides,
})

describe('PaymentPageView', () => {
  it('shows amount, simulator reference, deadline note and neutral waiting copy while pending', () => {
    const markup = render(<PaymentPageView {...pageProps()} />)
    expect(markup).toContain('PAY-7F3K9Q')
    expect(markup).toContain('Thời gian chính thức theo hệ thống')
    expect(markup).toContain('Đang chờ kết quả thanh toán…')
    expect(markup).not.toContain('thất bại')
  })

  it('does not use failure wording when the session status is unknown', () => {
    const markup = render(<PaymentPageView {...pageProps({ order: makeOrder({ payment_status: 'unknown' }) })} />)
    expect(markup).toContain('Đang chờ kết quả thanh toán…')
    expect(markup).toContain('Đừng thanh toán lại')
    expect(markup).not.toContain('Thanh toán thất bại')
  })

  it('shows success with the order number when confirmed', () => {
    const markup = render(<PaymentPageView {...pageProps({ order: makeOrder({ status: 'confirmed', payable: false, payment_status: 'succeeded' }) })} />)
    expect(markup).toContain('Thanh toán thành công')
    expect(markup).toContain('SS-1010-0042')
  })

  it('shows failure with a retry link back to the cart when payment_failed', () => {
    const markup = render(<PaymentPageView {...pageProps({ order: makeOrder({ status: 'payment_failed', payable: false, payment_status: 'failed' }) })} />)
    expect(markup).toContain('Thanh toán thất bại')
    expect(markup).toContain('href="/cart"')
    expect(markup).toContain('Đặt lại')
  })

  it('explains cancelled and expired orders and offers to buy again', () => {
    const cancelled = render(<PaymentPageView {...pageProps({ order: makeOrder({ status: 'cancelled', payable: false }) })} />)
    const expired = render(<PaymentPageView {...pageProps({ order: makeOrder({ status: 'expired', payable: false }) })} />)
    expect(cancelled).toContain('Đơn hàng đã được hủy')
    expect(cancelled).toContain('Mua lại')
    expect(expired).toContain('hết hạn thanh toán')
    expect(expired).toContain('Mua lại')
  })

  it('shows the release message once for an overdue pending order', () => {
    const markup = render(<PaymentPageView {...pageProps({ order: makeOrder({ payable: false }) })} />)
    expect(markup.split('hệ thống đang giải phóng hàng')).toHaveLength(2)
  })

  it('shows a refetch notice for ORDER_NOT_PAYABLE and a retry for transient failures', () => {
    const notPayable = render(<PaymentPageView {...pageProps({ sessionState: { session: null, failure: { kind: 'not-payable', text: 'Đơn hàng không còn có thể thanh toán.' }, retry: () => undefined } })} />)
    const transient = render(<PaymentPageView {...pageProps({ sessionState: { session: null, failure: { kind: 'error', text: 'Không thể tạo phiên thanh toán.' }, retry: () => undefined } })} />)
    expect(notPayable).toContain('không còn có thể thanh toán')
    expect(transient).toContain('Thử lại')
  })

  it('keeps a neutral review notice for reconciliation without claiming success', () => {
    const order = makeOrder({ status: 'expired', payable: false, payment_status: 'succeeded', reconciliation_required: true })
    const markup = render(<PaymentPageView {...pageProps({ order })} />)
    expect(markup).toContain('đang được kiểm tra')
    expect(markup).not.toContain('Thanh toán thành công')
  })

  it('renders loading and not-found states', () => {
    expect(render(<PaymentPageView {...pageProps({ order: undefined, loading: true })} />)).toContain('role="status"')
    const notFound = render(<PaymentPageView {...pageProps({ order: undefined, notFound: true, loadError: 'Không tìm thấy đơn hàng.' })} />)
    expect(notFound).toContain('Không tìm thấy đơn hàng')
  })
})

describe('PaymentSimulatorView', () => {
  const sim: SimulatorSession = {
    id: 's1',
    order_id: 'order-1',
    order_number: 'SS-1010-0042',
    amount: 848000,
    currency: 'VND',
    status: 'pending',
    simulator_reference: 'PAY-7F3K9Q',
    order_status: 'pending_payment',
    expires_at: '2026-10-10T13:46:00Z',
  }
  const base = { sessions: [sim], loading: false, refreshing: false, error: null, onRefresh: () => undefined, schedule: async () => ({ event_id: 'e', scheduled_at: '2026-10-10T13:43:05Z' }) }

  it('shows the test-only notice, a refresh button and one form per session', () => {
    const markup = render(<PaymentSimulatorView {...base} />)
    expect(markup).toContain('Chỉ dùng trong môi trường thử nghiệm')
    expect(markup).toContain('Làm mới')
    expect(markup).toContain('SS-1010-0042')
    expect(markup).toContain('Không gửi (mô phỏng mất callback)')
    expect(markup).toContain('Chữ ký sai')
    expect(markup).toContain('Ghi đè số tiền')
    expect(markup).toContain('Ghi đè tiền tệ')
    expect(markup).toContain('<details')
    expect(markup).not.toContain('<details open')
  })

  it('renders empty and disabled/forbidden error states', () => {
    expect(render(<PaymentSimulatorView {...base} sessions={[]} />)).toContain('Không có phiên thanh toán nào đang chờ.')
    const error = render(<PaymentSimulatorView {...base} sessions={[]} error="Công cụ mô phỏng thanh toán đang tắt trên môi trường này." />)
    expect(error).toContain('role="alert"')
    expect(error).toContain('đang tắt')
  })
})
