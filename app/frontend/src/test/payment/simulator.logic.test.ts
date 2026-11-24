import { describe, expect, it } from 'bun:test'

import { ApiError } from '@/lib/api-error'
import { buildSimulatorRequest, defaultSimulatorForm, describeSimulatorError } from '@/features/payment/simulator.logic'

describe('simulator form to request body', () => {
  it('uses defaults and omits blank overrides', () => {
    const result = buildSimulatorRequest('s1', defaultSimulatorForm)
    expect(result).toEqual({
      ok: true,
      body: { payment_session_id: 's1', event_type: 'succeeded', delay_ms: 0, repeat_count: 1, drop_delivery: false, signature_mode: 'valid' },
    })
    if (result.ok) {
      expect('amount_override' in result.body).toBe(false)
      expect('currency_override' in result.body).toBe(false)
    }
  })

  it('treats whitespace-only overrides as blank', () => {
    const result = buildSimulatorRequest('s1', { ...defaultSimulatorForm, amountOverride: '  ', currencyOverride: ' ' })
    expect(result.ok && 'amount_override' in result.body).toBe(false)
    expect(result.ok && 'currency_override' in result.body).toBe(false)
  })

  it('maps every option into the body', () => {
    const result = buildSimulatorRequest('s1', {
      eventType: 'failed',
      delayMs: '2000',
      repeatCount: '3',
      dropDelivery: true,
      invalidSignature: true,
      amountOverride: '1000',
      currencyOverride: ' USD ',
    })
    expect(result).toEqual({
      ok: true,
      body: {
        payment_session_id: 's1',
        event_type: 'failed',
        delay_ms: 2000,
        repeat_count: 3,
        drop_delivery: true,
        signature_mode: 'invalid',
        amount_override: 1000,
        currency_override: 'USD',
      },
    })
  })

  it('requires an amount override of at least 1', () => {
    const result = buildSimulatorRequest('s1', { ...defaultSimulatorForm, amountOverride: '0' })
    expect(result.ok).toBe(false)
    if (!result.ok) expect(result.errors.amountOverride).toBe('Số tiền ghi đè phải là số nguyên từ 1 trở lên.')
    const one = buildSimulatorRequest('s1', { ...defaultSimulatorForm, amountOverride: '1' })
    expect(one.ok && one.body.amount_override).toBe(1)
  })

  it('rejects invalid numbers with field errors and no body', () => {
    const result = buildSimulatorRequest('s1', { ...defaultSimulatorForm, delayMs: '-5', repeatCount: '0', amountOverride: '1.5' })
    expect(result.ok).toBe(false)
    if (!result.ok) expect(Object.keys(result.errors).sort()).toEqual(['amountOverride', 'delayMs', 'repeatCount'])
  })
})

describe('simulator error copy', () => {
  it('explains disabled simulator (404) and forbidden (403) without raw codes', () => {
    expect(describeSimulatorError(new ApiError({ status: 404, code: 'NOT_FOUND' }))).toContain('đang tắt')
    expect(describeSimulatorError(new ApiError({ status: 403, code: 'FORBIDDEN' }))).toContain('quản trị viên')
  })

  it('falls back to generic load copy', () => {
    expect(describeSimulatorError(new ApiError({ status: 500, code: 'INTERNAL' }))).toContain('Vui lòng thử lại')
  })
})

describe('simulator 422 mapping', () => {
  it('translates server violations into field messages', () => {
    const err = new ApiError({ status: 422, code: 'VALIDATION_FAILED', violations: [{ field: 'amount_override', reason: 'must be at least 1' }] })
    expect(describeSimulatorError(err)).toBe('Số tiền ghi đè phải từ 1 trở lên.')
  })

  it('falls back to generic copy when a 422 has no violations', () => {
    expect(describeSimulatorError(new ApiError({ status: 422, code: 'VALIDATION_FAILED' }))).toContain('không hợp lệ')
  })
})
