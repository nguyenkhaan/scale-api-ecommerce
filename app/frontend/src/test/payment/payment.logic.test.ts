import { describe, expect, it } from 'bun:test'

import { ApiError } from '@/lib/api-error'
import { createPaymentClient } from '@/features/payment/payment.service'
import { describeSessionFailure, selectPaymentView, shouldRequestSession } from '@/features/payment/payment.logic'

import { makeOrder, stubFetcher } from '../bill/order-fixtures'

describe('payment terminal view selection', () => {
  it('derives every view from the official order status', () => {
    expect(selectPaymentView(undefined)).toBe('loading')
    expect(selectPaymentView(makeOrder({ status: 'confirmed', payable: false }))).toBe('confirmed')
    expect(selectPaymentView(makeOrder({ status: 'payment_failed', payable: false }))).toBe('failed')
    expect(selectPaymentView(makeOrder({ status: 'cancelled', payable: false }))).toBe('cancelled')
    expect(selectPaymentView(makeOrder({ status: 'expired', payable: false }))).toBe('expired')
  })

  it('keeps pending with session pending or unknown neutral, never failed', () => {
    expect(selectPaymentView(makeOrder({ payment_status: 'pending' }))).toBe('waiting')
    expect(selectPaymentView(makeOrder({ payment_status: null }))).toBe('waiting')
    expect(selectPaymentView(makeOrder({ payment_status: 'unknown' }))).toBe('checking')
  })

  it('does not trust a session-level failure while the order is still pending', () => {
    expect(selectPaymentView(makeOrder({ payment_status: 'failed' }))).toBe('waiting')
  })

  it('treats pending but not payable as overdue', () => {
    expect(selectPaymentView(makeOrder({ payable: false }))).toBe('overdue')
  })
})

describe('payment session request rules', () => {
  it('requests once, and only for payable orders', () => {
    expect(shouldRequestSession(makeOrder(), false)).toBe(true)
    expect(shouldRequestSession(makeOrder(), true)).toBe(false)
    expect(shouldRequestSession(undefined, false)).toBe(false)
    expect(shouldRequestSession(makeOrder({ payable: false }), false)).toBe(false)
    expect(shouldRequestSession(makeOrder({ status: 'confirmed', payable: false }), false)).toBe(false)
  })

  it('maps ORDER_NOT_PAYABLE to a refetch-worthy failure without exposing the code', () => {
    const failure = describeSessionFailure(new ApiError({ status: 409, code: 'ORDER_NOT_PAYABLE' }))
    expect(failure.kind).toBe('not-payable')
    expect(failure.text).not.toContain('ORDER_NOT_PAYABLE')
    expect(describeSessionFailure(new ApiError({ status: 0, code: 'NETWORK_ERROR' })).kind).toBe('error')
  })
})

describe('payment service', () => {
  const session = { id: 's1', order_id: 'order-1', status: 'pending', amount: 848000, currency: 'VND', simulator_reference: 'PAY-7F3K9Q' }
  const clientWith = (responses: Array<{ status: number; body?: unknown }>) => {
    const { fetcher, calls } = stubFetcher(responses)
    return { client: createPaymentClient({ baseURL: '/api', fetcher, readCookie: () => 'csrf-9' }), calls }
  }

  it('creates a session with CSRF on POST and accepts both 201 and 200', async () => {
    const { client, calls } = clientWith([{ status: 201, body: session }, { status: 200, body: session }])
    await expect(client.createPaymentSession('order-1')).resolves.toEqual(session as never)
    await expect(client.createPaymentSession('order-1')).resolves.toEqual(session as never)
    expect(calls[0].url).toBe('/api/orders/order-1/payment-session')
    expect(calls[0].method).toBe('POST')
    expect(calls[0].headers['X-CSRF-Token']).toBe('csrf-9')
  })

  it('rejects a 409 ORDER_NOT_PAYABLE with its stable code', async () => {
    const { client } = clientWith([{ status: 409, body: { code: 'ORDER_NOT_PAYABLE' } }])
    const error = await client.createPaymentSession('order-1').catch((err: unknown) => err)
    expect((error as ApiError).status).toBe(409)
    expect((error as ApiError).code).toBe('ORDER_NOT_PAYABLE')
  })

  it('lists pending simulator sessions with the status and limit query', async () => {
    const { client, calls } = clientWith([{ status: 200, body: { items: [] } }])
    await client.listSimulatorSessions(10)
    expect(calls[0].url).toBe('/api/payment-simulator/sessions?status=pending&limit=10')
    expect(calls[0].method).toBe('GET')
  })

  it('posts simulator events with CSRF and returns the 202 receipt', async () => {
    const receipt = { event_id: 'e1', scheduled_at: '2026-10-10T13:43:05Z' }
    const { client, calls } = clientWith([{ status: 202, body: receipt }])
    const body = { payment_session_id: 's1', event_type: 'succeeded' as const, delay_ms: 0, repeat_count: 1, drop_delivery: false, signature_mode: 'valid' as const }
    await expect(client.scheduleSimulatorEvent(body)).resolves.toEqual(receipt)
    expect(calls[0].url).toBe('/api/payment-simulator/events')
    expect(calls[0].headers['X-CSRF-Token']).toBe('csrf-9')
    expect(calls[0].body).toEqual(body)
  })

  it('surfaces 404 (simulator disabled) and 403 (non-admin)', async () => {
    const { client } = clientWith([{ status: 404, body: { code: 'NOT_FOUND' } }, { status: 403, body: { code: 'FORBIDDEN' } }])
    expect(((await client.listSimulatorSessions().catch((err: unknown) => err)) as ApiError).status).toBe(404)
    expect(((await client.listSimulatorSessions().catch((err: unknown) => err)) as ApiError).status).toBe(403)
  })
})
