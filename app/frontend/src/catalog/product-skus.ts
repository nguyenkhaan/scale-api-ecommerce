import type { CatalogProduct, CatalogSku } from '../lib/catalog-api'
import { formatVND } from '../lib/format'

/** SKUs a customer or admin may pick: active SKU of an active product. */
export function activeSkus(product: CatalogProduct): CatalogSku[] {
  return product.active ? product.skus.filter((sku) => sku.active) : []
}

export type SkuOption = {
  skuId: string
  regularPrice: number
  /** "Tên sản phẩm — mã SKU — giá thường" */
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

/** Keeps SKUs already used by a campaign selectable (and labelled) even if they are no longer active. */
export function mergeCampaignSkuOptions(
  options: SkuOption[],
  items: Array<{ sku_id: string; sku_code: string; product_name: string; regular_price: number }>,
): SkuOption[] {
  const known = new Set(options.map((o) => o.skuId))
  const extra = items
    .filter((item) => !known.has(item.sku_id))
    .map((item) => ({
      skuId: item.sku_id,
      regularPrice: item.regular_price,
      label: `${item.product_name} — ${item.sku_code} — ${formatVND(item.regular_price)}`,
    }))
  return [...options, ...extra]
}
