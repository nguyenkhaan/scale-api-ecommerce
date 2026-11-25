import { describe, expect, it } from 'bun:test'

import { ApiError } from '@/lib/api-error'
import { createCheckoutClient } from '@/features/checkout/checkout.service'

import { makeOrder, stubFetcher } from '../bill/order-fixtures'

const request = {
  lines: [
    {
      cart_item_id: 'cart-1',
      cart_item_version: 3,
      sku_id: 'sku-1',
      quantity: 2,
      unit_price: 129000,
      subtotal: 258000,
      price_source: 'regular' as const,
    },
  ],
  currency: 'VND' as const,
  expected_total: 258000,
  recipient: { name: 'Nguyễn Minh Anh', phone: '0901234567', address: '12 Nguyễn Trãi' },
}

describe('checkout service', () => {
  it('places an order with idempotency and CSRF headers', async () => {
    const { fetcher, calls } = stubFetcher([{ status: 201, body: makeOrder() }])
    const client = createCheckoutClient({ baseURL: '/api', fetcher, readCookie: () => 'csrf-1' })

    const result = await client.placeOrder('key-12345', request)

    expect(result.order.id).toBe('order-1')
    expect(result.replayed).toBe(false)
    expect(calls[0]).toMatchObject({ url: '/api/orders', method: 'POST', body: request })
    expect(calls[0].headers['Idempotency-Key']).toBe('key-12345')
    expect(calls[0].headers['X-CSRF-Token']).toBe('csrf-1')
  })

  it('reports an idempotent replay from the response header', async () => {
    const { fetcher } = stubFetcher([{ status: 201, body: makeOrder() }])
    const client = createCheckoutClient({
      baseURL: '/api',
      fetcher: async (input, init) => {
        const response = await fetcher(input, init)
        return new Response(await response.text(), {
          status: 201,
          headers: { 'Content-Type': 'application/json', 'Idempotency-Replayed': 'true' },
        })
      },
      readCookie: () => 'csrf-1',
    })

    await expect(client.placeOrder('key-12345', request)).resolves.toMatchObject({ replayed: true })
  })

  it('gets an idempotency attempt without a CSRF header', async () => {
    const { fetcher, calls } = stubFetcher([
      {
        status: 200,
        body: { idempotency_key: 'key-12345', response_status: 201, response_body: makeOrder(), completed_at: '2026-11-25T08:00:00Z' },
      },
    ])
    const client = createCheckoutClient({ baseURL: '/api', fetcher, readCookie: () => 'csrf-1' })

    await client.getAttempt('key-12345')

    expect(calls[0].url).toBe('/api/idempotency/orders/key-12345')
    expect(calls[0].method).toBe('GET')
    expect(calls[0].headers['X-CSRF-Token']).toBeUndefined()
  })

  it('preserves ApiError details from a failed place order', async () => {
    const { fetcher } = stubFetcher([{ status: 409, body: { code: 'PRICE_CHANGED' } }])
    const client = createCheckoutClient({ baseURL: '/api', fetcher, readCookie: () => 'csrf-1' })

    const error = await client.placeOrder('key-12345', request).catch((value: unknown) => value)

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).code).toBe('PRICE_CHANGED')
  })
})
