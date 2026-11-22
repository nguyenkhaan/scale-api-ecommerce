import { Link, useParams } from 'react-router-dom'

import { Alert, Button, FeedbackState } from '@/components/common'
import { isApiError } from '@/lib/api-error'
import { useOrderQuery } from '@/features/bill/bill.hook'
import { describeOrderLoadError } from '@/features/bill/bill.logic'
import type { Order } from '@/features/bill/bill.type'
import { OrderStateNotices } from '@/features/bill/components/OrderStateNotices'

import { PaymentOutcome } from '../components/PaymentOutcome'
import { PaymentSummary } from '../components/PaymentSummary'
import { usePaymentSession, type PaymentSessionState } from '../payment.hook'
import { selectPaymentView } from '../payment.logic'

export type PaymentPageViewProps = {
  order: Order | undefined
  loading: boolean
  loadError: string | null
  notFound: boolean
  sessionState: PaymentSessionState
  onRetryLoad: () => void
}

export function PaymentPage() {
  const { orderId = '' } = useParams()
  const query = useOrderQuery(orderId, { poll: true })
  const { refetch } = query
  const sessionState = usePaymentSession(orderId, query.data, () => void refetch())

  return (
    <PaymentPageView
      order={query.data}
      loading={query.isPending}
      loadError={query.error ? describeOrderLoadError(query.error) : null}
      notFound={isApiError(query.error) && query.error.status === 404}
      sessionState={sessionState}
      onRetryLoad={() => void refetch()}
    />
  )
}

export function PaymentPageView({ order, loading, loadError, notFound, sessionState, onRetryLoad }: PaymentPageViewProps) {
  const failure = sessionState.failure
  return (
    <div className="mx-auto max-w-3xl space-y-4">
      <div className="border-b border-[var(--ss-color-border)] pb-4">
        <h1 className="text-2xl font-semibold tracking-tight text-[var(--ss-color-text)]">Thanh toán mô phỏng</h1>
        <p className="mt-1 text-sm text-[var(--ss-color-muted)]">Trạng thái hiển thị theo hệ thống và tự cập nhật.</p>
      </div>

      {loading && <FeedbackState kind="loading" title="Đang tải thông tin thanh toán" message="Vui lòng chờ trong giây lát." />}
      {!loading && notFound && (
        <FeedbackState kind="notFound" title="Không tìm thấy đơn hàng" message="Đơn hàng không tồn tại hoặc không thuộc tài khoản của bạn.">
          <Link to="/orders" className="text-sm font-semibold text-[var(--ss-color-primary-strong)] hover:underline">Về đơn của tôi</Link>
        </FeedbackState>
      )}
      {!loading && !notFound && loadError && !order && (
        <FeedbackState kind="error" title="Không thể tải thông tin thanh toán" message={loadError} action={{ label: 'Thử lại', onClick: onRetryLoad }} />
      )}
      {loadError && order && <Alert tone="warning">Chưa cập nhật được trạng thái mới nhất. Hệ thống sẽ tự thử lại.</Alert>}

      {order && (
        <>
          <PaymentSummary order={order} session={sessionState.session} onDeadlineReached={onRetryLoad} />
          {failure && (
            <Alert tone={failure.kind === 'not-payable' ? 'warning' : 'danger'}>
              <p>{failure.text}</p>
              {failure.kind === 'error' && <Button variant="secondary" className="mt-2" onClick={sessionState.retry}>Thử lại</Button>}
            </Alert>
          )}
          <OrderStateNotices order={order} showOverdue={false} />
          <PaymentOutcome kind={selectPaymentView(order)} order={order} />
        </>
      )}
    </div>
  )
}
