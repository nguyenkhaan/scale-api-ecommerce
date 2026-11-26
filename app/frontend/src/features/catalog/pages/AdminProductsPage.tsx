import { Link, useSearchParams } from 'react-router-dom'

import { Button, FeedbackState, LoadMoreButton } from '@/components/common'
import { useAdminProductsQuery } from '@/features/catalog/catalog.hook'
import { isCatalogApiError } from '../catalog.service'

export function AdminProductsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const status = (searchParams.get('status') ?? 'active') as 'active' | 'inactive' | 'all'
  const productsQuery = useAdminProductsQuery(status === 'all' ? undefined : status)
  const products = productsQuery.data?.pages.flatMap((page) => page.items) ?? []

  return (
    <section className="mx-auto max-w-6xl">
      <div className="flex flex-wrap items-end justify-between gap-4 border-b border-[var(--ss-color-border)] pb-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-[var(--ss-color-text)]">Sản phẩm và SKU</h1>
          <p className="mt-1 text-sm text-[var(--ss-color-muted)]">Quản lý thông tin hàng hóa và các phiên bản đang bán.</p>
        </div>
        <Link to="/admin/products/new"><Button>Tạo sản phẩm</Button></Link>
      </div>

      <label className="mt-4 flex max-w-xs flex-col gap-1 text-sm font-medium text-[var(--ss-color-text)]">
        Trạng thái sản phẩm
        <select aria-label="Trạng thái sản phẩm" value={status} onChange={(event) => setSearchParams(event.target.value === 'active' ? {} : { status: event.target.value })} className="min-h-11 rounded-[var(--ss-radius-control)] border border-[var(--ss-color-input-border)] bg-white px-3">
          <option value="active">Đang bán</option>
          <option value="inactive">Đã ẩn</option>
          <option value="all">Tất cả</option>
        </select>
      </label>

      {productsQuery.isPending && <div className="mt-5"><FeedbackState kind="loading" title="Đang tải sản phẩm" message="Vui lòng chờ trong giây lát." /></div>}
      {productsQuery.isError && <div className="mt-5"><FeedbackState kind={isCatalogApiError(productsQuery.error) && productsQuery.error.status === 403 ? 'forbidden' : 'error'} title="Không thể tải sản phẩm" message="Kiểm tra quyền quản trị hoặc thử tải lại." action={{ label: 'Thử lại', onClick: () => void productsQuery.refetch() }} /></div>}
      {!productsQuery.isPending && !productsQuery.isError && products.length === 0 && <div className="mt-5"><FeedbackState kind="empty" title="Chưa có sản phẩm trong trạng thái này." message="Tạo sản phẩm mới hoặc đổi bộ lọc trạng thái." /></div>}

      {products.length > 0 && (
        <ul className="mt-5 divide-y divide-[var(--ss-color-border)] border-y border-[var(--ss-color-border)]">
          {products.map((product) => (
            <li key={product.id} className="flex flex-wrap items-center justify-between gap-3 py-4">
              <div>
                <h2 className="font-semibold text-[var(--ss-color-text)]">{product.name}</h2>
                <p className="mt-1 text-xs text-[var(--ss-color-muted)]">{product.skus.length} SKU · {product.active ? 'Đang bán' : 'Đã ẩn'}</p>
              </div>
              <Link to={`/admin/products/${product.id}`} className="inline-flex min-h-11 items-center px-3 text-sm font-semibold text-[var(--ss-color-primary-strong)] hover:underline">Chỉnh sửa</Link>
            </li>
          ))}
        </ul>
      )}

      {productsQuery.hasNextPage && <div className="mt-5 text-center"><LoadMoreButton loading={productsQuery.isFetchingNextPage} onClick={() => void productsQuery.fetchNextPage()} /></div>}
    </section>
  )
}
