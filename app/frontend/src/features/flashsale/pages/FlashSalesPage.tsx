import { Link } from 'react-router-dom'
import type { AddToCartInput } from '@/features/cart/cart.logic'
import { FeedbackState, LoadMoreButton } from '@/components/common'
import { CartActionNotice } from '@/features/cart/components/CartActionNotice'
import { useAddToCart } from '@/features/cart/cart.hook'
import { FlashSaleCampaignCard } from '../components/FlashSaleCampaignCard'
import { useFlashSales } from '../flashsale.hook'
import type { FlashSaleCampaign } from '../flashsale.type'

type FlashSalesState = {
  items: FlashSaleCampaign[]
  loading: boolean
  loadingMore: boolean
  error: string | null
  hasMore: boolean
  reload: () => Promise<unknown>
  loadMore: () => Promise<unknown>
}

type CartActionState = {
  notice: ReturnType<typeof useAddToCart>['notice']
  busyKey: string | null
  add: (input: AddToCartInput, productName: string) => Promise<boolean>
}

export type FlashSalesPageViewProps = {
  sales: FlashSalesState
  cartAction: CartActionState
}

export function FlashSalesPage() {
  const sales = useFlashSales()
  const cartAction = useAddToCart()
  return <FlashSalesPageView sales={sales} cartAction={cartAction} />
}

export function FlashSalesPageView({ sales, cartAction }: FlashSalesPageViewProps) {
  return (
    <div className="mx-auto max-w-6xl">
      <div className="flex flex-wrap items-end justify-between gap-4 border-b border-[var(--ss-color-border)] pb-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-[var(--ss-color-text)]">Flash Sale</h1>
          <p className="mt-1 text-sm text-[var(--ss-color-muted)]">
            Săn deal chớp nhoáng với giá ưu đãi đặc biệt trong thời gian giới hạn.
          </p>
        </div>
        <Link
          to="/cart"
          className="inline-flex min-h-11 items-center gap-2 rounded-[var(--ss-radius-control)] bg-[var(--ss-color-primary)] px-4 py-2 text-sm font-semibold text-white hover:bg-[var(--ss-color-primary-strong)]"
        >
          Xem giỏ hàng
        </Link>
      </div>

      <CartActionNotice notice={cartAction.notice} loginNext="/flash-sales" />

      {sales.loading && (
        <div className="mt-6"><FeedbackState kind="loading" title="Đang tải danh sách flash sale" message="Vui lòng chờ trong giây lát." /></div>
      )}

      {sales.error && (
        <div className="mt-6">
          <FeedbackState kind="error" title="Không thể tải flash sale" message={sales.error} action={{ label: 'Thử lại', onClick: () => void sales.reload() }} />
        </div>
      )}

      {!sales.loading && !sales.error && sales.items.length === 0 && (
        <div className="mt-6"><FeedbackState kind="empty" title="Hiện tại không có chương trình Flash Sale nào." message="Vui lòng quay lại sau." /></div>
      )}

      <div className="mt-6 space-y-6">
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
          <LoadMoreButton loading={sales.loadingMore} onClick={() => void sales.loadMore()} />
        </div>
      )}
    </div>
  )
}
