import { describe, expect, it } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'

import { ProductDetailView } from '@/features/catalog/pages/ProductDetailPage'
import type { CatalogProduct } from '@/features/catalog/catalog.type'

const product: CatalogProduct = {
  id: 'product-1',
  name: 'Tai nghe SaleScale',
  description: 'Âm thanh rõ, thiết kế gọn.',
  image_url: null,
  active: true,
  category_id: 'category-1',
  skus: [{ id: 'sku-1', code: 'HEADPHONE-BLK', attributes: { color: 'black' }, regular_price: 450000, active: true }],
}

describe('ProductDetailView', () => {
  it('shows product details, category and active SKU purchase controls', () => {
    const markup = renderToStaticMarkup(
      <ProductDetailView product={product} categoryName="Điện tử" busy={false} onAdd={() => undefined} />,
    )

    expect(markup).toContain('Tai nghe SaleScale')
    expect(markup).toContain('Âm thanh rõ, thiết kế gọn.')
    expect(markup).toContain('Điện tử')
    expect(markup).toContain('HEADPHONE-BLK')
    expect(markup).toContain('Thêm vào giỏ')
  })

  it('shows unavailable state without purchase controls when no active SKU exists', () => {
    const unavailable = { ...product, skus: [{ ...product.skus[0], active: false }] }
    const markup = renderToStaticMarkup(
      <ProductDetailView product={unavailable} categoryName="Điện tử" busy={false} onAdd={() => undefined} />,
    )

    expect(markup).toContain('Hiện chưa có phiên bản đang bán.')
    expect(markup).not.toContain('Thêm vào giỏ')
  })
})
