import type { FlashSaleCampaign } from '../lib/flash-sale-api'
import { formatDateTime, formatVND } from '../lib/format'
import { Countdown } from './Countdown'
import { saleStatusBadgeClass, saleStatusLabel } from './flash-sale-status'

type Props = {
  campaign: FlashSaleCampaign
  addingSkuId: string | null
  onAdd: (skuId: string, flashSaleItemId: string, productName: string) => void
}

export function FlashSaleCampaignCard({ campaign, addingSkuId, onAdd }: Props) {
  const isActive = campaign.status === 'active'
  return (
    <section
      className="overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm"
      aria-labelledby={`campaign-title-${campaign.id}`}
    >
      <div className="border-b border-slate-100 bg-slate-50 px-4 py-4 sm:px-6">
        <div className="flex flex-wrap items-center gap-3">
          <h2 id={`campaign-title-${campaign.id}`} className="text-xl font-bold text-slate-900">
            {campaign.name}
          </h2>
          <span className={`inline-flex rounded-full px-3 py-1 text-xs font-bold ${saleStatusBadgeClass(campaign.status)}`}>
            {isActive ? '● ' : ''}
            {saleStatusLabel(campaign.status)}
          </span>
        </div>
        <p className="mt-1 text-xs text-slate-500">
          Thời gian: {formatDateTime(campaign.starts_at)} — {formatDateTime(campaign.ends_at)}
        </p>
        <Countdown status={campaign.status} startsAt={campaign.starts_at} endsAt={campaign.ends_at} />
      </div>

      <div className="p-4 sm:p-6">
        <h3 className="text-sm font-semibold uppercase tracking-wider text-slate-500">
          Sản phẩm ưu đãi ({campaign.items.length})
        </h3>
        <ul className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {campaign.items.map((item) => {
            const adding = addingSkuId === item.sku_id
            return (
              <li key={item.id} className="flex flex-col justify-between rounded border border-slate-200 p-4">
                <div>
                  <p className="break-words font-semibold text-slate-900">{item.product_name}</p>
                  <p className="font-mono text-xs text-slate-500">SKU: {item.sku_code}</p>
                  <p className="mt-2 text-sm text-slate-500 line-through">
                    <span className="sr-only">Giá thường </span>
                    {formatVND(item.regular_price)}
                  </p>
                  <p className="text-2xl font-bold text-rose-600">
                    <span className="sr-only">Giá flash sale </span>
                    {formatVND(item.sale_price)}
                  </p>
                  <p className="mt-1 text-xs text-slate-600">
                    Giới hạn mua: <span className="font-semibold">{item.purchase_limit} cái/khách</span>
                  </p>
                </div>
                {isActive && (
                  <div className="mt-4 border-t border-slate-100 pt-3">
                    <button
                      type="button"
                      disabled={adding}
                      onClick={() => onAdd(item.sku_id, item.id, item.product_name)}
                      className="min-h-11 w-full rounded bg-rose-600 px-3 py-2 text-xs font-bold text-white hover:bg-rose-500 focus-visible:outline-2 focus-visible:outline-rose-600 disabled:cursor-not-allowed disabled:bg-slate-300"
                    >
                      {adding ? 'Đang thêm...' : 'Thêm vào giỏ'}
                    </button>
                  </div>
                )}
              </li>
            )
          })}
        </ul>
      </div>
    </section>
  )
}
