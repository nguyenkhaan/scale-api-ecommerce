import type { KeyboardEvent } from 'react'

export function handleDialogKeyDown(event: KeyboardEvent<HTMLDivElement>, onClose: () => void) {
  if (event.key === 'Escape') {
    event.preventDefault()
    onClose()
    return
  }
  if (event.key !== 'Tab') return

  const focusable = [...event.currentTarget.querySelectorAll<HTMLElement>(
    'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
  )]
  if (focusable.length === 0) {
    event.preventDefault()
    event.currentTarget.focus()
    return
  }

  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  const active = document.activeElement
  if ((event.shiftKey && active === first) || (!event.shiftKey && active === last)) {
    event.preventDefault()
    ;(event.shiftKey ? last : first).focus()
  }
}
