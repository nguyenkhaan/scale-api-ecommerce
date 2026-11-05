import { describe, expect, it } from 'bun:test'
import { countdownTarget, formatCountdown, saleStatusLabel } from './flash-sale-status'

describe('saleStatusLabel', () => {
  it('labels every status, never calling an ended sale upcoming', () => {
    expect(saleStatusLabel('active')).toBe('ĐANG DIỄN RA')
    expect(saleStatusLabel('scheduled')).toBe('SẮP DIỄN RA')
    expect(saleStatusLabel('ended')).toBe('ĐÃ KẾT THÚC')
    expect(saleStatusLabel('stopped')).toBe('ĐÃ DỪNG')
    expect(saleStatusLabel('draft')).toBe('BẢN NHÁP')
  })
})

describe('countdown', () => {
  it('targets start for scheduled, end for active, nothing otherwise', () => {
    expect(countdownTarget('scheduled', 'S', 'E')).toEqual({ label: 'Bắt đầu sau', at: 'S' })
    expect(countdownTarget('active', 'S', 'E')).toEqual({ label: 'Kết thúc sau', at: 'E' })
    expect(countdownTarget('ended', 'S', 'E')).toBeNull()
  })

  it('formats remaining time and returns null once reached', () => {
    const now = Date.parse('2026-11-05T00:00:00Z')
    expect(formatCountdown('2026-11-05T01:02:03Z', now)).toBe('01:02:03')
    expect(formatCountdown('2026-11-07T00:00:05Z', now)).toBe('2 ngày 00:00:05')
    expect(formatCountdown('2026-11-05T00:00:00Z', now)).toBeNull()
    expect(formatCountdown('not-a-date', now)).toBeNull()
  })
})
