import { describe, expect, it } from 'bun:test'

import { ApiError } from '@/lib/api-error'
import type { CheckoutPreview } from '@/features/cart/cart.type'
import {
  buildPlaceOrderRequest,
  classifyPlaceOrderFailure,
  clearIntent,
  intentFingerprint,
  interpretAttempt,
  resolveIntent,
  validateRecipient,
} from '@/features/checkout/checkout.logic'
import { makeOrder } from '../bill/order-fixtures'

const preview: CheckoutPreview = {
  currency: 'VND',
  total_amount: 387000,
  previewed_at: '2026-11-25T08:00:00Z',
  lines: [
    {
      cart_item_id: 'cart-2',
      cart_item_version: 4,
      sku_id: 'sku-2',
      quantity: 1,
      unit_price: 129000,
      subtotal: 129000,
      price_source: 'regular',
    },
    {
      cart_item_id: 'cart-1',
      cart_item_version: 2,
      sku_id: 'sku-1',
      quantity: 2,
      unit_price: 129000,
      subtotal: 258000,
      price_source: 'flash_sale',
      flash_sale_item_id: 'sale-1',
    },
  ],
}

function storage(initial: string | null = null) {
  let value = initial
  return {
    getItem: (_key: string) => value,
    setItem: (_key: string, next: string) => {
      value = next
    },
    removeItem: (_key: string) => {
      value = null
    },
    read: () => value,
  }
}

describe('checkout logic', () => {
  it('builds the order request from the authoritative preview and trims recipient input', () => {
    expect(buildPlaceOrderRequest(preview, { name: '  Nguyễn Minh Anh ', phone: ' 0901234567 ', address: ' 12 Nguyễn Trãi ' })).toEqual({
      lines: preview.lines,
      currency: 'VND',
      expected_total: 387000,
      recipient: { name: 'Nguyễn Minh Anh', phone: '0901234567', address: '12 Nguyễn Trãi' },
    })
  })

  it('reuses an intent for the same request and replaces it when the fingerprint changes', () => {
    const state = storage()
    const firstRequest = buildPlaceOrderRequest(preview, { name: 'A', phone: '0901234567', address: 'A' })
    const changedRequest = buildPlaceOrderRequest(preview, { name: 'A', phone: '0901234568', address: 'A' })
    const first = resolveIntent(state, firstRequest, () => 'key-1')
    const same = resolveIntent(state, firstRequest, () => 'key-2')
    const changed = resolveIntent(state, changedRequest, () => 'key-3')

    expect(first.key).toBe('key-1')
    expect(same.key).toBe('key-1')
    expect(changed.key).toBe('key-3')
    expect(changed.fingerprint).toBe(intentFingerprint(changedRequest))
    expect(JSON.parse(state.read() ?? '{}')).toEqual(changed)
  })

  it('recovers from malformed storage and clears the persisted intent', () => {
    const state = storage('{not-json')
    const intent = resolveIntent(state, buildPlaceOrderRequest(preview, { name: 'A', phone: '0901234567', address: 'A' }), () => 'key-new')

    expect(intent.key).toBe('key-new')
    clearIntent(state)
    expect(state.read()).toBeNull()
  })

  it('validates recipient fields with the checkout copy', () => {
    expect(validateRecipient({ name: '', phone: '', address: '' })).toEqual({
      name: 'Vui lòng nhập họ tên.',
      phone: 'Vui lòng nhập số điện thoại.',
      address: 'Vui lòng nhập địa chỉ nhận hàng.',
    })
    expect(validateRecipient({ name: 'A'.repeat(121), phone: '1234567', address: 'A'.repeat(501) })).toEqual({
      name: 'Họ tên tối đa 120 ký tự.',
      phone: 'Số điện thoại cần 8–20 ký tự.',
      address: 'Địa chỉ tối đa 500 ký tự.',
    })
    expect(validateRecipient({ name: 'Nguyễn Minh Anh', phone: '0901234567', address: '12 Nguyễn Trãi' })).toEqual({})
  })

  it('classifies place-order outcomes by stable API contract', () => {
    expect(classifyPlaceOrderFailure(new ApiError({ status: 0, code: 'NETWORK_ERROR' }))).toEqual({ kind: 'unknown' })
    expect(classifyPlaceOrderFailure(new ApiError({ status: 503, code: 'INTERNAL_ERROR', requestID: 'req-1' }))).toEqual({ kind: 'unknown', requestID: 'req-1' })
    expect(classifyPlaceOrderFailure(new ApiError({ status: 409, code: 'IDEMPOTENCY_IN_PROGRESS' }))).toEqual({ kind: 'processing', retryAfterSeconds: 1 })
    expect(classifyPlaceOrderFailure(new ApiError({ status: 429, code: 'RATE_LIMITED', retryAfterSeconds: 4 }))).toEqual({ kind: 'overload', retryAfterSeconds: 4 })
    expect(classifyPlaceOrderFailure(new ApiError({ status: 409, code: 'PRICE_CHANGED', lineErrors: [] }))).toEqual({ kind: 'price_changed', lineErrors: [] })
    expect(classifyPlaceOrderFailure(new ApiError({ status: 409, code: 'CART_ITEM_CHANGED' }))).toEqual({ kind: 'cart_changed' })
    expect(classifyPlaceOrderFailure(new ApiError({ status: 409, code: 'INSUFFICIENT_STOCK', lineErrors: [] }))).toEqual({ kind: 'rejected', code: 'INSUFFICIENT_STOCK', lineErrors: [] })
    expect(classifyPlaceOrderFailure(new ApiError({ status: 409, code: 'IDEMPOTENCY_CONFLICT' }))).toEqual({ kind: 'intent_conflict' })
    expect(classifyPlaceOrderFailure(new ApiError({ status: 422, code: 'VALIDATION_ERROR', violations: [] }))).toEqual({ kind: 'invalid', violations: [] })
    expect(classifyPlaceOrderFailure(new ApiError({ status: 401, code: 'UNAUTHENTICATED' }))).toEqual({ kind: 'unauthenticated' })
  })

  it('interprets stored attempt responses as created or rejected', () => {
    const order = makeOrder()
    expect(interpretAttempt({ idempotency_key: 'key', response_status: 201, response_body: order, completed_at: 'now' })).toEqual({
      kind: 'created',
      order,
    })
    expect(interpretAttempt({ idempotency_key: 'key', response_status: 409, response_body: { code: 'SALE_QUOTA_EXCEEDED' }, completed_at: 'now' })).toEqual({
      kind: 'rejected',
      code: 'SALE_QUOTA_EXCEEDED',
      lineErrors: [],
    })
  })
})
