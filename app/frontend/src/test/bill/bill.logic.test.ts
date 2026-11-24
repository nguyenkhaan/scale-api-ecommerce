import { describe, expect, it } from 'bun:test'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'

import { StatusBadge } from '@/components/common'
import { ApiError } from '@/lib/api-error'
import {
  canCancelOrder,
  canPayOrder,
  describeCancelFailure,
  describeOrderLoadError,
  formatCountdown,
  itemAttributesText,
  nextOrderPollDelayMs,
  orderAvailability,
  orderTabs,
  tabToStatus,
} from '@/features/bill/bill.logic'
import type { OrderStatus } from '@/features/bill/bill.type'

describe('order status labels', () => {
  const labels: Record<OrderStatus, string> = {
    pending_payment: 'Chờ thanh toán',
    confirmed: 'Đã xác nhận',
    payment_failed: 'Thanh toán thất bại',
    cancelled: 'Đã hủy',
    expired: 'Hết hạn',
  }

  for (const [status, label] of Object.entries(labels)) {
    it(`renders "${label}" for ${status}`, () => {
      const markup = renderToStaticMarkup(createElement(StatusBadge, { domain: 'order', value: status }))
      expect(markup).toContain(label)
      expect(markup).toContain(`data-status-value="${status}"`)
    })
  }

  it('lists the status tabs in display order with an all tab', () => {
    expect(orderTabs.map((tab) => tab.label)).toEqual(['Tất cả', 'Chờ thanh toán', 'Đã xác nhận', 'Thanh toán thất bại', 'Đã hủy', 'Hết hạn'])
    expect(tabToStatus('all')).toBeUndefined()
    expect(tabToStatus('expired')).toBe('expired')
  })
})

describe('payable vs overdue pending display', () => {
  it('treats pending + payable as payable', () => {
    expect(orderAvailability({ status: 'pending_payment', payable: true })).toBe('payable')
  })

  it('treats pending without payable as overdue, not payable', () => {
    expect(orderAvailability({ status: 'pending_payment', payable: false })).toBe('overdue')
  })

  it('treats every other status as closed even if payable were true', () => {
    for (const status of ['confirmed', 'payment_failed', 'cancelled', 'expired'] as const) {
      expect(orderAvailability({ status, payable: true })).toBe('closed')
    }
  })

  it('shows pay and cancel only when payable', () => {
    expect(canPayOrder({ status: 'pending_payment', payable: true })).toBe(true)
    expect(canCancelOrder({ status: 'pending_payment', payable: true })).toBe(true)
    expect(canPayOrder({ status: 'pending_payment', payable: false })).toBe(false)
    expect(canCancelOrder({ status: 'confirmed', payable: false })).toBe(false)
  })
})

describe('countdown and attribute text', () => {
  const expiresAt = '2026-10-10T13:46:00Z'
  const expiresMs = new Date(expiresAt).getTime()

  it('formats remaining time as mm:ss', () => {
    expect(formatCountdown(expiresAt, expiresMs - 192_000)).toBe('03:12')
  })

  it('formats hours when more than an hour remains', () => {
    expect(formatCountdown(expiresAt, expiresMs - 3_725_000)).toBe('1:02:05')
  })

  it('returns null at or after the deadline and for invalid dates', () => {
    expect(formatCountdown(expiresAt, expiresMs)).toBeNull()
    expect(formatCountdown(expiresAt, expiresMs + 1000)).toBeNull()
    expect(formatCountdown('not-a-date', 0)).toBeNull()
  })

  it('joins attributes and tolerates a missing map', () => {
    expect(itemAttributesText({ Màu: 'Đen', 'Kích cỡ': 'M' })).toBe('Màu: Đen · Kích cỡ: M')
    expect(itemAttributesText(undefined)).toBe('')
  })
})

describe('order polling decision', () => {
  it('continues every 2 s while pending_payment', () => {
    expect(nextOrderPollDelayMs({ status: 'pending_payment', failures: 0 })).toBe(2000)
  })

  it('polls before the first response arrives', () => {
    expect(nextOrderPollDelayMs({ status: undefined, failures: 0 })).toBe(2000)
  })

  it('stops on every terminal status', () => {
    for (const status of ['confirmed', 'payment_failed', 'cancelled', 'expired'] as const) {
      expect(nextOrderPollDelayMs({ status, failures: 0 })).toBe(false)
    }
  })

  it('backs off exponentially on errors with a cap', () => {
    expect(nextOrderPollDelayMs({ status: 'pending_payment', failures: 1 })).toBe(4000)
    expect(nextOrderPollDelayMs({ status: 'pending_payment', failures: 2 })).toBe(8000)
    expect(nextOrderPollDelayMs({ status: 'pending_payment', failures: 20 })).toBe(30_000)
  })

  it('stops on a fatal error', () => {
    expect(nextOrderPollDelayMs({ status: undefined, failures: 1, fatal: true })).toBe(false)
  })
})

describe('error copy', () => {
  it('maps a 409 cancel to a conflict that asks for a refetch', () => {
    const failure = describeCancelFailure(new ApiError({ status: 409, code: 'STATE_TRANSITION_CONFLICT' }))
    expect(failure.kind).toBe('conflict')
    expect(failure.text).not.toContain('STATE_TRANSITION_CONFLICT')
  })

  it('maps other cancel failures to a generic error without raw codes', () => {
    expect(describeCancelFailure(new ApiError({ status: 0, code: 'NETWORK_ERROR' })).kind).toBe('error')
    expect(describeCancelFailure(new Error('boom')).text).toBe('Không thể hủy đơn. Vui lòng thử lại.')
  })

  it('maps 404 to the not-found copy', () => {
    expect(describeOrderLoadError(new ApiError({ status: 404, code: 'NOT_FOUND' }))).toBe('Không tìm thấy đơn hàng.')
    expect(describeOrderLoadError(new ApiError({ status: 500, code: 'INTERNAL' }))).toBe('Không thể tải đơn hàng. Vui lòng thử lại.')
  })
})
