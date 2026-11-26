import axios, { type AxiosResponse } from 'axios'

import { privateClient, publicClient } from '@/api/client'
import { csrfCookieName, readDocumentCookie } from '@/lib/csrf'

import type {
  AdminProductListParams,
  CatalogApiErrorInit,
  FieldViolation,
  CategoryPage,
  CreateProductRequest,
  CreateSkuInput,
  ProductListParams,
  CatalogProduct,
  ProductPage,
  UpdateProductRequest,
  UpdateSkuRequest,
  CatalogSku,
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

export async function listCategories(): Promise<CategoryPage> {
  return request(publicClient.get<CategoryPage>('/categories'))
}

export async function getProduct(productID: string): Promise<CatalogProduct> {
  return request(publicClient.get<CatalogProduct>(`/products/${encodeURIComponent(productID)}`))
}

export async function listAdminProducts(params: AdminProductListParams): Promise<ProductPage> {
  return request(privateClient.get<ProductPage>('/admin/products', { params }))
}

export async function getAdminProduct(productID: string): Promise<CatalogProduct> {
  return request(privateClient.get<CatalogProduct>(`/admin/products/${encodeURIComponent(productID)}`))
}

export async function createProduct(body: CreateProductRequest): Promise<CatalogProduct> {
  return request(privateClient.post<CatalogProduct>('/admin/products', body, { headers: csrfHeaders() }))
}

export async function updateProduct(productID: string, body: UpdateProductRequest): Promise<CatalogProduct> {
  return request(privateClient.patch<CatalogProduct>(`/admin/products/${encodeURIComponent(productID)}`, body, { headers: csrfHeaders() }))
}

export async function createSKU(productID: string, body: CreateSkuInput): Promise<CatalogSku> {
  return request(privateClient.post<CatalogSku>(`/admin/products/${encodeURIComponent(productID)}/skus`, body, { headers: csrfHeaders() }))
}

export async function updateSKU(skuID: string, body: UpdateSkuRequest): Promise<CatalogSku> {
  return request(privateClient.patch<CatalogSku>(`/admin/skus/${encodeURIComponent(skuID)}`, body, { headers: csrfHeaders() }))
}

function csrfHeaders() {
  return { 'X-CSRF-Token': readDocumentCookie(csrfCookieName) ?? '' }
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
