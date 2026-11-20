import type { AriaRole, ReactNode } from 'react'

export type AlertTone = 'info' | 'success' | 'warning' | 'danger'

export type AlertProps = {
  tone: AlertTone
  children: ReactNode
  role?: AriaRole
  className?: string
}

const toneClasses: Record<AlertTone, string> = {
  info: 'border-blue-200 bg-[var(--ss-color-info-surface)] text-blue-900',
  success: 'border-teal-200 bg-teal-50 text-[var(--ss-color-success-strong)]',
  warning: 'border-amber-300 bg-amber-50 text-[var(--ss-color-warning-strong)]',
  danger: 'border-red-200 bg-[var(--ss-color-danger-surface)] text-[var(--ss-color-danger)]',
}

export function Alert({ tone, children, role, className = '' }: AlertProps) {
  return (
    <div role={role ?? (tone === 'danger' ? 'alert' : 'status')} className={`border p-3 text-sm ${toneClasses[tone]} ${className}`}>
      {children}
    </div>
  )
}
