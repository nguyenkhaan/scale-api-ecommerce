import { afterEach, describe, expect, it } from 'bun:test'
import type { AxiosAdapter, AxiosResponse } from 'axios'

import { privateClient } from '@/api/client'
import { getCart, previewCheckout, removeCartItem, upsertCartItem } from '@/features/cart/cart.service'

type Call = { url?: string; method?: string; data?: unknown; headers?: unknown }

const initialAdapter = privateClient.defaults.adapter
const initialDocument = Object.getOwnPropertyDescriptor(globalThis, 'document')

afterEach(() => {
  privateClient.defaults.adapter = initialAdapter
  if (initialDocument) Object.defineProperty(globalThis, 'document', initialDocument)
  else delete (globalThis as { document?: unknown }).document
})

function respond(data: unknown): Call[] {
  const calls: Call[] = []
  const adapter: AxiosAdapter = async (config) => {
    calls.push({ url: config.url, method: config.method, data: config.data, headers: config.headers })
    return {
      data,
      status: 200,
      statusText: 'OK',
      headers: {},
      config,
    } as AxiosResponse
  }
  privateClient.defaults.adapter = adapter
  return calls
}

function withCsrfCookie() {
  Object.defineProperty(globalThis, 'document', {
    configurable: true,
    value: { cookie: 'salescale_csrf=cart-csrf' },
  })
}

describe('cart service', () => {
  it('gets the current cart from the OpenAPI route', async () => {
    const calls = respond({ id: 'cart-1', items: [] })

    await expect(getCart()).resolves.toEqual({ id: 'cart-1', items: [] })
    expect(calls).toHaveLength(1)
    expect(calls[0].url).toBe('/cart')
    expect(calls[0].method).toBe('get')
  })

  it('sends an absolute cart quantity with CSRF proof', async () => {
    withCsrfCookie()
    const calls = respond({ id: 'cart-1', items: [] })

    await upsertCartItem('sku-1', { quantity: 3, price_source: 'regular', expected_version: 7 })

    expect(calls[0].url).toBe('/cart/items/sku-1')
    expect(calls[0].method).toBe('put')
    expect(JSON.parse(String(calls[0].data))).toEqual({ quantity: 3, price_source: 'regular', expected_version: 7 })
    expect((calls[0].headers as Record<string, string>)['X-CSRF-Token']).toBe('cart-csrf')
  })

  it('uses the mutation routes for deleting a line and previewing selected lines', async () => {
    withCsrfCookie()
    const calls = respond({ currency: 'VND', total_amount: 0, lines: [], previewed_at: '2026-11-15T00:00:00Z' })

    await removeCartItem('sku-1')
    await previewCheckout({ items: [{ cart_item_id: 'line-1', version: 1 }] })

    expect(calls.map((call) => [call.url, call.method])).toEqual([
      ['/cart/items/sku-1', 'delete'],
      ['/checkout/preview', 'post'],
    ])
    expect((calls[1].headers as Record<string, string>)['X-CSRF-Token']).toBe('cart-csrf')
  })
})
