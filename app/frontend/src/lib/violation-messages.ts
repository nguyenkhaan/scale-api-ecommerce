import type { FieldViolation } from './api-error'

const fieldLabels: Record<string, string> = {
  name: 'Tên chiến dịch',
  starts_at: 'Thời gian bắt đầu',
  ends_at: 'Thời gian kết thúc',
  items: 'Danh sách SKU',
  sku_id: 'SKU',
  sale_price: 'Giá flash sale',
  quota_total: 'Tổng số suất',
  purchase_limit: 'Giới hạn mua',
  quantity: 'Số lượng',
  cart: 'Giỏ hàng',
  price_source: 'Loại giá',
  flash_sale_item_id: 'Sản phẩm flash sale',
  cart_item_id: 'Dòng giỏ hàng',
  version: 'Phiên bản dòng giỏ hàng',
  body: 'Dữ liệu gửi lên',
  cursor: 'Trang tiếp theo',
}

/** Server reasons are English, stable phrases; match by substring and translate. */
const reasonTranslations: Array<[string, string]> = [
  ['is required', 'là bắt buộc'],
  ['cannot be empty', 'không được để trống'],
  ['must be greater than 0', 'phải lớn hơn 0'],
  ['must be at least 1', 'phải từ 1 trở lên'],
  ['must be a valid UUID', 'không hợp lệ'],
  ['must be at most 200 characters', 'tối đa 200 ký tự'],
  ['must be strictly after starts_at', 'phải sau thời điểm bắt đầu'],
  ['must have at least one item', 'cần có ít nhất một SKU'],
  ['must contain at least one item', 'cần có ít nhất một mục'],
  ['cannot exceed 20', 'tối đa 20 mục'],
  ['contains duplicate sku_id', 'có SKU bị trùng'],
  ['duplicate cart item', 'có dòng bị chọn trùng'],
  ['do not exist or are not active', 'có SKU không tồn tại hoặc đã ngừng bán'],
  ['does not exist or is not active', 'không tồn tại hoặc đã ngừng bán'],
  ['sale_price must be strictly less', 'có giá flash sale chưa thấp hơn giá thường'],
  ['must be less than regular price', 'phải thấp hơn giá thường'],
  ['must be a whole number', 'phải là số nguyên'],
  ['exceeds maximum allowed quantity', 'vượt quá số lượng tối đa cho mỗi đơn hoặc giới hạn mua'],
  ['flash sale item is invalid or not active', 'không hợp lệ hoặc chương trình chưa mở'],
  ['must be regular or flash_sale', 'chỉ nhận giá thường hoặc flash sale'],
  ['is invalid or does not match', 'không hợp lệ'],
  ['malformed JSON', 'không đúng định dạng'],
]

/** `items[2].sale_price` -> `Dòng 3 — Giá flash sale`; unknown fields fall back to a neutral label. */
export function fieldLabel(field: string): string {
  const indexed = /^items\[(\d+)\]\.(.+)$/.exec(field)
  if (indexed) return `Dòng ${Number(indexed[1]) + 1} — ${fieldLabels[indexed[2]] ?? 'Thông tin'}`
  return fieldLabels[field] ?? 'Thông tin'
}

export function violationMessage(violation: FieldViolation): string {
  const { field, reason } = violation
  if (field === 'quantity' && reason.includes('exceeds maximum')) {
    return 'Số lượng vượt mức tối đa cho mỗi đơn hoặc vượt giới hạn mua.'
  }
  if (field === 'cart' && reason.includes('cannot exceed 20')) {
    return 'Giỏ hàng tối đa 20 SKU khác nhau.'
  }
  const translated = reasonTranslations.find(([needle]) => reason.includes(needle))?.[1]
  return `${fieldLabel(field)} ${translated ?? 'không hợp lệ'}.`
}

/** Messages for violations whose field is exactly `field`. */
export function fieldMessages(violations: FieldViolation[], field: string): string[] {
  return violations.filter((v) => v.field === field).map(violationMessage)
}

/** Violations that no rendered form control claims (so they must be shown in a summary). */
export function unclaimedMessages(violations: FieldViolation[], claimed: (field: string) => boolean): string[] {
  return violations.filter((v) => !claimed(v.field)).map(violationMessage)
}
