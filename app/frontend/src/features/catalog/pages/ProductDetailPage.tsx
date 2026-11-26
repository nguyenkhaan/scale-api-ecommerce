import { useState } from 'react'
import { useParams } from 'react-router-dom'

import { Button, FeedbackState, PriceBlock, QuantityStepper, Surface } from '@/components/common'
import { CartActionNotice } from '@/features/cart/components/CartActionNotice'
import { useAddToCart } from '@/features/cart/cart.hook'
import { activeSkus, describeProductsLoadError, useCategoriesQuery, useProductQuery } from '@/features/catalog/catalog.hook'
import type { CatalogProduct } from '@/features/catalog/catalog.type'
import { isCatalogApiError } from '../catalog.service'

export function ProductDetailPage() {
  const { productId = '' } = useParams()
  const productQuery = useProductQuery(productId)
  const categoriesQuery = useCategoriesQuery()
  const cartAction = useAddToCart()

  if (productQuery.isPending) {
    return <div className="mx-auto max-w-5xl"><FeedbackState kind="loading" title="Đang tải sản phẩm" message="Vui lòng chờ trong giây lát." /></div>
  }
  if (productQuery.isError || !productQuery.data) {
    const notFound = isCatalogApiError(productQuery.error) && productQuery.error.status === 404
    return <div className="mx-auto max-w-5xl"><FeedbackState kind={notFound ? 'notFound' : 'error'} title={notFound ? 'Không tìm thấy sản phẩm' : 'Không thể tải sản phẩm'} message={notFound ? 'Sản phẩm này không còn được bán hoặc không tồn tại.' : describeProductsLoadError(productQuery.error)} action={{ label: 'Thử lại', onClick: () => void productQuery.refetch() }} /></div>
  }

  const category = categoriesQuery.data?.items.find((item) => item.id === productQuery.data.category_id)
  return (
    <div className="mx-auto max-w-5xl">
      <a href="/products" className="inline-flex min-h-11 items-center text-sm font-semibold text-[var(--ss-color-primary-strong)] hover:underline">← Quay lại sản phẩm</a>
      <ProductDetailView
        product={productQuery.data}
        categoryName={category?.name ?? null}
        busy={cartAction.busyKey !== null}
        onAdd={(skuId, quantity, name) => void cartAction.add({ skuId, amount: quantity, priceSource: 'regular' }, name)}
      />
      <CartActionNotice notice={cartAction.notice} loginNext={`/products/${productId}`} />
    </div>
  )
}

export type ProductDetailViewProps = {
  product: CatalogProduct
  categoryName: string | null
  busy: boolean
  onAdd: (skuId: string, quantity: number, productName: string) => void
}

export function ProductDetailView({ product, categoryName, busy, onAdd }: ProductDetailViewProps) {
  const skus = activeSkus(product)
  const [skuID, setSkuID] = useState(skus[0]?.id ?? '')
  const [quantity, setQuantity] = useState(1)
  const selected = skus.find((sku) => sku.id === skuID) ?? skus[0]

  return (
    <article className="mt-3 grid gap-6 border-t border-[var(--ss-color-border)] pt-6 md:grid-cols-[minmax(0,1.2fr)_minmax(18rem,0.8fr)]">
      <div>
        {product.image_url && <img src={product.image_url} alt="" className="mb-5 aspect-[4/3] w-full border border-[var(--ss-color-border)] object-cover" />}
        <p className="text-xs font-semibold uppercase tracking-wide text-[var(--ss-color-primary-strong)]">{categoryName ?? 'Sản phẩm'}</p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight text-[var(--ss-color-text)]">{product.name}</h1>
        <p className="mt-4 whitespace-pre-line text-sm leading-6 text-[var(--ss-color-muted)]">{product.description || 'Thông tin sản phẩm đang được cập nhật.'}</p>
      </div>

      <Surface as="section" className="h-fit p-5">
        <h2 className="text-lg font-semibold text-[var(--ss-color-text)]">Chọn phiên bản</h2>
        {selected ? (
          <div className="mt-4 space-y-4">
            <label className="flex flex-col gap-1 text-sm font-medium text-[var(--ss-color-text)]">
              SKU
              <select aria-label="SKU" value={selected.id} onChange={(event) => setSkuID(event.target.value)} className="min-h-11 rounded-[var(--ss-radius-control)] border border-[var(--ss-color-input-border)] bg-white px-3">
                {skus.map((sku) => <option key={sku.id} value={sku.id}>{sku.code}{Object.keys(sku.attributes).length ? ` — ${Object.values(sku.attributes).join(', ')}` : ''}</option>)}
              </select>
            </label>
            <PriceBlock regularPrice={selected.regular_price} effectivePrice={selected.regular_price} source="regular" />
            <QuantityStepper value={quantity} min={1} max={99} busy={busy} label={`Số lượng ${product.name}`} onChange={setQuantity} />
            <Button disabled={!Number.isInteger(quantity) || quantity < 1 || quantity > 99} loading={busy} className="w-full" onClick={() => onAdd(selected.id, quantity, product.name)}>Thêm vào giỏ</Button>
          </div>
        ) : (
          <p className="mt-3 text-sm text-[var(--ss-color-muted)]">Hiện chưa có phiên bản đang bán.</p>
        )}
      </Surface>
    </article>
  )
}
