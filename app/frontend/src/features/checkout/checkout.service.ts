import { API_BASE_URL } from '@/api/base-url'
import { createJsonHttp, type JsonHttpOptions } from '@/lib/json-http'

import type { IdempotencyResult, Order, PlaceOrderRequest } from './checkout.type'

export function createCheckoutClient(options: JsonHttpOptions) {
  const http = createJsonHttp(options)

  return {
    async placeOrder(key: string, request: PlaceOrderRequest): Promise<{ order: Order; replayed: boolean }> {
      const response = await http.withHeaders<Order>('/orders', {
        method: 'POST',
        body: request,
        headers: { 'Idempotency-Key': key },
        timeoutMs: 15_000,
      })
      return {
        order: response.data,
        replayed: response.headers.get('Idempotency-Replayed')?.toLowerCase() === 'true',
      }
    },
    getAttempt(key: string): Promise<IdempotencyResult> {
      return http.json<IdempotencyResult>(`/idempotency/orders/${encodeURIComponent(key)}`, { method: 'GET' })
    },
  }
}

export const checkoutClient = createCheckoutClient({ baseURL: API_BASE_URL })
