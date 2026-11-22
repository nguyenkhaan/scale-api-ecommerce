import { isApiError } from '@/lib/api-error'
import { describeLoadError } from '@/lib/load-error'

import type { Order, OrderItem, OrderStatus } from './bill.type'

export type OrderTabKey = 'all' | OrderStatus

export const orderTabs: ReadonlyArray<{ key: OrderTabKey; label: string }> = [
  { key: 'all', label: 'Tất cả' },
  { key: 'pending_payment', label: 'Chờ thanh toán' },
  { key: 'confirmed', label: 'Đã xác nhận' },
  { key: 'payment_failed', label: 'Thanh toán thất bại' },
  { key: 'cancelled', label: 'Đã hủy' },
  { key: 'expired', label: 'Hết hạn' },
]

export function tabToStatus(tab: OrderTabKey): OrderStatus | undefined {
  return tab === 'all' ? undefined : tab
}

/**
 * `payable`: server says the order can still be paid.
 * `overdue`: still pending_payment but past the deadline, the worker has not released stock yet.
 * `closed`: any other status.
 */
export type OrderAvailability = 'payable' | 'overdue' | 'closed'

export function orderAvailability(order: Pick<Order, 'status' | 'payable'>): OrderAvailability {
  if (order.status !== 'pending_payment') return 'closed'
  return order.payable ? 'payable' : 'overdue'
}

export const canPayOrder = (order: Pick<Order, 'status' | 'payable'>) => orderAvailability(order) === 'payable'
export const canCancelOrder = canPayOrder

/** Display-only countdown as `mm:ss` (or `h:mm:ss`); null when the deadline is invalid or passed. */
export function formatCountdown(expiresAt: string, nowMs: number): string | null {
  const remaining = Math.floor((new Date(expiresAt).getTime() - nowMs) / 1000)
  if (!Number.isFinite(remaining) || remaining <= 0) return null
  const hours = Math.floor(remaining / 3600)
  const minutes = Math.floor((remaining % 3600) / 60)
  const seconds = remaining % 60
  const pad = (value: number) => String(value).padStart(2, '0')
  return hours > 0 ? `${hours}:${pad(minutes)}:${pad(seconds)}` : `${pad(minutes)}:${pad(seconds)}`
}

export function itemAttributesText(attributes: OrderItem['sku_attributes']): string {
  if (!attributes) return ''
  return Object.entries(attributes)
    .map(([key, value]) => `${key}: ${String(value)}`)
    .join(' · ')
}

export function totalQuantity(order: Pick<Order, 'items'>): number {
  return order.items.reduce((sum, item) => sum + item.quantity, 0)
}

export function describeOrderLoadError(err: unknown): string {
  if (isApiError(err) && err.status === 404) return 'Không tìm thấy đơn hàng.'
  return describeLoadError(err, 'đơn hàng')
}

export type CancelFailure = { kind: 'conflict' | 'error'; text: string }

/** A 409 means the order is no longer cancellable: the caller must refetch and show the real state. */
export function describeCancelFailure(err: unknown): CancelFailure {
  if (isApiError(err)) {
    if (err.status === 409) {
      return { kind: 'conflict', text: 'Đơn hàng đã thay đổi trạng thái nên không thể hủy. Đang hiển thị trạng thái mới nhất.' }
    }
    if (err.status === 404) return { kind: 'error', text: 'Không tìm thấy đơn hàng.' }
    if (err.code === 'NETWORK_ERROR') return { kind: 'error', text: 'Lỗi kết nối khi hủy đơn. Vui lòng thử lại.' }
    if (err.status === 429) return { kind: 'error', text: 'Bạn thao tác quá nhanh. Vui lòng thử lại sau.' }
  }
  return { kind: 'error', text: 'Không thể hủy đơn. Vui lòng thử lại.' }
}

export const ORDER_POLL_INTERVAL_MS = 2000
export const ORDER_POLL_MAX_BACKOFF_MS = 30_000

/**
 * Next delay for polling an order, or false to stop. Polls only while pending_payment; consecutive
 * failures back off exponentially; a fatal error (401/403/404) will not fix itself, so stop.
 */
export function nextOrderPollDelayMs(input: { status: OrderStatus | undefined; failures: number; fatal?: boolean }): number | false {
  if (input.fatal) return false
  if (input.status !== undefined && input.status !== 'pending_payment') return false
  const backoff = ORDER_POLL_INTERVAL_MS * 2 ** Math.max(0, input.failures)
  return Math.min(backoff, ORDER_POLL_MAX_BACKOFF_MS)
}

export function isFatalPollError(err: unknown): boolean {
  return isApiError(err) && (err.status === 401 || err.status === 403 || err.status === 404)
}
