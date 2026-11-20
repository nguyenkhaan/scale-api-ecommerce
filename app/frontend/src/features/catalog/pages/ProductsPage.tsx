import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'

import { FeedbackState, LoadMoreButton, SearchField } from '@/components/common'
import { describeProductsLoadError, useProductsQuery } from '@/features/catalog/catalog.hook'
import { ProductCard } from '@/features/catalog/components/ProductCard'
import { CartActionNotice } from '@/features/cart/components/CartActionNotice'
import { useAddToCart } from '@/features/cart/cart.hook'
import { isCatalogApiError } from '../catalog.service'

export function ProductsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const query = searchParams.get('query')?.trim() ?? ''
  const productsQuery = useProductsQuery(query)
  const cartAction = useAddToCart()
  const products = productsQuery.data?.pages.flatMap((page) => page.items) ?? []
  const error = productsQuery.isError ? describeProductsLoadError(productsQuery.error) : null
  const loading = productsQuery.isPending

  function submitSearch(searchInput: string) {
    const next = new URLSearchParams(searchParams)
    const normalized = searchInput.trim()
    if (normalized === '') next.delete('query')
    else next.set('query', normalized)
    setSearchParams(next)
  }

  const feedbackKind = isCatalogApiError(productsQuery.error) && productsQuery.error.status === 429 ? 'overload' : 'error'
  const loginNext = query === '' ? '/products' : `/products?query=${encodeURIComponent(query)}`

  return (
    <div className="mx-auto max-w-6xl">
      <div className="flex flex-wrap items-end justify-between gap-4 border-b border-[var(--ss-color-border)] pb-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-[var(--ss-color-text)]">Sản phẩm</h1>
          <p className="mt-1 text-sm text-[var(--ss-color-muted)]">Tìm và thêm sản phẩm vào giỏ hàng.</p>
        </div>
        <Link
          to="/cart"
          className="inline-flex min-h-11 items-center rounded-[var(--ss-radius-control)] bg-[var(--ss-color-primary)] px-4 py-2 text-sm font-semibold text-white hover:bg-[var(--ss-color-primary-strong)]"
        >
          Xem giỏ hàng
        </Link>
      </div>

      <div className="mt-4 max-w-xl">
        <CatalogSearchField key={query} query={query} onSubmit={submitSearch} />
      </div>

      <CartActionNotice notice={cartAction.notice} loginNext={loginNext} />

      {loading && (
        <div className="mt-6"><FeedbackState kind="loading" title="Đang tải danh sách sản phẩm" message="Vui lòng chờ trong giây lát." /></div>
      )}

      {error && (
        <div className="mt-6">
          <FeedbackState
            kind={feedbackKind}
            title="Không thể tải danh sách sản phẩm"
            message={error}
            action={{ label: 'Thử lại', onClick: () => void productsQuery.refetch() }}
          />
        </div>
      )}

      {!loading && !error && products.length === 0 && (
        <div className="mt-6"><FeedbackState kind="empty" title="Chưa có sản phẩm nào." message={query ? 'Thử tìm với từ khóa khác.' : 'Danh mục hiện chưa có sản phẩm đang bán.'} /></div>
      )}

      <ul aria-label="Danh sách sản phẩm" className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {products.map((product) => (
          <ProductCard
            key={product.id}
            product={product}
            busy={cartAction.busyKey !== null}
            onAdd={(skuId, amount, name) => void cartAction.add({ skuId, amount, priceSource: 'regular' }, name)}
          />
        ))}
      </ul>

      {productsQuery.hasNextPage && (
        <div className="mt-6 text-center">
          <LoadMoreButton loading={productsQuery.isFetchingNextPage} onClick={() => void productsQuery.fetchNextPage()} />
        </div>
      )}
    </div>
  )
}

function CatalogSearchField({ query, onSubmit }: { query: string; onSubmit: (value: string) => void }) {
  const [value, setValue] = useState(query)
  return <SearchField value={value} placeholder="Tìm theo tên sản phẩm hoặc SKU" onChange={setValue} onSubmit={() => onSubmit(value)} />
}
