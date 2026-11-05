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

const csrfOptions = { baseURL: 'http://api.test', readCookie: () => 'cart-csrf' }

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
    const cart = {
      id: 'cart-1',
      items: [
        {
          id: 'i1',
          sku_id: 's1',
          quantity: 2,
          version: 1,
          price_source: 'regular',
          flash_sale_item_id: null,
          product_name: 'Áo',
          sku_code: 'AO-1',
          reference_unit_price: 100,
        },
      ],
    }
    const { calls, fetcher } = recorder(() => Response.json(cart))
    const client = createCartClient({ ...csrfOptions, fetcher })

    const res = await client.upsertCartItem('s1', { quantity: 2, price_source: 'regular' })
    expect(res.items.length).toBe(1)
    expect(calls[0].url).toBe('http://api.test/api/v1/cart/items/s1')
    expect(calls[0].init.method).toBe('PUT')
    expect((calls[0].init.headers as Record<string, string>)['X-CSRF-Token']).toBe('cart-csrf')
  })

  it('sends expected_version in the absolute-quantity PUT body', async () => {
    const { calls, fetcher } = recorder(() => Response.json({ id: 'c', items: [] }))
    const client = createCartClient({ ...csrfOptions, fetcher })
    await client.upsertCartItem('s1', { quantity: 3, expected_version: 7 })
    expect(JSON.parse(String(calls[0].init.body))).toEqual({ quantity: 3, expected_version: 7 })
  })

  it('removes cart item from a 204 No Content response', async () => {
    const { calls, fetcher } = recorder(() => new Response(null, { status: 204 }))
    const client = createCartClient({ ...csrfOptions, fetcher })

    await expect(client.removeCartItem('s1')).resolves.toBeUndefined()
    expect(calls[0].url).toBe('http://api.test/api/v1/cart/items/s1')
    expect(calls[0].init.method).toBe('DELETE')
    expect((calls[0].init.headers as Record<string, string>)['X-CSRF-Token']).toBe('cart-csrf')
  })

  it('surfaces a failed delete as ApiError', async () => {
    const { fetcher } = recorder(() => Response.json({ code: 'RESOURCE_NOT_FOUND' }, { status: 404 }))
    const client = createCartClient({ ...csrfOptions, fetcher })
    await expect(client.removeCartItem('s1')).rejects.toMatchObject({ status: 404, code: 'RESOURCE_NOT_FOUND' })
  })

  it('parses line_errors from a 409 preview rejection', async () => {
    const body = {
      code: 'SALE_QUOTA_EXCEEDED',
      line_errors: [{ sku_id: 's1', requested_quantity: 3, code: 'SALE_QUOTA_EXCEEDED', available_quantity: 1 }],
    }
    const { fetcher } = recorder(() => Response.json(body, { status: 409 }))
    const client = createCartClient({ ...csrfOptions, fetcher })
    const error = await client.previewCheckout({ items: [{ cart_item_id: 'i1', version: 1 }] }).catch((e) => e)
    expect(error.lineErrors).toEqual(body.line_errors)
  })

  it('calls checkout preview', async () => {
    const preview = {
      currency: 'VND',
      total_amount: 100000,
      lines: [],
      previewed_at: '2026-11-03T12:00:00Z',
    }
    const { calls, fetcher } = recorder(() => Response.json(preview))
    const client = createCartClient({ ...csrfOptions, fetcher })

    const res = await client.previewCheckout({ items: [{ cart_item_id: 'i1', version: 1 }] })
    expect(res.total_amount).toBe(100000)
    expect(calls[0].url).toBe('http://api.test/api/v1/checkout/preview')
    expect(calls[0].init.method).toBe('POST')
  })
})
