import { useState } from 'react'
import { Button, FeedbackState, LoadMoreButton } from '@/components/common'
import type { AdminAction } from '../admin-flashsale.logic'
import { AdminCampaignCard } from '../components/AdminCampaignCard'
import { CampaignFormModal } from '../components/CampaignFormModal'
import { useAdminFlashSales } from '../admin-flashsale.hook'
import { useAdminSkuOptions } from '../admin-sku-options.hook'
import type { FlashSaleCampaign } from '@/features/flashsale/flashsale.type'

/** `null` = closed, `'new'` = create, a campaign = edit that draft. */
type FormTarget = null | 'new' | FlashSaleCampaign

export type AdminFlashSaleListState = {
  items: FlashSaleCampaign[]
  loading: boolean
  loadingMore: boolean
  error: string | null
  hasMore: boolean
  reload: () => Promise<unknown>
  loadMore: () => Promise<unknown>
}

export type AdminFlashSalePageViewProps = {
  list: AdminFlashSaleListState
  busyId: string | null
  notice: { kind: 'success' | 'error'; text: string } | null
  onCreate: () => void
  onAction: (campaign: FlashSaleCampaign, action: AdminAction) => void
  onLoadMore: () => void
}

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
    <>
      <AdminFlashSalePageView
        list={list}
        busyId={sales.busyId}
        notice={sales.notice}
        onCreate={() => setFormTarget('new')}
        onAction={handleAction}
        onLoadMore={() => void list.loadMore()}
      />

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
    </>
  )
}

export function AdminFlashSalePageView({ list, busyId, notice, onCreate, onAction, onLoadMore }: AdminFlashSalePageViewProps) {
  return (
    <div className="mx-auto max-w-6xl">
      <div className="flex flex-wrap items-end justify-between gap-4 border-b border-[var(--ss-color-border)] pb-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-[var(--ss-color-text)]">Quản lý Flash Sale</h1>
          <p className="mt-1 text-sm text-[var(--ss-color-muted)]">Tạo, lên lịch và kiểm soát các chiến dịch flash sale và phân bổ suất cho từng SKU.</p>
        </div>
        <Button onClick={onCreate}>Tạo chiến dịch mới</Button>
      </div>

      {notice && <div className="mt-4"><FeedbackState kind={notice.kind === 'error' ? 'error' : 'loading'} title={notice.kind === 'error' ? 'Thao tác chưa thành công' : 'Đã cập nhật'} message={notice.text} /></div>}

      {list.error && (
        <div className="mt-4">
          <FeedbackState kind="error" title="Không thể tải danh sách chiến dịch" message={list.error} action={{ label: 'Thử lại', onClick: () => void list.reload() }} />
        </div>
      )}

      {list.loading && <div className="mt-6"><FeedbackState kind="loading" title="Đang tải danh sách chiến dịch" message="Vui lòng chờ trong giây lát." /></div>}

      {!list.loading && !list.error && list.items.length === 0 && (
        <div className="mt-6"><FeedbackState kind="empty" title="Chưa có chiến dịch Flash Sale nào." message="Tạo bản nháp đầu tiên để bắt đầu." /></div>
      )}

      {list.items.length > 0 && (
        <ul className="mt-6 space-y-3">
          {list.items.map((campaign) => (
            <AdminCampaignCard key={campaign.id} campaign={campaign} busy={busyId === campaign.id} onAction={(action) => onAction(campaign, action)} />
          ))}
        </ul>
      )}

      {list.hasMore && (
        <div className="mt-6 text-center">
          <LoadMoreButton loading={list.loadingMore} onClick={onLoadMore} />
        </div>
      )}
    </div>
  )
}
