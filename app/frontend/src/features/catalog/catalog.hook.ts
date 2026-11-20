import { useInfiniteQuery } from '@tanstack/react-query'

import { formatVND } from '@/lib/format'

import { isCatalogApiError, listProducts } from './catalog.service'
import type { CatalogProduct, CatalogSku, ProductListParams } from './catalog.type'

const pageSize = 20

export const catalogQueryKeys = {
  products: (limit: number, query: string) => ['catalog', 'products', { limit, query }] as const,
}

export function productListParams(query: string, cursor?: string): ProductListParams {
  const params: ProductListParams = { limit: pageSize }
  const normalizedQuery = query.trim()
  if (normalizedQuery !== '') params.query = normalizedQuery
  if (cursor !== undefined) params.cursor = cursor
  return params
}

export function useProductsQuery(query: string) {
  return useInfiniteQuery({
    queryKey: catalogQueryKeys.products(pageSize, query),
    initialPageParam: undefined as string | undefined,
    queryFn: ({ pageParam }) => listProducts(productListParams(query, pageParam)),
    getNextPageParam: (page) => page.next_cursor ?? undefined,
  })
}

export function activeSkus(product: CatalogProduct): CatalogSku[] {
  return product.active ? product.skus.filter((sku) => sku.active) : []
}

export type SkuOption = {
  skuId: string
  regularPrice: number
  label: string
}

export function buildSkuOptions(products: CatalogProduct[]): SkuOption[] {
  return products.flatMap((product) =>
    activeSkus(product).map((sku) => ({
      skuId: sku.id,
      regularPrice: sku.regular_price,
      label: `${product.name} — ${sku.code} — ${formatVND(sku.regular_price)}`,
    })),
  )
}

type CampaignSku = {
  sku_id: string
  sku_code: string
  product_name: string
  regular_price: number
}

export function mergeCampaignSkuOptions(options: SkuOption[], items: CampaignSku[]): SkuOption[] {
  const known = new Set(options.map((option) => option.skuId))
  const extra = items
    .filter((item) => !known.has(item.sku_id))
    .map((item) => ({
      skuId: item.sku_id,
      regularPrice: item.regular_price,
      label: `${item.product_name} — ${item.sku_code} — ${formatVND(item.regular_price)}`,
    }))
  return [...options, ...extra]
}

export function describeProductsLoadError(error: unknown): string {
  if (isCatalogApiError(error)) {
    if (error.code === 'NETWORK_ERROR') return 'Lỗi kết nối khi tải danh sách sản phẩm.'
    if (error.status === 403) return 'Bạn không có quyền xem danh sách sản phẩm.'
    if (error.status === 429) return 'Bạn thao tác quá nhanh khi tải danh sách sản phẩm. Vui lòng thử lại sau.'
  }
  return 'Không thể tải danh sách sản phẩm. Vui lòng thử lại.'
}
