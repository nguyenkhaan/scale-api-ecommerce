import { useInfiniteQuery, useQuery } from '@tanstack/react-query'

import { formatVND } from '@/lib/format'

import { getAdminProduct, getProduct, isCatalogApiError, listAdminProducts, listCategories, listProducts } from './catalog.service'
import type { CatalogProduct, CatalogSku, ProductListParams } from './catalog.type'

const pageSize = 20

export const catalogQueryKeys = {
  categories: ['catalog', 'categories'] as const,
  product: (productID: string) => ['catalog', 'product', productID] as const,
  products: (limit: number, query: string, categoryID = '') => ['catalog', 'products', { limit, query, categoryID }] as const,
  adminProduct: (productID: string) => ['catalog', 'admin-product', productID] as const,
  adminProducts: (limit: number, status: string) => ['catalog', 'admin-products', { limit, status }] as const,
}

export function productListParams(query: string, cursor?: string, categoryID?: string): ProductListParams {
  const params: ProductListParams = { limit: pageSize }
  const normalizedQuery = query.trim()
  if (normalizedQuery !== '') params.query = normalizedQuery
  if (cursor !== undefined) params.cursor = cursor
  if (categoryID) params.category_id = categoryID
  return params
}

export function useProductsQuery(query: string, categoryID = '') {
  return useInfiniteQuery({
    queryKey: catalogQueryKeys.products(pageSize, query, categoryID),
    initialPageParam: undefined as string | undefined,
    queryFn: ({ pageParam }) => listProducts(productListParams(query, pageParam, categoryID)),
    getNextPageParam: (page) => page.next_cursor ?? undefined,
  })
}

export function useCategoriesQuery() {
  return useQuery({ queryKey: catalogQueryKeys.categories, queryFn: listCategories, retry: false })
}

export function useProductQuery(productID: string) {
  return useQuery({ queryKey: catalogQueryKeys.product(productID), queryFn: () => getProduct(productID), enabled: productID !== '', retry: false })
}

export function useAdminProductQuery(productID: string) {
  return useQuery({ queryKey: catalogQueryKeys.adminProduct(productID), queryFn: () => getAdminProduct(productID), enabled: productID !== '', retry: false })
}

export function useAdminProductsQuery(status: 'active' | 'inactive' | undefined) {
  return useInfiniteQuery({
    queryKey: catalogQueryKeys.adminProducts(pageSize, status ?? ''),
    initialPageParam: undefined as string | undefined,
    queryFn: ({ pageParam }) => listAdminProducts({ limit: pageSize, status, cursor: pageParam }),
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
