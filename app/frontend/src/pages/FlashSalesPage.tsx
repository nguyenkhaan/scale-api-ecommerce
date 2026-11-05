import { Link } from 'react-router-dom'
import { CartActionNotice } from '../cart/CartActionNotice'
import { useAddToCart } from '../cart/use-add-to-cart'
import { FlashSaleCampaignCard } from '../flash-sale/FlashSaleCampaignCard'
import { usePagedList } from '../lib/use-paged-list'
import { flashSaleClient } from '../lib/flash-sale-api'

const listPublic = (cursor?: string) => flashSaleClient.listPublicFlashSales({ cursor })

export function FlashSalesPage() {
  const sales = usePagedList(listPublic, 'chương trình flash sale')
  const cartAction = useAddToCart()

  return (
    <div className="mx-auto max-w-5xl">
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">Flash Sale</h1>
          <p className="mt-1 text-sm text-slate-600">
            Săn deal chớp nhoáng với giá ưu đãi đặc biệt trong thời gian giới hạn.
          </p>
        </div>
        <Link
          to="/cart"
          className="inline-flex min-h-11 items-center gap-2 rounded bg-amber-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-amber-500 focus-visible:outline-2 focus-visible:outline-amber-600"
        >
          Xem giỏ hàng
        </Link>
      </div>

      <CartActionNotice notice={cartAction.notice} loginNext="/flash-sales" />

      {sales.loading && (
        <div className="py-12 text-center text-slate-500" role="status">
          Đang tải danh sách sự kiện flash sale...
        </div>
      )}

      {sales.error && (
        <div role="alert" className="mt-6 rounded border border-rose-200 bg-rose-50 p-4 text-rose-800">
          <p className="font-semibold">{sales.error}</p>
          <button
            type="button"
            onClick={() => void sales.reload()}
            className="mt-2 min-h-9 text-sm font-semibold underline hover:text-rose-950"
          >
            Thử lại
          </button>
        </div>
      )}

      {!sales.loading && !sales.error && sales.items.length === 0 && (
        <div className="mt-8 rounded border border-dashed border-slate-300 p-12 text-center">
          <p className="text-base font-medium text-slate-700">Hiện tại không có chương trình Flash Sale nào.</p>
          <p className="mt-1 text-sm text-slate-500">Vui lòng quay lại sau!</p>
        </div>
      )}

      <div className="mt-6 space-y-8">
        {sales.items.map((campaign) => (
          <FlashSaleCampaignCard
            key={campaign.id}
            campaign={campaign}
            addingSkuId={cartAction.busyKey}
            onAdd={(skuId, flashSaleItemId, productName) =>
              void cartAction.add({ skuId, amount: 1, priceSource: 'flash_sale', flashSaleItemId }, productName)
            }
          />
        ))}
      </div>

      {sales.hasMore && (
        <div className="mt-6 text-center">
          <button
            type="button"
            disabled={sales.loadingMore}
            onClick={() => void sales.loadMore()}
            className="min-h-11 border border-slate-900 px-4 text-sm font-semibold text-slate-900 hover:bg-stone-100 focus-visible:outline-2 focus-visible:outline-amber-500 disabled:opacity-50"
          >
            {sales.loadingMore ? 'Đang tải...' : 'Tải thêm'}
          </button>
        </div>
      )}
    </div>
  )
}
