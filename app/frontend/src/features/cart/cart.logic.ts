import { isCartApiError } from './cart.service'
import type { Cart, CartItem, CheckoutPreview, LineError, PriceSource, UpsertCartItemRequest } from './cart.type'
import { formatVND } from '@/lib/format'
import { violationMessage } from '@/lib/violation-messages'

export type AddToCartInput = {
  skuId: string
  amount: number
  priceSource: PriceSource
  flashSaleItemId?: string
}

export type AddToCartPlan = {
  request: UpsertCartItemRequest
  switchedFrom: PriceSource | null
}

export function planAddToCart(existing: CartItem | undefined, input: AddToCartInput): AddToCartPlan {
  const request: UpsertCartItemRequest = {
    quantity: (existing?.quantity ?? 0) + input.amount,
    price_source: input.priceSource,
  }
  if (input.priceSource === 'flash_sale') request.flash_sale_item_id = input.flashSaleItemId ?? null
  if (existing) request.expected_version = existing.version
  return {
    request,
    switchedFrom: existing && existing.price_source !== input.priceSource ? existing.price_source : null,
  }
}

type CartReaderWriter = {
  getCart: () => Promise<Cart>
  upsertCartItem: (skuId: string, data: UpsertCartItemRequest) => Promise<Cart>
}

export type AddToCartResult = {
  cart: Cart
  quantity: number
  switchedFrom: PriceSource | null
}

export async function addToCart(client: CartReaderWriter, input: AddToCartInput): Promise<AddToCartResult> {
  for (let attempt = 0; ; attempt++) {
    const current = await client.getCart()
    const plan = planAddToCart(
      current.items.find((item) => item.sku_id === input.skuId),
      input,
    )
    try {
      const cart = await client.upsertCartItem(input.skuId, plan.request)
      return { cart, quantity: plan.request.quantity, switchedFrom: plan.switchedFrom }
    } catch (error) {
      if (attempt === 0 && isCartApiError(error) && error.code === 'CART_ITEM_CHANGED') continue
      throw error
    }
  }
}

const priceSourceLabels: Record<PriceSource, string> = { regular: 'giá thường', flash_sale: 'giá Flash Sale' }

export function describeAddResult(result: AddToCartResult, productName: string): string {
  const base = `Đã thêm "${productName}" vào giỏ hàng (số lượng trong giỏ: ${result.quantity}).`
  if (result.switchedFrom === null) return base
  const target: PriceSource = result.switchedFrom === 'regular' ? 'flash_sale' : 'regular'
  return `${base} Lưu ý: dòng này trước đó dùng ${priceSourceLabels[result.switchedFrom]}, nay đã chuyển sang ${priceSourceLabels[target]}.`
}

export const NO_HOLD_SUMMARY = 'Chưa giữ hàng nào; giỏ hàng vẫn giữ nguyên.'
export const CART_CHANGED_MESSAGE =
  'Giỏ hàng đã thay đổi ở nơi khác nên đã được tải lại. Vui lòng kiểm tra rồi xem trước lại.'

const regularFallbackCodes = new Set(['SALE_NOT_ACTIVE', 'SALE_QUOTA_EXCEEDED', 'PURCHASE_LIMIT_EXCEEDED'])

export function canSwitchToRegular(priceSource: string, code: string): boolean {
  return priceSource === 'flash_sale' && regularFallbackCodes.has(code)
}

type LineErrorLike = Omit<Pick<LineError, 'code' | 'available_quantity' | 'current_unit_price'>, 'code'> & {
  code: string
}

export function lineErrorMessage(error: LineErrorLike): string {
  const remaining = error.available_quantity
  switch (error.code) {
    case 'PURCHASE_LIMIT_EXCEEDED':
      return remaining === undefined ? 'Vượt giới hạn mua' : `Vượt giới hạn mua — còn được mua ${remaining}`
    case 'SALE_QUOTA_EXCEEDED':
      return remaining === undefined ? 'Đã hết suất giá flash sale' : `Đã hết suất giá flash sale — còn ${remaining}`
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
  cartChanged: boolean
  unauthenticated: boolean
  lineErrors: LineError[]
  lineCode?: string
}

const singleLineCodes = new Set([
  'ITEM_UNAVAILABLE',
  'SALE_NOT_ACTIVE',
  'SALE_QUOTA_EXCEEDED',
  'PURCHASE_LIMIT_EXCEEDED',
  'INSUFFICIENT_STOCK',
  'PRICE_CHANGED',
])

export function describeCartFailure(error: unknown): CartFailure {
  const base = { cartChanged: false, unauthenticated: false, lineErrors: [] as LineError[] }
  if (!isCartApiError(error)) return { ...base, text: 'Có lỗi không mong muốn. Vui lòng thử lại.' }
  if (error.code === 'NETWORK_ERROR') {
    return { ...base, text: 'Không thể kết nối tới máy chủ. Vui lòng kiểm tra mạng và thử lại.' }
  }
  if (error.status === 401) return { ...base, unauthenticated: true, text: 'Vui lòng đăng nhập để tiếp tục.' }
  if (error.status === 429) {
    const wait = error.retryAfterSeconds === undefined ? '' : ` Thử lại sau ${error.retryAfterSeconds} giây.`
    return { ...base, text: `Bạn thao tác quá nhanh.${wait}` }
  }
  if (error.code === 'CART_ITEM_CHANGED') return { ...base, cartChanged: true, text: CART_CHANGED_MESSAGE }
  if (error.lineErrors.length > 0) {
    return {
      ...base,
      lineErrors: error.lineErrors,
      text: `${NO_HOLD_SUMMARY} Một số sản phẩm cần xử lý — xem thông báo ở từng dòng.`,
    }
  }
  if (error.code === 'VALIDATION_ERROR' && error.violations.length > 0) {
    return { ...base, text: error.violations.map(violationMessage).join(' ') }
  }
  if (singleLineCodes.has(error.code)) return { ...base, lineCode: error.code, text: lineErrorMessage({ code: error.code }) }
  return { ...base, text: 'Không thể hoàn tất thao tác. Vui lòng thử lại sau.' }
}

export type PreviewChange = {
  sku_id: string
  oldUnitPrice: number
  newUnitPrice: number
  oldSource: PriceSource
  newSource: PriceSource
}

export function detectPreviewChanges(previous: CheckoutPreview | null, next: CheckoutPreview): PreviewChange[] {
  if (!previous) return []
  const before = new Map(previous.lines.map((line) => [line.sku_id, line]))
  return next.lines.flatMap((line): PreviewChange[] => {
    const old = before.get(line.sku_id)
    if (!old || (old.unit_price === line.unit_price && old.price_source === line.price_source)) return []
    return [{
      sku_id: line.sku_id,
      oldUnitPrice: old.unit_price,
      newUnitPrice: line.unit_price,
      oldSource: old.price_source,
      newSource: line.price_source,
    }]
  })
}

export function selectionSignature(items: CartItem[]): string {
  return items
    .map((item) => `${item.id}:${item.version}`)
    .sort()
    .join('|')
}
