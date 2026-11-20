import type { KeyboardEvent, ReactNode } from 'react'
import { IconButton } from './IconButton'

export type DialogProps = {
  open: boolean
  title: string
  onClose: () => void
  children: ReactNode
  footer?: ReactNode
}

export function Dialog({ open, title, onClose, children, footer }: DialogProps) {
  if (!open) return null

  function handleKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape') {
      event.preventDefault()
      onClose()
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" role="presentation">
      <div
        role="dialog"
        aria-modal="true"
        aria-label={title}
        tabIndex={-1}
        onKeyDown={handleKeyDown}
        className="max-h-[calc(100vh-2rem)] w-full max-w-xl overflow-y-auto rounded-[var(--ss-radius-card)] border border-[var(--ss-color-border)] bg-[var(--ss-color-surface)] shadow-xl"
      >
        <div className="flex items-center justify-between border-b border-[var(--ss-color-border)] px-4 py-3">
          <h2 className="text-lg font-semibold text-[var(--ss-color-text)]">{title}</h2>
          <IconButton label="Đóng" onClick={onClose} autoFocus>
            <span aria-hidden="true" className="text-xl leading-none">
              ×
            </span>
          </IconButton>
        </div>
        <div className="p-4">{children}</div>
        {footer && <div className="flex flex-wrap justify-end gap-2 border-t border-[var(--ss-color-border)] px-4 py-3">{footer}</div>}
      </div>
    </div>
  )
}
