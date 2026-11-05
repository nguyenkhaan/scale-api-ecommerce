import type { FlashSaleStatus } from '../lib/flash-sale-api'

const labels: Record<FlashSaleStatus, string> = {
  active: 'ĐANG DIỄN RA',
  scheduled: 'SẮP DIỄN RA',
  ended: 'ĐÃ KẾT THÚC',
  stopped: 'ĐÃ DỪNG',
  draft: 'BẢN NHÁP',
}

const badgeClasses: Record<FlashSaleStatus, string> = {
  active: 'bg-emerald-100 text-emerald-800',
  scheduled: 'bg-amber-100 text-amber-800',
  ended: 'bg-stone-200 text-stone-700',
  stopped: 'bg-rose-100 text-rose-800',
  draft: 'bg-slate-100 text-slate-800',
}

export function saleStatusLabel(status: FlashSaleStatus): string {
  return labels[status] ?? 'KHÔNG XÁC ĐỊNH'
}

export function saleStatusBadgeClass(status: FlashSaleStatus): string {
  return badgeClasses[status] ?? 'bg-slate-100 text-slate-800'
}

/** Display-only countdown target; the server remains the authority on whether a sale is open. */
export function countdownTarget(
  status: FlashSaleStatus,
  startsAt: string,
  endsAt: string,
): { label: string; at: string } | null {
  if (status === 'scheduled') return { label: 'Bắt đầu sau', at: startsAt }
  if (status === 'active') return { label: 'Kết thúc sau', at: endsAt }
  return null
}

/** `1 ngày 02:03:04` / `02:03:04`; null when the target is invalid, 0 or past. */
export function formatCountdown(targetISO: string, nowMs: number): string | null {
  const diff = new Date(targetISO).getTime() - nowMs
  if (!Number.isFinite(diff) || diff <= 0) return null
  const totalSeconds = Math.floor(diff / 1000)
  const days = Math.floor(totalSeconds / 86400)
  const pad = (n: number) => String(n).padStart(2, '0')
  const hms = `${pad(Math.floor((totalSeconds % 86400) / 3600))}:${pad(Math.floor((totalSeconds % 3600) / 60))}:${pad(totalSeconds % 60)}`
  return days > 0 ? `${days} ngày ${hms}` : hms
}
