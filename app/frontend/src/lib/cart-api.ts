import { createJsonHttp, type JsonHttpOptions } from './json-http'

export type PriceSource = 'regular' | 'flash_sale'

export type CartItem = {
  id: string
  sku_id: string
  quantity: number
  version: number
  price_source: PriceSource
  flash_sale_item_id: string | null
  product_name: string
  sku_code: string
  /** Display-only price; the authoritative price comes from the checkout preview. */
  reference_unit_price: number
}

export type Cart = {
  id: string
  items: CartItem[]
}

export type UpsertCartItemRequest = {
  quantity: number
  price_source?: PriceSource
  flash_sale_item_id?: string | null
  expected_version?: number
}

export type SelectedCartItem = {
  cart_item_id: string
  version: number
}

export type CheckoutSelection = {
  items: SelectedCartItem[]
}

export type CheckoutLine = {
  cart_item_id: string
  cart_item_version: number
  sku_id: string
  quantity: number
  unit_price: number
  subtotal: number
  price_source: PriceSource
  flash_sale_item_id: string | null
}

export type CheckoutPreview = {
  currency: string
  total_amount: number
  lines: CheckoutLine[]
  previewed_at: string
}

export type CartClientOptions = JsonHttpOptions

export function createCartClient(options: CartClientOptions) {
  const http = createJsonHttp(options)

  return {
    getCart: () => http.json<Cart>('/api/v1/cart', { method: 'GET' }),

    /** Sets the absolute quantity of one SKU line. */
    upsertCartItem: (skuId: string, data: UpsertCartItemRequest) =>
      http.json<Cart>(`/api/v1/cart/items/${skuId}`, { method: 'PUT', body: data }),

    /** The server answers 204 No Content; callers reload the cart with getCart(). */
    removeCartItem: (skuId: string) =>
      http.noContent(`/api/v1/cart/items/${skuId}`, { method: 'DELETE' }),

    previewCheckout: (selection: CheckoutSelection) =>
      http.json<CheckoutPreview>('/api/v1/checkout/preview', { method: 'POST', body: selection }),
  }
}

export type CartClient = ReturnType<typeof createCartClient>

export const cartClient = createCartClient({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
})
