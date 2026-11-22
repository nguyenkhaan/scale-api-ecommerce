import { StatusBadge, Surface } from '@/components/common'
import { OrderCountdown } from '@/features/bill/components/OrderCountdown'
import { canPayOrder } from '@/features/bill/bill.logic'
import type { Order } from '@/features/bill/bill.type'
import { formatDateTime, formatVND } from '@/lib/format'

import type { PaymentSession } from '../payment.type'

type Props = {
  order: Order
  session: PaymentSession | null
  onDeadlineReached: () => void
}

export function PaymentSummary({ order, session, onDeadlineReached }: Props) {
  return (
    <Surface as="section" aria-labelledby="payment-session-heading" className="p-4">
      <h2 id="payment-session-heading" className="text-lg font-semibold text-[var(--ss-color-text)]">Phiên thanh toán</h2>
      <dl className="mt-3 grid gap-3 text-sm sm:grid-cols-2">
        <Row label="Mã đơn" value={order.order_number} />
        <Row label="Số tiền" value={formatVND(session?.amount ?? order.total_amount)} />
        <Row label="Mã tham chiếu phiên" value={session?.simulator_reference ?? 'Đang tạo phiên…'} />
        <div>
          <dt className="text-[var(--ss-color-muted)]">Hạn thanh toán</dt>
          <dd className="font-medium text-[var(--ss-color-text)]">
            {formatDateTime(order.expires_at)}
            {canPayOrder(order) && <> · <OrderCountdown expiresAt={order.expires_at} onExpire={onDeadlineReached} /></>}
          </dd>
          <dd className="text-xs text-[var(--ss-color-muted)]">Thời gian chính thức theo hệ thống</dd>
        </div>
      </dl>
      <div className="mt-4 border-t border-[var(--ss-color-border)] pt-3">
        <p className="text-sm font-medium text-[var(--ss-color-text)]">Trạng thái chính thức từ hệ thống</p>
        <div className="mt-2 flex flex-wrap items-center gap-4 text-sm text-[var(--ss-color-muted)]">
          <span className="flex items-center gap-2">Đơn: <StatusBadge domain="order" value={order.status} /></span>
          {order.payment_status && (
            <span className="flex items-center gap-2">Phiên: <StatusBadge domain="payment" value={order.payment_status} /></span>
          )}
        </div>
      </div>
    </Surface>
  )
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-[var(--ss-color-muted)]">{label}</dt>
      <dd className="break-all font-medium text-[var(--ss-color-text)]">{value}</dd>
    </div>
  )
}
