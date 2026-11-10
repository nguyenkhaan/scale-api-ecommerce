import { describe, expect, it } from 'bun:test'
import { ApiError } from '../lib/api-error'
import { fieldMessages } from '../lib/violation-messages'
import { adminActionsFor, describeAdminFailure } from './admin-actions'
import { buildRequest, emptyForm, isClaimedField, type CampaignForm } from './campaign-form'

const valid = (): CampaignForm => ({
  name: 'Giờ vàng',
  startsAt: '2026-11-11T10:00',
  endsAt: '2026-11-11T12:00',
  items: [{ sku_id: 's1', sale_price: '500', quota_total: '10', purchase_limit: '2' }],
})

describe('adminActionsFor', () => {
  it('allows the right actions per status', () => {
    expect(adminActionsFor('draft')).toEqual(['edit', 'schedule'])
    expect(adminActionsFor('scheduled')).toEqual(['cancel_schedule'])
    expect(adminActionsFor('active')).toEqual(['stop'])
    expect(adminActionsFor('ended')).toEqual([])
    expect(adminActionsFor('stopped')).toEqual([])
  })
})

describe('describeAdminFailure', () => {
  it('maps overlap and state conflict to clear messages', () => {
    const overlap = describeAdminFailure(new ApiError({ status: 409, code: 'SALE_SCHEDULE_OVERLAP' }))
    expect(overlap.text).toContain('trùng khung giờ')
    const conflict = describeAdminFailure(new ApiError({ status: 409, code: 'STATE_TRANSITION_CONFLICT' }))
    expect(conflict.stateConflict).toBe(true)
    expect(conflict.text).not.toContain('STATE_TRANSITION_CONFLICT')
  })
})

describe('buildRequest', () => {
  it('builds an ISO request for a valid form', () => {
    const result = buildRequest(valid(), () => 1000)
    expect(result.ok).toBe(true)
    if (result.ok) {
      expect(result.request.items[0]).toEqual({ sku_id: 's1', sale_price: 500, quota_total: 10, purchase_limit: 2 })
      expect(result.request.starts_at).toBe(new Date('2026-11-11T10:00').toISOString())
    }
  })

  it('reports violations on the matching fields', () => {
    const result = buildRequest(emptyForm(), () => undefined)
    expect(result.ok).toBe(false)
    if (!result.ok) {
      const fields = result.violations.map((v) => v.field)
      expect(fields).toContain('name')
      expect(fields).toContain('starts_at')
      expect(fields).toContain('items[0].sku_id')
      expect(fieldMessages(result.violations, 'name')).toEqual(['Tên chiến dịch là bắt buộc.'])
    }
  })

  it('rejects a sale price not below the regular price and an end before the start', () => {
    const form = { ...valid(), endsAt: '2026-11-11T09:00', items: [{ ...valid().items[0], sale_price: '1000' }] }
    const result = buildRequest(form, () => 1000)
    expect(result.ok).toBe(false)
    if (!result.ok) {
      expect(result.violations.map((v) => v.field)).toEqual(['ends_at', 'items[0].sale_price'])
    }
  })
})

describe('isClaimedField', () => {
  it('claims known form fields and rows, leaving the rest for the summary', () => {
    expect(isClaimedField('items[0].sale_price', 1)).toBe(true)
    expect(isClaimedField('items[3].sale_price', 1)).toBe(false)
    expect(isClaimedField('body', 1)).toBe(false)
  })
})
