import { useCallback, useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient, type Query } from '@tanstack/react-query'

import { usePagedList } from '@/lib/use-paged-list'

import { describeCancelFailure, isFatalPollError, nextOrderPollDelayMs, orderTabs, tabToStatus, type OrderTabKey } from './bill.logic'
import { billClient, createOrderPageFetcher } from './bill.service'
import type { Order } from './bill.type'

export const orderQueryKeys = {
  detail: (orderId: string) => ['orders', orderId] as const,
}

const tabFetchers = Object.fromEntries(
  orderTabs.map((tab) => [tab.key, createOrderPageFetcher(billClient, tabToStatus(tab.key))]),
) as Record<OrderTabKey, ReturnType<typeof createOrderPageFetcher>>

/** First page on mount, "Tải thêm" appends by cursor. Remount (key) when the tab changes. */
export function useOrders(tab: OrderTabKey) {
  return usePagedList(tabFetchers[tab], 'danh sách đơn hàng')
}

/**
 * Polling interval callback that counts consecutive errors itself: TanStack Query resets
 * `fetchFailureCount` at the start of every fetch, so it cannot drive an exponential backoff.
 * The count is `errorUpdateCount` minus its value at the last successful update.
 */
export function createOrderRefetchInterval() {
  let lastSuccessUpdate = 0
  let errorsAtLastSuccess = 0
  return (query: Query<Order>): number | false => {
    const { dataUpdateCount, errorUpdateCount, data, error } = query.state
    if (dataUpdateCount !== lastSuccessUpdate) {
      lastSuccessUpdate = dataUpdateCount
      errorsAtLastSuccess = errorUpdateCount
    }
    return nextOrderPollDelayMs({
      status: data?.status,
      failures: errorUpdateCount - errorsAtLastSuccess,
      fatal: isFatalPollError(error),
    })
  }
}

/** Official order state. With `poll`, refetches every 2 s while pending_payment (exponential backoff on errors). */
export function useOrderQuery(orderId: string, options: { poll?: boolean } = {}) {
  const refetchInterval = useMemo(() => (options.poll ? createOrderRefetchInterval() : false), [options.poll])
  return useQuery({
    queryKey: orderQueryKeys.detail(orderId),
    queryFn: () => billClient.getOrder(orderId),
    retry: false,
    refetchInterval,
  })
}

export type CancelNotice = { kind: 'conflict' | 'error'; text: string }

export function useCancelOrder(orderId: string) {
  const queryClient = useQueryClient()
  const [notice, setNotice] = useState<CancelNotice | null>(null)
  const mutation = useMutation({
    mutationKey: ['orders', 'cancel'],
    mutationFn: () => billClient.cancelOrder(orderId),
  })

  const cancel = useCallback(async (): Promise<boolean> => {
    setNotice(null)
    try {
      const order = await mutation.mutateAsync()
      queryClient.setQueryData<Order>(orderQueryKeys.detail(orderId), order)
      return true
    } catch (err) {
      const failure = describeCancelFailure(err)
      setNotice(failure)
      if (failure.kind === 'conflict') await queryClient.invalidateQueries({ queryKey: orderQueryKeys.detail(orderId) })
      return false
    }
  }, [mutation, orderId, queryClient])

  return { cancel, cancelling: mutation.isPending, notice }
}

/** Ticks every second for the display-only countdown. */
export function useNow(intervalMs = 1000): number {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs])
  return now
}
