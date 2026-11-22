import { describe, expect, it } from 'bun:test'
import type { ReactNode } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryRouter } from 'react-router-dom'

import { OrderDetailView, type OrderDetailViewProps } from '@/features/bill/pages/OrderDetailPage'
import { OrdersListView, type OrderListState } from '@/features/bill/pages/OrdersPage'
import { OrderStatusTabs } from '@/features/bill/components/OrderStatusTabs'

import { makeOrder } from './order-fixtures'

const render = (node: ReactNode) => renderToStaticMarkup(<MemoryRouter>{node}</MemoryRouter>)

const noop = () => undefined
const detailProps = (overrides: Partial<OrderDetailViewProps> = {}): OrderDetailViewProps => ({
  order: makeOrder(),
  loading: false,
  loadError: null,
  notFound: false,
  notice: null,
  cancelOpen: false,
  cancelling: false,
  onRetry: noop,
  onOpenCancel: noop,
  onCloseCancel: noop,
  onConfirmCancel: noop,
  ...overrides,
})

describe('OrderDetailView', () => {
  it('shows snapshot, recipient, total, deadline note and pay/cancel when payable', () => {
    const markup = render(<OrderDetailView {...detailProps()} />)
    expect(markup).toContain('Áo thun basic')
    expect(markup).toContain('AT-DEN-M')
    expect(markup).toContain('Màu: Đen')
    expect(markup).toContain('Flash sale')
    expect(markup).toContain('Nguyễn Minh Anh')
    expect(markup).toContain('Thời gian chính thức theo hệ thống')
    expect(markup).toContain('Thanh toán</a>')
    expect(markup).toContain('href="/orders/order-1/payment"')
    expect(markup).toContain('Hủy đơn')
  })

  it('shows the release message and no pay/cancel when pending but overdue', () => {
    const markup = render(<OrderDetailView {...detailProps({ order: makeOrder({ payable: false }) })} />)
    expect(markup).toContain('Đã quá hạn thanh toán, hệ thống đang giải phóng hàng')
    expect(markup).not.toContain('/payment')
    expect(markup).not.toContain('Hủy đơn')
  })

  it('hides pay/cancel for a confirmed order', () => {
    const markup = render(<OrderDetailView {...detailProps({ order: makeOrder({ status: 'confirmed', payable: false, payment_status: 'succeeded' }) })} />)
    expect(markup).toContain('Đã xác nhận')
    expect(markup).not.toContain('Hủy đơn')
    expect(markup).not.toContain('/payment')
  })

  it('shows a neutral review notice and no success claim when reconciliation is required', () => {
    const order = makeOrder({ status: 'expired', payable: false, payment_status: 'succeeded', reconciliation_required: true })
    const markup = render(<OrderDetailView {...detailProps({ order })} />)
    expect(markup).toContain('đang được kiểm tra')
    expect(markup).toContain('Hết hạn')
    expect(markup).not.toContain('Đã xác nhận')
    expect(markup).not.toContain('thành công')
  })

  it('maps a missing order to the not-found message', () => {
    const markup = render(<OrderDetailView {...detailProps({ order: undefined, notFound: true, loadError: 'Không tìm thấy đơn hàng.' })} />)
    expect(markup).toContain('Không tìm thấy đơn hàng')
    expect(markup).toContain('role="alert"')
  })

  it('renders the cancel conflict notice as a status alert', () => {
    const markup = render(<OrderDetailView {...detailProps({ notice: { kind: 'conflict', text: 'Đơn hàng đã thay đổi trạng thái nên không thể hủy.' } })} />)
    expect(markup).toContain('Đơn hàng đã thay đổi trạng thái')
  })

  it('opens the confirm dialog when requested', () => {
    const markup = render(<OrderDetailView {...detailProps({ cancelOpen: true })} />)
    expect(markup).toContain('role="dialog"')
    expect(markup).toContain('Xác nhận hủy đơn')
  })
})

describe('OrdersListView and tabs', () => {
  const list = (overrides: Partial<OrderListState> = {}): OrderListState => ({
    items: [],
    loading: false,
    loadingMore: false,
    error: null,
    hasMore: false,
    reload: async () => undefined,
    loadMore: async () => undefined,
    ...overrides,
  })

  it('renders rows with number, status, total and a detail link', () => {
    const markup = render(<OrdersListView list={list({ items: [makeOrder()] })} filtered={false} />)
    expect(markup).toContain('SS-1010-0042')
    expect(markup).toContain('Chờ thanh toán')
    expect(markup).toContain('href="/orders/order-1"')
  })

  it('shows the countdown only for payable pending orders', () => {
    const payable = render(<OrdersListView list={list({ items: [makeOrder()] })} filtered={false} />)
    const overdue = render(<OrdersListView list={list({ items: [makeOrder({ payable: false })] })} filtered={false} />)
    expect(payable).toMatch(/Còn|Đã quá hạn/)
    expect(overdue).not.toContain('Còn ')
    expect(overdue).not.toContain('Đã quá hạn')
  })

  it('shows load more only while a next cursor exists', () => {
    expect(render(<OrdersListView list={list({ items: [makeOrder()], hasMore: true })} filtered={false} />)).toContain('Tải thêm')
    expect(render(<OrdersListView list={list({ items: [makeOrder()] })} filtered={false} />)).not.toContain('Tải thêm')
  })

  it('renders empty, loading and error states', () => {
    expect(render(<OrdersListView list={list()} filtered={false} />)).toContain('Bạn chưa có đơn hàng nào.')
    expect(render(<OrdersListView list={list()} filtered />)).toContain('Không có đơn hàng nào ở trạng thái này.')
    expect(render(<OrdersListView list={list({ loading: true })} filtered={false} />)).toContain('role="status"')
    const error = render(<OrdersListView list={list({ error: 'Không thể tải đơn hàng. Vui lòng thử lại.' })} filtered={false} />)
    expect(error).toContain('role="alert"')
    expect(error).toContain('Thử lại')
  })

  it('renders all six tabs and marks the active one', () => {
    const markup = render(<OrderStatusTabs active="confirmed" onChange={noop} />)
    for (const label of ['Tất cả', 'Chờ thanh toán', 'Đã xác nhận', 'Thanh toán thất bại', 'Đã hủy', 'Hết hạn']) {
      expect(markup).toContain(label)
    }
    expect(markup.match(/aria-pressed="true"/g)).toHaveLength(1)
  })
})

describe('reconciliation notice by status', () => {
  it('confirms the order is still confirmed when a conflicting event is under review', () => {
    const markup = render(<OrderDetailView {...detailProps({ order: makeOrder({ status: 'confirmed', payable: false, payment_status: 'succeeded', reconciliation_required: true }) })} />)
    expect(markup).toContain('Có sự kiện thanh toán mâu thuẫn đang được rà soát; đơn của bạn vẫn đã được xác nhận.')
    expect(markup).not.toContain('Đơn chưa được xác nhận')
  })

  it('keeps the late-payment wording for terminal non-confirmed orders', () => {
    for (const status of ['expired', 'cancelled', 'payment_failed'] as const) {
      const markup = render(<OrderDetailView {...detailProps({ order: makeOrder({ status, payable: false, reconciliation_required: true }) })} />)
      expect(markup).toContain('đến muộn hoặc không khớp')
      expect(markup).toContain('Đơn chưa được xác nhận')
    }
  })
})
