import { Alert } from '@/components/common'

import { orderAvailability } from '../bill.logic'
import type { Order } from '../bill.type'

/** Neutral notices derived from the official order state; never claims a purchase that is not confirmed. */
export function OrderStateNotices({ order, showOverdue = true }: { order: Order; showOverdue?: boolean }) {
  return (
    <>
      {showOverdue && orderAvailability(order) === 'overdue' && (
        <Alert tone="warning" role="status">
          <p className="font-semibold">Đã quá hạn thanh toán, hệ thống đang giải phóng hàng</p>
          <p className="mt-1">Đơn này không thể thanh toán nữa. Trạng thái đơn sẽ được cập nhật sau ít phút.</p>
        </Alert>
      )}
      {order.reconciliation_required && (
        <Alert tone="info" role="status">
          {order.status === 'confirmed'
            ? 'Có sự kiện thanh toán mâu thuẫn đang được rà soát; đơn của bạn vẫn đã được xác nhận.'
            : 'Khoản thanh toán của đơn này đến muộn hoặc không khớp với đơn hàng và đang được kiểm tra. Đơn chưa được xác nhận bởi khoản thanh toán này.'}
        </Alert>
      )}
    </>
  )
}
