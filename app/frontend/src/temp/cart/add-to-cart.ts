import { isApiError } from '../lib/api-error'
import type { Cart, CartClient, CartItem, PriceSource, UpsertCartItemRequest } from '../lib/cart-api'

export type AddToCartInput = {
  skuId: string
  /** How many units to add on top of what the cart already holds. */
  amount: number
  priceSource: PriceSource
  flashSaleItemId?: string
}

export type AddToCartPlan = {
  request: UpsertCartItemRequest
  /** Price source of the existing line when this add changes it; null otherwise. */
  switchedFrom: PriceSource | null
}

/** The PUT sets an absolute quantity, so a re-add must send existing quantity + amount, guarded by the line version. */
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

export type AddToCartResult = {
  cart: Cart
  quantity: number
  switchedFrom: PriceSource | null
}

/** Reads the cart, then upserts. A concurrent change of the line is retried once with fresh data. */
export async function addToCart(
  client: Pick<CartClient, 'getCart' | 'upsertCartItem'>,
  input: AddToCartInput,
): Promise<AddToCartResult> {
  for (let attempt = 0; ; attempt++) {
    const current = await client.getCart()
    const plan = planAddToCart(
      current.items.find((item) => item.sku_id === input.skuId),
      input,
    )
    try {
      const cart = await client.upsertCartItem(input.skuId, plan.request)
      return { cart, quantity: plan.request.quantity, switchedFrom: plan.switchedFrom }
    } catch (err) {
      if (attempt === 0 && isApiError(err) && err.code === 'CART_ITEM_CHANGED') continue
      throw err
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
