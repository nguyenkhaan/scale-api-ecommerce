import type { ButtonHTMLAttributes, ReactNode } from 'react'

export type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'ghost'

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant
  loading?: boolean
  children?: ReactNode
}

const variantClasses: Record<ButtonVariant, string> = {
  primary: 'bg-[var(--ss-color-primary)] text-white hover:bg-[var(--ss-color-primary-strong)]',
  secondary: 'border border-[var(--ss-color-input-border)] bg-white text-[var(--ss-color-text)] hover:bg-[var(--ss-color-app-bg)]',
  danger: 'bg-[var(--ss-color-danger)] text-white hover:bg-red-800',
  ghost: 'bg-transparent text-[var(--ss-color-primary-strong)] hover:bg-red-50',
}

export function Button({
  variant = 'primary',
  loading = false,
  disabled = false,
  className = '',
  type = 'button',
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      {...props}
      type={type}
      data-variant={variant}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      className={`inline-flex min-h-10 items-center justify-center gap-2 rounded-[var(--ss-radius-control)] px-4 py-2 text-sm font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${variantClasses[variant]} ${className}`}
    >
      {loading ? (
        <>
          <span aria-hidden="true" className="inline-block size-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
          <span>Đang xử lý…</span>
        </>
      ) : (
        children
      )}
    </button>
  )
}
