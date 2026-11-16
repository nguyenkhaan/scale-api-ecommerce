import axios, { type AxiosResponse } from 'axios'

import { privateClient, publicClient } from '@/api/client'

import type {
  AdminProductListParams,
  CatalogApiErrorInit,
  FieldViolation,
  ProductListParams,
  ProductPage,
} from './catalog.type'

type ProblemResponse = {
  code?: unknown
  request_id?: unknown
  violations?: unknown
}

export class CatalogApiError extends Error {
  readonly status: number
  readonly code: string
  readonly requestID?: string
  readonly violations: FieldViolation[]
  readonly retryAfterSeconds?: number

  constructor(init: CatalogApiErrorInit) {
    super(`api error ${init.code} (${init.status})`)
    this.name = 'CatalogApiError'
    this.status = init.status
    this.code = init.code
    this.requestID = init.requestID
    this.violations = init.violations ?? []
    this.retryAfterSeconds = init.retryAfterSeconds
  }
}

export function isCatalogApiError(error: unknown): error is CatalogApiError {
  return error instanceof CatalogApiError
}

export async function listProducts(params: ProductListParams): Promise<ProductPage> {
  return request(publicClient.get<ProductPage>('/products', { params }))
}

export async function listAdminProducts(params: AdminProductListParams): Promise<ProductPage> {
  return request(privateClient.get<ProductPage>('/admin/products', { params }))
}

function request<T>(operation: Promise<AxiosResponse<T>>): Promise<T> {
  return operation.then(
    (response) => response.data,
    (error: unknown) => Promise.reject(toCatalogApiError(error)),
  )
}

function toCatalogApiError(error: unknown): CatalogApiError {
  if (isCatalogApiError(error)) return error
  if (!axios.isAxiosError<ProblemResponse>(error) || error.response === undefined) {
    return new CatalogApiError({ status: 0, code: 'NETWORK_ERROR' })
  }

  const { data, headers, status } = error.response
  return new CatalogApiError({
    status,
    code: typeof data?.code === 'string' ? data.code : 'UNKNOWN_ERROR',
    requestID: typeof data?.request_id === 'string' ? data.request_id : headerValue(headers['x-request-id']),
    violations: parseViolations(data?.violations),
    retryAfterSeconds: parseRetryAfter(headers['retry-after']),
  })
}

function parseViolations(value: unknown): FieldViolation[] {
  if (!Array.isArray(value)) return []
  return value.flatMap((violation): FieldViolation[] => {
    if (!isRecord(violation) || typeof violation.field !== 'string' || typeof violation.reason !== 'string') return []
    return [{ field: violation.field, reason: violation.reason }]
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
