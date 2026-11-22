import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'

import { Alert, Button, FeedbackState, StatusBadge, Surface } from '@/components/common'
import { isApiError } from '@/lib/api-error'
import { formatDateTime, formatVND } from '@/lib/format'

import { useCancelOrder, useOrderQuery, type CancelNotice } from '../bill.hook'
import { canCancelOrder, canPayOrder, describeOrderLoadError } from '../bill.logic'
import type { Order } from '../bill.type'
import { CancelOrderDialog } from '../components/CancelOrderDialog'
import { OrderCountdown } from '../components/OrderCountdown'
import { OrderItemsList } from '../components/OrderItemsList'
import { OrderStateNotices } from '../components/OrderStateNotices'

export type OrderDetailViewProps = {
  order: Order | undefined
  loading: boolean
  loadError: string | null
  notFound: boolean
  notice: CancelNotice | null
  cancelOpen: boolean
  cancelling: boolean
  onRetry: () => void
  onOpenCancel: () => void
  onCloseCancel: () => void
  onConfirmCancel: () => void
  onDeadlineReached?: () => void
}

export function OrderDetailPage() {
  const { orderId = '' } = useParams()
  const query = useOrderQuery(orderId)
  const { cancel, cancelling, notice } = useCancelOrder(orderId)
  const [cancelOpen, setCancelOpen] = useState(false)

  const confirmCancel = async () => {
    await cancel()
    setCancelOpen(false)
  }

  return (
    <OrderDetailView
      order={query.data}
      loading={query.isPending}
      loadError={query.error ? describeOrderLoadError(query.error) : null}
      notFound={isApiError(query.error) && query.error.status === 404}
      notice={notice}
      cancelOpen={cancelOpen}
      cancelling={cancelling}
      onRetry={() => void query.refetch()}
      onOpenCancel={() => setCancelOpen(true)}
      onCloseCancel={() => setCancelOpen(false)}
      onConfirmCancel={() => void confirmCancel()}
      onDeadlineReached={() => void query.refetch()}
    />
  )
}

export function OrderDetailView(props: OrderDetailViewProps) {
  const { order, loading, loadError, notFound } = props
  return (
    <div className="mx-auto max-w-4xl">
      <Link to="/orders" className="inline-flex min-h-11 items-center text-sm font-semibold text-[var(--ss-color-primary-strong)] hover:underline">
        ← Đơn của tôi
      </Link>
      {loading && <FeedbackState kind="loading" title="Đang tải đơn hàng" message="Vui lòng chờ trong giây lát." />}
      {!loading && notFound && <FeedbackState kind="notFound" title="Không tìm thấy đơn hàng" message="Đơn hàng không tồn tại hoặc không thuộc tài khoản của bạn." />}
      {!loading && !notFound && loadError && (
        <FeedbackState kind="error" title="Không thể tải đơn hàng" message={loadError} action={{ label: 'Thử lại', onClick: props.onRetry }} />
      )}
      {order && <OrderDetailBody {...props} order={order} />}
    </div>
  )
}

function OrderDetailBody({ order, notice, cancelOpen, cancelling, onOpenCancel, onCloseCancel, onConfirmCancel, onDeadlineReached }: OrderDetailViewProps & { order: Order }) {
  const payable = canPayOrder(order)
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-4 border-b border-[var(--ss-color-border)] pb-4">
        <div className="space-y-1">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-2xl font-semibold tracking-tight text-[var(--ss-color-text)]">Đơn {order.order_number}</h1>
            <StatusBadge domain="order" value={order.status} />
          </div>
          <p className="text-sm text-[var(--ss-color-muted)]">Đặt lúc {formatDateTime(order.created_at)}</p>
          <p className="text-sm text-[var(--ss-color-muted)]">
            Hạn thanh toán: {formatDateTime(order.expires_at)}
            {payable && <> · <OrderCountdown expiresAt={order.expires_at} onExpire={onDeadlineReached} /></>}
          </p>
          <p className="text-xs text-[var(--ss-color-muted)]">Thời gian chính thức theo hệ thống</p>
          {order.payment_status && (
            <p className="flex items-center gap-2 text-sm text-[var(--ss-color-muted)]">
              Thanh toán: <StatusBadge domain="payment" value={order.payment_status} />
            </p>
          )}
        </div>
        {(payable || canCancelOrder(order)) && (
          <div className="flex flex-wrap gap-2">
            <Link
              to={`/orders/${order.id}/payment`}
              className="inline-flex min-h-11 items-center rounded-[var(--ss-radius-control)] bg-[var(--ss-color-primary)] px-4 py-2 text-sm font-semibold text-white hover:bg-[var(--ss-color-primary-strong)]"
            >
              Thanh toán
            </Link>
            <Button variant="secondary" className="min-h-11" onClick={onOpenCancel}>Hủy đơn</Button>
          </div>
        )}
      </div>

      {notice && <Alert tone={notice.kind === 'conflict' ? 'warning' : 'danger'}>{notice.text}</Alert>}
      <OrderStateNotices order={order} />
      <OrderItemsList items={order.items} />

      <div className="grid gap-4 sm:grid-cols-2">
        <Surface as="section" aria-labelledby="order-recipient-heading" className="p-4">
          <h2 id="order-recipient-heading" className="text-lg font-semibold text-[var(--ss-color-text)]">Người nhận</h2>
          <p className="mt-2 font-medium text-[var(--ss-color-text)]">{order.recipient.name}</p>
          <p className="text-sm text-[var(--ss-color-muted)]">{order.recipient.phone}</p>
          <p className="text-sm text-[var(--ss-color-muted)]">{order.recipient.address}</p>
        </Surface>
        <Surface as="section" aria-labelledby="order-total-heading" className="p-4">
          <h2 id="order-total-heading" className="text-lg font-semibold text-[var(--ss-color-text)]">Tổng tiền</h2>
          <p className="mt-2 text-2xl font-semibold text-[var(--ss-color-primary-strong)]">{formatVND(order.total_amount)}</p>
        </Surface>
      </div>

      <CancelOrderDialog open={cancelOpen} busy={cancelling} onConfirm={onConfirmCancel} onClose={onCloseCancel} />
    </div>
  )
}
