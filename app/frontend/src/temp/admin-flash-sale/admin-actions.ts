import { isApiError } from '../lib/api-error'
import type { FlashSaleStatus } from '../lib/flash-sale-api'
import { describeCartFailure } from '../cart/cart-messages'
import { violationMessage } from '../lib/violation-messages'

export type AdminAction = 'edit' | 'schedule' | 'cancel_schedule' | 'stop'

/** draft: edit + schedule; scheduled (upcoming): cancel schedule; active: stop; ended/stopped: view only. */
export function adminActionsFor(status: FlashSaleStatus): AdminAction[] {
  switch (status) {
    case 'draft':
      return ['edit', 'schedule']
    case 'scheduled':
      return ['cancel_schedule']
    case 'active':
      return ['stop']
    default:
      return []
  }
}

export const adminActionLabels: Record<AdminAction, string> = {
  edit: 'Chỉnh sửa',
  schedule: 'Lên lịch',
  cancel_schedule: 'Hủy lịch',
  stop: 'Dừng chiến dịch',
}

export const STATE_CONFLICT_MESSAGE =
  'Trạng thái chiến dịch đã thay đổi; danh sách đã được tải lại. Vui lòng kiểm tra rồi thử lại.'

export const SCHEDULE_OVERLAP_MESSAGE =
  'Có SKU đang nằm trong chiến dịch khác (đã lên lịch hoặc đã dừng) trùng khung giờ. Hãy đổi khung giờ hoặc bỏ SKU trùng.'

export type AdminFailure = { text: string; stateConflict: boolean }

export function describeAdminFailure(err: unknown): AdminFailure {
  if (isApiError(err)) {
    if (err.code === 'STATE_TRANSITION_CONFLICT') return { text: STATE_CONFLICT_MESSAGE, stateConflict: true }
    if (err.code === 'SALE_SCHEDULE_OVERLAP') return { text: SCHEDULE_OVERLAP_MESSAGE, stateConflict: false }
    if (err.status === 401 || err.status === 403) {
      return { text: 'Phiên đăng nhập đã hết hạn hoặc bạn không có quyền quản trị.', stateConflict: false }
    }
    if (err.code === 'RESOURCE_NOT_FOUND') {
      return { text: 'Chiến dịch không còn tồn tại; danh sách đã được tải lại.', stateConflict: true }
    }
    if (err.code === 'VALIDATION_ERROR' && err.violations.length > 0) {
      return { text: err.violations.map(violationMessage).join(' '), stateConflict: false }
    }
  }
  return { text: describeCartFailure(err).text, stateConflict: false }
}
