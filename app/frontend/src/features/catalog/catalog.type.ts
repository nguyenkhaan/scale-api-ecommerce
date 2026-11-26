import type { components, operations } from '@/contract/api-contract'

type ApiSchemas = components['schemas']

export type CatalogSku = ApiSchemas['Sku']
export type CatalogProduct = ApiSchemas['Product']
export type ProductPage = ApiSchemas['ProductPage']
export type CatalogCategory = ApiSchemas['Category']
export type CategoryPage = ApiSchemas['CategoryPage']
export type CreateProductRequest = ApiSchemas['CreateProductRequest']
export type UpdateProductRequest = ApiSchemas['UpdateProductRequest']
export type CreateSkuInput = ApiSchemas['CreateSkuInput']
export type UpdateSkuRequest = ApiSchemas['UpdateSkuRequest']
export type ProductListParams = NonNullable<operations['listProducts']['parameters']['query']>
export type AdminProductListParams = NonNullable<operations['listAdminProducts']['parameters']['query']>

export type FieldViolation = {
  field: string
  reason: string
}

export type CatalogApiErrorInit = {
  status: number
  code: string
  requestID?: string
  violations?: FieldViolation[]
  retryAfterSeconds?: number
}
