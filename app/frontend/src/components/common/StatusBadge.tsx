import type { ReactNode } from 'react'

export type StatusDomain = 'order' | 'payment' | 'flashSale' | 'priceSource'
export type StatusAudience = 'customer' | 'admin'

export type StatusBadgeProps = {
  domain: StatusDomain
  value: string
  audience?: StatusAudience
  className?: string
}

type StatusDefinition = {
  label: string | ((audience: StatusAudience) => string)
  tone: 'neutral' | 'warning' | 'primary' | 'success' | 'danger'
  icon: 'edit' | 'calendar' | 'bolt' | 'pause' | 'check' | 'clock' | 'alert' | 'tag'
}

const definitions: Record<StatusDomain, Record<string, StatusDefinition>> = {
  order: {
    pending_payment: { label: 'Chờ thanh toán', tone: 'primary', icon: 'clock' },
    confirmed: { label: 'Đã xác nhận', tone: 'success', icon: 'check' },
    payment_failed: { label: 'Thanh toán thất bại', tone: 'danger', icon: 'alert' },
    cancelled: { label: 'Đã hủy', tone: 'neutral', icon: 'pause' },
    expired: { label: 'Hết hạn', tone: 'neutral', icon: 'pause' },
  },
  payment: {
    pending: { label: 'Đang chờ', tone: 'primary', icon: 'clock' },
    unknown: { label: 'Chưa xác định', tone: 'warning', icon: 'alert' },
    succeeded: { label: 'Thành công', tone: 'success', icon: 'check' },
    failed: { label: 'Thất bại', tone: 'danger', icon: 'alert' },
  },
  flashSale: {
    draft: { label: 'Nháp', tone: 'neutral', icon: 'edit' },
    scheduled: {
      label: (audience) => (audience === 'admin' ? 'Đã lên lịch' : 'Sắp diễn ra'),
      tone: 'warning',
      icon: 'calendar',
    },
    active: { label: 'Đang diễn ra', tone: 'primary', icon: 'bolt' },
    stopped: { label: 'Đã dừng', tone: 'neutral', icon: 'pause' },
    ended: { label: 'Đã kết thúc', tone: 'neutral', icon: 'check' },
  },
  priceSource: {
    regular: { label: 'Giá thường', tone: 'neutral', icon: 'tag' },
    flash_sale: { label: 'Flash sale', tone: 'primary', icon: 'bolt' },
  },
}

const toneClasses: Record<StatusDefinition['tone'], string> = {
  neutral: 'border-[var(--ss-color-input-border)] bg-[var(--ss-color-app-bg)] text-[var(--ss-color-muted)]',
  warning: 'border-[var(--ss-color-warning)] bg-amber-50 text-[var(--ss-color-warning-strong)]',
  primary: 'border-red-200 bg-red-50 text-[var(--ss-color-primary-strong)]',
  success: 'border-teal-300 bg-teal-50 text-[var(--ss-color-success-strong)]',
  danger: 'border-red-300 bg-red-50 text-[var(--ss-color-danger)]',
}

function StatusIcon({ icon }: { icon: StatusDefinition['icon'] }) {
  return (
    <svg width="12" height="12" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" data-status-icon={icon}>
      {icon === 'check' && <path d="m3.5 8.25 2.5 2.5 6.5-6.5" />}
      {icon === 'alert' && <path d="M8 2.25 14.5 13.5h-13ZM8 6.5v3M8 11.6v.2" />}
      {icon === 'clock' && <><circle cx="8" cy="8" r="6.25" /><path d="M8 4.75V8l2.25 1.5" /></>}
      {icon === 'calendar' && <><path d="M2.5 3.5h11v10h-11Z M2.5 6.5h11 M5.5 2v3 M10.5 2v3" /></>}
      {icon === 'edit' && <path d="m10.5 2.5 3 3-8 8H2.5v-3Z" />}
      {icon === 'bolt' && <path d="M9 1.5 3.5 9H8l-1 5.5L12.5 7H8Z" />}
      {icon === 'pause' && <><circle cx="8" cy="8" r="6.25" /><path d="M6.5 5.5v5M9.5 5.5v5" /></>}
      {icon === 'tag' && <path d="M2.5 3.5v4l6 6 4-4-6-6Z M5.25 5.25h.01" />}
    </svg>
  )
}

export function StatusBadge({ domain, value, audience = 'customer', className = '' }: StatusBadgeProps) {
  const definition = definitions[domain][value] ?? { label: 'Không xác định', tone: 'neutral', icon: 'alert' as const }
  const label = typeof definition.label === 'function' ? definition.label(audience) : definition.label
  return (
    <span className={`inline-flex items-center gap-1 rounded-[var(--ss-radius-control)] border px-2 py-0.5 text-xs font-semibold ${toneClasses[definition.tone]} ${className}`} data-status-domain={domain} data-status-value={value}>
      <StatusIcon icon={definition.icon} />
      {label}
    </span>
  )
}

export type StatusIconNode = ReactNode
