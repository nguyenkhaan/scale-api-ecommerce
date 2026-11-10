import { describe, expect, it } from 'bun:test'
import type { CartItem, CheckoutLine, CheckoutPreview } from '../lib/cart-api'
import { detectPreviewChanges, selectionSignature } from './preview-changes'

const line = (sku: string, unit: number, quantity = 1, source: CheckoutLine['price_source'] = 'flash_sale'): CheckoutLine => ({
  cart_item_id: `i-${sku}`,
  cart_item_version: 1,
  sku_id: sku,
  quantity,
  unit_price: unit,
  subtotal: unit * quantity,
  price_source: source,
  flash_sale_item_id: null,
})
const preview = (...lines: CheckoutLine[]): CheckoutPreview => ({
  currency: 'VND',
  total_amount: lines.reduce((sum, l) => sum + l.subtotal, 0),
  lines,
  previewed_at: '2026-11-05T00:00:00Z',
})

describe('detectPreviewChanges', () => {
  it('reports no changes without a previous preview', () => {
    expect(detectPreviewChanges(null, preview(line('a', 100)))).toEqual([])
  })

  it('reports lines whose unit price changed (old -> new)', () => {
    const changes = detectPreviewChanges(preview(line('a', 100), line('b', 50)), preview(line('a', 120), line('b', 50)))
    expect(changes).toEqual([{ sku_id: 'a', oldUnitPrice: 100, newUnitPrice: 120, oldSource: 'flash_sale', newSource: 'flash_sale' }])
  })

  it('reports a price source change even when the unit price matches', () => {
    expect(detectPreviewChanges(preview(line('a', 100)), preview(line('a', 100, 1, 'regular')))).toHaveLength(1)
  })

  it('ignores quantity-only differences and lines absent from either side', () => {
    expect(detectPreviewChanges(preview(line('a', 100, 1)), preview(line('a', 100, 3), line('c', 10)))).toEqual([])
  })
})

describe('selectionSignature', () => {
  const item = (id: string, version: number) => ({ id, version }) as CartItem
  it('changes when a version or the selection changes, not with order', () => {
    expect(selectionSignature([item('a', 1), item('b', 2)])).toBe(selectionSignature([item('b', 2), item('a', 1)]))
    expect(selectionSignature([item('a', 1)])).not.toBe(selectionSignature([item('a', 2)]))
    expect(selectionSignature([item('a', 1)])).not.toBe(selectionSignature([item('a', 1), item('b', 1)]))
  })
})
