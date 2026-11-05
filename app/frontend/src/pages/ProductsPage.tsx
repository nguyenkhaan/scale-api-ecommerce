import { Link } from 'react-router-dom'
import { CartActionNotice } from '../cart/CartActionNotice'
import { useAddToCart } from '../cart/use-add-to-cart'
import { ProductCard } from '../catalog/ProductCard'
import { catalogClient } from '../lib/catalog-api'
import { usePagedList } from '../lib/use-paged-list'

const PAGE_SIZE = 20
const listProducts = (cursor?: string) => catalogClient.listProducts({ limit: PAGE_SIZE, cursor })

/** Minimal regular-price shopping entry point until the full catalog UI lands. */
export function ProductsPage() {
  const products = usePagedList(listProducts, 'danh sách sản phẩm')
  const cartAction = useAddToCart()

  return (
    <div className="mx-auto max-w-5xl">
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">Sản phẩm</h1>
          <p className="mt-1 text-sm text-slate-600">Thêm sản phẩm vào giỏ hàng với giá thường.</p>
        </div>
        <Link
          to="/cart"
          className="inline-flex min-h-11 items-center rounded bg-amber-600 px-4 py-2 text-sm font-semibold text-white hover:bg-amber-500 focus-visible:outline-2 focus-visible:outline-amber-600"
        >
          Xem giỏ hàng
        </Link>
      </div>

      <CartActionNotice notice={cartAction.notice} loginNext="/products" />

      {products.loading && (
        <div className="py-12 text-center text-slate-500" role="status">
          Đang tải danh sách sản phẩm...
        </div>
      )}

      {products.error && (
        <div role="alert" className="mt-6 rounded border border-rose-200 bg-rose-50 p-4 text-rose-800">
          <p className="font-semibold">{products.error}</p>
          <button type="button" onClick={() => void products.reload()} className="mt-2 min-h-9 text-sm font-semibold underline">
            Thử lại
          </button>
        </div>
      )}

      {!products.loading && !products.error && products.items.length === 0 && (
        <div className="mt-8 rounded border border-dashed border-slate-300 p-12 text-center text-slate-700">
          Chưa có sản phẩm nào.
        </div>
      )}

      <ul className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {products.items.map((product) => (
          <ProductCard
            key={product.id}
            product={product}
            busy={cartAction.busyKey !== null}
            onAdd={(skuId, amount, name) =>
              void cartAction.add({ skuId, amount, priceSource: 'regular' }, name)
            }
          />
        ))}
      </ul>

      {products.hasMore && (
        <div className="mt-6 text-center">
          <button
            type="button"
            disabled={products.loadingMore}
            onClick={() => void products.loadMore()}
            className="min-h-11 border border-slate-900 px-4 text-sm font-semibold text-slate-900 hover:bg-stone-100 focus-visible:outline-2 focus-visible:outline-amber-500 disabled:opacity-50"
          >
            {products.loadingMore ? 'Đang tải...' : 'Tải thêm'}
          </button>
        </div>
      )}
    </div>
  )
}
