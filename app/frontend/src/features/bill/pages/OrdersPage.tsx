import { useState } from 'react'

import { FeedbackState, LoadMoreButton } from '@/components/common'

import { useOrders } from '../bill.hook'
import type { OrderTabKey } from '../bill.logic'
import type { Order } from '../bill.type'
import { OrderListRow } from '../components/OrderListRow'
import { OrderStatusTabs } from '../components/OrderStatusTabs'

export type OrderListState = {
  items: Order[]
  loading: boolean
  loadingMore: boolean
  error: string | null
  hasMore: boolean
  reload: () => Promise<unknown>
  loadMore: () => Promise<unknown>
}

export function OrdersPage() {
  const [tab, setTab] = useState<OrderTabKey>('all')
  return (
    <div className="mx-auto max-w-4xl">
      <div className="border-b border-[var(--ss-color-border)] pb-4">
        <h1 className="text-2xl font-semibold tracking-tight text-[var(--ss-color-text)]">Đơn của tôi</h1>
        <p className="mt-1 text-sm text-[var(--ss-color-muted)]">Theo dõi trạng thái thanh toán và lịch sử đơn hàng.</p>
      </div>
      <div className="mt-4">
        <OrderStatusTabs active={tab} onChange={setTab} />
      </div>
      <OrdersList key={tab} tab={tab} />
    </div>
  )
}

/** Remounted per tab so each status filter starts from its own first page. */
function OrdersList({ tab }: { tab: OrderTabKey }) {
  return <OrdersListView list={useOrders(tab)} filtered={tab !== 'all'} />
}

export function OrdersListView({ list, filtered }: { list: OrderListState; filtered: boolean }) {
  return (
    <div className="mt-4">
      {list.loading && <FeedbackState kind="loading" title="Đang tải đơn hàng" message="Vui lòng chờ trong giây lát." />}

      {list.error && (
        <FeedbackState kind="error" title="Không thể tải đơn hàng" message={list.error} action={{ label: 'Thử lại', onClick: () => void list.reload() }} />
      )}

      {!list.loading && !list.error && list.items.length === 0 && (
        <FeedbackState
          kind="empty"
          title={filtered ? 'Không có đơn hàng nào ở trạng thái này.' : 'Bạn chưa có đơn hàng nào.'}
          message={filtered ? 'Thử chọn trạng thái khác.' : 'Đơn hàng sau khi đặt sẽ hiển thị tại đây.'}
        />
      )}

      {list.items.length > 0 && (
        <ul className="space-y-3">
          {list.items.map((order) => (
            <OrderListRow key={order.id} order={order} />
          ))}
        </ul>
      )}

      {list.hasMore && !list.loading && (
        <div className="mt-6 text-center">
          <LoadMoreButton loading={list.loadingMore} onClick={() => void list.loadMore()} />
        </div>
      )}
    </div>
  )
}
