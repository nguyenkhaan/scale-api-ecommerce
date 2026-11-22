import { describe, expect, it } from 'bun:test'
import { QueryClient, QueryObserver } from '@tanstack/react-query'

import { ApiError } from '@/lib/api-error'
import { createOrderRefetchInterval } from '@/features/bill/bill.hook'
import type { Order } from '@/features/bill/bill.type'

import { makeOrder } from './order-fixtures'

/** Drives a real QueryObserver through failing/succeeding fetches and records the interval after each. */
function harness(results: Array<Order | ApiError>) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false, gcTime: Infinity } } })
  const interval = createOrderRefetchInterval()
  let call = 0
  const observer = new QueryObserver<Order>(client, {
    queryKey: ['orders', 'poll-test'],
    queryFn: async () => {
      const next = results[Math.min(call++, results.length - 1)]
      if (next instanceof ApiError) throw next
      return next
    },
  })
  const step = async () => {
    await observer.refetch()
    return interval(observer.getCurrentQuery())
  }
  return { step, client }
}

const networkError = () => new ApiError({ status: 0, code: 'NETWORK_ERROR' })

describe('order polling backoff through a real query', () => {
  it('doubles the delay on consecutive errors up to the 30 s cap', async () => {
    const { step } = harness([networkError()])
    const delays = []
    for (let i = 0; i < 6; i++) delays.push(await step())
    expect(delays).toEqual([4000, 8000, 16000, 30000, 30000, 30000])
  })

  it('resets to 2 s after a success and stops on a terminal status', async () => {
    const { step } = harness([networkError(), networkError(), makeOrder(), networkError(), makeOrder({ status: 'confirmed', payable: false })])
    expect(await step()).toBe(4000)
    expect(await step()).toBe(8000)
    expect(await step()).toBe(2000)
    expect(await step()).toBe(4000)
    expect(await step()).toBe(false)
  })

  it('stops polling on a fatal 404', async () => {
    const { step } = harness([new ApiError({ status: 404, code: 'NOT_FOUND' })])
    expect(await step()).toBe(false)
  })
})
