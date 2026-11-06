import { useState } from 'react'
import type { AdminAction } from '../admin-flash-sale/admin-actions'
import { AdminCampaignCard } from '../admin-flash-sale/AdminCampaignCard'
import { CampaignFormModal } from '../admin-flash-sale/CampaignFormModal'
import { useAdminFlashSales } from '../admin-flash-sale/use-admin-flash-sales'
import { useAdminSkuOptions } from '../admin-flash-sale/use-admin-sku-options'
import type { FlashSaleCampaign } from '../lib/flash-sale-api'

/** `null` = closed, `'new'` = create, a campaign = edit that draft. */
type FormTarget = null | 'new' | FlashSaleCampaign

export function AdminFlashSalePage() {
  const sales = useAdminFlashSales()
  const skus = useAdminSkuOptions()
  const [formTarget, setFormTarget] = useState<FormTarget>(null)
  const { list } = sales

  const handleAction = (campaign: FlashSaleCampaign, action: AdminAction) => {
    if (action === 'edit') setFormTarget(campaign)
    else void sales.runAction(campaign, action)
  }

  return (
    <div className="mx-auto max-w-6xl">
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 pb-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">Quản lý Flash Sale</h1>
          <p className="mt-1 text-sm text-slate-600">
            Tạo, lên lịch và kiểm soát các chiến dịch flash sale và phân bổ suất cho từng SKU.
          </p>
        </div>
        <button
          type="button"
          onClick={() => setFormTarget('new')}
          className="min-h-11 rounded bg-slate-900 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-slate-800 focus-visible:outline-2 focus-visible:outline-amber-500"
        >
          + Tạo chiến dịch mới
        </button>
      </div>

      {sales.notice && (
        <div
          role={sales.notice.kind === 'error' ? 'alert' : 'status'}
          className={`mt-4 rounded border p-4 text-sm font-medium ${
            sales.notice.kind === 'error'
              ? 'border-rose-200 bg-rose-50 text-rose-800'
              : 'border-emerald-300 bg-emerald-50 text-emerald-900'
          }`}
        >
          {sales.notice.text}
        </div>
      )}

      {list.error && (
        <div role="alert" className="mt-4 rounded border border-rose-200 bg-rose-50 p-4 text-rose-800">
          <p className="font-semibold">{list.error}</p>
          <button type="button" onClick={() => void list.reload()} className="mt-2 min-h-9 text-sm font-semibold underline">
            Thử lại
          </button>
        </div>
      )}

      {list.loading && (
        <div className="py-12 text-center text-slate-500" role="status">
          Đang tải danh sách chiến dịch...
        </div>
      )}

      {!list.loading && !list.error && list.items.length === 0 && (
        <div className="mt-8 rounded border border-dashed border-slate-300 p-12 text-center">
          <p className="text-base font-medium text-slate-700">Chưa có chiến dịch Flash Sale nào.</p>
        </div>
      )}

      {list.items.length > 0 && (
        <ul className="mt-6 space-y-4">
          {list.items.map((campaign) => (
            <AdminCampaignCard
              key={campaign.id}
              campaign={campaign}
              busy={sales.busyId === campaign.id}
              onAction={(action) => handleAction(campaign, action)}
            />
          ))}
        </ul>
      )}

      {list.hasMore && (
        <div className="mt-6 text-center">
          <button
            type="button"
            disabled={list.loadingMore}
            onClick={() => void list.loadMore()}
            className="min-h-11 border border-slate-900 px-4 text-sm font-semibold text-slate-900 hover:bg-stone-100 focus-visible:outline-2 focus-visible:outline-amber-500 disabled:opacity-50"
          >
            {list.loadingMore ? 'Đang tải...' : 'Tải thêm'}
          </button>
        </div>
      )}

      {formTarget !== null && (
        <CampaignFormModal
          campaign={formTarget === 'new' ? undefined : formTarget}
          skuOptions={skus.options}
          skuLoadError={skus.error}
          onClose={() => setFormTarget(null)}
          onSubmit={async (request) => {
            await sales.save(request, formTarget === 'new' ? undefined : formTarget)
            setFormTarget(null)
          }}
        />
      )}
    </div>
  )
}
