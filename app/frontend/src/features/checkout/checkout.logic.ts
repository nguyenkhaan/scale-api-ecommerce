import { isApiError } from '@/lib/api-error'
import type { CheckoutPreview } from '@/features/cart/cart.type'

import type {
  CheckoutIntent,
  CheckoutOutcome,
  IdempotencyResult,
  Order,
  PlaceOrderRequest,
  RecipientErrors,
  RecipientInput,
} from './checkout.type'

export const CHECKOUT_INTENT_STORAGE_KEY = 'salescale.checkout-intent'

export function buildPlaceOrderRequest(preview: CheckoutPreview, recipient: RecipientInput): PlaceOrderRequest {
  return {
    lines: preview.lines,
    currency: 'VND',
    expected_total: preview.total_amount,
    recipient: {
      name: recipient.name.trim(),
      phone: recipient.phone.trim(),
      address: recipient.address.trim(),
    },
  }
}

export function intentFingerprint(request: PlaceOrderRequest): string {
  return JSON.stringify({
    ...request,
    lines: [...request.lines].sort((left, right) => left.cart_item_id.localeCompare(right.cart_item_id)),
  })
}

export function resolveIntent(
  storage: Pick<Storage, 'getItem' | 'setItem'>,
  request: PlaceOrderRequest,
  newKey: () => string = () => globalThis.crypto.randomUUID(),
): CheckoutIntent {
  const fingerprint = intentFingerprint(request)
  const stored = readIntent(storage.getItem(CHECKOUT_INTENT_STORAGE_KEY))
  if (stored?.fingerprint === fingerprint) return stored

  const intent = { key: newKey(), fingerprint }
  storage.setItem(CHECKOUT_INTENT_STORAGE_KEY, JSON.stringify(intent))
  return intent
}

export function clearIntent(storage: Pick<Storage, 'removeItem'>): void {
  storage.removeItem(CHECKOUT_INTENT_STORAGE_KEY)
}

export function validateRecipient(input: RecipientInput): RecipientErrors {
  const errors: RecipientErrors = {}
  const name = input.name.trim()
  const phone = input.phone.trim()
  const address = input.address.trim()

  if (name.length === 0) errors.name = 'Vui lòng nhập họ tên.'
  else if (name.length > 120) errors.name = 'Họ tên tối đa 120 ký tự.'
  if (phone.length === 0) errors.phone = 'Vui lòng nhập số điện thoại.'
  else if (phone.length < 8 || phone.length > 20) errors.phone = 'Số điện thoại cần 8–20 ký tự.'
  if (address.length === 0) errors.address = 'Vui lòng nhập địa chỉ nhận hàng.'
  else if (address.length > 500) errors.address = 'Địa chỉ tối đa 500 ký tự.'

  return errors
}

export function classifyPlaceOrderFailure(error: unknown): CheckoutOutcome {
  if (!isApiError(error)) return { kind: 'unknown' }
  if (error.status === 0 || error.code === 'INVALID_RESPONSE' || [500, 502, 503, 504].includes(error.status)) {
    return { kind: 'unknown', requestID: error.requestID }
  }
  if (error.status === 401) return { kind: 'unauthenticated' }
  if (error.status === 422) return { kind: 'invalid', violations: error.violations }
  if (error.status === 429) return { kind: 'overload', retryAfterSeconds: error.retryAfterSeconds ?? 1 }
  if (error.status !== 409) return { kind: 'unknown', requestID: error.requestID }

  if (error.code === 'IDEMPOTENCY_IN_PROGRESS') {
    return { kind: 'processing', retryAfterSeconds: error.retryAfterSeconds ?? 1 }
  }
  if (error.code === 'PRICE_CHANGED') return { kind: 'price_changed', lineErrors: error.lineErrors }
  if (error.code === 'CART_ITEM_CHANGED') return { kind: 'cart_changed' }
  if (error.code === 'IDEMPOTENCY_CONFLICT') return { kind: 'intent_conflict' }

  return { kind: 'rejected', code: error.code, lineErrors: error.lineErrors }
}

export function interpretAttempt(result: IdempotencyResult): CheckoutOutcome {
  if (result.response_status === 201) return { kind: 'created', order: result.response_body as Order }
  if (result.response_status === 409) {
    const body = result.response_body
    const code = typeof body.code === 'string' ? body.code : 'UNKNOWN_ERROR'
    return {
      kind: 'rejected',
      code,
      lineErrors: Array.isArray(body.line_errors) ? body.line_errors.filter(isLineError) : [],
    }
  }
  return { kind: 'unknown' }
}

function readIntent(value: string | null): CheckoutIntent | undefined {
  if (value === null) return undefined
  try {
    const parsed = JSON.parse(value) as unknown
    if (!isRecord(parsed) || typeof parsed.key !== 'string' || typeof parsed.fingerprint !== 'string') return undefined
    return { key: parsed.key, fingerprint: parsed.fingerprint }
  } catch {
    return undefined
  }
}

function isLineError(value: unknown): value is { sku_id: string; requested_quantity: number; code: string } {
  return isRecord(value) && typeof value.sku_id === 'string' && typeof value.requested_quantity === 'number' && typeof value.code === 'string'
}

function isRecord(value: unknown): value is Record<string, any> {
  return typeof value === 'object' && value !== null
}
