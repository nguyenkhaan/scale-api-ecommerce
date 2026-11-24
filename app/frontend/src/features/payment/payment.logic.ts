import { canPayOrder } from '@/features/bill/bill.logic'
import type { Order } from '@/features/bill/bill.type'
import { isApiError } from '@/lib/api-error'

/**
 * Which outcome the payment page shows. Always derived from the official order, never from the page:
 * pending + session pending/unknown stays neutral ("waiting"/"checking"), never failure wording.
 */
export type PaymentViewKind = 'loading' | 'confirmed' | 'failed' | 'cancelled' | 'expired' | 'overdue' | 'checking' | 'waiting'

export function selectPaymentView(order: Order | undefined): PaymentViewKind {
  if (!order) return 'loading'
  switch (order.status) {
    case 'confirmed':
      return 'confirmed'
    case 'payment_failed':
      return 'failed'
    case 'cancelled':
      return 'cancelled'
    case 'expired':
      return 'expired'
    case 'pending_payment':
      if (!order.payable) return 'overdue'
      return order.payment_status === 'unknown' ? 'checking' : 'waiting'
  }
}

/** A session is requested at most once, and only while the server says the order is payable. */
export function shouldRequestSession(order: Order | undefined, alreadyRequested: boolean): boolean {
  return order !== undefined && !alreadyRequested && canPayOrder(order)
}

export type SessionFailure = { kind: 'not-payable' | 'error'; text: string }

export function describeSessionFailure(err: unknown): SessionFailure {
  if (isApiError(err)) {
    if (err.status === 409 || err.code === 'ORDER_NOT_PAYABLE') {
      return { kind: 'not-payable', text: 'Đơn hàng không còn có thể thanh toán. Đang cập nhật trạng thái mới nhất.' }
    }
    if (err.status === 404) return { kind: 'error', text: 'Không tìm thấy đơn hàng.' }
    if (err.code === 'NETWORK_ERROR') return { kind: 'error', text: 'Lỗi kết nối khi tạo phiên thanh toán. Vui lòng thử lại.' }
  }
  return { kind: 'error', text: 'Không thể tạo phiên thanh toán. Vui lòng thử lại.' }
}
