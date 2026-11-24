import { API_BASE_URL } from '@/api/base-url'
import { createJsonHttp, type JsonHttpOptions } from '@/lib/json-http'

import type { Order, OrderListParams, OrderPage, OrderStatus } from './bill.type'

export function buildOrderListQuery(params: OrderListParams = {}): string {
  const query = new URLSearchParams()
  if (params.status) query.set('status', params.status)
  if (params.cursor) query.set('cursor', params.cursor)
  if (params.limit) query.set('limit', String(params.limit))
  const text = query.toString()
  return text === '' ? '' : `?${text}`
}

/** Order API client; the transport can be stubbed through `options.fetcher`. */
export function createBillClient(options: JsonHttpOptions) {
  const http = createJsonHttp(options)
  const orderPath = (orderId: string) => `/orders/${encodeURIComponent(orderId)}`

  return {
    listOrders: (params?: OrderListParams) => http.json<OrderPage>(`/orders${buildOrderListQuery(params)}`, { method: 'GET' }),
    getOrder: (orderId: string) => http.json<Order>(orderPath(orderId), { method: 'GET' }),
    /** Repeated cancel returns 200 with the order; 409 STATE_TRANSITION_CONFLICT means the order already left pending_payment. */
    cancelOrder: (orderId: string) => http.json<Order>(`${orderPath(orderId)}/cancel`, { method: 'POST' }),
  }
}

export const billClient = createBillClient({ baseURL: API_BASE_URL })

export const ORDER_PAGE_SIZE = 20

/** Binds a status filter so `usePagedList` can fetch (and append) pages by cursor. */
export function createOrderPageFetcher(client: Pick<typeof billClient, 'listOrders'>, status?: OrderStatus) {
  return (cursor?: string) => client.listOrders({ cursor, status, limit: ORDER_PAGE_SIZE })
}
