import { createJsonHttp, type JsonHttpOptions } from './json-http'

export type CatalogSku = {
  id: string
  code: string
  attributes: Record<string, unknown>
  regular_price: number
  active: boolean
}

export type CatalogProduct = {
  id: string
  name: string
  active: boolean
  image_url: string | null
  skus: CatalogSku[]
}

export type ProductPage = {
  items: CatalogProduct[]
  next_cursor: string | null
}

export type ProductListParams = { limit?: number; cursor?: string }

function query(params?: ProductListParams): string {
  const q = new URLSearchParams()
  if (params?.limit) q.set('limit', String(params.limit))
  if (params?.cursor) q.set('cursor', params.cursor)
  const str = q.toString()
  return str ? `?${str}` : ''
}

export function createCatalogClient(options: JsonHttpOptions) {
  const http = createJsonHttp(options)
  return {
    listProducts: (params?: ProductListParams) =>
      http.json<ProductPage>(`/api/v1/products${query(params)}`, { method: 'GET' }),
    listAdminProducts: (params?: ProductListParams) =>
      http.json<ProductPage>(`/api/v1/admin/products${query(params)}`, { method: 'GET' }),
  }
}

export const catalogClient = createCatalogClient({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
})
