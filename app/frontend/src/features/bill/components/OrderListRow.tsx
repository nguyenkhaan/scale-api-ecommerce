import { Link } from 'react-router-dom'

import { StatusBadge, Surface } from '@/components/common'
import { formatDateTime, formatVND } from '@/lib/format'

import { canPayOrder, totalQuantity } from '../bill.logic'
import type { Order } from '../bill.type'
import { OrderCountdown } from './OrderCountdown'

export function OrderListRow({ order }: { order: Order }) {
  return (
    <li>
      <Surface as="article" className="flex flex-wrap items-center justify-between gap-3 p-4">
        <div className="min-w-0 space-y-1">
          <p className="font-semibold text-[var(--ss-color-text)]">{order.order_number}</p>
          <p className="text-sm text-[var(--ss-color-muted)]">
            Đặt lúc {formatDateTime(order.created_at)} · {totalQuantity(order)} sản phẩm
          </p>
          <div className="flex flex-wrap items-center gap-2">
            <StatusBadge domain="order" value={order.status} />
            {canPayOrder(order) && <OrderCountdown expiresAt={order.expires_at} />}
          </div>
        </div>
        <div className="flex items-center gap-4">
          <span className="font-semibold text-[var(--ss-color-text)]">{formatVND(order.total_amount)}</span>
          <Link
            to={`/orders/${order.id}`}
            aria-label={`Xem đơn ${order.order_number}`}
            className="inline-flex min-h-11 items-center text-sm font-semibold text-[var(--ss-color-primary-strong)] hover:underline"
          >
            Xem đơn
          </Link>
        </div>
      </Surface>
    </li>
  )
}
