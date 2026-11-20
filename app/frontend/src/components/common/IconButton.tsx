import type { ButtonHTMLAttributes, ReactNode } from 'react'

export type IconButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  label: string
  pressed?: boolean
  children?: ReactNode
}

export function IconButton({
  label,
  pressed,
  className = '',
  type = 'button',
  children,
  ...props
}: IconButtonProps) {
  return (
    <button
      {...props}
      type={type}
      aria-label={label}
      aria-pressed={pressed}
      className={`inline-flex min-h-11 min-w-11 items-center justify-center rounded-[var(--ss-radius-control)] border border-transparent text-[var(--ss-color-text)] transition-colors hover:bg-[var(--ss-color-app-bg)] disabled:cursor-not-allowed disabled:opacity-50 ${className}`}
    >
      {children}
    </button>
  )
}
