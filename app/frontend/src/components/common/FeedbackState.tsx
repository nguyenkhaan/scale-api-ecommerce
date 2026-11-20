import type { ReactNode } from 'react'
import { Button } from './Button'

export type FeedbackKind = 'loading' | 'empty' | 'error' | 'forbidden' | 'notFound' | 'overload'

export type FeedbackStateProps = {
  kind: FeedbackKind
  title: string
  message: string
  action?: {
    label: string
    onClick: () => void
    disabled?: boolean
  }
  children?: ReactNode
}

export function FeedbackState({ kind, title, message, action, children }: FeedbackStateProps) {
  const isAlert = kind === 'error' || kind === 'forbidden' || kind === 'notFound'
  return (
    <section
      role={isAlert ? 'alert' : 'status'}
      aria-live={isAlert ? 'assertive' : 'polite'}
      data-kind={kind}
      className="flex flex-col items-center justify-center gap-2 border border-[var(--ss-color-border)] bg-[var(--ss-color-surface)] px-6 py-10 text-center"
    >
      <p className="font-semibold text-[var(--ss-color-text)]">{title}</p>
      <p className="max-w-prose text-sm text-[var(--ss-color-muted)]">{message}</p>
      {children}
      {action && (
        <Button variant="secondary" disabled={action.disabled} onClick={action.onClick}>
          {action.label}
        </Button>
      )}
    </section>
  )
}
