import { describe, expect, it } from 'bun:test'

import { ApiError } from '@/lib/api-error'
import { buildOrderListQuery, createBillClient, createOrderPageFetcher } from '@/features/bill/bill.service'

import { makeOrder, stubFetcher } from './order-fixtures'

const clientWith = (responses: Array<{ status: number; body?: unknown }>) => {
  const { fetcher, calls } = stubFetcher(responses)
  return { client: createBillClient({ baseURL: '/api', fetcher, readCookie: () => 'csrf-1' }), calls }
}

describe('bill service', () => {
  it('lists orders without CSRF and with the status/cursor/limit query', async () => {
    const { client, calls } = clientWith([{ status: 200, body: { items: [], next_cursor: null } }])
    await client.listOrders({ status: 'pending_payment', cursor: 'c1', limit: 20 })
    expect(calls[0].url).toBe('/api/orders?status=pending_payment&cursor=c1&limit=20')
    expect(calls[0].method).toBe('GET')
    expect(calls[0].headers['X-CSRF-Token']).toBeUndefined()
  })

  it('omits the query string when there are no filters', () => {
    expect(buildOrderListQuery()).toBe('')
    expect(buildOrderListQuery({})).toBe('')
  })

  it('gets one order by id', async () => {
    const order = makeOrder()
    const { client, calls } = clientWith([{ status: 200, body: order }])
    await expect(client.getOrder('order-1')).resolves.toEqual(order)
    expect(calls[0].url).toBe('/api/orders/order-1')
  })

  it('cancels with the CSRF header on POST', async () => {
    const { client, calls } = clientWith([{ status: 200, body: makeOrder({ status: 'cancelled', payable: false }) }])
    const order = await client.cancelOrder('order-1')
    expect(order.status).toBe('cancelled')
    expect(calls[0].url).toBe('/api/orders/order-1/cancel')
    expect(calls[0].method).toBe('POST')
    expect(calls[0].headers['X-CSRF-Token']).toBe('csrf-1')
  })

  it('rejects a 409 cancel with the stable error code', async () => {
    const { client } = clientWith([{ status: 409, body: { code: 'STATE_TRANSITION_CONFLICT' } }])
    const error = await client.cancelOrder('order-1').catch((err: unknown) => err)
    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).status).toBe(409)
    expect((error as ApiError).code).toBe('STATE_TRANSITION_CONFLICT')
  })

  it('surfaces a 404 for unknown orders', async () => {
    const { client } = clientWith([{ status: 404, body: { code: 'ORDER_NOT_FOUND' } }])
    const error = await client.getOrder('missing').catch((err: unknown) => err)
    expect((error as ApiError).status).toBe(404)
  })
})

describe('order list cursor paging', () => {
  it('requests the first page without a cursor, then the next page with next_cursor', async () => {
    const first = [makeOrder({ id: 'o1' })]
    const second = [makeOrder({ id: 'o2' })]
    const { client, calls } = clientWith([
      { status: 200, body: { items: first, next_cursor: 'cursor-2' } },
      { status: 200, body: { items: second, next_cursor: null } },
    ])
    const fetchPage = createOrderPageFetcher(client, 'confirmed')

    const page1 = await fetchPage()
    const page2 = await fetchPage(page1.next_cursor ?? undefined)

    expect(calls[0].url).toBe('/api/orders?status=confirmed&limit=20')
    expect(calls[1].url).toBe('/api/orders?status=confirmed&cursor=cursor-2&limit=20')
    expect(page2.next_cursor).toBeNull()
    expect([...page1.items, ...page2.items].map((order) => order.id)).toEqual(['o1', 'o2'])
  })

  it('does not send a status filter for the all tab', async () => {
    const { client, calls } = clientWith([{ status: 200, body: { items: [], next_cursor: null } }])
    await createOrderPageFetcher(client)()
    expect(calls[0].url).toBe('/api/orders?limit=20')
  })
})
