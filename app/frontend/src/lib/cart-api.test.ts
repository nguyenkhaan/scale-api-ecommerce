import { describe, expect, it } from 'bun:test'
import { createCartClient } from './cart-api'

type Call = { url: string; init: RequestInit }

function recorder(respond: () => Response) {
  const calls: Call[] = []
  const fetcher = async (input: RequestInfo | URL, init?: RequestInit) => {
    calls.push({ url: String(input), init: init ?? {} })
    return respond()
  }
  return { calls, fetcher }
}

describe('cartClient', () => {
  it('gets cart', async () => {
    const cart = { id: 'cart-1', items: [] }
    const { calls, fetcher } = recorder(() => Response.json(cart))
    const client = createCartClient({ baseURL: 'http://api.test', fetcher })

    const res = await client.getCart()
    expect(res).toEqual(cart)
    expect(calls[0].url).toBe('http://api.test/api/v1/cart')
    expect(calls[0].init.method).toBe('GET')
  })

  it('upserts cart item with CSRF token', async () => {
    const cart = { id: 'cart-1', items: [{ id: 'i1', sku_id: 's1', quantity: 2, version: 1, price_source: 'regular', flash_sale_item_id: null }] }
    const { calls, fetcher } = recorder(() => Response.json(cart))
    const client = createCartClient({
      baseURL: 'http://api.test',
      fetcher,
      readCookie: () => 'cart-csrf',
    })

    const res = await client.upsertCartItem('s1', { quantity: 2, price_source: 'regular' })
    expect(res.items.length).toBe(1)
    expect(calls[0].url).toBe('http://api.test/api/v1/cart/items/s1')
    expect(calls[0].init.method).toBe('PUT')
    expect((calls[0].init.headers as Record<string, string>)['X-CSRF-Token']).toBe('cart-csrf')
  })

  it('removes cart item', async () => {
    const cart = { id: 'cart-1', items: [] }
    const { calls, fetcher } = recorder(() => Response.json(cart))
    const client = createCartClient({
      baseURL: 'http://api.test',
      fetcher,
      readCookie: () => 'cart-csrf',
    })

    await client.removeCartItem('s1')
    expect(calls[0].url).toBe('http://api.test/api/v1/cart/items/s1')
    expect(calls[0].init.method).toBe('DELETE')
  })

  it('calls checkout preview', async () => {
    const preview = {
      currency: 'VND',
      total_amount: 100000,
      lines: [],
      previewed_at: '2026-11-03T12:00:00Z',
    }
    const { calls, fetcher } = recorder(() => Response.json(preview))
    const client = createCartClient({
      baseURL: 'http://api.test',
      fetcher,
      readCookie: () => 'cart-csrf',
    })

    const res = await client.previewCheckout({ items: [{ cart_item_id: 'i1', version: 1 }] })
    expect(res.total_amount).toBe(100000)
    expect(calls[0].url).toBe('http://api.test/api/v1/checkout/preview')
    expect(calls[0].init.method).toBe('POST')
  })
})
