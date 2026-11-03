import { errorFromResponse, invalidResponseError, networkError } from './api-error'
import { csrfCookieName, readDocumentCookie } from './auth-api'

export type CartItem = {
  id: string
  sku_id: string
  quantity: number
  version: number
  price_source: 'regular' | 'flash_sale'
  flash_sale_item_id: string | null
}

export type Cart = {
  id: string
  items: CartItem[]
}

export type UpsertCartItemRequest = {
  quantity: number
  price_source?: 'regular' | 'flash_sale'
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
  price_source: string
  flash_sale_item_id: string | null
}

export type CheckoutPreview = {
  currency: string
  total_amount: number
  lines: CheckoutLine[]
  previewed_at: string
}

type Fetcher = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>

export type CartClientOptions = {
  baseURL: string
  fetcher?: Fetcher
  readCookie?: (name: string) => string | null
}

export function createCartClient(options: CartClientOptions) {
  const baseURL = options.baseURL.replace(/\/$/, '')
  const fetcher: Fetcher = options.fetcher ?? ((input, init) => globalThis.fetch(input, init))
  const readCookie = options.readCookie ?? readDocumentCookie

  async function send(path: string, init: { method: string; body?: unknown; csrf?: string }) {
    const headers: Record<string, string> = { Accept: 'application/json' }
    if (init.body !== undefined) headers['Content-Type'] = 'application/json'
    if (init.csrf !== undefined) headers['X-CSRF-Token'] = init.csrf
    try {
      return await fetcher(`${baseURL}${path}`, {
        method: init.method,
        credentials: 'include',
        headers,
        body: init.body === undefined ? undefined : JSON.stringify(init.body),
      })
    } catch {
      throw networkError()
    }
  }

  async function readJSON<T>(response: Response): Promise<T> {
    try {
      return (await response.json()) as T
    } catch {
      throw invalidResponseError(response)
    }
  }

  function getCsrf() {
    return readCookie(csrfCookieName) ?? ''
  }

  return {
    async getCart(): Promise<Cart> {
      const res = await send('/api/v1/cart', { method: 'GET' })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<Cart>(res)
    },

    async upsertCartItem(skuId: string, data: UpsertCartItemRequest): Promise<Cart> {
      const res = await send(`/api/v1/cart/items/${skuId}`, {
        method: 'PUT',
        body: data,
        csrf: getCsrf(),
      })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<Cart>(res)
    },

    async removeCartItem(skuId: string): Promise<Cart> {
      const res = await send(`/api/v1/cart/items/${skuId}`, {
        method: 'DELETE',
        csrf: getCsrf(),
      })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<Cart>(res)
    },

    async previewCheckout(selection: CheckoutSelection): Promise<CheckoutPreview> {
      const res = await send('/api/v1/checkout/preview', {
        method: 'POST',
        body: selection,
        csrf: getCsrf(),
      })
      if (!res.ok) throw await errorFromResponse(res)
      return readJSON<CheckoutPreview>(res)
    },
  }
}

export const cartClient = createCartClient({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
})
