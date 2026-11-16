import axios, { type AxiosResponse } from 'axios'

import { privateClient } from '@/api/client'

import type {
  Cart,
  CartApiErrorInit,
  CheckoutPreview,
  CheckoutSelection,
  FieldViolation,
  LineError,
  UpsertCartItemRequest,
} from './cart.type'

const csrfCookieName = 'salescale_csrf'
const csrfHeaderName = 'X-CSRF-Token'

type ProblemResponse = {
  code?: unknown
  request_id?: unknown
  violations?: unknown
  line_errors?: unknown
}

export class CartApiError extends Error {
  readonly status: number
  readonly code: string
  readonly requestID?: string
  readonly violations: FieldViolation[]
  readonly lineErrors: LineError[]
  readonly retryAfterSeconds?: number

  constructor(init: CartApiErrorInit) {
    super(`api error ${init.code} (${init.status})`)
    this.name = 'CartApiError'
    this.status = init.status
    this.code = init.code
    this.requestID = init.requestID
    this.violations = init.violations ?? []
    this.lineErrors = init.lineErrors ?? []
    this.retryAfterSeconds = init.retryAfterSeconds
  }
}

export function isCartApiError(error: unknown): error is CartApiError {
  return error instanceof CartApiError
}

export async function getCart(): Promise<Cart> {
  return request(privateClient.get<Cart>('/cart'))
}

export async function upsertCartItem(skuId: string, data: UpsertCartItemRequest): Promise<Cart> {
  return request(
    privateClient.put<Cart>(`/cart/items/${skuId}`, data, {
      headers: { [csrfHeaderName]: readDocumentCookie(csrfCookieName) ?? '' },
    }),
  )
}

export async function removeCartItem(skuId: string): Promise<void> {
  await request(
    privateClient.delete(`/cart/items/${skuId}`, {
      headers: { [csrfHeaderName]: readDocumentCookie(csrfCookieName) ?? '' },
    }),
  )
}

export async function previewCheckout(selection: CheckoutSelection): Promise<CheckoutPreview> {
  return request(
    privateClient.post<CheckoutPreview>('/checkout/preview', selection, {
      headers: { [csrfHeaderName]: readDocumentCookie(csrfCookieName) ?? '' },
    }),
  )
}

function request<T>(operation: Promise<AxiosResponse<T>>): Promise<T> {
  return operation.then(
    (response) => response.data,
    (error: unknown) => Promise.reject(toCartApiError(error)),
  )
}

function toCartApiError(error: unknown): CartApiError {
  if (isCartApiError(error)) return error
  if (!axios.isAxiosError<ProblemResponse>(error) || error.response === undefined) {
    return new CartApiError({ status: 0, code: 'NETWORK_ERROR' })
  }

  const { data, headers, status } = error.response
  return new CartApiError({
    status,
    code: typeof data?.code === 'string' ? data.code : 'UNKNOWN_ERROR',
    requestID: typeof data?.request_id === 'string' ? data.request_id : headerValue(headers['x-request-id']),
    violations: parseViolations(data?.violations),
    lineErrors: parseLineErrors(data?.line_errors),
    retryAfterSeconds: parseRetryAfter(headers['retry-after']),
  })
}

function readDocumentCookie(name: string): string | null {
  if (typeof document === 'undefined') return null

  for (const cookie of document.cookie.split(';')) {
    const separator = cookie.indexOf('=')
    if (separator <= 0 || cookie.slice(0, separator).trim() !== name) continue
    try {
      return decodeURIComponent(cookie.slice(separator + 1).trim())
    } catch {
      return null
    }
  }
  return null
}

function parseViolations(value: unknown): FieldViolation[] {
  if (!Array.isArray(value)) return []
  return value.flatMap((violation): FieldViolation[] => {
    if (!isRecord(violation) || typeof violation.field !== 'string' || typeof violation.reason !== 'string') return []
    return [{ field: violation.field, reason: violation.reason }]
  })
}

function parseLineErrors(value: unknown): LineError[] {
  if (!Array.isArray(value)) return []
  return value.flatMap((line): LineError[] => {
    if (!isRecord(line) || typeof line.sku_id !== 'string' || typeof line.requested_quantity !== 'number' || typeof line.code !== 'string') {
      return []
    }
    const parsed: LineError = {
      sku_id: line.sku_id,
      requested_quantity: line.requested_quantity,
      code: line.code as LineError['code'],
    }
    if (typeof line.available_quantity === 'number') parsed.available_quantity = line.available_quantity
    if (typeof line.current_unit_price === 'number') parsed.current_unit_price = line.current_unit_price
    return [parsed]
  })
}

function parseRetryAfter(value: unknown): number | undefined {
  const raw = headerValue(value)
  return raw !== undefined && /^\d+$/.test(raw) ? Number(raw) : undefined
}

function headerValue(value: unknown): string | undefined {
  if (typeof value === 'string') return value
  if (Array.isArray(value) && typeof value[0] === 'string') return value[0]
  return undefined
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
