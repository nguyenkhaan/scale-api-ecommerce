import type { CartItem, CheckoutPreview, PriceSource } from '../lib/cart-api'

export type PreviewChange = {
  sku_id: string
  oldUnitPrice: number
  newUnitPrice: number
  oldSource: PriceSource
  newSource: PriceSource
}

/**
 * Lines present in both previews whose unit price or price source differs. Quantity-only differences are not
 * price changes, so adding units never forces an acknowledgement.
 */
export function detectPreviewChanges(previous: CheckoutPreview | null, next: CheckoutPreview): PreviewChange[] {
  if (!previous) return []
  const before = new Map(previous.lines.map((line) => [line.sku_id, line]))
  return next.lines.flatMap((line): PreviewChange[] => {
    const old = before.get(line.sku_id)
    if (!old) return []
    if (old.unit_price === line.unit_price && old.price_source === line.price_source) return []
    return [
      {
        sku_id: line.sku_id,
        oldUnitPrice: old.unit_price,
        newUnitPrice: line.unit_price,
        oldSource: old.price_source,
        newSource: line.price_source,
      },
    ]
  })
}

/** Identifies exactly which cart lines and versions a preview was computed for. */
export function selectionSignature(items: CartItem[]): string {
  return items
    .map((item) => `${item.id}:${item.version}`)
    .sort()
    .join('|')
}
