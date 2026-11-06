import { describe, expect, it } from 'bun:test'
import { createFlashSaleClient } from './flash-sale-api'

type Call = { url: string; init: RequestInit }

function recorder(respond: () => Response) {
  const calls: Call[] = []
  const fetcher = async (input: RequestInfo | URL, init?: RequestInit) => {
    calls.push({ url: String(input), init: init ?? {} })
    return respond()
  }
  return { calls, fetcher }
}

describe('flashSaleClient', () => {
  it('lists public flash sales with query parameters', async () => {
    const payload = { items: [], next_cursor: null }
    const { calls, fetcher } = recorder(() => Response.json(payload))
    const client = createFlashSaleClient({ baseURL: 'http://api.test', fetcher })

    const res = await client.listPublicFlashSales({ status: 'active', limit: 10 })
    expect(res).toEqual(payload)
    expect(calls[0].url).toBe('http://api.test/api/v1/flash-sales?status=active&limit=10')
    expect(calls[0].init.method).toBe('GET')
  })

  it('creates flash sale draft with CSRF token', async () => {
    const campaign = { id: 'c1', name: 'Sale 1', status: 'draft', starts_at: '', ends_at: '', items: [] }
    const { calls, fetcher } = recorder(() => Response.json(campaign, { status: 201 }))
    const client = createFlashSaleClient({
      baseURL: 'http://api.test',
      fetcher,
      readCookie: () => 'my-csrf-token',
    })

    const data = {
      name: 'Sale 1',
      starts_at: '2026-11-03T10:00:00Z',
      ends_at: '2026-11-03T12:00:00Z',
      items: [{ sku_id: 's1', sale_price: 1000, quota_total: 5, purchase_limit: 1 }],
    }
    const res = await client.createFlashSale(data)

    expect(res.id).toBe('c1')
    expect(calls[0].url).toBe('http://api.test/api/v1/admin/flash-sales')
    expect(calls[0].init.method).toBe('POST')
    expect((calls[0].init.headers as Record<string, string>)['X-CSRF-Token']).toBe('my-csrf-token')
  })

  it('schedules and stops flash sale campaign', async () => {
    const campaign = { id: 'c1', name: 'Sale 1', status: 'scheduled', starts_at: '', ends_at: '', items: [] }
    const { calls, fetcher } = recorder(() => Response.json(campaign))
    const client = createFlashSaleClient({
      baseURL: 'http://api.test',
      fetcher,
      readCookie: () => 'my-csrf-token',
    })

    await client.scheduleFlashSale('c1')
    expect(calls[0].url).toBe('http://api.test/api/v1/admin/flash-sales/c1/schedule')
    expect(calls[0].init.method).toBe('POST')

    await client.stopFlashSale('c1')
    expect(calls[1].url).toBe('http://api.test/api/v1/admin/flash-sales/c1/stop')
    expect(calls[1].init.method).toBe('POST')
  })
})
