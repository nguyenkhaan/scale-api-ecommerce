import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

import { Alert } from '@/components/common'
import { formatDateTime } from '@/lib/format'
import type { Order } from '@/features/bill/bill.type'

import type { PaymentViewKind } from '../payment.logic'

const linkClass = 'inline-flex min-h-11 items-center rounded-[var(--ss-radius-control)] px-4 py-2 text-sm font-semibold'
const primaryLink = `${linkClass} bg-[var(--ss-color-primary)] text-white hover:bg-[var(--ss-color-primary-strong)]`
const secondaryLink = `${linkClass} border border-[var(--ss-color-input-border)] text-[var(--ss-color-text)] hover:bg-[var(--ss-color-app-bg)]`

type Props = { kind: PaymentViewKind; order: Order }

/** Outcome panel for the official order state. Pending states never use failure wording. */
export function PaymentOutcome({ kind, order }: Props) {
  const viewOrder = (
    <Link to={`/orders/${order.id}`} className={kind === 'confirmed' ? primaryLink : secondaryLink}>
      Xem đơn
    </Link>
  )
  switch (kind) {
    case 'confirmed':
      return (
        <Outcome tone="success" title="Thanh toán thành công" text={`Đơn ${order.order_number} đã được xác nhận.`}>
          {viewOrder}
          <Link to="/products" className={secondaryLink}>Tiếp tục mua sắm</Link>
        </Outcome>
      )
    case 'failed':
      return (
        <Outcome tone="danger" title="Thanh toán thất bại" text="Đơn hàng chưa được xác nhận và hàng đã được trả lại kho. Bạn có thể đặt lại từ giỏ hàng.">
          <Link to="/cart" className={primaryLink}>Đặt lại</Link>
          {viewOrder}
        </Outcome>
      )
    case 'cancelled':
      return (
        <Outcome tone="info" title="Đơn hàng đã được hủy" text="Đơn này không còn cần thanh toán. Bạn có thể mua lại bất cứ lúc nào.">
          <Link to="/products" className={primaryLink}>Mua lại</Link>
          {viewOrder}
        </Outcome>
      )
    case 'expired':
      return (
        <Outcome tone="warning" title="Đơn hàng đã hết hạn thanh toán" text="Hàng đã được trả lại kho. Bạn có thể mua lại nếu vẫn muốn.">
          <Link to="/products" className={primaryLink}>Mua lại</Link>
          {viewOrder}
        </Outcome>
      )
    case 'overdue':
      return (
        <Outcome tone="warning" title="Đã quá hạn thanh toán, hệ thống đang giải phóng hàng" text="Đơn này không thể thanh toán nữa. Trạng thái đơn sẽ được cập nhật sau ít phút.">
          {viewOrder}
        </Outcome>
      )
    case 'checking':
      return (
        <Outcome
          tone="info"
          title="Đang chờ kết quả thanh toán…"
          text="Chưa nhận được kết quả chính thức từ cổng thanh toán. Đừng thanh toán lại để tránh bị trừ tiền hai lần. Trang sẽ tự cập nhật khi có kết quả."
        >
          {viewOrder}
        </Outcome>
      )
    case 'waiting':
      return (
        <Outcome tone="info" title="Đang chờ kết quả thanh toán…" text={`Hoàn tất thanh toán trước ${formatDateTime(order.expires_at)}. Trang sẽ tự cập nhật khi có kết quả.`}>
          {viewOrder}
        </Outcome>
      )
    case 'loading':
      return null
  }
}

type OutcomeProps = { tone: 'info' | 'success' | 'warning' | 'danger'; title: string; text: string; children: ReactNode }

function Outcome({ tone, title, text, children }: OutcomeProps) {
  return (
    <section aria-label="Kết quả thanh toán" className="space-y-3">
      <Alert tone={tone}>
        <p className="font-semibold">{title}</p>
        <p className="mt-1">{text}</p>
      </Alert>
      <div className="flex flex-wrap gap-2">{children}</div>
    </section>
  )
}
