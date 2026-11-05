import { isApiError, type LineError } from '../lib/api-error'
import { formatVND } from '../lib/format'
import { violationMessage } from '../lib/violation-messages'

export const NO_HOLD_SUMMARY = 'Chưa giữ hàng nào; giỏ hàng vẫn giữ nguyên.'
export const CART_CHANGED_MESSAGE =
  'Giỏ hàng đã thay đổi ở nơi khác nên đã được tải lại. Vui lòng kiểm tra rồi xem trước lại.'

/** Codes after which the customer may explicitly choose to pay the regular price instead. */
const regularFallbackCodes = new Set(['SALE_NOT_ACTIVE', 'SALE_QUOTA_EXCEEDED', 'PURCHASE_LIMIT_EXCEEDED'])

export function canSwitchToRegular(priceSource: string, code: string): boolean {
  return priceSource === 'flash_sale' && regularFallbackCodes.has(code)
}

type LineErrorLike = Pick<LineError, 'code' | 'available_quantity' | 'current_unit_price'>

export function lineErrorMessage(error: LineErrorLike): string {
  const remaining = error.available_quantity
  switch (error.code) {
    case 'PURCHASE_LIMIT_EXCEEDED':
      return remaining === undefined ? 'Vượt giới hạn mua' : `Vượt giới hạn mua — còn được mua ${remaining}`
    case 'SALE_QUOTA_EXCEEDED':
      return remaining === undefined
        ? 'Đã hết suất giá flash sale'
        : `Đã hết suất giá flash sale — còn ${remaining}`
    case 'SALE_NOT_ACTIVE':
      return 'Chương trình chưa mở hoặc đã kết thúc'
    case 'ITEM_UNAVAILABLE':
      return 'Sản phẩm đã ngừng bán'
    case 'INSUFFICIENT_STOCK':
      return remaining === undefined ? 'Không đủ tồn kho' : `Chỉ còn ${remaining} sản phẩm`
    case 'PRICE_CHANGED':
      return error.current_unit_price === undefined
        ? 'Giá đã thay đổi'
        : `Giá đã thay đổi — giá hiện tại ${formatVND(error.current_unit_price)}`
    default:
      return 'Dòng hàng này không thể tiếp tục'
  }
}

export type CartFailure = {
  text: string
  /** CART_ITEM_CHANGED: the caller must reload the cart. */
  cartChanged: boolean
  unauthenticated: boolean
  lineErrors: LineError[]
  /** A single-SKU rejection code the caller may attach to the line it was acting on. */
  lineCode?: string
}

const singleLineCodes = new Set(['ITEM_UNAVAILABLE', 'SALE_NOT_ACTIVE', 'SALE_QUOTA_EXCEEDED', 'PURCHASE_LIMIT_EXCEEDED', 'INSUFFICIENT_STOCK', 'PRICE_CHANGED'])

/** Turns any thrown value into Vietnamese copy plus the structured hints the UI acts on. Never exposes raw codes. */
export function describeCartFailure(err: unknown): CartFailure {
  const base = { cartChanged: false, unauthenticated: false, lineErrors: [] as LineError[] }
  if (!isApiError(err)) {
    return { ...base, text: 'Có lỗi không mong muốn. Vui lòng thử lại.' }
  }
  if (err.code === 'NETWORK_ERROR') {
    return { ...base, text: 'Không thể kết nối tới máy chủ. Vui lòng kiểm tra mạng và thử lại.' }
  }
  if (err.status === 401) {
    return { ...base, unauthenticated: true, text: 'Vui lòng đăng nhập để tiếp tục.' }
  }
  if (err.status === 429) {
    const wait = err.retryAfterSeconds === undefined ? '' : ` Thử lại sau ${err.retryAfterSeconds} giây.`
    return { ...base, text: `Bạn thao tác quá nhanh.${wait}` }
  }
  if (err.code === 'CART_ITEM_CHANGED') {
    return { ...base, cartChanged: true, text: CART_CHANGED_MESSAGE }
  }
  if (err.lineErrors.length > 0) {
    return {
      ...base,
      lineErrors: err.lineErrors,
      text: `${NO_HOLD_SUMMARY} Một số sản phẩm cần xử lý — xem thông báo ở từng dòng.`,
    }
  }
  if (err.code === 'VALIDATION_ERROR' && err.violations.length > 0) {
    return { ...base, text: err.violations.map(violationMessage).join(' ') }
  }
  if (singleLineCodes.has(err.code)) {
    return { ...base, lineCode: err.code, text: lineErrorMessage({ code: err.code }) }
  }
  return { ...base, text: 'Không thể hoàn tất thao tác. Vui lòng thử lại sau.' }
}
